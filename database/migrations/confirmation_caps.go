package migrations

import "github.com/AlexGalhardo/vukem/schema"

// confirmationCaps counts confirmation emails per address in a rolling 24 hour
// window (SA-001 in docs/security-audit-2026-11-07.md).
func confirmationCaps() schema.Migration {
	return schema.Migration{ID: "20261107000002_add_confirmation_caps", Up: func(s *schema.Schema) error {
		return s.Table("subscribers", func(t *schema.Table) {
			t.BigInteger("confirmations").Default(0)
			t.BigInteger("confirmations_since").Default(0)
		})
	}, Down: func(s *schema.Schema) error {
		return s.Table("subscribers", func(t *schema.Table) { t.DropColumn("confirmations", "confirmations_since") })
	}}
}
