package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BerylCAtieno/document-summarizer-api/internal/utils"

	"github.com/BerylCAtieno/document-summarizer-api/internal/analyzer"
	"github.com/BerylCAtieno/document-summarizer-api/internal/storage"

	"github.com/BerylCAtieno/document-summarizer-api/internal/extractor"

	"github.com/BerylCAtieno/document-summarizer-api/internal/config"

	"github.com/BerylCAtieno/document-summarizer-api/internal/models"
	"github.com/BerylCAtieno/document-summarizer-api/internal/repository"
)

type DocumentService interface {
	UploadDocument(ctx context.Context, req *models.UploadRequest) (*models.UploadResponse, error)
	AnalyzeDocument(ctx context.Context, id string) (*models.AnalysisResponse, error)
	GetDocument(ctx context.Context, id string) (*models.Document, error)
}

type documentService struct {
	repo     repository.Repository
	storage  storage.Storage
	analyzer analyzer.Analyzer
	logger   *utils.Logger
}

func NewService(repo repository.Repository, cfg *config.Config, logger *utils.Logger) DocumentService {
	s3Storage, err := storage.NewS3Storage(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize S3 storage", "error", err)
	}

	llmAnalyzer := analyzer.NewOpenRouterAnalyzer(cfg.OpenRouterAPIKey, cfg.OpenRouterModel, logger)

	return &documentService{
		repo:     repo,
		storage:  s3Storage,
		analyzer: llmAnalyzer,
		logger:   logger,
	}
}

func (s *documentService) UploadDocument(ctx context.Context, req *models.UploadRequest) (*models.UploadResponse, error) {
	docID := utils.GenerateID()

	var extractedText string
	var err error

	// Normalize content type and extract text
	switch {
	case req.ContentType == "application/pdf":
		extractedText, err = extractor.ExtractPDF(req.File)
	case isDOCXContentType(req.ContentType):
		extractedText, err = extractor.ExtractDOCX(req.File)
	default:
		s.logger.Warn("Unsupported content type", "content_type", req.ContentType, "filename", req.Filename)
		return nil, utils.NewBadRequestError(fmt.Sprintf("Unsupported file type '%s'. Only PDF and DOCX are allowed", req.ContentType))
	}

	if err != nil {
		s.logger.Error("Failed to extract text", "error", err, "content_type", req.ContentType, "filename", req.Filename)
		return nil, utils.NewInternalError(fmt.Sprintf("Failed to extract text from document: %v", err))
	}

	// Validate extracted text is not empty
	if strings.TrimSpace(extractedText) == "" {
		s.logger.Warn("No text extracted from document", "filename", req.Filename)
		return nil, utils.NewBadRequestError("No text could be extracted from the document. The file may be empty or corrupted")
	}

	s3Key := fmt.Sprintf("documents/%s/%s", docID, req.Filename)
	if err := s.storage.Upload(ctx, s3Key, req.File, req.ContentType); err != nil {
		s.logger.Error("Failed to upload to S3", "error", err, "s3_key", s3Key)
		return nil, utils.NewInternalError("Failed to store document")
	}

	now := time.Now()
	doc := &models.Document{
		ID:            docID,
		Filename:      req.Filename,
		FileSize:      int64(len(req.File)),
		ContentType:   normalizeContentType(req.ContentType),
		S3Key:         s3Key,
		ExtractedText: extractedText,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		s.logger.Error("Failed to save document to database", "error", err, "doc_id", docID)
		// Attempt to cleanup S3
		_ = s.storage.Delete(ctx, s3Key)
		return nil, utils.NewInternalError("Failed to save document metadata")
	}

	s.logger.Info("Document uploaded successfully",
		"id", docID,
		"filename", req.Filename,
		"content_type", req.ContentType,
		"text_length", len(extractedText))

	return &models.UploadResponse{
		ID:          docID,
		Filename:    req.Filename,
		FileSize:    doc.FileSize,
		ContentType: doc.ContentType,
		CreatedAt:   now,
		Message:     "Document uploaded successfully. Use /documents/{id}/analyze to analyze it.",
	}, nil
}

func (s *documentService) AnalyzeDocument(ctx context.Context, id string) (*models.AnalysisResponse, error) {
	// Get document from database
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get document", "error", err, "id", id)
		return nil, utils.NewInternalError("Failed to retrieve document")
	}
	if doc == nil {
		return nil, utils.NewNotFoundError("Document not found")
	}

	// Check if already analyzed
	if doc.AnalyzedAt != nil {
		s.logger.Info("Document already analyzed, returning cached results", "id", id)
		return &models.AnalysisResponse{
			ID:           doc.ID,
			Summary:      *doc.Summary,
			DocumentType: *doc.DocumentType,
			Metadata:     doc.Metadata,
			AnalyzedAt:   *doc.AnalyzedAt,
		}, nil
	}

	// Analyze with LLM
	s.logger.Info("Starting document analysis", "id", id, "text_length", len(doc.ExtractedText))
	result, err := s.analyzer.Analyze(ctx, doc.ExtractedText)
	if err != nil {
		s.logger.Error("Failed to analyze document", "error", err, "id", id)
		return nil, utils.NewInternalError("Failed to analyze document with LLM")
	}

	// Update database with analysis results
	if err := s.repo.UpdateAnalysis(ctx, id, result.Summary, result.DocumentType, result.Metadata); err != nil {
		s.logger.Error("Failed to update analysis", "error", err, "id", id)
		return nil, utils.NewInternalError("Failed to save analysis results")
	}

	s.logger.Info("Document analyzed successfully",
		"id", id,
		"type", result.DocumentType,
		"summary_length", len(result.Summary))

	return &models.AnalysisResponse{
		ID:           id,
		Summary:      result.Summary,
		DocumentType: result.DocumentType,
		Metadata:     result.Metadata,
		AnalyzedAt:   time.Now(),
	}, nil
}

func (s *documentService) GetDocument(ctx context.Context, id string) (*models.Document, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get document", "error", err, "id", id)
		return nil, utils.NewInternalError("Failed to retrieve document")
	}
	if doc == nil {
		return nil, utils.NewNotFoundError("Document not found")
	}

	return doc, nil
}

// isDOCXContentType checks if the content type is a DOCX file
// Handles various DOCX MIME type variations
func isDOCXContentType(contentType string) bool {
	docxTypes := []string{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.wordprocessingml",
		"application/docx",
		"application/x-docx",
	}

	for _, docxType := range docxTypes {
		if contentType == docxType {
			return true
		}
	}

	return false
}

// normalizeContentType normalizes content type to standard MIME types
func normalizeContentType(contentType string) string {
	if isDOCXContentType(contentType) {
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	return contentType
}
