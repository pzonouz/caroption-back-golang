-- =====================================================
-- Extensions
-- =====================================================
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =====================================================
-- Types
-- =====================================================
DO $ $ BEGIN IF NOT EXISTS (
    SELECT
        1
    FROM
        pg_type
    WHERE
        typname = 'voucher_status_enum'
) THEN CREATE TYPE voucher_status_enum AS ENUM ('Draft', 'Approved', 'Locked');

END IF;

IF NOT EXISTS (
    SELECT
        1
    FROM
        pg_type
    WHERE
        typname = 'transaction_type_enum'
) THEN CREATE TYPE transaction_type_enum AS ENUM ('Receipt', 'Payment', 'Transfer');

END IF;

IF NOT EXISTS (
    SELECT
        1
    FROM
        pg_type
    WHERE
        typname = 'payment_method_enum'
) THEN CREATE TYPE payment_method_enum AS ENUM ('Cash', 'Bank', 'Check');

END IF;

IF NOT EXISTS (
    SELECT
        1
    FROM
        pg_type
    WHERE
        typname = 'check_status_enum'
) THEN CREATE TYPE check_status_enum AS ENUM (
    'Registered',
    'In_Bank',
    'Cleared',
    'Bounced',
    'Returned'
);

END IF;

IF NOT EXISTS (
    SELECT
        1
    FROM
        pg_type
    WHERE
        typname = 'invoice_type'
) THEN CREATE TYPE invoice_type AS ENUM ('SELL', 'BUY');

END IF;

IF NOT EXISTS (
    SELECT
        1
    FROM
        pg_type
    WHERE
        typname = 'invoice_status_enum'
) THEN CREATE TYPE invoice_status_enum AS ENUM ('Draft', 'Issued', 'Paid', 'Cancelled');

END IF;

END $ $;

-- =====================================================
-- Tables
-- =====================================================
CREATE TABLE IF NOT EXISTS settings (
    id text PRIMARY KEY DEFAULT gen_random_uuid() :: text,
    key varchar(100) UNIQUE NOT NULL,
    value text
);

CREATE TABLE IF NOT EXISTS accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code varchar(20) UNIQUE NOT NULL,
    name varchar(100) UNIQUE NOT NULL,
    parent_id uuid NULL,
    locked boolean DEFAULT FALSE,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS persons (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name text,
    name text,
    address text,
    phone_number text UNIQUE,
    debit text DEFAULT '0',
    credit text DEFAULT '0',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS banks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text,
    branch text,
    number text,
    type text NOT NULL DEFAULT 'Jari',
    shaba text UNIQUE,
    first_balance text DEFAULT '0',
    balance text DEFAULT '0',
    account_id uuid NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text UNIQUE,
    description text,
    image_id uuid,
    priority varchar,
    parent_id uuid NULL,
    show boolean DEFAULT TRUE,
    slug text UNIQUE,
    generator boolean DEFAULT FALSE,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS entities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text UNIQUE,
    description text,
    image_id uuid,
    price text,
    priority varchar,
    parent_id uuid NULL,
    show boolean DEFAULT TRUE,
    keywords varchar [],
    entity_slug text UNIQUE,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS brands (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar UNIQUE,
    description text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS products (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text UNIQUE,
    description text,
    info text,
    price text,
    sell_price text,
    buy_price text,
    image_id uuid,
    entity_id uuid,
    category_id uuid,
    count text DEFAULT '0',
    is_service boolean DEFAULT FALSE,
    brand_id uuid,
    slug text UNIQUE,
    generated boolean DEFAULT FALSE,
    generatable boolean DEFAULT FALSE,
    keywords varchar [],
    show boolean DEFAULT TRUE,
    position varchar,
    code text UNIQUE,
    rank double precision,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS articles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar UNIQUE,
    description text,
    image_id uuid,
    slug text UNIQUE,
    show_in_products boolean DEFAULT FALSE,
    category_id uuid,
    keywords varchar [],
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar,
    image_url text,
    category_id uuid,
    product_id uuid,
    entity_id uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS parameter_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar UNIQUE,
    category_id uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS parameters (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar UNIQUE,
    description text,
    type varchar,
    parameter_group_id uuid,
    selectables varchar [],
    priority varchar DEFAULT '10000000',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS product_parameter_values (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid,
    parameter_id uuid,
    text_value varchar,
    bool_value boolean,
    selectable_value varchar,
    created_at timestamptz DEFAULT now(),
    UNIQUE (parameter_id, product_id)
);

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email varchar UNIQUE,
    password text,
    token text DEFAULT '',
    token_expires timestamptz,
    is_admin boolean DEFAULT FALSE,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS vouchers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text DEFAULT '',
    date timestamptz NOT NULL DEFAULT now(),
    voucher_number bigint GENERATED BY DEFAULT AS IDENTITY,
    description text,
    is_invoice boolean DEFAULT FALSE,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invoices (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    type invoice_type NOT NULL,
    person_id uuid NOT NULL,
    discount text DEFAULT '0',
    notes text DEFAULT '',
    description text NULL,
    date timestamptz DEFAULT now(),
    number bigint GENERATED BY DEFAULT AS IDENTITY,
    cleared boolean DEFAULT FALSE,
    total text DEFAULT '0',
    voucher_id uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invoice_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id uuid NOT NULL,
    product_id uuid NULL,
    price text NOT NULL,
    discount text NOT NULL DEFAULT '0',
    count text,
    description text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    UNIQUE (invoice_id, product_id)
);

CREATE TABLE IF NOT EXISTS voucher_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    voucher_id uuid NOT NULL,
    account_id uuid NULL,
    person_id uuid NULL,
    debit text DEFAULT '0',
    credit text DEFAULT '0',
    description varchar(255),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    CONSTRAINT check_positive_amounts CHECK (
        replace(coalesce(debit, '0'), ',', '') :: bigint >= 0
        AND replace(coalesce(credit, '0'), ',', '') :: bigint >= 0
    )
);

CREATE TABLE IF NOT EXISTS check_bands (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text,
    serial text,
    used_numbers text [],
    count text,
    start_number text,
    end_number text,
    bank_id uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS checks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    type text NOT NULL,
    check_number varchar(50) NOT NULL,
    bank_name varchar(100) NOT NULL,
    sayyad varchar(100) NOT NULL,
    due_date timestamptz NOT NULL,
    check_band_id uuid NULL,
    amount text NOT NULL,
    description text NULL,
    person_id uuid NOT NULL,
    status check_status_enum DEFAULT 'Registered',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE IF NOT EXISTS invoice_products_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text unique,
    product_ids text [],
    created_at timestamptz DEFAULT now()
);

-- =====================================================
-- Foreign keys
-- =====================================================
ALTER TABLE
    IF EXISTS accounts DROP CONSTRAINT IF EXISTS fk_accounts_parent;

ALTER TABLE
    IF EXISTS accounts
ADD
    CONSTRAINT fk_accounts_parent FOREIGN KEY (parent_id) REFERENCES accounts (id);

ALTER TABLE
    IF EXISTS banks DROP CONSTRAINT IF EXISTS fk_banks_account;

ALTER TABLE
    IF EXISTS banks
ADD
    CONSTRAINT fk_banks_account FOREIGN KEY (account_id) REFERENCES accounts (id);

ALTER TABLE
    IF EXISTS categories DROP CONSTRAINT IF EXISTS fk_categories_parent;

ALTER TABLE
    IF EXISTS categories
ADD
    CONSTRAINT fk_categories_parent FOREIGN KEY (parent_id) REFERENCES categories (id) ON DELETE RESTRICT;

ALTER TABLE
    IF EXISTS entities DROP CONSTRAINT IF EXISTS fk_entities_parent;

ALTER TABLE
    IF EXISTS entities
ADD
    CONSTRAINT fk_entities_parent FOREIGN KEY (parent_id) REFERENCES entities (id) ON DELETE RESTRICT;

ALTER TABLE
    IF EXISTS products DROP CONSTRAINT IF EXISTS fk_products_brand;

ALTER TABLE
    IF EXISTS products
ADD
    CONSTRAINT fk_products_brand FOREIGN KEY (brand_id) REFERENCES brands (id) ON DELETE
SET
    NULL ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS products DROP CONSTRAINT IF EXISTS fk_products_category;

ALTER TABLE
    IF EXISTS products
ADD
    CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE
SET
    NULL ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS products DROP CONSTRAINT IF EXISTS fk_products_entity;

ALTER TABLE
    IF EXISTS products
ADD
    CONSTRAINT fk_products_entity FOREIGN KEY (entity_id) REFERENCES entities (id) ON DELETE
SET
    NULL ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS articles DROP CONSTRAINT IF EXISTS fk_articles_category;

ALTER TABLE
    IF EXISTS articles
ADD
    CONSTRAINT fk_articles_category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE
SET
    NULL ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS parameter_groups DROP CONSTRAINT IF EXISTS fk_parameter_groups_category;

ALTER TABLE
    IF EXISTS parameter_groups
ADD
    CONSTRAINT fk_parameter_groups_category FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS parameters DROP CONSTRAINT IF EXISTS fk_parameters_group;

ALTER TABLE
    IF EXISTS parameters
ADD
    CONSTRAINT fk_parameters_group FOREIGN KEY (parameter_group_id) REFERENCES parameter_groups (id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS product_parameter_values DROP CONSTRAINT IF EXISTS fk_ppv_parameter;

ALTER TABLE
    IF EXISTS product_parameter_values
ADD
    CONSTRAINT fk_ppv_parameter FOREIGN KEY (parameter_id) REFERENCES parameters (id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS product_parameter_values DROP CONSTRAINT IF EXISTS fk_ppv_product;

ALTER TABLE
    IF EXISTS product_parameter_values
ADD
    CONSTRAINT fk_ppv_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE
    IF EXISTS invoices DROP CONSTRAINT IF EXISTS invoices_person_fk;

ALTER TABLE
    IF EXISTS invoices
ADD
    CONSTRAINT invoices_person_fk FOREIGN KEY (person_id) REFERENCES persons (id) ON DELETE RESTRICT;

ALTER TABLE
    IF EXISTS invoices DROP CONSTRAINT IF EXISTS invoices_voucher_fk;

ALTER TABLE
    IF EXISTS invoices
ADD
    CONSTRAINT invoices_voucher_fk FOREIGN KEY (voucher_id) REFERENCES vouchers (id) ON DELETE RESTRICT;

ALTER TABLE
    IF EXISTS invoice_items DROP CONSTRAINT IF EXISTS invoice_items_product_fk;

ALTER TABLE
    IF EXISTS invoice_items
ADD
    CONSTRAINT invoice_items_product_fk FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE RESTRICT;

ALTER TABLE
    IF EXISTS invoice_items DROP CONSTRAINT IF EXISTS invoice_items_invoice_fk;

ALTER TABLE
    IF EXISTS invoice_items
ADD
    CONSTRAINT invoice_items_invoice_fk FOREIGN KEY (invoice_id) REFERENCES invoices (id) ON DELETE CASCADE;

ALTER TABLE
    IF EXISTS voucher_items DROP CONSTRAINT IF EXISTS fk_voucher_items_voucher;

ALTER TABLE
    IF EXISTS voucher_items
ADD
    CONSTRAINT fk_voucher_items_voucher FOREIGN KEY (voucher_id) REFERENCES vouchers (id) ON DELETE CASCADE;

ALTER TABLE
    IF EXISTS voucher_items DROP CONSTRAINT IF EXISTS fk_voucher_items_account;

ALTER TABLE
    IF EXISTS voucher_items
ADD
    CONSTRAINT fk_voucher_items_account FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE
SET
    NULL;

ALTER TABLE
    IF EXISTS voucher_items DROP CONSTRAINT IF EXISTS fk_voucher_items_person;

ALTER TABLE
    IF EXISTS voucher_items
ADD
    CONSTRAINT fk_voucher_items_person FOREIGN KEY (person_id) REFERENCES persons (id) ON DELETE
SET
    NULL;

ALTER TABLE
    IF EXISTS checks DROP CONSTRAINT IF EXISTS fk_checks_person;

ALTER TABLE
    IF EXISTS checks
ADD
    CONSTRAINT fk_checks_person FOREIGN KEY (person_id) REFERENCES persons (id);

ALTER TABLE
    IF EXISTS checks DROP CONSTRAINT IF EXISTS fk_checks_check_band;

ALTER TABLE
    IF EXISTS checks
ADD
    CONSTRAINT fk_checks_check_band FOREIGN KEY (check_band_id) REFERENCES check_bands (id) ON DELETE
SET
    NULL;

ALTER TABLE
    IF EXISTS check_bands DROP CONSTRAINT IF EXISTS fk_check_bands_bank;

ALTER TABLE
    IF EXISTS check_bands
ADD
    CONSTRAINT fk_check_bands_bank FOREIGN KEY (bank_id) REFERENCES banks (id);

-- =====================================================
-- Seed data used by services
-- =====================================================
INSERT INTO
    settings (key, value)
VALUES
    ('AVVAL_DOREH', 'YES') ON CONFLICT (key) DO NOTHING;

-- =====================================================
-- Triggers and functions
-- =====================================================
CREATE
OR REPLACE FUNCTION prevent_bad_insert() RETURNS TRIGGER AS $ $ DECLARE v_date date;

BEGIN
SELECT
    date INTO v_date
FROM
    vouchers
WHERE
    voucher_number = 1;

IF v_date IS NOT NULL
AND NEW.date < v_date + INTERVAL '1 day' THEN RAISE EXCEPTION 'Cannot insert row: created date (%) is before voucher 1 date (%)',
NEW.date,
v_date;

END IF;

RETURN NEW;

END;

$ $ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_prevent_bad_insert ON invoices;

CREATE TRIGGER trg_prevent_bad_insert BEFORE
INSERT
    ON invoices FOR EACH ROW EXECUTE FUNCTION prevent_bad_insert();

DROP TRIGGER IF EXISTS trg_prevent_bad_insert ON vouchers;

CREATE TRIGGER trg_prevent_bad_insert BEFORE
INSERT
    ON vouchers FOR EACH ROW EXECUTE FUNCTION prevent_bad_insert();

CREATE
OR REPLACE FUNCTION check_voucher_date_order() RETURNS TRIGGER AS $ $ BEGIN IF EXISTS (
    SELECT
        1
    FROM
        vouchers v
    WHERE
        v.voucher_number < NEW.voucher_number
        AND v.date > NEW.date
) THEN RAISE EXCEPTION 'voucher date cannot be earlier than first voucher';

END IF;

RETURN NEW;

END;

$ $ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_check_voucher_date_order ON vouchers;

CREATE TRIGGER trg_check_voucher_date_order BEFORE
INSERT
    OR
UPDATE
    ON vouchers FOR EACH ROW EXECUTE FUNCTION check_voucher_date_order();