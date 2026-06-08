CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE OR REPLACE FUNCTION normalize_persian (text)
    RETURNS text
    AS $$
DECLARE
    txt text := $1;
BEGIN
    txt := replace(txt, 'ي', 'ی');
    txt := replace(txt, 'ك', 'ک');
    txt := replace(txt, 'ة', 'ه');
    txt := replace(txt, 'ۀ', 'ه');
    txt := replace(txt, 'ؤ', 'و');
    txt := replace(txt, 'إ', 'ا');
    txt := replace(txt, 'أ', 'ا');
    txt := replace(txt, 'آ', 'ا');
    txt := regexp_replace(txt, '[\u064B-\u065F]', '', 'g');
    txt := regexp_replace(txt, '\s+', ' ', 'g');
    RETURN trim(txt);
END;
$$
LANGUAGE plpgsql
IMMUTABLE;

ALTER TABLE products
    ADD COLUMN fts tsvector;

UPDATE
    products
SET
    fts = to_tsvector('simple', unaccent (normalize_persian (coalesce(name, '') || ' ' || coalesce(description, '') || ' ' || coalesce((
                    SELECT
                        string_agg(normalize_persian (coalesce(p.name, '') || ' ' || coalesce(ppv.text_value, '') || ' ' || coalesce(ppv.selectable_value, '')), ' ')
                    FROM product_parameter_values ppv
                    JOIN parameters p ON p.id = ppv.parameter_id
                    WHERE
                        ppv.product_id = products.id), ''))));

CREATE OR REPLACE FUNCTION update_product_fts ()
    RETURNS TRIGGER
    AS $$
BEGIN
    UPDATE
        products
    SET
        fts = to_tsvector('simple', unaccent (normalize_persian (coalesce(name, '') || ' ' || coalesce(description, '') || ' ' || coalesce((
                        SELECT
                            string_agg(normalize_persian (coalesce(p.name, '') || ' ' || coalesce(ppv.text_value, '') || ' ' || coalesce(ppv.selectable_value, '')), ' ')
                        FROM product_parameter_values ppv
                        JOIN parameters p ON p.id = ppv.parameter_id
                        WHERE
                            ppv.product_id = NEW.product_id), ''))))
    WHERE
        id = NEW.product_id;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trg_update_product_fts_on_ppv
    AFTER INSERT OR UPDATE OR DELETE ON product_parameter_values
    FOR EACH ROW
    EXECUTE FUNCTION update_product_fts ();

CREATE INDEX idx_products_fts ON products USING GIN (fts);

CREATE INDEX idx_products_trgm ON products USING gin (normalize_persian (name) gin_trgm_ops);

UPDATE
    products
SET
    fts = setweight(to_tsvector('simple', normalize_persian (coalesce(name, ''))), 'A') || setweight(to_tsvector('simple', normalize_persian (coalesce(description, ''))), 'B') || setweight(to_tsvector('simple', normalize_persian ((
                SELECT
                    string_agg(normalize_persian (coalesce(p.name, '') || ' ' || coalesce(ppv.text_value, '') || ' ' || coalesce(ppv.selectable_value, '')), ' ')
                FROM product_parameter_values ppv
                JOIN parameters p ON p.id = ppv.parameter_id
                WHERE
                    ppv.product_id = products.id))), 'C');

CREATE OR REPLACE FUNCTION convert_english_digits_to_persian ()
    RETURNS TRIGGER
    AS $$
BEGIN
    NEW.name := translate(NEW.name, '0123456789', '۰۱۲۳۴۵۶۷۸۹');
    NEW.slug := translate(NEW.slug, '0123456789', '۰۱۲۳۴۵۶۷۸۹');
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER convert_digits_before_insert
    BEFORE INSERT ON products
    FOR EACH ROW
    EXECUTE FUNCTION convert_english_digits_to_persian ();

UPDATE
    categories
SET
    priority = translate(priority, '۰۱۲۳۴۵۶۷۸۹', '0123456789');

CREATE OR REPLACE FUNCTION convert_persian_digits (text)
    RETURNS text
    AS $$
BEGIN
    RETURN translate($1, '۰۱۲۳۴۵۶۷۸۹', '0123456789');
END;
$$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION normalize_priority ()
    RETURNS TRIGGER
    AS $$
BEGIN
    NEW.priority := convert_persian_digits (NEW.priority);
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trigger_normalize_priority
    BEFORE INSERT OR UPDATE ON categories
    FOR EACH ROW
    EXECUTE FUNCTION normalize_priority ();

