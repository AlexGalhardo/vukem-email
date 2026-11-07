package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/orm"
)

// SA-001: past the cooldown, an unconfirmed address still gets at most three confirmations a day.
func TestConfirmationsCappedPerAddress(t *testing.T) {
	app, _ := setup(t)
	db, _ := app.DB()
	client := app.Client(t)
	for range 5 {
		client.PostJSON("/subscribe/hexascore", map[string]string{"email": "victim@example.com"}).AssertStatus(202)
		if _, err := orm.QueryOf[models.Subscriber](db).Where("email", "=", "victim@example.com").Update(t.Context(), orm.Set{"last_requested_at": int64(0)}); err != nil {
			t.Fatal(err)
		}
	}
	if got := len(app.Mails().SentTo("victim@example.com")); got != 3 {
		t.Fatalf("confirmations sent: %d, want 3", got)
	}
}
