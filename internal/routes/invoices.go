package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

func GenerateInvoiceRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.With(middlewares.AdminOnly).Route("/invoices", func(router chi.Router) {
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			service.ListInvoicesWithSortFilterPagination(
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
			utils.ObjectFromQueryToResponse(service.GetInvoice, r, w, stringId)
		})

		router.Post("/", func(w http.ResponseWriter, r *http.Request) {
			invoice, err := utils.DecodeBody[services.Invoice](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			invoiceNumber, voucherNumber, invoiceId, err := service.CreateInvoice(
				invoice,
				r.Context(),
			)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			type InvoiceCreateResualt struct {
				InvoiceNumber string `json:"invoiceNumber"`
				InvoiceId     string `json:"invoiceId"`
				VoucherNumber string `json:"voucherNumber"`
			}

			var invoiceCreateResualt InvoiceCreateResualt

			invoiceCreateResualt.InvoiceNumber = invoiceNumber
			invoiceCreateResualt.InvoiceId = invoiceId
			invoiceCreateResualt.VoucherNumber = voucherNumber

			err = json.NewEncoder(w).Encode(invoiceCreateResualt)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})

		router.Patch("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			invoice, err := utils.DecodeBody[services.Invoice](r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}

			err = service.EditInvoice(id, invoice)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)

				return
			}
		})
		router.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")

			err := service.DeleteInvoice(id, r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		})
	})
}
