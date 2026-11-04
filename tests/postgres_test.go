package tests

import (
	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem-email/bootstrap"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/vukemtest"
	"os"
	"strconv"
	"testing"
)

func TestPostgresWorkflow(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL must point to a disposable PostgreSQL database")
	}
	t.Chdir("..")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": databaseURL, "QUEUE_CONNECTION": "database", "QUEUE_IN_SERVE": "false", "CACHE_STORE": "database"}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	p := models.Project{Name: "Postgres", Slug: "postgres", Domain: "news.example.com", FromEmail: "hi@example.com"}
	if err := orm.Create(t.Context(), db, &p); err != nil {
		t.Fatal(err)
	}
	client := app.Client(t)
	client.PostJSON("/subscribe/postgres", map[string]string{"email": "pg@example.com"}).AssertStatus(202)
	s, err := orm.QueryOf[models.Subscriber](db).Where("project_id", "=", p.ID).First(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	q, err := app.Queue()
	if err != nil {
		t.Fatal(err)
	}
	if err := q.Work(t.Context(), queue.WorkOptions{StopWhenEmpty: true, Concurrency: 1}); err != nil {
		t.Fatal(err)
	}
	if len(app.Mails().SentTo(s.Email)) != 1 {
		t.Fatal("PostgreSQL confirmation queue did not deliver")
	}
	if _, err := orm.QueryOf[models.Subscriber](db).Where("id", "=", s.ID).Update(t.Context(), orm.Set{"status": "confirmed", "token_hash": ""}); err != nil {
		t.Fatal(err)
	}
	c := models.Campaign{ProjectID: p.ID, Subject: "Postgres works", Mode: "markdown", Source: "Hi", BodyHTML: "<p>Hello Postgres</p>", Status: "draft"}
	if err := orm.Create(t.Context(), db, &c); err != nil {
		t.Fatal(err)
	}
	admin(t, app).PostJSON("/api/campaigns/"+strconv.FormatInt(c.ID, 10)+"/send", map[string]string{}).AssertStatus(202)
	if err := q.Work(t.Context(), queue.WorkOptions{StopWhenEmpty: true, Concurrency: 1}); err != nil {
		t.Fatal(err)
	}
	d, err := orm.QueryOf[models.Delivery](db).Where("campaign_id", "=", c.ID).First(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != "accepted" {
		t.Fatalf("delivery = %q", d.Status)
	}
	admin(t, app).Get("/?project="+strconv.FormatInt(p.ID, 10)).AssertInertia("Dashboard").PropEquals("stats.confirmed", 1)
}
