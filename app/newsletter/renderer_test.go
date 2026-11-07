package newsletter

import (
	"strings"
	"testing"
)

func TestRenderEnvOmitsSecrets(t *testing.T) {
	for _, name := range []string{"APP_KEY", "RESEND_API_KEY", "DATABASE_URL", "SOCIAL_GOOGLE_CLIENT_SECRET", "RESEND_WEBHOOK_SECRET"} {
		t.Setenv(name, "secret-value")
	}
	t.Setenv("PATH", "/bin")
	env := strings.Join(renderEnv(), "\n")
	if strings.Contains(env, "secret-value") {
		t.Fatalf("renderer environment leaks a secret: %s", env)
	}
	if !strings.Contains(env, "PATH=/bin") {
		t.Fatal("renderer environment must keep PATH")
	}
}
