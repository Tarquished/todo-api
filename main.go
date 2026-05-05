package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type Todo struct {
	gorm.Model
	UserID    uint `json:"user_id"`
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

type User struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
}

type InputAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func getUserID(r *http.Request) uint {
	claims := r.Context().Value("claims").(*jwt.MapClaims)
	userID := uint((*claims)["user_id"].(float64))
	return userID
}

func handlerRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendError(w, "method harus POST", 405)
		return
	}

	var input InputAuth
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		sendError(w, "format JSON tidak valid", 400)
		return
	}

	if input.Username == "" {
		sendError(w, "mohon isi username", 400)
		return
	}
	if input.Password == "" {
		sendError(w, "mohon isi password", 400)
		return
	}

	var user User

	results := db.Where("username = ?", input.Username).First(&user)
	if results.Error == nil {
		sendError(w, "username sudah ada", 400)
		return
	}
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
	if err != nil {
		sendError(w, "error saat hashing password", 400)
		return
	}
	db.Create(&User{
		Username: input.Username,
		Password: string(hashPassword),
	})
	succesRespon := map[string]any{
		"pesan": "Berhasil menambahkan username ke database",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(succesRespon)
}

func handlerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendError(w, "method harus POST", 405)
		return
	}

	var input InputAuth
	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		sendError(w, "format JSON salah", 400)
		return
	}
	if input.Username == "" {
		sendError(w, "mohon isi username", 400)
		return
	}
	if input.Password == "" {
		sendError(w, "mohon isi password", 400)
		return
	}

	var user User

	results := db.Where("username = ?", input.Username).First(&user)
	if results.Error != nil {
		sendError(w, "username belum ada", 400)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		sendError(w, "password salah", 400)
		return
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretkey := os.Getenv("JWT_SECRET")
	if secretkey == "" {
		secretkey = "test1625jason34"
	}
	tokenString, err := token.SignedString([]byte(secretkey))

	if err != nil {
		sendError(w, "gagal generate token", 400)
		return
	}
	succesRespon := map[string]any{
		"pesan": "Berhasil login!",
		"token": tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(succesRespon)
}

func algoritma(t *jwt.Token) (interface{}, error) {
	secretkey := os.Getenv("JWT_SECRET")
	if secretkey == "" {
		secretkey = "test1625jason34"
	}
	return []byte(secretkey), nil
}

func verifyToken(r *http.Request) (*jwt.MapClaims, error) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims := &jwt.MapClaims{}

	responToken, err := jwt.ParseWithClaims(token, claims, algoritma)

	if err != nil {
		return nil, err
	}
	if !responToken.Valid {
		return nil, fmt.Errorf("token tidak valid")
	}
	return claims, nil
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := verifyToken(r)
		if err != nil {
			sendError(w, "tidak valid", 401)
			return
		}
		ctx := context.WithValue(r.Context(), "claims", claims)
		next(w, r.WithContext(ctx))
	}
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

	userID := getUserID(r)

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
		UserID:    userID,
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
	userID := getUserID(r)
	db.Where("user_id = ?", userID).Find(&todos)

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
	userID := getUserID(r)
	db.Where("id = ? AND user_id = ?", id, userID).Delete(&Todo{})

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
	userID := getUserID(r)
	db.Model(&Todo{}).Where("id = ? AND user_id = ?", id, userID).Updates(map[string]any{
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
	db.AutoMigrate(&User{})
	http.HandleFunc("/register", handlerRegister)
	http.HandleFunc("/login", handlerLogin)
	http.HandleFunc("/tambah-todo", authMiddleware(handlerTodoSingle))
	http.HandleFunc("/tambah-todo-batch", authMiddleware(handlerTodoBatch))
	http.HandleFunc("/todos", authMiddleware(handlerTodos))
	http.HandleFunc("/hapus-todo", authMiddleware(handlerHapusTodo))
	http.HandleFunc("/update-todo", authMiddleware(handlerUpdateTodo))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server jalan di port", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
