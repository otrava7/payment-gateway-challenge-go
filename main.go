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

//	@title			Payment Gateway Challenge Go
//	@description	Interview challenge for building a Payment Gateway - Go version

//	@host		localhost:8090
//	@BasePath	/

// @securityDefinitions.basic	BasicAuth
func main() {
	slog.SetDefault(logging.New(os.Stdout))
	slog.Info("starting payment gateway", "version", version, "commit", commit, "date", date)
	docs.SwaggerInfo.Version = version

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

	api := api.New()
	if err := api.Run(ctx, ":8090"); err != nil {
		return err
	}

	return nil
}
