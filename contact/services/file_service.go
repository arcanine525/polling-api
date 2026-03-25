package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"polling-system/config"
	"polling-system/contact/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var nonDigitRe = regexp.MustCompile(`[\s\-\(\)]+`)

// NormalizePhone converts a raw phone string to a normalized E.164-like format.
func NormalizePhone(raw string) string {
	cleaned := nonDigitRe.ReplaceAllString(strings.TrimSpace(raw), "")

	if len(cleaned) == 9 && isAllDigits(cleaned) {
		return "+61" + cleaned
	}
	if len(cleaned) == 10 && strings.HasPrefix(cleaned, "0") {
		return "+61" + cleaned[1:]
	}
	if strings.HasPrefix(cleaned, "+") {
		return cleaned
	}
	return "+" + cleaned
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ProcessCSVUpload reads a CSV file, saves it to disk, and persists file + contacts to DB.
func ProcessCSVUpload(db *gorm.DB, cfg *config.Config, fileHeader *multipart.FileHeader, userID string) (*models.File, error) {
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

	// Strip UTF-8 BOM if present.
	text := string(content)
	text = strings.TrimPrefix(text, "\xEF\xBB\xBF")

	reader := csv.NewReader(strings.NewReader(text))
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("CSV file is empty or unreadable")
	}

	// Find phone and name columns (case-insensitive).
	phoneIdx := -1
	nameIdx := -1
	for i, col := range header {
		lower := strings.TrimSpace(strings.ToLower(col))
		switch lower {
		case "phone", "phone_number", "phonenumber":
			phoneIdx = i
		case "name", "full_name", "fullname":
			nameIdx = i
		}
	}
	if phoneIdx == -1 {
		return nil, fmt.Errorf("CSV must contain a 'Phone' column")
	}

	// Parse rows into contacts, deduplicating by phone number within the file.
	var contacts []models.Contact
	seenPhones := make(map[string]bool)
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

		rawPhone := strings.TrimSpace(row[phoneIdx])
		if rawPhone == "" {
			continue
		}

		if seenPhones[rawPhone] {
			continue
		}
		seenPhones[rawPhone] = true

		var name *string
		if nameIdx >= 0 && len(row) > nameIdx {
			n := strings.TrimSpace(row[nameIdx])
			if n != "" {
				name = &n
			}
		}

		contacts = append(contacts, models.Contact{
			ID:              uuid.New().String(),
			Name:            name,
			PhoneNumber:     rawPhone,
			PhoneNormalized: NormalizePhone(rawPhone),
		})
	}

	// Save file to disk.
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	savedName := uuid.New().String() + "_" + fileHeader.Filename
	filePath := filepath.Join(cfg.UploadDir, savedName)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to save file to disk: %w", err)
	}

	// Persist to database inside a transaction.
	uploadedFile := models.File{
		FileName:    fileHeader.Filename,
		RecordCount: len(contacts),
		CreatedBy:   userID,
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&uploadedFile).Error; err != nil {
			return err
		}
		// Set file_id on all contacts and batch insert.
		const batchSize = 1000
		for i := 0; i < len(contacts); i += batchSize {
			end := i + batchSize
			if end > len(contacts) {
				end = len(contacts)
			}
			batch := contacts[i:end]
			for j := range batch {
				batch[j].FileID = &uploadedFile.ID
			}
			if err := tx.Create(&batch).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		// Clean up the saved file on DB error.
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save to database: %w", err)
	}

	return &uploadedFile, nil
}

// GetFiles returns a paginated list of uploaded files.
func GetFiles(db *gorm.DB, page, size int) (*models.PaginatedFiles, error) {
	query := db.Model(&models.File{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(size)))
	}

	var items []models.File
	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}

	return &models.PaginatedFiles{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
		Pages: pages,
	}, nil
}
