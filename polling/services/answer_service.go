package services

import (
	"fmt"
	"math"
	"strings"

	"polling-system/polling/models"

	"gorm.io/gorm"
)

// CreateAnswer submits an answer for a question by a candidate.
func CreateAnswer(db *gorm.DB, questionID string, data models.AnswerCreate) (*models.Answer, error) {
	// Validate question exists
	if err := db.Where("id = ?", questionID).First(&models.PollQuestion{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("question not found")
		}
		return nil, err
	}

	// Validate candidate exists
	if err := db.Where("id = ?", data.CandidateID).First(&models.Candidate{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("candidate not found")
		}
		return nil, err
	}

	answer := models.Answer{
		QuestionID:  questionID,
		CandidateID: data.CandidateID,
		Answer:      data.Answer,
	}

	if data.Source != nil {
		answer.Source = *data.Source
	}

	if err := db.Create(&answer).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
			return nil, fmt.Errorf("candidate has already answered this question")
		}
		return nil, err
	}
	return &answer, nil
}

// GetAnswersByQuestion returns a paginated list of answers for a question.
func GetAnswersByQuestion(db *gorm.DB, questionID string, page, size int) (*models.PaginatedAnswers, error) {
	query := db.Model(&models.Answer{}).Where("question_id = ?", questionID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(size)))
	}

	var items []models.Answer
	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}

	return &models.PaginatedAnswers{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
		Pages: pages,
	}, nil
}

// GetAnswersByCandidate returns a paginated list of answers by a candidate.
func GetAnswersByCandidate(db *gorm.DB, candidateID string, page, size int) (*models.PaginatedAnswers, error) {
	query := db.Model(&models.Answer{}).Where("candidate_id = ?", candidateID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(size)))
	}

	var items []models.Answer
	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}

	return &models.PaginatedAnswers{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
		Pages: pages,
	}, nil
}

// GetAnswerByID returns an answer by ID.
func GetAnswerByID(db *gorm.DB, id string) (*models.Answer, error) {
	var answer models.Answer
	if err := db.Where("id = ?", id).First(&answer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("answer not found")
		}
		return nil, err
	}
	return &answer, nil
}

// DeleteAnswer removes an answer by ID.
func DeleteAnswer(db *gorm.DB, id string) error {
	result := db.Where("id = ?", id).Delete(&models.Answer{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("answer not found")
	}
	return nil
}
