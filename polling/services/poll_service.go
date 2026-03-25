package services

import (
	"fmt"
	"math"
	"time"

	"polling-system/polling/models"

	"gorm.io/gorm"
)

// CreatePoll creates a new poll owned by the given user.
func CreatePoll(db *gorm.DB, userID string, data models.PollCreate) (*models.Poll, error) {
	poll := models.Poll{
		PollName:             data.PollName,
		PollerName:           data.PollerName,
		FirstTypeformSMSText: data.FirstTypeformSMSText,
		CreatedBy:            userID,
	}

	if data.Status != nil {
		poll.Status = *data.Status
	}
	if data.VoicePollDelayMinutes != nil {
		poll.VoicePollDelayMinutes = *data.VoicePollDelayMinutes
	}
	if data.VoicePollMethod != nil {
		poll.VoicePollMethod = *data.VoicePollMethod
	}
	if data.CustomVoicePollInstructions != nil {
		poll.CustomVoicePollInstructions = data.CustomVoicePollInstructions
	}
	if data.ScheduledToStartAt != nil {
		poll.ScheduledToStartAt = data.ScheduledToStartAt
	}
	if data.TypeformForms != nil {
		poll.TypeformForms = data.TypeformForms
	}
	if data.BlandPathwayIDs != nil {
		poll.BlandPathwayIDs = data.BlandPathwayIDs
	}
	if data.DontSendSMS != nil {
		poll.DontSendSMS = *data.DontSendSMS
	}
	if data.OnlyDayHours != nil {
		poll.OnlyDayHours = *data.OnlyDayHours
	}
	if data.TimeZone != nil {
		poll.TimeZone = *data.TimeZone
	}

	if err := db.Create(&poll).Error; err != nil {
		return nil, err
	}
	return &poll, nil
}

// GetPolls returns a paginated list of polls, optionally filtered by status.
func GetPolls(db *gorm.DB, page, size int, status string) (*models.PaginatedPolls, error) {
	query := db.Model(&models.Poll{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(size)))
	}

	var items []models.Poll
	offset := (page - 1) * size
	if err := query.Preload("Creator").Order("created_at DESC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}

	return &models.PaginatedPolls{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
		Pages: pages,
	}, nil
}

// GetPollByID returns a poll by ID with its creator preloaded.
func GetPollByID(db *gorm.DB, id string) (*models.Poll, error) {
	var poll models.Poll
	if err := db.Preload("Creator").Where("id = ?", id).First(&poll).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("poll not found")
		}
		return nil, err
	}
	return &poll, nil
}

// UpdatePoll performs a partial update on a poll.
func UpdatePoll(db *gorm.DB, id string, data models.PollUpdate) (*models.Poll, error) {
	poll, err := GetPollByID(db, id)
	if err != nil {
		return nil, err
	}

	if data.PollName != nil {
		poll.PollName = *data.PollName
	}
	if data.PollerName != nil {
		poll.PollerName = *data.PollerName
	}
	if data.Status != nil {
		poll.Status = *data.Status
	}
	if data.VoicePollDelayMinutes != nil {
		poll.VoicePollDelayMinutes = *data.VoicePollDelayMinutes
	}
	if data.FirstTypeformSMSText != nil {
		poll.FirstTypeformSMSText = *data.FirstTypeformSMSText
	}
	if data.VoicePollMethod != nil {
		poll.VoicePollMethod = *data.VoicePollMethod
	}
	if data.CustomVoicePollInstructions != nil {
		poll.CustomVoicePollInstructions = data.CustomVoicePollInstructions
	}
	if data.ScheduledToStartAt != nil {
		poll.ScheduledToStartAt = data.ScheduledToStartAt
	}
	if data.TypeformForms != nil {
		poll.TypeformForms = data.TypeformForms
	}
	if data.BlandPathwayIDs != nil {
		poll.BlandPathwayIDs = data.BlandPathwayIDs
	}
	if data.DontSendSMS != nil {
		poll.DontSendSMS = *data.DontSendSMS
	}
	if data.OnlyDayHours != nil {
		poll.OnlyDayHours = *data.OnlyDayHours
	}
	if data.TimeZone != nil {
		poll.TimeZone = *data.TimeZone
	}

	now := time.Now().UTC()
	poll.UpdatedAt = &now

	if err := db.Save(poll).Error; err != nil {
		return nil, err
	}
	return poll, nil
}

// DeletePoll removes a poll by ID (cascades via FK constraints).
func DeletePoll(db *gorm.DB, id string) error {
	result := db.Where("id = ?", id).Delete(&models.Poll{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("poll not found")
	}
	return nil
}

// IsPollOwner checks if the given user owns the poll.
func IsPollOwner(db *gorm.DB, pollID, userID string) bool {
	var count int64
	db.Model(&models.Poll{}).Where("id = ? AND created_by = ?", pollID, userID).Count(&count)
	return count > 0
}
