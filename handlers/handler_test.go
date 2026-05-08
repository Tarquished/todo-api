package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"todo-api/models"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ============================================
// Mock Repositories
// ============================================

type MockTodoRepository struct {
	CreateTodoError error
	GetTodosError   error
	HapusTodoError  error
	UpdateTodoError error
}

func (m *MockTodoRepository) CreateTodo(todo models.Todo) error {
	return m.CreateTodoError
}

func (m *MockTodoRepository) GetTodos(userID uint, limit int, offset int) ([]models.Todo, int64, error) {
	return nil, 0, m.GetTodosError
}

func (m *MockTodoRepository) DeleteTodo(id int, userID uint) error {
	return m.HapusTodoError
}

func (m *MockTodoRepository) UpdateTodo(id int, userID uint, judul string, prioritas string) error {
	return m.UpdateTodoError
}

type MockUserRepository struct {
	CheckUserData  models.User
	CheckUserError error
	RegisterError  error
}

func (m *MockUserRepository) CheckUser(username string) (models.User, error) {
	return m.CheckUserData, m.CheckUserError
}

func (m *MockUserRepository) RegisterUser(user models.User) error {
	return m.RegisterError
}

// ============================================
// Setup
// ============================================

func TestMain(m *testing.M) {
	Validate = validator.New()
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	os.Exit(m.Run())
}

// helper: bikin request dengan JWT claims
func newRequestWithClaims(method, url string, body string) (*http.Request, *httptest.ResponseRecorder) {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, url, strings.NewReader(body))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	req.Header.Set("Content-Type", "application/json")

	claims := jwt.MapClaims{"user_id": float64(1)}
	ctx := context.WithValue(req.Context(), "claims", &claims)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	return req, w
}

// ============================================
// FormatValidationError Tests
// ============================================

func TestFormatValidationError(t *testing.T) {
	tests := []struct {
		nama     string
		input    models.ListTodo
		expected []string
	}{
		{
			nama:     "judul kosong",
			input:    models.ListTodo{Judul: "", Prioritas: "tinggi"},
			expected: []string{"judul harus terisi"},
		},
		{
			nama:     "prioritas kosong",
			input:    models.ListTodo{Judul: "judul", Prioritas: ""},
			expected: []string{"prioritas harus terisi"},
		},
		{
			nama:     "prioritas tidak valid",
			input:    models.ListTodo{Judul: "judul", Prioritas: "ada"},
			expected: []string{"prioritas harus berupa tinggi sedang rendah"},
		},
		{
			nama:     "judul dan prioritas kosong",
			input:    models.ListTodo{Judul: "", Prioritas: ""},
			expected: []string{"judul harus terisi", "prioritas harus terisi"},
		},
		{
			nama:     "valid",
			input:    models.ListTodo{Judul: "judul", Prioritas: "tinggi"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			err := Validate.Struct(tt.input)

			if tt.expected == nil {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Errorf("expected error %v, got nil", tt.expected)
				return
			}

			result := FormatValidationError(err)

			if len(result) != len(tt.expected) {
				t.Errorf("got %d errors, want %d. result: %v", len(result), len(tt.expected), result)
				return
			}

			for i, msg := range tt.expected {
				if result[i] != msg {
					t.Errorf("got %q, want %q", result[i], msg)
				}
			}
		})
	}
}

// ============================================
// HandlerTodoSingle Tests
// ============================================

func TestHandlerTodoSingle_Success(t *testing.T) {
	Repo = &MockTodoRepository{CreateTodoError: nil}
	req, w := newRequestWithClaims("POST", "/tambah-todo", `{"judul":"Belajar Go","prioritas":"tinggi"}`)
	HandlerTodoSingle(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerTodoSingle_DatabaseError(t *testing.T) {
	Repo = &MockTodoRepository{CreateTodoError: errors.New("database error")}
	req, w := newRequestWithClaims("POST", "/tambah-todo", `{"judul":"Belajar Go","prioritas":"tinggi"}`)
	HandlerTodoSingle(w, req)
	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestHandlerTodoSingle_WrongMethod(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("GET", "/tambah-todo", `{"judul":"Belajar Go","prioritas":"tinggi"}`)
	HandlerTodoSingle(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlerTodoSingle_JudulKosong(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("POST", "/tambah-todo", `{"judul":"","prioritas":"tinggi"}`)
	HandlerTodoSingle(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerTodoSingle_JSONCacat(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("POST", "/tambah-todo", `{"judul":"Belajar Go", prioritas... rusak`)
	HandlerTodoSingle(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ============================================
// HandlerTodoBatch Tests
// ============================================

func TestHandlerTodoBatch_Success(t *testing.T) {
	Repo = &MockTodoRepository{CreateTodoError: nil}
	req, w := newRequestWithClaims("POST", "/tambah-todo-batch", `[{"judul":"Belajar Go","prioritas":"tinggi"},{"judul":"Push up","prioritas":"sedang"}]`)
	HandlerTodoBatch(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerTodoBatch_MixedValid(t *testing.T) {
	Repo = &MockTodoRepository{CreateTodoError: nil}
	req, w := newRequestWithClaims("POST", "/tambah-todo-batch", `[{"judul":"Belajar Go","prioritas":"tinggi"},{"judul":"","prioritas":"sedang"},{"judul":"Makan","prioritas":"xyz"}]`)
	HandlerTodoBatch(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerTodoBatch_DatabaseError(t *testing.T) {
	Repo = &MockTodoRepository{CreateTodoError: errors.New("database error")}
	req, w := newRequestWithClaims("POST", "/tambah-todo-batch", `[{"judul":"Belajar Go","prioritas":"tinggi"}]`)
	HandlerTodoBatch(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerTodoBatch_WrongMethod(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("GET", "/tambah-todo-batch", "")
	HandlerTodoBatch(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlerTodoBatch_BodyKosong(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("POST", "/tambah-todo-batch", `[]`)
	HandlerTodoBatch(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerTodoBatch_JSONCacat(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("POST", "/tambah-todo-batch", `[{"judul":"Belajar", prioritas... rusak`)
	HandlerTodoBatch(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ============================================
// HandlerTodos Tests
// ============================================

func TestHandlerTodos_WrongMethod(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("POST", "/todos", "")
	HandlerTodos(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlerTodos_DatabaseError(t *testing.T) {
	Repo = &MockTodoRepository{GetTodosError: errors.New("database error")}
	req, w := newRequestWithClaims("GET", "/todos", "")
	HandlerTodos(w, req)
	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestHandlerTodos_Success(t *testing.T) {
	Repo = &MockTodoRepository{GetTodosError: nil}
	req, w := newRequestWithClaims("GET", "/todos", "")
	HandlerTodos(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerTodos_InvalidPagination(t *testing.T) {
	Repo = &MockTodoRepository{GetTodosError: nil}
	req, w := newRequestWithClaims("GET", "/todos?page=abc&limit=-12", "")
	HandlerTodos(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ============================================
// HandlerHapusTodo Tests
// ============================================

func TestHandlerHapusTodo_Success(t *testing.T) {
	Repo = &MockTodoRepository{HapusTodoError: nil}
	req, w := newRequestWithClaims("DELETE", "/hapus-todo?id=1", "")
	HandlerHapusTodo(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerHapusTodo_ID0(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("DELETE", "/hapus-todo?id=0", "")
	HandlerHapusTodo(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerHapusTodo_WrongMethod(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("GET", "/hapus-todo?id=1", "")
	HandlerHapusTodo(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlerHapusTodo_DatabaseError(t *testing.T) {
	Repo = &MockTodoRepository{HapusTodoError: errors.New("database error")}
	req, w := newRequestWithClaims("DELETE", "/hapus-todo?id=1", "")
	HandlerHapusTodo(w, req)
	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestHandlerHapusTodo_InvalidID(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("DELETE", "/hapus-todo?id=abc", "")
	HandlerHapusTodo(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerHapusTodo_IDKosong(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("DELETE", "/hapus-todo?id=", "")
	HandlerHapusTodo(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ============================================
// HandlerUpdateTodo Tests
// ============================================

func TestHandlerUpdateTodo_Success(t *testing.T) {
	Repo = &MockTodoRepository{UpdateTodoError: nil}
	req, w := newRequestWithClaims("PUT", "/update-todo?id=1", `{"judul":"Belajar Go Part 2","prioritas":"rendah"}`)
	HandlerUpdateTodo(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_ID0(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("PUT", "/update-todo?id=0", `{"judul":"Belajar Go Part 2","prioritas":"rendah"}`)
	HandlerUpdateTodo(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_InvalidID(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("PUT", "/update-todo?id=abc", `{"judul":"Belajar Go Part 2","prioritas":"rendah"}`)
	HandlerUpdateTodo(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_WrongMethod(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("DELETE", "/update-todo?id=1", `{"judul":"Belajar Go Part 2","prioritas":"rendah"}`)
	HandlerUpdateTodo(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_JudulKosong(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("PUT", "/update-todo?id=1", `{"judul":"","prioritas":"rendah"}`)
	HandlerUpdateTodo(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_PrioritasKosong(t *testing.T) {
	Repo = &MockTodoRepository{}
	req, w := newRequestWithClaims("PUT", "/update-todo?id=1", `{"judul":"Belajar Go Part 2","prioritas":""}`)
	HandlerUpdateTodo(w, req)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerUpdateTodo_DatabaseError(t *testing.T) {
	Repo = &MockTodoRepository{UpdateTodoError: errors.New("database error")}
	req, w := newRequestWithClaims("PUT", "/update-todo?id=1", `{"judul":"Belajar Go Part 2","prioritas":"rendah"}`)
	HandlerUpdateTodo(w, req)
	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ============================================
// HandlerRegister Tests
// ============================================

func TestHandlerRegister(t *testing.T) {
	tests := []struct {
		nama           string
		method         string
		body           string
		expectedStatus int
		mockRepo       *MockUserRepository
	}{
		{"body kosong", "POST", "", 400, nil},
		{"method salah", "GET", "", 405, nil},
		{"username kosong", "POST", `{"username":"","password":"rahasia"}`, 400, nil},
		{"password kosong", "POST", `{"username":"jason111","password":""}`, 400, nil},
		{"json cacat", "POST", `{"username":"jason111", passw`, 400, nil},
		{"success", "POST", `{"username":"jason111","password":"rahasiaa"}`, 200, &MockUserRepository{
			CheckUserError: errors.New("user tidak ditemukan"),
			RegisterError:  nil,
		}},
		{"username sudah ada", "POST", `{"username":"jason111","password":"rahasiaa"}`, 400, &MockUserRepository{
			CheckUserError: nil,
		}},
		{"database error", "POST", `{"username":"jason111","password":"rahasiaa"}`, 500, &MockUserRepository{
			CheckUserError: errors.New("user tidak ditemukan"),
			RegisterError:  errors.New("database error"),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			UserRepo = tt.mockRepo

			var body *strings.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			} else {
				body = strings.NewReader("")
			}
			req := httptest.NewRequest(tt.method, "/register", body)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			HandlerRegister(rr, req)
			if rr.Code != tt.expectedStatus {
				t.Errorf("got %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}

// ============================================
// HandlerLogin Tests
// ============================================

func TestHandlerLogin(t *testing.T) {
	passwordAsli := "rahasia123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(passwordAsli), bcrypt.MinCost)
	tests := []struct {
		nama           string
		method         string
		body           string
		expectedStatus int
		mockRepo       *MockUserRepository
	}{
		{"method salah", "GET", "", 405, nil},
		{"body kosong", "POST", "", 400, nil},
		{"username kosong", "POST", `{"username":"", "password":"rahasia"}`, 400, nil},
		{"password kosong", "POST", `{"username":"jason","password":""}`, 400, nil},
		{"json cacat", "POST", `{"username":"jason", pass`, 400, nil},
		{"success", "POST", `{"username":"jason", "password":"rahasia123"}`, 200, &MockUserRepository{
			CheckUserData: models.User{
				Username: "jason",
				Password: string(hashedPassword),
			},
			CheckUserError: nil,
		}},
		{"username tidak ditemukan", "POST", `{"username":"jason123", "password":"rahasia123"}`, 400, &MockUserRepository{
			CheckUserData: models.User{
				Username: "jason",
				Password: string(hashedPassword),
			},
			CheckUserError: errors.New("username tidak ditemukan"),
		}},
		{"password salah", "POST", `{"username":"jason", "password":"rahasia1231"}`, 400, &MockUserRepository{
			CheckUserData: models.User{
				Username: "jason",
				Password: string(hashedPassword),
			},
			CheckUserError: nil,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			UserRepo = tt.mockRepo
			var body *strings.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			} else {
				body = strings.NewReader("")
			}

			req := httptest.NewRequest(tt.method, "/login", body)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			HandlerLogin(rr, req)
			if rr.Code != tt.expectedStatus {
				t.Errorf("got %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}
