package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"polling-system/contact/models"
	"polling-system/contact/services"
	"polling-system/database"

	"github.com/gin-gonic/gin"
)

// CreateContact handles POST /api/v1/contacts
// @Summary      Create a contact
// @Description  Creates a new contact. Phone normalization is calculated automatically.
// @Tags         Contacts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  models.ContactCreate  true  "Contact to create"
// @Success      201  {object}  models.Contact
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /contacts [post]
func CreateContact(c *gin.Context) {
	var data models.ContactCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contact, err := services.CreateContact(database.DB, data)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, contact)
}

// ListContacts handles GET /api/v1/contacts
// @Summary      List contacts
// @Description  Returns a paginated list of contacts, optionally filtered by file_id, poll_id, name, or phone number.
// @Tags         Contacts
// @Produce      json
// @Security     BearerAuth
// @Param        page     query  int     false  "Page number (min: 1)"          default(1)
// @Param        size     query  int     false  "Items per page (1-100)"        default(20)
// @Param        file_id  query  string  false  "Filter by file UUID"
// @Param        poll_id  query  string  false  "Filter by poll UUID"
// @Param        name     query  string  false  "Filter by name (partial match)"
// @Param        phone    query  string  false  "Filter by phone (partial match)"
// @Success      200  {object}  models.PaginatedContacts
// @Failure      500  {object}  models.ErrorResponse
// @Router       /contacts [get]
func ListContacts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	fileID := c.Query("file_id")
	pollID := c.Query("poll_id")
	name := c.Query("name")
	phone := c.Query("phone")

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	result, err := services.GetContacts(database.DB, page, size, name, phone, fileID, pollID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// UpdateContact handles PATCH /api/v1/contacts/:id
// @Summary      Update a contact
// @Description  Updates the name and/or phone number of an existing contact. Phone normalization is recalculated automatically.
// @Tags         Contacts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string               true  "Contact UUID"
// @Param        body  body  models.ContactUpdate  true  "Fields to update"
// @Success      200  {object}  models.Contact
// @Failure      400  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /contacts/{id} [patch]
func UpdateContact(c *gin.Context) {
	contactID := c.Param("id")

	var data models.ContactUpdate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contact, err := services.UpdateContact(database.DB, contactID, data)
	if err != nil {
		if err.Error() == "contact not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contact)
}

// DeleteContact handles DELETE /api/v1/contacts/:id
// @Summary      Delete a contact
// @Description  Permanently removes a contact from the database.
// @Tags         Contacts
// @Security     BearerAuth
// @Param        id  path  string  true  "Contact UUID"
// @Success      204  "No Content"
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /contacts/{id} [delete]
func DeleteContact(c *gin.Context) {
	contactID := c.Param("id")

	err := services.DeleteContact(database.DB, contactID)
	if err != nil {
		if err.Error() == "contact not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// BulkCreateContacts handles POST /api/v1/contacts/bulk
// @Summary      Bulk create contacts
// @Description  Creates multiple contacts from a list of phone numbers for a specific poll. Duplicates and invalid numbers are skipped and reported.
// @Tags         Contacts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  models.BulkContactCreate  true  "Bulk contact creation request"
// @Success      201  {object}  models.BulkContactResult
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /contacts/bulk [post]
func BulkCreateContacts(c *gin.Context) {
	var data models.BulkContactCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := services.BulkCreateContacts(database.DB, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}
