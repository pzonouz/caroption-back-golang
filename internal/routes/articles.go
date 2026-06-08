package routes

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

func GenerateArticleRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.Get("/recently_added_articles", func(w http.ResponseWriter, r *http.Request) {
		utils.ListFromQueryToResponse(service.RecentlyArticles, r, w)
	})
	mainRouter.Get("/article_by_slug/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		slug, err := url.QueryUnescape(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		utils.ObjectFromQueryToResponse(service.GetArticleBySlug, r, w, slug)
	})
	mainRouter.With(middlewares.AdminOrReadOnly).Route("/articles", func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			utils.ListFromQueryToResponse(service.ListArticles, r, w)
		})

		router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			utils.ObjectFromQueryToResponse(service.GetArticle, r, w, id)
		})

		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			article, err := utils.DecodeBody[services.Article](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.CreateArticle(article)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Patch("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			article, err := utils.DecodeBody[services.Article](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.EditArticle(id, article)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			err := service.DeleteArticle(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		})
	})
}
