package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/BerylCAtieno/document-summarizer-api/internal/models"
	"github.com/BerylCAtieno/document-summarizer-api/internal/services"
	"github.com/BerylCAtieno/document-summarizer-api/internal/utils"
	"github.com/gorilla/mux"
)

type DocumentHandler struct {
	service services.DocumentService
	logger  *utils.Logger
}

func NewDocumentHandler(service services.DocumentService, logger *utils.Logger) *DocumentHandler {
	return &DocumentHandler{
		service: service,
		logger:  logger,
	}
}

func (h *DocumentHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 5MB)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		h.respondError(w, utils.NewBadRequestError("File too large or invalid form data"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.respondError(w, utils.NewBadRequestError("No file provided"))
		return
	}
	defer file.Close()

	// Determine content type with fallback to file extension
	contentType := determineContentType(header.Filename, header.Header.Get("Content-Type"))

	h.logger.Info("File upload attempt",
		"filename", header.Filename,
		"reported_content_type", header.Header.Get("Content-Type"),
		"determined_content_type", contentType)

	// Validate content type
	if !isValidContentType(contentType) {
		h.respondError(w, utils.NewBadRequestError("Only PDF and DOCX files are allowed"))
		return
	}

	// Read file data
	data, err := io.ReadAll(file)
	if err != nil {
		h.respondError(w, utils.NewInternalError("Failed to read file"))
		return
	}

	// Validate file size (5MB)
	if len(data) > 5*1024*1024 {
		h.respondError(w, utils.NewBadRequestError("File size exceeds 5MB limit"))
		return
	}

	// Validate file is not empty
	if len(data) == 0 {
		h.respondError(w, utils.NewBadRequestError("Uploaded file is empty"))
		return
	}

	// Process upload
	req := &models.UploadRequest{
		File:        data,
		Filename:    header.Filename,
		ContentType: contentType,
	}

	resp, err := h.service.UploadDocument(r.Context(), req)
	if err != nil {
		h.respondError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, resp)
}

func (h *DocumentHandler) AnalyzeDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.respondError(w, utils.NewBadRequestError("Document ID is required"))
		return
	}

	resp, err := h.service.AnalyzeDocument(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.respondError(w, utils.NewBadRequestError("Document ID is required"))
		return
	}

	doc, err := h.service.GetDocument(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, doc)
}

// determineContentType determines the content type from filename extension
// with fallback to the provided content type header
func determineContentType(filename, headerContentType string) string {
	// Get file extension
	ext := strings.ToLower(filepath.Ext(filename))

	// Map extensions to MIME types
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".doc":
		// Note: .doc is not supported, but we can give a better error message
		return "application/msword"
	}

	// If no extension match, use the header content type if valid
	if isValidContentType(headerContentType) {
		return headerContentType
	}

	// Return the header content type anyway (will be validated later)
	return headerContentType
}

// isValidContentType checks if the content type is supported
func isValidContentType(contentType string) bool {
	validTypes := map[string]bool{
		"application/pdf": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		// Some browsers might send these variants for DOCX
		"application/vnd.openxmlformats-officedocument.wordprocessingml": true,
	}

	return validTypes[contentType]
}

func (h *DocumentHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

func (h *DocumentHandler) respondError(w http.ResponseWriter, err error) {
	var status int
	var message string

	switch e := err.(type) {
	case *utils.AppError:
		status = e.StatusCode
		message = e.Message
	default:
		status = http.StatusInternalServerError
		message = "Internal server error"
	}

	h.logger.Error("Request error", "status", status, "error", message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
