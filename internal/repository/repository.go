package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/BerylCAtieno/document-summarizer-api/internal/models"
	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, doc *models.Document) error
	GetByID(ctx context.Context, id string) (*models.Document, error)
	Update(ctx context.Context, doc *models.Document) error
	UpdateAnalysis(ctx context.Context, id, summary, docType string, metadata map[string]interface{}) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, doc *models.Document) error {
	query := `
		INSERT INTO documents (id, filename, file_size, content_type, s3_key, extracted_text, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		doc.ID,
		doc.Filename,
		doc.FileSize,
		doc.ContentType,
		doc.S3Key,
		doc.ExtractedText,
		doc.CreatedAt,
		doc.UpdatedAt,
	)

	return err
}

func (r *repository) GetByID(ctx context.Context, id string) (*models.Document, error) {
	var doc models.Document
	var metadataJSON sql.NullString

	query := `
		SELECT id, filename, file_size, content_type, s3_key, extracted_text, 
		       summary, document_type, metadata, created_at, updated_at, analyzed_at
		FROM documents
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&doc.ID,
		&doc.Filename,
		&doc.FileSize,
		&doc.ContentType,
		&doc.S3Key,
		&doc.ExtractedText,
		&doc.Summary,
		&doc.DocumentType,
		&metadataJSON,
		&doc.CreatedAt,
		&doc.UpdatedAt,
		&doc.AnalyzedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if metadataJSON.Valid && metadataJSON.String != "" {
		if err := json.Unmarshal([]byte(metadataJSON.String), &doc.Metadata); err != nil {
			return nil, err
		}
	}

	return &doc, nil
}

func (r *repository) Update(ctx context.Context, doc *models.Document) error {
	query := `
		UPDATE documents
		SET filename = $2, file_size = $3, content_type = $4, extracted_text = $5, updated_at = $6
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		doc.ID,
		doc.Filename,
		doc.FileSize,
		doc.ContentType,
		doc.ExtractedText,
		time.Now(),
	)

	return err
}

func (r *repository) UpdateAnalysis(ctx context.Context, id, summary, docType string, metadata map[string]interface{}) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	query := `
		UPDATE documents
		SET summary = $2, document_type = $3, metadata = $4, analyzed_at = $5, updated_at = $6
		WHERE id = $1
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query, id, summary, docType, metadataJSON, now, now)

	return err
}
