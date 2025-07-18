package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
)

func HandleSignup(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	if r.Method != http.MethodPost {
		slog.Warn("non-POST method", slog.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		slog.Error("invalid request body", slog.Any("error", err), slog.String("handler", "HandleSignup"))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	claims := &Claims{
		Email: creds.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenstr, err := token.SignedString(jwtKey)
	if err != nil {
		slog.Error("could not generate token", slog.Any("error", err), slog.String("handler", "HandleSignup"))
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	_, err = Db.Exec(`
       INSERT INTO user_tokens (email, token)
        VALUES ($1, $2)`, creds.Email, tokenstr)

	if err != nil {
		slog.Error("failed to store token", slog.Any("error", err.Error), slog.String("handler", "HandleSignup"))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenstr,
		"email": creds.Email,
	})

}

func AuthenticateJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			slog.Warn("authorization token missing")
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		tokenstr := strings.TrimPrefix(authHeader, "Bearer ")
		tokenstr = strings.TrimSpace(tokenstr)

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenstr, claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			slog.Error("invalid token", slog.Any("error", err))
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		var storedToken string
		err = Db.QueryRow(`
            SELECT token FROM user_tokens
            WHERE email = $1`, claims.Email).Scan(&storedToken)

		if err != nil || storedToken != tokenstr {
			slog.Error("invalid token", slog.Any("error", err))
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		// Store email in context or pass it along
		ctx := context.WithValue(r.Context(), "email", claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
