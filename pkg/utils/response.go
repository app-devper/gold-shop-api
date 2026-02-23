package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	ErrorCode string      `json:"errorCode,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    PageMeta    `json:"meta"`
}

// PageMeta contains pagination metadata
type PageMeta struct {
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	PerPage int   `json:"per_page"`
	Pages   int64 `json:"pages"`
}

// SuccessResponse sends a success response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// CreatedResponse sends a created response
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// MessageResponse sends a message response
func MessageResponse(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, errorCode string, status int, message string) {
	c.JSON(status, Response{
		Success:   false,
		Error:     message,
		ErrorCode: errorCode,
	})
}

// BadRequestResponse sends a bad request response
func BadRequestResponse(c *gin.Context, errorCode string, message string) {
	ErrorResponse(c, errorCode, http.StatusBadRequest, message)
}

// UnauthorizedResponse sends an unauthorized response
func UnauthorizedResponse(c *gin.Context, errorCode string, message string) {
	ErrorResponse(c, errorCode, http.StatusUnauthorized, message)
}

// ForbiddenResponse sends a forbidden response
func ForbiddenResponse(c *gin.Context, errorCode string, message string) {
	ErrorResponse(c, errorCode, http.StatusForbidden, message)
}

// NotFoundResponse sends a not found response
func NotFoundResponse(c *gin.Context, errorCode string, message string) {
	ErrorResponse(c, errorCode, http.StatusNotFound, message)
}

// InternalErrorResponse sends an internal server error response
func InternalErrorResponse(c *gin.Context, errorCode string, message string) {
	ErrorResponse(c, errorCode, http.StatusInternalServerError, message)
}

// PaginatedSuccessResponse sends a paginated success response
func PaginatedSuccessResponse(c *gin.Context, data interface{}, total int64, page, perPage int) {
	pages := total / int64(perPage)
	if total%int64(perPage) > 0 {
		pages++
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    data,
		Meta: PageMeta{
			Total:   total,
			Page:    page,
			PerPage: perPage,
			Pages:   pages,
		},
	})
}
