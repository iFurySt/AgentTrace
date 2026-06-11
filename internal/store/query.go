package store

import (
	"context"

	"gorm.io/gorm"
)

func (db *DB) Projects(ctx context.Context) ([]Project, error) {
	var projects []Project
	err := db.gorm.WithContext(ctx).Order("name asc").Find(&projects).Error
	return projects, err
}

func (db *DB) Traces(ctx context.Context, projectName string, limit int) ([]Trace, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	query := db.gorm.WithContext(ctx).Model(&Trace{}).Order("start_time desc").Limit(limit)
	if projectName != "" {
		query = query.Joins("join projects on projects.id = traces.project_id").Where("projects.name = ?", projectName)
	}
	var traces []Trace
	err := query.Find(&traces).Error
	return traces, err
}

func (db *DB) TraceByID(ctx context.Context, traceID string) (*Trace, []Span, error) {
	var trace Trace
	if err := db.gorm.WithContext(ctx).Where("trace_id = ?", traceID).First(&trace).Error; err != nil {
		return nil, nil, err
	}
	spans, err := db.Spans(ctx, traceID, 1000)
	return &trace, spans, err
}

func (db *DB) Spans(ctx context.Context, traceID string, limit int) ([]Span, error) {
	if limit <= 0 || limit > 5000 {
		limit = 1000
	}
	query := db.gorm.WithContext(ctx).Order("start_time asc").Limit(limit)
	if traceID != "" {
		query = query.Where("otel_trace_id = ?", traceID)
	}
	var spans []Span
	err := query.Find(&spans).Error
	return spans, err
}

func IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}
