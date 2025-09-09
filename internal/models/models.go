package models

// Config represents PostgreSQL configuration parameters
type Config struct {
	SharedBuffers              string `json:"shared_buffers"`
	EffectiveCacheSize         string `json:"effective_cache_size"`
	MaintenanceWorkMem         string `json:"maintenance_work_mem"`
	CheckpointCompletionTarget string `json:"checkpoint_completion_target"`
	WalBuffers                 string `json:"wal_buffers"`
	DefaultStatisticsTarget    string `json:"default_statistics_target"`
	RandomPageCost             string `json:"random_page_cost"`
	EffectiveIOConcurrency     string `json:"effective_io_concurrency"`
	WorkMem                    string `json:"work_mem"`
	MinWalSize                 string `json:"min_wal_size"`
	MaxWalSize                 string `json:"max_wal_size"`
}

// ServerInfo represents server information
type ServerInfo struct {
	Version  string
	Host     string
	Database string
}

// ServerData represents the overall server configuration and info
type ServerData struct {
	Config      Config     `json:"config"`
	Environment string     `json:"environment"`
	ServerInfo  ServerInfo `json:"server_info"`
}

// Recommendation represents a configuration recommendation message
type Recommendation struct {
	Content        string `json:"content"`
	Criticality    string `json:"criticality"`
	Recommendation string `json:"recommendation"`
}

// QueryReviewRequest represents a single SQL query review request
type QueryReviewRequest struct {
	SQL         string      `json:"sql"`
	QueryPlan   interface{} `json:"query_plan,omitempty"`
	Tables      []TableInfo `json:"tables,omitempty"`
	ServerInfo  ServerInfo  `json:"server_info,omitempty"`
	ThreadID    string      `json:"thread_id,omitempty"`
	Environment string      `json:"environment,omitempty"`
}

// BatchReviewRequest represents a batch of SQL queries for review
type BatchReviewRequest struct {
	Queries     []QueryReviewRequest `json:"queries"`
	Environment string               `json:"environment,omitempty"`
}

// MigrationReviewRequest represents a migration SQL script for review
type MigrationReviewRequest struct {
	SQL         string `json:"sql"`
	Environment string `json:"environment,omitempty"`
}

// QueryReviewResponse represents the response for a single query review
type QueryReviewResponse struct {
	Score           int      `json:"overall_score"`
	Recommendations []string `json:"recommendations"`
	Issues          []string `json:"issues"`
}

// BatchReviewResponse represents the response for batch query review
type BatchReviewResponse struct {
	Results []QueryReviewResponse `json:"results"`
}

// MigrationReviewResponse represents the response for migration review
type MigrationReviewResponse struct {
	Score           int      `json:"overall_score"`
	Recommendations []string `json:"recommendations"`
	Issues          []string `json:"issues"`
	Warnings        []string `json:"warnings"`
}

// TableInfo represents information about a database table
type TableInfo struct {
	Name     string   `json:"name"`
	Schema   string   `json:"schema,omitempty"`
	RowCount int64    `json:"row_count,omitempty"`
	Indexes  []string `json:"indexes,omitempty"`
}
