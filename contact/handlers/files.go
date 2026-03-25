package handlers

import (
	"net/http"
	"strconv"

	"polling-system/config"
	_ "polling-system/contact/models" // swagger model references
	"polling-system/contact/services"
	"polling-system/database"
	"polling-system/middleware"

	"github.com/gin-gonic/gin"
)

// ListFiles handles GET /api/v1/files
// @Summary      List uploaded files
// @Description  Returns a paginated list of uploaded CSV files.
// @Tags         Files
// @Produce      json
// @Security     BearerAuth
// @Param        page  query  int  false  "Page number (min: 1)"    default(1)
// @Param        size  query  int  false  "Items per page (1-100)"  default(20)
// @Success      200  {object}  models.PaginatedFiles
// @Failure      500  {object}  models.ErrorResponse
// @Router       /files [get]
func ListFiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	result, err := services.GetFiles(database.DB, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// UploadCSV handles POST /api/v1/files/upload
// @Summary      Upload a CSV file of contacts
// @Description  Parses the CSV, saves the file to disk, and creates contact records in the database. The uploading user is determined from the auth token.
// @Tags         Files
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "CSV file to upload (.csv)"
// @Success      201  {object}  models.File
// @Failure      400  {object}  models.ErrorResponse
// @Router       /files/upload [post]
func UploadCSV(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}

		userID := middleware.GetUserID(c)

		uploaded, err := services.ProcessCSVUpload(database.DB, cfg, fileHeader, userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, uploaded)
	}
}
