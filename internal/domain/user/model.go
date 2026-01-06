package user

import "time"

// User represents a user in the system
type User struct {
	ID        string     `json:"id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Email     string     `json:"email"`
	Password  string     `json:"-"` // Never expose password in JSON
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	IsBlocked bool       `json:"is_blocked"`
	BlockedAt *time.Time `json:"blocked_at,omitempty"`
	BlockedBy *string    `json:"blocked_by,omitempty"`
	CreatedBy *string    `json:"created_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedBy *string    `json:"updated_by,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedBy *string    `json:"deleted_by,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
