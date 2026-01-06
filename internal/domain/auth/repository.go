package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetActiveSessionIDsByUserID returns active session IDs for a user.
// Used for cache invalidation and should be treated as best-effort.
func (r *Repository) GetActiveSessionIDsByUserID(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT id
		FROM user_sessions
		WHERE user_id = $1 AND is_active = true
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query session ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan session id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("session ids rows: %w", err)
	}
	return ids, nil
}

// -------------------------
// User Auth Queries
// -------------------------

// GetUserByEmail retrieves a user by email with password and role
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*UserWithAuth, error) {
	query := `
		SELECT u.id, u.first_name, u.last_name, u.email, u.password, 
		       u.is_active, u.is_blocked, u.created_at, u.updated_at, u.deleted_at,
		       ro.name as role_name
		FROM users u
		JOIN roles ro ON u.role_id = ro.id
		WHERE u.email = $1 AND u.deleted_at IS NULL
	`

	var user UserWithAuth
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.IsActive,
		&user.IsBlocked,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Role,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID with role
func (r *Repository) GetUserByID(ctx context.Context, userID string) (*UserWithAuth, error) {
	query := `
		SELECT u.id, u.first_name, u.last_name, u.email, u.password, 
		       u.is_active, u.is_blocked, u.created_at, u.updated_at, u.deleted_at,
		       ro.name as role_name
		FROM users u
		JOIN roles ro ON u.role_id = ro.id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`

	var user UserWithAuth
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.IsActive,
		&user.IsBlocked,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Role,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// CreateUser creates a new user
func (r *Repository) CreateUser(ctx context.Context, firstName, lastName, email, hashedPassword string) (string, error) {
	query := `
		INSERT INTO users (first_name, last_name, email, password, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, '00000000-0000-0000-0000-000000000004', true, NOW(), NOW())
		RETURNING id
	`

	var userID string
	err := r.db.QueryRow(ctx, query, firstName, lastName, email, hashedPassword).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

// UpdateUserPassword updates a user's password
func (r *Repository) UpdateUserPassword(ctx context.Context, userID, hashedPassword string) error {
	query := `
		UPDATE users 
		SET password = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, hashedPassword, userID)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}

// BlockUser blocks a user and marks them as inactive
func (r *Repository) BlockUser(ctx context.Context, userID, blockedBy string) error {
	query := `
		UPDATE users 
		SET is_blocked = true, blocked_at = NOW(), blocked_by = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, blockedBy, userID)
	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	return nil
}

// UnblockUser unblocks a user
func (r *Repository) UnblockUser(ctx context.Context, userID string) error {
	query := `
		UPDATE users 
		SET is_blocked = false, blocked_at = NULL, blocked_by = NULL, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}

	return nil
}

// -------------------------
// Session Management
// -------------------------

// CreateSession creates a new session
func (r *Repository) CreateSession(ctx context.Context, session *Session) error {
	deviceInfoJSON, err := json.Marshal(session.DeviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}

	query := `
		INSERT INTO user_sessions (user_id, refresh_token_hash, device_info, ip_address, user_agent, is_active, last_activity_at, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err = r.db.QueryRow(ctx, query,
		session.UserID,
		session.RefreshTokenHash,
		deviceInfoJSON,
		session.IPAddress,
		session.UserAgent,
		session.IsActive,
		session.LastActivityAt,
		session.ExpiresAt,
		session.CreatedAt,
	).Scan(&session.ID)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByRefreshTokenHash retrieves a session by refresh token hash
func (r *Repository) GetSessionByRefreshTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, device_info, ip_address, user_agent, 
		       is_active, last_activity_at, expires_at, created_at
		FROM user_sessions
		WHERE refresh_token_hash = $1
	`

	var session Session
	var deviceInfoJSON []byte

	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&deviceInfoJSON,
		&session.IPAddress,
		&session.UserAgent,
		&session.IsActive,
		&session.LastActivityAt,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if err := json.Unmarshal(deviceInfoJSON, &session.DeviceInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
	}

	return &session, nil
}

// GetSessionByID retrieves a session by ID
func (r *Repository) GetSessionByID(ctx context.Context, sessionID string) (*Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, device_info, ip_address, user_agent, 
		       is_active, last_activity_at, expires_at, created_at
		FROM user_sessions
		WHERE id = $1
	`

	var session Session
	var deviceInfoJSON []byte

	err := r.db.QueryRow(ctx, query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&deviceInfoJSON,
		&session.IPAddress,
		&session.UserAgent,
		&session.IsActive,
		&session.LastActivityAt,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if err := json.Unmarshal(deviceInfoJSON, &session.DeviceInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
	}

	return &session, nil
}

// GetUserSessions retrieves all active sessions for a user
func (r *Repository) GetUserSessions(ctx context.Context, userID string) ([]*Session, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, device_info, ip_address, user_agent, 
		       is_active, last_activity_at, expires_at, created_at
		FROM user_sessions
		WHERE user_id = $1 AND is_active = true AND expires_at > NOW()
		ORDER BY last_activity_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var session Session
		var deviceInfoJSON []byte

		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshTokenHash,
			&deviceInfoJSON,
			&session.IPAddress,
			&session.UserAgent,
			&session.IsActive,
			&session.LastActivityAt,
			&session.ExpiresAt,
			&session.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		if err := json.Unmarshal(deviceInfoJSON, &session.DeviceInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
		}

		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// UpdateSessionRefreshToken updates the refresh token hash for a session
func (r *Repository) UpdateSessionRefreshToken(ctx context.Context, sessionID, newTokenHash string) error {
	query := `
		UPDATE user_sessions
		SET refresh_token_hash = $1, last_activity_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, newTokenHash, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session refresh token: %w", err)
	}

	return nil
}

// InvalidateSession marks a session as inactive
func (r *Repository) InvalidateSession(ctx context.Context, sessionID string) error {
	query := `
		UPDATE user_sessions
		SET is_active = false
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}

// InvalidateAllUserSessions marks all sessions for a user as inactive
func (r *Repository) InvalidateAllUserSessions(ctx context.Context, userID string) error {
	query := `
		UPDATE user_sessions
		SET is_active = false
		WHERE user_id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate all user sessions: %w", err)
	}

	return nil
}

// -------------------------
// Password Reset Tokens
// -------------------------

// CreatePasswordResetToken creates a new password reset token
func (r *Repository) CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token_hash, otp, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.QueryRow(ctx, query,
		token.UserID,
		token.TokenHash,
		token.OTP,
		token.ExpiresAt,
		token.CreatedAt,
	).Scan(&token.ID)

	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	return nil
}

// GetPasswordResetToken retrieves a password reset token by OTP and email
func (r *Repository) GetPasswordResetToken(ctx context.Context, email, otp string) (*PasswordResetToken, error) {
	query := `
		SELECT prt.id, prt.user_id, prt.token_hash, prt.otp, prt.expires_at, prt.used_at, prt.created_at
		FROM password_reset_tokens prt
		JOIN users u ON prt.user_id = u.id
		WHERE u.email = $1 AND prt.otp = $2 AND prt.used_at IS NULL AND prt.expires_at > NOW()
		ORDER BY prt.created_at DESC
		LIMIT 1
	`

	var token PasswordResetToken
	err := r.db.QueryRow(ctx, query, email, otp).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.OTP,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}

	return &token, nil
}

// MarkPasswordResetTokenAsUsed marks a password reset token as used
func (r *Repository) MarkPasswordResetTokenAsUsed(ctx context.Context, tokenID string) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark password reset token as used: %w", err)
	}

	return nil
}

// -------------------------
// Helper Types
// -------------------------

type UserWithAuth struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
	Password  string
	Role      string
	IsActive  bool
	IsBlocked bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
