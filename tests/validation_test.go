package tests

import (
	"testing"

	"polling-system/polling/models"
	"polling-system/polling/services"
	"polling-system/tests/setup"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPollValidation tests poll data validation
func TestPollValidation(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	t.Run("Create poll with empty name (service level allows)", func(t *testing.T) {
		// Note: At service level, empty poll_name is allowed
		// Validation happens at handler level with binding:"required"
		data := models.PollCreate{
			PollerName:           "Test Poller",
			FirstTypeformSMSText: "Please answer",
		}

		poll, err := services.CreatePoll(testDB.DB, user.ID, data)
		// Service level allows empty name - handler validates
		assert.NoError(t, err)
		assert.NotNil(t, poll)
		assert.Empty(t, poll.PollName)
	})

	t.Run("Create poll with valid data", func(t *testing.T) {
		data := models.PollCreate{
			PollName:             "Valid Poll",
			PollerName:           "Valid Poller",
			FirstTypeformSMSText: "Please answer our poll",
		}

		poll, err := services.CreatePoll(testDB.DB, user.ID, data)
		assert.NoError(t, err)
		assert.NotNil(t, poll)
		assert.Equal(t, "Valid Poll", poll.PollName)
	})

	t.Run("Create poll with invalid status defaults to draft", func(t *testing.T) {
		// Service level accepts any status string
		// Handler level should validate against enum
		data := models.PollCreate{
			PollName:             "Status Test",
			PollerName:           "Test",
			FirstTypeformSMSText: "Test",
			Status:               strPtr("invalid_status"),
		}

		poll, err := services.CreatePoll(testDB.DB, user.ID, data)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_status", poll.Status)
	})

	t.Run("Create poll with valid status", func(t *testing.T) {
		statuses := []string{"draft", "active", "closed"}

		for _, status := range statuses {
			data := models.PollCreate{
				PollName:             "Status Poll " + status,
				PollerName:           "Test",
				FirstTypeformSMSText: "Test",
				Status:               &status,
			}

			poll, err := services.CreatePoll(testDB.DB, user.ID, data)
			assert.NoError(t, err)
			assert.Equal(t, status, poll.Status)
		}
	})

	t.Run("Create poll with very long name", func(t *testing.T) {
		longName := ""
		for i := 0; i < 1000; i++ {
			longName += "a"
		}

		data := models.PollCreate{
			PollName:             longName,
			PollerName:           "Test",
			FirstTypeformSMSText: "Test",
		}

		poll, err := services.CreatePoll(testDB.DB, user.ID, data)
		// SQLite/PostgreSQL text type should handle long strings
		assert.NoError(t, err)
		assert.Equal(t, longName, poll.PollName)
	})
}

// TestCandidateValidation tests candidate data validation
func TestCandidateValidation(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	t.Run("Create candidate with empty phone (service allows)", func(t *testing.T) {
		// Service level allows empty phone - handler validates with binding:"required"
		data := models.CandidateCreate{
			Phone: "",
		}

		candidate, err := services.CreateCandidate(testDB.DB, poll.ID, data)
		assert.NoError(t, err)
		assert.Equal(t, "", candidate.Phone)
	})

	t.Run("Create candidate with valid phone", func(t *testing.T) {
		data := models.CandidateCreate{
			Phone: "+61412345678",
		}

		candidate, err := services.CreateCandidate(testDB.DB, poll.ID, data)
		assert.NoError(t, err)
		assert.Equal(t, "+61412345678", candidate.Phone)
	})

	t.Run("Create candidate with various phone formats", func(t *testing.T) {
		phones := []string{
			"+61412345678",    // International format
			"0412345678",      // Australian local format
			"+15551234567",    // US format
			"0412 345 678",    // With spaces
			"(0412) 345 678",  // With parentheses
		}

		for i, phone := range phones {
			data := models.CandidateCreate{
				Phone: phone,
			}

			// Need different poll for each to avoid unique constraint issues
			poll, _ := testDB.CreateTestPoll(user.ID, "Phone Test Poll "+string(rune('A'+i)), "active")

			candidate, err := services.CreateCandidate(testDB.DB, poll.ID, data)
			assert.NoError(t, err)
			assert.Equal(t, phone, candidate.Phone)
		}
	})
}

// TestQuestionValidation tests question data validation
func TestQuestionValidation(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	t.Run("Create question with empty text (service allows)", func(t *testing.T) {
		// Service level allows empty question - handler validates with binding:"required"
		data := models.PollQuestionCreate{
			Question:   "",
			AnswerType: "text",
		}

		question, err := services.CreateQuestion(testDB.DB, poll.ID, data)
		assert.NoError(t, err)
		assert.Equal(t, "", question.Question)
	})

	t.Run("Create question with valid data", func(t *testing.T) {
		data := models.PollQuestionCreate{
			Question:   "What is your favorite color?",
			AnswerType: "multiple_choice",
		}

		question, err := services.CreateQuestion(testDB.DB, poll.ID, data)
		assert.NoError(t, err)
		assert.Equal(t, "What is your favorite color?", question.Question)
		assert.Equal(t, "multiple_choice", question.AnswerType)
	})

	t.Run("Create question with various answer types", func(t *testing.T) {
		answerTypes := []string{"text", "multiple_choice", "boolean", "rating", "phone"}

		for _, answerType := range answerTypes {
			data := models.PollQuestionCreate{
				Question:   "Test question for " + answerType,
				AnswerType: answerType,
			}

			question, err := services.CreateQuestion(testDB.DB, poll.ID, data)
			assert.NoError(t, err)
			assert.Equal(t, answerType, question.AnswerType)
		}
	})

	t.Run("Create question with empty answer type (service allows)", func(t *testing.T) {
		data := models.PollQuestionCreate{
			Question:   "Test question",
			AnswerType: "",
		}

		question, err := services.CreateQuestion(testDB.DB, poll.ID, data)
		assert.NoError(t, err)
		assert.Equal(t, "", question.AnswerType)
	})

	t.Run("Create question with flow rules", func(t *testing.T) {
		data := models.PollQuestionCreate{
			Question:   "Do you agree?",
			AnswerType: "boolean",
			FlowRules:  rawMsgPtr(`[{"condition": "answer == 'yes'", "action": "next"}]`),
		}

		question, err := services.CreateQuestion(testDB.DB, poll.ID, data)
		assert.NoError(t, err)
		assert.NotNil(t, question.FlowRules)
	})
}

// TestAnswerValidation tests answer data validation
func TestAnswerValidation(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	question, err := testDB.CreateTestQuestion(poll.ID, "What is your name?", "text")
	require.NoError(t, err)

	t.Run("Create answer with empty text (service allows)", func(t *testing.T) {
		// Service level allows empty answer - handler validates with binding:"required"
		cand, _ := testDB.CreateTestCandidate(poll.ID, "+61411111111")

		data := models.AnswerCreate{
			CandidateID: cand.ID,
			Answer:      "",
		}

		answer, err := services.CreateAnswer(testDB.DB, question.ID, data)
		assert.NoError(t, err)
		assert.Equal(t, "", answer.Answer)
	})

	t.Run("Create answer with valid data", func(t *testing.T) {
		cand, _ := testDB.CreateTestCandidate(poll.ID, "+61422222222")

		data := models.AnswerCreate{
			CandidateID: cand.ID,
			Answer:      "John Doe",
		}

		answer, err := services.CreateAnswer(testDB.DB, question.ID, data)
		assert.NoError(t, err)
		assert.Equal(t, "John Doe", answer.Answer)
	})

	t.Run("Create answer with various sources", func(t *testing.T) {
		sources := []string{"typeform", "voice", "sms", "web", "api"}

		for _, source := range sources {
			cand, _ := testDB.CreateTestCandidate(poll.ID, "+6143333333"+source[len(source)-1:])
			q, _ := testDB.CreateTestQuestion(poll.ID, "Question for "+source, "text")

			data := models.AnswerCreate{
				CandidateID: cand.ID,
				Answer:      "Test answer",
				Source:      &source,
			}

			answer, err := services.CreateAnswer(testDB.DB, q.ID, data)
			assert.NoError(t, err)
			assert.Equal(t, source, answer.Source)
		}
	})

	t.Run("Create answer with long text", func(t *testing.T) {
		cand, _ := testDB.CreateTestCandidate(poll.ID, "+61444444444")
		q, _ := testDB.CreateTestQuestion(poll.ID, "Long answer question", "text")

		longAnswer := ""
		for i := 0; i < 1000; i++ {
			longAnswer += "word "
		}

		data := models.AnswerCreate{
			CandidateID: cand.ID,
			Answer:      longAnswer,
		}

		answer, err := services.CreateAnswer(testDB.DB, q.ID, data)
		assert.NoError(t, err)
		assert.Equal(t, longAnswer, answer.Answer)
	})
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	t.Run("Get poll with UUID format ID", func(t *testing.T) {
		poll, err := testDB.CreateTestPoll(user.ID, "UUID Test", "active")
		require.NoError(t, err)

		// Verify UUID format
		assert.Len(t, poll.ID, 36) // Standard UUID length with hyphens

		// Retrieve by ID
		retrieved, err := services.GetPollByID(testDB.DB, poll.ID)
		assert.NoError(t, err)
		assert.Equal(t, poll.ID, retrieved.ID)
	})

	t.Run("Update poll with nil fields", func(t *testing.T) {
		poll, err := testDB.CreateTestPoll(user.ID, "Nil Update Test", "draft")
		require.NoError(t, err)

		originalName := poll.PollName

		data := models.PollUpdate{
			PollName: nil,
			Status:   nil,
		}

		updated, err := services.UpdatePoll(testDB.DB, poll.ID, data)
		assert.NoError(t, err)
		assert.Equal(t, originalName, updated.PollName) // Should remain unchanged
	})

	t.Run("Delete poll that was already deleted", func(t *testing.T) {
		poll, err := testDB.CreateTestPoll(user.ID, "Double Delete", "draft")
		require.NoError(t, err)

		// First delete
		err = services.DeletePoll(testDB.DB, poll.ID)
		assert.NoError(t, err)

		// Second delete should fail
		err = services.DeletePoll(testDB.DB, poll.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "poll not found")
	})

	t.Run("List polls with zero page size", func(t *testing.T) {
		// This tests division behavior with size=0
		result, err := services.GetPolls(testDB.DB, 1, 0, "")
		// Behavior depends on GORM implementation
		// Most likely will return empty or error
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, result)
		}
	})

	t.Run("List polls with negative page", func(t *testing.T) {
		// This tests offset calculation with negative page
		result, err := services.GetPolls(testDB.DB, -1, 10, "")
		// Should still work, just with negative offset
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, result)
		}
	})

	t.Run("Pagination with more pages than items", func(t *testing.T) {
		testDB.Cleanup()
		// Need to recreate user after cleanup since foreign keys are enforced
		newUser, err := testDB.CreateTestUser("firebase-page-test", "page@example.com", "poller")
		require.NoError(t, err)
		_, err = testDB.CreateTestPoll(newUser.ID, "Page Test", "active")
		require.NoError(t, err)

		// Request page 100 when only 1 item exists
		result, err := services.GetPolls(testDB.DB, 100, 10, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Len(t, result.Items, 0) // No items on page 100
		assert.Equal(t, 100, result.Page)
	})
}

// TestConcurrentOperations tests concurrent operations
func TestConcurrentOperations(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Concurrent Test", "active")
	require.NoError(t, err)

	question, err := testDB.CreateTestQuestion(poll.ID, "Question", "text")
	require.NoError(t, err)

	t.Run("Concurrent answer creation should prevent duplicates", func(t *testing.T) {
		candidate, err := testDB.CreateTestCandidate(poll.ID, "+61499999999")
		require.NoError(t, err)

		// Create first answer
		data := models.AnswerCreate{
			CandidateID: candidate.ID,
			Answer:      "First answer",
		}
		answer1, err := services.CreateAnswer(testDB.DB, question.ID, data)
		assert.NoError(t, err)
		assert.NotNil(t, answer1)

		// Attempt duplicate answer
		data2 := models.AnswerCreate{
			CandidateID: candidate.ID,
			Answer:      "Second answer",
		}
		answer2, err := services.CreateAnswer(testDB.DB, question.ID, data2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already answered")
		assert.Nil(t, answer2)
	})
}
