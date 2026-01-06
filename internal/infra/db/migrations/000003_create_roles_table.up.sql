-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert default roles with fixed UUIDs
INSERT INTO roles (id, name, description) VALUES
    ('00000000-0000-0000-0000-000000000001', 'owner', 'Super user with full system access'),
    ('00000000-0000-0000-0000-000000000002', 'admin', 'Administrator with limited access (cannot modify owner)'),
    ('00000000-0000-0000-0000-000000000003', 'system', 'System user for automated tasks'),
    ('00000000-0000-0000-0000-000000000004', 'customer', 'Regular customer user')
ON CONFLICT (id) DO NOTHING;

