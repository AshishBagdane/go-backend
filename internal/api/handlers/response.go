package handlers

import (
	"github.com/AshishBagdane/go-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type APIResponse[T any] struct {
	Response T      `json:"response"`
	Status   int    `json:"status"`
	Message  string `json:"message"`
}

// Swagger-friendly concrete types (swag does not support generics).
type APIResponseTodo struct {
	Response models.Todo `json:"response"`
	Status   int         `json:"status"`
	Message  string      `json:"message"`
}

type APIResponseTodos struct {
	Response []models.Todo `json:"response"`
	Status   int           `json:"status"`
	Message  string        `json:"message"`
}

type APIResponseEmpty struct {
	Response any    `json:"response"`
	Status   int    `json:"status"`
	Message  string `json:"message"`
}

func Respond[T any](c *gin.Context, status int, message string, data T) {
	c.JSON(status, APIResponse[T]{
		Response: data,
		Status:   status,
		Message:  message,
	})
}
