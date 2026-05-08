package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"todo-api/models"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func SendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.ResponError{Error: message})
}

func FormatValidationError(err error) []string {
	var validationErrs validator.ValidationErrors
	var errorMessages []string

	if errors.As(err, &validationErrs) {
		for _, e := range validationErrs {
			switch e.Tag() {
			case "required":
				pesan := fmt.Sprintf("%s harus terisi", e.Field())
				errorMessages = append(errorMessages, pesan)
			case "oneof":
				pesan := fmt.Sprintf("%s harus berupa %s", e.Field(), e.Param())
				errorMessages = append(errorMessages, pesan)
			}
		}
	}

	return errorMessages
}
