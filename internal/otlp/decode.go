package otlp

import (
	"encoding/hex"
	"time"

	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"

	"github.com/iFurySt/AgentTrace/internal/store"
)

func ProjectNameFromResource(resource *resourcepb.Resource) string {
	attrs := decodeAttributes(resource.GetAttributes())
	for _, key := range []string{"service.name", "service.namespace"} {
		if value, ok := attrs[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func DecodeResourceSpans(resourceSpans *tracepb.ResourceSpans) []store.IngestSpan {
	resourceAttrs := decodeAttributes(resourceSpans.GetResource().GetAttributes())
	var spans []store.IngestSpan
	for _, scopeSpans := range resourceSpans.GetScopeSpans() {
		scope := scopeSpans.GetScope()
		for _, span := range scopeSpans.GetSpans() {
			spans = append(spans, decodeSpan(resourceAttrs, scope.GetName(), scope.GetVersion(), span))
		}
	}
	return spans
}

func decodeSpan(resourceAttrs map[string]any, scopeName, scopeVersion string, span *tracepb.Span) store.IngestSpan {
	return store.IngestSpan{
		TraceID:            hex.EncodeToString(span.GetTraceId()),
		SpanID:             hex.EncodeToString(span.GetSpanId()),
		ParentSpanID:       hex.EncodeToString(span.GetParentSpanId()),
		Name:               span.GetName(),
		OTelSpanKind:       span.GetKind().String(),
		StatusCode:         decodeStatus(span.GetStatus().GetCode()),
		StatusMessage:      span.GetStatus().GetMessage(),
		StartTime:          unixNano(span.GetStartTimeUnixNano()),
		EndTime:            unixNano(span.GetEndTimeUnixNano()),
		ScopeName:          scopeName,
		ScopeVersion:       scopeVersion,
		ResourceAttributes: resourceAttrs,
		Attributes:         decodeAttributes(span.GetAttributes()),
		Events:             decodeEvents(span.GetEvents()),
	}
}

func decodeStatus(code tracepb.Status_StatusCode) string {
	switch code {
	case tracepb.Status_STATUS_CODE_OK:
		return "OK"
	case tracepb.Status_STATUS_CODE_ERROR:
		return "ERROR"
	default:
		return "UNSET"
	}
}

func unixNano(value uint64) time.Time {
	if value == 0 {
		return time.Time{}
	}
	return time.Unix(0, int64(value)).UTC()
}

func decodeEvents(events []*tracepb.Span_Event) []map[string]any {
	out := make([]map[string]any, 0, len(events))
	for _, event := range events {
		out = append(out, map[string]any{
			"name":       event.GetName(),
			"timestamp":  unixNano(event.GetTimeUnixNano()).Format(time.RFC3339Nano),
			"attributes": decodeAttributes(event.GetAttributes()),
		})
	}
	return out
}

func decodeAttributes(attrs []*commonpb.KeyValue) map[string]any {
	out := make(map[string]any, len(attrs))
	for _, kv := range attrs {
		out[kv.GetKey()] = decodeValue(kv.GetValue())
	}
	return out
}

func decodeValue(value *commonpb.AnyValue) any {
	switch v := value.GetValue().(type) {
	case *commonpb.AnyValue_StringValue:
		return v.StringValue
	case *commonpb.AnyValue_BoolValue:
		return v.BoolValue
	case *commonpb.AnyValue_IntValue:
		return v.IntValue
	case *commonpb.AnyValue_DoubleValue:
		return v.DoubleValue
	case *commonpb.AnyValue_BytesValue:
		return v.BytesValue
	case *commonpb.AnyValue_ArrayValue:
		values := v.ArrayValue.GetValues()
		out := make([]any, 0, len(values))
		for _, item := range values {
			out = append(out, decodeValue(item))
		}
		return out
	case *commonpb.AnyValue_KvlistValue:
		return decodeAttributes(v.KvlistValue.GetValues())
	default:
		return nil
	}
}
