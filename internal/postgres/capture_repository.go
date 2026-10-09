package postgres

import (
	"context"
	"errors"
	"go-workbench/internal/capture"

	"github.com/jackc/pgx/v5"
)

type CaptureRepository struct {
	conn *pgx.Conn
}

func NewCaptureRepository(conn *pgx.Conn) *CaptureRepository {
	return &CaptureRepository{conn: conn}
}

func (r *CaptureRepository) Create(
	ctx context.Context,
	record capture.CreateRecord,
) (capture.Capture, error) {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return capture.Capture{}, err
	}
	defer tx.Rollback(ctx)

	var created capture.Capture
	err = tx.QueryRow(ctx, `
		INSERT INTO captures (
			owner_id,
			idempotency_key,
			input_hash,
			input_text,
			source_type
		)
		VALUES ($1::uuid, $2, $3, $4, $5)
		ON CONFLICT (owner_id, idempotency_key) DO NOTHING
		RETURNING id::text, owner_id::text, idempotency_key, input_hash, input_text, source_type, status, created_at, updated_at
	`, record.OwnerID, record.IdempotencyKey, record.InputHash, record.InputText, record.SourceType).Scan(
		&created.ID,
		&created.OwnerID,
		&created.IdempotencyKey,
		&created.InputHash,
		&created.InputText,
		&created.SourceType,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err == nil {
		if _, err = tx.Exec(ctx, `
			INSERT INTO source_evidence (
				owner_id,
				capture_id,
				source_type,
				external_ref,
				source_url,
				excerpt,
				consent_scope
			)
			VALUES (
				$1::uuid,
				$2::uuid,
				$3,
				$4,
				$5,
				$6,
				$7
			)
		`,
			record.OwnerID,
			created.ID,
			record.SourceType,
			record.ExternalRef,
			record.SourceURL,
			record.Excerpt,
			record.ConsentScope); err != nil {
			return capture.Capture{}, err
		}

		if err = tx.Commit(ctx); err != nil {
			return capture.Capture{}, err
		}

		return created, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		// if err is not pgx.ErrNoRows, it must be a real database error
		return capture.Capture{}, err
	}

	// pgx.ErrNoRows means conflict path: fetch existing row
	var existing capture.Capture
	err = tx.QueryRow(ctx, `
		SELECT
			id::text,
			owner_id::text,
			idempotency_key,
			input_hash,
			input_text,
			source_type,
			status,
			created_at,
			updated_at
		FROM captures
		WHERE owner_id = $1::uuid
			AND idempotency_key = $2
		LIMIT 1
	`, record.OwnerID, record.IdempotencyKey).Scan(
		&existing.ID,
		&existing.OwnerID,
		&existing.IdempotencyKey,
		&existing.InputHash,
		&existing.InputText,
		&existing.SourceType,
		&existing.Status,
		&existing.CreatedAt,
		&existing.UpdatedAt,
	)

	if err != nil {
		return capture.Capture{}, err
	}
	if existing.InputHash != record.InputHash {
		return capture.Capture{}, capture.ErrIdempotencyConflict
	}
	if err := tx.Commit(ctx); err != nil {
		return capture.Capture{}, err
	}

	return existing, nil
}
