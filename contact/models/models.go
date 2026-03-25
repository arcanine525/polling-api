package models

import (
	"time"

	authmodels "polling-system/auth/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// File represents an uploaded CSV file record.
type File struct {
	ID          string           `gorm:"type:text;primaryKey" json:"id"`
	FileName    string           `gorm:"type:text;not null" json:"file_name"`
	RecordCount int              `gorm:"not null" json:"record_count"`
	CreatedAt   time.Time        `gorm:"autoCreateTime;not null" json:"created_at"`
	CreatedBy   string           `gorm:"type:text;not null" json:"created_by"`
	Creator     *authmodels.User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	UpdatedAt   *time.Time       `json:"updated_at"`
	UpdatedBy   *string          `json:"updated_by"`
	Notes       *string          `gorm:"type:text" json:"notes"`
	Contacts    []Contact        `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"contacts,omitempty"`
}

// BeforeCreate generates a UUID for new files.
func (f *File) BeforeCreate(tx *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuid.New().String()
	}
	return nil
}

func (File) TableName() string {
	return "files"
}

// Contact represents a single contact record.
type Contact struct {
	ID              string     `gorm:"type:text;primaryKey" json:"id"`
	Name            *string    `gorm:"type:text;index" json:"name"`
	PhoneNumber     string     `gorm:"type:text;not null;uniqueIndex:idx_poll_phone" json:"phone_number"`
	PhoneNormalized string     `gorm:"type:text;not null;index" json:"phone_normalized"`
	CreatedAt       time.Time  `gorm:"autoCreateTime;not null" json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	FileID          *string    `gorm:"type:text" json:"file_id"`
	PollID          *string    `gorm:"type:text;uniqueIndex:idx_poll_phone" json:"poll_id"`
}

func (Contact) TableName() string {
	return "contacts"
}

// BeforeCreate generates a UUID for new contacts.
func (c *Contact) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// ContactCreate holds the fields for creating a new contact.
type ContactCreate struct {
	Name        *string `json:"name"`
	PhoneNumber string  `json:"phone_number" binding:"required"`
	FileID      *string `json:"file_id"`
	PollID      *string `json:"poll_id"`
}

// BulkContactCreate holds the fields for bulk creating contacts.
type BulkContactCreate struct {
	PhoneNumbers []string `json:"phone_numbers" binding:"required,min=1"`
	PollID       string   `json:"poll_id" binding:"required"`
}

// BulkContactResult holds the result of bulk contact creation.
type BulkContactResult struct {
	Created   []Contact `json:"created"`
	Duplicate []string  `json:"duplicate"`
	Invalid   []string  `json:"invalid"`
}

// ContactUpdate holds the fields that can be updated.
type ContactUpdate struct {
	Name        *string `json:"name"`
	PhoneNumber *string `json:"phone_number"`
}

// PaginatedContacts is the response for paginated contact lists.
type PaginatedContacts struct {
	Items []Contact `json:"items"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Size  int       `json:"size"`
	Pages int       `json:"pages"`
}

// PaginatedFiles is the response for paginated file lists.
type PaginatedFiles struct {
	Items []File `json:"items"`
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
	Pages int    `json:"pages"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}
