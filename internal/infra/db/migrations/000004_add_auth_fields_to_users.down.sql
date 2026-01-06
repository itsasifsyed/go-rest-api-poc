-- Remove indexes
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_email;

-- Remove authentication and authorization fields from users table
ALTER TABLE users DROP COLUMN IF EXISTS blocked_by;
ALTER TABLE users DROP COLUMN IF EXISTS blocked_at;
ALTER TABLE users DROP COLUMN IF EXISTS is_blocked;
ALTER TABLE users DROP COLUMN IF EXISTS is_active;
ALTER TABLE users DROP COLUMN IF EXISTS role_id;
ALTER TABLE users DROP COLUMN IF EXISTS password;

