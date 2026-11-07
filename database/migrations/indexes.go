package migrations

import "github.com/AlexGalhardo/vukem/schema"

// indexes covers the hot queries: the delivery worker counts queued rows per
// campaign after every job, and the dashboard groups deliveries by status and
// lists campaigns per project.
func indexes() schema.Migration {
	return schema.Migration{ID: "20261107000001_add_query_indexes", Up: func(s *schema.Schema) error {
		if err := s.Table("deliveries", func(t *schema.Table) { t.Index("campaign_id", "status") }); err != nil {
			return err
		}
		return s.Table("campaigns", func(t *schema.Table) { t.Index("project_id", "id") })
	}, Down: func(s *schema.Schema) error {
		if err := s.Table("deliveries", func(t *schema.Table) { t.DropIndex("deliveries_campaign_id_status_index") }); err != nil {
			return err
		}
		return s.Table("campaigns", func(t *schema.Table) { t.DropIndex("campaigns_project_id_id_index") })
	}}
}
