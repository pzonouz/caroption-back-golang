package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Service) ReceiptFromPersonOperation(rfp ReceiptFromPerson,
	ctx context.Context,
) (string, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var bankBalance int

	bankRaw := strings.TrimSpace(rfp.Bank.String)

	if bankRaw == "" {
		bankBalance = 0
	} else {
		bankBalance, err = strconv.Atoi(bankRaw)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}
	}

	var cash int

	cashRaw := strings.TrimSpace(rfp.Cash.String)
	if cashRaw == "" {
		cash = 0
	} else {
		cash, err = strconv.Atoi(cashRaw)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}
	}

	if cash+bankBalance <= 0 {
		return "", nil
	}

	var firstName, lastName string

	searchPersonQuery := `
		SELECT
		    name,
		    first_name
		FROM
		    persons
		WHERE
		    id = $1`

	err = tx.QueryRow(ctx, searchPersonQuery, rfp.PersonId).Scan(&lastName, &firstName)
	if err != nil {
		tx.Rollback(ctx)

		return "", err
	}

	createVoucherQuery := `
		INSERT INTO vouchers (id, date, description)
		    VALUES ($1, $2, $3)`
	voucherId := uuid.New()
	description := "دریافتی"

	_, err = tx.Exec(
		ctx,
		createVoucherQuery,
		voucherId,
		rfp.Date,
		description,
	)
	if err != nil {
		tx.Rollback(ctx)

		return "", err
	}

	var bankBranch, bankName, bankNumber, accountId *string

	if bankBalance > 0 {
		bankSearchQuery := `
			SELECT
			    name,
			    branch,
			    number,
			    account_id
			FROM
			    banks
			WHERE
			    id = $1`

		err = tx.QueryRow(ctx, bankSearchQuery, rfp.BankId).
			Scan(&bankName, &bankBranch, &bankNumber, &accountId)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}

		id := uuid.New()
		bankVoucherItemQuery := `
			INSERT INTO voucher_items (id, account_id, debit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			bankVoucherItemQuery,
			id,
			accountId,
			rfp.Bank,
			voucherId,
			"دریافت از"+" "+lastName+"-"+firstName,
		)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}
	}

	if cash > 0 {
		id := uuid.New()

		var accountId string

		accountSearchQuery := `
			SELECT
			    id
			FROM
			    accounts
			WHERE
			    code = $1`

		err = tx.QueryRow(ctx, accountSearchQuery, "110101").Scan(&accountId)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}

		cashVoucherItemQuery := `
			INSERT INTO voucher_items (id, account_id, debit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			cashVoucherItemQuery,
			id,
			accountId,
			rfp.Cash,
			voucherId,
			fmt.Sprintf("دریافتی از %s-%s", lastName, firstName),
		)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}
	}

	personBalance := (cash + bankBalance)
	if personBalance > 0 {
		id := uuid.New()

		personBalance := strconv.Itoa(cash + bankBalance)
		personIdVoucherItemQuery := `
			INSERT INTO voucher_items (id, person_id, credit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			personIdVoucherItemQuery,
			id,
			rfp.PersonId,
			personBalance,
			voucherId,
			"دریافت",
		)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}

		err = tx.Commit(ctx)
		if err != nil {
			return "", err
		}

		return voucherId.String(), nil
	}

	return "", nil
}

func (s *Service) PaymentToPersonOperation(ptp PaymentToPerson,
	ctx context.Context,
) (string, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var bankBalance int

	bankRaw := strings.TrimSpace(ptp.Bank.String)

	if bankRaw == "" {
		bankBalance = 0
	} else {
		bankBalance, err = strconv.Atoi(bankRaw)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}
	}

	var cash int

	cashRaw := strings.TrimSpace(ptp.Cash.String)
	if cashRaw == "" {
		cash = 0
	} else {
		cash, err = strconv.Atoi(cashRaw)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}
	}

	if cash+bankBalance <= 0 {
		return "", nil
	}

	var firstName, lastName string

	searchPersonQuery := `
		SELECT
		    name,
		    first_name
		FROM
		    persons
		WHERE
		    id = $1`

	err = tx.QueryRow(ctx, searchPersonQuery, ptp.PersonId).Scan(&lastName, &firstName)
	if err != nil {
		tx.Rollback(ctx)

		return "", err
	}

	createVoucherQuery := `
		INSERT INTO vouchers (id, date, description)
		    VALUES ($1, $2, $3)`
	voucherId := uuid.New()
	description := "پرداخت"

	_, err = tx.Exec(
		ctx,
		createVoucherQuery,
		voucherId,
		ptp.Date,
		description,
	)
	if err != nil {
		tx.Rollback(ctx)

		return "", err
	}

	for checkId := range ptp.Checks {
		id := uuid.New()
		checkVoucherItemQuery := `
			INSERT INTO voucher_items (id, account_id, credit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			checkVoucherItemQuery,
			id,
			"210201",
			checkId,
			checkId,
			"پرداخت"+" "+lastName+" "+firstName,
		)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}

		checkEditQuery := `
			UPDATE
			    checks
			SET
			    status = $1
			WHERE
			    id = $2`

		_, err = tx.Exec(
			ctx,
			checkEditQuery,
			"Transferred",
			checkId,
		)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}
	}

	var bankBranch, bankName, bankNumber, accountId string

	if bankBalance > 0 {
		bankSearchQuery := `
			SELECT
			    name,
			    branch,
			    number,
			    account_id
			FROM
			    banks
			WHERE
			    id = $1`

		err = tx.QueryRow(ctx, bankSearchQuery, ptp.BankId).
			Scan(&bankName, &bankBranch, &bankNumber, &accountId)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}

		id := uuid.New()
		bankVoucherItemQuery := `
			INSERT INTO voucher_items (id, account_id, credit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			bankVoucherItemQuery,
			id,
			accountId,
			ptp.Bank,
			voucherId,
			"پرداخت"+" "+lastName+" "+firstName,
		)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}
	}

	if cash > 0 {
		id := uuid.New()

		var accountId string

		accountSearchQuery := `
			SELECT
			    id
			FROM
			    accounts
			WHERE
			    code = $1`

		err = tx.QueryRow(ctx, accountSearchQuery, "110101").Scan(&accountId)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}

		cashVoucherItemQuery := `
			INSERT INTO voucher_items (id, account_id, credit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			cashVoucherItemQuery,
			id,
			accountId,
			ptp.Cash,
			voucherId,
			fmt.Sprintf("پرداختی %s - %s", lastName, firstName),
		)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}
	}

	personBalance := (cash + bankBalance)
	if personBalance > 0 {
		id := uuid.New()

		personBalance := strconv.Itoa(cash + bankBalance)
		personIdVoucherItemQuery := `
			INSERT INTO voucher_items (id, person_id, debit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			personIdVoucherItemQuery,
			id,
			ptp.PersonId,
			personBalance,
			voucherId,
			"پرداخت",
		)
		if err != nil {
			tx.Rollback(ctx)

			return "", err
		}

		err = tx.Commit(ctx)
		if err != nil {
			return "", err
		}

		return voucherId.String(), nil
	}

	return "", nil
}

func (s *Service) BuyInvoiceSettlementOperation(
	bis BuyInvoiceSettlement,
	ctx context.Context,
) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var invoice Invoice

	getInvoiceQuery := `
		SELECT
		    number,
		    date
		FROM
		    invoices
		WHERE
		    id = $1`

	err = tx.QueryRow(ctx, getInvoiceQuery, bis.InvoiceId).Scan(&invoice.Number, &invoice.Date)
	if err != nil {
		tx.Rollback(ctx)

		return err
	}

	var bank int

	bankRaw := strings.TrimSpace(bis.Bank.String)

	if bankRaw == "" {
		bank = 0
	} else {
		bank, err = strconv.Atoi(bankRaw)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}
	}

	var cash int

	cashRaw := strings.TrimSpace(bis.Cash.String)
	if cashRaw == "" {
		cash = 0
	} else {
		cash, err = strconv.Atoi(cashRaw)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}
	}

	if cash+bank <= 0 {
		return nil
	}

	createVoucherQuery := `
		INSERT INTO vouchers (id, date, description)
		    VALUES ($1, $2, $3)`
	voucherId := uuid.New()

	_, err = tx.Exec(
		ctx,
		createVoucherQuery,
		voucherId,
		invoice.Date,
		fmt.Sprintf("پرداختی فاکتور خرید شماره %s", invoice.Number.String),
	)
	if err != nil {
		tx.Rollback(ctx)

		return err
	}

	if bank > 0 {
		var accountId string

		accountSearchQuery := `
			SELECT
			    account_id
			FROM
			    banks
			WHERE
			    id = $1`

		err = tx.QueryRow(ctx, accountSearchQuery, bis.BankId).Scan(&accountId)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}

		id := uuid.New()
		bankVoucherItemQuery := `
			INSERT INTO voucher_items (id, account_id, credit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			bankVoucherItemQuery,
			id,
			accountId,
			bis.Bank,
			voucherId,
			fmt.Sprintf("پرداختی فاکتور خرید شماره %s", invoice.Number.String),
		)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}
	}

	if cash > 0 {
		id := uuid.New()

		var accountId string

		accountSearchQuery := `
			SELECT
			    id
			FROM
			    accounts
			WHERE
			    code = $1`

		err = tx.QueryRow(ctx, accountSearchQuery, "110101").Scan(&accountId)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}

		cashVoucherItemQuery := `
			INSERT INTO voucher_items (id, account_id, credit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			cashVoucherItemQuery,
			id,
			accountId,
			bis.Cash,
			voucherId,
			fmt.Sprintf("پرداخت فاکتور خرید شماره %s", invoice.Number.String),
		)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}
	}

	personBalance := (cash + bank)
	if personBalance > 0 {
		var personId string

		personSearchQuery := `
			SELECT
			    person_id
			FROM
			    invoices
			WHERE
			    id = $1`

		err = tx.QueryRow(ctx, personSearchQuery, bis.InvoiceId).Scan(&personId)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}

		id := uuid.New()

		personBalance := strconv.Itoa(cash + bank)
		personIdVoucherItemQuery := `
			INSERT INTO voucher_items (id, person_id, debit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			personIdVoucherItemQuery,
			id,
			personId,
			personBalance,
			voucherId,
			fmt.Sprintf("پرداخت فاکتور خرید شماره %s", invoice.Number.String),
		)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) SellInvoiceSettlementOperation(
	rfp SellInvoiceSettlement,
	ctx context.Context,
) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var invoice Invoice

	getInvoiceQuery := `
		SELECT
		    number,
		    date
		FROM
		    invoices
		WHERE
		    id = $1`

	err = tx.QueryRow(ctx, getInvoiceQuery, rfp.InvoiceId).Scan(&invoice.Number, &invoice.Date)
	if err != nil {
		tx.Rollback(ctx)

		return err
	}

	var bank int

	bankRaw := strings.TrimSpace(rfp.Bank.String)

	if bankRaw == "" {
		bank = 0
	} else {
		bank, err = strconv.Atoi(bankRaw)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}
	}

	var cash int

	cashRaw := strings.TrimSpace(rfp.Cash.String)
	if cashRaw == "" {
		cash = 0
	} else {
		cash, err = strconv.Atoi(cashRaw)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}
	}

	if cash+bank <= 0 {
		return nil
	}

	createVoucherQuery := `
		INSERT INTO vouchers (id, date, description)
		    VALUES ($1, $2, $3)`
	voucherId := uuid.New()

	_, err = tx.Exec(
		ctx,
		createVoucherQuery,
		voucherId,
		invoice.Date,
		fmt.Sprintf(" دریافت فاکتور فروش  شماره %s", invoice.Number.String),
	)
	if err != nil {
		tx.Rollback(ctx)

		return err
	}

	if bank > 0 {
		var accountId string

		accountSearchQuery := `
			SELECT
			    account_id
			FROM
			    banks
			WHERE
			    id = $1`

		err = tx.QueryRow(ctx, accountSearchQuery, rfp.BankId).Scan(&accountId)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}

		id := uuid.New()
		bankVoucherItemQuery := `
			INSERT INTO voucher_items (id, account_id, debit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			bankVoucherItemQuery,
			id,
			accountId,
			rfp.Bank,
			voucherId,
			fmt.Sprintf("پرداخت فاکتور فروش  شماره %s", invoice.Number.String),
		)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}
	}

	if cash > 0 {
		id := uuid.New()

		var accountId string

		accountSearchQuery := `
			SELECT
			    id
			FROM
			    accounts
			WHERE
			    code = $1`

		err = tx.QueryRow(ctx, accountSearchQuery, "110101").Scan(&accountId)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}

		cashVoucherItemQuery := `
			INSERT INTO voucher_items (id, account_id, debit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			cashVoucherItemQuery,
			id,
			accountId,
			rfp.Cash,
			voucherId,
			fmt.Sprintf("پرداخت فاکتور فروش  شماره %s", invoice.Number.String),
		)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}
	}

	personBalance := (cash + bank)
	if personBalance > 0 {
		var personId string

		personSearchQuery := `
			SELECT
			    person_id
			FROM
			    invoices
			WHERE
			    id = $1`

		err = tx.QueryRow(ctx, personSearchQuery, rfp.InvoiceId).Scan(&personId)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}

		id := uuid.New()

		personBalance := strconv.Itoa(cash + bank)
		personIdVoucherItemQuery := `
			INSERT INTO voucher_items (id, person_id, credit, voucher_id, description)
			    VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(
			ctx,
			personIdVoucherItemQuery,
			id,
			personId,
			personBalance,
			voucherId,
			fmt.Sprintf("پرداخت فاکتور فروش  شماره %s", invoice.Number.String),
		)
		if err != nil {
			tx.Rollback(ctx)

			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) BankToBankOperation(btp BankToBank, ctx context.Context) error {
	voucherId := uuid.New()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var originBank Bank

	err = tx.QueryRow(ctx, `
		SELECT
		    name,
		    branch,
		    number,
		    account_id
		FROM
		    banks
		WHERE
		    id = $1`, btp.OriginBankId).
		Scan(&originBank.Name, &originBank.Branch, &originBank.Number, &originBank.AccountId)
	if err != nil {
		return err
	}

	var destinationBank Bank

	err = tx.QueryRow(ctx, `
		SELECT
		    name,
		    branch,
		    number,
		    account_id
		FROM
		    banks
		WHERE
		    id = $1`, btp.DestinationBankId).
		Scan(&destinationBank.Name, &destinationBank.Branch, &destinationBank.Number, &destinationBank.AccountId)
	if err != nil {
		return err
	}

	createVoucherQuery := `
		INSERT INTO vouchers (id, date, description)
		    VALUES ($1, $2, $3)`

	_, err = tx.Exec(ctx, createVoucherQuery, voucherId, btp.Date, btp.Description)
	if err != nil {
		return err
	}

	voucherItemId := uuid.New()

	originBankQuery := `
		INSERT INTO voucher_items (id, voucher_id, account_id, credit, description)
		    VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(
		ctx,
		originBankQuery,
		voucherItemId,
		voucherId,
		originBank.AccountId,
		btp.Total,
		fmt.Sprintf(
			"پرداخت به %s %s %s",
			destinationBank.Name.String,
			destinationBank.Branch.String,
			destinationBank.Number.String,
		),
	)
	if err != nil {
		return err
	}

	voucherItemId = uuid.New()

	destinationBankQuery := `
		INSERT INTO voucher_items (id, voucher_id, account_id, debit, description)
		    VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(
		ctx,
		destinationBankQuery,
		voucherItemId,
		voucherId,
		destinationBank.AccountId,
		btp.Total,
		fmt.Sprintf(
			"دریافت از %s %s %s",
			originBank.Name.String,
			originBank.Branch.String,
			originBank.Number.String,
		),
	)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CashToBankOperation(ctb CashToBank, ctx context.Context) error {
	voucherId := uuid.New()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var bank Bank

	err = tx.QueryRow(ctx, `
		SELECT
		    name,
		    branch,
		    number,
		    account_id
		FROM
		    banks
		WHERE
		    id = $1`, ctb.BankId).
		Scan(&bank.Name, &bank.Branch, &bank.Number, &bank.AccountId)
	if err != nil {
		return err
	}

	createVoucherQuery := `
		INSERT INTO vouchers (id, date, description)
		    VALUES ($1, $2, $3)`

	_, err = tx.Exec(ctx, createVoucherQuery, voucherId, ctb.Date, ctb.Description)
	if err != nil {
		return err
	}

	voucherItemId := uuid.New()

	bankQuery := `
		INSERT INTO voucher_items (id, voucher_id, account_id, debit, description)
		    VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(
		ctx,
		bankQuery,
		voucherItemId,
		voucherId,
		bank.AccountId,
		ctb.Total,
		"واریز نقدی از صندوق",
	)
	if err != nil {
		return err
	}

	var accountId string

	searchAccountIdQuery := `
		SELECT
		    id
		FROM
		    accounts
		WHERE
		    code = '110101'`

	err = tx.QueryRow(ctx, searchAccountIdQuery).Scan(&accountId)
	if err != nil {
		return err
	}

	voucherItemId = uuid.New()

	cashQuery := `
		INSERT INTO voucher_items (id, voucher_id, account_id, credit, description)
		    VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(
		ctx,
		cashQuery,
		voucherItemId,
		voucherId,
		accountId,
		ctb.Total,
		fmt.Sprintf(
			"واریز به %s %s %s",
			bank.Name.String,
			bank.Branch.String,
			bank.Number.String,
		),
	)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) BankToCashOperation(btc BankToCash, ctx context.Context) error {
	voucherId := uuid.New()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var bank Bank

	err = tx.QueryRow(ctx, `
		SELECT
		    name,
		    branch,
		    number,
		    account_id
		FROM
		    banks
		WHERE
		    id = $1`, btc.BankId).
		Scan(&bank.Name, &bank.Branch, &bank.Number, &bank.AccountId)
	if err != nil {
		return err
	}

	createVoucherQuery := `
		INSERT INTO vouchers (id, date, description)
		    VALUES ($1, $2, $3)`

	_, err = tx.Exec(ctx, createVoucherQuery, voucherId, btc.Date, btc.Description)
	if err != nil {
		return err
	}

	voucherItemId := uuid.New()

	bankQuery := `
		INSERT INTO voucher_items (id, voucher_id, account_id, credit, description)
		    VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(
		ctx,
		bankQuery,
		voucherItemId,
		voucherId,
		bank.AccountId,
		btc.Total,
		"پرداخت نقدی به صندوق",
	)
	if err != nil {
		return err
	}

	var accountId string

	searchAccountIdQuery := `
		SELECT
		    id
		FROM
		    accounts
		WHERE
		    code = '110101'`

	err = tx.QueryRow(ctx, searchAccountIdQuery).Scan(&accountId)
	if err != nil {
		return err
	}

	voucherItemId = uuid.New()

	cashQuery := `
		INSERT INTO voucher_items (id, voucher_id, account_id, debit, description)
		    VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(
		ctx,
		cashQuery,
		voucherItemId,
		voucherId,
		accountId,
		btc.Total,
		fmt.Sprintf(
			"برداشت نقد از %s %s %s",
			bank.Name.String,
			bank.Branch.String,
			bank.Number.String,
		),
	)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) BankToCostOperation(btc BankToCost, ctx context.Context) error {
	voucherId := uuid.New()

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var bank Bank

	err = tx.QueryRow(ctx, `
		SELECT
		    name,
		    branch,
		    number,
		    account_id
		FROM
		    banks
		WHERE
		    id = $1`, btc.BankId).
		Scan(&bank.Name, &bank.Branch, &bank.Number, &bank.AccountId)
	if err != nil {
		return err
	}

	createVoucherQuery := `
		INSERT INTO vouchers (id, date, description)
		    VALUES ($1, $2, $3)`

	_, err = tx.Exec(ctx, createVoucherQuery, voucherId, btc.Date, btc.Description)
	if err != nil {
		return err
	}

	var account Account

	searchAccountQuery := `
		SELECT
		    id,
		    name,
		    code
		FROM
		    accounts
		WHERE
		    id = $1`

	err = tx.QueryRow(ctx, searchAccountQuery, btc.AccountId).
		Scan(&account.ID, &account.Name, &account.Code)
	if err != nil {
		return err
	}

	voucherItemId := uuid.New()

	bankQuery := `
		INSERT INTO voucher_items (id, voucher_id, account_id, credit, description)
		    VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(
		ctx,
		bankQuery,
		voucherItemId,
		voucherId,
		bank.AccountId,
		btc.Total,
		fmt.Sprintf("پرداخت به %s", account.Name.String),
	)
	if err != nil {
		return err
	}

	voucherItemId = uuid.New()

	costAccountQuery := `
		INSERT INTO voucher_items (id, voucher_id, account_id, debit, description)
		    VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(
		ctx,
		costAccountQuery,
		voucherItemId,
		voucherId,
		btc.AccountId,
		btc.Total,
		"")
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
