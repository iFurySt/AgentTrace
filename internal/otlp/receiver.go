package otlp

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"strings"

	collectortracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/iFurySt/AgentTrace/internal/config"
	"github.com/iFurySt/AgentTrace/internal/store"
)

type Receiver struct {
	collectortracepb.UnimplementedTraceServiceServer

	DB             *store.DB
	DefaultProject string
	Logger         *slog.Logger
}

func (r *Receiver) RegisterHTTP(mux *http.ServeMux) {
	mux.HandleFunc("/v1/traces", r.handleHTTPTraces)
}

func (r *Receiver) ServeGRPC(ctx context.Context, addr string) error {
	if strings.TrimSpace(addr) == "" || strings.EqualFold(addr, "off") {
		return nil
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	collectortracepb.RegisterTraceServiceServer(server, r)
	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()
	r.log().Info("otlp grpc receiver listening", "addr", addr)
	err = server.Serve(listener)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

func (r *Receiver) Export(ctx context.Context, req *collectortracepb.ExportTraceServiceRequest) (*collectortracepb.ExportTraceServiceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "missing request")
	}
	if err := r.ingest(ctx, req.GetResourceSpans(), ""); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &collectortracepb.ExportTraceServiceResponse{}, nil
}

func (r *Receiver) handleHTTPTraces(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/x-protobuf" {
		http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
		return
	}
	body, err := readEncodedBody(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}
	var exportReq collectortracepb.ExportTraceServiceRequest
	if err := proto.Unmarshal(body, &exportReq); err != nil {
		http.Error(w, "request body is invalid ExportTraceServiceRequest", http.StatusUnprocessableEntity)
		return
	}
	if err := r.ingest(req.Context(), exportReq.GetResourceSpans(), req.Header.Get("x-project-name")); err != nil {
		r.log().Error("failed to ingest otlp traces", "error", err)
		http.Error(w, "failed to ingest traces", http.StatusInternalServerError)
		return
	}
	resp, _ := proto.Marshal(&collectortracepb.ExportTraceServiceResponse{})
	w.Header().Set("Content-Type", "application/x-protobuf")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

func (r *Receiver) ingest(ctx context.Context, resourceSpans []*tracepb.ResourceSpans, projectHeader string) error {
	defaultProject := strings.TrimSpace(projectHeader)
	if defaultProject == "" {
		defaultProject = r.defaultProject()
	}
	total := 0
	for _, rs := range resourceSpans {
		projectName := strings.TrimSpace(projectHeader)
		if projectName == "" {
			projectName = ProjectNameFromResource(rs.GetResource())
		}
		if projectName == "" {
			projectName = defaultProject
		}
		spans := DecodeResourceSpans(rs)
		count, err := r.DB.Ingest(ctx, projectName, spans)
		if err != nil {
			return err
		}
		total += count
	}
	r.log().Info("ingested otlp spans", "count", total)
	return nil
}

func readEncodedBody(req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(strings.TrimSpace(req.Header.Get("Content-Encoding"))) {
	case "":
		return body, nil
	case "gzip":
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	case "deflate":
		reader, err := zlib.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return io.ReadAll(reader)
	default:
		return nil, errors.New("unsupported content encoding")
	}
}

func (r *Receiver) defaultProject() string {
	if strings.TrimSpace(r.DefaultProject) != "" {
		return r.DefaultProject
	}
	return config.DefaultProjectName
}

func (r *Receiver) log() *slog.Logger {
	if r.Logger != nil {
		return r.Logger
	}
	return slog.Default()
}
