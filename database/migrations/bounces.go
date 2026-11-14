package migrations

import "github.com/AlexGalhardo/vukem/schema"

// bounces logs every bounce and complaint reported by Resend. event_id is the
// Svix message id, so a redelivered event is stored once.
func bounces() schema.Migration {
	return schema.Migration{ID: "20261114000000_create_bounces", Up: func(s *schema.Schema) error {
		return s.Create("bounces", func(t *schema.Table) {
			t.ID()
			t.ForeignID("project_id").Constrained()
			t.String("email")
			t.BigInteger("subscriber_id").Default(0)
			t.BigInteger("delivery_id").Default(0)
			t.String("type")
			t.Text("reason")
			t.String("event_id").Unique()
			t.Timestamps()
			t.Index("project_id", "email", "type")
		})
	}, Down: func(s *schema.Schema) error { return s.DropIfExists("bounces") }}
}
