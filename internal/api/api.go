package api

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/sync/errgroup"
)

// PaymentsService describes the business-logic operations the HTTP layer
// depends on. Defining it here (consumer side) lets the API be tested against a
// mock and lets the concrete implementation grow — e.g. an HTTP client to an
// acquiring bank for processing payments — without changing this package.
type PaymentsService interface {
	GetPayment(id string) *models.PaymentResponse
}

type Api struct {
	router          *chi.Mux
	paymentsService PaymentsService
}

func New() *Api {
	a := &Api{}
	a.paymentsService = service.NewPaymentsService()
	a.setupRouter()

	return a
}

func (a *Api) Run(ctx context.Context, addr string) error {
	httpServer := &http.Server{
		Addr:        addr,
		Handler:     a.router,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-ctx.Done()
		fmt.Printf("shutting down HTTP server\n")
		return httpServer.Shutdown(ctx)
	})

	g.Go(func() error {
		fmt.Printf("starting HTTP server on %s\n", addr)
		err := httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	return g.Wait()
}

func (a *Api) setupRouter() {
	a.router = chi.NewRouter()
	a.router.Use(middleware.Logger)

	a.router.Get("/ping", a.PingHandler())
	a.router.Get("/swagger/*", a.SwaggerHandler())

	a.router.Get("/api/payments/{id}", a.GetPaymentHandler())
}
