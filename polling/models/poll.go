package models

import (
	"encoding/json"
	"time"

	authmodels "polling-system/auth/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Poll represents a polling campaign.
type Poll struct {
	ID                          string           `gorm:"type:text;primaryKey" json:"id"`
	PollName                    string           `gorm:"type:text;not null" json:"poll_name"`
	Status                      string           `gorm:"type:text;not null;default:'draft'" json:"status"`
	CreatedAt                   time.Time        `gorm:"autoCreateTime;not null" json:"created_at"`
	UpdatedAt                   *time.Time       `json:"updated_at"`
	CreatedBy                   string           `gorm:"type:text;not null" json:"created_by"`
	Creator                     *authmodels.User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	PollerName                  string           `gorm:"type:text;not null" json:"poller_name"`
	VoicePollDelayMinutes       int              `gorm:"not null;default:5" json:"voice_poll_delay_minutes"`
	FirstTypeformSMSText        string           `gorm:"type:text;not null" json:"first_typeform_sms_text"`
	VoicePollMethod             string           `gorm:"type:text;not null;default:'livekit'" json:"voice_poll_method"`
	CustomVoicePollInstructions *string          `gorm:"type:text" json:"custom_voice_poll_instructions"`
	ScheduledToStartAt          *time.Time       `json:"scheduled_to_start_at"`
	TypeformForms               *json.RawMessage `gorm:"type:jsonb;serializer:json" json:"typeform_forms" swaggertype:"object"`
	BlandPathwayIDs             *json.RawMessage `gorm:"type:jsonb;serializer:json" json:"bland_pathway_ids" swaggertype:"object"`
	DontSendSMS                 bool             `gorm:"not null;default:false" json:"dont_send_sms"`
	OnlyDayHours                bool             `gorm:"not null;default:true" json:"only_day_hours"`
	TimeZone                    string           `gorm:"type:text;not null;default:'Australia/Sydney'" json:"time_zone"`
	Candidates                  []Candidate      `gorm:"foreignKey:PollID;constraint:OnDelete:CASCADE" json:"candidates,omitempty"`
	Questions                   []PollQuestion   `gorm:"foreignKey:PollID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
}

// BeforeCreate generates a UUID for new polls.
func (p *Poll) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

func (Poll) TableName() string {
	return "polls"
}

// PollCreate holds fields for creating a new poll.
type PollCreate struct {
	PollName                    string           `json:"poll_name" binding:"required"`
	PollerName                  string           `json:"poller_name"`
	FirstTypeformSMSText        string           `json:"first_typeform_sms_text"`
	Status                      *string          `json:"status"`
	VoicePollDelayMinutes       *int             `json:"voice_poll_delay_minutes"`
	VoicePollMethod             *string          `json:"voice_poll_method"`
	CustomVoicePollInstructions *string          `json:"custom_voice_poll_instructions"`
	ScheduledToStartAt          *time.Time       `json:"scheduled_to_start_at"`
	TypeformForms               *json.RawMessage `json:"typeform_forms" swaggertype:"object"`
	BlandPathwayIDs             *json.RawMessage `json:"bland_pathway_ids" swaggertype:"object"`
	DontSendSMS                 *bool            `json:"dont_send_sms"`
	OnlyDayHours                *bool            `json:"only_day_hours"`
	TimeZone                    *string          `json:"time_zone"`
}

// PollUpdate holds fields for updating an existing poll.
type PollUpdate struct {
	PollName                    *string          `json:"poll_name"`
	PollerName                  *string          `json:"poller_name"`
	Status                      *string          `json:"status"`
	VoicePollDelayMinutes       *int             `json:"voice_poll_delay_minutes"`
	FirstTypeformSMSText        *string          `json:"first_typeform_sms_text"`
	VoicePollMethod             *string          `json:"voice_poll_method"`
	CustomVoicePollInstructions *string          `json:"custom_voice_poll_instructions"`
	ScheduledToStartAt          *time.Time       `json:"scheduled_to_start_at"`
	TypeformForms               *json.RawMessage `json:"typeform_forms" swaggertype:"object"`
	BlandPathwayIDs             *json.RawMessage `json:"bland_pathway_ids" swaggertype:"object"`
	DontSendSMS                 *bool            `json:"dont_send_sms"`
	OnlyDayHours                *bool            `json:"only_day_hours"`
	TimeZone                    *string          `json:"time_zone"`
}

// PaginatedPolls is the response for paginated poll lists.
type PaginatedPolls struct {
	Items []Poll `json:"items"`
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
	Pages int    `json:"pages"`
}
