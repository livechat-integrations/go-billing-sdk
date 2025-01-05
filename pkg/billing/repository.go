package billing

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/livechat-integrations/go-billing-sdk/pkg/billing/db"
	"time"
)

type PostgresqlStorage struct {
	queries *db.Queries
}

func NewPostgresqlStorage(queries *db.Queries) *PostgresqlStorage {
	return &PostgresqlStorage{queries: queries}
}

func (r *PostgresqlStorage) CreateCharge(ctx context.Context, ic InstallationCharge) error {
	rawPayload, err := json.Marshal(ic.Charge.Payload)
	if err != nil {
		return err
	}

	if err = r.queries.CreateCharge(ctx, db.CreateChargeParams{
		ID:      ic.Charge.ID,
		Type:    string(ic.Charge.Type),
		Payload: rawPayload,
	}); err != nil {
		return err
	}

	if err = r.queries.CreateInstallationCharge(ctx, db.CreateInstallationChargeParams{
		InstallationID: ic.InstallationID,
		ChargeID:       pgtype.Text{String: ic.Charge.ID, Valid: true},
	}); err != nil {
		return err
	}

	return nil
}

func (r *PostgresqlStorage) GetCharge(ctx context.Context, id string) (*InstallationCharge, error) {
	row, err := r.queries.GetChargeByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var canceledAt *time.Time
	if row.DeletedAt.Valid {
		canceledAt = &row.DeletedAt.Time
	}

	charge := &Charge{
		ID:         row.ID,
		Type:       ChargeType(row.Type),
		CreatedAt:  row.CreatedAt.Time,
		CanceledAt: canceledAt,
	}

	_ = json.Unmarshal(row.Payload, &charge.Payload)

	return &InstallationCharge{
		InstallationID: row.InstallationID.String,
		Charge:         charge,
	}, nil
}

func (r *PostgresqlStorage) GetChargeByInstallationID(ctx context.Context, lcID string) (*InstallationCharge, error) {
	row, err := r.queries.GetChargeByInstallationID(ctx, lcID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if !row.ID.Valid {
		return &InstallationCharge{InstallationID: lcID}, nil
	}

	var canceledAt *time.Time
	if row.DeletedAt.Valid {
		canceledAt = &row.DeletedAt.Time
	}

	charge := &Charge{
		ID:         row.ID.String,
		Type:       ChargeType(row.Type.String),
		CreatedAt:  row.CreatedAt.Time,
		CanceledAt: canceledAt,
	}

	_ = json.Unmarshal(row.Payload, &charge.Payload)

	return &InstallationCharge{
		InstallationID: row.InstallationID,
		Charge:         charge,
	}, nil
}

func (r *PostgresqlStorage) UpdateChargePayload(ctx context.Context, id string, payload BaseCharge) error {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.queries.UpdateCharge(ctx, db.UpdateChargeParams{
		ID:      id,
		Payload: rawPayload,
	})
}
