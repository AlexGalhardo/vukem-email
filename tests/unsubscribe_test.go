package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem-email/app/newsletter"
	"github.com/AlexGalhardo/vukem-email/bootstrap"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/vukemtest"
)

func oneClickURL(t *testing.T, app *vukemtest.App, domain string, id int64) string {
	t.Helper()
	key, _ := app.DeriveKey("newsletter.unsubscribe", 32)
	u := newsletter.OneClickURL(domain, id, newsletter.UnsubscribeToken(key, id))
	return strings.TrimPrefix(u, "https://"+domain)
}

func TestOneClickUnsubscribe(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	s := models.Subscriber{ProjectID: p.ID, Email: "one@example.com", Status: "confirmed"}
	if err := orm.Create(t.Context(), db, &s); err != nil {
		t.Fatal(err)
	}
	target := oneClickURL(t, app, p.Domain, s.ID)
	form := func(c *vukemtest.Client, target, body string) *vukemtest.Response {
		return c.Do(http.MethodPost, target, strings.NewReader(body), "application/x-www-form-urlencoded")
	}
	status := func() string {
		got, _ := orm.QueryOf[models.Subscriber](db).Find(t.Context(), s.ID)
		return got.Status
	}
	app.Client(t).Get(target).AssertStatus(405)
	form(app.Client(t), target, "").AssertStatus(400)
	form(app.Client(t), target, "List-Unsubscribe=Nope").AssertStatus(400)
	form(app.Client(t), strings.Replace(target, "token=", "token=00", 1), "List-Unsubscribe=One-Click").AssertStatus(403)
	form(app.Client(t), "/unsubscribe/one-click?subscriber="+strconv.FormatInt(s.ID+1, 10)+"&token="+url.QueryEscape("zz"), "List-Unsubscribe=One-Click").AssertStatus(403)
	if status() != "confirmed" {
		t.Fatal("rejected requests must not unsubscribe")
	}
	// Mailbox providers send no cookies, Origin or Sec-Fetch-Site.
	form(app.Client(t), target, "List-Unsubscribe=One-Click").AssertStatus(200).AssertHeaderMissing("Set-Cookie")
	if status() != "unsubscribed" {
		t.Fatal("one-click POST must unsubscribe")
	}
	form(app.Client(t), target, "List-Unsubscribe=One-Click").AssertStatus(200)
}

func TestCampaignCarriesListUnsubscribeHeaders(t *testing.T) {
	var mu sync.Mutex
	var got []map[string]any
	var keys []string
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		got = append(got, body)
		keys = append(keys, r.Header.Get("Idempotency-Key")+"|"+r.Header.Get("Authorization"))
		mu.Unlock()
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":"em_1"}`))
	}))
	defer fake.Close()
	t.Chdir("..")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"MAIL_MAILER": "resend", "RESEND_API_KEY": "re_test", "RESEND_BASE_URL": fake.URL}})
	db, _ := app.DB()
	p := seedProject(t, db)
	s := models.Subscriber{ProjectID: p.ID, Email: "hdr@example.com", Status: "confirmed"}
	c := models.Campaign{ProjectID: p.ID, Subject: "Headers", Mode: "markdown", Source: "Hi", BodyHTML: "<p>Hi</p>", Status: "draft"}
	if err := orm.Create(t.Context(), db, &s); err != nil {
		t.Fatal(err)
	}
	if err := orm.Create(t.Context(), db, &c); err != nil {
		t.Fatal(err)
	}
	admin(t, app).PostJSON("/api/campaigns/"+strconv.FormatInt(c.ID, 10)+"/send", map[string]string{}).AssertStatus(202)
	mu.Lock()
	defer mu.Unlock()
	if len(got) != 1 {
		t.Fatalf("resend requests: %d", len(got))
	}
	h, _ := got[0]["headers"].(map[string]any)
	post, _ := h["List-Unsubscribe-Post"].(string)
	link, _ := h["List-Unsubscribe"].(string)
	if post != "List-Unsubscribe=One-Click" || !strings.HasPrefix(link, "<https://news.hexascore.app/unsubscribe/one-click?subscriber=") || !strings.HasSuffix(link, ">") {
		t.Fatalf("headers: %v", h)
	}
	if keys[0] != "newsletter-"+strconv.FormatInt(c.ID, 10)+"-"+strconv.FormatInt(s.ID, 10)+"|Bearer re_test" {
		t.Fatalf("idempotency/auth: %s", keys[0])
	}
	// The advertised URL must actually work.
	target := strings.TrimSuffix(strings.TrimPrefix(link, "<https://news.hexascore.app"), ">")
	app.Client(t).Do(http.MethodPost, target, strings.NewReader("List-Unsubscribe=One-Click"), "application/x-www-form-urlencoded").AssertStatus(200)
}
