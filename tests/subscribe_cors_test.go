package tests

import (
	"net/http"
	"testing"

	"github.com/AlexGalhardo/vukem-email/bootstrap"
	"github.com/AlexGalhardo/vukem/vukemtest"
)

func corsApp(t *testing.T) *vukemtest.App {
	t.Helper()
	t.Chdir("..")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"NEWSLETTER_ALLOWED_ORIGINS": "https://hexascore.app, https://www.hexascore.app"}})
	db, _ := app.DB()
	seedProject(t, db)
	return app
}

func TestSubscribeCORSAllowedOrigin(t *testing.T) {
	app := corsApp(t)
	pre := app.Client(t).WithHeader("Origin", "https://hexascore.app").WithHeader("Access-Control-Request-Method", "POST").WithHeader("Access-Control-Request-Headers", "content-type").Do(http.MethodOptions, "/subscribe/hexascore", nil, "")
	pre.AssertStatus(204).AssertHeader("Access-Control-Allow-Origin", "https://hexascore.app").AssertHeaderContains("Access-Control-Allow-Methods", "POST").AssertHeaderMissing("Access-Control-Allow-Credentials")
	r := app.Client(t).WithHeader("Origin", "https://hexascore.app").WithHeader("Sec-Fetch-Site", "cross-site").PostJSON("/subscribe/hexascore", map[string]string{"email": "web@example.com"})
	r.AssertStatus(202).AssertHeader("Access-Control-Allow-Origin", "https://hexascore.app").AssertHeaderMissing("Access-Control-Allow-Credentials").AssertHeaderMissing("Set-Cookie")
	r.AssertJSONPath("message", "If this address can be subscribed, you will receive a confirmation link. Check your spam folder too.")
	if len(app.Mails().SentTo("web@example.com")) != 1 {
		t.Fatal("cross-origin subscribe must send the confirmation email")
	}
	app.Client(t).WithHeader("Origin", "https://www.hexascore.app").PostJSON("/subscribe/hexascore", map[string]string{"email": "bad"}).AssertStatus(422).AssertHeader("Access-Control-Allow-Origin", "https://www.hexascore.app")
}

func TestSubscribeCORSRejectsOtherOrigins(t *testing.T) {
	app := corsApp(t)
	for _, origin := range []string{"https://evil.example", "http://hexascore.app", "https://hexascore.app.evil.example", "null"} {
		c := app.Client(t).WithHeader("Origin", origin).WithHeader("Sec-Fetch-Site", "cross-site")
		c.PostJSON("/subscribe/hexascore", map[string]string{"email": "x@example.com"}).AssertStatus(403).AssertHeaderMissing("Access-Control-Allow-Origin")
		pre := app.Client(t).WithHeader("Origin", origin).WithHeader("Access-Control-Request-Method", "POST").Do(http.MethodOptions, "/subscribe/hexascore", nil, "")
		pre.AssertHeaderMissing("Access-Control-Allow-Origin")
	}
	if len(app.Mails().Sent()) != 0 {
		t.Fatal("disallowed origins must not trigger mail")
	}
}

func TestCORSDoesNotReachDashboardOrOtherRoutes(t *testing.T) {
	app := corsApp(t)
	for _, path := range []string{"/", "/api/projects", "/confirm", "/unsubscribe"} {
		r := app.Client(t).WithHeader("Origin", "https://hexascore.app").Do(http.MethodOptions, path, nil, "")
		r.AssertHeaderMissing("Access-Control-Allow-Origin")
	}
	app.Client(t).WithHeader("Origin", "https://hexascore.app").WithHeader("Sec-Fetch-Site", "cross-site").PostJSON("/api/projects", map[string]string{}).AssertStatus(403)
}
