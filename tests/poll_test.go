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

// TestCreatePoll tests poll creation functionality
func TestCreatePoll(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	tests := []struct {
		name        string
		data        models.PollCreate
		wantErr     bool
		errContains string
	}{
		{
			name: "Create poll with required fields only",
			data: models.PollCreate{
				PollName:             "Test Poll",
				PollerName:           "Test Poller",
				FirstTypeformSMSText: "Please answer",
			},
			wantErr: false,
		},
		{
			name: "Create poll with all optional fields",
			data: models.PollCreate{
				PollName:                    "Full Poll",
				PollerName:                  "Full Poller",
				FirstTypeformSMSText:        "Please answer",
				Status:                      strPtr("draft"),
				VoicePollDelayMinutes:       intPtr(10),
				VoicePollMethod:             strPtr("livekit"),
				CustomVoicePollInstructions: strPtr("Custom instructions"),
				DontSendSMS:                 boolPtr(true),
				OnlyDayHours:                boolPtr(false),
				TimeZone:                    strPtr("Australia/Sydney"),
			},
			wantErr: false,
		},
		{
			name: "Create poll with JSON fields",
			data: models.PollCreate{
				PollName:             "JSON Poll",
				PollerName:           "JSON Poller",
				FirstTypeformSMSText: "Please answer",
				TypeformForms:        rawMsgPtr(`{"form1": "abc123"}`),
				BlandPathwayIDs:      rawMsgPtr(`["path1", "path2"]`),
			},
			wantErr: false,
		},
		{
			name: "Create poll with empty name should still work (validation at handler level)",
			data: models.PollCreate{
				PollerName:           "Test Poller",
				FirstTypeformSMSText: "Please answer",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDB.Cleanup()
			user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
			require.NoError(t, err)

			poll, err := services.CreatePoll(testDB.DB, user.ID, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, poll)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, poll)
				assert.NotEmpty(t, poll.ID)
				assert.Equal(t, user.ID, poll.CreatedBy)

				if tt.data.PollName != "" {
					assert.Equal(t, tt.data.PollName, poll.PollName)
				}
				if tt.data.Status != nil {
					assert.Equal(t, *tt.data.Status, poll.Status)
				}
				assert.NotZero(t, poll.CreatedAt)
			}
		})
	}
}

// TestGetPolls tests listing polls
func TestGetPolls(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	// Create test user
	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	// Create test polls with different statuses
	_, err = testDB.CreateTestPoll(user.ID, "Draft Poll 1", "draft")
	require.NoError(t, err)
	_, err = testDB.CreateTestPoll(user.ID, "Active Poll 1", "active")
	require.NoError(t, err)
	_, err = testDB.CreateTestPoll(user.ID, "Active Poll 2", "active")
	require.NoError(t, err)
	_, err = testDB.CreateTestPoll(user.ID, "Closed Poll 1", "closed")
	require.NoError(t, err)

	tests := []struct {
		name         string
		page         int
		size         int
		status       string
		expectTotal  int64
		expectPages  int
		expectLength int
	}{
		{
			name:         "List all polls",
			page:         1,
			size:         10,
			status:       "",
			expectTotal:  4,
			expectPages:  1,
			expectLength: 4,
		},
		{
			name:         "List only active polls",
			page:         1,
			size:         10,
			status:       "active",
			expectTotal:  2,
			expectPages:  1,
			expectLength: 2,
		},
		{
			name:         "List only draft polls",
			page:         1,
			size:         10,
			status:       "draft",
			expectTotal:  1,
			expectPages:  1,
			expectLength: 1,
		},
		{
			name:         "List only closed polls",
			page:         1,
			size:         10,
			status:       "closed",
			expectTotal:  1,
			expectPages:  1,
			expectLength: 1,
		},
		{
			name:         "List with pagination - page 1",
			page:         1,
			size:         2,
			status:       "",
			expectTotal:  4,
			expectPages:  2,
			expectLength: 2,
		},
		{
			name:         "List with pagination - page 2",
			page:         2,
			size:         2,
			status:       "",
			expectTotal:  4,
			expectPages:  2,
			expectLength: 2,
		},
		{
			name:         "Filter non-existent status returns empty",
			page:         1,
			size:         10,
			status:       "nonexistent",
			expectTotal:  0,
			expectPages:  0,
			expectLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := services.GetPolls(testDB.DB, tt.page, tt.size, tt.status)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectTotal, result.Total)
			assert.Equal(t, tt.page, result.Page)
			assert.Equal(t, tt.size, result.Size)
			assert.Equal(t, tt.expectPages, result.Pages)
			assert.Len(t, result.Items, tt.expectLength)
		})
	}
}

// TestGetPollByID tests retrieving a single poll
func TestGetPollByID(t *testing.T) {
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
		wantErr     bool
		errContains string
	}{
		{
			name:    "Get existing poll",
			pollID:  poll.ID,
			wantErr: false,
		},
		{
			name:        "Get non-existent poll",
			pollID:      "non-existent-id",
			wantErr:     true,
			errContains: "poll not found",
		},
		{
			name:        "Get with empty ID",
			pollID:      "",
			wantErr:     true,
			errContains: "poll not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := services.GetPollByID(testDB.DB, tt.pollID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.pollID, result.ID)
				assert.Equal(t, "Test Poll", result.PollName)
			}
		})
	}
}

// TestUpdatePoll tests updating a poll
func TestUpdatePoll(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	tests := []struct {
		name         string
		setupPoll    func() *models.Poll
		updateData   models.PollUpdate
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, poll *models.Poll)
	}{
		{
			name: "Update poll name",
			setupPoll: func() *models.Poll {
				p, _ := testDB.CreateTestPoll(user.ID, "Original Name", "draft")
				return p
			},
			updateData: models.PollUpdate{
				PollName: strPtr("Updated Name"),
			},
			wantErr: false,
			validateFunc: func(t *testing.T, poll *models.Poll) {
				assert.Equal(t, "Updated Name", poll.PollName)
			},
		},
		{
			name: "Update poll status (close poll)",
			setupPoll: func() *models.Poll {
				p, _ := testDB.CreateTestPoll(user.ID, "Test Poll", "active")
				return p
			},
			updateData: models.PollUpdate{
				Status: strPtr("closed"),
			},
			wantErr: false,
			validateFunc: func(t *testing.T, poll *models.Poll) {
				assert.Equal(t, "closed", poll.Status)
			},
		},
		{
			name: "Update multiple fields at once",
			setupPoll: func() *models.Poll {
				p, _ := testDB.CreateTestPoll(user.ID, "Test Poll", "draft")
				return p
			},
			updateData: models.PollUpdate{
				PollName:              strPtr("Multi Update"),
				Status:                strPtr("active"),
				VoicePollDelayMinutes: intPtr(15),
				DontSendSMS:           boolPtr(true),
			},
			wantErr: false,
			validateFunc: func(t *testing.T, poll *models.Poll) {
				assert.Equal(t, "Multi Update", poll.PollName)
				assert.Equal(t, "active", poll.Status)
				assert.Equal(t, 15, poll.VoicePollDelayMinutes)
				assert.True(t, poll.DontSendSMS)
			},
		},
		{
			name: "Update non-existent poll",
			setupPoll: func() *models.Poll {
				return &models.Poll{ID: "non-existent-id"}
			},
			updateData: models.PollUpdate{
				PollName: strPtr("Updated"),
			},
			wantErr:     true,
			errContains: "poll not found",
		},
		{
			name: "Update with empty data (no changes)",
			setupPoll: func() *models.Poll {
				p, _ := testDB.CreateTestPoll(user.ID, "Unchanged", "draft")
				return p
			},
			updateData: models.PollUpdate{},
			wantErr:    false,
			validateFunc: func(t *testing.T, poll *models.Poll) {
				assert.Equal(t, "Unchanged", poll.PollName)
				assert.NotNil(t, poll.UpdatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poll := tt.setupPoll()

			result, err := services.UpdatePoll(testDB.DB, poll.ID, tt.updateData)

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

// TestDeletePoll tests deleting a poll
func TestDeletePoll(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user, err := testDB.CreateTestUser("firebase-123", "test@example.com", "poller")
	require.NoError(t, err)

	t.Run("Delete existing poll", func(t *testing.T) {
		poll, err := testDB.CreateTestPoll(user.ID, "To Delete", "draft")
		require.NoError(t, err)

		err = services.DeletePoll(testDB.DB, poll.ID)
		assert.NoError(t, err)

		// Verify poll is deleted
		_, err = services.GetPollByID(testDB.DB, poll.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "poll not found")
	})

	t.Run("Delete non-existent poll", func(t *testing.T) {
		err := services.DeletePoll(testDB.DB, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "poll not found")
	})

	t.Run("Delete poll cascades to related entities", func(t *testing.T) {
		poll, err := testDB.CreateTestPoll(user.ID, "Cascade Test", "active")
		require.NoError(t, err)

		candidate, err := testDB.CreateTestCandidate(poll.ID, "+61412345678")
		require.NoError(t, err)

		question, err := testDB.CreateTestQuestion(poll.ID, "What do you think?", "text")
		require.NoError(t, err)

		answer, err := testDB.CreateTestAnswer(question.ID, candidate.ID, "My answer")
		require.NoError(t, err)

		// Delete the poll
		err = services.DeletePoll(testDB.DB, poll.ID)
		assert.NoError(t, err)

		// Verify all related entities are deleted
		var count int64
		testDB.DB.Model(&models.Candidate{}).Where("id = ?", candidate.ID).Count(&count)
		assert.Equal(t, int64(0), count)

		testDB.DB.Model(&models.PollQuestion{}).Where("id = ?", question.ID).Count(&count)
		assert.Equal(t, int64(0), count)

		testDB.DB.Model(&models.Answer{}).Where("id = ?", answer.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

// TestIsPollOwner tests poll ownership check
func TestIsPollOwner(t *testing.T) {
	testDB, err := setup.NewTestDB()
	require.NoError(t, err)
	defer testDB.Close()

	user1, err := testDB.CreateTestUser("firebase-1", "user1@example.com", "poller")
	require.NoError(t, err)

	user2, err := testDB.CreateTestUser("firebase-2", "user2@example.com", "poller")
	require.NoError(t, err)

	poll, err := testDB.CreateTestPoll(user1.ID, "Test Poll", "draft")
	require.NoError(t, err)

	tests := []struct {
		name     string
		pollID   string
		userID   string
		expected bool
	}{
		{
			name:     "Owner is the creator",
			pollID:   poll.ID,
			userID:   user1.ID,
			expected: true,
		},
		{
			name:     "Non-owner user",
			pollID:   poll.ID,
			userID:   user2.ID,
			expected: false,
		},
		{
			name:     "Non-existent poll",
			pollID:   "non-existent",
			userID:   user1.ID,
			expected: false,
		},
		{
			name:     "Non-existent user",
			pollID:   poll.ID,
			userID:   "non-existent",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := services.IsPollOwner(testDB.DB, tt.pollID, tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func rawMsgPtr(s string) *json.RawMessage {
	raw := json.RawMessage(s)
	return &raw
}
