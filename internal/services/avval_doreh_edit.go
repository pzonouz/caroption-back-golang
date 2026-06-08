package services

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func (s *Service) IncomeUpdate(tx pgx.Tx, ctx context.Context) error {
	var incomeSum, tmp1, tmp2 string

	incomeCalculateQuery := `
		SELECT
		    SUM(REPLACE(voucher_items.debit, ',', '')::BIGINT) AS debit_sum,
		    SUM(REPLACE(voucher_items.credit, ',', '')::BIGINT) AS credit_sum,
		    SUM(REPLACE(voucher_items.debit, ',', '')::BIGINT) - SUM(REPLACE(voucher_items.credit, ',', '')::BIGINT) AS different
		FROM
		    voucher_items
		    LEFT JOIN vouchers ON vouchers.id = voucher_items.voucher_id
		    LEFT JOIN accounts ON accounts.id = voucher_items.account_id
		WHERE
		    vouchers.voucher_number = '1'
		    AND accounts.code != '310101'`

	tx.QueryRow(ctx, incomeCalculateQuery).Scan(&tmp1, &tmp2, &incomeSum)

	incomeUpdateQuery := `
		UPDATE
		    voucher_items
		SET
		    credit = $1
		FROM
		    vouchers,
		    accounts
		WHERE
		    voucher_items.voucher_id = vouchers.id
		    AND vouchers.voucher_number = '1'
		    AND voucher_items.account_id = accounts.id
		    AND accounts.code = '310101'`

	_, err := tx.Exec(ctx, incomeUpdateQuery, incomeSum)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) AvvalDorehProductEdit(
	ctx context.Context,
	id string,
	count string,
	buy_price string,
) error {
	query := `
		UPDATE
		    products
		SET
		    buy_price = $1,
		    count = $2
		WHERE
		    id = $3`

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, query, buy_price, count, id)
	if err != nil {
		return err
	}

	var productBuyPriceCountSum int64

	productsBuyPriceCountSumQuery := `
		SELECT
		    SUM(REPLACE(buy_price, ',', '')::BIGINT * count::BIGINT)
		FROM
		    products
		WHERE
		    GENERATED IS FALSE`

	err = tx.QueryRow(ctx, productsBuyPriceCountSumQuery).Scan(&productBuyPriceCountSum)
	if err != nil {
		return err
	}

	productBuyPriceCountSumString := strconv.Itoa(int(productBuyPriceCountSum))
	updateVoucherItemQuery := `
		UPDATE
		    voucher_items
		SET
		    debit = $1
		FROM
		    vouchers,
		    accounts
		WHERE
		    voucher_items.voucher_id = vouchers.id
		    AND voucher_items.account_id = accounts.id
		    AND vouchers.voucher_number = $2
		    AND accounts.code = $3`

	_, err = tx.Exec(ctx, updateVoucherItemQuery, productBuyPriceCountSumString, "1", "110601")
	if err != nil {
		return err
	}

	err = s.IncomeUpdate(tx, ctx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) AvvalDorehPersonEdit(
	ctx context.Context,
	id string,
	debit string,
	credit string,
) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	personUpdateQuery := `
		UPDATE
		    persons
		SET
		    debit = $1,
		    credit = $2
		WHERE
		    id = $3`

	_, err = tx.Exec(ctx, personUpdateQuery, debit, credit, id)
	if err != nil {
		return err
	}

	debitPersonsSumQuery := `
		SELECT
		    SUM(REPLACE(debit, ',', '')::BIGINT)
		FROM
		    persons`

	var debitPersonsSum string

	err = tx.QueryRow(ctx, debitPersonsSumQuery).Scan(&debitPersonsSum)
	if err != nil {
		return err
	}

	creditPersonsSumQuery := `
		SELECT
		    SUM(REPLACE(credit, ',', '')::BIGINT)
		FROM
		    persons`

	var creditPersonsSum string

	err = tx.QueryRow(ctx, creditPersonsSumQuery).Scan(&creditPersonsSum)
	if err != nil {
		return err
	}

	updateVoucherItemQuery := `
		UPDATE
		    voucher_items
		SET
		    debit = $1,
		    credit = $2
		FROM
		    vouchers,
		    accounts
		WHERE
		    voucher_items.voucher_id = vouchers.id
		    AND voucher_items.account_id = accounts.id
		    AND accounts.code = $3`

	_, err = tx.Exec(ctx, updateVoucherItemQuery, debitPersonsSum, "0", "110401")
	if err != nil {
		return err
	}

	updateVoucherItemQuery = `
		UPDATE
		    voucher_items
		SET
		    debit = $1,
		    credit = $2
		FROM
		    vouchers,
		    accounts
		WHERE
		    voucher_items.voucher_id = vouchers.id
		    AND voucher_items.account_id = accounts.id
		    AND accounts.code = $3`

	_, err = tx.Exec(ctx, updateVoucherItemQuery, "0", creditPersonsSum, "210101")
	if err != nil {
		return err
	}

	err = s.IncomeUpdate(tx, ctx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) AvvalDorehBankEdit(
	ctx context.Context,
	id string,
	firstBalance string,
) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	bankUpdateQuery := `
		UPDATE
		    banks
		SET
		    first_balance = $1
		WHERE
		    id = $2`

	_, err = tx.Exec(ctx, bankUpdateQuery, firstBalance, id)
	if err != nil {
		return err
	}

	var accountId string

	bankAccountIdGetQuery := `
		SELECT
		    account_id
		FROM
		    banks
		WHERE
		    id = $1`

	err = tx.QueryRow(ctx, bankAccountIdGetQuery, id).Scan(&accountId)
	if err != nil {
		return err
	}

	updateVoucherItemQuery := `
		UPDATE
		    voucher_items
		SET
		    debit = $1,
		    credit = '0'
		WHERE
		    voucher_items.account_id = $2`

	_, err = tx.Exec(ctx, updateVoucherItemQuery, firstBalance, accountId)
	if err != nil {
		return err
	}

	err = s.IncomeUpdate(tx, ctx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) AvvalDorehCashEdit(
	ctx context.Context,
	accountCode string,
	firstBalance string,
) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	updateVoucherItemQuery := `
		UPDATE
		    voucher_items
		SET
		    debit = $1,
		    credit = '0'
		FROM
		    vouchers,
		    accounts
		WHERE
		    vouchers.id = voucher_items.voucher_id
		    AND voucher_items.account_id = accounts.id
		    AND vouchers.voucher_number = '1'
		    AND accounts.code = $2`

	_, err = tx.Exec(ctx, updateVoucherItemQuery, firstBalance, accountCode)
	if err != nil {
		return err
	}

	err = s.IncomeUpdate(tx, ctx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
