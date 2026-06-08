package routes

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

func GenerateEntityRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.Get("/parent_entities", func(w http.ResponseWriter, r *http.Request) {
		utils.ListFromQueryToResponse(service.ListParentEntities, r, w)
	})
	mainRouter.Get("/entity_by_slug/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		slug, err := url.QueryUnescape(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		utils.ObjectFromQueryToResponse(service.GetEntityBySlug, r, w, slug)
	})
	mainRouter.Get("/products_in_entity/{id}", func(w http.ResponseWriter, r *http.Request) {
		stringId := chi.URLParam(r, "id")
		utils.ListFromQueryToResponseById(
			service.ProductsInEntity,
			r,
			w,
			stringId,
		)
	})

	mainRouter.With(middlewares.AdminOrReadOnly).Route("/entities", func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			utils.ListFromQueryToResponse(service.ListEntities, r, w)
		})

		router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			stringId := chi.URLParam(r, "id")
			utils.ObjectFromQueryToResponse(service.GetEntity, r, w, stringId)
		})

		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			Entity, err := utils.DecodeBody[services.Entity](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.CreateEntity(Entity)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})

		router.Patch("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			Entity, err := utils.DecodeBody[services.Entity](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.EditEntity(id, Entity)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			err := service.DeleteEntity(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		})
	})
}
