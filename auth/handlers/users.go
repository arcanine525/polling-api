package handlers

import (
	"net/http"
	"strings"

	"polling-system/auth/models"
	"polling-system/auth/services"
	"polling-system/database"
	"polling-system/handlers"
	"polling-system/middleware"

	"github.com/gin-gonic/gin"
)

// GetMe handles GET /api/v1/users/me
// @Summary      Get current user profile
// @Description  Returns the authenticated user's profile.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  models.User
// @Failure      401  {object}  models.ErrorResponse
// @Router       /users/me [get]
func GetMe(c *gin.Context) {
	user := middleware.GetUser(c)
	c.JSON(http.StatusOK, user)
}

// UpdateMe handles PATCH /api/v1/users/me
// @Summary      Update own profile
// @Description  Updates the authenticated user's display name.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  models.UserUpdate  true  "Fields to update"
// @Success      200  {object}  models.User
// @Failure      400  {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Router       /users/me [patch]
func UpdateMe(c *gin.Context) {
	var data models.UserUpdate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	user, err := services.UpdateUserDisplayName(database.DB, userID, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// ListUsers handles GET /api/v1/users
// @Summary      List all users
// @Description  Returns a paginated list of all users. Admin only.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        page  query  int  false  "Page number (min: 1)"    default(1)
// @Param        size  query  int  false  "Items per page (1-100)"  default(20)
// @Success      200  {object}  models.PaginatedUsers
// @Failure      403  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /users [get]
func ListUsers(c *gin.Context) {
	page, size := handlers.ParsePagination(c)

	result, err := services.GetUsers(database.DB, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetUser handles GET /api/v1/users/:id
// @Summary      Get user by ID
// @Description  Returns a user by database ID. Admin only.
// @Tags         Users
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "User UUID"
// @Success      200  {object}  models.User
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /users/{id} [get]
func GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := services.GetUserByID(database.DB, id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// CreateUser handles POST /api/v1/users
// @Summary      Create user
// @Description  Creates a user manually. Admin only.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  models.UserCreate  true  "User to create"
// @Success      201  {object}  models.User
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      409  {object}  models.ErrorResponse
// @Router       /users [post]
func CreateUser(c *gin.Context) {
	var data models.UserCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := services.CreateUser(database.DB, data)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

// UpdateUser handles PATCH /api/v1/users/:id
// @Summary      Update a user
// @Description  Updates a user. Admin only.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string             true  "User UUID"
// @Param        body  body  models.UserUpdate  true  "Fields to update"
// @Success      200  {object}  models.User
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /users/{id} [patch]
func UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var data models.UserUpdate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := services.UpdateUser(database.DB, id, data)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateUserRole handles PATCH /api/v1/users/:id/role
// @Summary      Update user role
// @Description  Updates a user's role. Admin only.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                 true  "User UUID"
// @Param        body  body  models.UserRoleUpdate  true  "New role"
// @Success      200  {object}  models.User
// @Failure      400  {object}  models.ErrorResponse
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /users/{id}/role [patch]
func UpdateUserRole(c *gin.Context) {
	id := c.Param("id")

	var data models.UserRoleUpdate
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := services.UpdateUserRole(database.DB, id, data)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// DeleteUser handles DELETE /api/v1/users/:id
// @Summary      Delete a user
// @Description  Deletes a user. Admin only.
// @Tags         Users
// @Security     BearerAuth
// @Param        id  path  string  true  "User UUID"
// @Success      204  "No Content"
// @Failure      403  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /users/{id} [delete]
func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := services.DeleteUser(database.DB, id); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
