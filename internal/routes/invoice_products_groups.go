package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

func GenerateInvoiceProductsGroupsRoutes(mainRouter *chi.Mux, service services.Service) {

	mainRouter.With(middlewares.AdminOrReadOnly).Route("/invoice_products_groups", func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			utils.ListFromQueryToResponse(service.ListInvoiceProductsGroups, r, w)
		})
		router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			stringId := chi.URLParam(r, "id")
			utils.ObjectFromQueryToResponse(service.GetInvoiceProductsGroup, r, w, stringId)
		})

		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			invoiceProductsGroup, err := utils.DecodeBody[services.InvoiceProductsGroup](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.CreateInvoiceProductsGroup(invoiceProductsGroup, r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Patch("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			invoiceProductsGroup, err := utils.DecodeBody[services.InvoiceProductsGroup](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.EditInvoiceProductsGroup(id, invoiceProductsGroup)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			err := service.DeleteInvoiceProductsGroup(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		})
	})
}
