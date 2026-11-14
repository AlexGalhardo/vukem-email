package tests

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/vukemtest"
)

func issueToken(t *testing.T, app *vukemtest.App, projectID int64, abilities string) (string, string) {
	t.Helper()
	res := admin(t, app).PostJSON("/api/projects/"+strconv.FormatInt(projectID, 10)+"/tokens", map[string]string{"name": "ci", "abilities": abilities}).AssertStatus(201)
	var out struct {
		Token string `json:"token"`
		ID    string `json:"id"`
	}
	res.JSON(&out)
	if !strings.HasPrefix(out.Token, "vke_") {
		t.Fatalf("token response: %+v", out)
	}
	return out.Token, out.ID
}

func bearer(c *vukemtest.Client, token string) *vukemtest.Client {
	return c.WithHeader("Authorization", "Bearer "+token)
}

func TestAPITokensScopeIsolationAndRevocation(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	other := models.Project{Name: "Other", Slug: "other", Domain: "news.other.app", FromEmail: "hi@other.app"}
	if err := orm.Create(t.Context(), db, &other); err != nil {
		t.Fatal(err)
	}
	pid := strconv.FormatInt(p.ID, 10)

	// Only the admin session can issue or revoke tokens.
	app.Client(t).PostJSON("/api/projects/"+pid+"/tokens", map[string]string{"name": "x", "abilities": "subscribers:write"}).AssertStatus(401)
	admin(t, app).PostJSON("/api/projects/"+pid+"/tokens", map[string]string{"name": "x", "abilities": "*"}).AssertStatus(422)
	admin(t, app).PostJSON("/api/projects/"+pid+"/tokens", map[string]string{"name": "x"}).AssertStatus(422)

	writer, writerID := issueToken(t, app, p.ID, "subscribers:write")
	sender, _ := issueToken(t, app, p.ID, "campaigns:send")
	foreign, _ := issueToken(t, app, other.ID, "subscribers:write,campaigns:send")

	// The plain token is never stored.
	rows, err := db.QueryContext(t.Context(), "SELECT hash, abilities FROM access_tokens")
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var h, a string
		if err := rows.Scan(&h, &a); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(writer, h) || len(h) != 64 {
			t.Fatalf("stored hash looks like the secret: %s", h)
		}
	}
	rows.Close()

	body := map[string]string{"email": "New@Example.com"}
	app.Client(t).PostJSON("/api/v1/subscribers", body).AssertStatus(401).AssertHeader("WWW-Authenticate", `Bearer realm="api"`)
	bearer(app.Client(t), "vke_bad.token").PostJSON("/api/v1/subscribers", body).AssertStatus(401)
	// Ability check: a send-only token cannot write subscribers.
	bearer(app.Client(t), sender).PostJSON("/api/v1/subscribers", body).AssertStatus(403)
	bearer(app.Client(t), writer).PostJSON("/api/v1/subscribers", map[string]string{"email": "nope"}).AssertStatus(422)

	res := bearer(app.Client(t), writer).PostJSON("/api/v1/subscribers", body).AssertStatus(202)
	res.AssertJSONPath("status", "pending")
	res.AssertJSONPath("email", "new@example.com")
	if len(app.Mails().SentTo("new@example.com")) != 1 {
		t.Fatal("API subscribe must send the double opt-in confirmation")
	}
	sub, err := orm.QueryOf[models.Subscriber](db).Where("email", "=", "new@example.com").First(t.Context())
	if err != nil || sub.ProjectID != p.ID || sub.Status != "pending" {
		t.Fatalf("subscriber = %+v, %v", sub, err)
	}
	sid := strconv.FormatInt(sub.ID, 10)

	// A token of another project cannot touch this subscriber or campaign.
	bearer(app.Client(t), foreign).Do(http.MethodPatch, "/api/v1/subscribers/"+sid, strings.NewReader(`{"status":"unsubscribed"}`), "application/json").AssertStatus(404)
	bearer(app.Client(t), writer).Do(http.MethodPatch, "/api/v1/subscribers/"+sid, strings.NewReader(`{"status":"confirmed"}`), "application/json").AssertStatus(422)
	bearer(app.Client(t), writer).Do(http.MethodPatch, "/api/v1/subscribers/"+sid, strings.NewReader(`{"status":"unsubscribed"}`), "application/json").AssertStatus(200).AssertJSONPath("status", "unsubscribed")

	c := models.Campaign{ProjectID: p.ID, Subject: "S", Mode: "markdown", Source: "x", BodyHTML: "<p>x</p>", Status: "draft"}
	live := models.Subscriber{ProjectID: p.ID, Email: "live@example.com", Status: "confirmed"}
	for _, v := range []any{&c, &live} {
		switch m := v.(type) {
		case *models.Campaign:
			if err := orm.Create(t.Context(), db, m); err != nil {
				t.Fatal(err)
			}
		case *models.Subscriber:
			if err := orm.Create(t.Context(), db, m); err != nil {
				t.Fatal(err)
			}
		}
	}
	send := "/api/v1/campaigns/" + strconv.FormatInt(c.ID, 10) + "/send"
	bearer(app.Client(t), writer).PostJSON(send, map[string]string{}).AssertStatus(403)
	bearer(app.Client(t), foreign).PostJSON(send, map[string]string{}).AssertStatus(404)
	bearer(app.Client(t), sender).PostJSON(send, map[string]string{}).AssertStatus(202)
	if got, _ := orm.QueryOf[models.Campaign](db).Find(t.Context(), c.ID); got.Status == "draft" {
		t.Fatal("campaign was not queued")
	}

	// Revocation is immediate.
	admin(t, app).PostJSON("/api/projects/"+pid+"/tokens/"+writerID+"/revoke", map[string]string{}).AssertStatus(200)
	bearer(app.Client(t), writer).PostJSON("/api/v1/subscribers", body).AssertStatus(401)
	admin(t, app).PostJSON("/api/projects/"+strconv.FormatInt(other.ID, 10)+"/tokens/"+writerID+"/revoke", map[string]string{}).AssertStatus(404)
}

func TestAPIRateLimitsPerToken(t *testing.T) {
	app, p := setup(t)
	token, _ := issueToken(t, app, p.ID, "subscribers:write")
	var limited bool
	for range 70 {
		if app.Client(t).WithHeader("Authorization", "Bearer "+token).PostJSON("/api/v1/subscribers", map[string]string{"email": "x"}).Status() == 429 {
			limited = true
			break
		}
	}
	if !limited {
		t.Fatal("expected a 429 after 60 requests per minute for one token")
	}
}
