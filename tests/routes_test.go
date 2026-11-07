package tests

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem-email/app/newsletter"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/vukemtest"
)

func TestPublicRoutes(t *testing.T) {
	app, _ := setup(t)
	for path, page := range map[string]string{"/login": "Login", "/subscribe/hexascore": "Subscribe", "https://news.hexascore.app/subscribe": "Subscribe", "/confirm": "SubscriptionAction", "/unsubscribe": "SubscriptionAction"} {
		app.Client(t).Get(path).AssertStatus(200).AssertInertia(page)
	}
	app.Client(t).Get("/confirm?token=abc").AssertHeader("Referrer-Policy", "no-referrer")
	app.Client(t).Get("/unsubscribe?token=abc").AssertHeader("Referrer-Policy", "no-referrer")
	app.Client(t).Get("/subscribe/missing").AssertStatus(404)
	app.Client(t).Get("https://unknown.example.com/subscribe").AssertStatus(404)
	app.Client(t).Get("/robots.txt").AssertStatus(200)
	app.Client(t).Get("/does-not-exist").AssertStatus(404)
	app.Client(t).Get("/auth/google/redirect").AssertStatus(503)
}

func TestDashboardRoutesRequireTheAdmin(t *testing.T) {
	app, p := setup(t)
	id := strconv.FormatInt(p.ID, 10)
	pages := []string{"/", "/subscribers", "/campaigns", "/campaigns/new", "/projects", "/settings"}
	apis := []struct{ method, path string }{
		{"POST", "/api/projects"}, {"POST", "/api/campaigns"}, {"POST", "/api/campaigns/1/send"}, {"POST", "/api/preview"},
		{"POST", "/api/subscribers/1/unsubscribe"}, {"GET", "/api/subscribers/export?project=" + id}, {"GET", "/api/preview/sometoken"},
	}
	sessions := map[string]map[string]any{
		"anonymous": nil,
		"other":     {"admin": "intruder@example.com"},
		"empty":     {"admin": ""},
		"flag":      {"admin": true},
	}
	visitor := func(session map[string]any) *vukemtest.Client {
		c := app.Client(t)
		if session != nil {
			c = c.WithSession(session)
		}
		return c
	}
	for name, session := range sessions {
		for _, path := range pages {
			visitor(session).Get(path).AssertRedirect("/login")
		}
		for _, a := range apis {
			var got int
			if a.method == "POST" {
				got = visitor(session).PostJSON(a.path, map[string]string{}).Status()
			} else {
				got = visitor(session).GetJSON(a.path).Status()
			}
			if got != 401 {
				t.Fatalf("%s %s %s = %d, want 401", name, a.method, a.path, got)
			}
		}
	}
	a := admin(t, app)
	for _, path := range pages {
		a.Get(path).AssertStatus(200).AssertInertia("Dashboard")
	}
	a.Get("/?project=999999").AssertStatus(404)
	a.Get("/api/preview/unknown").AssertStatus(404)
	a.GetJSON("/api/subscribers/export?project=999999").AssertStatus(404)
	a.PostJSON("/api/subscribers/999999/unsubscribe", map[string]string{}).AssertStatus(404)
}

func TestAPIErrorShapes(t *testing.T) {
	app, p := setup(t)
	a := admin(t, app)
	valid := map[string]string{"project_id": strconv.FormatInt(p.ID, 10), "subject": "Hello", "mode": "markdown", "source": "Body"}
	with := func(k, v string) map[string]string {
		m := map[string]string{}
		for key, val := range valid {
			m[key] = val
		}
		m[k] = v
		return m
	}
	project := func(k, v string) map[string]string {
		m := map[string]string{"name": "Ok", "slug": "ok", "domain": "a.example.com", "from_email": "a@example.com"}
		m[k] = v
		return m
	}
	cases := []struct {
		name string
		body map[string]string
		path string
		want int
	}{
		{"subject empty", with("subject", ""), "/api/campaigns", 422},
		{"subject too long", with("subject", strings.Repeat("x", 201)), "/api/campaigns", 422},
		{"subject header injection", with("subject", "Hi\r\nBcc: x@example.com"), "/api/campaigns", 422},
		{"bad mode", with("mode", "html"), "/api/campaigns", 422},
		{"empty source", with("source", ""), "/api/campaigns", 422},
		{"unknown project", with("project_id", "999"), "/api/campaigns", 404},
		{"bad slug", project("slug", "Bad Slug"), "/api/projects", 422},
		{"bad domain", project("domain", "https://a.example.com"), "/api/projects", 422},
		{"bad sender", project("from_email", "nope"), "/api/projects", 422},
		{"duplicate slug", project("slug", "hexascore"), "/api/projects", 409},
		{"duplicate domain", project("domain", "news.hexascore.app"), "/api/projects", 409},
	}
	for _, tc := range cases {
		r := a.PostJSON(tc.path, tc.body).AssertStatus(tc.want)
		r.AssertHeaderContains("Content-Type", "application/problem+json")
		r.AssertJSONPath("status", float64(tc.want))
		if _, ok := r.JSONPath("request_id"); !ok {
			t.Fatalf("%s: problem body lacks request_id", tc.name)
		}
	}
	a.PostJSON("/api/projects", map[string]string{"name": "New", "slug": "new-one", "domain": "new.example.com", "from_email": "hi@example.com"}).AssertStatus(201)
	a.PostJSON("/api/campaigns/999/send", map[string]string{}).AssertStatus(404)
}

func TestSubscribeEdgeCases(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	for _, email := range []string{"", "plain", "@b.co", "a b@example.com", "Name <a@example.com>", strings.Repeat("a", 250) + "@example.com"} {
		app.Client(t).PostJSON("/subscribe/hexascore", map[string]string{"email": email}).AssertStatus(422)
	}
	app.Client(t).Do(http.MethodPost, "/subscribe/hexascore", strings.NewReader("{broken"), "application/json").AssertStatus(400)
	big := `{"email":"a@example.com","pad":"` + strings.Repeat("x", 2<<20) + `"}`
	if got := app.Client(t).Do(http.MethodPost, "/subscribe/hexascore", strings.NewReader(big), "application/json").Status(); got != 413 {
		t.Fatalf("oversized body = %d, want 413", got)
	}
	app.Client(t).PostForm("/subscribe/hexascore", url.Values{"email": {"form@example.com"}}).AssertStatus(202)
	if len(app.Mails().SentTo("form@example.com")) != 1 {
		t.Fatal("form-encoded subscribe must work")
	}
	n, _ := orm.QueryOf[models.Subscriber](db).Where("project_id", "=", p.ID).Count(t.Context())
	if n != 1 {
		t.Fatalf("invalid requests must not create subscribers, have %d", n)
	}
}

func TestTokenEdgeCases(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	client := app.Client(t)
	for _, token := range []string{"", "short", strings.Repeat("a", 129), strings.Repeat("a", 30)} {
		client.PostJSON("/confirm", map[string]string{"token": token}).AssertStatus(422).AssertHeaderContains("Content-Type", "problem+json")
	}
	s := models.Subscriber{ProjectID: p.ID, Email: "tok@example.com", Status: "pending", TokenHash: newsletter.Hash("a-valid-looking-token-12345"), ExpiresAt: time.Now().Add(time.Hour).Unix()}
	if err := orm.Create(t.Context(), db, &s); err != nil {
		t.Fatal(err)
	}
	client.PostJSON("/confirm", map[string]string{"token": "a-valid-looking-token-12346"}).AssertStatus(422)
	client.PostJSON("/confirm", map[string]string{"token": "a-valid-looking-token-12345"}).AssertStatus(200)
	client.PostJSON("/confirm", map[string]string{"token": "a-valid-looking-token-12345"}).AssertStatus(422)

}

// Separate app: the public endpoints share a 10 requests per minute limiter.
func TestUnsubscribeTokenEdgeCases(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	client := app.Client(t)
	s := models.Subscriber{ProjectID: p.ID, Email: "tok@example.com", Status: "confirmed"}
	if err := orm.Create(t.Context(), db, &s); err != nil {
		t.Fatal(err)
	}
	id := strconv.FormatInt(s.ID, 10)
	key, _ := app.DeriveKey("newsletter.unsubscribe", 32)
	good := newsletter.UnsubscribeToken(key, s.ID)
	for name, body := range map[string]map[string]string{
		"missing token":       {"subscriber": id},
		"non hex":             {"subscriber": id, "token": "zzzz"},
		"wrong subscriber":    {"subscriber": strconv.FormatInt(s.ID+1, 10), "token": good},
		"zero subscriber":     {"subscriber": "0", "token": good},
		"negative subscriber": {"subscriber": "-1", "token": good},
		"truncated token":     {"subscriber": id, "token": good[:32]},
		"missing subscriber":  {"token": good},
		"hash of the token":   {"subscriber": id, "token": newsletter.Hash(good)},
	} {
		client.PostJSON("/unsubscribe", body).AssertStatus(403)
		if got := subscriberStatus(t, app, s.ID); got != "confirmed" {
			t.Fatalf("%s changed status to %s", name, got)
		}
	}
	client.PostJSON("/unsubscribe", map[string]string{"subscriber": id, "token": good}).AssertStatus(200)
	if got := subscriberStatus(t, app, s.ID); got != "unsubscribed" {
		t.Fatalf("valid token left status %s", got)
	}
}
