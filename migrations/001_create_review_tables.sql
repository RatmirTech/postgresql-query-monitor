-- SQL Query Review Storage Schema
-- This script creates tables to store SQL query review results

-- Main table to store review requests and responses
CREATE TABLE IF NOT EXISTS query_reviews (
    id BIGSERIAL PRIMARY KEY,
    
    -- Source information
    source_database VARCHAR(255) NOT NULL,
    environment VARCHAR(100) NOT NULL DEFAULT 'production',
    
    -- Request data
    sql_query TEXT NOT NULL,
    query_plan TEXT,
    request_json JSONB NOT NULL,
    
    -- Response data  
    response_json JSONB NOT NULL,
    thread_id UUID,
    overall_score INTEGER,
    notes TEXT,
    errors TEXT[],
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Table to store table structure information
CREATE TABLE IF NOT EXISTS table_structures (
    id BIGSERIAL PRIMARY KEY,
    review_id BIGINT REFERENCES query_reviews(id) ON DELETE CASCADE,
    
    table_name VARCHAR(255) NOT NULL,
    columns_info JSONB NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_query_reviews_source_db ON query_reviews(source_database);
CREATE INDEX IF NOT EXISTS idx_query_reviews_environment ON query_reviews(environment);
CREATE INDEX IF NOT EXISTS idx_query_reviews_created_at ON query_reviews(created_at);
CREATE INDEX IF NOT EXISTS idx_query_reviews_thread_id ON query_reviews(thread_id);
CREATE INDEX IF NOT EXISTS idx_query_reviews_overall_score ON query_reviews(overall_score);

-- GIN index for JSON columns for fast search
CREATE INDEX IF NOT EXISTS idx_query_reviews_request_json ON query_reviews USING GIN(request_json);
CREATE INDEX IF NOT EXISTS idx_query_reviews_response_json ON query_reviews USING GIN(response_json);

-- Index for table structures
CREATE INDEX IF NOT EXISTS idx_table_structures_review_id ON table_structures(review_id);
CREATE INDEX IF NOT EXISTS idx_table_structures_table_name ON table_structures(table_name);

-- Comment on tables
COMMENT ON TABLE query_reviews IS 'Stores SQL query review requests and responses from external review API';
COMMENT ON TABLE table_structures IS 'Stores table structure information for reviewed queries';

-- Comment on important columns
COMMENT ON COLUMN query_reviews.source_database IS 'Name of the database that was monitored';
COMMENT ON COLUMN query_reviews.sql_query IS 'The SQL query that was reviewed';
COMMENT ON COLUMN query_reviews.request_json IS 'Complete JSON request sent to review API';
COMMENT ON COLUMN query_reviews.response_json IS 'Complete JSON response received from review API';
COMMENT ON COLUMN query_reviews.thread_id IS 'Review thread ID from API response';
COMMENT ON COLUMN query_reviews.overall_score IS 'Overall score from review API (0-100)';
