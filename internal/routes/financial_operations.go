package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/services"
	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
	"github.com/pzonouz/pzonouz-caroption-back-golang/middlewares"
)

func GenerateFinancialOperationsRoutes(mainRouter *chi.Mux, service services.Service) {
	mainRouter.With(middlewares.AdminOnly).
		Route("/financial_operations", func(router chi.Router) {
			router.Post("/buy_invoice_settlement", func(w http.ResponseWriter, r *http.Request) {
				buyInvoiceSettlement, err := utils.DecodeBody[services.BuyInvoiceSettlement](r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				err = service.BuyInvoiceSettlementOperation(buyInvoiceSettlement, r.Context())
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}
			})
			router.Post("/receipt_from_person", func(w http.ResponseWriter, r *http.Request) {
				receiptFromPerson, err := utils.DecodeBody[services.ReceiptFromPerson](r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				voucherNumber, err := service.ReceiptFromPersonOperation(
					receiptFromPerson,
					r.Context(),
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				type ReceiptFromPersonVoucherNumber struct {
					VoucherNumber string `json:"voucherNumber"`
				}

				var receiptFromPersonVoucherNumber ReceiptFromPersonVoucherNumber

				receiptFromPersonVoucherNumber.VoucherNumber = voucherNumber

				json.NewEncoder(w).Encode(receiptFromPersonVoucherNumber)
			})
			router.Post("/sell_invoice_settlement", func(w http.ResponseWriter, r *http.Request) {
				sellInvoiceSettlement, err := utils.DecodeBody[services.SellInvoiceSettlement](r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				err = service.SellInvoiceSettlementOperation(sellInvoiceSettlement, r.Context())
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}
			})
			router.Post("/receipt_from_person", func(w http.ResponseWriter, r *http.Request) {
				receiptFromPerson, err := utils.DecodeBody[services.ReceiptFromPerson](r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				voucherNumber, err := service.ReceiptFromPersonOperation(
					receiptFromPerson,
					r.Context(),
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				type ReceiptFromPersonVoucherNumber struct {
					VoucherNumber string `json:"voucherNumber"`
				}

				var receiptFromPersonVoucherNumber ReceiptFromPersonVoucherNumber

				receiptFromPersonVoucherNumber.VoucherNumber = voucherNumber

				json.NewEncoder(w).Encode(receiptFromPersonVoucherNumber)
			})

			router.Post("/payment_to_person", func(w http.ResponseWriter, r *http.Request) {
				paymentToPerson, err := utils.DecodeBody[services.PaymentToPerson](r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				voucherNumber, err := service.PaymentToPersonOperation(
					paymentToPerson,
					r.Context(),
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				type PaymentToPersonVoucherNumber struct {
					VoucherNumber string `json:"voucherNumber"`
				}

				var paymentToPersonVoucherNumber PaymentToPersonVoucherNumber

				paymentToPersonVoucherNumber.VoucherNumber = voucherNumber

				json.NewEncoder(w).Encode(paymentToPersonVoucherNumber)
			})

			router.Post("/bank_to_bank", func(w http.ResponseWriter, r *http.Request) {
				bankToBank, err := utils.DecodeBody[services.BankToBank](r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				err = service.BankToBankOperation(
					bankToBank,
					r.Context(),
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}
			})

			router.Post("/cash_to_bank", func(w http.ResponseWriter, r *http.Request) {
				cashToBank, err := utils.DecodeBody[services.CashToBank](r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				err = service.CashToBankOperation(
					cashToBank,
					r.Context(),
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}
			})

			router.Post("/bank_to_cash", func(w http.ResponseWriter, r *http.Request) {
				bankToCash, err := utils.DecodeBody[services.BankToCash](r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				err = service.BankToCashOperation(
					bankToCash,
					r.Context(),
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}
			})

			router.Post("/bank_to_cost", func(w http.ResponseWriter, r *http.Request) {
				bankToCost, err := utils.DecodeBody[services.BankToCost](r, w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				err = service.BankToCostOperation(
					bankToCost,
					r.Context(),
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}
			})
		})
}
