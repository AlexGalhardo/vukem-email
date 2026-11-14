// Package commands holds the application CLI commands.
package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem-email/app/controllers"
	"github.com/AlexGalhardo/vukem-email/app/models"
)

// IntegrationSeed creates a project and prints a new API token for it, so the
// local compose stack can be wired to other apps without Google sign-in.
type IntegrationSeed struct {
	name, slug, domain, from, abilities string
}

func (*IntegrationSeed) Name() string { return "integration:seed" }
func (*IntegrationSeed) Description() string {
	return "Create a project if missing and print a new API token (local integration only)"
}

func (c *IntegrationSeed) Flags(fs *flag.FlagSet) {
	fs.StringVar(&c.name, "name", "", "project name")
	fs.StringVar(&c.slug, "slug", "", "project slug used in /subscribe/{slug}")
	fs.StringVar(&c.domain, "domain", "", "project domain")
	fs.StringVar(&c.from, "from", "", "sender address")
	fs.StringVar(&c.abilities, "abilities", controllers.AbilitySubscribersWrite, "comma separated token abilities")
}

func (c *IntegrationSeed) Run(ctx context.Context, app *vukem.App, _ []string) error {
	if app.Config.IsProduction() {
		return errors.New("integration:seed is for local integration and is refused in production: create projects and tokens in the dashboard")
	}
	if c.name == "" || c.slug == "" || c.domain == "" || c.from == "" {
		return errors.New("integration:seed needs --name, --slug, --domain and --from")
	}
	db, err := app.DB()
	if err != nil {
		return err
	}
	token, err := (&controllers.NewsletterController{App: app, DB: db}).SeedProjectToken(ctx, models.Project{Name: c.name, Slug: c.slug, Domain: c.domain, FromEmail: c.from}, strings.Split(c.abilities, ","))
	if err != nil {
		return err
	}
	fmt.Println(token)
	return nil
}
