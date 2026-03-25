package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"polling-system/auth/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetUserByID returns a user by database ID.
func GetUserByID(db *gorm.DB, id string) (*models.User, error) {
	var user models.User
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByFirebaseUID returns a user by Firebase UID.
func GetUserByFirebaseUID(db *gorm.DB, uid string) (*models.User, error) {
	var user models.User
	if err := db.Where("firebase_uid = ?", uid).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetUsers returns a paginated list of users.
func GetUsers(db *gorm.DB, page, size int) (*models.PaginatedUsers, error) {
	var total int64
	if err := db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, err
	}

	pages := 0
	if total > 0 {
		pages = int(math.Ceil(float64(total) / float64(size)))
	}

	var items []models.User
	offset := (page - 1) * size
	if err := db.Order("created_at DESC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, err
	}

	return &models.PaginatedUsers{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
		Pages: pages,
	}, nil
}

// CreateUser creates a new user from admin input.
func CreateUser(db *gorm.DB, data models.UserCreate) (*models.User, error) {
	role := data.Role
	if role == "" {
		role = "viewer"
	}

	user := models.User{
		FirebaseUID: data.FirebaseUID,
		Email:       data.Email,
		DisplayName: data.DisplayName,
		Role:        role,
	}

	if err := db.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("user with this firebase_uid already exists")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUserDisplayName updates the display name of a user.
func UpdateUserDisplayName(db *gorm.DB, id string, data models.UserUpdate) (*models.User, error) {
	user, err := GetUserByID(db, id)
	if err != nil {
		return nil, err
	}

	if data.DisplayName != nil {
		user.DisplayName = data.DisplayName
	}

	now := time.Now().UTC()
	user.UpdatedAt = &now

	if err := db.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUser updates a user (admin). Currently supports display_name.
func UpdateUser(db *gorm.DB, id string, data models.UserUpdate) (*models.User, error) {
	return UpdateUserDisplayName(db, id, data)
}

// UpdateUserRole updates the role of a user.
func UpdateUserRole(db *gorm.DB, id string, data models.UserRoleUpdate) (*models.User, error) {
	user, err := GetUserByID(db, id)
	if err != nil {
		return nil, err
	}

	user.Role = data.Role

	now := time.Now().UTC()
	user.UpdatedAt = &now

	if err := db.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// DeleteUser deletes a user by ID.
func DeleteUser(db *gorm.DB, id string) error {
	result := db.Where("id = ?", id).Delete(&models.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// ProvisionUser upserts a user by firebase_uid.
func ProvisionUser(db *gorm.DB, data models.ProvisionRequest) (*models.User, error) {
	user := models.User{
		FirebaseUID: data.FirebaseUID,
		Email:       data.Email,
	}

	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "firebase_uid"}},
		DoUpdates: clause.AssignmentColumns([]string{"email", "updated_at"}),
	}).Create(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	// Re-read full record to get defaults (role, created_at)
	var fullUser models.User
	if err := db.Where("firebase_uid = ?", data.FirebaseUID).First(&fullUser).Error; err != nil {
		return nil, err
	}
	return &fullUser, nil
}
