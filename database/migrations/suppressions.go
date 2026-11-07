package migrations

import "github.com/AlexGalhardo/vukem/schema"

func suppressions() schema.Migration {
	return schema.Migration{ID: "20261107000000_create_suppressions", Up: func(s *schema.Schema) error {
		return s.Create("suppressions", func(t *schema.Table) {
			t.ID()
			t.ForeignID("project_id").Constrained()
			t.String("email")
			t.String("reason")
			t.Timestamps()
			t.UniqueIndex("project_id", "email")
		})
	}, Down: func(s *schema.Schema) error { return s.DropIfExists("suppressions") }}
}
