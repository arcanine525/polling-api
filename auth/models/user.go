package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an application user linked to Firebase Auth.
type User struct {
	ID          string     `gorm:"type:text;primaryKey" json:"id"`
	FirebaseUID string     `gorm:"type:text;uniqueIndex;not null" json:"firebase_uid"`
	Email       string     `gorm:"type:text;not null" json:"email"`
	DisplayName *string    `gorm:"type:text" json:"display_name"`
	Role        string     `gorm:"type:text;not null;default:'viewer'" json:"role"`
	CreatedAt   time.Time  `gorm:"autoCreateTime;not null" json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

// BeforeCreate generates a UUID for new users.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

func (User) TableName() string {
	return "users"
}

// UserCreate holds fields for admin manual user creation.
type UserCreate struct {
	FirebaseUID string  `json:"firebase_uid" binding:"required"`
	Email       string  `json:"email" binding:"required,email"`
	DisplayName *string `json:"display_name"`
	Role        string  `json:"role" binding:"omitempty,oneof=admin poller viewer"`
}

// UserUpdate holds fields for updating own profile.
type UserUpdate struct {
	DisplayName *string `json:"display_name"`
}

// UserRoleUpdate holds fields for admin role changes.
type UserRoleUpdate struct {
	Role string `json:"role" binding:"required,oneof=admin poller viewer"`
}

// PaginatedUsers is the response for paginated user lists.
type PaginatedUsers struct {
	Items []User `json:"items"`
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
	Pages int    `json:"pages"`
}

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// ProvisionRequest is the internal provisioning request.
type ProvisionRequest struct {
	FirebaseUID string `json:"firebase_uid" binding:"required"`
	Email       string `json:"email" binding:"required"`
}

// LoginRequest holds email/password credentials.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse returns the authenticated user profile with Firebase tokens.
type LoginResponse struct {
	User         User   `json:"user"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
}
