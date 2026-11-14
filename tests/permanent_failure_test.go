package tests

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem-email/bootstrap"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/vukemtest"
)

func TestPermanentResendFailureSettlesDeliveryAndCompletesCampaign(t *testing.T) {
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		_, _ = w.Write([]byte(`{"message":"invalid"}`))
	}))
	defer fake.Close()
	t.Chdir("..")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"QUEUE_CONNECTION": "memory", "MAIL_MAILER": "resend", "RESEND_API_KEY": "re_test", "RESEND_BASE_URL": fake.URL}})
	db, _ := app.DB()
	p := seedProject(t, db)
	s := models.Subscriber{ProjectID: p.ID, Email: "perm@example.com", Status: "confirmed"}
	c := models.Campaign{ProjectID: p.ID, Subject: "Perm", Mode: "markdown", Source: "Hi", BodyHTML: "<p>Hi</p>", Status: "draft"}
	if err := orm.Create(t.Context(), db, &s); err != nil {
		t.Fatal(err)
	}
	if err := orm.Create(t.Context(), db, &c); err != nil {
		t.Fatal(err)
	}
	admin(t, app).PostJSON("/api/campaigns/"+strconv.FormatInt(c.ID, 10)+"/send", map[string]string{}).AssertStatus(202)
	q, err := app.Queue()
	if err != nil {
		t.Fatal(err)
	}
	// A real worker retries; only attempt 1 runs here, so the permanent error must settle the delivery itself.
	if err := q.Work(t.Context(), queue.WorkOptions{StopWhenEmpty: true, MaxJobs: 1}); err != nil {
		t.Fatal(err)
	}
	d, err := orm.QueryOf[models.Delivery](db).Where("campaign_id", "=", c.ID).First(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != "failed" {
		t.Fatalf("delivery = %s, want failed", d.Status)
	}
	got, _ := orm.QueryOf[models.Campaign](db).Find(t.Context(), c.ID)
	if got.Status != "completed" {
		t.Fatalf("campaign = %s, want completed", got.Status)
	}
}
