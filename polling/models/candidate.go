package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Candidate represents a person targeted by a poll.
type Candidate struct {
	ID                 string     `gorm:"type:text;primaryKey" json:"id"`
	PollID             string     `gorm:"type:text;not null" json:"poll_id"`
	Poll               *Poll      `gorm:"foreignKey:PollID;constraint:OnDelete:CASCADE" json:"poll,omitempty"`
	Phone              string     `gorm:"type:text;not null" json:"phone"`
	CreatedAt          time.Time  `gorm:"autoCreateTime;not null" json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
	VoicePollStartedAt *time.Time `json:"voice_poll_started_at"`
	VoicePollSentiment *string    `gorm:"type:text" json:"voice_poll_sentiment"`
	VoicemailDetected  bool       `gorm:"not null;default:false" json:"voicemail_detected"`
	Answers            []Answer   `gorm:"foreignKey:CandidateID;constraint:OnDelete:CASCADE" json:"answers,omitempty"`
}

// BeforeCreate generates a UUID for new candidates.
func (ca *Candidate) BeforeCreate(tx *gorm.DB) error {
	if ca.ID == "" {
		ca.ID = uuid.New().String()
	}
	return nil
}

func (Candidate) TableName() string {
	return "candidates"
}

// CandidateCreate holds fields for adding a candidate to a poll.
type CandidateCreate struct {
	Phone string `json:"phone" binding:"required"`
}

// CandidateUpdate holds optional fields for updating a candidate.
type CandidateUpdate struct {
	Phone              *string    `json:"phone"`
	VoicePollStartedAt *time.Time `json:"voice_poll_started_at"`
	VoicePollSentiment *string    `json:"voice_poll_sentiment"`
	VoicemailDetected  *bool      `json:"voicemail_detected"`
}

// PaginatedCandidates is the response for paginated candidate lists.
type PaginatedCandidates struct {
	Items []Candidate `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Pages int         `json:"pages"`
}

// BulkCandidateCreate holds a list of phone numbers for bulk candidate insertion.
type BulkCandidateCreate struct {
	PhoneNumbers []string `json:"phone_numbers" binding:"required,min=1"`
}

// BulkCandidateResult is the response for bulk candidate creation.
type BulkCandidateResult struct {
	Created    int      `json:"created"`
	Duplicates []string `json:"duplicates,omitempty"`
	Invalid    []string `json:"invalid,omitempty"`
}
