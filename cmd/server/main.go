package main

import (
	"net/http"

	_ "github.com/lib/pq"     
	_ "subscriptions/docs"
	"subscriptions/internal/config"
	"subscriptions/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"
)
// @title Subscriptions API
// @version 1.0
// @description API for managing user subscriptions

// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.Load()

	r := chi.NewRouter()


	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})


	handler.RegisterSubscriptionRoutes(r, cfg)


	r.Get("/swagger/*", httpSwagger.WrapHandler)

	log.Info().Msg("server started at :" + cfg.Port)

	http.ListenAndServe(":"+cfg.Port, r)
}
