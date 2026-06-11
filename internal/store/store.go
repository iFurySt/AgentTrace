package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type DB struct {
	gorm *gorm.DB
}

func Open(driver, dsn string) (*DB, error) {
	driver = strings.ToLower(strings.TrimSpace(driver))
	switch driver {
	case "", "sqlite", "sqlite3":
		if err := ensureSQLiteDir(dsn); err != nil {
			return nil, err
		}
		return open(sqlite.Open(dsn))
	case "postgres", "postgresql":
		return open(postgres.Open(dsn))
	default:
		return nil, errors.New("unsupported database driver: " + driver)
	}
}

func open(dialector gorm.Dialector) (*DB, error) {
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Project{}, &Trace{}, &Span{}); err != nil {
		return nil, err
	}
	return &DB{gorm: db}, nil
}

func ensureSQLiteDir(dsn string) error {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" || dsn == ":memory:" || strings.HasPrefix(dsn, "file:") {
		return nil
	}
	return os.MkdirAll(filepath.Dir(dsn), 0o755)
}

func (db *DB) Close() error {
	if db == nil || db.gorm == nil {
		return nil
	}
	sqlDB, err := db.gorm.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	sqlDB, err := db.gorm.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (db *DB) Ingest(ctx context.Context, projectName string, spans []IngestSpan) (int, error) {
	projectName = strings.TrimSpace(projectName)
	if projectName == "" {
		projectName = "default"
	}
	if len(spans) == 0 {
		return 0, nil
	}
	err := db.gorm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var project Project
		if err := tx.Where("name = ?", projectName).FirstOrCreate(&project, Project{Name: projectName}).Error; err != nil {
			return err
		}
		for _, ingest := range spans {
			if strings.TrimSpace(ingest.TraceID) == "" || strings.TrimSpace(ingest.SpanID) == "" {
				continue
			}
			trace, err := upsertTrace(tx, project.ID, ingest)
			if err != nil {
				return err
			}
			row, err := buildSpan(project.ID, trace.ID, ingest)
			if err != nil {
				return err
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "otel_trace_id"}, {Name: "span_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"project_id", "trace_row_id", "parent_span_id", "name",
					"otel_span_kind", "status_code", "status_message", "start_time", "end_time",
					"duration_millis", "scope_name", "scope_version", "resource_attributes",
					"attributes", "events", "gen_ai_operation_name",
					"gen_ai_provider_name", "gen_ai_request_model", "gen_ai_response_model",
					"input_tokens", "output_tokens", "updated_at",
				}),
			}).Create(&row).Error; err != nil {
				return err
			}
			if err := refreshTraceSummary(tx, trace.ID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(spans), nil
}

func upsertTrace(tx *gorm.DB, projectID uint, ingest IngestSpan) (*Trace, error) {
	sessionID := firstString(ingest.Attributes, "gen_ai.conversation.id")
	var trace Trace
	err := tx.Where("project_id = ? AND trace_id = ?", projectID, ingest.TraceID).First(&trace).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		trace = Trace{
			ProjectID: projectID,
			TraceID:   ingest.TraceID,
			SessionID: sessionID,
			StartTime: ingest.StartTime,
			EndTime:   ingest.EndTime,
		}
		trace.DurationMillis = millis(trace.StartTime, trace.EndTime)
		return &trace, tx.Create(&trace).Error
	}
	if err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if sessionID != "" && trace.SessionID == "" {
		updates["session_id"] = sessionID
	}
	if !ingest.StartTime.IsZero() && (trace.StartTime.IsZero() || ingest.StartTime.Before(trace.StartTime)) {
		updates["start_time"] = ingest.StartTime
		trace.StartTime = ingest.StartTime
	}
	if !ingest.EndTime.IsZero() && ingest.EndTime.After(trace.EndTime) {
		updates["end_time"] = ingest.EndTime
		trace.EndTime = ingest.EndTime
	}
	if len(updates) > 0 {
		updates["duration_millis"] = millis(trace.StartTime, trace.EndTime)
		if err := tx.Model(&trace).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return &trace, nil
}

func buildSpan(projectID, traceRowID uint, ingest IngestSpan) (Span, error) {
	attrs := ingest.Attributes
	attrJSON, err := marshalJSON(attrs)
	if err != nil {
		return Span{}, err
	}
	resourceJSON, err := marshalJSON(ingest.ResourceAttributes)
	if err != nil {
		return Span{}, err
	}
	eventsJSON, err := marshalJSON(ingest.Events)
	if err != nil {
		return Span{}, err
	}
	inputTokens := firstInt(attrs, "gen_ai.usage.input_tokens")
	outputTokens := firstInt(attrs, "gen_ai.usage.output_tokens")
	return Span{
		ProjectID:          projectID,
		TraceRowID:         traceRowID,
		OTelTraceID:        ingest.TraceID,
		SpanID:             ingest.SpanID,
		ParentSpanID:       ingest.ParentSpanID,
		Name:               ingest.Name,
		OTelSpanKind:       ingest.OTelSpanKind,
		StatusCode:         normalizeStatus(ingest.StatusCode),
		StatusMessage:      ingest.StatusMessage,
		StartTime:          ingest.StartTime,
		EndTime:            ingest.EndTime,
		DurationMillis:     millis(ingest.StartTime, ingest.EndTime),
		ScopeName:          ingest.ScopeName,
		ScopeVersion:       ingest.ScopeVersion,
		ResourceAttributes: resourceJSON,
		Attributes:         attrJSON,
		Events:             eventsJSON,
		GenAIOperationName: firstString(attrs, "gen_ai.operation.name"),
		GenAIProviderName:  firstString(attrs, "gen_ai.provider.name"),
		GenAIRequestModel:  firstString(attrs, "gen_ai.request.model"),
		GenAIResponseModel: firstString(attrs, "gen_ai.response.model"),
		InputTokens:        inputTokens,
		OutputTokens:       outputTokens,
	}, nil
}

func refreshTraceSummary(tx *gorm.DB, traceRowID uint) error {
	var spans []Span
	if err := tx.Where("trace_row_id = ?", traceRowID).Find(&spans).Error; err != nil {
		return err
	}
	var start, end time.Time
	var errorsCount, inputTokens, outputTokens int
	for _, span := range spans {
		if start.IsZero() || (!span.StartTime.IsZero() && span.StartTime.Before(start)) {
			start = span.StartTime
		}
		if span.EndTime.After(end) {
			end = span.EndTime
		}
		if span.StatusCode == "ERROR" {
			errorsCount++
		}
		inputTokens += span.InputTokens
		outputTokens += span.OutputTokens
	}
	return tx.Model(&Trace{}).Where("id = ?", traceRowID).Updates(map[string]any{
		"start_time":      start,
		"end_time":        end,
		"span_count":      len(spans),
		"error_count":     errorsCount,
		"input_tokens":    inputTokens,
		"output_tokens":   outputTokens,
		"duration_millis": millis(start, end),
	}).Error
}

func marshalJSON(value any) (datatypes.JSON, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(data), nil
}

func millis(start, end time.Time) int64 {
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return 0
	}
	return end.Sub(start).Milliseconds()
}

func normalizeStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "OK", "ERROR":
		return strings.ToUpper(strings.TrimSpace(status))
	default:
		return "UNSET"
	}
}

func firstString(attrs map[string]any, keys ...string) string {
	for _, key := range keys {
		switch value := attrs[key].(type) {
		case string:
			if strings.TrimSpace(value) != "" {
				return value
			}
		case []byte:
			if strings.TrimSpace(string(value)) != "" {
				return string(value)
			}
		}
	}
	return ""
}

func firstInt(attrs map[string]any, keys ...string) int {
	for _, key := range keys {
		switch value := attrs[key].(type) {
		case int:
			return value
		case int64:
			return int(value)
		case float64:
			return int(value)
		case float32:
			return int(value)
		case json.Number:
			i, _ := value.Int64()
			return int(i)
		}
	}
	return 0
}
