package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/AshishBagdane/go-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	service *service.TodoService
}

func NewTodoHandler(s *service.TodoService) *TodoHandler {
	return &TodoHandler{service: s}
}

// CreateTodo godoc
// @Summary Create a new todo
// @Description create a todo item
// @Tags todos
// @Accept json
// @Produce json
// @Param todo body object true "Todo title"
// @Success 200 {object} APIResponseTodo
// @Router /todos [post]
func (h *TodoHandler) CreateTodo(c *gin.Context) {

	var req struct {
		Title string `json:"title"`
	}

	if err := c.BindJSON(&req); err != nil {
		Respond[any](c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		Respond[any](c, http.StatusBadRequest, "title is required", nil)
		return
	}

	todo, err := h.service.Create(req.Title)

	if err != nil {
		Respond[any](c, http.StatusInternalServerError, "failed to create todo", nil)
		return
	}

	Respond(c, http.StatusOK, "todo created", todo)
}

// GetTodo godoc
// @Summary Get todo by ID
// @Description fetch todo by ID
// @Tags todos
// @Produce json
// @Param id path string true "Todo ID"
// @Success 200 {object} APIResponseTodo
// @Router /todos/{id} [get]
func (h *TodoHandler) GetTodo(c *gin.Context) {

	id := c.Param("id")

	todo, err := h.service.GetTodo(id)

	if err != nil {
		if err == sql.ErrNoRows {
			Respond[any](c, http.StatusNotFound, "not found", nil)
		} else {
			Respond[any](c, http.StatusInternalServerError, "failed to fetch todo", nil)
		}
		return
	}

	Respond(c, http.StatusOK, "ok", todo)
}

// GetTodos godoc
// @Summary Get todos
// @Description fetch todos
// @Tags todos
// @Produce json
// @Param limit query int false "Limit (max 100)"
// @Param offset query int false "Offset"
// @Success 200 {object} APIResponseTodos
// @Router /todos [get]
func (h *TodoHandler) GetTodos(c *gin.Context) {

	limit := 20
	offset := 0

	if v := c.Query("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		} else {
			Respond[any](c, http.StatusBadRequest, "invalid limit", nil)
			return
		}
	}
	if v := c.Query("offset"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed >= 0 {
			offset = parsed
		} else {
			Respond[any](c, http.StatusBadRequest, "invalid offset", nil)
			return
		}
	}

	if limit > 100 {
		limit = 100
	}

	todos, err := h.service.GetTodos(limit, offset)

	if err != nil {
		Respond[any](c, http.StatusInternalServerError, "failed to fetch todos", nil)
		return
	}

	Respond(c, http.StatusOK, "ok", todos)
}

// UpdateTodo godoc
// @Summary Update todo by ID
// @Description update todo by ID
// @Tags todos
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Param todo body object true "Todo title and completed status"
// @Success 200 {object} APIResponseTodo
// @Router /todos/{id} [put]
func (h *TodoHandler) UpdateTodo(c *gin.Context) {

	id := c.Param("id")

	var req struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}

	if err := c.BindJSON(&req); err != nil {
		Respond[any](c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		Respond[any](c, http.StatusBadRequest, "title is required", nil)
		return
	}

	todo, err := h.service.UpdateTodo(id, req.Title, req.Completed)

	if err != nil {
		if err == sql.ErrNoRows {
			Respond[any](c, http.StatusNotFound, "not found", nil)
		} else {
			Respond[any](c, http.StatusInternalServerError, "failed to update todo", nil)
		}
		return
	}

	Respond(c, http.StatusOK, "todo updated", todo)
}
