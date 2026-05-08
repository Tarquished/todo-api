package main

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"time"
	"todo-api/config"
	_ "todo-api/docs"
	"todo-api/handlers"
	"todo-api/middleware"
	"todo-api/repository"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title           Todo API
// @version         1.0
// @description     REST API untuk manajemen todo list dengan JWT authentication
// @description     Setiap user punya todo list sendiri (terisolasi by user_id)

// @contact.name    Jason
// @contact.url     https://github.com/Tarquished

// @host            todo-api-production-74d1.up.railway.app
// @schemes         https
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     Masukkan token dengan format: Bearer <token>

func main() {
	// Logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Config
	err := config.LoadConfig()
	if err != nil {
		log.Warn().Err(err).Msg("file .env tidak ditemukan, menggunakan variabel sistem")
	}

	// Database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		viper.GetString("DB_HOST"),
		viper.GetString("DB_USER"),
		viper.GetString("DB_PASSWORD"),
		viper.GetString("DB_NAME"),
		viper.GetString("DB_PORT"),
	)
	if viper.GetString("DB_HOST") == "" {
		dsn = viper.GetString("DATABASE_URL")
	}

	var db *gorm.DB
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Warn().Int("attempt", i+1).Err(err).Msg("database belum ready, retry...")
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("gagal konek ke database")
	}

	// Validator
	handlers.Validate = validator.New()
	handlers.Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Repository
	handlers.Repo = repository.NewPostgresTodoRepository(db)
	handlers.UserRepo = repository.NewPostgresUserRepository(db)

	// Routes
	http.HandleFunc("/register", middleware.CorsMiddleware(middleware.RecoveryMiddleware(handlers.HandlerRegister)))
	http.HandleFunc("/login", middleware.CorsMiddleware(middleware.RecoveryMiddleware(handlers.HandlerLogin)))
	http.HandleFunc("/tambah-todo", middleware.CorsMiddleware(middleware.RecoveryMiddleware(middleware.AuthMiddleware(handlers.HandlerTodoSingle))))
	http.HandleFunc("/tambah-todo-batch", middleware.CorsMiddleware(middleware.RecoveryMiddleware(middleware.AuthMiddleware(handlers.HandlerTodoBatch))))
	http.HandleFunc("/todos", middleware.CorsMiddleware(middleware.RecoveryMiddleware(middleware.AuthMiddleware(handlers.HandlerTodos))))
	http.HandleFunc("/hapus-todo", middleware.CorsMiddleware(middleware.RecoveryMiddleware(middleware.AuthMiddleware(handlers.HandlerHapusTodo))))
	http.HandleFunc("/update-todo", middleware.CorsMiddleware(middleware.RecoveryMiddleware(middleware.AuthMiddleware(handlers.HandlerUpdateTodo))))
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Start
	port := viper.GetString("PORT")
	if port == "" {
		port = "8080"
	}

	log.Info().Str("port", port).Msg("server started")
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}
