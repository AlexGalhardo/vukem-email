package newsletter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"time"
)

// Render compiles author content in a bounded subprocess. The renderer evaluates authored MDX in an isolated WebAssembly interpreter
// without host APIs, then validates and sanitizes the resulting element data.
func Render(ctx context.Context, mode, source string) (string, error) {
	if len(source) > 100000 {
		return "", errors.New("O conteúdo excede 100 KB.")
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	payload, _ := json.Marshal(map[string]string{"mode": mode, "source": source})
	cmd := exec.CommandContext(ctx, "node", "--max-old-space-size=128", "--import", "tsx", "scripts/render-email.tsx")
	cmd.Stdin = bytes.NewReader(payload)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", errors.New("Conteúdo inválido. Revise a sintaxe e use os componentes React Email disponíveis e imports de React ou @react-email/components.")
	}
	if out.Len() > 1000000 {
		return "", errors.New("O HTML gerado excede o limite de 1 MB.")
	}
	return out.String(), nil
}
