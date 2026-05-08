package models

import "gorm.io/gorm"

type Todo struct {
	gorm.Model
	UserID    uint `json:"user_id"`
	Judul     string
	Prioritas string
}

type User struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
}

type ListTodo struct {
	Judul     string `json:"judul" validate:"required"`
	Prioritas string `json:"prioritas" validate:"required,oneof=tinggi sedang rendah"`
}

type ListTodoBatch struct {
	Judul     string `json:"judul"`
	Prioritas string `json:"prioritas,omitempty"`
	Status    string `json:"status,omitempty"`
	Error     string `json:"error,omitempty"`
}

type GetTodo struct {
	ID        int    `json:"id"`
	Judul     string `json:"judul"`
	Prioritas string `json:"prioritas"`
}

type InputAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponError struct {
	Error string `json:"error"`
}

type ResponPesan struct {
	Pesan string `json:"pesan" example:"Berhasil menambahkan username ke database"`
}

type SuccessResponLogin struct {
	Pesan string `json:"pesan" example:"Berhasil login!"`
	Token string `json:"token"`
}

type SuccessHandlerTodoSingle struct {
	Pesan     string `json:"pesan"`
	Judul     string `json:"judul"`
	Prioritas string `json:"prioritas"`
}

type SuccessHandlerHapusUpdateTodo struct {
	Pesan string `json:"pesan"`
}

type PageData struct {
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
	Total int64     `json:"total"`
	Data  []GetTodo `json:"data"`
}
