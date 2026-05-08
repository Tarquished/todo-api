package repository

import (
	"todo-api/models"

	"gorm.io/gorm"
)

// ============ Todo ============

type PostgresTodoRepository struct {
	DB *gorm.DB
}

func NewPostgresTodoRepository(db *gorm.DB) TodoRepository {
	return &PostgresTodoRepository{DB: db}
}

func (r *PostgresTodoRepository) CreateTodo(todo models.Todo) error {
	result := r.DB.Create(&todo)
	return result.Error
}

func (r *PostgresTodoRepository) GetTodos(userID uint, limit int, offset int) ([]models.Todo, int64, error) {
	var todos []models.Todo
	var total int64

	r.DB.Model(&models.Todo{}).Where("user_id = ?", userID).Count(&total)
	result := r.DB.Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&todos)

	return todos, total, result.Error
}

func (r *PostgresTodoRepository) DeleteTodo(id int, userID uint) error {
	result := r.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Todo{})
	return result.Error
}

func (r *PostgresTodoRepository) UpdateTodo(id int, userID uint, judul string, prioritas string) error {
	result := r.DB.Model(&models.Todo{}).Where("id = ? AND user_id = ?", id, userID).Updates(map[string]any{
		"judul":     judul,
		"prioritas": prioritas,
	})
	return result.Error
}

// ============ User ============

type PostgresUserRepository struct {
	DB *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepository{DB: db}
}

func (r *PostgresUserRepository) CheckUser(username string) (models.User, error) {
	var user models.User
	result := r.DB.Where("username = ?", username).First(&user)
	return user, result.Error
}

func (r *PostgresUserRepository) RegisterUser(user models.User) error {
	result := r.DB.Create(&user)
	return result.Error
}
