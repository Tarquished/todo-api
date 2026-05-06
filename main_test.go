package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidasiTodo(t *testing.T) {
	tests := []struct {
		nama     string
		input    listTodo
		expected string
	}{
		{
			nama:     "judul kosong",
			input:    listTodo{Judul: "", Prioritas: "tinggi"},
			expected: "judul harus terisi",
		},
		{
			nama:     "prioritas kosong",
			input:    listTodo{Judul: "judul", Prioritas: ""},
			expected: "prioritas harus terisi",
		},
		{
			nama:     "prioritas tidak valid",
			input:    listTodo{Judul: "judul", Prioritas: "ada"},
			expected: "prioritas harus berupa tinggi/sedang/rendah",
		},
		{
			nama:     "valid",
			input:    listTodo{Judul: "judul", Prioritas: "tinggi"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			result := validasiTodo(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
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
	}{
		{"body kosong", "POST", "", 400},
		{"method salah", "GET", "", 405},
		{"username kosong", "POST", `{"username":"","password":"rahasia"}`, 400},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
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
	tests := []struct {
		nama           string
		method         string
		body           string
		expectedStatus int
	}{
		{"method salah", "GET", "", 405},
		{"body kosong", "POST", "", 400},
		{"username kosong", "POST", `{"username":"", "password":"rahasia"}`, 400},
	}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			var body *strings.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			} else {
				body = strings.NewReader("")
			}

			req := httptest.NewRequest(tt.method, "/register", body)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handlerLogin(rr, req)
			if rr.Code != tt.expectedStatus {
				t.Errorf("got %d, want %d", rr.Code, tt.expectedStatus)
			}
		})
	}

}
