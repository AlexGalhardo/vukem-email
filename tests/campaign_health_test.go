package tests

import (
	"strconv"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/orm"
)

func TestCampaignHealthReportsDeliveriesAndStalls(t *testing.T) {
	app, p := setup(t)
	db, _ := app.DB()
	ctx := t.Context()

	c := models.Campaign{ProjectID: p.ID, Subject: "Health", Mode: "markdown", Source: "Hi", BodyHTML: "<p>Hi</p>", Status: "queued", Recipients: 3}
	if err := orm.Create(ctx, db, &c); err != nil {
		t.Fatal(err)
	}

	for i, status := range []string{"queued", "accepted", "accepted"} {
		sub := models.Subscriber{ProjectID: p.ID, Email: "h" + strconv.Itoa(i) + "@example.com", Status: "confirmed"}
		if err := orm.Create(ctx, db, &sub); err != nil {
			t.Fatal(err)
		}

		d := models.Delivery{CampaignID: c.ID, SubscriberID: sub.ID, Status: status}
		if err := orm.Create(ctx, db, &d); err != nil {
			t.Fatal(err)
		}

		if status == "queued" {
			old := time.Now().Add(-time.Hour).UTC()
			if _, err := orm.QueryOf[models.Delivery](db).Where("id", "=", d.ID).Update(ctx, orm.Set{"created_at": old}); err != nil {
				t.Fatal(err)
			}
		}
	}

	path := "/api/campaigns/" + strconv.FormatInt(c.ID, 10) + "/health"

	app.Client(t).GetJSON(path).AssertStatus(401)
	admin(t, app).GetJSON("/api/campaigns/999999/health").AssertStatus(404)

	res := admin(t, app).GetJSON(path).AssertStatus(200)
	res.AssertJSONPath("recipients", float64(3))
	res.AssertJSONPath("deliveries.queued", float64(1))
	res.AssertJSONPath("deliveries.accepted", float64(2))
	res.AssertJSONPath("stalled", true)

	if _, err := orm.QueryOf[models.Delivery](db).Where("campaign_id", "=", c.ID).Where("status", "=", "queued").Update(ctx, orm.Set{"status": "accepted"}); err != nil {
		t.Fatal(err)
	}

	admin(t, app).GetJSON(path).AssertStatus(200).AssertJSONPath("stalled", false)
}
