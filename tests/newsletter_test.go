package tests

import (
	"context"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem-email/app/controllers"
	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem-email/app/newsletter"
	"github.com/AlexGalhardo/vukem-email/bootstrap"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/vukemtest"
)

func setup(t *testing.T) (*vukemtest.App, models.Project) {
	t.Helper()
	t.Chdir("..")
	app := vukemtest.New(t, bootstrap.New)
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	return app, seedProject(t, db)
}
func seedProject(t *testing.T, db *database.DB) models.Project {
	t.Helper()
	p := models.Project{Name: "Hexascore", Slug: "hexascore", Domain: "news.hexascore.app", FromEmail: "newsletter@hexascore.app", Description: "Project updates"}
	if err := orm.Create(t.Context(), db, &p); err != nil {
		t.Fatal(err)
	}
	return p
}
func admin(t *testing.T, app *vukemtest.App) *vukemtest.Client {
	return app.Client(t).WithSession(map[string]any{"admin": controllers.AdminEmail})
}
func TestSecurity_AdminAccess(t *testing.T) {
	app, _ := setup(t)
	for _, path := range []string{"/", "/subscribers", "/campaigns", "/campaigns/new", "/projects", "/settings"} {
		app.Client(t).Get(path).AssertRedirect("/login")
	}
	for _, path := range []string{"/api/projects", "/api/campaigns", "/api/campaigns/1/send", "/api/preview", "/api/subscribers/1/unsubscribe"} {
		app.Client(t).PostJSON(path, map[string]string{}).AssertStatus(401)
	}
	app.Client(t).GetJSON("/api/subscribers/export?project=1").AssertStatus(401)
	app.Client(t).WithSession(map[string]any{"admin": "intruder@example.com"}).Get("/").AssertRedirect("/login")
	admin(t, app).Get("/").AssertStatus(200).AssertInertia("Dashboard").PropEquals("project.name", "Hexascore")
	app.Client(t).Get("/register").AssertStatus(404)
	for _, tc := range []struct {
		email          string
		verified, want bool
	}{{controllers.AdminEmail, true, true}, {controllers.AdminEmail, false, false}, {"intruder@example.com", true, false}, {"ALEEXGVIEIRA@gmail.com", true, true}, {"aleexgvieira+admin@gmail.com", true, false}} {
		if got := controllers.AllowedGoogle(tc.email, tc.verified); got != tc.want {
			t.Fatalf("allow %q verified=%v = %v", tc.email, tc.verified, got)
		}
	}
}
func TestDoubleOptInAndTokenLifecycle(t *testing.T) {
	app, p := setup(t)
	client := app.Client(t)
	db, _ := app.DB()
	client.PostJSON("/subscribe/hexascore", map[string]string{"email": "ANA@example.com"}).AssertStatus(202)
	sent := app.Mails().SentTo("ana@example.com")
	if len(sent) != 1 {
		t.Fatalf("sent %d messages", len(sent))
	}
	match := regexp.MustCompile(`https://news.hexascore.app/confirm\?token=([^"< ]+)`).FindStringSubmatch(sent[0].HTML)
	if len(match) != 2 {
		t.Fatal("missing confirmation URL")
	}
	token := html.UnescapeString(match[1])
	s, err := orm.QueryOf[models.Subscriber](db).Where("project_id", "=", p.ID).First(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if s.Status != "pending" || s.TokenHash == token || s.TokenHash != newsletter.Hash(token) {
		t.Fatal("token must be hashed; subscriber must start pending")
	}
	client.Get("/confirm?token=" + token).AssertStatus(200)
	s, _ = orm.QueryOf[models.Subscriber](db).Find(t.Context(), s.ID)
	if s.Status != "pending" {
		t.Fatal("GET must not confirm email")
	}
	client.PostJSON("/confirm", map[string]string{"token": token}).AssertStatus(200)
	client.PostJSON("/confirm", map[string]string{"token": token}).AssertStatus(422)
	client.PostJSON("/subscribe/hexascore", map[string]string{"email": "ana@example.com"}).AssertStatus(202)
	if len(app.Mails().SentTo("ana@example.com")) != 1 {
		t.Fatal("confirmed/rapid duplicate signups must not send more mail")
	}
	s, _ = orm.QueryOf[models.Subscriber](db).Find(t.Context(), s.ID)
	if s.Status != "confirmed" || s.TokenHash != "" {
		t.Fatal("confirmation must consume token")
	}
	key, _ := app.DeriveKey("newsletter.unsubscribe", 32)
	signature := newsletter.UnsubscribeToken(key, s.ID)
	id := strconv.FormatInt(s.ID, 10)
	client.PostJSON("/unsubscribe", map[string]string{"subscriber": id, "token": strings.Repeat("a", 64)}).AssertStatus(403)
	client.Get("/unsubscribe?subscriber=" + id + "&token=" + signature).AssertStatus(200)
	s, _ = orm.QueryOf[models.Subscriber](db).Find(t.Context(), s.ID)
	if s.Status != "confirmed" {
		t.Fatal("GET must not unsubscribe email")
	}
	client.PostJSON("/unsubscribe", map[string]string{"subscriber": id, "token": signature}).AssertStatus(200)
	s, _ = orm.QueryOf[models.Subscriber](db).Find(t.Context(), s.ID)
	if s.Status != "unsubscribed" {
		t.Fatal("unsubscribe failed")
	}
}
func TestExpiredAndRotatedTokens(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	s := models.Subscriber{ProjectID: p.ID, Email: "expired@example.com", Status: "pending", TokenHash: newsletter.Hash("this-is-an-expired-test-token"), ExpiresAt: time.Now().Add(-time.Hour).Unix()}
	if err := orm.Create(t.Context(), db, &s); err != nil {
		t.Fatal(err)
	}
	app.Client(t).PostJSON("/confirm", map[string]string{"token": "this-is-an-expired-test-token"}).AssertStatus(422)
	app.Client(t).PostJSON("/subscribe/hexascore", map[string]string{"email": s.Email}).AssertStatus(202)
	changed, _ := orm.QueryOf[models.Subscriber](db).Find(t.Context(), s.ID)
	if changed.TokenHash == s.TokenHash || changed.ExpiresAt <= time.Now().Unix() {
		t.Fatal("resubscribe must rotate expired token")
	}
	app.Client(t).PostJSON("/confirm", map[string]string{"token": "this-is-an-expired-test-token"}).AssertStatus(422)
}
func TestProjectIsolationAndCampaignEligibility(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	ctx := t.Context()
	other := models.Project{Name: "Other", Slug: "other", Domain: "news.example.com", FromEmail: "hi@example.com"}
	if err := orm.Create(ctx, db, &other); err != nil {
		t.Fatal(err)
	}
	for _, s := range []models.Subscriber{{ProjectID: p.ID, Email: "yes@example.com", Status: "confirmed"}, {ProjectID: p.ID, Email: "pending@example.com", Status: "pending"}, {ProjectID: p.ID, Email: "removed@example.com", Status: "unsubscribed"}, {ProjectID: other.ID, Email: "other@example.com", Status: "confirmed"}} {
		if err := orm.Create(ctx, db, &s); err != nil {
			t.Fatal(err)
		}
	}
	c := models.Campaign{ProjectID: p.ID, Subject: "Hello", Mode: "markdown", Source: "Hello", BodyHTML: "<p>Only to confirmed subscribers.</p>", Status: "draft"}
	if err := orm.Create(ctx, db, &c); err != nil {
		t.Fatal(err)
	}
	client := admin(t, app)
	path := "/api/campaigns/" + strconv.FormatInt(c.ID, 10) + "/send"
	client.PostJSON(path, map[string]string{}).AssertStatus(202)
	client.PostJSON(path, map[string]string{}).AssertStatus(409)
	if len(app.Mails().Sent()) != 1 {
		t.Fatalf("campaign must send to exactly one eligible subscriber: %d", len(app.Mails().Sent()))
	}
	sent := app.Mails().SentTo("yes@example.com")
	if len(sent) != 1 || !strings.Contains(sent[0].HTML, "unsubscribe?") {
		t.Fatal("missing recipient or unsubscribe link")
	}
	client.GetJSON("/api/subscribers/export?project=" + strconv.FormatInt(p.ID, 10)).AssertStatus(200)
	body := client.GetJSON("/api/subscribers/export?project=" + strconv.FormatInt(p.ID, 10)).Body()
	if strings.Contains(body, "other@example.com") {
		t.Fatal("cross-project CSV leak")
	}
	client.Get("/?project="+strconv.FormatInt(other.ID, 10)).AssertInertia("Dashboard").PropEquals("stats.confirmed", 1)
}
func TestEmptyCampaignRollsBack(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	c := models.Campaign{ProjectID: p.ID, Subject: "Empty", Mode: "markdown", Source: "Hello", BodyHTML: "<p>Hello</p>", Status: "draft"}
	if err := orm.Create(t.Context(), db, &c); err != nil {
		t.Fatal(err)
	}
	admin(t, app).PostJSON("/api/campaigns/"+strconv.FormatInt(c.ID, 10)+"/send", map[string]string{}).AssertStatus(422)
	c, _ = orm.QueryOf[models.Campaign](db).Find(t.Context(), c.ID)
	if c.Status != "draft" {
		t.Fatal("empty audience must leave campaign editable")
	}
}
func TestSubscriptionValidationAndHoneypot(t *testing.T) {
	app, _ := setup(t)
	app.Client(t).PostJSON("/subscribe/hexascore", map[string]string{"email": "not-an-email"}).AssertStatus(422)
	app.Client(t).PostJSON("/subscribe/hexascore", map[string]string{"email": "bot@example.com", "website": "https://bot.example"}).AssertStatus(202)
	if len(app.Mails().Sent()) != 0 {
		t.Fatal("honeypot must suppress mail")
	}
	app.Client(t).Get("https://news.hexascore.app/subscribe").AssertStatus(200).AssertInertia("Subscribe").PropEquals("project.slug", "hexascore")
	app.Client(t).PostJSON("/subscribe/missing", map[string]string{"email": "a@example.com"}).AssertStatus(404)
	app.Client(t).WithHeader("Sec-Fetch-Site", "cross-site").PostJSON("/subscribe/hexascore", map[string]string{"email": "a@example.com"}).AssertStatus(403)
}
func TestPublicPageContainsNoSubscriberData(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	s := models.Subscriber{ProjectID: p.ID, Email: "private@example.com", Status: "confirmed"}
	if err := orm.Create(context.Background(), db, &s); err != nil {
		t.Fatal(err)
	}
	body := app.Client(t).Get("/subscribe/hexascore").AssertStatus(200).Body()
	if strings.Contains(body, s.Email) {
		t.Fatal("subscriber leaked on public landing page")
	}
	// A query string must never grant dashboard access.
	app.Client(t).Get("/?admin=" + url.QueryEscape(controllers.AdminEmail)).AssertRedirect("/login")
}

func TestEditorPreviewAndSave(t *testing.T) {
	app, p := setup(t)
	client := admin(t, app)
	data := map[string]string{"project_id": strconv.FormatInt(p.ID, 10), "subject": "Preview test", "mode": "mdx", "source": "# Hello\n\n<Button href=\"https://example.com\">Read more</Button>"}
	r := client.PostJSON("/api/preview", data).AssertStatus(200)
	var preview struct {
		URL string `json:"url"`
	}
	r.JSON(&preview)
	doc := client.Get(preview.URL).AssertStatus(200)
	if !strings.Contains(doc.Body(), "Read more") {
		t.Fatal("preview missing rendered content")
	}
	app.Client(t).GetJSON(preview.URL).AssertStatus(401)
	if !strings.Contains(doc.Header("Content-Security-Policy"), "script-src 'none'") {
		t.Fatal("preview scripts must be disabled")
	}
	main := client.Get("/").AssertStatus(200)
	if strings.Contains(main.Header("Content-Security-Policy"), "unsafe-inline") {
		t.Fatal("global CSP must remain strict")
	}

	var saved models.Campaign
	client.PostJSON("/api/campaigns", data).AssertStatus(200).JSON(&saved)
	if saved.ID == 0 || saved.Status != "draft" {
		t.Fatal("missing saved draft")
	}
	data["source"] = "{process.env.SECRET}"
	client.PostJSON("/api/preview", data).AssertStatus(422)
	client.PostJSON("/api/campaigns", data).AssertStatus(422)
}
func TestDurableQueueRechecksUnsubscribe(t *testing.T) {
	t.Chdir("..")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"QUEUE_CONNECTION": "database", "QUEUE_IN_SERVE": "false", "CACHE_STORE": "database"}})
	db, _ := app.DB()
	ctx := t.Context()
	p := models.Project{Name: "Durable", Slug: "durable", Domain: "news.example.com", FromEmail: "newsletter@example.com"}
	if err := orm.Create(ctx, db, &p); err != nil {
		t.Fatal(err)
	}
	s := models.Subscriber{ProjectID: p.ID, Email: "confirmed@example.com", Status: "confirmed"}
	if err := orm.Create(ctx, db, &s); err != nil {
		t.Fatal(err)
	}
	c := models.Campaign{ProjectID: p.ID, Subject: "Check unsubscribe", Mode: "markdown", Source: "Hi", BodyHTML: "<p>Hi</p>", Status: "draft"}
	if err := orm.Create(ctx, db, &c); err != nil {
		t.Fatal(err)
	}
	client := admin(t, app)
	client.PostJSON("/api/campaigns/"+strconv.FormatInt(c.ID, 10)+"/send", map[string]string{}).AssertStatus(202)
	if len(app.Mails().Sent()) != 0 {
		t.Fatal("persistent queue should not send during request")
	}
	client.PostJSON("/api/subscribers/"+strconv.FormatInt(s.ID, 10)+"/unsubscribe", map[string]string{}).AssertStatus(200)
	q, err := app.Queue()
	if err != nil {
		t.Fatal(err)
	}
	if err := q.Work(ctx, queue.WorkOptions{StopWhenEmpty: true, Concurrency: 1}); err != nil {
		t.Fatal(err)
	}
	if len(app.Mails().Sent()) != 0 {
		t.Fatal("queued mail must honor unsubscribe before worker sends")
	}
	d, err := orm.QueryOf[models.Delivery](db).Where("campaign_id", "=", c.ID).First(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != "skipped" {
		t.Fatalf("got delivery status %q", d.Status)
	}
}
