package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PollQuestion represents a question belonging to a poll.
type PollQuestion struct {
	ID                    string           `gorm:"type:text;primaryKey" json:"id"`
	PollID                string           `gorm:"type:text;not null" json:"poll_id"`
	Poll                  *Poll            `gorm:"foreignKey:PollID;constraint:OnDelete:CASCADE" json:"poll,omitempty"`
	Question              string           `gorm:"type:text;not null" json:"question"`
	AnswerType            string           `gorm:"type:text;not null" json:"answer_type"`
	CreatedAt             time.Time        `gorm:"autoCreateTime;not null" json:"created_at"`
	UpdatedAt             *time.Time       `json:"updated_at"`
	TypeformFieldID       *string          `gorm:"type:text" json:"typeform_field_id"`
	AnswerOptions         *json.RawMessage `gorm:"type:jsonb;serializer:json" json:"answer_options" swaggertype:"object"`
	DefaultNextQuestionID *string          `gorm:"type:text" json:"default_next_question_id"`
	DefaultNextQuestion   *PollQuestion    `gorm:"foreignKey:DefaultNextQuestionID;constraint:OnDelete:SET NULL" json:"default_next_question,omitempty"`
	FlowRules             json.RawMessage  `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"flow_rules" swaggertype:"object"`
	RandomizeAnswers      bool             `gorm:"not null;default:false" json:"randomize_answers"`
	Answers               []Answer         `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE" json:"answers,omitempty"`
}

// BeforeCreate generates a UUID for new questions.
func (q *PollQuestion) BeforeCreate(tx *gorm.DB) error {
	if q.ID == "" {
		q.ID = uuid.New().String()
	}
	return nil
}

func (PollQuestion) TableName() string {
	return "poll_questions"
}

// PollQuestionCreate holds fields for adding a question to a poll.
type PollQuestionCreate struct {
	Question              string           `json:"question" binding:"required"`
	AnswerType            string           `json:"answer_type" binding:"required"`
	TypeformFieldID       *string          `json:"typeform_field_id"`
	AnswerOptions         *json.RawMessage `json:"answer_options" swaggertype:"object"`
	DefaultNextQuestionID *string          `json:"default_next_question_id"`
	FlowRules             *json.RawMessage `json:"flow_rules" swaggertype:"object"`
	RandomizeAnswers      *bool            `json:"randomize_answers"`
}

// PollQuestionUpdate holds optional fields for updating a question.
type PollQuestionUpdate struct {
	Question              *string          `json:"question"`
	AnswerType            *string          `json:"answer_type"`
	TypeformFieldID       *string          `json:"typeform_field_id"`
	AnswerOptions         *json.RawMessage `json:"answer_options" swaggertype:"object"`
	DefaultNextQuestionID *string          `json:"default_next_question_id"`
	FlowRules             *json.RawMessage `json:"flow_rules" swaggertype:"object"`
	RandomizeAnswers      *bool            `json:"randomize_answers"`
}

// PaginatedQuestions is the response for paginated question lists.
type PaginatedQuestions struct {
	Items []PollQuestion `json:"items"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
	Pages int            `json:"pages"`
}
