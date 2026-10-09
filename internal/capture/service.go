package capture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

type Status string

const (
	StatusQueued     Status = "queued"
	StatusProcessing Status = "processing"
	StatusProcessed  Status = "processed"
	StatusFailed     Status = "failed"
)

var (
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrInvalidInput        = errors.New("invalid input")
)

type Capture struct {
	ID             string
	OwnerID        string
	IdempotencyKey string
	InputHash      string
	InputText      string
	SourceType     string
	Status         Status
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateInput struct {
	OwnerID        string
	IdempotencyKey string
	InputText      string
	SourceType     string
	ExternalRef    *string
	SourceURL      *string
	Excerpt        *string
	ConsentScope   *string
}

type CreateRecord struct {
	OwnerID        string
	IdempotencyKey string
	InputHash      string
	InputText      string
	SourceType     string
	ExternalRef    *string
	SourceURL      *string
	Excerpt        *string
	ConsentScope   *string
}

type Repository interface {
	Create(ctx context.Context, record CreateRecord) (Capture, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Capture, error) {
	text := strings.TrimSpace(input.InputText)
	sum := sha256.Sum256([]byte(text))
	inputHash := hex.EncodeToString(sum[:])

	return s.repository.Create(ctx, CreateRecord{
		OwnerID:        input.OwnerID,
		IdempotencyKey: input.IdempotencyKey,
		InputHash:      inputHash,
		InputText:      text,
		SourceType:     input.SourceType,
		ExternalRef:    input.ExternalRef,
		SourceURL:      input.SourceURL,
		Excerpt:        input.Excerpt,
		ConsentScope:   input.ConsentScope,
	})
}
