package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cko-recruitment/payment-gateway-challenge-go/docs"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/logging"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// defaultAddr is the address the HTTP server binds to when ADDR is not set.
const defaultAddr = ":8090"

//	@title			Payment Gateway API
//	@version		1.0
//	@description	Processes card payments through an acquiring bank and retrieves previously made payments.
//	@description	A payment ends in one of three states: Authorized, Declined, or Rejected (failed validation, never sent to the bank).

//	@host		localhost:8090
//	@BasePath	/
//	@schemes	http

//	@tag.name			payments
//	@tag.description	Process and retrieve card payments
func main() {
	slog.SetDefault(logging.New(os.Stdout))
	slog.Info("starting payment gateway", "version", version, "commit", commit, "date", date)
	docs.SwaggerInfo.Version = version
	// The @host annotation bakes in a default; override it at runtime so the
	// Swagger UI points at wherever the service is actually reachable.
	if host := os.Getenv("SWAGGER_HOST"); host != "" {
		docs.SwaggerInfo.Host = host
	}

	if err := run(); err != nil {
		slog.Error("fatal API error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		// graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		slog.Info("received shutdown signal")
		cancel()
	}()

	defer func() {
		// recover after panic
		if x := recover(); x != nil {
			slog.Error("run time panic", "panic", x)
			panic(x)
		}
	}()

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	api := api.New()
	if err := api.Run(ctx, addr); err != nil {
		return err
	}

	return nil
}
