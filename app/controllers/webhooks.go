package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/web"
	"github.com/AlexGalhardo/vukem/webhook"
)

const (
	// SoftBounceLimit is how many soft bounces (Resend "Transient" or
	// "Undetermined") within SoftBounceWindow suppress an address for a project.
	SoftBounceLimit  = 3
	SoftBounceWindow = 30 * 24 * time.Hour
	maxBounceReason  = 1000
)

type resendEvent struct {
	Type string `json:"type"`
	Data struct {
		To    []string          `json:"to"`
		Tags  map[string]string `json:"tags"`
		Click struct {
			Link string `json:"link"`
		} `json:"click"`
		Bounce struct {
			Type    string `json:"type"`
			Message string `json:"message"`
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
// deduplication are enforced by the route middleware. Every bounce and
// complaint is logged in the bounces table. A hard bounce or a complaint
// suppresses the address for the project that sent the message; so do
// SoftBounceLimit soft bounces within SoftBounceWindow.
func (n *NewsletterController) ResendWebhook(c *web.Ctx) error {
	var ev resendEvent
	if err := json.NewDecoder(io.LimitReader(c.Request().Body, 256<<10)).Decode(&ev); err != nil {
		return errs.New(400, "Invalid webhook payload.")
	}
	if ev.Type == "email.opened" || ev.Type == "email.clicked" {
		return n.recordEngagement(c, ev)
	}
	status := map[string]string{"email.delivered": "delivered", "email.bounced": "bounced", "email.complained": "complained"}[ev.Type]
	if status == "" || len(ev.Data.To) == 0 {
		return c.JSON(200, web.Map{"ignored": true})
	}
	ctx := c.Context()
	deliveryID, _ := strconv.ParseInt(ev.Data.Tags["delivery"], 10, 64)
	if deliveryID > 0 {
		if _, err := orm.QueryOf[models.Delivery](n.DB).Where("id", "=", deliveryID).WhereIn("status", weakerThan(status)...).Update(ctx, orm.Set{"status": status}); err != nil {
			return err
		}
	}
	projectID, err := strconv.ParseInt(ev.Data.Tags["project"], 10, 64)
	if ev.Type == "email.delivered" || err != nil || projectID <= 0 {
		return c.JSON(200, web.Map{"ok": true})
	}
	email := strings.ToLower(strings.TrimSpace(ev.Data.To[0]))
	kind := "complaint"
	if ev.Type == "email.bounced" {
		kind = "soft"
		if strings.EqualFold(ev.Data.Bounce.Type, "Permanent") {
			kind = "hard"
		}
	}
	err = n.DB.Tx(ctx, func(tx *database.Tx) error {
		bounce := models.Bounce{ProjectID: projectID, Email: email, DeliveryID: deliveryID, Type: kind, Reason: truncate(ev.Data.Bounce.Message, maxBounceReason), EventID: webhook.StandardID(c.Request())}
		if sub, err := orm.QueryOf[models.Subscriber](tx).Where("project_id", "=", projectID).Where("email", "=", email).First(ctx); err == nil {
			bounce.SubscriberID = sub.ID
		}
		if err := orm.Upsert(ctx, tx, &bounce, orm.OnConflict{Columns: []string{"event_id"}, DoNothing: true}); err != nil {
			return err
		}
		reason := map[string]string{"hard": "hard_bounce", "complaint": "complaint"}[kind]
		if kind == "soft" {
			soft, err := orm.QueryOf[models.Bounce](tx).Where("project_id", "=", projectID).Where("email", "=", email).Where("type", "=", "soft").Where("created_at", ">", time.Now().Add(-SoftBounceWindow).UTC()).Count(ctx)
			if err != nil {
				return err
			}
			if soft < SoftBounceLimit {
				return nil
			}
			reason = "soft_bounce_limit"
		}
		// Svix retries and concurrent events for one address race on the unique (project_id, email) index; the first row wins.
		sup := models.Suppression{ProjectID: projectID, Email: email, Reason: reason}
		if err := orm.Upsert(ctx, tx, &sup, orm.OnConflict{Columns: []string{"project_id", "email"}, DoNothing: true}); err != nil {
			return err
		}
		_, err := orm.QueryOf[models.Subscriber](tx).Where("project_id", "=", projectID).Where("email", "=", email).Update(ctx, orm.Set{"status": "suppressed", "token_hash": "", "expires_at": int64(0)})
		return err
	})
	if err != nil {
		return err
	}
	return c.JSON(200, web.Map{"ok": true})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return strings.ToValidUTF8(s[:n], "")
}

// recordEngagement stores an open or click for its delivery (the "delivery"
// tag). Events for unknown deliveries are ignored. The click URL loses its
// query string and fragment: links in campaigns can carry personal tokens.
func (n *NewsletterController) recordEngagement(c *web.Ctx, ev resendEvent) error {
	ctx := c.Context()
	deliveryID, _ := strconv.ParseInt(ev.Data.Tags["delivery"], 10, 64)
	d, err := orm.QueryOf[models.Delivery](n.DB).Find(ctx, deliveryID)
	if errors.Is(err, orm.ErrNotFound) {
		return c.JSON(200, web.Map{"ignored": true})
	}
	if err != nil {
		return err
	}
	e := models.DeliveryEvent{DeliveryID: d.ID, CampaignID: d.CampaignID, Type: strings.TrimPrefix(ev.Type, "email."), EventID: webhook.StandardID(c.Request())}
	e.Type = map[string]string{"opened": "opened", "clicked": "clicked"}[e.Type]
	if e.Type == "clicked" {
		e.URL = stripQuery(ev.Data.Click.Link)
	}
	if err := orm.Upsert(ctx, n.DB, &e, orm.OnConflict{Columns: []string{"event_id"}, DoNothing: true}); err != nil {
		return err
	}
	return c.JSON(200, web.Map{"ok": true})
}

func stripQuery(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	u.RawQuery, u.Fragment, u.User = "", "", nil
	return truncate(u.String(), 2048)
}

// trackingStats counts opens and clicks (total and per distinct delivery) for the given campaigns.
func (n *NewsletterController) trackingStats(ctx context.Context, projectID int64) (map[int64]map[string]int64, error) {
	rows, err := n.DB.QueryContext(ctx, "SELECT e.campaign_id, e.type, COUNT(*), COUNT(DISTINCT e.delivery_id) FROM delivery_events e JOIN campaigns c ON c.id = e.campaign_id WHERE c.project_id = "+n.DB.Dialect().Placeholder(1)+" GROUP BY e.campaign_id, e.type", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]map[string]int64{}
	for rows.Next() {
		var campaignID, total, unique int64
		var kind string
		if err := rows.Scan(&campaignID, &kind, &total, &unique); err != nil {
			return nil, err
		}
		if out[campaignID] == nil {
			out[campaignID] = map[string]int64{}
		}
		prefix := map[string]string{"opened": "opens", "clicked": "clicks"}[kind]
		out[campaignID][prefix], out[campaignID]["unique_"+prefix] = total, unique
	}
	return out, rows.Err()
}
