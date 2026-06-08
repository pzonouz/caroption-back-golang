package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

func GenerateReportsRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.With(middlewares.AdminOnly).Route("/reports", func(router chi.Router) {
		router.Get(
			"/product_activities_details/{product_id}/{from_date}/{to_date}",
			func(w http.ResponseWriter, r *http.Request) {
				productID := chi.URLParam(r, "product_id")
				fromDate := chi.URLParam(r, "from_date")

				parsedFromDate, err := time.Parse("2006-01-02", fromDate)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				toDate := chi.URLParam(r, "to_date")

				parsedToDate, err := time.Parse("2006-01-02", toDate)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				productActivitiesDetail, err := service.GetProductActivitiesDetail(
					productID,
					r.Context(), parsedFromDate, parsedToDate,
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}

				w.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(w).Encode(productActivitiesDetail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}
			},
		)
		router.Get(
			"/person_activities_details/{person_id}/{from_date}/{to_date}",
			func(w http.ResponseWriter, r *http.Request) {
				personID := chi.URLParam(r, "person_id")
				fromDate := chi.URLParam(r, "from_date")

				parsedFromDate, err := time.Parse("2006-01-02", fromDate)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				toDate := chi.URLParam(r, "to_date")

				parsedToDate, err := time.Parse("2006-01-02", toDate)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				personActivitiesDetail, err := service.GetPersonActivitiesDetail(
					personID,
					r.Context(),
					parsedFromDate, parsedToDate,
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}

				w.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(w).Encode(personActivitiesDetail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}
			},
		)

		router.Get(
			"/bank_activities_details/{bank_id}/{from_date}/{to_date}",
			func(w http.ResponseWriter, r *http.Request) {
				bankID := chi.URLParam(r, "bank_id")
				fromDate := chi.URLParam(r, "from_date")

				parsedFromDate, err := time.Parse("2006-01-02", fromDate)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				toDate := chi.URLParam(r, "to_date")

				parsedToDate, err := time.Parse("2006-01-02", toDate)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				bankActivitiesDetail, err := service.GetBankActivitiesDetail(
					bankID,
					parsedFromDate, parsedToDate,
					r.Context(),
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}

				w.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(w).Encode(bankActivitiesDetail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}
			},
		)

		router.Get(
			"/cash_activities_details/{code}/{from_date}/{to_date}",
			func(w http.ResponseWriter, r *http.Request) {
				code := chi.URLParam(r, "code")
				fromDate := chi.URLParam(r, "from_date")

				parsedFromDate, err := time.Parse("2006-01-02", fromDate)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				toDate := chi.URLParam(r, "to_date")

				parsedToDate, err := time.Parse("2006-01-02", toDate)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				cashActivitiesDetail, err := service.GetCashActivitiesDetail(
					r.Context(),
					code,
					parsedFromDate, parsedToDate,
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}

				w.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(w).Encode(cashActivitiesDetail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}
			},
		)

		router.Get(
			"/daily_activities_details/{date}",
			func(w http.ResponseWriter, r *http.Request) {
				dateString := chi.URLParam(r, "date")

				date, err := time.Parse("2006-01-02", dateString)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				dailyActivitiesDetail, err := service.GetDailyActivitiesDetail(
					r.Context(),
					date,
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}

				w.Header().Set("Content-Type", "application/json")

				if err := json.NewEncoder(w).Encode(dailyActivitiesDetail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}
			},
		)
	})
}
