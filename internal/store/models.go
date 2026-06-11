package store

import (
	"time"

	"gorm.io/datatypes"
)

type Project struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null;uniqueIndex"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Trace struct {
	ID             uint      `gorm:"primaryKey"`
	ProjectID      uint      `gorm:"not null;index;uniqueIndex:idx_traces_project_trace"`
	TraceID        string    `gorm:"not null;index;uniqueIndex:idx_traces_project_trace"`
	SessionID      string    `gorm:"index"`
	StartTime      time.Time `gorm:"index"`
	EndTime        time.Time
	SpanCount      int
	ErrorCount     int
	InputTokens    int
	OutputTokens   int
	DurationMillis int64 `gorm:"index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Span struct {
	ID                 uint   `gorm:"primaryKey"`
	ProjectID          uint   `gorm:"not null;index"`
	TraceRowID         uint   `gorm:"not null;index"`
	OTelTraceID        string `gorm:"column:otel_trace_id;not null;index;uniqueIndex:idx_spans_trace_span"`
	SpanID             string `gorm:"not null;index;uniqueIndex:idx_spans_trace_span"`
	ParentSpanID       string `gorm:"index"`
	Name               string `gorm:"not null;index"`
	OTelSpanKind       string `gorm:"column:otel_span_kind;index"`
	StatusCode         string `gorm:"not null;index"`
	StatusMessage      string
	StartTime          time.Time `gorm:"index"`
	EndTime            time.Time
	DurationMillis     int64 `gorm:"index"`
	ScopeName          string
	ScopeVersion       string
	ResourceAttributes datatypes.JSON
	Attributes         datatypes.JSON
	Events             datatypes.JSON
	GenAIOperationName string `gorm:"index"`
	GenAIProviderName  string `gorm:"column:gen_ai_provider_name;index"`
	GenAIRequestModel  string `gorm:"column:gen_ai_request_model;index"`
	GenAIResponseModel string `gorm:"column:gen_ai_response_model;index"`
	InputTokens        int
	OutputTokens       int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type IngestSpan struct {
	TraceID            string
	SpanID             string
	ParentSpanID       string
	Name               string
	OTelSpanKind       string
	StatusCode         string
	StatusMessage      string
	StartTime          time.Time
	EndTime            time.Time
	ScopeName          string
	ScopeVersion       string
	ResourceAttributes map[string]any
	Attributes         map[string]any
	Events             []map[string]any
}
