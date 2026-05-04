package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type Todo struct {
	gorm.Model
	Judul     string
	Prioritas string
}

type listTodo struct {
	Judul     string `json:"judul"`
	Prioritas string `json:"prioritas"`
}

type listTodoBatch struct {
	Judul     string `json:"judul"`
	Prioritas string `json:"prioritas,omitempty"`
	Status    string `json:"status,omitempty"`
	Error     string `json:"error,omitempty"`
}

type getTodo struct {
	ID        int    `json:"id"`
	Judul     string `json:"judul"`
	Prioritas string `json:"prioritas"`
}

type ResponError struct {
	Error string `json:"error"`
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ResponError{Error: message})
}

func validasiTodo(inputs listTodo) string {
	if inputs.Judul == "" {
		return "judul harus terisi"
	}
	if inputs.Prioritas == "" {
		return "prioritas harus terisi"
	}
	if inputs.Prioritas != "tinggi" && inputs.Prioritas != "sedang" && inputs.Prioritas != "rendah" {
		return "prioritas harus berupa tinggi/sedang/rendah"
	}
	return ""
}

func handlerTodoSingle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendError(w, "method harus POST", 405)
		return
	}

	var inputs listTodo
	err := json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		sendError(w, "format JSON tidak valid", 400)
		return
	}

	if pesan := validasiTodo(inputs); pesan != "" {
		sendError(w, pesan, 400)
		return
	}

	db.Create(&Todo{
		Judul:     inputs.Judul,
		Prioritas: inputs.Prioritas,
	})

	hasil := map[string]any{
		"pesan":     "Todo berhasil ditambahkan",
		"judul":     inputs.Judul,
		"prioritas": inputs.Prioritas,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hasil)
}

func handlerTodoBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendError(w, "method harus POST", 405)
		return
	}

	var inputs []listTodo
	err := json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		sendError(w, "format JSON tidak valid", 400)
		return
	}

	if len(inputs) == 0 {
		sendError(w, "data tidak boleh kosong", 400)
		return
	}

	var hasil []listTodoBatch
	for _, v := range inputs {
		if pesan := validasiTodo(v); pesan != "" {
			hasil = append(hasil, listTodoBatch{
				Judul: v.Judul,
				Error: pesan,
			})
			continue
		}
		hasil = append(hasil, listTodoBatch{
			Judul:     v.Judul,
			Prioritas: v.Prioritas,
			Status:    "berhasil",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hasil)
}

func handlerTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		sendError(w, "method harus GET", 405)
		return
	}

	var todos []Todo
	db.Find(&todos)

	var hasil []getTodo
	for _, v := range todos {
		hasil = append(hasil, getTodo{
			ID:        int(v.ID),
			Judul:     v.Judul,
			Prioritas: v.Prioritas,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hasil)
}

func handlerHapusTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		sendError(w, "method harus DELETE", 405)
		return
	}

	strID := r.URL.Query().Get("id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		sendError(w, "ID tidak valid", 400)
		return
	}
	if id == 0 {
		sendError(w, "ID tidak terdaftar", 400)
		return
	}

	db.Delete(&Todo{}, id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"pesan": "Todo berhasil dihapus",
	})
}

func handlerUpdateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		sendError(w, "method harus PUT", 405)
		return
	}

	strID := r.URL.Query().Get("id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		sendError(w, "ID tidak valid", 400)
		return
	}
	if id == 0 {
		sendError(w, "ID tidak terdaftar", 400)
		return
	}

	var inputs listTodo
	err = json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		sendError(w, "format JSON tidak valid", 400)
		return
	}

	if pesan := validasiTodo(inputs); pesan != "" {
		sendError(w, pesan, 400)
		return
	}

	db.Model(&Todo{}).Where("id = ?", id).Updates(map[string]any{
		"judul":     inputs.Judul,
		"prioritas": inputs.Prioritas,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"pesan": "Todo berhasil diupdate",
	})
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=test162534 dbname=todoapp port=5432 sslmode=disable"
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Gagal konek ke database:", err)
		return
	}

	db.AutoMigrate(&Todo{})

	http.HandleFunc("/tambah-todo", handlerTodoSingle)
	http.HandleFunc("/tambah-todo-batch", handlerTodoBatch)
	http.HandleFunc("/todos", handlerTodos)
	http.HandleFunc("/hapus-todo", handlerHapusTodo)
	http.HandleFunc("/update-todo", handlerUpdateTodo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server jalan di port", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
