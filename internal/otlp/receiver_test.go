package otlp_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"

	"github.com/iFurySt/AgentTrace/internal/httpapi"
	"github.com/iFurySt/AgentTrace/internal/otlp"
	"github.com/iFurySt/AgentTrace/internal/store"
)

func TestHTTPReceiverIngestsGenAISpan(t *testing.T) {
	db, err := store.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mux := http.NewServeMux()
	httpapi.API{DB: db}.Register(mux)
	(&otlp.Receiver{DB: db, DefaultProject: "default"}).RegisterHTTP(mux)

	payload, err := proto.Marshal(sampleExportRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	var gzipped bytes.Buffer
	gzipWriter := gzip.NewWriter(&gzipped)
	if _, err := gzipWriter.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/traces", &gzipped)
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "application/x-protobuf" {
		t.Fatalf("content-type = %q", got)
	}

	projects, err := db.Projects(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 || projects[0].Name != "heyyod" {
		t.Fatalf("projects = %+v, want heyyod", projects)
	}

	traces, err := db.Traces(context.Background(), "heyyod", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(traces) != 1 {
		t.Fatalf("trace count = %d, want 1", len(traces))
	}
	if traces[0].SpanCount != 1 || traces[0].InputTokens != 12 || traces[0].OutputTokens != 8 {
		t.Fatalf("trace summary = %+v", traces[0])
	}

	spans, err := db.Spans(context.Background(), traces[0].TraceID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("span count = %d, want 1", len(spans))
	}
	span := spans[0]
	if span.GenAIOperationName != "audio_transcription" {
		t.Fatalf("gen ai operation = %q", span.GenAIOperationName)
	}
	if span.GenAIProviderName != "gemini" {
		t.Fatalf("provider = %q", span.GenAIProviderName)
	}
	if span.OpenInferenceSpanKind != "LLM" {
		t.Fatalf("openinference span kind = %q", span.OpenInferenceSpanKind)
	}
	if span.GenAIRequestModel != "gemini-2.5-flash" || span.InputTokens != 12 || span.OutputTokens != 8 {
		t.Fatalf("gen ai indexed fields = %+v", span)
	}
}

func sampleExportRequest(t *testing.T) *collectortracepb.ExportTraceServiceRequest {
	t.Helper()
	traceID, err := hex.DecodeString("00112233445566778899aabbccddeeff")
	if err != nil {
		t.Fatal(err)
	}
	spanID, err := hex.DecodeString("0011223344556677")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 6, 11, 10, 0, 0, 0, time.UTC)
	end := start.Add(150 * time.Millisecond)
	return &collectortracepb.ExportTraceServiceRequest{
		ResourceSpans: []*tracepb.ResourceSpans{
			{
				Resource: &resourcepb.Resource{Attributes: []*commonpb.KeyValue{
					kv("openinference.project.name", stringValue("heyyod")),
					kv("service.name", stringValue("heyyod")),
				}},
				ScopeSpans: []*tracepb.ScopeSpans{
					{
						Scope: &commonpb.InstrumentationScope{Name: "heyyod-test", Version: "v0"},
						Spans: []*tracepb.Span{
							{
								TraceId:           traceID,
								SpanId:            spanID,
								Name:              "audio_transcription gemini-2.5-flash",
								Kind:              tracepb.Span_SPAN_KIND_CLIENT,
								StartTimeUnixNano: uint64(start.UnixNano()),
								EndTimeUnixNano:   uint64(end.UnixNano()),
								Status:            &tracepb.Status{Code: tracepb.Status_STATUS_CODE_OK},
								Attributes: []*commonpb.KeyValue{
									kv("gen_ai.operation.name", stringValue("audio_transcription")),
									kv("gen_ai.provider.name", stringValue("gemini")),
									kv("gen_ai.request.model", stringValue("gemini-2.5-flash")),
									kv("gen_ai.response.model", stringValue("gemini-2.5-flash")),
									kv("gen_ai.usage.input_tokens", intValue(12)),
									kv("gen_ai.usage.output_tokens", intValue(8)),
									kv("gen_ai.conversation.id", stringValue("session-1")),
									kv("input.value", stringValue("transcribe this audio")),
									kv("output.value", stringValue("hello world")),
								},
							},
						},
					},
				},
			},
		},
	}
}

func kv(key string, value *commonpb.AnyValue) *commonpb.KeyValue {
	return &commonpb.KeyValue{Key: key, Value: value}
}

func stringValue(value string) *commonpb.AnyValue {
	return &commonpb.AnyValue{Value: &commonpb.AnyValue_StringValue{StringValue: value}}
}

func intValue(value int64) *commonpb.AnyValue {
	return &commonpb.AnyValue{Value: &commonpb.AnyValue_IntValue{IntValue: value}}
}
