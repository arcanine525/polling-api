package handlers

import (
	"net/http"
	"strings"

	"polling-system/database"
	"polling-system/handlers"
	"polling-system/polling/models"
	"polling-system/polling/services"

	"github.com/gin-gonic/gin"
)

// CreateAnswer handles POST /api/v1/questions/:questionId/answers
// @Summary      Submit answer for a question
// @Description  Submits an answer from a candidate. Each candidate can answer a question only once.
// @Tags         Answers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        questionId  path  string               true  "Question UUID"
// @Param        body        body  models.AnswerCreate   true  "Answer to submit"
// @Success      201  {object}  models.Answer
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      409  {object}  models.ErrorResponse
// @Router       /questions/{questionId}/answers [post]
func CreateAnswer(c *gin.Context) {
	questionID := c.Param("id")

	// Ownership check: question → poll → created_by
	question, err := services.GetQuestionByID(database.DB, questionID)
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

	var data models.AnswerCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	answer, err := services.CreateAnswer(database.DB, questionID, data)
	if err != nil {
		if strings.Contains(err.Error(), "already answered") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, answer)
}

// ListAnswersByQuestion handles GET /api/v1/questions/:questionId/answers
// @Summary      List answers for a question
// @Description  Returns a paginated list of answers for the given question.
// @Tags         Answers
// @Produce      json
// @Security     BearerAuth
// @Param        questionId  path   string  true   "Question UUID"
// @Param        page        query  int     false  "Page number (min: 1)"    default(1)
// @Param        size        query  int     false  "Items per page (1-100)"  default(20)
// @Success      200  {object}  models.PaginatedAnswers
// @Failure      500  {object}  models.ErrorResponse
// @Router       /questions/{questionId}/answers [get]
func ListAnswersByQuestion(c *gin.Context) {
	questionID := c.Param("id")
	page, size := handlers.ParsePagination(c)

	result, err := services.GetAnswersByQuestion(database.DB, questionID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ListAnswersByCandidate handles GET /api/v1/candidates/:candidateId/answers
// @Summary      List answers by a candidate
// @Description  Returns a paginated list of all answers submitted by a candidate.
// @Tags         Answers
// @Produce      json
// @Security     BearerAuth
// @Param        candidateId  path   string  true   "Candidate UUID"
// @Param        page         query  int     false  "Page number (min: 1)"    default(1)
// @Param        size         query  int     false  "Items per page (1-100)"  default(20)
// @Success      200  {object}  models.PaginatedAnswers
// @Failure      500  {object}  models.ErrorResponse
// @Router       /candidates/{candidateId}/answers [get]
func ListAnswersByCandidate(c *gin.Context) {
	candidateID := c.Param("id")
	page, size := handlers.ParsePagination(c)

	result, err := services.GetAnswersByCandidate(database.DB, candidateID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetAnswer handles GET /api/v1/answers/:id
// @Summary      Get answer by ID
// @Description  Returns a single answer.
// @Tags         Answers
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Answer UUID"
// @Success      200  {object}  models.Answer
// @Failure      404  {object}  models.ErrorResponse
// @Router       /answers/{id} [get]
func GetAnswer(c *gin.Context) {
	id := c.Param("id")

	answer, err := services.GetAnswerByID(database.DB, id)
	if err != nil {
		if err.Error() == "answer not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, answer)
}

// DeleteAnswer handles DELETE /api/v1/answers/:id
// @Summary      Delete an answer
// @Description  Deletes an answer. Admin only.
// @Tags         Answers
// @Security     BearerAuth
// @Param        id  path  string  true  "Answer UUID"
// @Success      204  "No Content"
// @Failure      404  {object}  models.ErrorResponse
// @Router       /answers/{id} [delete]
func DeleteAnswer(c *gin.Context) {
	id := c.Param("id")

	if err := services.DeleteAnswer(database.DB, id); err != nil {
		if err.Error() == "answer not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
