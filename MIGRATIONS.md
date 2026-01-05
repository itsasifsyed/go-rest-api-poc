# Database Migrations & Seeds

## Quick Start

```bash
# Just run your app - migrations and seeds run automatically
go run cmd/api/main.go
```

That's it! Your database is set up with tables and test data.

---

## Structure

```
internal/infra/db/
├── migrations/          # Schema changes (CREATE TABLE, ALTER, etc.)
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_create_products_table.up.sql
│   └── 000002_create_products_table.down.sql
└── seeds/              # Test data (dev/staging only)
    ├── 001_users.sql
    └── 002_products.sql
```

---

## Migrations vs Seeds

| | Migrations | Seeds |
|---|---|---|
| **Purpose** | Schema changes | Test data |
| **Runs in** | All environments | Dev/staging only |
| **Tracked** | Yes (versioned) | No |
| **Format** | `.up.sql` + `.down.sql` | `.sql` only |

---

## Creating a Migration

**1. Create files:**
```bash
touch internal/infra/db/migrations/000003_add_user_phone.up.sql
touch internal/infra/db/migrations/000003_add_user_phone.down.sql
```

**2. Write the change:**
```sql
-- 000003_add_user_phone.up.sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

**3. Write the rollback:**
```sql
-- 000003_add_user_phone.down.sql
ALTER TABLE users DROP COLUMN phone;
```

**4. Run:**
```bash
go run cmd/api/main.go  # Automatic!
```

---

## Creating a Seed

**1. Create file:**
```bash
touch internal/infra/db/seeds/003_categories.sql
```

**2. Write the data (use ON CONFLICT for idempotency):**
```sql
-- 003_categories.sql
INSERT INTO categories (id, name) VALUES
    ('cat-1', 'Electronics'),
    ('cat-2', 'Books')
ON CONFLICT (id) DO NOTHING;
```

**3. Run:**
```bash
go run cmd/api/main.go  # Automatic!
```

---

## Current Schema

### Users Table
- `id` (UUID, primary key)
- `first_name`, `last_name`, `email`
- Audit: `created_by`, `created_at`, `updated_by`, `updated_at`, `deleted_by`, `deleted_at`

### Products Table
- `id` (UUID, primary key)
- `name`, `price`
- Audit: `created_by`, `created_at`, `updated_by`, `updated_at`, `deleted_by`, `deleted_at`

**Note:** All tables have auto-update triggers for `updated_at`

---

## Environment Control

Seeds automatically skip in production:

```bash
ENV=dev go run cmd/api/main.go        # Seeds run ✅
ENV=production go run cmd/api/main.go # Seeds skip ❌
```

---

## Manual Commands

```bash
# Apply migrations manually
go run cmd/migrate/main.go up

# Rollback last migration (dev only!)
go run cmd/migrate/main.go down

# Check status
go run cmd/migrate/main.go status
```

---

## Checking Database

```bash
psql $DATABASE_URL

# View migration history
SELECT * FROM schema_migrations;

# View tables
\dt

# View data
SELECT * FROM users;
SELECT * FROM products;
```

---

## Best Practices

### ✅ DO
- Keep migrations small (one change per file)
- Use `ON CONFLICT DO NOTHING` in seeds
- Test both up and down migrations
- Put schema changes in migrations
- Put test data in seeds

### ❌ DON'T
- Modify applied migrations (create new ones)
- Put seed data in migrations
- Rollback in production (create forward migrations)
- Skip version numbers

---

## Examples

### Add a column
```sql
-- migrations/000003_add_status.up.sql
ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active';

-- migrations/000003_add_status.down.sql
ALTER TABLE users DROP COLUMN status;
```

### Add a table
```sql
-- migrations/000004_create_categories.up.sql
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL
);

-- migrations/000004_create_categories.down.sql
DROP TABLE categories;
```

### Seed data
```sql
-- seeds/003_more_users.sql
INSERT INTO users (id, first_name, last_name, email) VALUES
    ('uuid-here', 'Test', 'User', 'test@example.com')
ON CONFLICT (id) DO NOTHING;
```

---

## Troubleshooting

**Migrations not running?**
- Check `schema_migrations` table in database
- Ensure connection string is correct

**Seeds not running?**
- Check `ENV` is not set to `production`
- Seeds run every time (they're idempotent)

**Reset database (dev only):**
```bash
psql $DATABASE_URL -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
go run cmd/api/main.go
```

---

That's all you need to know. Keep it simple! 🚀
