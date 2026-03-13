package repository

import (
	"backend/internal/models"

	"github.com/jmoiron/sqlx"
)

type TodoRepository struct {
	db *sqlx.DB
}

func NewTodoRepository(db *sqlx.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) Create(todo *models.Todo) error {

	query := `
	INSERT INTO todos (id, title, completed, created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5)`

	_, err := r.db.Exec(query,
		todo.ID,
		todo.Title,
		todo.Completed,
		todo.CreatedAt,
		todo.UpdatedAt,
	)

	return err
}

func (r *TodoRepository) GetByID(id string) (*models.Todo, error) {

	var todo models.Todo

	err := r.db.Get(&todo,
		"SELECT * FROM todos WHERE id=$1",
		id,
	)

	return &todo, err
}

func (r *TodoRepository) GetAll(limit int, offset int) ([]models.Todo, error) {

	var todos []models.Todo

	err := r.db.Select(&todos,
		"SELECT * FROM todos ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		limit,
		offset,
	)

	return todos, err
}

func (r *TodoRepository) Update(todo *models.Todo) error {

	query := `
	UPDATE todos
	SET title=$1, completed=$2, updated_at=$3
	WHERE id=$4
	`

	_, err := r.db.Exec(
		query,
		todo.Title,
		todo.Completed,
		todo.UpdatedAt,
		todo.ID,
	)

	return err
}
