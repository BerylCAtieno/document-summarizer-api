# Document Analysis API

A Go-based REST API for document upload, text extraction, and AI-powered analysis using OpenRouter LLM.

## Features

- Upload PDF and DOCX files (max 5MB)
- Automatic text extraction
- AI-powered document analysis (summary, type detection, metadata extraction)
- S3/Minio storage for raw files
- Database storage for metadata and analysis results


## Prerequisites

- Go 1.21+
- SQlite
- Minio
- OpenRouter API key

## Setup

### 1. Environment Variables

Create a `.env` file:

```bash
PORT=8080
DATABASE_URL=sqliteurl
LOG_LEVEL=info

# S3/Minio
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_BUCKET_NAME=documents
S3_USE_SSL=false

# OpenRouter
OPENROUTER_API_KEY=your_openrouter_api_key
OPENROUTER_MODEL=openai/gpt-4o-mini
```

### 2. Install Dependencies

```bash
go mod download
```

### 4. Setup Minio (For Local Development)

```bash
# Using Docker
docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"
```

### 5. Run the API

```bash
go run cmd/server/main.go
```

## API Endpoints

### Health Check

```bash
GET /api/v1/health
```

### Upload Document

```bash
POST /api/v1/documents/upload
Content-Type: multipart/form-data

Form data:
- file: PDF or DOCX file (max 5MB)

Response:
{
  "id": "abc123...",
  "filename": "document.pdf",
  "file_size": 123456,
  "content_type": "application/pdf",
  "created_at": "2024-01-01T12:00:00Z",
  "message": "Document uploaded successfully. Use /documents/{id}/analyze to analyze it."
}
```

### Analyze Document

```bash
POST /api/v1/documents/{id}/analyze

Response:
{
  "id": "abc123...",
  "summary": "This is a concise summary of the document...",
  "document_type": "invoice",
  "metadata": {
    "date": "2024-01-01",
    "sender": "Company Inc.",
    "amount": "1500.00",
    "currency": "USD"
  },
  "analyzed_at": "2024-01-01T12:00:30Z"
}
```

### Get Document

```bash
GET /api/v1/documents/{id}

Response:
{
  "id": "abc123...",
  "filename": "document.pdf",
  "file_size": 123456,
  "content_type": "application/pdf",
  "s3_key": "documents/abc123.../document.pdf",
  "extracted_text": "Full extracted text...",
  "summary": "This is a concise summary...",
  "document_type": "invoice",
  "metadata": {
    "date": "2024-01-01",
    "amount": "1500.00"
  },
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:30Z",
  "analyzed_at": "2024-01-01T12:00:30Z"
}
```

## Testing with cURL

### Upload a PDF

```bash
curl -X POST http://localhost:8080/api/v1/documents/upload \
  -F "file=@document.pdf"
```

### Analyze Document

```bash
curl -X POST http://localhost:8080/api/v1/documents/{id}/analyze
```

### Get Document

```bash
curl http://localhost:8080/api/v1/documents/{id}
```