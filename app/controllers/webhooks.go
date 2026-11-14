package controllers

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/web"
)

type resendEvent struct {
	Type string `json:"type"`
	Data struct {
		To     []string          `json:"to"`
		Tags   map[string]string `json:"tags"`
		Bounce struct {
			Type string `json:"type"`
		} `json:"bounce"`
	} `json:"data"`
}

// deliveryRank orders the delivery states Resend events can set. An event
// never moves a delivery to a weaker state, so out-of-order or repeated
// events cannot lose information (a late "delivered" after "bounced").
var deliveryRank = map[string]int{"queued": 0, "accepted": 1, "delivered": 2, "bounced": 3, "complained": 4}

func weakerThan(status string) []any {
	var out []any
	for s, rank := range deliveryRank {
		if rank < deliveryRank[status] {
			out = append(out, s)
		}
	}
	return out
}

// ResendWebhook records delivery outcomes. Signature, replay window and
// deduplication are enforced by the route middleware. Hard bounces and
// complaints suppress the address for the project that sent the message.
func (n *NewsletterController) ResendWebhook(c *web.Ctx) error {
	var ev resendEvent
	if err := json.NewDecoder(io.LimitReader(c.Request().Body, 256<<10)).Decode(&ev); err != nil {
		return errs.New(400, "Invalid webhook payload.")
	}
	status := map[string]string{"email.delivered": "delivered", "email.bounced": "bounced", "email.complained": "complained"}[ev.Type]
	if status == "" || len(ev.Data.To) == 0 {
		return c.JSON(200, web.Map{"ignored": true})
	}
	ctx := c.Context()
	if id, err := strconv.ParseInt(ev.Data.Tags["delivery"], 10, 64); err == nil {
		if _, err := orm.QueryOf[models.Delivery](n.DB).Where("id", "=", id).WhereIn("status", weakerThan(status)...).Update(ctx, orm.Set{"status": status}); err != nil {
			return err
		}
	}
	hard := (ev.Type == "email.bounced" && strings.EqualFold(ev.Data.Bounce.Type, "Permanent")) || ev.Type == "email.complained"
	projectID, err := strconv.ParseInt(ev.Data.Tags["project"], 10, 64)
	if !hard || err != nil || projectID <= 0 {
		return c.JSON(200, web.Map{"ok": true})
	}
	email := strings.ToLower(strings.TrimSpace(ev.Data.To[0]))
	reason := map[string]string{"email.bounced": "hard_bounce", "email.complained": "complaint"}[ev.Type]
	// Svix retries and concurrent events for one address race on the unique (project_id, email) index; the first row wins.
	sup := models.Suppression{ProjectID: projectID, Email: email, Reason: reason}
	if err := orm.Upsert(ctx, n.DB, &sup, orm.OnConflict{Columns: []string{"project_id", "email"}, DoNothing: true}); err != nil {
		return err
	}
	if _, err := orm.QueryOf[models.Subscriber](n.DB).Where("project_id", "=", projectID).Where("email", "=", email).Update(ctx, orm.Set{"status": "suppressed", "token_hash": "", "expires_at": int64(0)}); err != nil {
		return err
	}
	return c.JSON(200, web.Map{"ok": true})
}
