package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

func GenerateParametersRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.With(middlewares.AdminOrReadOnly).Route("/parameters", func(router chi.Router) {
		router.Get("/by-category/{id}", func(w http.ResponseWriter, r *http.Request) {
			categoryId := chi.URLParam(r, "id")
			utils.ListFromQueryToResponseById(service.ListParametersByCategory, r, w, categoryId)
		})
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			service.ListParametersWithSortFilterPagination(
				utils.DefaultInput(r.URL.Query().Get("sort"), ""),
				utils.DefaultInput(r.URL.Query().Get("sort_direction"), ""),
				r.URL.Query()["filter"],
				r.URL.Query()["filter_operand"],
				r.URL.Query()["filter_condition"],
				r.URL.Query().Get("count_in_page"),
				r.URL.Query().Get("offset"),
				w,
			)
		})
		router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			stringId := chi.URLParam(r, "id")
			utils.ObjectFromQueryToResponse(service.GetParameter, r, w, stringId)
		})

		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			parameter, err := utils.DecodeBody[services.Parameter](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.CreateParameter(parameter)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Patch("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			parameter, err := utils.DecodeBody[services.Parameter](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.EditParameter(id, parameter)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			err := service.DeleteParameter(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		})
	})
}
