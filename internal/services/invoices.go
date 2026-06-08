package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/pzonouz/pzonouz-caroption-back-golang/internal/utils"
)

func (s *Service) CreateInvoice(inv Invoice, ctx context.Context) (string, string, string, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", "", "", err
	}
	defer tx.Rollback(ctx)

	var invoiceNumber int

	query := `
		INSERT INTO invoices (id, person_id, type, discount, notes, date, total, description, created_at, updated_at)
		    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING number`
	invoiceId := uuid.New()

	err = tx.QueryRow(ctx, query,
		invoiceId,
		inv.PersonID,
		inv.Type,
		inv.Discount,
		inv.Notes,
		inv.Date,
		inv.Total,
		inv.Description,
		time.Now(),
		time.Now(),
	).Scan(&invoiceNumber)
	if err != nil {
		tx.Rollback(ctx)

		return "", "", "", err
	}

	for _, item := range inv.Items {
		itemQuery := `
			INSERT INTO invoice_items (id, invoice_id, description, price, product_id, count, discount)
			    VALUES ($1, $2, $3, $4, $5, $6, $7)`
		id := uuid.New()

		_, err = tx.Exec(ctx, itemQuery,
			id,
			invoiceId,
			item.Description,
			item.Price,
			item.ProductID,
			item.Count,
			item.Discount,
		)
		if err != nil {
			tx.Rollback(ctx)

			return "", "", "", err
		}
		// Update buy price of product
		if inv.Type == "BUY" {
			updateProductBuyPrice := `UPDATE products SET buy_price=$1 WHERE id=$2`
			_, err = tx.Exec(ctx, updateProductBuyPrice, item.Price, item.ProductID)
			if err != nil {
				tx.Rollback(ctx)

				return "", "", "", err
			}
		}
	}

	// ---------------------------------------------------------
	// -----------------VoucherCreate
	// ---------------------------------------------------------
	var voucherNumber string

	voucherId := uuid.New()

	var voucherDescription string

	productActivityRelatedAcccountVoucherCreateQuery := `INSERT INTO vouchers (id,date,description,is_invoice) VALUES ($1,$2,$3,$4) RETURNING voucher_number`

	if inv.Type == "SELL" {
		voucherDescription = fmt.Sprintf("فاکتورفروش شماره %d ", invoiceNumber)
	}

	if inv.Type == "SELL" {
		voucherDescription = fmt.Sprintf("فاکتورخریدشماره %d ", invoiceNumber)
	}

	err = tx.QueryRow(
		ctx,
		productActivityRelatedAcccountVoucherCreateQuery,
		voucherId,
		inv.Date,
		voucherDescription,
		true,
	).Scan(&voucherNumber)
	if err != nil {
		tx.Rollback(ctx)

		return "", "", "", err
	}

	var accountId string

	if inv.Type == "SELL" {
		getAccountIdOfProductSell := `SELECT id from accounts WHERE code='4101'`
		err = tx.QueryRow(ctx, getAccountIdOfProductSell).Scan(&accountId)
	}

	if inv.Type == "BUY" {
		getAccountIdOfProductBuy := `SELECT id from accounts WHERE code='5101'`
		err = tx.QueryRow(ctx, getAccountIdOfProductBuy).Scan(&accountId)
	}

	if err != nil {
		tx.Rollback(ctx)

		return "", "", "", err
	}

	voucherItemId := uuid.New()

	var description string

	if inv.Type == "SELL" {
		productSellRelatedAcccountVoucherItemCreateQuery := `INSERT INTO voucher_items (id,voucher_id,account_id,credit,description) VALUES ($1,$2,$3,$4,$5)`
		description = fmt.Sprintf("فاکتورفروش شماره %d", invoiceNumber)
		_, err = tx.Exec(
			ctx,
			productSellRelatedAcccountVoucherItemCreateQuery,
			voucherItemId,
			voucherId,
			accountId,
			inv.Total,
			description,
		)
	}

	if inv.Type == "BUY" {
		productBuyRelatedAcccountVoucherItemCreateQuery := `INSERT INTO voucher_items (id,voucher_id,account_id,debit,description) VALUES ($1,$2,$3,$4,$5)`
		description = fmt.Sprintf("فاکتور خرید شماره %d", invoiceNumber)
		_, err = tx.Exec(
			ctx,
			productBuyRelatedAcccountVoucherItemCreateQuery,
			voucherItemId,
			voucherId,
			accountId,
			inv.Total,
			description,
		)
	}

	if err != nil {
		tx.Rollback(ctx)

		return "", "", "", err
	}

	voucherItemId = uuid.New()

	if inv.Type == "SELL" {
		personDebitRelatedAcccountVoucherItemCreateQuery := `INSERT INTO voucher_items (id,voucher_id,person_id,debit,description) VALUES ($1,$2,$3,$4,$5)`

		_, err = tx.Exec(
			ctx,
			personDebitRelatedAcccountVoucherItemCreateQuery,
			voucherItemId,
			voucherId,
			inv.PersonID,
			inv.Total,
			description,
		)
	}

	if inv.Type == "BUY" {
		personCreditRelatedAcccountVoucherItemCreateQuery := `INSERT INTO voucher_items (id,voucher_id,person_id,credit,description) VALUES ($1,$2,$3,$4,$5)`

		_, err = tx.Exec(
			ctx,
			personCreditRelatedAcccountVoucherItemCreateQuery,
			voucherItemId,
			voucherId,
			inv.PersonID,
			inv.Total,
			description,
		)
	}

	if err != nil {
		tx.Rollback(ctx)

		return "", "", "", err
	}

	updateVoucherIdOfInvoiceQuery := `UPDATE invoices SET voucher_id=$1 WHERE id=$2`

	_, err = tx.Exec(ctx, updateVoucherIdOfInvoiceQuery, voucherId, invoiceId)
	if err != nil {
		return "", "", "", err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return "", "", "", err
	}

	return strconv.Itoa(invoiceNumber), voucherNumber, invoiceId.String(), nil
}

func (s *Service) GetInvoice(id string) (Invoice, error) {
	query := `
	SELECT
    invoices.id,
    invoices.person_id,
    CONCAT(persons.name, ' ', persons.first_name) AS person_name,
    invoices.type,
    invoices.discount,
    invoices.notes,
    invoices.number,
	  vouchers.voucher_number,
    invoices.total,
    COALESCE(json_agg(
        json_build_object(
            'id', invoice_items.id,
            'invoiceID', invoice_items.invoice_id,
            'description', invoice_items.description,
            'price', invoice_items.price,
            'productID', invoice_items.product_id,
            'productName', products.name,
            'count', invoice_items.count,
            'discount', invoice_items.discount
        )
    )FILTER (WHERE invoice_items.id IS NOT NULL),'[]') AS items,
    invoices.date,
    invoices.created_at,
    invoices.updated_at
FROM
    invoices
    LEFT JOIN persons ON invoices.person_id = persons.id
    LEFT JOIN vouchers ON invoices.voucher_id = vouchers.id
    LEFT JOIN invoice_items ON invoices.id = invoice_items.invoice_id
    LEFT JOIN products ON invoice_items.product_id = products.id
WHERE
    invoices.id = $1
GROUP BY
    invoices.id,
    persons.name,
    persons.first_name,
	  vouchers.voucher_number;
	`

	var inv Invoice

	err := s.db.QueryRow(context.Background(), query, id).Scan(
		&inv.ID,
		&inv.PersonID,
		&inv.PersonName,
		&inv.Type,
		&inv.Discount,
		&inv.Notes,
		&inv.Number,
		&inv.VoucherNumber,
		&inv.Total,
		&inv.Items,
		&inv.Date,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if err != nil {
		return Invoice{}, err
	}

	return inv, nil
}

func (s *Service) ListInvoicesWithSortFilterPagination(
	sort string,
	sortDirection string,
	filters []string,
	filterOperands []string,
	filterConditions []string,
	countInPage string,
	offset string,
	w http.ResponseWriter,
) {
	pagedBy := ""
	offsetNum := 0

	if countInPage != "" {
		limit, err := strconv.Atoi(countInPage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		if offset != "" {
			offsetNum, err = strconv.Atoi(offset)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)

				return
			}
		}

		pagedBy = fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offsetNum)
	}

	orderByClause := utils.BuildOrderBy("invoices", sort, sortDirection)

	whereClause, args := utils.BuildWhere(
		"invoices",
		filters,
		filterOperands,
		filterConditions,
		1,
	)

	query := fmt.Sprintf(`
		SELECT
    invoices.id,
    invoices.person_id,
    CONCAT(persons.name, ' ', persons.first_name) AS person_name,
    invoices.type,
    invoices.discount,
    invoices.notes,
    invoices.number,
	  vouchers.voucher_number,
    invoices.total,
    COALESCE(json_agg(
        json_build_object(
            'id', invoice_items.id,
            'invoiceID', invoice_items.invoice_id,
            'description', invoice_items.description,
            'price', invoice_items.price,
            'productID', invoice_items.product_id,
            'count', invoice_items.count,
		        'productName', products.name,
		        'discount', invoice_items.discount
        )
    )FILTER (WHERE invoice_items.id IS NOT NULL),'[]') AS items,
    invoices.date,
    invoices.created_at,
    invoices.updated_at
FROM
    invoices
    LEFT JOIN persons ON invoices.person_id = persons.id
    LEFT JOIN vouchers ON invoices.voucher_id = vouchers.id
    LEFT JOIN invoice_items ON invoices.id = invoice_items.invoice_id
    LEFT JOIN products ON invoice_items.product_id = products.id
    %s
GROUP BY
    invoices.id,
    persons.name,
		vouchers.voucher_number,
    persons.first_name
    %s %s;
		`,
		whereClause, orderByClause, pagedBy)

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var invoices []Invoice

	for rows.Next() {
		var invoice Invoice
		if err := rows.Scan(&invoice.ID, &invoice.PersonID, &invoice.PersonName, &invoice.Type, &invoice.Discount, &invoice.Notes, &invoice.Number, &invoice.VoucherNumber, &invoice.Total, &invoice.Items, &invoice.Date, &invoice.CreatedAt, &invoice.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		invoices = append(invoices, invoice)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM
		    invoices
		    LEFT JOIN persons ON invoices.person_id = persons.id
		%s
		`, whereClause)
	row := s.db.QueryRow(context.Background(), newQuery, args...)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var invoicesWithTotalCount struct {
		Rows       []Invoice `json:"rows"`
		TotalCount int32     `json:"totalCount"`
	}

	invoicesWithTotalCount.Rows = invoices
	invoicesWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(invoicesWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) EditInvoice(id string, invoice Invoice) error {
	ctx := context.Background()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	// ✅ Defer rollback ensures the tx is closed on panic or early return.
	// (If tx is already committed, rollback safely does nothing in pgx)
	defer tx.Rollback(ctx)

	updateInvoiceQuery := `
		UPDATE
		    invoices
		SET
		    person_id = $1,
		    type = $2,
		    notes = $3,
	      date  = $4,
	      total  = $5
		WHERE
		    id = $6`

	_, err = tx.Exec(ctx, updateInvoiceQuery,
		invoice.PersonID,
		invoice.Type,
		invoice.Notes,
		invoice.Date,
		invoice.Total,
		id,
	)
	if err != nil {
		return err
	}

	itemIDs := make([]uuid.UUID, 0, len(invoice.Items))

	for _, item := range invoice.Items {
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}

		itemIDs = append(itemIDs, item.ID)

		itemQuery := `
			INSERT INTO invoice_items (id, invoice_id, description, price, product_id, count, discount)
			    VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id)
			    DO UPDATE SET
			        description = EXCLUDED.description,
			        price = EXCLUDED.price,
			        product_id = EXCLUDED.product_id,
			        count = EXCLUDED.count,
			        discount = EXCLUDED.discount`

		_, err = tx.Exec(ctx, itemQuery,
			item.ID,
			id,
			item.Description,
			item.Price,
			item.ProductID,
			item.Count,
			item.Discount,
		)
		if err != nil {
			return err
		}
	}

	// ✅ Safer array comparison for empty slices
	deleteQuery := `
		DELETE FROM invoice_items
		WHERE invoice_id = $1
		    AND id != ALL($2::UUID[])`

	_, err = tx.Exec(ctx, deleteQuery, id, itemIDs) // ✅ Changed from invoice.ID to id
	if err != nil {
		return err
	}

	// ---------------------------------------------------------------
	// ------------------voucher Edit---------------------------------
	// ---------------------------------------------------------------

	// 1. Get voucher id
	var voucherId string

	err = tx.QueryRow(ctx, `
        SELECT voucher_id FROM invoices WHERE id = $1
    `, id).Scan(&voucherId)
	if err != nil {
		tx.Rollback(ctx)

		return err
	}

	// 2. Update voucher
	_, err = tx.Exec(ctx, `
        UPDATE vouchers
        SET date = $1,
            description = $2,
            updated_at = NOW()
        WHERE id = $3
    `,
		invoice.Date,
		invoice.Description,
		voucherId,
	)
	if err != nil {
		tx.Rollback(ctx)

		return err
	}

	// 3. Update voucher Items

	debit := "0"
	credit := "0"
	// person Side
	if invoice.Type == "BUY" {
		credit = invoice.Total
		debit = "0"
	}

	if invoice.Type == "SELL" {
		debit = invoice.Total
		credit = "0"
	}

	_, err = tx.Exec(ctx, `
        UPDATE voucher_items
        SET person_id = $1,
            debit = $2,
            credit = $3,
            updated_at = NOW()
        WHERE voucher_id = $4 AND person_id IS NOT NULL
    `, invoice.PersonID, debit, credit, voucherId,
	)
	if err != nil {
		tx.Rollback(ctx)

		return err
	}
	// account Side
	if invoice.Type == "SELL" {
		credit = invoice.Total
		debit = "0"
	}

	if invoice.Type == "BUY" {
		debit = invoice.Total
		credit = "0"
	}

	_, err = tx.Exec(ctx, `
        UPDATE voucher_items
        SET debit = $1,
            credit = $2,
            updated_at = NOW()
        WHERE voucher_id = $3 AND person_id IS NULL
    `, debit, credit, voucherId,
	)
	if err != nil {
		tx.Rollback(ctx)

		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) DeleteInvoice(id string, ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		SELECT voucher_id FROM invoices
		WHERE id = $1`

	var voucherId string

	err = tx.QueryRow(ctx, query, id).Scan(&voucherId)
	if err != nil {
		tx.Rollback(ctx)

		return err
	}

	query = `
		DELETE FROM invoices
		WHERE id = $1`

	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		tx.Rollback(ctx)

		return err
	}

	query = `
		DELETE FROM vouchers
		WHERE id = $1`

	_, err = tx.Exec(ctx, query, voucherId)
	if err != nil {
		tx.Rollback(ctx)

		return err
	}

	return tx.Commit(ctx)
}
