package tests

import (
	"testing"

	"polling-system/polling/models"
	"polling-system/polling/services"
	"polling-system/tests/setup"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateAnswer tests answer creation functionality (voting in a poll)
func TestCreateAnswer(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	question, err := testDB.CreateTestQuestion(poll.ID, "What is your favorite color?", "multiple_choice")
	require.NoError(t, err)

	candidate, err := testDB.CreateTestCandidate(poll.ID, "+61412345678")
	require.NoError(t, err)

	t.Run("Submit valid answer", func(t *testing.T) {
		data := models.AnswerCreate{
			CandidateID: candidate.ID,
			Answer:      "Blue",
		}

		answer, err := services.CreateAnswer(testDB.DB, question.ID, data)
		assert.NoError(t, err)
		assert.NotNil(t, answer)
		assert.NotEmpty(t, answer.ID)
		assert.Equal(t, question.ID, answer.QuestionID)
		assert.Equal(t, candidate.ID, answer.CandidateID)
		assert.Equal(t, "Blue", answer.Answer)
		assert.Equal(t, "typeform", answer.Source)
	})

	t.Run("Submit answer with source", func(t *testing.T) {
		cand, _ := testDB.CreateTestCandidate(poll.ID, "+61411111111")
		q, _ := testDB.CreateTestQuestion(poll.ID, "Question with source", "text")

		source := "voice"
		data := models.AnswerCreate{
			CandidateID: cand.ID,
			Answer:      "Red",
			Source:      &source,
		}

		answer, err := services.CreateAnswer(testDB.DB, q.ID, data)
		assert.NoError(t, err)
		assert.NotNil(t, answer)
		assert.Equal(t, "voice", answer.Source)
	})

	t.Run("Submit answer for non-existent question", func(t *testing.T) {
		cand, _ := testDB.CreateTestCandidate(poll.ID, "+61422222222")

		data := models.AnswerCreate{
			CandidateID: cand.ID,
			Answer:      "Blue",
		}

		answer, err := services.CreateAnswer(testDB.DB, "non-existent-question-id", data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "question not found")
		assert.Nil(t, answer)
	})

	t.Run("Submit answer for non-existent candidate", func(t *testing.T) {
		q, _ := testDB.CreateTestQuestion(poll.ID, "Question for non-existent candidate", "text")

		data := models.AnswerCreate{
			CandidateID: "non-existent-candidate-id",
			Answer:      "Blue",
		}

		answer, err := services.CreateAnswer(testDB.DB, q.ID, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "candidate not found")
		assert.Nil(t, answer)
	})

	t.Run("Submit duplicate answer (same candidate, same question)", func(t *testing.T) {
		cand, _ := testDB.CreateTestCandidate(poll.ID, "+61433333333")
		q, _ := testDB.CreateTestQuestion(poll.ID, "Duplicate test question", "text")

		// First answer should succeed
		data1 := models.AnswerCreate{
			CandidateID: cand.ID,
			Answer:      "First answer",
		}
		answer1, err := services.CreateAnswer(testDB.DB, q.ID, data1)
		assert.NoError(t, err)
		assert.NotNil(t, answer1)

		// Second answer from same candidate to same question should fail
		data2 := models.AnswerCreate{
			CandidateID: cand.ID,
			Answer:      "Second answer",
		}
		answer2, err := services.CreateAnswer(testDB.DB, q.ID, data2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "candidate has already answered this question")
		assert.Nil(t, answer2)
	})
}

// TestGetAnswersByQuestion tests listing answers for a question (viewing poll results)
func TestGetAnswersByQuestion(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	question1, err := testDB.CreateTestQuestion(poll.ID, "Question 1", "multiple_choice")
	require.NoError(t, err)

	question2, err := testDB.CreateTestQuestion(poll.ID, "Question 2", "text")
	require.NoError(t, err)

	// Create candidates and answers for question1
	for i := 0; i < 5; i++ {
		candidate, err := testDB.CreateTestCandidate(poll.ID, "+6141234567"+string(rune('0'+i)))
		require.NoError(t, err)
		_, err = testDB.CreateTestAnswer(question1.ID, candidate.ID, "Answer "+string(rune('A'+i)))
		require.NoError(t, err)
	}

	// Create answers for question2
	for i := 0; i < 3; i++ {
		candidate, err := testDB.CreateTestCandidate(poll.ID, "+6149876543"+string(rune('0'+i)))
		require.NoError(t, err)
		_, err = testDB.CreateTestAnswer(question2.ID, candidate.ID, "Text answer "+string(rune('1'+i)))
		require.NoError(t, err)
	}

	tests := []struct {
		name         string
		questionID   string
		page         int
		size         int
		expectTotal  int64
		expectLength int
	}{
		{
			name:         "List all answers for question1",
			questionID:   question1.ID,
			page:         1,
			size:         10,
			expectTotal:  5,
			expectLength: 5,
		},
		{
			name:         "List all answers for question2",
			questionID:   question2.ID,
			page:         1,
			size:         10,
			expectTotal:  3,
			expectLength: 3,
		},
		{
			name:         "List with pagination - first page",
			questionID:   question1.ID,
			page:         1,
			size:         2,
			expectTotal:  5,
			expectLength: 2,
		},
		{
			name:         "List with pagination - second page",
			questionID:   question1.ID,
			page:         2,
			size:         2,
			expectTotal:  5,
			expectLength: 2,
		},
		{
			name:         "List answers for question with no answers",
			questionID:   question2.ID,
			page:         1,
			size:         10,
			expectTotal:  3,
			expectLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := services.GetAnswersByQuestion(testDB.DB, tt.questionID, tt.page, tt.size)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectTotal, result.Total)
			assert.Equal(t, tt.page, result.Page)
			assert.Equal(t, tt.size, result.Size)
			assert.Len(t, result.Items, tt.expectLength)
		})
	}
}

// TestGetAnswersByCandidate tests listing answers by a candidate
func TestGetAnswersByCandidate(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	// Create multiple questions
	questions := make([]*models.PollQuestion, 5)
	for i := 0; i < 5; i++ {
		q, err := testDB.CreateTestQuestion(poll.ID, "Question "+string(rune('A'+i)), "text")
		require.NoError(t, err)
		questions[i] = q
	}

	// Create candidate1 who answers 3 questions
	candidate1, err := testDB.CreateTestCandidate(poll.ID, "+61412345678")
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		_, err := testDB.CreateTestAnswer(questions[i].ID, candidate1.ID, "Answer "+string(rune('A'+i)))
		require.NoError(t, err)
	}

	// Create candidate2 who answers 2 questions
	candidate2, err := testDB.CreateTestCandidate(poll.ID, "+61498765432")
	require.NoError(t, err)
	for i := 0; i < 2; i++ {
		_, err := testDB.CreateTestAnswer(questions[i].ID, candidate2.ID, "Answer "+string(rune('0'+i)))
		require.NoError(t, err)
	}

	// Create candidate3 with no answers
	candidate3, err := testDB.CreateTestCandidate(poll.ID, "+61455555555")
	require.NoError(t, err)

	tests := []struct {
		name         string
		candidateID  string
		page         int
		size         int
		expectTotal  int64
		expectLength int
	}{
		{
			name:         "List all answers for candidate1",
			candidateID:  candidate1.ID,
			page:         1,
			size:         10,
			expectTotal:  3,
			expectLength: 3,
		},
		{
			name:         "List all answers for candidate2",
			candidateID:  candidate2.ID,
			page:         1,
			size:         10,
			expectTotal:  2,
			expectLength: 2,
		},
		{
			name:         "List answers for candidate with no answers",
			candidateID:  candidate3.ID,
			page:         1,
			size:         10,
			expectTotal:  0,
			expectLength: 0,
		},
		{
			name:         "List with pagination",
			candidateID:  candidate1.ID,
			page:         1,
			size:         2,
			expectTotal:  3,
			expectLength: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := services.GetAnswersByCandidate(testDB.DB, tt.candidateID, tt.page, tt.size)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectTotal, result.Total)
			assert.Equal(t, tt.page, result.Page)
			assert.Equal(t, tt.size, result.Size)
			assert.Len(t, result.Items, tt.expectLength)
		})
	}
}

// TestGetAnswerByID tests retrieving a single answer
func TestGetAnswerByID(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	question, err := testDB.CreateTestQuestion(poll.ID, "What is your name?", "text")
	require.NoError(t, err)

	candidate, err := testDB.CreateTestCandidate(poll.ID, "+61412345678")
	require.NoError(t, err)

	answer, err := testDB.CreateTestAnswer(question.ID, candidate.ID, "John Doe")
	require.NoError(t, err)

	tests := []struct {
		name        string
		answerID    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "Get existing answer",
			answerID: answer.ID,
			wantErr:  false,
		},
		{
			name:        "Get non-existent answer",
			answerID:    "non-existent-id",
			wantErr:     true,
			errContains: "answer not found",
		},
		{
			name:        "Get with empty ID",
			answerID:    "",
			wantErr:     true,
			errContains: "answer not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := services.GetAnswerByID(testDB.DB, tt.answerID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.answerID, result.ID)
				assert.Equal(t, "John Doe", result.Answer)
			}
		})
	}
}

// TestDeleteAnswer tests deleting an answer
func TestDeleteAnswer(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	question, err := testDB.CreateTestQuestion(poll.ID, "What is your name?", "text")
	require.NoError(t, err)

	candidate, err := testDB.CreateTestCandidate(poll.ID, "+61412345678")
	require.NoError(t, err)

	t.Run("Delete existing answer", func(t *testing.T) {
		answer, err := testDB.CreateTestAnswer(question.ID, candidate.ID, "To delete")
		require.NoError(t, err)

		err = services.DeleteAnswer(testDB.DB, answer.ID)
		assert.NoError(t, err)

		// Verify answer is deleted
		_, err = services.GetAnswerByID(testDB.DB, answer.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "answer not found")
	})

	t.Run("Delete non-existent answer", func(t *testing.T) {
		err := services.DeleteAnswer(testDB.DB, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "answer not found")
	})
}

// TestAnswerStatistics tests answer statistics calculation (poll results)
func TestAnswerStatistics(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	question, err := testDB.CreateTestQuestion(poll.ID, "What is your favorite color?", "multiple_choice")
	require.NoError(t, err)

	// Create candidates and answers with different choices
	answers := []string{"Blue", "Red", "Blue", "Green", "Blue", "Red", "Blue"}
	for i, answer := range answers {
		candidate, err := testDB.CreateTestCandidate(poll.ID, "+6141234567"+string(rune('0'+i)))
		require.NoError(t, err)
		_, err = testDB.CreateTestAnswer(question.ID, candidate.ID, answer)
		require.NoError(t, err)
	}

	t.Run("Calculate answer distribution", func(t *testing.T) {
		result, err := services.GetAnswersByQuestion(testDB.DB, question.ID, 1, 100)
		require.NoError(t, err)

		// Count answers
		answerCounts := make(map[string]int)
		for _, ans := range result.Items {
			answerCounts[ans.Answer]++
		}

		// Verify counts
		assert.Equal(t, 4, answerCounts["Blue"])
		assert.Equal(t, 2, answerCounts["Red"])
		assert.Equal(t, 1, answerCounts["Green"])
		assert.Equal(t, int64(7), result.Total)

		// Calculate percentages
		bluePercent := float64(answerCounts["Blue"]) / float64(result.Total) * 100
		redPercent := float64(answerCounts["Red"]) / float64(result.Total) * 100
		greenPercent := float64(answerCounts["Green"]) / float64(result.Total) * 100

		assert.InDelta(t, 57.14, bluePercent, 0.1)
		assert.InDelta(t, 28.57, redPercent, 0.1)
		assert.InDelta(t, 14.29, greenPercent, 0.1)
	})
}

// TestViewPollResults_NoVotes tests viewing results for a poll with no votes
func TestViewPollResults_NoVotes(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Empty Poll", "active")
	require.NoError(t, err)

	question, err := testDB.CreateTestQuestion(poll.ID, "What is your favorite color?", "multiple_choice")
	require.NoError(t, err)

	t.Run("View results for poll with no votes", func(t *testing.T) {
		result, err := services.GetAnswersByQuestion(testDB.DB, question.ID, 1, 10)
		require.NoError(t, err)

		assert.Equal(t, int64(0), result.Total)
		assert.Len(t, result.Items, 0)
		assert.Equal(t, 0, result.Pages)
	})
}

// TestDuplicateAnswerPrevention tests that a candidate cannot answer the same question twice
func TestDuplicateAnswerPrevention(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	question, err := testDB.CreateTestQuestion(poll.ID, "Question", "text")
	require.NoError(t, err)

	question2, err := testDB.CreateTestQuestion(poll.ID, "Another question", "text")
	require.NoError(t, err)

	candidate, err := testDB.CreateTestCandidate(poll.ID, "+61412345678")
	require.NoError(t, err)

	// First answer should succeed
	answer1, err := services.CreateAnswer(testDB.DB, question.ID, models.AnswerCreate{
		CandidateID: candidate.ID,
		Answer:      "First answer",
	})
	require.NoError(t, err)
	require.NotNil(t, answer1)

	// Second answer from same candidate to same question should fail
	answer2, err := services.CreateAnswer(testDB.DB, question.ID, models.AnswerCreate{
		CandidateID: candidate.ID,
		Answer:      "Second answer",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "candidate has already answered this question")
	assert.Nil(t, answer2)

	// But answering a different question should succeed
	answer3, err := services.CreateAnswer(testDB.DB, question2.ID, models.AnswerCreate{
		CandidateID: candidate.ID,
		Answer:      "Answer to different question",
	})
	assert.NoError(t, err)
	assert.NotNil(t, answer3)
}
