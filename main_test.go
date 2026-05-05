package main

import "testing"

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
