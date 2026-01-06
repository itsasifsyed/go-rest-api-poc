# JWT Authentication System - Testing Guide

## Prerequisites

1. Ensure PostgreSQL is running
2. Set up environment variables (copy from .env.example if needed)
3. Run migrations: `make migrate-up`
4. Seed database: `make seed`

## Environment Variables Required

```bash
JWT_SECRET=your-super-secret-key-change-in-production-min-32-chars
JWT_ISSUER=go-rest-api-poc
JWT_AUDIENCE=go-rest-api-poc
ACCESS_TOKEN_LIFETIME=15m
REFRESH_TOKEN_LIFETIME=168h
STAY_SIGNED_IN_LIFETIME=720h
PASSWORD_RESET_OTP_LIFETIME=15m
```

## Seeded Users

After running migrations and seeds, you'll have these test users:

| Email | Password | Role | Description |
|-------|----------|------|-------------|
| owner@example.com | password123 | owner | Super user with full access |
| admin@example.com | password123 | admin | Admin (cannot modify owner) |
| system@example.com | password123 | system | System user for automated tasks |
| alice@example.com | password123 | customer | Regular customer |
| bob@example.com | password123 | customer | Regular customer |

## API Endpoints

### Public Endpoints (No Authentication Required)

#### 1. Register New User
```bash
POST http://localhost:8080/v1/auth/register
Content-Type: application/json

{
  "first_name": "Test",
  "last_name": "User",
  "email": "test@example.com",
  "password": "password123"
}
```

#### 2. Login
```bash
POST http://localhost:8080/v1/auth/login
Content-Type: application/json

{
  "email": "owner@example.com",
  "password": "password123",
  "stay_signed_in": false
}
```

**Response:**
- Sets `access_token` cookie (15 min)
- Sets `refresh_token` cookie (7 days or 30 days if stay_signed_in=true)
- Returns user info and tokens in body (for Bearer token support)

#### 3. Request Password Reset
```bash
POST http://localhost:8080/v1/auth/reset-password
Content-Type: application/json

{
  "email": "owner@example.com"
}
```

**Note:** OTP will be logged to console. Check server logs.

#### 4. Verify Password Reset
```bash
POST http://localhost:8080/v1/auth/reset-password/verify
Content-Type: application/json

{
  "email": "owner@example.com",
  "otp": "123456",
  "new_password": "newpassword123"
}
```

### Protected Endpoints (Authentication Required)

**Note:** For Postman/API clients, you can use either:
- Cookies (automatic after login)
- Bearer token: `Authorization: Bearer <access_token>`

#### 5. Get Current User Info
```bash
GET http://localhost:8080/v1/auth/me
```

#### 6. Refresh Access Token
```bash
POST http://localhost:8080/v1/auth/refresh
```

**Note:** Automatically rotates refresh token for security.

#### 7. Change Password
```bash
POST http://localhost:8080/v1/auth/change-password
Content-Type: application/json

{
  "current_password": "password123",
  "new_password": "newpassword456"
}
```

#### 8. Get All Active Sessions
```bash
GET http://localhost:8080/v1/auth/sessions
```

**Response:**
```json
[
  {
    "id": "session-uuid",
    "device_name": "Chrome on Desktop",
    "ip_address": "192.168.1.100",
    "last_activity_at": "2026-01-06T10:30:00Z",
    "expires_at": "2026-01-13T10:30:00Z",
    "created_at": "2026-01-06T10:30:00Z",
    "is_current": true
  }
]
```

#### 9. Delete Specific Session
```bash
DELETE http://localhost:8080/v1/auth/sessions/{session_id}
```

#### 10. Logout Current Session
```bash
POST http://localhost:8080/v1/auth/logout
```

#### 11. Logout All Devices (Current User)
```bash
POST http://localhost:8080/v1/auth/logout-all
```

**Note:** This logs out all devices for the currently authenticated user.

### Admin Endpoints (Requires owner, admin, or system role)

#### 12. Logout All User Sessions (Admin)
```bash
POST http://localhost:8080/v1/auth/logout-all-user-sessions/{user_id}
```

**Effect:**
- Admin can logout all sessions for any user
- Useful for security incidents or support requests
- Does not require the target user to be logged in

**Example:**
```bash
# Admin logs out all sessions for user alice
POST http://localhost:8080/v1/auth/logout-all-user-sessions/550e8400-e29b-41d4-a716-446655440004
Authorization: Bearer <admin_token>
```

#### 14. Block User
```bash
POST http://localhost:8080/v1/auth/block-user/{user_id}
```

**Effect:** 
- Marks user as blocked
- Invalidates all their active sessions
- User cannot login or access resources

#### 15. Unblock User
```bash
POST http://localhost:8080/v1/auth/unblock-user/{user_id}
```

### Protected Resource Endpoints

All existing endpoints now require authentication:

```bash
GET http://localhost:8080/v1/users
GET http://localhost:8080/v1/users/{id}
POST http://localhost:8080/v1/users
PUT http://localhost:8080/v1/users/{id}
DELETE http://localhost:8080/v1/users/{id}

GET http://localhost:8080/v1/products
GET http://localhost:8080/v1/products/{id}
POST http://localhost:8080/v1/products
PUT http://localhost:8080/v1/products/{id}
DELETE http://localhost:8080/v1/products/{id}
```

## Testing Scenarios

### Scenario 1: Basic Login Flow
1. Login with `owner@example.com`
2. Verify you receive access and refresh tokens
3. Call `/v1/auth/me` to verify authentication
4. Call `/v1/users` to verify protected routes work

### Scenario 2: Token Expiration & Refresh
1. Login
2. Wait 16 minutes (access token expires after 15 min)
3. Try to access `/v1/auth/me` - should get 401
4. Call `/v1/auth/refresh` - should get new tokens
5. Retry `/v1/auth/me` - should work

### Scenario 3: Stay Signed In
1. Login with `stay_signed_in: true`
2. Verify refresh token expires in 30 days instead of 7

### Scenario 4: Multi-Device Sessions
1. Login from Postman (or browser)
2. Login from another client (different IP/user-agent if possible)
3. Call `/v1/auth/sessions` - should see 2 sessions
4. Delete one session
5. Verify that session is invalidated

### Scenario 5: User Blocking
1. Login as `owner@example.com`
2. Block user `alice@example.com` via `/v1/auth/block-user/{alice_id}`
3. Try to login as Alice - should fail with "User account has been blocked"
4. If Alice was logged in, her sessions should be invalidated

### Scenario 6: Password Reset
1. Request password reset for `bob@example.com`
2. Check server console for OTP
3. Verify password reset with OTP
4. Try logging in with old password - should fail
5. Login with new password - should work
6. All Bob's previous sessions should be invalidated

### Scenario 7: Logout All Devices
1. Login from multiple clients (simulate multiple devices)
2. Verify multiple sessions in `/v1/auth/sessions`
3. Call `/v1/auth/logout-all`
4. Verify all sessions are invalidated
5. Try to access protected route - should fail

### Scenario 8: Role-Based Access Control
1. Login as `customer` (alice@example.com)
2. Try to block another user - should get 403 Forbidden
3. Login as `admin` (admin@example.com)
4. Block a customer - should work
5. Try to block `owner` - should work (admin can block owner)

### Scenario 9: Bearer Token Support (Postman)
1. Login and copy `access_token` from response
2. Remove cookies
3. Add header: `Authorization: Bearer <access_token>`
4. Access protected routes - should work

### Scenario 10: Cookie-Based Auth (Browser)
1. Login from browser
2. Cookies are automatically set
3. Access protected routes - should work automatically
4. No need to manually add headers

## Common Issues & Troubleshooting

### Issue: "Missing authentication token"
- **Cause:** No cookie or Authorization header
- **Fix:** Ensure cookies are enabled in Postman or add Bearer token header

### Issue: "Token has expired"
- **Cause:** Access token expired (15 min)
- **Fix:** Call `/v1/auth/refresh` to get new tokens

### Issue: "Session is inactive"
- **Cause:** Session was invalidated (logout, user blocked, etc.)
- **Fix:** Login again

### Issue: "User account is blocked"
- **Cause:** Admin blocked the user
- **Fix:** Contact admin to unblock

### Issue: "Invalid or expired OTP"
- **Cause:** OTP expired (15 min) or wrong OTP
- **Fix:** Request new password reset

## Security Features Implemented

✅ **HttpOnly Cookies** - Prevents XSS attacks
✅ **Secure Cookies** - HTTPS only in production
✅ **SameSite Cookies** - CSRF protection
✅ **Bcrypt Password Hashing** - Secure password storage
✅ **JWT Token Signing** - Prevents token tampering
✅ **Refresh Token Rotation** - Prevents replay attacks
✅ **Session Tracking** - Multi-device management
✅ **Instant User Blocking** - Immediate access revocation
✅ **Short-lived Access Tokens** - Limits damage from token theft
✅ **Long-lived Refresh Tokens** - Better UX without compromising security
✅ **Device Fingerprinting** - Track login devices
✅ **IP Tracking** - Audit trail for security

## Next Steps

1. Integrate email service for password reset OTPs
2. Add rate limiting on login/register endpoints
3. Add 2FA support
4. Add OAuth providers (Google, GitHub, etc.)
5. Add password strength validation
6. Add email verification on registration
7. Add "remember this device" feature
8. Add suspicious login detection

