package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func algoritma(t *jwt.Token) (interface{}, error) {
	secretkey := viper.GetString("JWT_SECRET")
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

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := verifyToken(r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(401)
			json.NewEncoder(w).Encode(map[string]string{"error": "tidak valid"})
			return
		}
		ctx := context.WithValue(r.Context(), "claims", claims)
		next(w, r.WithContext(ctx))
	}
}

func GetUserID(r *http.Request) uint {
	claims := r.Context().Value("claims").(*jwt.MapClaims)
	userID := uint((*claims)["user_id"].(float64))
	return userID
}
