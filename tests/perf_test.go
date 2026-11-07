package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/orm"
)

// Set RUN_PERF_TESTS=1 to print dashboard timings against a large audience.
func TestDashboardPerformance(t *testing.T) {
	if os.Getenv("RUN_PERF_TESTS") != "1" {
		t.Skip("set RUN_PERF_TESTS=1 to measure the dashboard on 30k subscribers")
	}
	app, p := setup(t)
	db, _ := app.DB()
	const n = 30000
	subs := make([]models.Subscriber, n)
	for i := range subs {
		status := []string{"confirmed", "pending", "unsubscribed"}[i%3]
		subs[i] = models.Subscriber{ProjectID: p.ID, Email: fmt.Sprintf("user%d@example.com", i), Status: status}
	}
	for i := 0; i < n; i += 1000 {
		if err := orm.CreateMany(t.Context(), db, subs[i:i+1000]); err != nil {
			t.Fatal(err)
		}
	}
	c := models.Campaign{ProjectID: p.ID, Subject: "S", Mode: "markdown", Source: "x", BodyHTML: "x", Status: "completed"}
	if err := orm.Create(t.Context(), db, &c); err != nil {
		t.Fatal(err)
	}
	ds := make([]models.Delivery, 0, 10000)
	for i := 1; i <= 10000; i++ {
		ds = append(ds, models.Delivery{CampaignID: c.ID, SubscriberID: int64(i), Status: "accepted"})
	}
	for i := 0; i < len(ds); i += 1000 {
		if err := orm.CreateMany(t.Context(), db, ds[i:i+1000]); err != nil {
			t.Fatal(err)
		}
	}
	client := admin(t, app)
	client.Get("/").AssertStatus(200)
	start := time.Now()
	const runs = 20
	for range runs {
		client.Get("/?status=confirmed&page=3").AssertStatus(200)
	}
	t.Logf("dashboard: %v per request (%d subscribers)", time.Since(start)/runs, n)
}

func TestStaticAssetsAreImmutableAndPagesAreNotShared(t *testing.T) {
	app, _ := setup(t)
	files, _ := filepath.Glob("resources/public/build/assets/*.js")
	if len(files) == 0 {
		t.Skip("run npm run build first")
	}
	app.Client(t).Get("/build/assets/"+filepath.Base(files[0])).AssertStatus(200).AssertHeaderContains("Cache-Control", "immutable")
	// HTML embeds a per-response CSP nonce, so it must never be served from a shared cache.
	r := app.Client(t).Get("/subscribe/hexascore").AssertStatus(200)
	if strings.Contains(r.Header("Cache-Control"), "public") {
		t.Fatal("pages with CSP nonces must not be publicly cacheable")
	}
}
