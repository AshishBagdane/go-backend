package service

import (
	"encoding/json"
	"time"

	"backend/internal/cache"
	"backend/internal/models"
	"backend/internal/repository"

	"github.com/google/uuid"
)

type TodoService struct {
	repo       *repository.TodoRepository
	localCache *cache.LocalCache
	redisCache *cache.RedisCache
}

func NewTodoService(
	repo *repository.TodoRepository,
	local *cache.LocalCache,
	redis *cache.RedisCache,
) *TodoService {

	return &TodoService{
		repo:       repo,
		localCache: local,
		redisCache: redis,
	}
}

func (s *TodoService) Create(title string) (*models.Todo, error) {

	todo := &models.Todo{
		ID:        uuid.New().String(),
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.repo.Create(todo)

	if err == nil {
		bytes, _ := json.Marshal(todo)
		s.redisCache.Set(todo.ID, string(bytes))
		s.localCache.Set(todo.ID, todo)

		// Invalidate list caches to avoid stale empty results.
		s.redisCache.Delete("todos")
		s.localCache.Delete("todos")
	}

	return todo, err
}

func (s *TodoService) GetTodo(id string) (*models.Todo, error) {

	if val, ok := s.localCache.Get(id); ok {
		return val.(*models.Todo), nil
	}

	val, err := s.redisCache.Get(id)

	if err == nil {

		var todo models.Todo
		json.Unmarshal([]byte(val), &todo)

		s.localCache.Set(id, &todo)

		return &todo, nil
	}

	todo, err := s.repo.GetByID(id)

	if err == nil {
		bytes, _ := json.Marshal(todo)
		s.redisCache.Set(id, string(bytes))
		s.localCache.Set(id, todo)
	}

	return todo, err
}

func (s *TodoService) GetTodos(limit int, offset int) ([]models.Todo, error) {
	todos, err := s.repo.GetAll(limit, offset)
	if todos == nil {
		todos = []models.Todo{}
	}
	return todos, err
}

func (s *TodoService) UpdateTodo(id string, title string, completed bool) (*models.Todo, error) {

	// Fetch existing todo from DB
	todo, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	todo.Title = title
	todo.Completed = completed
	todo.UpdatedAt = time.Now()

	// Persist update in DB
	err = s.repo.Update(todo)
	if err != nil {
		return nil, err
	}

	// Serialize updated object
	bytes, err := json.Marshal(todo)
	if err == nil {
		s.redisCache.Set(id, string(bytes))
	}

	// Update local cache
	s.localCache.Set(id, todo)
	// Invalidate list caches to avoid stale results.
	s.redisCache.Delete("todos")
	s.localCache.Delete("todos")

	return todo, nil
}
