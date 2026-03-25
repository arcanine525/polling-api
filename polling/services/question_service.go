package services

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"polling-system/polling/models"

	"gorm.io/gorm"
)

// CreateQuestion adds a question to a poll.
func CreateQuestion(db *gorm.DB, pollID string, data models.PollQuestionCreate) (*models.PollQuestion, error) {
	// Validate poll exists
	if err := db.Where("id = ?", pollID).First(&models.Poll{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("poll not found")
		}
		return nil, err
	}

	// Validate default_next_question_id belongs to same poll
	if data.DefaultNextQuestionID != nil {
		var count int64
		db.Model(&models.PollQuestion{}).Where("id = ? AND poll_id = ?", *data.DefaultNextQuestionID, pollID).Count(&count)
		if count == 0 {
			return nil, fmt.Errorf("default_next_question_id must belong to the same poll")
		}
	}

	question := models.PollQuestion{
		PollID:     pollID,
		Question:   data.Question,
		AnswerType: data.AnswerType,
	}

	if data.TypeformFieldID != nil {
		question.TypeformFieldID = data.TypeformFieldID
	}
	if data.AnswerOptions != nil {
		question.AnswerOptions = data.AnswerOptions
	}
	if data.DefaultNextQuestionID != nil {
		question.DefaultNextQuestionID = data.DefaultNextQuestionID
	}
	if data.FlowRules != nil {
		question.FlowRules = *data.FlowRules
	} else {
		empty := json.RawMessage([]byte("[]"))
		question.FlowRules = empty
	}
	if data.RandomizeAnswers != nil {
		question.RandomizeAnswers = *data.RandomizeAnswers
	}

	if err := db.Create(&question).Error; err != nil {
		return nil, err
	}
	return &question, nil
}

// GetQuestionsByPoll returns a paginated list of questions for a poll.
func GetQuestionsByPoll(db *gorm.DB, pollID string, page, size int) (*models.PaginatedQuestions, error) {
	query := db.Model(&models.PollQuestion{}).Where("poll_id = ?", pollID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(size)))
	}

	var items []models.PollQuestion
	offset := (page - 1) * size
	if err := query.Order("created_at ASC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}

	return &models.PaginatedQuestions{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
		Pages: pages,
	}, nil
}

// GetQuestionByID returns a question by ID.
func GetQuestionByID(db *gorm.DB, id string) (*models.PollQuestion, error) {
	var question models.PollQuestion
	if err := db.Where("id = ?", id).First(&question).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("question not found")
		}
		return nil, err
	}
	return &question, nil
}

// UpdateQuestion performs a partial update on a question.
func UpdateQuestion(db *gorm.DB, id string, data models.PollQuestionUpdate) (*models.PollQuestion, error) {
	question, err := GetQuestionByID(db, id)
	if err != nil {
		return nil, err
	}

	// Validate default_next_question_id belongs to same poll
	if data.DefaultNextQuestionID != nil {
		var count int64
		db.Model(&models.PollQuestion{}).Where("id = ? AND poll_id = ?", *data.DefaultNextQuestionID, question.PollID).Count(&count)
		if count == 0 {
			return nil, fmt.Errorf("default_next_question_id must belong to the same poll")
		}
	}

	if data.Question != nil {
		question.Question = *data.Question
	}
	if data.AnswerType != nil {
		question.AnswerType = *data.AnswerType
	}
	if data.TypeformFieldID != nil {
		question.TypeformFieldID = data.TypeformFieldID
	}
	if data.AnswerOptions != nil {
		question.AnswerOptions = data.AnswerOptions
	}
	if data.DefaultNextQuestionID != nil {
		question.DefaultNextQuestionID = data.DefaultNextQuestionID
	}
	if data.FlowRules != nil {
		question.FlowRules = *data.FlowRules
	}
	if data.RandomizeAnswers != nil {
		question.RandomizeAnswers = *data.RandomizeAnswers
	}

	now := time.Now().UTC()
	question.UpdatedAt = &now

	if err := db.Save(question).Error; err != nil {
		return nil, err
	}
	return question, nil
}

// DeleteQuestion removes a question by ID.
func DeleteQuestion(db *gorm.DB, id string) error {
	result := db.Where("id = ?", id).Delete(&models.PollQuestion{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("question not found")
	}
	return nil
}
