package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"todo-api/middleware"
	"todo-api/models"
	"todo-api/repository"

	"github.com/rs/zerolog/log"
)

var Repo repository.TodoRepository

// HandlerTodoSingle godoc
// @Summary      Membuat To-Do secara single
// @Description  Mendaftarkan To-Do ke dalam database dengan melakukan verify JWT terlebih dahulu
// @Tags         Todo
// @Accept       json
// @Produce      json
// @Param        request  body  models.ListTodo  true  "Judul dan Prioritas"
// @Security     BearerAuth
// @Success      200  {object}  models.SuccessHandlerTodoSingle
// @Failure      400  {object}  models.ResponError
// @Failure      405  {object}  models.ResponError
// @Router       /tambah-todo [post]
func HandlerTodoSingle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		SendError(w, "method harus POST", 405)
		return
	}

	userID := middleware.GetUserID(r)

	var inputs models.ListTodo
	err := json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		SendError(w, "format JSON tidak valid", 400)
		return
	}

	err = Validate.Struct(inputs)
	if err != nil {
		pesanError := FormatValidationError(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "fail",
			"errors": pesanError,
		})
		return
	}

	err = Repo.CreateTodo(models.Todo{
		Judul:     inputs.Judul,
		Prioritas: inputs.Prioritas,
		UserID:    userID,
	})

	if err != nil {
		log.Error().
			Err(err).
			Uint("userID", userID).
			Str("handler", "HandlerTodoSingle").
			Msg("gagal menyimpan todo")
		SendError(w, "gagal menyimpan data", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessHandlerTodoSingle{
		Pesan:     "Todo berhasil ditambahkan",
		Judul:     inputs.Judul,
		Prioritas: inputs.Prioritas,
	})
}

// HandlerTodoBatch godoc
// @Summary      Membuat To-Do secara batch
// @Description  Mendaftarkan To-Do batch ke dalam database dengan melakukan verify JWT terlebih dahulu
// @Tags         Todo
// @Accept       json
// @Produce      json
// @Param        request  body  []models.ListTodo  true  "List todo"
// @Security     BearerAuth
// @Success      200  {array}   models.ListTodoBatch
// @Failure      400  {object}  models.ResponError
// @Failure      405  {object}  models.ResponError
// @Router       /tambah-todo-batch [post]
func HandlerTodoBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		SendError(w, "method harus POST", 405)
		return
	}

	userID := middleware.GetUserID(r)

	var inputs []models.ListTodo
	err := json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		SendError(w, "format JSON tidak valid", 400)
		return
	}

	if len(inputs) == 0 {
		SendError(w, "data tidak boleh kosong", 400)
		return
	}

	hasil := make([]models.ListTodoBatch, 0, len(inputs))

	for _, v := range inputs {
		if err := Validate.Struct(v); err != nil {
			pesanErrors := FormatValidationError(err)
			hasil = append(hasil, models.ListTodoBatch{
				Judul: v.Judul,
				Error: strings.Join(pesanErrors, ", "),
			})
			continue
		}

		err := Repo.CreateTodo(models.Todo{
			Judul:     v.Judul,
			Prioritas: v.Prioritas,
			UserID:    userID,
		})
		if err != nil {
			log.Error().
				Err(err).
				Uint("userID", userID).
				Str("handler", "HandlerTodoBatch").
				Msg("gagal menyimpan todo")
			hasil = append(hasil, models.ListTodoBatch{
				Judul: v.Judul,
				Error: "gagal menyimpan data",
			})
			continue
		}

		hasil = append(hasil, models.ListTodoBatch{
			Judul:     v.Judul,
			Prioritas: v.Prioritas,
			Status:    "berhasil",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hasil)
}

// HandlerTodos godoc
// @Summary      Menampilkan semua To-Do User
// @Description  Menampilkan isi dari semua To-Do yang dimiliki oleh user
// @Tags         Todo
// @Produce      json
// @Param        page   query  int  false  "Nomor halaman (default: 1)"
// @Param        limit  query  int  false  "Jumlah data per halaman (default: 10)"
// @Security     BearerAuth
// @Success      200  {object}  models.PageData
// @Failure      405  {object}  models.ResponError
// @Router       /todos [get]
func HandlerTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		SendError(w, "method harus GET", 405)
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
	userID := middleware.GetUserID(r)

	todos, total, err := Repo.GetTodos(userID, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Uint("userID", userID).
			Str("handler", "HandlerTodos").
			Msg("gagal mengambil todos")
		SendError(w, "gagal mengambil data", 500)
		return
	}

	var hasil []models.GetTodo
	for _, v := range todos {
		hasil = append(hasil, models.GetTodo{
			ID:        int(v.ID),
			Judul:     v.Judul,
			Prioritas: v.Prioritas,
		})
	}

	pageData := models.PageData{
		Page:  page,
		Limit: limit,
		Total: total,
		Data:  hasil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pageData)
}

// HandlerHapusTodo godoc
// @Summary      Menghapus To-Do berdasarkan ID
// @Description  Untuk menghapus suatu To-Do dengan ID yang diberikan
// @Tags         Todo
// @Produce      json
// @Param        id  query  int  true  "ID"
// @Security     BearerAuth
// @Success      200  {object}  models.SuccessHandlerHapusUpdateTodo
// @Failure      400  {object}  models.ResponError
// @Failure      405  {object}  models.ResponError
// @Router       /hapus-todo [delete]
func HandlerHapusTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		SendError(w, "method harus DELETE", 405)
		return
	}

	strID := r.URL.Query().Get("id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		SendError(w, "ID tidak valid", 400)
		return
	}
	if id == 0 {
		SendError(w, "ID tidak terdaftar", 400)
		return
	}

	userID := middleware.GetUserID(r)
	err = Repo.DeleteTodo(id, userID)
	if err != nil {
		log.Error().
			Err(err).
			Uint("userID", userID).
			Str("handler", "HandlerHapusTodo").
			Msg("gagal menghapus todo")
		SendError(w, "gagal menghapus data", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessHandlerHapusUpdateTodo{
		Pesan: "Todo berhasil dihapus",
	})
}

// HandlerUpdateTodo godoc
// @Summary      Mengupdate To-Do berdasarkan ID
// @Description  Untuk mengubah suatu To-Do dengan ID yang diberikan
// @Tags         Todo
// @Produce      json
// @Accept       json
// @Param        id       query  int             true  "ID"
// @Param        request  body   models.ListTodo  true  "listTodo"
// @Security     BearerAuth
// @Success      200  {object}  models.SuccessHandlerHapusUpdateTodo
// @Failure      400  {object}  models.ResponError
// @Failure      405  {object}  models.ResponError
// @Router       /update-todo [put]
func HandlerUpdateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		SendError(w, "method harus PUT", 405)
		return
	}

	strID := r.URL.Query().Get("id")
	id, err := strconv.Atoi(strID)
	if err != nil {
		SendError(w, "ID tidak valid", 400)
		return
	}
	if id == 0 {
		SendError(w, "ID tidak terdaftar", 400)
		return
	}

	var inputs models.ListTodo
	err = json.NewDecoder(r.Body).Decode(&inputs)
	if err != nil {
		SendError(w, "format JSON tidak valid", 400)
		return
	}

	err = Validate.Struct(inputs)
	if err != nil {
		pesanError := FormatValidationError(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "fail",
			"errors": pesanError,
		})
		return
	}

	userID := middleware.GetUserID(r)
	err = Repo.UpdateTodo(id, userID, inputs.Judul, inputs.Prioritas)
	if err != nil {
		log.Error().
			Err(err).
			Uint("userID", userID).
			Str("handler", "HandlerUpdateTodo").
			Msg("gagal mengupdate todo")
		SendError(w, "gagal mengupdate data", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SuccessHandlerHapusUpdateTodo{
		Pesan: "Todo berhasil diupdate",
	})
}
