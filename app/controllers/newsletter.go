package controllers

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"errors"
	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem-email/app/jobs"
	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem-email/app/newsletter"
	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/web"
	"html/template"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const AdminEmail = "aleexgvieira@gmail.com"

func AllowedGoogle(email string, verified bool) bool {
	return verified && strings.EqualFold(email, AdminEmail)
}
func Protected(next web.HandlerFunc) web.HandlerFunc {
	return func(c *web.Ctx) error {
		if c.Session() == nil || c.Session().Get("admin") != AdminEmail {
			if strings.HasPrefix(c.Request().URL.Path, "/api/") {
				return errs.ErrUnauthorized
			}
			return c.Redirect("/login")
		}
		return next(c)
	}
}

type NewsletterController struct {
	App            *vukem.App
	DB             *database.DB
	AllowedOrigins []string
}

// originAllowed is the CSRF substitute for the public subscribe endpoint: a
// browser always sends Origin on cross-origin POSTs, so anything that is
// neither same-origin nor allowlisted is rejected.
func (n *NewsletterController) originAllowed(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return r.Header.Get("Sec-Fetch-Site") != "cross-site"
	}
	if slices.Contains(n.AllowedOrigins, origin) {
		return true
	}
	u, err := url.Parse(origin)
	return err == nil && strings.EqualFold(u.Host, r.Host)
}

func number(s string) int64 { n, _ := strconv.ParseInt(s, 10, 64); return n }
func input(c *web.Ctx) (map[string]string, error) {
	v, err := c.Input()
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	for k, value := range v {
		if s, ok := value.(string); ok {
			out[k] = strings.TrimSpace(s)
		}
	}
	return out, nil
}
func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s && len(s) <= 254
}
func (n *NewsletterController) project(c *web.Ctx) (models.Project, error) {
	q := orm.QueryOf[models.Project](n.DB)
	if slug := c.Param("slug"); slug != "" {
		return q.Where("slug", "=", slug).First(c.Context())
	}
	host := strings.ToLower(c.Request().Host)
	if h, _, ok := strings.Cut(host, ":"); ok {
		host = h
	}
	return q.Where("domain", "=", host).First(c.Context())
}
func (n *NewsletterController) Landing(c *web.Ctx) error {
	p, err := n.project(c)
	if err != nil {
		return err
	}
	return inertia.Render(c, "Subscribe", inertia.Props{"project": p})
}
func (n *NewsletterController) Subscribe(c *web.Ctx) error {
	if !n.originAllowed(c.Request()) {
		return errs.ErrForbidden
	}
	p, err := n.project(c)
	if err != nil {
		return err
	}
	in, err := input(c)
	if err != nil {
		return err
	}
	email := strings.ToLower(in["email"])
	if !validEmail(email) {
		return errs.New(422, "Enter a valid email address.")
	}
	suppressed, err := orm.QueryOf[models.Suppression](n.DB).Where("project_id", "=", p.ID).Where("email", "=", email).Count(c.Context())
	if err != nil {
		return err
	}
	if in["website"] == "" && suppressed == 0 {
		err = n.DB.Tx(c.Context(), func(tx *database.Tx) error {
			if _, err := orm.QueryOf[models.Project](tx).Where("id", "=", p.ID).LockForUpdate().First(c.Context()); err != nil {
				return err
			}
			s, err := orm.QueryOf[models.Subscriber](tx).Where("project_id", "=", p.ID).Where("email", "=", email).First(c.Context())
			if err != nil && !errors.Is(err, orm.ErrNotFound) {
				return err
			}
			now := time.Now().Unix()
			if err == nil && (s.Status == "confirmed" || now-s.LastRequestedAt < 60) {
				return nil
			}
			// Caps keep the public endpoint from mailing a third party at will; the 202 below stays the same.
			if now-s.ConfirmationsSince >= 86400 {
				s.Confirmations, s.ConfirmationsSince = 0, now
			}
			if s.Confirmations >= maxConfirmationsPerDay {
				return nil
			}
			recent, err := orm.QueryOf[models.Subscriber](tx).Where("project_id", "=", p.ID).Where("last_requested_at", ">", now-3600).Count(c.Context())
			if err != nil {
				return err
			}
			if recent >= maxConfirmationsPerProjectHour {
				return nil
			}
			s.Confirmations++
			token := rand.Text()
			s.ProjectID = p.ID
			s.Email = email
			s.Status = "pending"
			s.TokenHash = newsletter.Hash(token)
			s.ExpiresAt = now + 86400
			s.LastRequestedAt = now
			if s.ID == 0 {
				err = orm.Create(c.Context(), tx, &s)
			} else {
				err = orm.Update(c.Context(), tx, &s)
			}
			if err != nil {
				return err
			}
			return n.App.MailLaterTx(c.Context(), tx, newsletter.Confirmation{Project: p, Email: email, URL: "https://" + p.Domain + "/confirm?token=" + token})
		})
		if err != nil {
			return err
		}
	}
	return c.JSON(http.StatusAccepted, web.Map{"message": "If this address can be subscribed, you will receive a confirmation link. Check your spam folder too."})
}

const (
	maxConfirmationsPerDay         = 3
	maxConfirmationsPerProjectHour = 500
)

func (n *NewsletterController) ConfirmPage(c *web.Ctx) error {
	c.Writer().Header().Set("Referrer-Policy", "no-referrer")
	return inertia.Render(c, "SubscriptionAction", inertia.Props{"action": "confirm", "token": c.Query("token"), "subscriber": ""})
}
func (n *NewsletterController) Confirm(c *web.Ctx) error {
	in, err := input(c)
	if err != nil {
		return err
	}
	if len(in["token"]) < 20 || len(in["token"]) > 128 {
		return errs.New(422, "Invalid or expired link. Please subscribe again.")
	}
	count, err := orm.QueryOf[models.Subscriber](n.DB).Where("token_hash", "=", newsletter.Hash(in["token"])).Where("status", "=", "pending").Where("expires_at", ">", time.Now().Unix()).Update(c.Context(), orm.Set{"status": "confirmed", "token_hash": "", "expires_at": int64(0)})
	if err != nil {
		return err
	}
	if count == 0 {
		return errs.New(422, "Invalid, used, or expired link. Please subscribe again.")
	}
	return c.JSON(200, web.Map{"message": "Subscription confirmed. Welcome aboard!"})
}
func (n *NewsletterController) UnsubscribePage(c *web.Ctx) error {
	c.Writer().Header().Set("Referrer-Policy", "no-referrer")
	return inertia.Render(c, "SubscriptionAction", inertia.Props{"action": "unsubscribe", "token": c.Query("token"), "subscriber": c.Query("subscriber")})
}
func (n *NewsletterController) Unsubscribe(c *web.Ctx) error {
	in, err := input(c)
	if err != nil {
		return err
	}
	return n.unsubscribe(c, in)
}

// OneClickUnsubscribe implements RFC 8058: mail providers POST
// "List-Unsubscribe=One-Click" to the signed URL from the List-Unsubscribe header.
func (n *NewsletterController) OneClickUnsubscribe(c *web.Ctx) error {
	in, err := input(c)
	if err != nil {
		return err
	}
	if in["List-Unsubscribe"] != "One-Click" {
		return errs.New(400, "Missing List-Unsubscribe=One-Click body.")
	}
	return n.unsubscribe(c, in)
}
func (n *NewsletterController) unsubscribe(c *web.Ctx, in map[string]string) error {
	id := number(in["subscriber"])
	key, err := n.App.DeriveKey("newsletter.unsubscribe", 32)
	if err != nil {
		return err
	}
	got, err := hex.DecodeString(in["token"])
	if err != nil {
		return errs.ErrForbidden
	}
	keys := append([][]byte{key}, n.App.DerivePreviousKeys("newsletter.unsubscribe", 32)...)
	valid := false
	for _, candidate := range keys {
		want, _ := hex.DecodeString(newsletter.UnsubscribeToken(candidate, id))
		if hmac.Equal(got, want) {
			valid = true
		}
	}
	if id <= 0 || !valid {
		return errs.ErrForbidden
	}
	_, err = orm.QueryOf[models.Subscriber](n.DB).Where("id", "=", id).Update(c.Context(), orm.Set{"status": "unsubscribed", "token_hash": "", "expires_at": int64(0)})
	if err != nil {
		return err
	}
	return c.JSON(200, web.Map{"message": "Subscription canceled. You will not receive new campaigns."})
}
func (n *NewsletterController) CreateProject(c *web.Ctx) error {
	in, err := input(c)
	if err != nil {
		return err
	}
	if len(in["name"]) < 2 || len(in["name"]) > 100 || !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(in["slug"]) || len(in["slug"]) > 63 {
		return errs.New(422, "Enter a name and valid slug (lowercase letters, numbers, and hyphens).")
	}
	domain := strings.ToLower(in["domain"])
	if len(domain) > 253 || !regexp.MustCompile(`^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`).MatchString(domain) || !validEmail(in["from_email"]) {
		return errs.New(422, "Enter a domain without https:// and a valid sender email.")
	}
	if len(in["description"]) > 500 {
		return errs.New(422, "Keep the description within 500 characters.")
	}
	exists, err := orm.QueryOf[models.Project](n.DB).Where("slug", "=", in["slug"]).Where("id", "!=", number(in["id"])).Count(c.Context())
	if err != nil {
		return err
	}
	other, err := orm.QueryOf[models.Project](n.DB).Where("domain", "=", domain).Where("id", "!=", number(in["id"])).Count(c.Context())
	if err != nil {
		return err
	}
	if exists+other > 0 {
		return errs.New(409, "A project with this slug or domain already exists.")
	}
	p := models.Project{Name: in["name"], Slug: in["slug"], Domain: domain, FromEmail: in["from_email"], Description: in["description"]}
	if id := number(in["id"]); id > 0 {
		existing, err := orm.QueryOf[models.Project](n.DB).Find(c.Context(), id)
		if err != nil {
			return err
		}
		p.Model = existing.Model
		if err := orm.Update(c.Context(), n.DB, &p); err != nil {
			return err
		}
	} else if err := orm.Create(c.Context(), n.DB, &p); err != nil {
		return err
	}
	return c.JSON(201, p)
}
func (n *NewsletterController) Dashboard(c *web.Ctx) error {
	ctx := c.Context()
	projects, err := orm.QueryOf[models.Project](n.DB).OrderBy("name", orm.Asc).Get(ctx)
	if err != nil {
		return err
	}
	var p models.Project
	id := number(c.Query("project"))
	for _, project := range projects {
		if project.ID == id || (id == 0 && p.ID == 0) {
			p = project
		}
	}
	if id != 0 && p.ID == 0 {
		return errs.ErrNotFound
	}
	stats := map[string]int64{"pending": 0, "confirmed": 0, "unsubscribed": 0, "suppressed": 0}
	// One grouped query replaces one count per status.
	statusRows, err := n.DB.QueryContext(ctx, "SELECT status, COUNT(*) FROM subscribers WHERE project_id = "+n.DB.Dialect().Placeholder(1)+" GROUP BY status", p.ID)
	if err != nil {
		return err
	}
	for statusRows.Next() {
		var status string
		var count int64
		if err := statusRows.Scan(&status, &count); err != nil {
			statusRows.Close()
			return err
		}
		stats[status] = count
	}
	if err := statusRows.Err(); err != nil {
		statusRows.Close()
		return err
	}
	statusRows.Close()
	q := orm.QueryOf[models.Subscriber](n.DB).Where("project_id", "=", p.ID)
	if search := c.Query("q"); search != "" {
		q = q.Where("email", "LIKE", "%"+search+"%")
	}
	if filter := c.Query("status"); filter != "" {
		q = q.Where("status", "=", filter)
	}
	page := min(max(number(c.Query("page")), 1), 1000000)
	total, err := q.Count(ctx)
	if err != nil {
		return err
	}
	subscribers, err := q.OrderBy("id", orm.Desc).Limit(20).Offset(int((page - 1) * 20)).Get(ctx)
	if err != nil {
		return err
	}
	campaigns, err := orm.QueryOf[models.Campaign](n.DB).Where("project_id", "=", p.ID).OrderBy("id", orm.Desc).Limit(100).Get(ctx)
	if err != nil {
		return err
	}
	deliveryStats := map[int64]map[string]int64{}
	// One aggregate query avoids one round trip per status and campaign.
	rows, err := n.DB.QueryContext(ctx, "SELECT d.campaign_id, d.status, COUNT(*) FROM deliveries d JOIN campaigns c ON c.id = d.campaign_id WHERE c.project_id = "+n.DB.Dialect().Placeholder(1)+" GROUP BY d.campaign_id, d.status", p.ID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var campaignID, count int64
		var status string
		if err := rows.Scan(&campaignID, &status, &count); err != nil {
			rows.Close()
			return err
		}
		if deliveryStats[campaignID] == nil {
			deliveryStats[campaignID] = map[string]int64{}
		}
		deliveryStats[campaignID][status] = count
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	stats["campaigns"], err = orm.QueryOf[models.Campaign](n.DB).Where("project_id", "=", p.ID).Count(ctx)
	if err != nil {
		return err
	}

	return inertia.Render(c, "Dashboard", inertia.Props{"projects": projects, "project": p, "subscribers": subscribers, "campaigns": campaigns, "stats": stats, "deliveryStats": deliveryStats, "total": total, "page": page, "filters": web.Map{"q": c.Query("q"), "status": c.Query("status")}, "admin": AdminEmail})
}
func (n *NewsletterController) campaignInput(c *web.Ctx) (models.Project, map[string]string, string, error) {
	in, err := input(c)
	if err != nil {
		return models.Project{}, nil, "", err
	}
	p, err := orm.QueryOf[models.Project](n.DB).Find(c.Context(), number(in["project_id"]))
	if err != nil {
		return p, nil, "", err
	}
	if in["subject"] == "" || len(in["subject"]) > 200 || strings.ContainsAny(in["subject"], "\r\n") {
		return p, nil, "", errs.New(422, "Enter a subject of up to 200 characters.")
	}
	if in["source"] == "" || (in["mode"] != "markdown" && in["mode"] != "mdx" && in["mode"] != "components") {
		return p, nil, "", errs.New(422, "Choose a format and write some content.")
	}
	html, err := newsletter.Render(c.Context(), in["mode"], in["source"])
	if err != nil {
		return p, nil, "", errs.New(422, err.Error())
	}
	return p, in, html, nil
}
func (n *NewsletterController) SaveCampaign(c *web.Ctx) error {
	p, in, html, err := n.campaignInput(c)
	if err != nil {
		return err
	}
	campaign := models.Campaign{ProjectID: p.ID, Subject: in["subject"], Mode: in["mode"], Source: in["source"], BodyHTML: html, Status: "draft"}
	if id := number(in["id"]); id > 0 {
		count, err := orm.QueryOf[models.Campaign](n.DB).Where("id", "=", id).Where("project_id", "=", p.ID).Where("status", "=", "draft").Update(c.Context(), orm.Set{"subject": campaign.Subject, "mode": campaign.Mode, "source": campaign.Source, "body_html": html})
		if err != nil {
			return err
		}
		if count == 0 {
			return errs.New(409, "The campaign was already sent or was not found.")
		}
		campaign.ID = id
	} else if err := orm.Create(c.Context(), n.DB, &campaign); err != nil {
		return err
	}
	return c.JSON(200, campaign)
}
func (n *NewsletterController) Preview(c *web.Ctx) error {
	p, in, body, err := n.campaignInput(c)
	if err != nil {
		return err
	}
	mailer, err := n.App.Mailer()
	if err != nil {
		return err
	}
	message, err := mailer.Render(c.Context(), newsletter.Newsletter{Project: p, Email: AdminEmail, Subject: in["subject"], Body: template.HTML(body), UnsubscribeURL: "https://" + p.Domain + "/subscribe"})
	if err != nil {
		return err
	}
	store, err := n.App.Cache()
	if err != nil {
		return err
	}
	token := rand.Text()
	if err := store.Put(c.Context(), "email-preview:"+token, message.HTML, 10*time.Minute); err != nil {
		return err
	}
	return c.JSON(200, web.Map{"url": "/api/preview/" + token})
}
func (n *NewsletterController) SendCampaign(c *web.Ctx) error {
	q, err := n.App.Queue()
	if err != nil {
		return err
	}
	err = n.DB.Tx(c.Context(), func(tx *database.Tx) error {
		campaign, err := orm.QueryOf[models.Campaign](tx).Where("id", "=", number(c.Param("id"))).LockForUpdate().First(c.Context())
		if err != nil {
			return err
		}
		if campaign.Status != "draft" {
			return errs.New(409, "This campaign is already queued for delivery.")
		}
		count, err := orm.QueryOf[models.Campaign](tx).Where("id", "=", campaign.ID).Where("status", "=", "draft").Update(c.Context(), orm.Set{"status": "queued"})
		if err != nil {
			return err
		}
		if count == 0 {
			return errs.New(409, "This campaign is already queued.")
		}
		var recipients, lastID int64
		for {
			subs, err := orm.QueryOf[models.Subscriber](tx).Where("project_id", "=", campaign.ProjectID).Where("status", "=", "confirmed").Where("id", ">", lastID).OrderBy("id", orm.Asc).Limit(500).Get(c.Context())
			if err != nil {
				return err
			}
			if len(subs) == 0 {
				break
			}
			for _, s := range subs {
				d := models.Delivery{CampaignID: campaign.ID, SubscriberID: s.ID, Status: "queued"}
				if err := orm.Create(c.Context(), tx, &d); err != nil {
					return err
				}
				if err := q.DispatchTx(c.Context(), tx, jobs.SendDelivery{DeliveryID: d.ID}); err != nil {
					return err
				}
				lastID = s.ID
				recipients++
			}
		}
		if recipients == 0 {
			return errs.New(422, "This project has no confirmed subscribers yet.")
		}
		_, err = orm.QueryOf[models.Campaign](tx).Where("id", "=", campaign.ID).Update(c.Context(), orm.Set{"recipients": recipients})
		return err
	})
	if err != nil {
		return err
	}
	return c.JSON(202, web.Map{"message": "Campaign queued. Delivery respects confirmation and unsubscribe status."})
}

// stalledAfter is how long a delivery may wait in "queued" before the campaign
// is reported as stalled: the worker normally drains a campaign in minutes.
const stalledAfter = 15 * time.Minute

// CampaignHealth reports how a campaign's deliveries are progressing, so a
// stuck queue shows without opening the database.
func (n *NewsletterController) CampaignHealth(c *web.Ctx) error {
	ctx := c.Context()
	campaign, err := orm.QueryOf[models.Campaign](n.DB).Find(ctx, number(c.Param("id")))
	if err != nil {
		return err
	}
	deliveries, err := orm.QueryOf[models.Delivery](n.DB).Where("campaign_id", "=", campaign.ID).GroupCount[string](ctx, "status")
	if err != nil {
		return err
	}
	var oldest int64
	if d, err := orm.QueryOf[models.Delivery](n.DB).Where("campaign_id", "=", campaign.ID).Where("status", "=", "queued").OrderBy("created_at", orm.Asc).First(ctx); err == nil {
		oldest = int64(time.Since(d.CreatedAt).Seconds())
	} else if !errors.Is(err, orm.ErrNotFound) {
		return err
	}
	return c.JSON(200, web.Map{"status": campaign.Status, "recipients": campaign.Recipients, "deliveries": deliveries, "oldest_queued_seconds": oldest, "stalled": oldest > int64(stalledAfter.Seconds())})
}
func (n *NewsletterController) Export(c *web.Ctx) error {
	id := number(c.Query("project"))
	if _, err := orm.QueryOf[models.Project](n.DB).Find(c.Context(), id); err != nil {
		return err
	}
	c.Writer().Header().Set("Content-Disposition", `attachment; filename="subscribers.csv"`)
	c.Writer().Header().Set("Content-Type", "text/csv; charset=utf-8")
	w := csv.NewWriter(c.Writer())
	if err := w.Write([]string{"email", "status", "created_at"}); err != nil {
		return err
	}
	var last int64
	for {
		rows, err := orm.QueryOf[models.Subscriber](n.DB).Where("project_id", "=", id).Where("id", ">", last).OrderBy("id", orm.Asc).Limit(500).Get(c.Context())
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		for _, s := range rows {
			email := s.Email
			if email != "" && strings.ContainsAny(email[:1], "=+-@\t\r") {
				email = "'" + email
			}
			if err := w.Write([]string{email, s.Status, s.CreatedAt.Format(time.RFC3339)}); err != nil {
				return err
			}
			last = s.ID
		}
	}
	w.Flush()
	return w.Error()
}

func (n *NewsletterController) AdminUnsubscribe(c *web.Ctx) error {
	count, err := orm.QueryOf[models.Subscriber](n.DB).Where("id", "=", number(c.Param("id"))).Update(c.Context(), orm.Set{"status": "unsubscribed", "token_hash": "", "expires_at": int64(0)})
	if err != nil {
		return err
	}
	if count == 0 {
		return errs.ErrNotFound
	}
	return c.JSON(200, web.Map{"message": "Subscription canceled."})
}

// PreviewDocument is the only document allowed email-specific styles and remote
// images. It requires the admin session and a short-lived random cache key.
// Both its CSP sandbox and the iframe sandbox disable scripts, forms and navigation.
func (n *NewsletterController) PreviewDocument(c *web.Ctx) error {
	store, err := n.App.Cache()
	if err != nil {
		return err
	}
	body, ok, err := cache.Get[string](c.Context(), store, "email-preview:"+c.Param("token"))
	if err != nil {
		return err
	}
	if !ok {
		return errs.ErrNotFound
	}
	c.SetHeader("Content-Security-Policy", "default-src 'none'; script-src 'none'; style-src 'unsafe-inline'; img-src https: data:; base-uri 'none'; form-action 'none'; frame-ancestors 'self'; sandbox")
	c.SetHeader("X-Frame-Options", "SAMEORIGIN")
	c.SetHeader("Cache-Control", "no-store")
	c.SetHeader("Referrer-Policy", "no-referrer")
	return c.HTML(200, body)
}
