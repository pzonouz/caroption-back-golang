package services

import (
	"context"
	"fmt"
	"time"
)

func (s *Service) GetProductActivitiesDetail(
	productId string,
	ctx context.Context,
	fromDate time.Time, toDate time.Time,
) ([]ProductActivitiesDetail, error) {
	query := `
		SELECT 
    buy_price,
    in_count,
    sell_price,
    out_count,
    description,
    voucher_number,
    invoice_number,
    voucher_date,
    count
FROM (
    SELECT 
        p.id AS product_id,
        p.buy_price,
        p.count AS in_count,
        p.sell_price,
        '0' AS out_count,                                   
        'اول دوره' AS description,
        '1' AS voucher_number,
	      NULL AS invoice_number,
        (SELECT date FROM vouchers WHERE vouchers.voucher_number = '1') AS voucher_date,
        COALESCE(p.count,'0') AS count
    FROM products AS p

    UNION ALL

    SELECT
        p.id AS product_id,
        CASE WHEN i.type='BUY'  THEN ii.price ELSE '0' END AS buy_price,  
        CASE WHEN i.type='BUY'  THEN ii.count ELSE '0' END AS in_count,   
        CASE WHEN i.type='SELL' THEN ii.price ELSE '0' END AS sell_price, 
        CASE WHEN i.type='SELL' THEN ii.count ELSE '0' END AS out_count,  

        CASE 
            WHEN i.type='BUY' 
                THEN CONCAT('فاکتور خرید ', i.number, ' از ', pe.name)
            ELSE 
                CONCAT('فاکتور فروش ', i.number, ' به ', pe.name)
        END AS description,

        v.voucher_number::text AS voucher_number,
        i.number AS invoice_number,
        v.date AS voucher_date,

    CAST(
	COALESCE(p.count,'0')::bigint +
    SUM(
        CASE WHEN i.type='SELL' THEN -(COALESCE(ii.count,'0'))::bigint ELSE 0 END +
        CASE WHEN i.type='BUY'  THEN  COALESCE(ii.count,'0')::bigint ELSE 0 END
    ) OVER (PARTITION BY p.id ORDER BY v.date, v.voucher_number)
    AS text) AS count
    FROM invoices i
    LEFT JOIN invoice_items ii ON ii.invoice_id = i.id
    LEFT JOIN products p ON ii.product_id = p.id
    LEFT JOIN vouchers v ON i.voucher_id = v.id
    LEFT JOIN persons pe ON i.person_id = pe.id
) t
WHERE t.product_id = $1
  AND t.voucher_date >= $2
  AND t.voucher_date <= $3
	ORDER BY t.voucher_date::date,voucher_number::bigint;
`

	formattedFromDate := fromDate.Format("2006-01-02")

	toDateTomorrow := toDate.AddDate(0, 0, 1)
	formattedToDate := toDateTomorrow.Format("2006-01-02")

	rows, err := s.db.Query(ctx, query, productId, formattedFromDate, formattedToDate)
	if err != nil {
		return []ProductActivitiesDetail{}, err
	}

	defer rows.Close()

	var productPriceActivitiesDetailsRows []ProductActivitiesDetail

	for rows.Next() {
		var productPriceActivitiesDetailsRow ProductActivitiesDetail

		err := rows.Scan(
			&productPriceActivitiesDetailsRow.InPrice,
			&productPriceActivitiesDetailsRow.InCount,
			&productPriceActivitiesDetailsRow.OutPrice,
			&productPriceActivitiesDetailsRow.OutCount,
			&productPriceActivitiesDetailsRow.Description,
			&productPriceActivitiesDetailsRow.VoucherNumber,
			&productPriceActivitiesDetailsRow.InvoiceNumber,
			&productPriceActivitiesDetailsRow.VoucherDate,
			&productPriceActivitiesDetailsRow.Count,
		)
		if err != nil {
			return nil, err
		}

		productPriceActivitiesDetailsRows = append(
			productPriceActivitiesDetailsRows,
			productPriceActivitiesDetailsRow,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return productPriceActivitiesDetailsRows, nil
}

func (s *Service) GetPersonActivitiesDetail(
	personId string,
	ctx context.Context,
	fromDate time.Time, toDate time.Time,
) ([]PersonActivitiesDetail, error) {
	query := `SELECT 
    debit,credit,description,voucher_number,voucher_date,

    -- Net running balance (can be negative internally)
    net_balance,

    -- Always positive displayed balance
    ABS(net_balance) AS balance,

    -- Balance type
    CASE 
        WHEN net_balance > 0 THEN 'debit'
        WHEN net_balance < 0 THEN 'credit'
        ELSE 'zero'
    END AS balance_type

FROM (
    SELECT 
        t.*,
        SUM(
	          COALESCE(REPLACE(t.debit,',',''),'0')::bigint -
	          COALESCE(REPLACE(t.credit,',',''),'0')::bigint) 
	OVER (PARTITION BY t.person_id ORDER BY t.voucher_date::date,t.voucher_number::bigint
	) AS net_balance
    FROM (
        SELECT 
            p.debit,
            p.credit,
            'مانده اول دوره' AS description,
            p.id AS person_id,
            '1' AS voucher_number,
            vo.date as voucher_date
        FROM persons as p
				JOIN vouchers as vo ON vo.voucher_number='1'

        UNION ALL

        SELECT
            vi.debit,
            vi.credit,
            vi.description,
            vi.person_id,
            v.voucher_number,
            v.date AS voucher_date
        FROM voucher_items vi
        LEFT JOIN vouchers v ON vi.voucher_id = v.id
    ) t
) t
WHERE t.person_id = $1 AND t.voucher_date >= $2 AND t.voucher_date <= $3
	ORDER BY t.voucher_date::date,voucher_number::bigint;
`

	formattedFromDate := fromDate.Format("2006-01-02")
	toDateTomorrow := toDate.AddDate(0, 0, 1)
	formattedToDate := toDateTomorrow.Format("2006-01-02")

	rows, err := s.db.Query(ctx, query, personId, formattedFromDate, formattedToDate)
	if err != nil {
		return []PersonActivitiesDetail{}, err
	}

	defer rows.Close()

	var personActivitiesDetailRows []PersonActivitiesDetail

	for rows.Next() {
		var personActivitiesDetailRow PersonActivitiesDetail

		err := rows.Scan(
			&personActivitiesDetailRow.Debit,
			&personActivitiesDetailRow.Credit,
			&personActivitiesDetailRow.Description,
			&personActivitiesDetailRow.VoucherNumber,
			&personActivitiesDetailRow.VoucherDate,
			&personActivitiesDetailRow.NetBalance,
			&personActivitiesDetailRow.Balance,
			&personActivitiesDetailRow.BalanceType,
		)
		if err != nil {
			return nil, err
		}

		personActivitiesDetailRows = append(personActivitiesDetailRows, personActivitiesDetailRow)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return personActivitiesDetailRows, nil
}

func (s *Service) GetBankActivitiesDetail(
	bankId string,
	fromDate time.Time, toDate time.Time,
	ctx context.Context,
) ([]BankActivitiesDetail, error) {
	query := `SELECT 
    debit,credit,description,voucher_number,voucher_date,bank_id,

    -- Net running balance (can be negative internally)
    net_balance,

    -- Always positive displayed balance
    ABS(net_balance) AS balance,

    -- Balance type
    CASE 
        WHEN net_balance > 0 THEN 'debit'
        WHEN net_balance < 0 THEN 'credit'
        ELSE 'zero'
    END AS balance_type

FROM (
    SELECT 
        t.*,
        SUM(REPLACE(t.debit,',','')::bigint - REPLACE(t.credit,',','')::bigint) 
	OVER (PARTITION BY t.bank_id ORDER BY voucher_date::date,voucher_number::bigint) AS net_balance
    FROM (
        (SELECT 
            b.first_balance as debit,
            '0' as credit,
            'مانده اول دوره' AS description,
            '1' AS voucher_number,
            vo.date as voucher_date,
						b.id AS bank_id
        FROM banks as b
				JOIN vouchers as vo ON vo.voucher_number='1'
				WHERE b.id=$1	
      	LIMIT 1)

        UNION ALL

        SELECT
            vi.debit,
            vi.credit,
            vi.description,
            v.voucher_number,
            v.date AS voucher_date,
						b.id AS bank_id
        FROM voucher_items vi
        LEFT JOIN vouchers v ON vi.voucher_id = v.id
				LEFT JOIN accounts a ON a.id = vi.account_id
				LEFT JOIN banks b ON b.account_id = a.id
        WHERE v.voucher_number != '1' 
    ) t
) t
WHERE t.bank_id = $1 AND t.voucher_date >= $2 AND t.voucher_date <= $3
	ORDER BY t.voucher_date::date,voucher_number::bigint;
`
	formattedFromDate := fromDate.Format("2006-01-02")
	toDateTomorrow := toDate.AddDate(0, 0, 1)
	formattedToDate := toDateTomorrow.Format("2006-01-02")

	rows, err := s.db.Query(ctx, query, bankId, formattedFromDate, formattedToDate)
	if err != nil {
		return []BankActivitiesDetail{}, err
	}

	defer rows.Close()

	var bankActivitiesDetailRows []BankActivitiesDetail

	for rows.Next() {
		var bankActivitiesDetailRow BankActivitiesDetail

		err := rows.Scan(
			&bankActivitiesDetailRow.Debit,
			&bankActivitiesDetailRow.Credit,
			&bankActivitiesDetailRow.Description,
			&bankActivitiesDetailRow.VoucherNumber,
			&bankActivitiesDetailRow.VoucherDate,
			&bankActivitiesDetailRow.BankId,
			&bankActivitiesDetailRow.NetBalance,
			&bankActivitiesDetailRow.Balance,
			&bankActivitiesDetailRow.BalanceType,
		)
		if err != nil {
			return nil, err
		}

		bankActivitiesDetailRows = append(bankActivitiesDetailRows, bankActivitiesDetailRow)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bankActivitiesDetailRows, nil
}

func (s *Service) GetCashActivitiesDetail(
	ctx context.Context, code string, fromDate time.Time, toDate time.Time,
) ([]CashActivitiesDetail, error) {
	query := `
	SELECT 
    debit,credit,description,voucher_number,voucher_date,account_code,

    -- Net running balance (can be negative internally)
    net_balance,

    -- Always positive displayed balance
    ABS(net_balance) AS balance,

    -- Balance type
    CASE 
        WHEN net_balance > 0 THEN 'debit'
        WHEN net_balance < 0 THEN 'credit'
        ELSE 'zero'
    END AS balance_type

FROM (
    SELECT 
        t.*,
        SUM(REPLACE(t.debit,',','')::bigint - REPLACE(t.credit,',','')::bigint) 
            OVER (PARTITION BY t.account_code ORDER BY t.voucher_date::date, t.voucher_number::bigint) AS net_balance
    FROM (
        SELECT 
            vi.debit as debit,
            '0' as credit,
            'مانده اول دوره' AS description,
            '1' AS voucher_number,
            vo.date as voucher_date,
			'1' AS account_code
        FROM voucher_items as vi
				JOIN accounts as a ON vi.account_id=a.id ANd a.code=$1 
				JOIN vouchers as vo ON vo.voucher_number='1'

        UNION ALL

        SELECT
            vi.debit,
            vi.credit,
            vi.description,
            v.voucher_number,
            v.date AS voucher_date,
		    a.code AS account_code
        FROM voucher_items vi
        LEFT JOIN vouchers v ON vi.voucher_id = v.id
				LEFT JOIN accounts a ON a.id = vi.account_id
    ) t
) t
WHERE t.account_code = $1 AND t.voucher_date >= $2 AND t.voucher_date <= $3
	ORDER BY t.voucher_date::date,voucher_number::bigint;
`
	fmt.Println(query)
	formattedFromDate := fromDate.Format("2006-01-02")
	toDateTomorrow := toDate.AddDate(0, 0, 1)
	formattedToDate := toDateTomorrow.Format("2006-01-02")

	rows, err := s.db.Query(ctx, query, code, formattedFromDate, formattedToDate)
	if err != nil {
		return []CashActivitiesDetail{}, err
	}

	defer rows.Close()

	var cashActivitiesDetailRows []CashActivitiesDetail

	for rows.Next() {
		var cashActivitiesDetailRow CashActivitiesDetail

		err := rows.Scan(
			&cashActivitiesDetailRow.Debit,
			&cashActivitiesDetailRow.Credit,
			&cashActivitiesDetailRow.Description,
			&cashActivitiesDetailRow.VoucherNumber,
			&cashActivitiesDetailRow.VoucherDate,
			&cashActivitiesDetailRow.AccountCode,
			&cashActivitiesDetailRow.NetBalance,
			&cashActivitiesDetailRow.Balance,
			&cashActivitiesDetailRow.BalanceType,
		)
		if err != nil {
			return nil, err
		}

		cashActivitiesDetailRows = append(cashActivitiesDetailRows, cashActivitiesDetailRow)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cashActivitiesDetailRows, nil
}

func (s *Service) GetDailyActivitiesDetail(
	ctx context.Context, date time.Time,
) ([]VoucherForView, error) {
	query := `
	SELECT
    vouchers.id,
    vouchers.voucher_number,
	  vouchers.date,
    vouchers.description,
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
    vouchers.date >= $1 AND vouchers.date < $2
GROUP BY
    vouchers.id
ORDER BY vouchers.date,vouchers.voucher_number
	`
	formattedDate := date.Format("2006-01-02")
	nextDay := date.AddDate(0, 0, 1)
	formattedNextDay := nextDay.Format("2006-01-02")

	rows, err := s.db.Query(ctx, query, formattedDate, formattedNextDay)
	if err != nil {
		return []VoucherForView{}, err
	}

	defer rows.Close()

	var voucherForViews []VoucherForView

	for rows.Next() {
		var voucherForView VoucherForView

		err := rows.Scan(
			&voucherForView.ID,
			&voucherForView.VoucherNumber,
			&voucherForView.Date,
			&voucherForView.Description,
			&voucherForView.Items,
			&voucherForView.CreatedAt,
			&voucherForView.UpdatedAt,
		)
		if err != nil {
			return []VoucherForView{}, err
		}

		voucherForViews = append(voucherForViews, voucherForView)
	}

	if err := rows.Err(); err != nil {
		return []VoucherForView{}, err
	}

	return voucherForViews, nil
}
