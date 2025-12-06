package router

import (
	"net/http"

	"github.com/BerylCAtieno/document-summarizer-api/internal/handlers"
	"github.com/BerylCAtieno/document-summarizer-api/internal/middleware"
	"github.com/BerylCAtieno/document-summarizer-api/internal/services"
	"github.com/BerylCAtieno/document-summarizer-api/internal/utils"

	"github.com/gorilla/mux"
)

func NewRouter(docService services.DocumentService, logger *utils.Logger) http.Handler {
	r := mux.NewRouter()

	// Middlewares
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery(logger))

	// Document handler
	docHandler := handlers.NewDocumentHandler(docService, logger)

	// Routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods(http.MethodGet)

	// Document endpoints
	api.HandleFunc("/documents/upload", docHandler.UploadDocument).Methods(http.MethodPost)
	api.HandleFunc("/documents/{id}/analyze", docHandler.AnalyzeDocument).Methods(http.MethodPost)
	api.HandleFunc("/documents/{id}", docHandler.GetDocument).Methods(http.MethodGet)

	return r
}
