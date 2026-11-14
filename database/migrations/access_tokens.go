package migrations

import "github.com/AlexGalhardo/vukem/schema"

// accessTokens is the table behind auth.Tokens (see the vukem API tokens guide).
// user_id holds the project id: a token acts for exactly one project.
func accessTokens() schema.Migration {
	return schema.Migration{ID: "20261114000002_create_access_tokens", Up: func(s *schema.Schema) error {
		return s.Create("access_tokens", func(t *schema.Table) {
			t.String("id").Unique()
			t.String("user_id").Index()
			t.String("name")
			t.String("hash")
			t.Text("abilities")
			t.Timestamp("last_used_at").Nullable()
			t.Timestamp("expires_at").Nullable()
			t.Timestamp("revoked_at").Nullable()
			t.Timestamp("created_at")
		})
	}, Down: func(s *schema.Schema) error { return s.DropIfExists("access_tokens") }}
}
