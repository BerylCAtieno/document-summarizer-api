package models

import (
	"time"
)

type Document struct {
	ID            string                 `json:"id" db:"id"`
	Filename      string                 `json:"filename" db:"filename"`
	FileSize      int64                  `json:"file_size" db:"file_size"`
	ContentType   string                 `json:"content_type" db:"content_type"`
	S3Key         string                 `json:"s3_key" db:"s3_key"`
	ExtractedText string                 `json:"extracted_text,omitempty" db:"extracted_text"`
	Summary       *string                `json:"summary,omitempty" db:"summary"`
	DocumentType  *string                `json:"document_type,omitempty" db:"document_type"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
	AnalyzedAt    *time.Time             `json:"analyzed_at,omitempty" db:"analyzed_at"`
}

type UploadRequest struct {
	File        []byte
	Filename    string
	ContentType string
}

type UploadResponse struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	FileSize    int64     `json:"file_size"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
	Message     string    `json:"message"`
}

type AnalysisResponse struct {
	ID           string                 `json:"id"`
	Summary      string                 `json:"summary"`
	DocumentType string                 `json:"document_type"`
	Metadata     map[string]interface{} `json:"metadata"`
	AnalyzedAt   time.Time              `json:"analyzed_at"`
}

type LLMAnalysisResult struct {
	Summary      string                 `json:"summary"`
	DocumentType string                 `json:"document_type"`
	Metadata     map[string]interface{} `json:"metadata"`
}
