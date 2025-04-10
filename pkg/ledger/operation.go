package ledger

import (
	"encoding/json"
	"time"
)

type Operation struct {
	ID               string          `json:"id"`
	LCOrganizationID string          `json:"lc_organization_id"`
	Amount           float32         `json:"amount"`
	Payload          json.RawMessage `json:"payload"`
	CreatedAt        time.Time       `json:"created_at"`
}
