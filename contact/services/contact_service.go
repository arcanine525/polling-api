package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"polling-system/contact/models"

	"gorm.io/gorm"
)

// GetContacts returns a paginated, optionally filtered list of contacts.
func GetContacts(db *gorm.DB, page, size int, name, phone, fileID, pollID string) (*models.PaginatedContacts, error) {
	query := db.Model(&models.Contact{})

	if fileID != "" {
		query = query.Where("file_id = ?", fileID)
	}
	if pollID != "" {
		query = query.Where("poll_id = ?", pollID)
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if phone != "" {
		normalized := NormalizePhone(phone)
		query = query.Where("phone_normalized LIKE ? OR phone_number LIKE ?", "%"+normalized+"%", "%"+phone+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(size)))
	}

	var items []models.Contact
	offset := (page - 1) * size
	if err := query.Order("created_at DESC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}

	return &models.PaginatedContacts{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
		Pages: pages,
	}, nil
}

// CreateContact creates a new contact record.
func CreateContact(db *gorm.DB, data models.ContactCreate) (*models.Contact, error) {
	contact := models.Contact{
		Name:            data.Name,
		PhoneNumber:     data.PhoneNumber,
		PhoneNormalized: NormalizePhone(data.PhoneNumber),
		FileID:          data.FileID,
		PollID:          data.PollID,
	}

	if err := db.Create(&contact).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("contact with this phone number already exists for the given poll")
		}
		return nil, err
	}
	return &contact, nil
}

// GetContactByID returns a single contact or an error if not found.
func GetContactByID(db *gorm.DB, contactID string) (*models.Contact, error) {
	var contact models.Contact
	if err := db.Where("id = ?", contactID).First(&contact).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("contact not found")
		}
		return nil, err
	}
	return &contact, nil
}

// UpdateContact updates the name and/or phone_number of a contact.
func UpdateContact(db *gorm.DB, contactID string, data models.ContactUpdate) (*models.Contact, error) {
	contact, err := GetContactByID(db, contactID)
	if err != nil {
		return nil, err
	}

	if data.Name != nil {
		contact.Name = data.Name
	}
	if data.PhoneNumber != nil {
		contact.PhoneNumber = *data.PhoneNumber
		contact.PhoneNormalized = NormalizePhone(*data.PhoneNumber)
	}

	now := time.Now().UTC()
	contact.UpdatedAt = &now

	if err := db.Save(contact).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("contact with this phone number already exists for the given file")
		}
		return nil, err
	}
	return contact, nil
}

// DeleteContact removes a contact by ID.
func DeleteContact(db *gorm.DB, contactID string) error {
	contact, err := GetContactByID(db, contactID)
	if err != nil {
		return err
	}
	return db.Delete(contact).Error
}

// BulkCreateContacts creates multiple contacts from phone numbers for a specific poll.
func BulkCreateContacts(db *gorm.DB, data models.BulkContactCreate) (*models.BulkContactResult, error) {
	result := &models.BulkContactResult{
		Created:   []models.Contact{},
		Duplicate: []string{},
		Invalid:   []string{},
	}

	for _, phone := range data.PhoneNumbers {
		phone = strings.TrimSpace(phone)
		if phone == "" {
			continue
		}

		normalized := NormalizePhone(phone)
		if normalized == "" || len(normalized) < 5 {
			result.Invalid = append(result.Invalid, phone)
			continue
		}

		contact := models.Contact{
			PhoneNumber:     phone,
			PhoneNormalized: normalized,
			PollID:          &data.PollID,
		}

		if err := db.Create(&contact).Error; err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key") {
				result.Duplicate = append(result.Duplicate, phone)
			} else {
				result.Invalid = append(result.Invalid, phone)
			}
			continue
		}
		result.Created = append(result.Created, contact)
	}

	return result, nil
}
