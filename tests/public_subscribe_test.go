package tests

import (
	"net/http"
	"testing"

	"github.com/AlexGalhardo/vukem-email/app/controllers"
	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/orm"
)

func TestPublicSubscribeRateLimitedPerIP(t *testing.T) {
	app := corsApp(t)
	var limited bool
	for range 15 {
		r := app.Client(t).WithHeader("Origin", "https://hexascore.app").PostJSON("/subscribe/hexascore", map[string]string{"email": "rl@example.com"})
		if r.Status() == http.StatusTooManyRequests {
			r.AssertHeaderContains("Content-Type", "problem+json")
			limited = true
			break
		}
	}
	if !limited {
		t.Fatal("expected 429 after 10 requests per minute from one IP")
	}
}

func TestPublicSubscribeHoneypotSendsNothing(t *testing.T) {
	app := corsApp(t)
	app.Client(t).WithHeader("Origin", "https://hexascore.app").PostJSON("/subscribe/hexascore", map[string]string{"email": "bot@example.com", "website": "http://spam.example"}).AssertStatus(202)
	if len(app.Mails().Sent()) != 0 {
		t.Fatal("honeypot hit must not send mail")
	}
	db, _ := app.DB()
	if n, _ := orm.QueryOf[models.Subscriber](db).Count(t.Context()); n != 0 {
		t.Fatalf("honeypot hit created %d subscribers", n)
	}
}

func TestSeedProjectTokenIsIdempotentAndWorks(t *testing.T) {
	app, _ := setup(t)
	db, _ := app.DB()
	n := &controllers.NewsletterController{App: app.App, DB: db}
	p := models.Project{Name: "Eletroradar", Slug: "eletroradar", Domain: "eletroradar.test", FromEmail: "news@eletroradar.test"}
	token, err := n.SeedProjectToken(t.Context(), p, []string{"subscribers:write"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := n.SeedProjectToken(t.Context(), p, []string{"subscribers:write"}); err != nil {
		t.Fatal(err)
	}
	if _, err := n.SeedProjectToken(t.Context(), p, []string{"everything"}); err == nil {
		t.Fatal("unknown ability must be refused")
	}
	app.Client(t).WithHeader("Authorization", "Bearer "+token).PostJSON("/api/v1/subscribers", map[string]string{"email": "seed@example.com"}).AssertStatus(202)
	if c, _ := orm.QueryOf[models.Project](db).Where("slug", "=", "eletroradar").Count(t.Context()); c != 1 {
		t.Fatalf("projects = %d", c)
	}
}
