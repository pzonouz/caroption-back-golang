WITH
-- ---------------------------
-- 1. Invoices summary
-- ---------------------------
invoice_events AS (
    SELECT
        i.person_id,
        i.date AS event_date,
        'INVOICE' AS event_type,
        i.id AS event_id,
        i.type AS invoice_type,
        i.number AS invoice_number,
        i.total AS amount,
        NULL::uuid AS product_id,
        NULL::text AS product_name,
        NULL::text AS qty,
        NULL::text AS price,
        NULL::text AS discount,
        NULL::uuid AS voucher_id,
        NULL::uuid AS check_id,
        NULL::uuid AS transaction_id,
        i.description AS description
    FROM
        invoices i
),
-- ---------------------------
-- 2. Invoice items
-- ---------------------------
invoice_item_events AS (
    SELECT
        inv.person_id,
        inv.date AS event_date,
        'INVOICE_ITEM' AS event_type,
        ii.id AS event_id,
        inv.type AS invoice_type,
        inv.number AS invoice_number,
        NULL AS amount,
        ii.product_id,
        p.name AS product_name,
        ii.quantity AS qty,
        ii.price,
        ii.discount,
        NULL::uuid AS voucher_id,
        NULL::uuid AS check_id,
        NULL::uuid AS transaction_id,
        ii.description
    FROM
        invoice_items ii
        JOIN invoices inv ON inv.id = ii.invoice_id
        LEFT JOIN products p ON p.id = ii.product_id
),
-- ---------------------------
-- 3. Voucher entries related to this person
-- ---------------------------
voucher_events AS (
    SELECT
        vi.person_id,
        v.date AS event_date,
        'VOUCHER' AS event_type,
        vi.id AS event_id,
        NULL AS invoice_type,
        NULL AS invoice_number,
        CASE WHEN REPLACE(vi.debit, ',', '')::bigint > 0 THEN
            vi.debit
        ELSE
            vi.credit
        END AS amount,
        NULL::uuid AS product_id,
        NULL::text AS product_name,
        NULL::text AS qty,
        NULL::text AS price,
        NULL::text AS discount,
        v.id AS voucher_id,
        NULL::uuid AS check_id,
        NULL::uuid AS transaction_id,
        vi.description
    FROM
        voucher_items vi
        JOIN vouchers v ON v.id = vi.voucher_id
    WHERE
        vi.person_id IS NOT NULL
),
-- ---------------------------
-- 4. Checks
-- ---------------------------
check_events AS (
    SELECT
        c.person_id,
        c.due_date AS event_date,
        'CHECK' AS event_type,
        c.id AS event_id,
        c.type AS check_type,
        NULL AS invoice_number,
        c.amount AS amount,
        NULL AS product_id,
        NULL AS product_name,
        NULL AS qty,
        NULL AS price,
        NULL AS discount,
        NULL AS voucher_id,
        c.id AS check_id,
        NULL::uuid AS transaction_id,
        c.description
    FROM
        checks c
),
-- ---------------------------
-- Final UNION
-- ---------------------------
SELECT
    *
FROM (
    SELECT
        *
    FROM
        invoice_events
    UNION ALL
    SELECT
        *
    FROM
        invoice_item_events
    UNION ALL
    SELECT
        *
    FROM
        voucher_events
    UNION ALL
    SELECT
        *
    FROM
        check_events
) AS all_events
WHERE
    person_id = $1 -- person ID HERE
ORDER BY
    event_date,
    event_type;

