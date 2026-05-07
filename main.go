package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"context"
	"log"
	"strings"
	"time"

	_ "todo-api/docs"

	"github.com/golang-jwt/jwt/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var repo TodoRepository

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

type ResponPesan struct {
	Pesan string `json:"pesan" example:"Berhasil menambahkan username ke database"`
}

func getUserID(r *http.Request) uint {
	claims := r.Context().Value("claims").(*jwt.MapClaims)
	userID := uint((*claims)["user_id"].(float64))
	return userID
}

// handlerRegister godoc
// @Summary      Register user baru
// @Description  Mendaftarkan user baru dengan username dan password
// @Description  Password akan di-hash menggunakan bcrypt sebelum disimpan
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      InputAuth      true  "Username dan password"
// @Success      200      {object}  ResponPesan
// @Failure      400      {object}  ResponError
// @Failure      405      {object}  ResponError
// @Router       /register [post]
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ResponPesan{
		Pesan: "Berhasil menambahkan username ke database",
	})
}

type SuccessResponLogin struct {
	Pesan string `json:"pesan" example:"Berhasil login!"`
	Token string `json:"token"`
}

// handlerLogin godoc
// @Summary 	Login user
// @Description	Melakukan login dengan username dan password yang sudah ada dalam database
// @Description	Password akan diverify dan akan diberikan JWT Token
// @Tags		Auth
// @Accept		json
// @Produce		json
// @Param		request	body		InputAuth	true	"Username dan password"
// @Success		200		{object}	SuccessResponLogin
// @Failure		400		{object}	ResponError
// @Failure		405		{object}	ResponError
// @Router		/login	[post]
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessResponLogin{
		Pesan: "Berhasil login!",
		Token: tokenString,
	})
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

type SuccessHandlerTodoSingle struct {
	Pesan     string `json:"pesan"`
	Judul     string `json:"judul"`
	Prioritas string `json:"prioritas"`
}

// handlerTodoSingle godoc
// @Summary	Membuat To-Do secara single
// @Description	Mendaftarkan To-Do ke dalam database dengan melakukan verify JWT terlebih dahulu
// @Tags		Todo
// @Accept		json
// @Produce		json
// @Param		request	body	listTodo	true	"Judul dan Prioritas"
// @Security 	BearerAuth
// @Success		200		{object}	SuccessHandlerTodoSingle
// @Failure		400		{object}	ResponError
// @Failure		405		{object}	ResponError
// @Router		/tambah-todo	[post]
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

	err = repo.CreateTodo(Todo{
		Judul:     inputs.Judul,
		Prioritas: inputs.Prioritas,
		UserID:    userID,
	})

	if err != nil {
		log.Printf("ERROR handlerTodoSingle - db.Create failed: %v, userID: %d", err, userID)
		sendError(w, "gagal menyimpan data", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessHandlerTodoSingle{
		Pesan:     "Todo berhasil ditambahkan",
		Judul:     inputs.Judul,
		Prioritas: inputs.Prioritas,
	})
}

// handlerTodoBatch godoc
// @Summary	Membuat To-Do secara batch
// @Description	Mendaftarkan To-Do batch ke dalam database dengan melakukan verify JWT terlebih dahulu
// @Tags		Todo
// @Accept		json
// @Produce		json
// @Param		request	body	[]listTodo	true	"List todo"
// @Security 	BearerAuth
// @Success		200		{array}		listTodoBatch
// @Failure		400		{object}	ResponError
// @Failure		405		{object}	ResponError
// @Router		/tambah-todo-batch	[post]
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

type PageData struct {
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
	Total int64     `json:"total"`
	Data  []getTodo `json:"data"`
}

// handlerTodos godoc
// @Summary		Menampilkan semua To-Do User
// @Description	Menampilkan isi dari semua To-Do yang dimiliki oleh user
// @Description	Memverifikasi ID dan token yang dipakai melalui JWT verification
// @Tags		Todo
// @Produce		json
// @Param		page	query	int	false	"Nomor halaman (default: 1)"
// @Param		limit	query	int	false	"Jumlah data per halaman (default: 10)"
// @Security 	BearerAuth
// @Success		200		{object}	PageData
// @Failure		405		{object}	ResponError
// @Router		/todos	[get]
func handlerTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		sendError(w, "method harus GET", 405)
		return
	}

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit
	userID := getUserID(r)

	todos, total, err := repo.GetTodos(userID, limit, offset)

	if err != nil {
		log.Printf("ERROR handlerTodos - GetTodos failed: %v, userID: %d", err, userID)
		sendError(w, "gagal mengambil data", 500)
		return
	}

	var hasil []getTodo
	for _, v := range todos {
		hasil = append(hasil, getTodo{
			ID:        int(v.ID),
			Judul:     v.Judul,
			Prioritas: v.Prioritas,
		})
	}
	page_data := PageData{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  hasil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(page_data)
}

type SuccessHandlerHapusUpdateTodo struct {
	Pesan string `json:"pesan"`
}

// handlerHapusTodo godoc
// @Summary		Menghapus To-Do berdasarkan ID
// @Description	Untuk menghapus suatu To-Do dengan ID yang diberikan
// @Description	Diverifikasi melalui JWT Token untuk mengecek kepemilikan
// @Tags		Todo
// @Produce		json
// @Param		id	query	int	true	"ID"
// @Security	BearerAuth
// @Success		200		{object}	SuccessHandlerHapusUpdateTodo
// @Failure		400		{object}	ResponError
// @Failure		405		{object}	ResponError
// @Router		/hapus-todo	[delete]
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
	err = repo.DeleteTodo(id, userID)
	if err != nil {
		log.Printf("ERROR handlerHapusTodo - db.Delete failed: %v, userID: %d", err, userID)
		sendError(w, "gagal menghapus data", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessHandlerHapusUpdateTodo{
		Pesan: "Todo berhasil dihapus",
	})
}

// handlerUpdateTodo godoc
// @Summary		Mengupdate To-Do berdasarkan ID
// @Description	Untuk mengubah suatu To-Do dengan ID yang diberikan
// @Description	Diverifikasi melalui JWT Token untuk mengecek kepemilikan
// @Tags		Todo
// @Produce		json
// @Accept		json
// @Param		id	query	int	true	"ID"
// @Param		request	body	listTodo	true	"listTodo"
// @Security	BearerAuth
// @Success		200		{object}	SuccessHandlerHapusUpdateTodo
// @Failure		400		{object}	ResponError
// @Failure		405		{object}	ResponError
// @Router		/update-todo	[put]
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
	err = repo.UpdateTodo(id, userID, inputs.Judul, inputs.Prioritas)

	if err != nil {
		log.Printf("ERROR handlerUpdateTodo - db.Update failed: %v, userID: %d", err, userID)
		sendError(w, "gagal mengupdate data", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessHandlerHapusUpdateTodo{
		Pesan: "Todo berhasil diupdate",
	})
}

func recoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("ERROR recoveryMiddlewar:%v", err)
				sendError(w, "terjadi kesalahan", 500)
			}
		}()
		next(w, r)
	}

}

// @title           Todo API
// @version         1.0
// @description     REST API untuk manajemen todo list dengan JWT authentication
// @description     Setiap user punya todo list sendiri (terisolasi by user_id)

// @contact.name    Jason
// @contact.url     https://github.com/Tarquished

// @host            todo-api-production-74d1.up.railway.app
// @schemes 		https
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Masukkan token dengan format: Bearer <token>

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
	repo = NewPostgresTodoRepository(db)

	db.AutoMigrate(&Todo{})
	db.AutoMigrate(&User{})
	http.HandleFunc("/register", recoveryMiddleware(handlerRegister))
	http.HandleFunc("/login", recoveryMiddleware(handlerLogin))
	http.HandleFunc("/tambah-todo", recoveryMiddleware(authMiddleware(handlerTodoSingle)))
	http.HandleFunc("/tambah-todo-batch", recoveryMiddleware(authMiddleware(handlerTodoBatch)))
	http.HandleFunc("/todos", recoveryMiddleware(authMiddleware(handlerTodos)))
	http.HandleFunc("/hapus-todo", recoveryMiddleware(authMiddleware(handlerHapusTodo)))
	http.HandleFunc("/update-todo", recoveryMiddleware(authMiddleware(handlerUpdateTodo)))
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server jalan di port", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
