package controllers

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/AlexGalhardo/vukem-email/app/models"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/web"
)

// API token abilities. A token acts for exactly one project and carries only
// the abilities the admin ticked when creating it; there is no wildcard.
const (
	AbilitySubscribersWrite = "subscribers:write"
	AbilityCampaignsSend    = "campaigns:send"
	tokenPrefix             = "vke_"
	maxTokensPerProject     = 20
)

var allAbilities = []string{AbilitySubscribersWrite, AbilityCampaignsSend}

type projectKey struct{}

func (n *NewsletterController) tokens() *auth.Tokens {
	t := auth.NewTokens(n.DB)
	t.Prefix, t.Now, t.Log = tokenPrefix, n.App.Now, n.App.Log
	return t
}

// APIToken authenticates "Authorization: Bearer" requests against the
// project API tokens and requires ability. The token's project reaches the
// handler through apiProject.
func (n *NewsletterController) APIToken(ability string, next web.HandlerFunc) web.HandlerFunc {
	unauthorized := errs.ErrUnauthorized.WithHeader("WWW-Authenticate", `Bearer realm="api"`)
	return func(c *web.Ctx) error {
		plain, ok := auth.BearerToken(c.Request())
		if !ok {
			return unauthorized
		}
		tok, err := n.tokens().Find(c.Context(), plain)
		if err != nil {
			// auth.Tokens does not export its bad-token error; store failures carry this prefix.
			if strings.HasPrefix(err.Error(), "auth: find token") {
				return errs.Wrap(http.StatusServiceUnavailable, err)
			}
			return unauthorized
		}
		projectID, err := strconv.ParseInt(tok.UserID, 10, 64)
		if err != nil || projectID <= 0 {
			return unauthorized
		}
		if !tok.Can(ability) {
			return errs.New(http.StatusForbidden, "This token cannot perform this action.")
		}
		c.SetRequest(c.Request().WithContext(context.WithValue(c.Context(), projectKey{}, projectID)))
		return next(c)
	}
}

func apiProject(c *web.Ctx) int64 {
	id, _ := c.Context().Value(projectKey{}).(int64)
	return id
}

// APISubscribe creates a subscriber of the token's project (the project is
// its list) and sends the double opt-in confirmation. The API cannot mark an
// address confirmed: consent still comes from the link in the email.
func (n *NewsletterController) APISubscribe(c *web.Ctx) error {
	p, err := orm.QueryOf[models.Project](n.DB).Find(c.Context(), apiProject(c))
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
	blocked, err := orm.QueryOf[models.Suppression](n.DB).Where("project_id", "=", p.ID).Where("email", "=", email).Count(c.Context())
	if err != nil {
		return err
	}
	if blocked > 0 {
		return errs.New(409, "This address is suppressed for this project (bounce or complaint).")
	}
	if err := n.requestConfirmation(c.Context(), p, email); err != nil {
		return err
	}
	s, err := orm.QueryOf[models.Subscriber](n.DB).Where("project_id", "=", p.ID).Where("email", "=", email).First(c.Context())
	if errors.Is(err, orm.ErrNotFound) {
		return errs.New(429, "Confirmation limit reached for this project. Retry later.")
	}
	if err != nil {
		return err
	}
	return c.JSON(http.StatusAccepted, web.Map{"id": s.ID, "email": s.Email, "status": s.Status})
}

// APIUpdateSubscriber lets a token unsubscribe one of its project's
// subscribers. Other transitions are refused: confirming needs the email link,
// and suppression belongs to the bounce webhook.
func (n *NewsletterController) APIUpdateSubscriber(c *web.Ctx) error {
	in, err := input(c)
	if err != nil {
		return err
	}
	if in["status"] != "unsubscribed" {
		return errs.New(422, `Only {"status":"unsubscribed"} is supported.`)
	}
	s, err := orm.QueryOf[models.Subscriber](n.DB).Where("project_id", "=", apiProject(c)).Find(c.Context(), number(c.Param("id")))
	if err != nil {
		return err
	}
	if err := n.cancelSubscription(c, s.ID, s.ProjectID); err != nil {
		return err
	}
	s, err = orm.QueryOf[models.Subscriber](n.DB).Find(c.Context(), s.ID)
	if err != nil {
		return err
	}
	return c.JSON(200, web.Map{"id": s.ID, "email": s.Email, "status": s.Status})
}

// APISendCampaign queues a draft campaign of the token's project.
func (n *NewsletterController) APISendCampaign(c *web.Ctx) error {
	return n.queueCampaign(c, number(c.Param("id")), apiProject(c))
}

// CreateAPIToken issues a token for a project. The plain value is returned
// once; only a hash is stored.
func (n *NewsletterController) CreateAPIToken(c *web.Ctx) error {
	p, err := orm.QueryOf[models.Project](n.DB).Find(c.Context(), number(c.Param("id")))
	if err != nil {
		return err
	}
	in, err := input(c)
	if err != nil {
		return err
	}
	if len(in["name"]) < 1 || len(in["name"]) > 100 {
		return errs.New(422, "Name the token (up to 100 characters).")
	}
	var abilities []string
	for a := range strings.SplitSeq(in["abilities"], ",") {
		if a = strings.TrimSpace(a); a == "" || slices.Contains(abilities, a) {
			continue
		}
		if !slices.Contains(allAbilities, a) {
			return errs.New(422, "Unknown ability "+strconv.Quote(a)+".")
		}
		abilities = append(abilities, a)
	}
	if len(abilities) == 0 {
		return errs.New(422, "Choose at least one ability: "+strings.Join(allAbilities, ", ")+".")
	}
	owner := strconv.FormatInt(p.ID, 10)
	live, err := n.liveTokens(c.Context(), owner)
	if err != nil {
		return err
	}
	if len(live) >= maxTokensPerProject {
		return errs.New(409, "This project has too many active tokens. Revoke one first.")
	}
	plain, tok, err := n.tokens().Create(c.Context(), owner, in["name"], abilities, 0)
	if err != nil {
		return err
	}
	return c.JSON(201, web.Map{"token": plain, "id": tok.ID, "name": tok.Name, "abilities": tok.Abilities})
}

func (n *NewsletterController) liveTokens(ctx context.Context, owner string) ([]auth.Token, error) {
	all, err := n.tokens().List(ctx, owner)
	if err != nil {
		return nil, err
	}
	// Never nil: it is an Inertia prop and a JSON null would crash the page.
	return slices.DeleteFunc(append([]auth.Token{}, all...), func(t auth.Token) bool { return t.RevokedAt != nil }), nil
}

// RevokeAPIToken revokes a token of the project in the path.
func (n *NewsletterController) RevokeAPIToken(c *web.Ctx) error {
	revoked, err := n.tokens().RevokeFor(c.Context(), strconv.FormatInt(number(c.Param("id")), 10), c.Param("token"))
	if err != nil {
		return err
	}
	if !revoked {
		return errs.ErrNotFound
	}
	return c.JSON(200, web.Map{"message": "Token revoked."})
}

// SeedProjectToken returns a new API token for the project with that slug,
// creating the project first when it does not exist. It backs the local
// integration seed command and is never routed.
func (n *NewsletterController) SeedProjectToken(ctx context.Context, p models.Project, abilities []string) (string, error) {
	existing, err := orm.QueryOf[models.Project](n.DB).Where("slug", "=", p.Slug).First(ctx)
	switch {
	case err == nil:
		p = existing
	case errors.Is(err, orm.ErrNotFound):
		if err := orm.Create(ctx, n.DB, &p); err != nil {
			return "", err
		}
	default:
		return "", err
	}
	for i, a := range abilities {
		abilities[i] = strings.TrimSpace(a)
		if !slices.Contains(allAbilities, abilities[i]) {
			return "", errors.New("unknown ability " + strconv.Quote(a))
		}
	}
	plain, _, err := n.tokens().Create(ctx, strconv.FormatInt(p.ID, 10), "integration", abilities, 0)
	return plain, err
}
