package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

func GenerateAvvalDorehEditRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.With(middlewares.AdminOrReadOnly).
		Route("/avval_doreh_edit", func(router chi.Router) {
			router.Post(
				"/product/{id}",
				func(w http.ResponseWriter, r *http.Request) {
					type AvvalDorehEditProductType struct {
						Count    string `json:"count"`
						BuyPrice string `json:"buyPrice"`
					}

					id := chi.URLParam(r, "id")

					var avvalDorehEditProduct AvvalDorehEditProductType
					json.NewDecoder(r.Body).Decode(&avvalDorehEditProduct)

					err := service.AvvalDorehProductEdit(
						r.Context(),
						id,
						avvalDorehEditProduct.Count,
						avvalDorehEditProduct.BuyPrice,
					)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
					}
				},
			)
			router.Post(
				"/person/{id}",
				func(w http.ResponseWriter, r *http.Request) {
					type AvvalDorehEditPersonType struct {
						Debit  string `json:"debit"`
						Credit string `json:"credit"`
					}

					id := chi.URLParam(r, "id")

					var AvvalDorehEditPerson AvvalDorehEditPersonType
					json.NewDecoder(r.Body).Decode(&AvvalDorehEditPerson)

					err := service.AvvalDorehPersonEdit(
						r.Context(),
						id,
						AvvalDorehEditPerson.Debit,
						AvvalDorehEditPerson.Credit,
					)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
					}
				},
			)
			router.Post(
				"/bank/{id}",
				func(w http.ResponseWriter, r *http.Request) {
					type AvvalDorehEditBankType struct {
						FirstBalance string `json:"firstBalance"`
					}

					id := chi.URLParam(r, "id")

					var AvvalDorehEditBank AvvalDorehEditBankType
					json.NewDecoder(r.Body).Decode(&AvvalDorehEditBank)

					err := service.AvvalDorehBankEdit(
						r.Context(),
						id,
						AvvalDorehEditBank.FirstBalance,
					)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
					}
				},
			)
			router.Post(
				"/cash",
				func(w http.ResponseWriter, r *http.Request) {
					type AvvalDorehEditCashType struct {
						FirstBalance string `json:"firstBalance"`
						AccountCode  string `json:"accountCode"`
					}

					var AvvalDorehEditCash AvvalDorehEditCashType
					json.NewDecoder(r.Body).Decode(&AvvalDorehEditCash)

					err := service.AvvalDorehCashEdit(
						r.Context(),
						AvvalDorehEditCash.AccountCode,
						AvvalDorehEditCash.FirstBalance,
					)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
					}
				},
			)
		})
}
