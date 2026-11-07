package jobs

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem-email/app/newsletter"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/queue"
)

type SendDelivery struct{ DeliveryID int64 }

func (SendDelivery) Kind() string { return "newsletter.delivery" }

func deliver(app *vukem.App, ctx context.Context, job SendDelivery) (result error) {
	db, err := app.DB()
	if err != nil {
		return err
	}
	d, err := orm.QueryOf[models.Delivery](db).Find(ctx, job.DeliveryID)
	if err != nil {
		return err
	}
	defer func() {
		if result != nil {
			if attempt, ok := queue.AttemptFrom(ctx); ok && attempt.Last() {
				if _, updateErr := orm.QueryOf[models.Delivery](db).Where("id", "=", d.ID).Where("status", "=", "queued").Update(ctx, orm.Set{"status": "failed"}); updateErr != nil {
					result = fmt.Errorf("settle delivery: %w", updateErr)
					return
				}
			}
		}
		remaining, countErr := orm.QueryOf[models.Delivery](db).Where("campaign_id", "=", d.CampaignID).Where("status", "=", "queued").Count(ctx)
		if countErr != nil {
			if result == nil {
				result = countErr
			}
			return
		}
		if remaining == 0 {
			if _, updateErr := orm.QueryOf[models.Campaign](db).Where("id", "=", d.CampaignID).Update(ctx, orm.Set{"status": "completed"}); result == nil {
				result = updateErr
			}
		}
	}()
	if d.Status == "accepted" || d.Status == "skipped" {
		return nil
	}
	if time.Since(d.CreatedAt) > 23*time.Hour {
		_, err = orm.QueryOf[models.Delivery](db).Where("id", "=", d.ID).Update(ctx, orm.Set{"status": "failed"})
		if err != nil {
			return err
		}
		return fmt.Errorf("delivery exceeded idempotency window: %w", queue.ErrPermanent)
	}
	s, err := orm.QueryOf[models.Subscriber](db).Find(ctx, d.SubscriberID)
	if err != nil {
		return err
	}
	c, err := orm.QueryOf[models.Campaign](db).Find(ctx, d.CampaignID)
	if err != nil {
		return err
	}
	if s.Status != "confirmed" || s.ProjectID != c.ProjectID {
		_, err = orm.QueryOf[models.Delivery](db).Where("id", "=", d.ID).Update(ctx, orm.Set{"status": "skipped"})
		return err
	}
	p, err := orm.QueryOf[models.Project](db).Find(ctx, c.ProjectID)
	if err != nil {
		return err
	}
	key, err := app.DeriveKey("newsletter.unsubscribe", 32)
	if err != nil {
		return err
	}
	url := "https://" + p.Domain + "/unsubscribe?subscriber=" + strconv.FormatInt(s.ID, 10) + "&token=" + newsletter.UnsubscribeToken(key, s.ID)
	oneClick := newsletter.OneClickURL(p.Domain, s.ID, newsletter.UnsubscribeToken(key, s.ID))
	mailer, err := app.Mailer()
	if err != nil {
		return err
	}
	msg, err := mailer.Render(ctx, newsletter.Newsletter{Project: p, Email: s.Email, Subject: c.Subject, Body: template.HTML(c.BodyHTML), UnsubscribeURL: url})
	if err != nil {
		return err
	}
	msg.ID = fmt.Sprintf("newsletter-%d-%d", c.ID, s.ID)
	if err = newsletter.SendWithHeaders(ctx, app, msg, newsletter.ListUnsubscribeHeaders(oneClick)); err != nil {
		if attempt, ok := queue.AttemptFrom(ctx); ok && attempt.Last() {
			if _, updateErr := orm.QueryOf[models.Delivery](db).Where("id", "=", d.ID).Update(ctx, orm.Set{"status": "failed"}); updateErr != nil {
				return updateErr
			}
		}
		return err
	}
	_, err = orm.QueryOf[models.Delivery](db).Where("id", "=", d.ID).Update(ctx, orm.Set{"status": "accepted"})
	return err
}
