package tests

import (
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/orm"
)

func TestEndToEndSubscriberFlow(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	client := app.Client(t)
	subscribe := func() {
		client.PostJSON("/subscribe/hexascore", map[string]string{"email": "flow@example.com"}).AssertStatus(202)
	}
	confirmURL := regexp.MustCompile(`https://news\.hexascore\.app/confirm\?token=([^"<\s]+)`)

	subscribe()
	msgs := app.Mails().SentTo("flow@example.com")
	if len(msgs) != 1 {
		t.Fatalf("confirmation mails: %d", len(msgs))
	}
	conf := msgs[0]
	if !strings.Contains(conf.Subject, "Confirm your subscription to Hexascore") || conf.Text == "" || conf.Tags["project"] != strconv.FormatInt(p.ID, 10) {
		t.Fatalf("confirmation message: subject=%q text=%d tags=%v", conf.Subject, len(conf.Text), conf.Tags)
	}
	for _, part := range []string{conf.HTML, conf.Text} {
		if !confirmURL.MatchString(part) {
			t.Fatalf("confirmation part lacks the confirm link: %.200s", part)
		}
	}
	token := html.UnescapeString(confirmURL.FindStringSubmatch(conf.HTML)[1])
	client.PostJSON("/confirm", map[string]string{"token": token}).AssertStatus(200)

	c := models.Campaign{ProjectID: p.ID, Subject: "Weekly update", Mode: "markdown", Source: "Hello", BodyHTML: "<p>Hello there</p>", Status: "draft"}
	if err := orm.Create(t.Context(), db, &c); err != nil {
		t.Fatal(err)
	}
	send := func(c models.Campaign) {
		admin(t, app).PostJSON("/api/campaigns/"+strconv.FormatInt(c.ID, 10)+"/send", map[string]string{}).AssertStatus(202)
	}
	send(c)
	var campaignMail = app.Mails().SentTo("flow@example.com")
	if len(campaignMail) != 2 {
		t.Fatalf("expected confirmation + campaign, got %d", len(campaignMail))
	}
	camp := campaignMail[1]
	link := regexp.MustCompile(`https://news\.hexascore\.app/unsubscribe\?subscriber=\d+&(?:amp;)?token=[0-9a-f]{64}`)
	for name, part := range map[string]string{"html": camp.HTML, "text": camp.Text} {
		if !link.MatchString(part) {
			t.Fatalf("campaign %s lacks the unsubscribe link: %.300s", name, part)
		}
	}
	if !strings.Contains(camp.HTML, "Hello there") || !strings.Contains(camp.Text, "Hello there") {
		t.Fatal("campaign body missing from a part")
	}

	u, _ := url.Parse(html.UnescapeString(link.FindString(camp.HTML)))
	q := u.Query()
	client.Get(u.RequestURI()).AssertStatus(200)
	client.PostJSON("/unsubscribe", map[string]string{"subscriber": q.Get("subscriber"), "token": q.Get("token")}).AssertStatus(200)
	s, _ := orm.QueryOf[models.Subscriber](db).Where("email", "=", "flow@example.com").First(t.Context())
	if s.Status != "unsubscribed" {
		t.Fatalf("status after unsubscribe: %s", s.Status)
	}

	// Resubscribe needs a fresh confirmation, and only then campaigns resume.
	if _, err := orm.QueryOf[models.Subscriber](db).Where("id", "=", s.ID).Update(t.Context(), orm.Set{"last_requested_at": int64(0)}); err != nil {
		t.Fatal(err)
	}
	subscribe()
	msgs = app.Mails().SentTo("flow@example.com")
	if len(msgs) != 3 {
		t.Fatalf("resubscribe must send a new confirmation, mails=%d", len(msgs))
	}
	token2 := html.UnescapeString(confirmURL.FindStringSubmatch(msgs[2].HTML)[1])
	if token2 == token {
		t.Fatal("resubscribe must issue a new token")
	}
	client.PostJSON("/confirm", map[string]string{"token": token}).AssertStatus(422)
	client.PostJSON("/confirm", map[string]string{"token": token2}).AssertStatus(200)
	c2 := models.Campaign{ProjectID: p.ID, Subject: "Second", Mode: "markdown", Source: "Hi", BodyHTML: "<p>Welcome back</p>", Status: "draft"}
	if err := orm.Create(t.Context(), db, &c2); err != nil {
		t.Fatal(err)
	}
	send(c2)
	if got := app.Mails().SentTo("flow@example.com"); len(got) != 4 || !strings.Contains(got[3].HTML, "Welcome back") {
		t.Fatalf("resubscribed address must receive the next campaign: %d", len(got))
	}
}
