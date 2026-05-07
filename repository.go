package main

import "gorm.io/gorm"

// ============================================
// 1. Interface (kontrak)
// ============================================
type TodoRepository interface {
	CreateTodo(todo Todo) error
	GetTodos(userID uint, limit int, offset int) ([]Todo, int64, error)
	DeleteTodo(id int, userID uint) error
	UpdateTodo(id int, userID uint, judul string, prioritas string) error
}

// ============================================
// 2. Implementasi GORM (yang memenuhi kontrak)
// ============================================
type PostgresTodoRepository struct {
	db *gorm.DB
}

func NewPostgresTodoRepository(db *gorm.DB) TodoRepository {
	return &PostgresTodoRepository{db: db}
}

func (r *PostgresTodoRepository) CreateTodo(todo Todo) error {
	result := r.db.Create(&todo)
	return result.Error
}

func (r *PostgresTodoRepository) GetTodos(userID uint, limit int, offset int) ([]Todo, int64, error) {
	var todos []Todo
	var total int64

	r.db.Model(&Todo{}).Where("user_id = ?", userID).Count(&total)
	result := r.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&todos)

	return todos, total, result.Error
}

func (r *PostgresTodoRepository) DeleteTodo(id int, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Todo{})
	return result.Error
}

func (r *PostgresTodoRepository) UpdateTodo(id int, userID uint, judul string, prioritas string) error {
	result := r.db.Model(&Todo{}).Where("id = ? AND user_id = ?", id, userID).Updates(map[string]any{
		"judul":     judul,
		"prioritas": prioritas,
	})
	return result.Error
}
