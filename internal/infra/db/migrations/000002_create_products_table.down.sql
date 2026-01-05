-- Drop trigger
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

-- Drop indexes
DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_deleted_at;
DROP INDEX IF EXISTS idx_products_name;

-- Drop table
DROP TABLE IF EXISTS products;

