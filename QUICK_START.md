# Quick Start Guide - JWT Authentication

## Prerequisites
- PostgreSQL running
- Go 1.24.5+
- Make

## Setup (5 minutes)

### 1. Configure Environment Variables

Add these to your `.env` file:

```bash
# JWT Authentication (REQUIRED)
JWT_SECRET=your-super-secret-key-change-in-production-min-32-chars
JWT_ISSUER=go-rest-api-poc
JWT_AUDIENCE=go-rest-api-poc
ACCESS_TOKEN_LIFETIME=15m
REFRESH_TOKEN_LIFETIME=168h
STAY_SIGNED_IN_LIFETIME=720h
PASSWORD_RESET_OTP_LIFETIME=15m
```

### 2. Run Database Migrations

```bash
make migrate-up
```

This creates:
- `roles` table (owner, admin, system, customer)
- `user_sessions` table (session tracking)
- `password_reset_tokens` table (password reset)
- Updates `users` table with auth fields

### 3. Seed Database

```bash
make seed
```

This creates 5 test users with password `password123`:
- owner@example.com (owner role)
- admin@example.com (admin role)
- system@example.com (system role)
- alice@example.com (customer role)
- bob@example.com (customer role)

### 4. Start Server

```bash
make run
```

Server starts on `http://localhost:8080`

## Quick Test (2 minutes)

### Using cURL:

```bash
# 1. Login
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "owner@example.com",
    "password": "password123",
    "stay_signed_in": false
  }'

# Copy the access_token from response

# 2. Get your user info
curl http://localhost:8080/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN_HERE"

# 3. List all users (protected route)
curl http://localhost:8080/v1/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN_HERE"
```

### Using Postman:

1. **Login:**
   - Method: POST
   - URL: `http://localhost:8080/v1/auth/login`
   - Body (JSON):
     ```json
     {
       "email": "owner@example.com",
       "password": "password123"
     }
     ```
   - Cookies are automatically saved!

2. **Access Protected Route:**
   - Method: GET
   - URL: `http://localhost:8080/v1/auth/me`
   - No headers needed (cookies work automatically)
   - OR use: `Authorization: Bearer <access_token>`

## What Changed?

### ✅ All routes now require authentication!

**Before:**
```bash
curl http://localhost:8080/v1/users  # ✅ Worked
```

**After:**
```bash
curl http://localhost:8080/v1/users  # ❌ 401 Unauthorized
```

You must login first and include the token.

### ✅ New auth endpoints available:

- `/v1/auth/login` - Login
- `/v1/auth/register` - Register
- `/v1/auth/logout` - Logout
- `/v1/auth/me` - Get current user
- `/v1/auth/sessions` - View all sessions
- `/v1/auth/block-user/:id` - Block user (admin only)
- And more! (See AUTH_TESTING_GUIDE.md)

## Key Features

✅ **Dual Token System** - Access (15 min) + Refresh (7-30 days)
✅ **Cookie + Bearer Token** - Works with browsers and API clients
✅ **Multi-Device Sessions** - Track and manage all login sessions
✅ **Instant User Blocking** - Invalidate all sessions immediately
✅ **Role-Based Access** - Owner, Admin, System, Customer roles
✅ **Password Reset** - OTP-based (logged to console)
✅ **Stay Signed In** - 30-day sessions option

## Common Issues

### "Missing authentication token"
**Solution:** Login first to get tokens

### "Token has expired"
**Solution:** Call `/v1/auth/refresh` to get new tokens

### "User account is blocked"
**Solution:** Contact admin to unblock

## Next Steps

1. **Read the full testing guide:** `AUTH_TESTING_GUIDE.md`
2. **Read implementation details:** `AUTH_IMPLEMENTATION_SUMMARY.md`
3. **Test all auth flows** with different users and roles
4. **Integrate email service** for password reset OTPs
5. **Add rate limiting** for production

## Need Help?

- Full testing guide: `AUTH_TESTING_GUIDE.md`
- Implementation details: `AUTH_IMPLEMENTATION_SUMMARY.md`
- Check server logs for OTPs and errors

## Test Users

| Email | Password | Role | Can Block Users? |
|-------|----------|------|------------------|
| owner@example.com | password123 | owner | ✅ Yes |
| admin@example.com | password123 | admin | ✅ Yes |
| system@example.com | password123 | system | ❌ No |
| alice@example.com | password123 | customer | ❌ No |
| bob@example.com | password123 | customer | ❌ No |

Happy coding! 🚀

