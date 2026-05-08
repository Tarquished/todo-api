package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"todo-api/models"
	"todo-api/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

var UserRepo repository.UserRepository

// HandlerRegister godoc
// @Summary      Register user baru
// @Description  Mendaftarkan user baru dengan username dan password
// @Description  Password akan di-hash menggunakan bcrypt sebelum disimpan
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.InputAuth      true  "Username dan password"
// @Success      200      {object}  models.ResponPesan
// @Failure      400      {object}  models.ResponError
// @Failure      405      {object}  models.ResponError
// @Router       /register [post]
func HandlerRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		SendError(w, "method harus POST", 405)
		return
	}

	var input models.InputAuth
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		SendError(w, "format JSON tidak valid", 400)
		return
	}

	if input.Username == "" {
		SendError(w, "mohon isi username", 400)
		return
	}
	if input.Password == "" {
		SendError(w, "mohon isi password", 400)
		return
	}

	_, results := UserRepo.CheckUser(input.Username)
	if results == nil {
		SendError(w, "username sudah ada", 400)
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
	if err != nil {
		SendError(w, "error saat hashing password", 400)
		return
	}

	err = UserRepo.RegisterUser(models.User{
		Username: input.Username,
		Password: string(hashPassword),
	})

	if err != nil {
		SendError(w, "gagal menyimpan user", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ResponPesan{
		Pesan: "Berhasil menambahkan username ke database",
	})
}

// HandlerLogin godoc
// @Summary      Login user
// @Description  Melakukan login dengan username dan password yang sudah ada dalam database
// @Description  Password akan diverify dan akan diberikan JWT Token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.InputAuth      true  "Username dan password"
// @Success      200      {object}  models.SuccessResponLogin
// @Failure      400      {object}  models.ResponError
// @Failure      405      {object}  models.ResponError
// @Router       /login [post]
func HandlerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		SendError(w, "method harus POST", 405)
		return
	}

	var input models.InputAuth
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		SendError(w, "format JSON salah", 400)
		return
	}

	if input.Username == "" {
		SendError(w, "mohon isi username", 400)
		return
	}
	if input.Password == "" {
		SendError(w, "mohon isi password", 400)
		return
	}

	user, results := UserRepo.CheckUser(input.Username)
	if results != nil {
		SendError(w, "username belum ada", 400)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		SendError(w, "password salah", 400)
		return
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretkey := viper.GetString("JWT_SECRET")
	if secretkey == "" {
		secretkey = "test1625jason34"
	}
	tokenString, err := token.SignedString([]byte(secretkey))
	if err != nil {
		SendError(w, "gagal generate token", 400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessResponLogin{
		Pesan: "Berhasil login!",
		Token: tokenString,
	})
}
