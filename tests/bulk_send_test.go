package tests

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/orm"
)

func TestSendCampaignCreatesEveryDeliveryAcrossBatches(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	const n = 1203
	subs := make([]models.Subscriber, n)
	for i := range subs {
		subs[i] = models.Subscriber{ProjectID: p.ID, Email: fmt.Sprintf("bulk%d@example.com", i), Status: "confirmed"}
	}
	for i := 0; i < n; i += 400 {
		if err := orm.CreateMany(t.Context(), db, subs[i:min(i+400, n)]); err != nil {
			t.Fatal(err)
		}
	}
	// A pending subscriber must not receive the campaign.
	if err := orm.Create(t.Context(), db, &models.Subscriber{ProjectID: p.ID, Email: "pending@example.com", Status: "pending"}); err != nil {
		t.Fatal(err)
	}
	c := models.Campaign{ProjectID: p.ID, Subject: "Bulk", Mode: "markdown", Source: "x", BodyHTML: "<p>x</p>", Status: "draft"}
	if err := orm.Create(t.Context(), db, &c); err != nil {
		t.Fatal(err)
	}
	admin(t, app).PostJSON("/api/campaigns/"+strconv.FormatInt(c.ID, 10)+"/send", map[string]string{}).AssertStatus(202)
	got, _ := orm.QueryOf[models.Campaign](db).Find(t.Context(), c.ID)
	if got.Recipients != n {
		t.Fatalf("recipients = %d, want %d", got.Recipients, n)
	}
	total, _ := orm.QueryOf[models.Delivery](db).Where("campaign_id", "=", c.ID).Count(t.Context())
	accepted, _ := orm.QueryOf[models.Delivery](db).Where("campaign_id", "=", c.ID).Where("status", "=", "accepted").Count(t.Context())
	if total != n || accepted != n || len(app.Mails().Sent()) != n {
		t.Fatalf("deliveries = %d, accepted = %d, mails = %d, want %d", total, accepted, len(app.Mails().Sent()), n)
	}
}
