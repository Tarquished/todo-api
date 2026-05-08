package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
)

func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Interface("panic", err).
					Msg("panic recovered")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(500)
				json.NewEncoder(w).Encode(map[string]string{"error": "terjadi kesalahan"})
			}
		}()
		next(w, r)
	}
}
