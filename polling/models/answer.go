package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Answer represents a recorded answer from a candidate to a poll question.
type Answer struct {
	ID          string        `gorm:"type:text;primaryKey" json:"id"`
	QuestionID  string        `gorm:"type:text;not null;uniqueIndex:answers_candidate_question_uniq" json:"question_id"`
	Question    *PollQuestion `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE" json:"question,omitempty"`
	CandidateID string        `gorm:"type:text;not null;uniqueIndex:answers_candidate_question_uniq" json:"candidate_id"`
	Candidate   *Candidate    `gorm:"foreignKey:CandidateID;constraint:OnDelete:CASCADE" json:"candidate,omitempty"`
	Answer      string        `gorm:"type:text;not null" json:"answer"`
	CreatedAt   time.Time     `gorm:"autoCreateTime;not null" json:"created_at"`
	Source      string        `gorm:"type:text;not null;default:'typeform'" json:"source"`
}

// BeforeCreate generates a UUID for new answers.
func (a *Answer) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

func (Answer) TableName() string {
	return "answers"
}

// AnswerCreate holds fields for submitting an answer.
type AnswerCreate struct {
	CandidateID string  `json:"candidate_id" binding:"required"`
	Answer      string  `json:"answer" binding:"required"`
	Source      *string `json:"source"`
}

// PaginatedAnswers is the response for paginated answer lists.
type PaginatedAnswers struct {
	Items []Answer `json:"items"`
	Total int64    `json:"total"`
	Page  int      `json:"page"`
	Size  int      `json:"size"`
	Pages int      `json:"pages"`
}
