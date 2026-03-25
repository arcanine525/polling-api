package setup

import (
	"polling-system/auth/models"
	pollingmodels "polling-system/polling/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestDB wraps the test database connection
type TestDB struct {
	DB *gorm.DB
}

// NewTestDB creates a new in-memory SQLite database for testing
func NewTestDB() (*TestDB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:?_foreign_keys=on"), &gorm.Config{
		Logger: logger.Default,
	})
	if err != nil {
		return nil, err
	}

	// Auto migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&pollingmodels.Poll{},
		&pollingmodels.Candidate{},
		&pollingmodels.PollQuestion{},
		&pollingmodels.Answer{},
	)
	if err != nil {
		return nil, err
	}

	return &TestDB{DB: db}, nil
}

// Close closes the database connection
func (t *TestDB) Close() error {
	sqlDB, err := t.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Cleanup clears all tables
func (t *TestDB) Cleanup() {
	t.DB.Exec("DELETE FROM answers")
	t.DB.Exec("DELETE FROM poll_questions")
	t.DB.Exec("DELETE FROM candidates")
	t.DB.Exec("DELETE FROM polls")
	t.DB.Exec("DELETE FROM users")
}

// CreateTestUser creates a test user and returns the user
func (t *TestDB) CreateTestUser(firebaseUID, email, role string) (*models.User, error) {
	user := models.User{
		FirebaseUID: firebaseUID,
		Email:       email,
		Role:        role,
	}
	if err := t.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateTestPoll creates a test poll and returns it
func (t *TestDB) CreateTestPoll(userID, pollName, status string) (*pollingmodels.Poll, error) {
	poll := pollingmodels.Poll{
		PollName:             pollName,
		Status:               status,
		CreatedBy:            userID,
		PollerName:           "Test Poller",
		FirstTypeformSMSText: "Please answer our poll",
		VoicePollMethod:      "livekit",
	}
	if err := t.DB.Create(&poll).Error; err != nil {
		return nil, err
	}
	return &poll, nil
}

// CreateTestCandidate creates a test candidate and returns it
func (t *TestDB) CreateTestCandidate(pollID, phone string) (*pollingmodels.Candidate, error) {
	candidate := pollingmodels.Candidate{
		PollID: pollID,
		Phone:  phone,
	}
	if err := t.DB.Create(&candidate).Error; err != nil {
		return nil, err
	}
	return &candidate, nil
}

// CreateTestQuestion creates a test question and returns it
func (t *TestDB) CreateTestQuestion(pollID, question, answerType string) (*pollingmodels.PollQuestion, error) {
	q := pollingmodels.PollQuestion{
		PollID:     pollID,
		Question:   question,
		AnswerType: answerType,
	}
	if err := t.DB.Create(&q).Error; err != nil {
		return nil, err
	}
	return &q, nil
}

// CreateTestAnswer creates a test answer and returns it
func (t *TestDB) CreateTestAnswer(questionID, candidateID, answer string) (*pollingmodels.Answer, error) {
	a := pollingmodels.Answer{
		QuestionID:  questionID,
		CandidateID: candidateID,
		Answer:      answer,
		Source:      "typeform",
	}
	if err := t.DB.Create(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}
