-- AI Knowledge App Database Schema
-- Generated from GORM models
-- Requires PostgreSQL with pgvector extension

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Knowledge table - 知识库主表
CREATE TABLE IF NOT EXISTS knowledges (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    summary TEXT,
    is_published BOOLEAN DEFAULT true,
    view_count INTEGER DEFAULT 0,
    content_vector vector(1536),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index for soft delete
CREATE INDEX IF NOT EXISTS idx_knowledges_deleted_at ON knowledges(deleted_at);

-- Tags table - 标签表
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    color VARCHAR(7),
    usage_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index for soft delete
CREATE INDEX IF NOT EXISTS idx_tags_deleted_at ON tags(deleted_at);

-- Query History table - AI查询历史表
CREATE TABLE IF NOT EXISTS query_histories (
    id SERIAL PRIMARY KEY,
    query TEXT NOT NULL,
    response TEXT,
    knowledge_id INTEGER,
    model VARCHAR(50),
    tokens INTEGER DEFAULT 0,
    duration INTEGER DEFAULT 0, -- 毫秒
    is_success BOOLEAN DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (knowledge_id) REFERENCES knowledges(id)
);

-- Create indexes for query history
CREATE INDEX IF NOT EXISTS idx_query_histories_deleted_at ON query_histories(deleted_at);
CREATE INDEX IF NOT EXISTS idx_query_histories_knowledge_id ON query_histories(knowledge_id);

-- Knowledge Tag junction table - 知识标签关联表
CREATE TABLE IF NOT EXISTS knowledge_tags (
    knowledge_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (knowledge_id, tag_id),
    FOREIGN KEY (knowledge_id) REFERENCES knowledges(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Documents table - 文档表
CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    original_name VARCHAR(255),
    file_name VARCHAR(255),
    file_type VARCHAR(100),
    file_path VARCHAR(500), -- Stores S3 object key for S3-compatible storage
    file_size BIGINT,
    file_hash VARCHAR(255),
    mime_type VARCHAR(100),
    extension VARCHAR(20),
    description TEXT,
    status VARCHAR(50) DEFAULT 'completed',
    raw_text TEXT,
    cleaned_text TEXT,
    chunk_count INTEGER,
    error TEXT,
    ref_count INTEGER DEFAULT 1, -- Reference counting for deduplication
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Upload Sessions table - 上传会话表
CREATE TABLE IF NOT EXISTS upload_sessions (
    id VARCHAR(255) PRIMARY KEY,
    file_name VARCHAR(255),
    file_size BIGINT,
    file_hash VARCHAR(255),
    chunk_size BIGINT,
    total_chunks INTEGER,
    uploaded_size BIGINT,
    temp_dir VARCHAR(500),
    upload_id VARCHAR(255), -- MinIO multipart upload ID for S3-compatible storage
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Document Chunks table - 文档分块表
CREATE TABLE IF NOT EXISTS document_chunks (
    id VARCHAR(36) PRIMARY KEY,
    document_id VARCHAR(36) NOT NULL,
    content TEXT NOT NULL,
    chunk_index INTEGER NOT NULL,
    start_offset INTEGER NOT NULL,
    end_offset INTEGER NOT NULL,
    metadata TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for document chunks
CREATE INDEX IF NOT EXISTS idx_document_chunks_document_id ON document_chunks(document_id);
CREATE INDEX IF NOT EXISTS idx_document_chunks_chunk_index ON document_chunks(chunk_index);

-- Document Processing Status table - 文档处理状态表
CREATE TABLE IF NOT EXISTS document_processing_status (
    id VARCHAR(36) PRIMARY KEY,
    document_id VARCHAR(36) NOT NULL UNIQUE,
    preprocess_status VARCHAR(20) NOT NULL,
    vectorization_status VARCHAR(20) NOT NULL DEFAULT 'not_started',
    progress DECIMAL(5,2) DEFAULT 0.00,
    vectorization_progress DECIMAL(5,2) DEFAULT 0.00,
    error_message TEXT,
    vectorization_error TEXT,
    processing_options TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for document processing status
CREATE UNIQUE INDEX IF NOT EXISTS idx_document_processing_status_document_id ON document_processing_status(document_id);
CREATE INDEX IF NOT EXISTS idx_document_processing_status_preprocess_status ON document_processing_status(preprocess_status);

-- Document Embeddings table - 文档嵌入表（预留）
CREATE TABLE IF NOT EXISTS document_embeddings (
    id VARCHAR(36) PRIMARY KEY,
    chunk_id VARCHAR(36) NOT NULL,
    vector_data TEXT, -- 简化为text类型存储
    model_name VARCHAR(100) NOT NULL,
    dimensions INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for document embeddings
CREATE INDEX IF NOT EXISTS idx_document_embeddings_chunk_id ON document_embeddings(chunk_id);
CREATE INDEX IF NOT EXISTS idx_document_embeddings_model_name ON document_embeddings(model_name);

-- Create additional useful indexes
CREATE INDEX IF NOT EXISTS idx_knowledges_title ON knowledges(title);
CREATE INDEX IF NOT EXISTS idx_knowledges_is_published ON knowledges(is_published);
CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_documents_file_hash ON documents(file_hash);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
CREATE INDEX IF NOT EXISTS idx_upload_sessions_expires_at ON upload_sessions(expires_at);

-- Create vector similarity search index (using HNSW for better performance)
CREATE INDEX IF NOT EXISTS idx_knowledges_content_vector ON knowledges 
USING hnsw (content_vector vector_cosine_ops);

-- Update triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update triggers to tables with updated_at columns
CREATE TRIGGER update_knowledges_updated_at BEFORE UPDATE ON knowledges 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tags_updated_at BEFORE UPDATE ON tags 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_query_histories_updated_at BEFORE UPDATE ON query_histories 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_upload_sessions_updated_at BEFORE UPDATE ON upload_sessions 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_document_chunks_updated_at BEFORE UPDATE ON document_chunks 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_document_processing_status_updated_at BEFORE UPDATE ON document_processing_status 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_document_embeddings_updated_at BEFORE UPDATE ON document_embeddings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();