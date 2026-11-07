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
