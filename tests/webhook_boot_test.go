package tests

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem-email/bootstrap"
	"github.com/AlexGalhardo/vukem/vukemtest"
)

func TestProductionBootRequiresWebhookSecret(t *testing.T) {
	t.Chdir("..")
	_, err := vukemtest.Build(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"APP_ENV": "production"}})
	if err == nil || !strings.Contains(err.Error(), "RESEND_WEBHOOK_SECRET") {
		t.Fatalf("production boot without the webhook secret must fail naming it: %v", err)
	}
	secret := "whsec_" + base64.StdEncoding.EncodeToString(webhookKey)
	if _, err := vukemtest.Build(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"APP_ENV": "production", "RESEND_WEBHOOK_SECRET": secret}}); err != nil && strings.Contains(err.Error(), "RESEND_WEBHOOK_SECRET") {
		t.Fatalf("a configured secret must not trip the check: %v", err)
	}
}
