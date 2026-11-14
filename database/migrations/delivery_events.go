package migrations

import "github.com/AlexGalhardo/vukem/schema"

// deliveryEvents stores the Resend open and click events per delivery. event_id
// is the Svix message id, so a redelivered event is stored once.
func deliveryEvents() schema.Migration {
	return schema.Migration{ID: "20261114000001_create_delivery_events", Up: func(s *schema.Schema) error {
		return s.Create("delivery_events", func(t *schema.Table) {
			t.ID()
			t.ForeignID("delivery_id").Constrained()
			t.ForeignID("campaign_id").Constrained()
			t.String("type")
			t.Text("url")
			t.String("event_id").Unique()
			t.Timestamps()
			t.Index("campaign_id", "type")
			t.Index("delivery_id")
		})
	}, Down: func(s *schema.Schema) error { return s.DropIfExists("delivery_events") }}
}
