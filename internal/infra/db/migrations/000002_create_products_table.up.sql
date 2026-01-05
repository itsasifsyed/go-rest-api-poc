-- Create products table with audit columns
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    
    -- Audit columns
    created_by UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_by UUID,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_by UUID,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create index on name for faster searches
CREATE INDEX idx_products_name ON products(name) WHERE deleted_at IS NULL;

-- Create index on deleted_at for soft delete queries
CREATE INDEX idx_products_deleted_at ON products(deleted_at);

-- Create index on price for range queries
CREATE INDEX idx_products_price ON products(price) WHERE deleted_at IS NULL;

-- Add trigger to automatically update updated_at timestamp
CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

