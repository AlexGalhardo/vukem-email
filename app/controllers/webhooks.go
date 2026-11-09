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
		if _, err := orm.QueryOf[models.Delivery](n.DB).Where("id", "=", id).WhereIn("status", "queued", "accepted", "delivered").Update(ctx, orm.Set{"status": status}); err != nil {
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
	existing, err := orm.QueryOf[models.Suppression](n.DB).Where("project_id", "=", projectID).Where("email", "=", email).Count(ctx)
	if err != nil {
		return err
	}
	if existing == 0 {
		sup := models.Suppression{ProjectID: projectID, Email: email, Reason: reason}
		if err := orm.Create(ctx, n.DB, &sup); err != nil {
			return err
		}
	}
	if _, err := orm.QueryOf[models.Subscriber](n.DB).Where("project_id", "=", projectID).Where("email", "=", email).Update(ctx, orm.Set{"status": "suppressed", "token_hash": "", "expires_at": int64(0)}); err != nil {
		return err
	}
	return c.JSON(200, web.Map{"ok": true})
}
