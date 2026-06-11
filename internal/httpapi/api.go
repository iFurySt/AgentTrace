package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/iFurySt/AgentTrace/internal/store"
)

type API struct {
	DB *store.DB
}

func (api API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", api.handleHealthz)
	mux.HandleFunc("/api/projects", api.handleProjects)
	mux.HandleFunc("/api/traces/", api.handleTrace)
	mux.HandleFunc("/api/traces", api.handleTraces)
	mux.HandleFunc("/api/spans", api.handleSpans)
}

func (api API) handleHealthz(w http.ResponseWriter, req *http.Request) {
	if err := api.DB.Ping(req.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (api API) handleProjects(w http.ResponseWriter, req *http.Request) {
	projects, err := api.DB.Projects(req.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": projects})
}

func (api API) handleTraces(w http.ResponseWriter, req *http.Request) {
	traces, err := api.DB.Traces(req.Context(), req.URL.Query().Get("project"), limit(req, 100))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": traces})
}

func (api API) handleTrace(w http.ResponseWriter, req *http.Request) {
	traceID := strings.TrimPrefix(req.URL.Path, "/api/traces/")
	if traceID == "" {
		writeError(w, http.StatusNotFound, "trace not found")
		return
	}
	trace, spans, err := api.DB.TraceByID(req.Context(), traceID)
	if store.IsNotFound(err) {
		writeError(w, http.StatusNotFound, "trace not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"trace": trace, "spans": spans}})
}

func (api API) handleSpans(w http.ResponseWriter, req *http.Request) {
	spans, err := api.DB.Spans(req.Context(), req.URL.Query().Get("trace_id"), limit(req, 1000))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": spans})
}

func limit(req *http.Request, fallback int) int {
	value, err := strconv.Atoi(req.URL.Query().Get("limit"))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
