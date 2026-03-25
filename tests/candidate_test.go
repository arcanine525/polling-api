package tests

import (
	"testing"

	"polling-system/polling/models"
	"polling-system/polling/services"
	"polling-system/tests/setup"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateCandidate tests candidate creation functionality
func TestCreateCandidate(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	tests := []struct {
		name        string
		pollID      string
		data        models.CandidateCreate
		wantErr     bool
		errContains string
	}{
		{
			name:   "Create candidate with valid phone",
			pollID: poll.ID,
			data: models.CandidateCreate{
				Phone: "+61412345678",
			},
			wantErr: false,
		},
		{
			name:   "Create candidate with different phone format",
			pollID: poll.ID,
			data: models.CandidateCreate{
				Phone: "0412345678",
			},
			wantErr: false,
		},
		{
			name:   "Create candidate with international phone",
			pollID: poll.ID,
			data: models.CandidateCreate{
				Phone: "+15551234567",
			},
			wantErr: false,
		},
		{
			name:   "Create candidate for non-existent poll",
			pollID: "non-existent-poll-id",
			data: models.CandidateCreate{
				Phone: "+61412345678",
			},
			wantErr:     true,
			errContains: "poll not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate, err := services.CreateCandidate(testDB.DB, tt.pollID, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, candidate)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, candidate)
				assert.NotEmpty(t, candidate.ID)
				assert.Equal(t, tt.pollID, candidate.PollID)
				assert.Equal(t, tt.data.Phone, candidate.Phone)
				assert.NotZero(t, candidate.CreatedAt)
			}
		})
	}
}

// TestGetCandidatesByPoll tests listing candidates for a poll
func TestGetCandidatesByPoll(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll1, err := testDB.CreateTestPoll(user.ID, "Poll 1", "active")
	require.NoError(t, err)

	poll2, err := testDB.CreateTestPoll(user.ID, "Poll 2", "active")
	require.NoError(t, err)

	// Create candidates for poll1
	for i := 0; i < 5; i++ {
		_, err := testDB.CreateTestCandidate(poll1.ID, "+6141234567"+string(rune('0'+i)))
		require.NoError(t, err)
	}

	// Create candidates for poll2
	for i := 0; i < 3; i++ {
		_, err := testDB.CreateTestCandidate(poll2.ID, "+6149876543"+string(rune('0'+i)))
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
			name:         "List all candidates for poll1",
			pollID:       poll1.ID,
			page:         1,
			size:         10,
			expectTotal:  5,
			expectLength: 5,
		},
		{
			name:         "List all candidates for poll2",
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
			name:         "List with pagination - last page",
			pollID:       poll1.ID,
			page:         3,
			size:         2,
			expectTotal:  5,
			expectLength: 1,
		},
		{
			name:         "List candidates for poll with no candidates",
			pollID:       poll2.ID,
			page:         1,
			size:         10,
			expectTotal:  3,
			expectLength: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := services.GetCandidatesByPoll(testDB.DB, tt.pollID, tt.page, tt.size)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectTotal, result.Total)
			assert.Equal(t, tt.page, result.Page)
			assert.Equal(t, tt.size, result.Size)
			assert.Len(t, result.Items, tt.expectLength)
		})
	}
}

// TestGetCandidateByID tests retrieving a single candidate
func TestGetCandidateByID(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	candidate, err := testDB.CreateTestCandidate(poll.ID, "+61412345678")
	require.NoError(t, err)

	tests := []struct {
		name         string
		candidateID  string
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, c *models.Candidate)
	}{
		{
			name:        "Get existing candidate",
			candidateID: candidate.ID,
			wantErr:     false,
			validateFunc: func(t *testing.T, c *models.Candidate) {
				assert.Equal(t, "+61412345678", c.Phone)
				assert.Equal(t, poll.ID, c.PollID)
			},
		},
		{
			name:         "Get non-existent candidate",
			candidateID:  "non-existent-id",
			wantErr:      true,
			errContains:  "candidate not found",
		},
		{
			name:         "Get with empty ID",
			candidateID:  "",
			wantErr:      true,
			errContains:  "candidate not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := services.GetCandidateByID(testDB.DB, tt.candidateID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.candidateID, result.ID)
				if tt.validateFunc != nil {
					tt.validateFunc(t, result)
				}
			}
		})
	}
}

// TestUpdateCandidate tests updating a candidate
func TestUpdateCandidate(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	sentiment := "positive"

	tests := []struct {
		name         string
		setupFunc    func() *models.Candidate
		updateData   models.CandidateUpdate
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, c *models.Candidate)
	}{
		{
			name: "Update phone number",
			setupFunc: func() *models.Candidate {
				c, _ := testDB.CreateTestCandidate(poll.ID, "+61412345678")
				return c
			},
			updateData: models.CandidateUpdate{
				Phone: strPtr("+61498765432"),
			},
			wantErr: false,
			validateFunc: func(t *testing.T, c *models.Candidate) {
				assert.Equal(t, "+61498765432", c.Phone)
			},
		},
		{
			name: "Update voice poll sentiment",
			setupFunc: func() *models.Candidate {
				c, _ := testDB.CreateTestCandidate(poll.ID, "+61412345678")
				return c
			},
			updateData: models.CandidateUpdate{
				VoicePollSentiment: &sentiment,
			},
			wantErr: false,
			validateFunc: func(t *testing.T, c *models.Candidate) {
				assert.NotNil(t, c.VoicePollSentiment)
				assert.Equal(t, "positive", *c.VoicePollSentiment)
			},
		},
		{
			name: "Update voicemail detected",
			setupFunc: func() *models.Candidate {
				c, _ := testDB.CreateTestCandidate(poll.ID, "+61412345678")
				return c
			},
			updateData: models.CandidateUpdate{
				VoicemailDetected: boolPtr(true),
			},
			wantErr: false,
			validateFunc: func(t *testing.T, c *models.Candidate) {
				assert.True(t, c.VoicemailDetected)
			},
		},
		{
			name: "Update multiple fields",
			setupFunc: func() *models.Candidate {
				c, _ := testDB.CreateTestCandidate(poll.ID, "+61412345678")
				return c
			},
			updateData: models.CandidateUpdate{
				Phone:              strPtr("+61499999999"),
				VoicemailDetected:  boolPtr(true),
				VoicePollSentiment: strPtr("neutral"),
			},
			wantErr: false,
			validateFunc: func(t *testing.T, c *models.Candidate) {
				assert.Equal(t, "+61499999999", c.Phone)
				assert.True(t, c.VoicemailDetected)
				assert.NotNil(t, c.VoicePollSentiment)
				assert.Equal(t, "neutral", *c.VoicePollSentiment)
			},
		},
		{
			name: "Update non-existent candidate",
			setupFunc: func() *models.Candidate {
				return &models.Candidate{ID: "non-existent-id"}
			},
			updateData: models.CandidateUpdate{
				Phone: strPtr("+61499999999"),
			},
			wantErr:     true,
			errContains: "candidate not found",
		},
		{
			name: "Update with empty data",
			setupFunc: func() *models.Candidate {
				c, _ := testDB.CreateTestCandidate(poll.ID, "+61412345678")
				return c
			},
			updateData: models.CandidateUpdate{},
			wantErr:    false,
			validateFunc: func(t *testing.T, c *models.Candidate) {
				assert.Equal(t, "+61412345678", c.Phone)
				assert.NotNil(t, c.UpdatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := tt.setupFunc()

			result, err := services.UpdateCandidate(testDB.DB, candidate.ID, tt.updateData)

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

// TestDeleteCandidate tests deleting a candidate
func TestDeleteCandidate(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
	require.NoError(t, err)

	t.Run("Delete existing candidate", func(t *testing.T) {
		candidate, err := testDB.CreateTestCandidate(poll.ID, "+61412345678")
		require.NoError(t, err)

		err = services.DeleteCandidate(testDB.DB, candidate.ID)
		assert.NoError(t, err)

		// Verify candidate is deleted
		_, err = services.GetCandidateByID(testDB.DB, candidate.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "candidate not found")
	})

	t.Run("Delete non-existent candidate", func(t *testing.T) {
		err := services.DeleteCandidate(testDB.DB, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "candidate not found")
	})

	t.Run("Delete candidate cascades to answers", func(t *testing.T) {
		candidate, err := testDB.CreateTestCandidate(poll.ID, "+61412345679")
		require.NoError(t, err)

		question, err := testDB.CreateTestQuestion(poll.ID, "What do you think?", "text")
		require.NoError(t, err)

		answer, err := testDB.CreateTestAnswer(question.ID, candidate.ID, "My answer")
		require.NoError(t, err)

		// Delete the candidate
		err = services.DeleteCandidate(testDB.DB, candidate.ID)
		assert.NoError(t, err)

		// Verify answer is deleted
		var count int64
		testDB.DB.Model(&models.Answer{}).Where("id = ?", answer.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}
