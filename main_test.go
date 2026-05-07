package main

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	CheckUserData  User
	CheckUserError error
	RegisterError  error
}

func (m *MockUserRepository) CheckUser(username string) (User, error) {
	return m.CheckUserData, m.CheckUserError
}

func (m *MockUserRepository) RegisterUser(user User) error {
	return m.RegisterError
}

func TestFormatValidationError(t *testing.T) {
	tests := []struct {
		nama     string
		input    listTodo
		expected []string
	}{
		{
			nama:     "judul kosong",
			input:    listTodo{Judul: "", Prioritas: "tinggi"},
			expected: []string{"judul harus terisi"},
		},
		{
			nama:     "prioritas kosong",
			input:    listTodo{Judul: "judul", Prioritas: ""},
			expected: []string{"prioritas harus terisi"},
		},
		{
			nama:     "prioritas tidak valid",
			input:    listTodo{Judul: "judul", Prioritas: "ada"},
			expected: []string{"prioritas harus berupa tinggi sedang rendah"},
		},
		{
			nama:     "judul dan prioritas kosong",
			input:    listTodo{Judul: "", Prioritas: ""},
			expected: []string{"judul harus terisi", "prioritas harus terisi"},
		},
		{
			nama:     "valid",
			input:    listTodo{Judul: "judul", Prioritas: "tinggi"},
			expected: nil, // gak ada error
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			err := validate.Struct(tt.input)
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
			handlerRegister(rr, req)
			if rr.Code != tt.expectedStatus {
				t.Errorf("got %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}
}

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
			CheckUserData: User{
				Username: "jason",
				Password: string(hashedPassword),
			},
			CheckUserError: nil,
		}},
		{"username tidak ditemukan", "POST", `{"username":"jason123", "password":"rahasia123"}`, 400, &MockUserRepository{
			CheckUserData: User{
				Username: "jason",
				Password: string(hashedPassword),
			},
			CheckUserError: errors.New("username tidak ditemukan"),
		}},
		{"password salah", "POST", `{"username":"jason", "password":"rahasia1231"}`, 400, &MockUserRepository{
			CheckUserData: User{
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
			handlerLogin(rr, req)
			if rr.Code != tt.expectedStatus {
				t.Errorf("got %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}

}
