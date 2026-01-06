-- Seed users table with role-based users for development
-- Password for all users: "password123" (bcrypt hashed)
-- Hash generated with: bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

INSERT INTO users (id, first_name, last_name, email, password, role_id, is_active, is_blocked, created_by, created_at, updated_at) VALUES
    -- Owner user (super user)
    ('550e8400-e29b-41d4-a716-446655440001', 'Bruce', 'Wayne', 'bruce.wayne@example.com', '$2a$10$CY3wFpNeNO.869l4yJs3cOJ5BHSnNx8.kXldTWSKPpfrJv3CSLow.', '00000000-0000-0000-0000-000000000001', true, false, '550e8400-e29b-41d4-a716-446655440001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    
    -- Admin users
    ('550e8400-e29b-41d4-a716-446655440002', 'John', 'Wick', 'john.wick@example.com', '$2a$10$CY3wFpNeNO.869l4yJs3cOJ5BHSnNx8.kXldTWSKPpfrJv3CSLow.', '00000000-0000-0000-0000-000000000002', true, false, '550e8400-e29b-41d4-a716-446655440001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    
    -- System user (for automated tasks)
    ('550e8400-e29b-41d4-a716-446655440003', 'System', 'Bot', 'system@example.com', '$2a$10$CY3wFpNeNO.869l4yJs3cOJ5BHSnNx8.kXldTWSKPpfrJv3CSLow.', '00000000-0000-0000-0000-000000000003', true, false, '550e8400-e29b-41d4-a716-446655440001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    
    -- Customer users
    ('550e8400-e29b-41d4-a716-446655440004', 'Alice', 'Customer', 'alice@example.com', '$2a$10$CY3wFpNeNO.869l4yJs3cOJ5BHSnNx8.kXldTWSKPpfrJv3CSLow.', '00000000-0000-0000-0000-000000000004', true, false, '550e8400-e29b-41d4-a716-446655440001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('550e8400-e29b-41d4-a716-446655440005', 'Bob', 'Customer', 'bob@example.com', '$2a$10$CY3wFpNeNO.869l4yJs3cOJ5BHSnNx8.kXldTWSKPpfrJv3CSLow.', '00000000-0000-0000-0000-000000000004', true, false, '550e8400-e29b-41d4-a716-446655440001', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;
