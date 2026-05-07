package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os" // Tambahin import os
	"strings"
	"testing" // Tambahin import os

	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
)

// Mock TodoRepository
type MockTodoRepository struct {
	// Kita simpan "hasil" yang mau di-return mock
	// Bisa diatur per test
	CreateTodoError error
	GetTodosError   error
	HapusTodoError  error
	UpdateTodoError error
}

func (m *MockTodoRepository) CreateTodo(todo Todo) error {
	return m.CreateTodoError
}

func (m *MockTodoRepository) GetTodos(userID uint, limit int, offset int) ([]Todo, int64, error) {
	return nil, 0, m.GetTodosError
}

func (m *MockTodoRepository) DeleteTodo(id int, userID uint) error {
	return m.HapusTodoError
}

func (m *MockTodoRepository) UpdateTodo(id int, userID uint, judul string, prioritas string) error {
	return m.UpdateTodoError
}
func TestMain(m *testing.M) {
	// Setup validator yang sama persis kayak di main()
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Jalankan semua test
	os.Exit(m.Run())
}
func TestHandlerTodoSingle_Success(t *testing.T) {
	// 1. Setup mock — pura-pura database berhasil
	mock := &MockTodoRepository{
		CreateTodoError: nil,
	}
	repo = mock // inject mock ke handler

	// 2. Bikin fake request
	body := `{"judul":"Belajar Go","prioritas":"tinggi"}`
	req := httptest.NewRequest("POST", "/tambah-todo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	// 4. Jalanin handler
	w := httptest.NewRecorder()
	handlerTodoSingle(w, req)

	// 5. Cek hasilnya
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerTodoSingle_DatabaseError(t *testing.T) {
	mock := &MockTodoRepository{
		CreateTodoError: errors.New("database error"),
	}
	repo = mock // inject mock ke handler

	// 2. Bikin fake request
	body := `{"judul":"Belajar Go","prioritas":"tinggi"}`
	req := httptest.NewRequest("POST", "/tambah-todo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	// 4. Jalanin handler
	w := httptest.NewRecorder()
	handlerTodoSingle(w, req)

	// 5. Cek hasilnya
	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestHandlerTodoSingle_WrongMethod(t *testing.T) {
	mock := &MockTodoRepository{
		CreateTodoError: nil,
	}
	repo = mock // inject mock ke handler

	// 2. Bikin fake request
	body := `{"judul":"Belajar Go","prioritas":"tinggi"}`
	req := httptest.NewRequest("GET", "/tambah-todo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	// 4. Jalanin handler
	w := httptest.NewRecorder()
	handlerTodoSingle(w, req)

	// 5. Cek hasilnya
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlerTodoSingle_JudulKosong(t *testing.T) {
	mock := &MockTodoRepository{
		CreateTodoError: nil,
	}
	repo = mock // inject mock ke handler

	// 2. Bikin fake request
	body := `{"judul":"","prioritas":"tinggi"}`
	req := httptest.NewRequest("POST", "/tambah-todo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	// 4. Jalanin handler
	w := httptest.NewRecorder()
	handlerTodoSingle(w, req)

	// 5. Cek hasilnya
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerTodoSingle_JSONCacat(t *testing.T) {
	mock := &MockTodoRepository{}
	repo = mock

	// Sengaja kirim string yang kurung kurawalnya tidak tutup / bukan format JSON
	body := `{"judul":"Belajar Go", prioritas... rusak`
	req := httptest.NewRequest("POST", "/tambah-todo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerTodoSingle(w, req)

	// Harus return 400 karena format JSON salah
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerTodoBatch_Success(t *testing.T) {
	mock := &MockTodoRepository{
		CreateTodoError: nil,
	}
	repo = mock

	body := `[
		{"judul":"Belajar Go","prioritas":"tinggi"},
		{"judul":"Push up","prioritas":"sedang"}
	]`
	req := httptest.NewRequest("POST", "/tambah-todo-batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerTodoBatch(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerTodoBatch_MixedValid(t *testing.T) {
	mock := &MockTodoRepository{
		CreateTodoError: nil,
	}
	repo = mock

	// 1 valid, 1 judul kosong, 1 prioritas invalid
	body := `[
		{"judul":"Belajar Go","prioritas":"tinggi"},
		{"judul":"","prioritas":"sedang"},
		{"judul":"Makan","prioritas":"xyz"}
	]`
	req := httptest.NewRequest("POST", "/tambah-todo-batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerTodoBatch(w, req)

	// Status tetep 200 karena handler proses semua, tiap item dapet status sendiri
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerTodoBatch_DatabaseError(t *testing.T) {
	mock := &MockTodoRepository{
		CreateTodoError: errors.New("database error"),
	}
	repo = mock

	body := `[{"judul":"Belajar Go","prioritas":"tinggi"}]`
	req := httptest.NewRequest("POST", "/tambah-todo-batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerTodoBatch(w, req)

	// Status tetep 200 — error per item dilaporin di body, bukan status code
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerTodoBatch_WrongMethod(t *testing.T) {
	mock := &MockTodoRepository{}
	repo = mock

	req := httptest.NewRequest("GET", "/tambah-todo-batch", nil)
	req.Header.Set("Content-Type", "application/json")

	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerTodoBatch(w, req)

	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlerTodoBatch_BodyKosong(t *testing.T) {
	mock := &MockTodoRepository{}
	repo = mock

	// Array kosong
	body := `[]`
	req := httptest.NewRequest("POST", "/tambah-todo-batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerTodoBatch(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerTodoBatch_JSONCacat(t *testing.T) {
	mock := &MockTodoRepository{}
	repo = mock

	body := `[{"judul":"Belajar", prioritas... rusak`
	req := httptest.NewRequest("POST", "/tambah-todo-batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerTodoBatch(w, req)

	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerTodos_WrongMethod(t *testing.T) {
	mock := &MockTodoRepository{
		CreateTodoError: nil,
	}
	repo = mock
	req := httptest.NewRequest("POST", "/todos", nil)
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerTodos(w, req)

	// 5. Cek hasilnya
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlerTodos_DatabaseError(t *testing.T) {
	mock := &MockTodoRepository{
		GetTodosError: errors.New("database error"),
	}
	repo = mock
	req := httptest.NewRequest("GET", "/todos", nil)
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerTodos(w, req)

	// 5. Cek hasilnya
	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestHandlerTodos_Success(t *testing.T) {
	mock := &MockTodoRepository{
		GetTodosError: nil,
	}
	repo = mock
	req := httptest.NewRequest("GET", "/todos", nil)
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerTodos(w, req)

	// 5. Cek hasilnya
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerTodos_InvalidPagination(t *testing.T) {
	mock := &MockTodoRepository{
		GetTodosError: nil,
	}
	repo = mock
	req := httptest.NewRequest("GET", "/todos?page=abc&limit=-12", nil)
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerTodos(w, req)

	// 5. Cek hasilnya
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerHapusTodo_Success(t *testing.T) {
	mock := &MockTodoRepository{
		HapusTodoError: nil,
	}
	repo = mock
	req := httptest.NewRequest("DELETE", "/hapus-todo?id=1", nil)
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerHapusTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerHapusTodo_ID0(t *testing.T) {
	mock := &MockTodoRepository{
		HapusTodoError: nil,
	}
	repo = mock
	req := httptest.NewRequest("DELETE", "/hapus-todo?id=0", nil)
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerHapusTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerHapusTodo_WrongMethod(t *testing.T) {
	mock := &MockTodoRepository{
		HapusTodoError: nil,
	}
	repo = mock
	req := httptest.NewRequest("GET", "/hapus-todo?id=1", nil)
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerHapusTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlerHapusTodo_DatabaseError(t *testing.T) {
	mock := &MockTodoRepository{
		HapusTodoError: errors.New("database error"),
	}
	repo = mock
	req := httptest.NewRequest("DELETE", "/hapus-todo?id=1", nil)
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerHapusTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestHandlerHapusTodo_InvalidID(t *testing.T) {
	mock := &MockTodoRepository{
		HapusTodoError: nil,
	}
	repo = mock
	req := httptest.NewRequest("DELETE", "/hapus-todo?id=abc", nil)
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerHapusTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerHapusTodo_IDKosong(t *testing.T) {
	mock := &MockTodoRepository{
		HapusTodoError: nil,
	}
	repo = mock
	req := httptest.NewRequest("DELETE", "/hapus-todo?id=", nil)
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerHapusTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_Success(t *testing.T) {
	mock := &MockTodoRepository{
		UpdateTodoError: nil,
	}
	repo = mock

	body := `{"judul":"Belajar Go Part 2","prioritas":"rendah"}`
	req := httptest.NewRequest("PUT", "/update-todo?id=1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerUpdateTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_ID0(t *testing.T) {
	mock := &MockTodoRepository{
		UpdateTodoError: nil,
	}
	repo = mock

	body := `{"judul":"Belajar Go Part 2","prioritas":"rendah"}`
	req := httptest.NewRequest("PUT", "/update-todo?id=0", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerUpdateTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_InvalidID(t *testing.T) {
	mock := &MockTodoRepository{
		UpdateTodoError: nil,
	}
	repo = mock

	body := `{"judul":"Belajar Go Part 2","prioritas":"rendah"}`
	req := httptest.NewRequest("PUT", "/update-todo?id=abc", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerUpdateTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_WrongMethod(t *testing.T) {
	mock := &MockTodoRepository{
		UpdateTodoError: nil,
	}
	repo = mock

	body := `{"judul":"Belajar Go Part 2","prioritas":"rendah"}`
	req := httptest.NewRequest("DELETE", "/update-todo?id=1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerUpdateTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_JudulKosong(t *testing.T) {
	mock := &MockTodoRepository{
		UpdateTodoError: nil,
	}
	repo = mock

	body := `{"judul":"","prioritas":"rendah"}`
	req := httptest.NewRequest("PUT", "/update-todo?id=1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerUpdateTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_PrioritasKosong(t *testing.T) {
	mock := &MockTodoRepository{
		UpdateTodoError: nil,
	}
	repo = mock

	body := `{"judul":"Belajar Go Part 2","prioritas":""}`
	req := httptest.NewRequest("PUT", "/update-todo?id=1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerUpdateTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_DatabaseError(t *testing.T) {
	mock := &MockTodoRepository{
		UpdateTodoError: errors.New("database error"),
	}
	repo = mock

	body := `{"judul":"Belajar Go Part 2","prioritas":"rendah"}`
	req := httptest.NewRequest("PUT", "/update-todo?id=1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 3. Inject fake JWT claims ke context
	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlerUpdateTodo(w, req)

	// 5. Cek hasilnya
	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
