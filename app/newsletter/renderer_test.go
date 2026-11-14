package newsletter

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
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

func TestRenderSeparatesInfrastructureFailuresFromInvalidContent(t *testing.T) {
	t.Chdir("../..")
	var contentErr *ContentError

	if _, err := Render(t.Context(), "markdown", strings.Repeat("x", 100001)); !errors.As(err, &contentErr) {
		t.Fatalf("oversized source must be a content error: %v", err)
	}
	if _, err := Render(t.Context(), "mdx", "export default () => {"); !errors.As(err, &contentErr) {
		t.Fatalf("a syntax error must be a content error: %v", err)
	}

	old := nodeBinary
	t.Cleanup(func() { nodeBinary = old })
	nodeBinary = "node-does-not-exist"
	if _, err := Render(t.Context(), "markdown", "hi"); err == nil || errors.As(err, &contentErr) {
		t.Fatalf("missing node must be an infrastructure error: %v", err)
	}
}

func TestRenderTimeoutIsNotAContentError(t *testing.T) {
	t.Chdir("../..")
	old := renderTimeout
	t.Cleanup(func() { renderTimeout = old })
	renderTimeout = time.Millisecond
	_, err := Render(t.Context(), "markdown", "hi")
	var contentErr *ContentError
	if err == nil || errors.As(err, &contentErr) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout must be an infrastructure error wrapping DeadlineExceeded: %v", err)
	}
}
