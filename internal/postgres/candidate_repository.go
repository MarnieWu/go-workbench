package postgres

import (
	"context"
	"go-workbench/internal/candidate"

	"github.com/jackc/pgx/v5"
)

type CandidateRepository struct {
	conn *pgx.Conn
}

func NewCandidateRepository(conn *pgx.Conn) *CandidateRepository {
	return &CandidateRepository{conn: conn}
}

func (r *CandidateRepository) ListPending(ctx context.Context, ownerID string) ([]candidate.Candidate, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT id::text, owner_id::text, capture_id::text, proposed_title,
			proposed_description, proposed_project_id::text, labels, status, created_at, updated_at
		FROM candidates
		WHERE owner_id = $1::uuid AND status = 'pending_review'
		ORDER BY created_at DESC, id DESC
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]candidate.Candidate, 0)
	for rows.Next() {
		var item candidate.Candidate
		if err := rows.Scan(
			&item.ID,
			&item.OwnerID,
			&item.CaptureID,
			&item.ProposedTitle,
			&item.ProposedDescription,
			&item.ProposedProjectID,
			&item.Labels,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
