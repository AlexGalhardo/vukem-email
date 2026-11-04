package routes

import (
	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem-email/app/controllers"
	"github.com/AlexGalhardo/vukem-email/resources"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/social"
	"github.com/AlexGalhardo/vukem/web"
	"strings"
	"time"
)

func Web(r *router.Router, app *vukem.App) error {
	if err := vukem.DefaultMiddleware(app, r); err != nil {
		return err
	}
	r.UseServer(middleware.Gzip())
	r.UseServer(middleware.StaticFS(resources.Public(), middleware.StaticConfig{Prefix: "/"}))
	r.Use(app.Inertia.Middleware)
	if err := app.RequireMailer(); err != nil {
		return err
	}
	db, err := app.DB()
	if err != nil {
		return err
	}
	c := &controllers.NewsletterController{App: app, DB: db}
	key, err := app.DeriveKey("social.state", 32)
	if err != nil {
		return err
	}
	google := social.Google(social.Config{ClientID: app.Source.Get("SOCIAL_GOOGLE_CLIENT_ID"), ClientSecret: app.Source.Get("SOCIAL_GOOGLE_CLIENT_SECRET"), RedirectURL: strings.TrimRight(app.Source.Get("APP_URL"), "/") + "/auth/google/callback"})
	s := social.New(key, google).WithPreviousKeys(app.DerivePreviousKeys("social.state", 32)...)
	s.InsecureCookies = app.Config.Env != "production"
	cache, err := app.Cache()
	if err != nil {
		return err
	}
	s.States = cache.Store()
	limit := ratelimit.Middleware(ratelimit.NewMemory(10, time.Minute), ratelimit.ByIP)
	r.Get("/login", func(c *web.Ctx) error {
		return inertia.Render(c, "Login", inertia.Props{"configured": app.Source.Get("SOCIAL_GOOGLE_CLIENT_ID") != ""})
	}).Name("login")
	r.Get("/auth/google/redirect", func(c *web.Ctx) error {
		if app.Source.Get("SOCIAL_GOOGLE_CLIENT_ID") == "" {
			return errs.New(503, "Configure as credenciais Google OAuth para entrar.")
		}
		return s.Redirect(c, "google")
	}).Use(limit).Name("google.redirect")
	r.Get("/auth/google/callback", func(c *web.Ctx) error {
		u, err := s.User(c, "google")
		if err != nil {
			return err
		}
		if !controllers.AllowedGoogle(u.Email, u.EmailVerified) {
			return errs.New(403, "Esta conta não tem acesso ao dashboard.")
		}
		c.Session().Regenerate()
		c.Session().Put("admin", controllers.AdminEmail)
		return c.Redirect("/")
	}).Use(limit).Name("google.callback")
	r.Post("/logout", func(c *web.Ctx) error { c.Session().Invalidate(); inertia.ClearHistory(c); return c.Redirect("/login") }).Name("logout")
	for i, path := range []string{"/", "/subscribers", "/campaigns", "/campaigns/new", "/projects", "/settings"} {
		r.Get(path, controllers.Protected(c.Dashboard)).Name([]string{"dashboard", "subscribers", "campaigns", "campaigns.create", "projects", "settings"}[i])
	}
	r.Post("/api/subscribers/{id}/unsubscribe", controllers.Protected(c.AdminUnsubscribe)).Name("subscribers.unsubscribe")
	r.Post("/api/projects", controllers.Protected(c.CreateProject)).Name("projects.store")
	r.Post("/api/campaigns", controllers.Protected(c.SaveCampaign)).Name("campaigns.store")
	r.Post("/api/campaigns/{id}/send", controllers.Protected(c.SendCampaign)).Name("campaigns.send")
	r.Get("/api/preview/{token}", controllers.Protected(c.PreviewDocument)).Name("campaigns.preview.document")
	r.Post("/api/preview", controllers.Protected(c.Preview)).Name("campaigns.preview")
	r.Get("/api/subscribers/export", controllers.Protected(c.Export)).Name("subscribers.export")
	r.Get("/subscribe", c.Landing).Name("subscribe.host")
	r.Get("/subscribe/{slug}", c.Landing).Name("subscribe.show")
	r.Post("/subscribe/{slug}", c.Subscribe).Use(limit).Name("subscribe.store")
	r.Get("/confirm", c.ConfirmPage).Name("confirm.show")
	r.Post("/confirm", c.Confirm).Use(limit).Name("confirm.store")
	r.Get("/unsubscribe", c.UnsubscribePage).Name("unsubscribe.show")
	r.Post("/unsubscribe", c.Unsubscribe).Use(limit).Name("unsubscribe.store")
	return nil
}
