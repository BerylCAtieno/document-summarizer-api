package services

import (
	"context"
	"fmt"
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

func NewService(repo repository.Repository, cfg *config.Config, logger *utils.Logger) *documentService {
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

	switch req.ContentType {
	case "application/pdf":
		extractedText, err = extractor.ExtractPDF(req.File)
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		extractedText, err = extractor.ExtractDOCX(req.File)
	default:
		return nil, utils.NewBadRequestError("Unsupported file type. Only PDF and DOCX are allowed")
	}

	if err != nil {
		s.logger.Error("Failed to extract text", "error", err)
		return nil, utils.NewInternalError("Failed to extract text from document")
	}

	s3Key := fmt.Sprintf("documents/%s/%s", docID, req.Filename)
	if err := s.storage.Upload(ctx, s3Key, req.File, req.ContentType); err != nil {
		s.logger.Error("Failed to upload to S3", "error", err)
		return nil, utils.NewInternalError("Failed to store document")
	}

	now := time.Now()
	doc := &models.Document{
		ID:            docID,
		Filename:      req.Filename,
		FileSize:      int64(len(req.File)),
		ContentType:   req.ContentType,
		S3Key:         s3Key,
		ExtractedText: extractedText,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		s.logger.Error("Failed to save document to database", "error", err)
		_ = s.storage.Delete(ctx, s3Key)
		return nil, utils.NewInternalError("Failed to save document metadata")
	}

	s.logger.Info("Document uploaded successfully", "id", docID, "filename", req.Filename)

	return &models.UploadResponse{
		ID:          docID,
		Filename:    req.Filename,
		FileSize:    doc.FileSize,
		ContentType: req.ContentType,
		CreatedAt:   now,
		Message:     "Document uploaded successfully. Use /documents/{id}/analyze to analyze it.",
	}, nil
}

func (s *documentService) AnalyzeDocument(ctx context.Context, id string) (*models.AnalysisResponse, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get document", "error", err)
		return nil, utils.NewInternalError("Failed to retrieve document")
	}
	if doc == nil {
		return nil, utils.NewNotFoundError("Document not found")
	}

	if doc.AnalyzedAt != nil {
		return &models.AnalysisResponse{
			ID:           doc.ID,
			Summary:      *doc.Summary,
			DocumentType: *doc.DocumentType,
			Metadata:     doc.Metadata,
			AnalyzedAt:   *doc.AnalyzedAt,
		}, nil
	}

	result, err := s.analyzer.Analyze(ctx, doc.ExtractedText)
	if err != nil {
		s.logger.Error("Failed to analyze document", "error", err)
		return nil, utils.NewInternalError("Failed to analyze document with LLM")
	}

	if err := s.repo.UpdateAnalysis(ctx, id, result.Summary, result.DocumentType, result.Metadata); err != nil {
		s.logger.Error("Failed to update analysis", "error", err)
		return nil, utils.NewInternalError("Failed to save analysis results")
	}

	s.logger.Info("Document analyzed successfully", "id", id, "type", result.DocumentType)

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
		s.logger.Error("Failed to get document", "error", err)
		return nil, utils.NewInternalError("Failed to retrieve document")
	}
	if doc == nil {
		return nil, utils.NewNotFoundError("Document not found")
	}

	return doc, nil
}
