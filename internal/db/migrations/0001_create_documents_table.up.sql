CREATE TABLE IF NOT EXISTS documents (
    id VARCHAR(255) PRIMARY KEY,
    filename VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    s3_key VARCHAR(1000) NOT NULL,
    extracted_text TEXT,
    summary TEXT,
    document_type VARCHAR(100),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    analyzed_at TIMESTAMP
);

CREATE INDEX idx_documents_created_at ON documents(created_at);
CREATE INDEX idx_documents_document_type ON documents(document_type);
CREATE INDEX idx_documents_analyzed_at ON documents(analyzed_at);