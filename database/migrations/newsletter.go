package migrations

import "github.com/AlexGalhardo/vukem/schema"

func newsletter() schema.Migration {
	return schema.Migration{ID: "20261104000000_create_newsletter", Up: func(s *schema.Schema) error {
		if err := s.Create("projects", func(t *schema.Table) {
			t.ID()
			t.String("name")
			t.String("slug").Unique()
			t.String("domain").Unique()
			t.String("from_email")
			t.Text("description")
			t.Timestamps()
		}); err != nil {
			return err
		}
		if err := s.Create("subscribers", func(t *schema.Table) {
			t.ID()
			t.ForeignID("project_id").Constrained()
			t.String("email")
			t.String("status")
			t.String("token_hash")
			t.BigInteger("expires_at")
			t.BigInteger("last_requested_at")
			t.Timestamps()
			t.UniqueIndex("project_id", "email")
			t.Index("token_hash")
			t.Index("project_id", "status")
		}); err != nil {
			return err
		}
		if err := s.Create("campaigns", func(t *schema.Table) {
			t.ID()
			t.ForeignID("project_id").Constrained()
			t.String("subject")
			t.String("mode")
			t.Text("source")
			t.Text("body_html")
			t.String("status")
			t.BigInteger("recipients")
			t.Timestamps()
		}); err != nil {
			return err
		}
		return s.Create("deliveries", func(t *schema.Table) {
			t.ID()
			t.ForeignID("campaign_id").Constrained()
			t.ForeignID("subscriber_id").Constrained()
			t.String("status")
			t.Timestamps()
			t.UniqueIndex("campaign_id", "subscriber_id")
		})
	}, Down: func(s *schema.Schema) error {
		for _, table := range []string{"deliveries", "campaigns", "subscribers", "projects"} {
			if err := s.DropIfExists(table); err != nil {
				return err
			}
		}
		return nil
	}}
}
