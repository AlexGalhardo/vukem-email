package newsletter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/mail"
	"github.com/AlexGalhardo/vukem/queue"
)

// ListUnsubscribeHeaders returns the RFC 8058 one-click headers. Mail clients
// POST "List-Unsubscribe=One-Click" to the HTTPS URL without a session.
func ListUnsubscribeHeaders(oneClickURL string) map[string]string {
	return map[string]string{
		"List-Unsubscribe":      "<" + oneClickURL + ">",
		"List-Unsubscribe-Post": "List-Unsubscribe=One-Click",
	}
}

var resendClient = &http.Client{Timeout: 15 * time.Second}

// SendWithHeaders sends msg with extra MIME headers. vukem's mail.Message has no
// header field (see VE-2 in vukem-feedback.md), so with MAIL_MAILER=resend this
// posts to the Resend API directly; other mailers (array, log, smtp) send
// without the headers. Remove when vukem supports headers.
func SendWithHeaders(ctx context.Context, app *vukem.App, msg *mail.Message, headers map[string]string) error {
	if app.Source.Get("MAIL_MAILER") != "resend" {
		mailer, err := app.Mailer()
		if err != nil {
			return err
		}
		return mailer.Transport().Send(ctx, msg)
	}
	base := strings.TrimRight(app.Source.Get("RESEND_BASE_URL"), "/")
	if base == "" {
		base = "https://api.resend.com"
	}
	payload := map[string]any{"from": msg.From, "to": msg.To, "subject": msg.Subject, "html": msg.HTML, "text": msg.Text, "headers": headers}
	if len(msg.ReplyTo) > 0 {
		payload["reply_to"] = msg.ReplyTo
	}
	if len(msg.Tags) > 0 {
		tags := make([]map[string]string, 0, len(msg.Tags))
		for k, v := range msg.Tags {
			tags = append(tags, map[string]string{"name": k, "value": v})
		}
		payload["tags"] = tags
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+app.Source.Get("RESEND_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", msg.ID)
	resp, err := resendClient.Do(req)
	if err != nil {
		return fmt.Errorf("resend request: %w", err)
	}
	defer resp.Body.Close()
	detail, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	switch {
	case resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
		return fmt.Errorf("resend %d: %s", resp.StatusCode, detail)
	default:
		return fmt.Errorf("resend %d: %s: %w", resp.StatusCode, detail, errors.Join(queue.ErrPermanent))
	}
}

func OneClickURL(domain string, id int64, token string) string {
	return "https://" + domain + "/unsubscribe/one-click?subscriber=" + strconv.FormatInt(id, 10) + "&token=" + token
}
