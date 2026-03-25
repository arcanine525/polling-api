package handlers

import (
	"net/http"

	"polling-system/database"
	"polling-system/handlers"
	"polling-system/middleware"
	"polling-system/polling/models"
	"polling-system/polling/services"

	"github.com/gin-gonic/gin"
)

// CreatePoll handles POST /api/v1/polls
// @Summary      Create a poll
// @Description  Creates a new poll. Sets created_by from the authenticated user.
// @Tags         Polls
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  models.PollCreate  true  "Poll to create"
// @Success      201  {object}  models.Poll
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /polls [post]
func CreatePoll(c *gin.Context) {
	var data models.PollCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	poll, err := services.CreatePoll(database.DB, userID, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, poll)
}

// ListPolls handles GET /api/v1/polls
// @Summary      List polls
// @Description  Returns a paginated list of polls, optionally filtered by status.
// @Tags         Polls
// @Produce      json
// @Security     BearerAuth
// @Param        page    query  int     false  "Page number (min: 1)"    default(1)
// @Param        size    query  int     false  "Items per page (1-100)"  default(20)
// @Param        status  query  string  false  "Filter by status"
// @Success      200  {object}  models.PaginatedPolls
// @Failure      500  {object}  models.ErrorResponse
// @Router       /polls [get]
func ListPolls(c *gin.Context) {
	page, size := handlers.ParsePagination(c)
	status := c.Query("status")

	result, err := services.GetPolls(database.DB, page, size, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetPoll handles GET /api/v1/polls/:id
// @Summary      Get poll by ID
// @Description  Returns a single poll with its creator.
// @Tags         Polls
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Poll UUID"
// @Success      200  {object}  models.Poll
// @Failure      404  {object}  models.ErrorResponse
// @Router       /polls/{id} [get]
func GetPoll(c *gin.Context) {
	id := c.Param("id")

	poll, err := services.GetPollByID(database.DB, id)
	if err != nil {
		if err.Error() == "poll not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, poll)
}

// UpdatePoll handles PATCH /api/v1/polls/:id
// @Summary      Update a poll
// @Description  Partially updates a poll. Admins can update any poll; pollers can only update their own.
// @Tags         Polls
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string            true  "Poll UUID"
// @Param        body  body  models.PollUpdate  true  "Fields to update"
// @Success      200  {object}  models.Poll
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /polls/{id} [patch]
func UpdatePoll(c *gin.Context) {
	id := c.Param("id")

	if !CheckPollOwnership(c, id) {
		return
	}

	var data models.PollUpdate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	poll, err := services.UpdatePoll(database.DB, id, data)
	if err != nil {
		if err.Error() == "poll not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, poll)
}

// DeletePoll handles DELETE /api/v1/polls/:id
// @Summary      Delete a poll
// @Description  Deletes a poll and all its candidates, questions, and answers (cascade). Admin only.
// @Tags         Polls
// @Security     BearerAuth
// @Param        id  path  string  true  "Poll UUID"
// @Success      204  "No Content"
// @Failure      404  {object}  models.ErrorResponse
// @Router       /polls/{id} [delete]
func DeletePoll(c *gin.Context) {
	id := c.Param("id")

	if err := services.DeletePoll(database.DB, id); err != nil {
		if err.Error() == "poll not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// CheckPollOwnership verifies that a poller owns the poll. Admins always pass.
// Returns false and writes an error response if ownership check fails.
func CheckPollOwnership(c *gin.Context, pollID string) bool {
	role := middleware.GetUserRole(c)
	if role == "admin" {
		return true
	}
	userID := middleware.GetUserID(c)
	if !services.IsPollOwner(database.DB, pollID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: you do not own this poll"})
		return false
	}
	return true
}
