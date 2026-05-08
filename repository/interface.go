package repository

import "todo-api/models"

type TodoRepository interface {
	CreateTodo(todo models.Todo) error
	GetTodos(userID uint, limit int, offset int) ([]models.Todo, int64, error)
	DeleteTodo(id int, userID uint) error
	UpdateTodo(id int, userID uint, judul string, prioritas string) error
}

type UserRepository interface {
	CheckUser(username string) (models.User, error)
	RegisterUser(user models.User) error
}
