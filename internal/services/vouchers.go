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

func (s *Service) CreateVoucher(voucher Voucher) error {
	tx, err := s.db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	query := `
		INSERT INTO vouchers (id, date, description, created_at, updated_at)
		    VALUES ($1, $2, $3, $4, $5) RETURNING voucher_number`
	voucherId := uuid.New()

	var voucherNumber int

	err = tx.QueryRow(context.Background(), query,
		voucherId,
		voucher.Date,
		voucher.Description,
		time.Now(),
		time.Now(),
	).Scan(&voucherNumber)
	if err != nil {
		tx.Rollback(context.Background())

		return err
	}

	if voucherNumber == 1 {
		query = `UPDATE settings SET value='NO' WHERE key='AVVAL_DOREH'`

		_, err = tx.Exec(context.Background(), query)
		if err != nil {
			tx.Rollback(context.Background())

			return err
		}
	}

	for _, item := range voucher.Items {
		itemQuery := `
			INSERT INTO voucher_items (id, voucher_id, account_id, person_id, debit, credit, description)
			    VALUES ($1, $2, $3, $4, $5, $6, $7)`
		id := uuid.New()

		_, err = tx.Exec(context.Background(), itemQuery,
			id,
			voucherId,
			item.AccountId,
			item.PersonId,
			item.Debit,
			item.Credit,
			item.Description,
		)
		if err != nil {
			tx.Rollback(context.Background())

			return err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetVoucher(id string) (Voucher, error) {
	query := `
	SELECT
    vouchers.id,
    vouchers.voucher_number,
    vouchers.date,
    vouchers.description,
		vouchers.is_invoice,
    COALESCE(json_agg(
        json_build_object(
            'id', voucher_items.id,
            'voucherId', voucher_items.voucher_id,
            'accountId', voucher_items.account_id,
            'personId', voucher_items.person_id,
            'debit', voucher_items.debit,
            'credit', voucher_items.credit,
            'description', voucher_items.description,
            'created_at', voucher_items.created_at,
            'updated_at', voucher_items.updated_at
        )
    )FILTER (WHERE voucher_items.id IS NOT NULL),'[]') AS items,
    vouchers.created_at,
    vouchers.updated_at
FROM
    vouchers
    LEFT JOIN voucher_items ON vouchers.id = voucher_items.voucher_id
WHERE
    vouchers.id = $1
GROUP BY
    vouchers.id
	`

	var voucher Voucher

	err := s.db.QueryRow(context.Background(), query, id).Scan(
		&voucher.ID,
		&voucher.VoucherNumber,
		&voucher.Date,
		&voucher.Description,
		&voucher.IsInvoice,
		&voucher.Items,
		&voucher.CreatedAt,
		&voucher.UpdatedAt,
	)
	if err != nil {
		return Voucher{}, err
	}

	return voucher, nil
}

func (s *Service) GetVoucherForView(id string) (VoucherForView, error) {
	query := `
	SELECT
    vouchers.id,
    vouchers.voucher_number,
    vouchers.date,
    vouchers.description,
		vouchers.is_invoice,
    COALESCE(json_agg(
        json_build_object(
            'id', voucher_items.id,
            'voucherId', voucher_items.voucher_id,
            'accountName', a.name,
            'accountCode', a.code,
            'personName', p.name,
            'personFirstName', p.first_name,
            'personPhone', p.phone_number,
            'debit', voucher_items.debit,
            'credit', voucher_items.credit,
            'description', voucher_items.description,
            'created_at', voucher_items.created_at,
            'updated_at', voucher_items.updated_at
        )
    )FILTER (WHERE voucher_items.id IS NOT NULL),'[]') AS items,
    vouchers.created_at,
    vouchers.updated_at
FROM
    vouchers
    LEFT JOIN voucher_items ON vouchers.id = voucher_items.voucher_id
    LEFT JOIN persons  p ON p.id  = voucher_items.person_id
    LEFT JOIN accounts a ON a.id = voucher_items.account_id
WHERE
    vouchers.id = $1
GROUP BY
    vouchers.id
	`

	var voucher VoucherForView

	err := s.db.QueryRow(context.Background(), query, id).Scan(
		&voucher.ID,
		&voucher.VoucherNumber,
		&voucher.Date,
		&voucher.Description,
		&voucher.IsInvoice,
		&voucher.Items,
		&voucher.CreatedAt,
		&voucher.UpdatedAt,
	)
	if err != nil {
		return VoucherForView{}, err
	}

	return voucher, nil
}

func (s *Service) ListVouchersWithSortFilterPagination(
	sort string,
	sortDirection string,
	filters []string,
	filterOperands []string,
	filterConditions []string,
	countInPage string,
	offset string,
	w http.ResponseWriter,
) {
	orderByClause := utils.BuildOrderBy("vouchers", sort, sortDirection)

	whereClause, args := utils.BuildWhere(
		"vouchers",
		filters,
		filterOperands,
		filterConditions,
		1,
	)

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

	query := fmt.Sprintf(`
	SELECT
    vouchers.id,
    vouchers.voucher_number,
    vouchers.date,
    vouchers.description,
		vouchers.is_invoice,
    COALESCE(json_agg(
        json_build_object(
            'id', voucher_items.id,
            'voucherId', voucher_items.voucher_id,
            'accountId', accounts.id,
            'debit', voucher_items.debit,
            'credit', voucher_items.credit,
            'description', voucher_items.description,
            'created_at', voucher_items.created_at,
            'updated_at', voucher_items.updated_at
        )
    )FILTER (WHERE voucher_items.id IS NOT NULL),'[]') AS items,
    vouchers.created_at,
    vouchers.updated_at
FROM
    vouchers
    LEFT JOIN voucher_items ON vouchers.id = voucher_items.voucher_id
		LEFT JOIN accounts ON accounts.id = voucher_items.account_id
		%s
GROUP BY
    vouchers.id %s %s
		`,
		whereClause, orderByClause, pagedBy)

	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	defer rows.Close()

	var vouchers []Voucher

	for rows.Next() {
		var voucher Voucher
		if err := rows.Scan(&voucher.ID, &voucher.VoucherNumber, &voucher.Date, &voucher.Description, &voucher.IsInvoice, &voucher.Items, &voucher.CreatedAt, &voucher.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		vouchers = append(vouchers, voucher)
	}

	w.Header().Add("Content-Type", "application/json")

	newQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM
		    vouchers
		%s
		`, whereClause)
	row := s.db.QueryRow(context.Background(), newQuery, args...)

	var Count int32

	err = row.Scan(&Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	var vouchersWithTotalCount struct {
		Rows       []Voucher `json:"rows"`
		TotalCount int32     `json:"totalCount"`
	}

	vouchersWithTotalCount.Rows = vouchers
	vouchersWithTotalCount.TotalCount = Count

	err = json.NewEncoder(w).Encode(vouchersWithTotalCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
}

func (s *Service) EditVoucher(id string, voucher Voucher) error {
	ctx := context.Background()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	updateVoucherQuery := `
		UPDATE
		    vouchers
		SET
		    date = $1,
		    description = $2
		WHERE
		    id = $3`

	_, err = tx.Exec(ctx, updateVoucherQuery,
		voucher.Date,
		voucher.Description,
		id,
	)
	if err != nil {
		return err
	}

	itemIDs := make([]uuid.UUID, 0, len(voucher.Items))

	for _, item := range voucher.Items {
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}

		itemIDs = append(itemIDs, item.ID)

		itemQuery := `
			INSERT INTO voucher_items (id, voucher_id, account_id, person_id , debit , credit, description)
			    VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id)
			    DO UPDATE SET
			        voucher_id = EXCLUDED.voucher_id,
			        account_id = EXCLUDED.account_id ,
			        person_id = EXCLUDED.person_id ,
			        debit = EXCLUDED.debit,
			        credit = EXCLUDED.credit,
			        description = EXCLUDED.description`

		_, err = tx.Exec(ctx, itemQuery,
			item.ID,
			id,
			item.AccountId,
			item.PersonId,
			item.Debit,
			item.Credit,
			item.Description,
		)
		if err != nil {
			return err
		}
	}

	// ✅ Safer array comparison for empty slices
	deleteQuery := `
		DELETE FROM voucher_items
		WHERE voucher_id = $1
		    AND id != ALL($2::UUID[])`

	_, err = tx.Exec(ctx, deleteQuery, id, itemIDs) // ✅ Changed from voucher.ID to id
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) DeleteVoucher(id string) error {
	query := `
		DELETE FROM vouchers
		WHERE id = $1`

	_, err := s.db.Exec(context.Background(), query, id)
	if err != nil {
		return err
	}

	return nil
}
