package tests

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem-email/app/controllers"
	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem-email/bootstrap"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/vukemtest"
)

var webhookKey = []byte("0123456789abcdef0123456789abcdef")

func signed(c *vukemtest.Client, id string, at time.Time, body string) *vukemtest.Client {
	ts := strconv.FormatInt(at.Unix(), 10)
	m := hmac.New(sha256.New, webhookKey)
	m.Write([]byte(id + "." + ts + "." + body))
	return c.WithHeader("svix-id", id).WithHeader("svix-timestamp", ts).WithHeader("svix-signature", "v1,"+base64.StdEncoding.EncodeToString(m.Sum(nil)))
}

func post(c *vukemtest.Client, body string) *vukemtest.Response {
	return c.Do(http.MethodPost, "/webhooks/resend", strings.NewReader(body), "application/json")
}

func event(kind, email string, projectID int64, extra string) string {
	return fmt.Sprintf(`{"type":%q,"created_at":"2026-11-07T00:00:00Z","data":{"to":[%q],"tags":{"project":"%d"%s}%s}}`, kind, email, projectID, extra, "")
}

func webhookApp(t *testing.T) (*vukemtest.App, models.Project) {
	t.Helper()
	t.Chdir("..")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"RESEND_WEBHOOK_SECRET": "whsec_" + base64.StdEncoding.EncodeToString(webhookKey)}})
	db, _ := app.DB()
	return app, seedProject(t, db)
}

func subscriberStatus(t *testing.T, app *vukemtest.App, id int64) string {
	t.Helper()
	db, _ := app.DB()
	s, err := orm.QueryOf[models.Subscriber](db).Find(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	return s.Status
}

func TestWebhookRejectsForgedStaleAndUnsignedRequests(t *testing.T) {
	app, p := webhookApp(t)
	db, _ := app.DB()
	s := models.Subscriber{ProjectID: p.ID, Email: "victim@example.com", Status: "confirmed"}
	if err := orm.Create(t.Context(), db, &s); err != nil {
		t.Fatal(err)
	}
	body := event("email.complained", s.Email, p.ID, "")
	post(app.Client(t), body).AssertStatus(401)
	forged := app.Client(t).WithHeader("svix-id", "msg_1").WithHeader("svix-timestamp", strconv.FormatInt(time.Now().Unix(), 10)).WithHeader("svix-signature", "v1,"+base64.StdEncoding.EncodeToString([]byte("forged")))
	post(forged, body).AssertStatus(401)
	post(signed(app.Client(t), "msg_2", time.Now(), body), body+" ").AssertStatus(401)
	post(signed(app.Client(t), "msg_3", time.Now().Add(-10*time.Minute), body), body).AssertStatus(401)
	post(signed(app.Client(t), "msg_4", time.Now().Add(10*time.Minute), body), body).AssertStatus(401)
	if subscriberStatus(t, app, s.ID) != "confirmed" {
		t.Fatal("unverified webhooks must not change subscribers")
	}
}

func TestWebhookSuppressesOnHardBounceAndComplaintOnly(t *testing.T) {
	app, p := webhookApp(t)
	db, _ := app.DB()
	other := models.Project{Name: "Other", Slug: "other", Domain: "news.example.com", FromEmail: "hi@example.com"}
	if err := orm.Create(t.Context(), db, &other); err != nil {
		t.Fatal(err)
	}
	mk := func(pid int64, email string) models.Subscriber {
		s := models.Subscriber{ProjectID: pid, Email: email, Status: "confirmed"}
		if err := orm.Create(t.Context(), db, &s); err != nil {
			t.Fatal(err)
		}
		return s
	}
	hard, soft, spam, bystander, unrelated := mk(p.ID, "hard@example.com"), mk(p.ID, "soft@example.com"), mk(p.ID, "spam@example.com"), mk(other.ID, "hard@example.com"), mk(p.ID, "fine@example.com")
	send := func(id, body string) {
		t.Helper()
		post(signed(app.Client(t), id, time.Now(), body), body).AssertStatus(200)
	}
	bounce := func(typ string) string {
		return `"bounce":{"type":"` + typ + `","subType":"General"}`
	}
	withBounce := func(email, typ string) string {
		return fmt.Sprintf(`{"type":"email.bounced","data":{"to":[%q],"tags":{"project":"%d"},%s}}`, email, p.ID, bounce(typ))
	}
	send("evt_soft", withBounce(soft.Email, "Transient"))
	send("evt_hard", withBounce(hard.Email, "Permanent"))
	send("evt_spam", event("email.complained", spam.Email, p.ID, ""))
	send("evt_unknown", `{"type":"email.opened","data":{"to":["fine@example.com"]}}`)
	for id, want := range map[int64]string{hard.ID: "suppressed", spam.ID: "suppressed", soft.ID: "confirmed", bystander.ID: "confirmed", unrelated.ID: "confirmed"} {
		if got := subscriberStatus(t, app, id); got != want {
			t.Fatalf("subscriber %d = %s, want %s", id, got, want)
		}
	}
	n, _ := orm.QueryOf[models.Suppression](db).Count(t.Context())
	if n != 2 {
		t.Fatalf("suppressions = %d", n)
	}
	app.Client(t).PostJSON("/subscribe/hexascore", map[string]string{"email": hard.Email}).AssertStatus(202)
	if len(app.Mails().SentTo(hard.Email)) != 0 || subscriberStatus(t, app, hard.ID) != "suppressed" {
		t.Fatal("suppressed address must not be re-sent a confirmation")
	}
}

func TestWebhookDedupesRedeliveryAndUpdatesDelivery(t *testing.T) {
	app, p := webhookApp(t)
	db, _ := app.DB()
	s := models.Subscriber{ProjectID: p.ID, Email: "dup@example.com", Status: "confirmed"}
	c := models.Campaign{ProjectID: p.ID, Subject: "S", Mode: "markdown", Source: "x", BodyHTML: "x", Status: "completed"}
	if err := orm.Create(t.Context(), db, &s); err != nil {
		t.Fatal(err)
	}
	if err := orm.Create(t.Context(), db, &c); err != nil {
		t.Fatal(err)
	}
	d := models.Delivery{CampaignID: c.ID, SubscriberID: s.ID, Status: "accepted"}
	if err := orm.Create(t.Context(), db, &d); err != nil {
		t.Fatal(err)
	}
	delivered := fmt.Sprintf(`{"type":"email.delivered","data":{"to":[%q],"tags":{"project":"%d","delivery":"%d"}}}`, s.Email, p.ID, d.ID)
	post(signed(app.Client(t), "evt_d", time.Now(), delivered), delivered).AssertStatus(200)
	got, _ := orm.QueryOf[models.Delivery](db).Find(t.Context(), d.ID)
	if got.Status != "delivered" {
		t.Fatalf("delivery = %s", got.Status)
	}
	body := event("email.complained", s.Email, p.ID, "")
	post(signed(app.Client(t), "evt_c", time.Now(), body), body).AssertStatus(200)
	if _, err := orm.QueryOf[models.Subscriber](db).Where("id", "=", s.ID).Update(t.Context(), orm.Set{"status": "confirmed"}); err != nil {
		t.Fatal(err)
	}
	post(signed(app.Client(t), "evt_c", time.Now(), body), body).AssertStatus(200).AssertHeader("Webhook-Duplicate", "true")
	if subscriberStatus(t, app, s.ID) != "confirmed" {
		t.Fatal("a replayed event id must not be processed twice")
	}
}

func TestDeliverySkipsSuppressedAddress(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	s := models.Subscriber{ProjectID: p.ID, Email: "blocked@example.com", Status: "confirmed"}
	c := models.Campaign{ProjectID: p.ID, Subject: "S", Mode: "markdown", Source: "x", BodyHTML: "<p>x</p>", Status: "draft"}
	sup := models.Suppression{ProjectID: p.ID, Email: s.Email, Reason: "hard_bounce"}
	if err := orm.Create(t.Context(), db, &s); err != nil {
		t.Fatal(err)
	}
	if err := orm.Create(t.Context(), db, &c); err != nil {
		t.Fatal(err)
	}
	if err := orm.Create(t.Context(), db, &sup); err != nil {
		t.Fatal(err)
	}
	admin(t, app).PostJSON("/api/campaigns/"+strconv.FormatInt(c.ID, 10)+"/send", map[string]string{}).AssertStatus(202)
	if len(app.Mails().Sent()) != 0 {
		t.Fatal("suppressed address must be skipped even if still marked confirmed")
	}
}

func TestWebhookDisabledWithoutSecret(t *testing.T) {
	app, _ := setup(t)
	post(app.Client(t), "{}").AssertStatus(404)
}

func TestWebhookNeverRegressesDeliveryState(t *testing.T) {
	app, p := webhookApp(t)
	db, _ := app.DB()
	s := models.Subscriber{ProjectID: p.ID, Email: "order@example.com", Status: "confirmed"}
	c := models.Campaign{ProjectID: p.ID, Subject: "S", Mode: "markdown", Source: "x", BodyHTML: "x", Status: "completed"}
	if err := orm.Create(t.Context(), db, &s); err != nil {
		t.Fatal(err)
	}
	if err := orm.Create(t.Context(), db, &c); err != nil {
		t.Fatal(err)
	}
	d := models.Delivery{CampaignID: c.ID, SubscriberID: s.ID, Status: "accepted"}
	if err := orm.Create(t.Context(), db, &d); err != nil {
		t.Fatal(err)
	}
	send := func(id, kind, extra string) {
		t.Helper()
		body := fmt.Sprintf(`{"type":%q,"data":{"to":[%q],"tags":{"project":"%d","delivery":"%d"}%s}}`, kind, s.Email, p.ID, d.ID, extra)
		post(signed(app.Client(t), id, time.Now(), body), body).AssertStatus(200)
	}
	status := func() string {
		got, _ := orm.QueryOf[models.Delivery](db).Find(t.Context(), d.ID)
		return got.Status
	}
	hard := `,"bounce":{"type":"Permanent"}`
	for _, step := range []struct{ id, kind, extra, want string }{
		{"o1", "email.bounced", hard, "bounced"},
		{"o2", "email.delivered", "", "bounced"},
		{"o3", "email.complained", "", "complained"},
		{"o4", "email.bounced", hard, "complained"},
		{"o5", "email.delivered", "", "complained"},
	} {
		send(step.id, step.kind, step.extra)
		if got := status(); got != step.want {
			t.Fatalf("after %s delivery = %s, want %s", step.kind, got, step.want)
		}
	}
}

func TestWebhookSuppressionIsIdempotent(t *testing.T) {
	app, p := webhookApp(t)
	db, _ := app.DB()
	body := fmt.Sprintf(`{"type":"email.bounced","data":{"to":["twice@example.com"],"tags":{"project":"%d"},"bounce":{"type":"Permanent"}}}`, p.ID)
	complaint := event("email.complained", "twice@example.com", p.ID, "")
	post(signed(app.Client(t), "s1", time.Now(), body), body).AssertStatus(200)
	post(signed(app.Client(t), "s2", time.Now(), body), body).AssertStatus(200)
	post(signed(app.Client(t), "s3", time.Now(), complaint), complaint).AssertStatus(200)
	rows, err := orm.QueryOf[models.Suppression](db).Where("email", "=", "twice@example.com").Get(t.Context())
	if err != nil || len(rows) != 1 || rows[0].Reason != "hard_bounce" {
		t.Fatalf("suppressions = %+v, %v", rows, err)
	}
}

func TestWebhookLogsBouncesAndSuppressesAfterSoftBounceLimit(t *testing.T) {
	app, p := webhookApp(t)
	db, _ := app.DB()
	s := models.Subscriber{ProjectID: p.ID, Email: "flaky@example.com", Status: "confirmed"}
	if err := orm.Create(t.Context(), db, &s); err != nil {
		t.Fatal(err)
	}
	soft := fmt.Sprintf(`{"type":"email.bounced","data":{"to":[%q],"tags":{"project":"%d"},"bounce":{"type":"Transient","message":"mailbox full"}}}`, s.Email, p.ID)
	for i := 1; i < controllers.SoftBounceLimit; i++ {
		post(signed(app.Client(t), fmt.Sprintf("soft_%d", i), time.Now(), soft), soft).AssertStatus(200)
		if subscriberStatus(t, app, s.ID) != "confirmed" {
			t.Fatalf("suppressed after only %d soft bounces", i)
		}
	}
	// A redelivered event id must not count twice.
	post(signed(app.Client(t), "soft_1", time.Now(), soft), soft).AssertStatus(200)
	if subscriberStatus(t, app, s.ID) != "confirmed" {
		t.Fatal("a duplicate event counted as a new soft bounce")
	}
	post(signed(app.Client(t), "soft_last", time.Now(), soft), soft).AssertStatus(200)
	if subscriberStatus(t, app, s.ID) != "suppressed" {
		t.Fatalf("subscriber must be suppressed after %d soft bounces", controllers.SoftBounceLimit)
	}
	sup, err := orm.QueryOf[models.Suppression](db).Where("email", "=", s.Email).First(t.Context())
	if err != nil || sup.Reason != "soft_bounce_limit" {
		t.Fatalf("suppression = %+v, %v", sup, err)
	}
	rows, err := orm.QueryOf[models.Bounce](db).Where("email", "=", s.Email).Get(t.Context())
	if err != nil || len(rows) != controllers.SoftBounceLimit || rows[0].Type != "soft" || rows[0].Reason != "mailbox full" || rows[0].SubscriberID != s.ID {
		t.Fatalf("bounces = %+v, %v", rows, err)
	}
	hard := fmt.Sprintf(`{"type":"email.bounced","data":{"to":["gone@example.com"],"tags":{"project":"%d"},"bounce":{"type":"Permanent"}}}`, p.ID)
	post(signed(app.Client(t), "hard_1", time.Now(), hard), hard).AssertStatus(200)
	if n, _ := orm.QueryOf[models.Bounce](db).Where("type", "=", "hard").Count(t.Context()); n != 1 {
		t.Fatalf("hard bounces logged = %d", n)
	}
}
