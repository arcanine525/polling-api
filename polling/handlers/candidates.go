package handlers

import (
	"net/http"

	"polling-system/database"
	"polling-system/handlers"
	"polling-system/polling/models"
	"polling-system/polling/services"

	"github.com/gin-gonic/gin"
)

// CreateCandidate handles POST /api/v1/polls/:pollId/candidates
// @Summary      Add candidate to poll
// @Description  Adds a candidate to a poll. Requires poll ownership for pollers.
// @Tags         Candidates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        pollId  path  string                 true  "Poll UUID"
// @Param        body    body  models.CandidateCreate  true  "Candidate to add"
// @Success      201  {object}  models.Candidate
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /polls/{pollId}/candidates [post]
func CreateCandidate(c *gin.Context) {
	pollID := c.Param("id")

	if !CheckPollOwnership(c, pollID) {
		return
	}

	var data models.CandidateCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	candidate, err := services.CreateCandidate(database.DB, pollID, data)
	if err != nil {
		if err.Error() == "poll not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, candidate)
}

// ListCandidates handles GET /api/v1/polls/:pollId/candidates
// @Summary      List candidates for a poll
// @Description  Returns a paginated list of candidates for the given poll.
// @Tags         Candidates
// @Produce      json
// @Security     BearerAuth
// @Param        pollId  path   string  true   "Poll UUID"
// @Param        page    query  int     false  "Page number (min: 1)"    default(1)
// @Param        size    query  int     false  "Items per page (1-100)"  default(20)
// @Success      200  {object}  models.PaginatedCandidates
// @Failure      500  {object}  models.ErrorResponse
// @Router       /polls/{pollId}/candidates [get]
func ListCandidates(c *gin.Context) {
	pollID := c.Param("id")
	page, size := handlers.ParsePagination(c)

	result, err := services.GetCandidatesByPoll(database.DB, pollID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetCandidate handles GET /api/v1/candidates/:id
// @Summary      Get candidate by ID
// @Description  Returns a single candidate.
// @Tags         Candidates
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Candidate UUID"
// @Success      200  {object}  models.Candidate
// @Failure      404  {object}  models.ErrorResponse
// @Router       /candidates/{id} [get]
func GetCandidate(c *gin.Context) {
	id := c.Param("id")

	candidate, err := services.GetCandidateByID(database.DB, id)
	if err != nil {
		if err.Error() == "candidate not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, candidate)
}

// UpdateCandidate handles PATCH /api/v1/candidates/:id
// @Summary      Update a candidate
// @Description  Partially updates a candidate. Requires ownership of the parent poll for pollers.
// @Tags         Candidates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                   true  "Candidate UUID"
// @Param        body  body  models.CandidateUpdate    true  "Fields to update"
// @Success      200  {object}  models.Candidate
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /candidates/{id} [patch]
func UpdateCandidate(c *gin.Context) {
	id := c.Param("id")

	// Ownership check: candidate → poll → created_by
	candidate, err := services.GetCandidateByID(database.DB, id)
	if err != nil {
		if err.Error() == "candidate not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !CheckPollOwnership(c, candidate.PollID) {
		return
	}

	var data models.CandidateUpdate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := services.UpdateCandidate(database.DB, id, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// DeleteCandidate handles DELETE /api/v1/candidates/:id
// @Summary      Delete a candidate
// @Description  Deletes a candidate and its answers. Requires ownership of the parent poll for pollers.
// @Tags         Candidates
// @Security     BearerAuth
// @Param        id  path  string  true  "Candidate UUID"
// @Success      204  "No Content"
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /candidates/{id} [delete]
func DeleteCandidate(c *gin.Context) {
	id := c.Param("id")

	// Ownership check
	candidate, err := services.GetCandidateByID(database.DB, id)
	if err != nil {
		if err.Error() == "candidate not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !CheckPollOwnership(c, candidate.PollID) {
		return
	}

	if err := services.DeleteCandidate(database.DB, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// QueryCandidates handles GET /api/v1/candidates?poll_id=xxx
// @Summary      Query candidates
// @Description  Returns a paginated list of candidates, filtered by poll_id.
// @Tags         Candidates
// @Produce      json
// @Security     BearerAuth
// @Param        poll_id  query  string  true   "Poll UUID to filter by"
// @Param        page     query  int     false  "Page number (min: 1)"    default(1)
// @Param        size     query  int     false  "Items per page (1-100)"  default(20)
// @Success      200  {object}  models.PaginatedCandidates
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /candidates [get]
func QueryCandidates(c *gin.Context) {
	pollID := c.Query("poll_id")
	if pollID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "poll_id query parameter is required"})
		return
	}

	page, size := handlers.ParsePagination(c)

	result, err := services.ListCandidates(database.DB, pollID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// BulkCreateCandidates handles POST /api/v1/polls/:id/candidates/bulk
// @Summary      Bulk add candidates to a poll
// @Description  Creates candidates from a list of phone numbers. Deduplicates and normalizes phones.
// @Tags         Candidates
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        pollId  path  string                      true  "Poll UUID"
// @Param        body    body  models.BulkCandidateCreate   true  "Phone numbers to add"
// @Success      201  {object}  models.BulkCandidateResult
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /polls/{pollId}/candidates/bulk [post]
func BulkCreateCandidates(c *gin.Context) {
	pollID := c.Param("id")

	if !CheckPollOwnership(c, pollID) {
		return
	}

	var data models.BulkCandidateCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := services.BulkCreateCandidates(database.DB, pollID, data.PhoneNumbers)
	if err != nil {
		if err.Error() == "poll not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

// UploadCandidateCSV handles POST /api/v1/polls/:id/candidates/upload
// @Summary      Upload CSV to bulk add candidates
// @Description  Parses a CSV file with a 'phone' column and creates candidates for the poll. Deduplicates and normalizes phones.
// @Tags         Candidates
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        pollId  path      string  true  "Poll UUID"
// @Param        file    formData  file    true  "CSV file with phone column (.csv)"
// @Success      201  {object}  models.BulkCandidateResult
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /polls/{pollId}/candidates/upload [post]
func UploadCandidateCSV(c *gin.Context) {
	pollID := c.Param("id")

	if !CheckPollOwnership(c, pollID) {
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	result, err := services.ProcessCandidateCSVUpload(database.DB, fileHeader, pollID)
	if err != nil {
		if err.Error() == "poll not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}
