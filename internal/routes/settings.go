package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

func GenerateSettingsRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.With(middlewares.AdminOnly).
		Get("/settings", func(w http.ResponseWriter, r *http.Request) {
			utils.ListFromQueryToResponse(service.ListSettings, r, w)
		})
}
