package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/iFurySt/AgentTrace/internal/config"
	"github.com/iFurySt/AgentTrace/internal/httpapi"
	"github.com/iFurySt/AgentTrace/internal/otlp"
	"github.com/iFurySt/AgentTrace/internal/store"
)

var version = "dev"

func Execute() {
	if err := rootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCommand() *cobra.Command {
	cfg := config.Load()
	cmd := &cobra.Command{
		Use:   "agenttrace",
		Short: "A small OTLP trace store for agent and GenAI telemetry",
	}
	cmd.PersistentFlags().StringVar(&cfg.DatabaseDriver, "database-driver", cfg.DatabaseDriver, "database driver: sqlite or postgres")
	cmd.PersistentFlags().StringVar(&cfg.DatabaseDSN, "database-dsn", cfg.DatabaseDSN, "database DSN or SQLite path")
	cmd.PersistentFlags().StringVar(&cfg.DefaultProject, "default-project", cfg.DefaultProject, "project name used when OTLP data does not set one")

	cmd.AddCommand(serveCommand(&cfg), migrateCommand(&cfg), versionCommand())
	return cmd
}

func serveCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run OTLP receivers and query API",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd.Context(), *cfg)
		},
	}
	cmd.Flags().StringVar(&cfg.HTTPAddr, "http-addr", cfg.HTTPAddr, "HTTP listen address for /v1/traces and query API")
	cmd.Flags().StringVar(&cfg.GRPCAddr, "grpc-addr", cfg.GRPCAddr, "gRPC OTLP listen address, or off")
	return cmd
}

func migrateCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Open the database and run GORM migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := store.Open(cfg.DatabaseDriver, cfg.DatabaseDSN)
			if err != nil {
				return err
			}
			return db.Close()
		},
	}
}

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version)
		},
	}
}

func runServe(ctx context.Context, cfg config.Config) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	db, err := store.Open(cfg.DatabaseDriver, cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	api := httpapi.API{DB: db}
	api.Register(mux)
	receiver := &otlp.Receiver{DB: db, DefaultProject: cfg.DefaultProject, Logger: logger}
	receiver.RegisterHTTP(mux)

	httpServer := &http.Server{Addr: cfg.HTTPAddr, Handler: mux}
	errCh := make(chan error, 2)
	go func() {
		logger.Info("http receiver listening", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	go func() {
		if err := receiver.ServeGRPC(ctx, cfg.GRPCAddr); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
