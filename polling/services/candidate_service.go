package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"strings"
	"time"

	contactServices "polling-system/contact/services"
	"polling-system/polling/models"

	"gorm.io/gorm"
)

// CreateCandidate adds a candidate to a poll.
func CreateCandidate(db *gorm.DB, pollID string, data models.CandidateCreate) (*models.Candidate, error) {
	// Validate poll exists
	if err := db.Where("id = ?", pollID).First(&models.Poll{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("poll not found")
		}
		return nil, err
	}

	candidate := models.Candidate{
		PollID: pollID,
		Phone:  data.Phone,
	}

	if err := db.Create(&candidate).Error; err != nil {
		return nil, err
	}
	return &candidate, nil
}

// GetCandidatesByPoll returns a paginated list of candidates for a poll.
func GetCandidatesByPoll(db *gorm.DB, pollID string, page, size int) (*models.PaginatedCandidates, error) {
	query := db.Model(&models.Candidate{}).Where("poll_id = ?", pollID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(size)))
	}

	var items []models.Candidate
	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}

	return &models.PaginatedCandidates{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
		Pages: pages,
	}, nil
}

// GetCandidateByID returns a candidate by ID.
func GetCandidateByID(db *gorm.DB, id string) (*models.Candidate, error) {
	var candidate models.Candidate
	if err := db.Where("id = ?", id).First(&candidate).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("candidate not found")
		}
		return nil, err
	}
	return &candidate, nil
}

// UpdateCandidate performs a partial update on a candidate.
func UpdateCandidate(db *gorm.DB, id string, data models.CandidateUpdate) (*models.Candidate, error) {
	candidate, err := GetCandidateByID(db, id)
	if err != nil {
		return nil, err
	}

	if data.Phone != nil {
		candidate.Phone = *data.Phone
	}
	if data.VoicePollStartedAt != nil {
		candidate.VoicePollStartedAt = data.VoicePollStartedAt
	}
	if data.VoicePollSentiment != nil {
		candidate.VoicePollSentiment = data.VoicePollSentiment
	}
	if data.VoicemailDetected != nil {
		candidate.VoicemailDetected = *data.VoicemailDetected
	}

	now := time.Now().UTC()
	candidate.UpdatedAt = &now

	if err := db.Save(candidate).Error; err != nil {
		return nil, err
	}
	return candidate, nil
}

// DeleteCandidate removes a candidate by ID.
func DeleteCandidate(db *gorm.DB, id string) error {
	result := db.Where("id = ?", id).Delete(&models.Candidate{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("candidate not found")
	}
	return nil
}

// ListCandidates returns a paginated list of candidates, optionally filtered by poll_id.
func ListCandidates(db *gorm.DB, pollID string, page, size int) (*models.PaginatedCandidates, error) {
	query := db.Model(&models.Candidate{})
	if pollID != "" {
		query = query.Where("poll_id = ?", pollID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(size)))
	}

	var items []models.Candidate
	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}

	return &models.PaginatedCandidates{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
		Pages: pages,
	}, nil
}

// BulkCreateCandidates creates candidates from a list of phone numbers for a given poll.
func BulkCreateCandidates(db *gorm.DB, pollID string, phones []string) (*models.BulkCandidateResult, error) {
	if err := db.Where("id = ?", pollID).First(&models.Poll{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("poll not found")
		}
		return nil, err
	}

	result := &models.BulkCandidateResult{}
	seen := make(map[string]bool)
	var candidates []models.Candidate

	for _, phone := range phones {
		phone = strings.TrimSpace(phone)
		if phone == "" {
			continue
		}

		normalized := contactServices.NormalizePhone(phone)
		if normalized == "" || len(normalized) < 5 {
			result.Invalid = append(result.Invalid, phone)
			continue
		}

		if seen[normalized] {
			result.Duplicates = append(result.Duplicates, phone)
			continue
		}
		seen[normalized] = true

		candidates = append(candidates, models.Candidate{
			PollID: pollID,
			Phone:  normalized,
		})
	}

	if len(candidates) == 0 {
		return result, nil
	}

	const batchSize = 1000
	err := db.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(candidates); i += batchSize {
			end := i + batchSize
			if end > len(candidates) {
				end = len(candidates)
			}
			if err := tx.Create(candidates[i:end]).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create candidates: %w", err)
	}

	result.Created = len(candidates)
	return result, nil
}

// ProcessCandidateCSVUpload parses a CSV file and bulk-inserts candidates for a poll.
func ProcessCandidateCSVUpload(db *gorm.DB, fileHeader *multipart.FileHeader, pollID string) (*models.BulkCandidateResult, error) {
	if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".csv") {
		return nil, fmt.Errorf("only CSV files are accepted")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %w", err)
	}

	text := strings.TrimPrefix(string(content), "\xEF\xBB\xBF")

	reader := csv.NewReader(strings.NewReader(text))
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("CSV file is empty or unreadable")
	}

	phoneIdx := -1
	for i, col := range header {
		lower := strings.TrimSpace(strings.ToLower(col))
		if lower == "phone" || lower == "phone_number" || lower == "phonenumber" {
			phoneIdx = i
			break
		}
	}
	if phoneIdx == -1 {
		return nil, fmt.Errorf("CSV must contain a 'phone' column")
	}

	var phones []string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		if len(row) <= phoneIdx {
			continue
		}
		phone := strings.TrimSpace(row[phoneIdx])
		if phone != "" {
			phones = append(phones, phone)
		}
	}

	if len(phones) == 0 {
		return nil, fmt.Errorf("no phone numbers found in CSV")
	}

	return BulkCreateCandidates(db, pollID, phones)
}
