package handlers

import (
	"net/http"

	"polling-system/database"
	"polling-system/handlers"
	"polling-system/polling/models"
	"polling-system/polling/services"

	"github.com/gin-gonic/gin"
)

// CreateQuestion handles POST /api/v1/polls/:pollId/questions
// @Summary      Add question to poll
// @Description  Adds a question to a poll. Requires poll ownership for pollers.
// @Tags         Questions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        pollId  path  string                     true  "Poll UUID"
// @Param        body    body  models.PollQuestionCreate   true  "Question to add"
// @Success      201  {object}  models.PollQuestion
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /polls/{pollId}/questions [post]
func CreateQuestion(c *gin.Context) {
	pollID := c.Param("id")

	if !CheckPollOwnership(c, pollID) {
		return
	}

	var data models.PollQuestionCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	question, err := services.CreateQuestion(database.DB, pollID, data)
	if err != nil {
		if err.Error() == "poll not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "default_next_question_id must belong to the same poll" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, question)
}

// ListQuestions handles GET /api/v1/polls/:pollId/questions
// @Summary      List questions for a poll
// @Description  Returns a paginated list of questions for the given poll.
// @Tags         Questions
// @Produce      json
// @Security     BearerAuth
// @Param        pollId  path   string  true   "Poll UUID"
// @Param        page    query  int     false  "Page number (min: 1)"    default(1)
// @Param        size    query  int     false  "Items per page (1-100)"  default(20)
// @Success      200  {object}  models.PaginatedQuestions
// @Failure      500  {object}  models.ErrorResponse
// @Router       /polls/{pollId}/questions [get]
func ListQuestions(c *gin.Context) {
	pollID := c.Param("id")
	page, size := handlers.ParsePagination(c)

	result, err := services.GetQuestionsByPoll(database.DB, pollID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetQuestion handles GET /api/v1/questions/:id
// @Summary      Get question by ID
// @Description  Returns a single question.
// @Tags         Questions
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Question UUID"
// @Success      200  {object}  models.PollQuestion
// @Failure      404  {object}  models.ErrorResponse
// @Router       /questions/{id} [get]
func GetQuestion(c *gin.Context) {
	id := c.Param("id")

	question, err := services.GetQuestionByID(database.DB, id)
	if err != nil {
		if err.Error() == "question not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, question)
}

// UpdateQuestion handles PATCH /api/v1/questions/:id
// @Summary      Update a question
// @Description  Partially updates a question. Requires ownership of the parent poll for pollers.
// @Tags         Questions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                      true  "Question UUID"
// @Param        body  body  models.PollQuestionUpdate    true  "Fields to update"
// @Success      200  {object}  models.PollQuestion
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /questions/{id} [patch]
func UpdateQuestion(c *gin.Context) {
	id := c.Param("id")

	// Ownership check: question → poll → created_by
	question, err := services.GetQuestionByID(database.DB, id)
	if err != nil {
		if err.Error() == "question not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !CheckPollOwnership(c, question.PollID) {
		return
	}

	var data models.PollQuestionUpdate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := services.UpdateQuestion(database.DB, id, data)
	if err != nil {
		if err.Error() == "default_next_question_id must belong to the same poll" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// DeleteQuestion handles DELETE /api/v1/questions/:id
// @Summary      Delete a question
// @Description  Deletes a question and its answers. Requires ownership of the parent poll for pollers.
// @Tags         Questions
// @Security     BearerAuth
// @Param        id  path  string  true  "Question UUID"
// @Success      204  "No Content"
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /questions/{id} [delete]
func DeleteQuestion(c *gin.Context) {
	id := c.Param("id")

	// Ownership check
	question, err := services.GetQuestionByID(database.DB, id)
	if err != nil {
		if err.Error() == "question not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !CheckPollOwnership(c, question.PollID) {
		return
	}

	if err := services.DeleteQuestion(database.DB, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
