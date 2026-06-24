package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

type pong struct {
	Message string `json:"message"`
}

// PingHandler returns an http.HandlerFunc that handles HTTP Ping GET requests.
//
// @Produce	json
// @Success	200	{object}	pong
// @Router	/ping [get]
func (a *Api) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(pong{Message: "pong"}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// SwaggerHandler returns an http.HandlerFunc that handles HTTP Swagger related requests.
func (a *Api) SwaggerHandler() http.HandlerFunc {
	return httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s/swagger/doc.json", docs.SwaggerInfo.Host)),
		// Render request/response bodies as the annotated schema (with field
		// descriptions and validation constraints) by default, rather than the
		// bare example JSON, and expand it so the fields are visible without extra
		// clicks. The standalone Models section is hidden (-1); schemas are shown
		// inline on each operation instead.
		httpSwagger.UIConfig(map[string]string{
			"defaultModelRendering":    `"model"`,
			"defaultModelExpandDepth":  "3",
			"defaultModelsExpandDepth": "-1",
		}),
	)
}
