package newsletter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"
)

// ContentError means the author's content was rejected: a syntax error, an
// unsupported import or a size limit. Every other Render error is an
// infrastructure failure (timeout, missing node, OOM kill) and must not be
// reported to the author as invalid content.
type ContentError struct{ Message string }

func (e *ContentError) Error() string { return e.Message }

const (
	// invalidContentExit is the exit code scripts/render-email.tsx uses for rejected content;
	// any other non-zero exit (uncaught error, OOM abort, signal) is an infrastructure failure.
	invalidContentExit = 65
	maxStderrLogged    = 2048
)

// Overridable by tests.
var (
	nodeBinary    = "node"
	renderTimeout = 15 * time.Second
)

// Render compiles author content in a bounded subprocess. The renderer evaluates authored MDX in an isolated WebAssembly interpreter
// without host APIs, then validates and sanitizes the resulting element data.
func Render(ctx context.Context, mode, source string) (string, error) {
	if len(source) > 100000 {
		return "", &ContentError{"Content exceeds 100 KB."}
	}
	ctx, cancel := context.WithTimeout(ctx, renderTimeout)
	defer cancel()
	payload, _ := json.Marshal(map[string]string{"mode": mode, "source": source})
	cmd := exec.CommandContext(ctx, nodeBinary, "--max-old-space-size=128", "--import", "tsx", "scripts/render-email.tsx")
	cmd.Env = renderEnv()
	cmd.Stdin = bytes.NewReader(payload)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() == invalidContentExit && ctx.Err() == nil {
			return "", &ContentError{"Invalid content. Review the syntax and use supported React Email components and imports from React or @react-email/components."}
		}
		cause := err
		if ctx.Err() != nil {
			cause = ctx.Err()
		}
		slog.ErrorContext(ctx, "email renderer failed", "error", err, "stderr", truncate(stderr.String(), maxStderrLogged))
		return "", fmt.Errorf("render email: %w", cause)
	}
	if out.Len() > 1000000 {
		return "", &ContentError{"Generated HTML exceeds the 1 MB limit."}
	}
	return out.String(), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}

// renderEnv is the allowlisted environment of the renderer subprocess. It never
// carries APP_KEY, provider keys or DATABASE_URL, so a bug on the Node side
// cannot leak them.
func renderEnv() []string {
	var env []string
	for _, name := range []string{"PATH", "SystemRoot", "TEMP", "TMP", "TMPDIR", "HOME", "USERPROFILE", "NODE_ENV", "LANG"} {
		if v, ok := os.LookupEnv(name); ok {
			env = append(env, name+"="+v)
		}
	}
	return env
}
