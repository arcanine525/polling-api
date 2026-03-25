package tests

import (
	"encoding/json"
	"testing"

	"polling-system/polling/models"
	"polling-system/polling/services"
	"polling-system/tests/setup"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateQuestion tests question creation functionality
func TestCreateQuestion(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	// Create a question to reference as default next question
	existingQuestion, err := testDB.CreateTestQuestion(poll.ID, "Existing question", "text")
	require.NoError(t, err)

	tests := []struct {
		name        string
		pollID      string
		data        models.PollQuestionCreate
		wantErr     bool
		errContains string
	}{
		{
			name:   "Create text question",
			pollID: poll.ID,
			data: models.PollQuestionCreate{
				Question:   "What is your name?",
				AnswerType: "text",
			},
			wantErr: false,
		},
		{
			name:   "Create multiple choice question",
			pollID: poll.ID,
			data: models.PollQuestionCreate{
				Question:   "What is your favorite color?",
				AnswerType: "multiple_choice",
				AnswerOptions: rawMsgPtr(`["Red", "Blue", "Green", "Yellow"]`),
			},
			wantErr: false,
		},
		{
			name:   "Create question with all optional fields",
			pollID: poll.ID,
			data: models.PollQuestionCreate{
				Question:              "Full question?",
				AnswerType:            "text",
				TypeformFieldID:       strPtr("field_abc123"),
				AnswerOptions:         rawMsgPtr(`["Option A", "Option B"]`),
				DefaultNextQuestionID: &existingQuestion.ID,
				FlowRules:             rawMsgPtr(`[{"condition": "answer == 'yes'", "next_question_id": "` + existingQuestion.ID + `"}]`),
				RandomizeAnswers:      boolPtr(true),
			},
			wantErr: false,
		},
		{
			name:   "Create question for non-existent poll",
			pollID: "non-existent-poll-id",
			data: models.PollQuestionCreate{
				Question:   "Test question?",
				AnswerType: "text",
			},
			wantErr:     true,
			errContains: "poll not found",
		},
		{
			name:   "Create question with invalid default_next_question_id",
			pollID: poll.ID,
			data: models.PollQuestionCreate{
				Question:              "Test question?",
				AnswerType:            "text",
				DefaultNextQuestionID: strPtr("non-existent-question-id"),
			},
			wantErr:     true,
			errContains: "default_next_question_id must belong to the same poll",
		},
		{
			name:   "Create question with default_next_question_id from different poll",
			pollID: poll.ID,
			data: models.PollQuestionCreate{
				Question:              "Test question?",
				AnswerType:            "text",
				DefaultNextQuestionID: strPtr("different-poll-question-id"),
			},
			wantErr:     true,
			errContains: "default_next_question_id must belong to the same poll",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			question, err := services.CreateQuestion(testDB.DB, tt.pollID, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, question)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, question)
				assert.NotEmpty(t, question.ID)
				assert.Equal(t, tt.pollID, question.PollID)
				assert.Equal(t, tt.data.Question, question.Question)
				assert.Equal(t, tt.data.AnswerType, question.AnswerType)
				assert.NotZero(t, question.CreatedAt)

				// Check that FlowRules defaults to empty array if not provided
				if tt.data.FlowRules == nil {
					assert.Equal(t, json.RawMessage("[]"), question.FlowRules)
				}
			}
		})
	}
}

// TestGetQuestionsByPoll tests listing questions for a poll
func TestGetQuestionsByPoll(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll1, err := testDB.CreateTestPoll(user.ID, "Poll 1", "active")
	require.NoError(t, err)

	poll2, err := testDB.CreateTestPoll(user.ID, "Poll 2", "active")
	require.NoError(t, err)

	// Create questions for poll1
	for i := 0; i < 5; i++ {
		_, err := testDB.CreateTestQuestion(poll1.ID, "Question "+string(rune('A'+i)), "text")
		require.NoError(t, err)
	}

	// Create questions for poll2
	for i := 0; i < 3; i++ {
		_, err := testDB.CreateTestQuestion(poll2.ID, "Question "+string(rune('1'+i)), "text")
		require.NoError(t, err)
	}

	tests := []struct {
		name         string
		pollID       string
		page         int
		size         int
		expectTotal  int64
		expectLength int
	}{
		{
			name:         "List all questions for poll1",
			pollID:       poll1.ID,
			page:         1,
			size:         10,
			expectTotal:  5,
			expectLength: 5,
		},
		{
			name:         "List all questions for poll2",
			pollID:       poll2.ID,
			page:         1,
			size:         10,
			expectTotal:  3,
			expectLength: 3,
		},
		{
			name:         "List with pagination - first page",
			pollID:       poll1.ID,
			page:         1,
			size:         2,
			expectTotal:  5,
			expectLength: 2,
		},
		{
			name:         "List with pagination - second page",
			pollID:       poll1.ID,
			page:         2,
			size:         2,
			expectTotal:  5,
			expectLength: 2,
		},
		{
			name:         "List questions for poll with no questions",
			pollID:       poll2.ID,
			page:         1,
			size:         10,
			expectTotal:  3,
			expectLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := services.GetQuestionsByPoll(testDB.DB, tt.pollID, tt.page, tt.size)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectTotal, result.Total)
			assert.Equal(t, tt.page, result.Page)
			assert.Equal(t, tt.size, result.Size)
			assert.Len(t, result.Items, tt.expectLength)
		})
	}
}

// TestGetQuestionByID tests retrieving a single question
func TestGetQuestionByID(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	question, err := testDB.CreateTestQuestion(poll.ID, "What is your name?", "text")
	require.NoError(t, err)

	tests := []struct {
		name        string
		questionID  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "Get existing question",
			questionID: question.ID,
			wantErr:    false,
		},
		{
			name:        "Get non-existent question",
			questionID:  "non-existent-id",
			wantErr:     true,
			errContains: "question not found",
		},
		{
			name:        "Get with empty ID",
			questionID:  "",
			wantErr:     true,
			errContains: "question not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := services.GetQuestionByID(testDB.DB, tt.questionID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.questionID, result.ID)
				assert.Equal(t, "What is your name?", result.Question)
			}
		})
	}
}

// TestUpdateQuestion tests updating a question
func TestUpdateQuestion(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	// Create questions for testing
	question1, err := testDB.CreateTestQuestion(poll.ID, "Original question", "text")
	require.NoError(t, err)

	question2, err := testDB.CreateTestQuestion(poll.ID, "Another question", "text")
	require.NoError(t, err)

	tests := []struct {
		name         string
		setupFunc    func() *models.PollQuestion
		updateData   models.PollQuestionUpdate
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, q *models.PollQuestion)
	}{
		{
			name: "Update question text",
			setupFunc: func() *models.PollQuestion {
				return question1
			},
			updateData: models.PollQuestionUpdate{
				Question: strPtr("Updated question text"),
			},
			wantErr: false,
			validateFunc: func(t *testing.T, q *models.PollQuestion) {
				assert.Equal(t, "Updated question text", q.Question)
			},
		},
		{
			name: "Update answer type",
			setupFunc: func() *models.PollQuestion {
				q, _ := testDB.CreateTestQuestion(poll.ID, "Test", "text")
				return q
			},
			updateData: models.PollQuestionUpdate{
				AnswerType: strPtr("multiple_choice"),
			},
			wantErr: false,
			validateFunc: func(t *testing.T, q *models.PollQuestion) {
				assert.Equal(t, "multiple_choice", q.AnswerType)
			},
		},
		{
			name: "Update with valid default_next_question_id",
			setupFunc: func() *models.PollQuestion {
				q, _ := testDB.CreateTestQuestion(poll.ID, "Test", "text")
				return q
			},
			updateData: models.PollQuestionUpdate{
				DefaultNextQuestionID: &question2.ID,
			},
			wantErr: false,
			validateFunc: func(t *testing.T, q *models.PollQuestion) {
				assert.NotNil(t, q.DefaultNextQuestionID)
				assert.Equal(t, question2.ID, *q.DefaultNextQuestionID)
			},
		},
		{
			name: "Update with invalid default_next_question_id",
			setupFunc: func() *models.PollQuestion {
				q, _ := testDB.CreateTestQuestion(poll.ID, "Test", "text")
				return q
			},
			updateData: models.PollQuestionUpdate{
				DefaultNextQuestionID: strPtr("non-existent-id"),
			},
			wantErr:     true,
			errContains: "default_next_question_id must belong to the same poll",
		},
		{
			name: "Update multiple fields",
			setupFunc: func() *models.PollQuestion {
				q, _ := testDB.CreateTestQuestion(poll.ID, "Test", "text")
				return q
			},
			updateData: models.PollQuestionUpdate{
				Question:         strPtr("Multi update question"),
				AnswerType:       strPtr("multiple_choice"),
				RandomizeAnswers: boolPtr(true),
			},
			wantErr: false,
			validateFunc: func(t *testing.T, q *models.PollQuestion) {
				assert.Equal(t, "Multi update question", q.Question)
				assert.Equal(t, "multiple_choice", q.AnswerType)
				assert.True(t, q.RandomizeAnswers)
			},
		},
		{
			name: "Update non-existent question",
			setupFunc: func() *models.PollQuestion {
				return &models.PollQuestion{ID: "non-existent-id"}
			},
			updateData: models.PollQuestionUpdate{
				Question: strPtr("Updated"),
			},
			wantErr:     true,
			errContains: "question not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			question := tt.setupFunc()

			result, err := services.UpdateQuestion(testDB.DB, question.ID, tt.updateData)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateFunc != nil {
					tt.validateFunc(t, result)
				}
			}
		})
	}
}

// TestDeleteQuestion tests deleting a question
func TestDeleteQuestion(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	t.Run("Delete existing question", func(t *testing.T) {
		question, err := testDB.CreateTestQuestion(poll.ID, "To delete", "text")
		require.NoError(t, err)

		err = services.DeleteQuestion(testDB.DB, question.ID)
		assert.NoError(t, err)

		// Verify question is deleted
		_, err = services.GetQuestionByID(testDB.DB, question.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "question not found")
	})

	t.Run("Delete non-existent question", func(t *testing.T) {
		err := services.DeleteQuestion(testDB.DB, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "question not found")
	})

	t.Run("Delete question cascades to answers", func(t *testing.T) {
		question, err := testDB.CreateTestQuestion(poll.ID, "Cascade test", "text")
		require.NoError(t, err)

		candidate, err := testDB.CreateTestCandidate(poll.ID, "+61412345678")
		require.NoError(t, err)

		answer, err := testDB.CreateTestAnswer(question.ID, candidate.ID, "My answer")
		require.NoError(t, err)

		// Delete the question
		err = services.DeleteQuestion(testDB.DB, question.ID)
		assert.NoError(t, err)

		// Verify answer is deleted
		var count int64
		testDB.DB.Model(&models.Answer{}).Where("id = ?", answer.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}
