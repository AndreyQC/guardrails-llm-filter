package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cloud-ru-tech/guardrails-llm-filter/internal/app"
	"github.com/cloud-ru-tech/guardrails-llm-filter/internal/config"
	"github.com/cloud-ru-tech/guardrails-llm-filter/internal/logging"
)

func main() {
	// Healthcheck
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := app.RunSelfCheck(selfCheckAddr()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		logging.Fatal(ctx, "Failed to load config", err)
	}

	logging.Setup(cfg.LogLevel, cfg.LogFormat)

	logging.Info(ctx, "Starting guardrails")
	application := app.New(cfg)
	if err := application.Start(ctx); err != nil {
		logging.Fatal(ctx, "Failed to start app", err)
	}

	<-ctx.Done()
	stop()

	logging.Info(context.Background(), "Shutting down")
	if err := application.Stop(); err != nil {
		logging.Error(context.Background(), "Shutdown error", err)
	}
}

// selfCheckAddr resolves the probe target from the same environment variable
// as the data-plane listener, falling back to the config default ":8080".
func selfCheckAddr() string {
	if addr := strings.TrimSpace(os.Getenv("GUARDRAILS_LISTEN_ADDR")); addr != "" {
		return addr
	}
	return ":8080"
}
