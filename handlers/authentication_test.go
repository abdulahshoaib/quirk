package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

var db *sql.DB

// Dummy next handler to confirm middleware passed
func dummyHandler(w http.ResponseWriter, r *http.Request) {
	email := r.Context().Value("email").(string)
	fmt.Fprintf(w, "Authenticated: %s", email)
}

func generateValidToken(email string) (string, error) {
	claims := &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func TestAuthenticateJWT_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()
	Db = mockDB

	tokenStr, err := generateValidToken("test@example.com")
	assert.NoError(t, err)

	// Expect query
	mock.ExpectQuery(`SELECT token FROM user_tokens WHERE email = \$1`).
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"token"}).AddRow(tokenStr))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()

	handler := AuthenticateJWT(dummyHandler)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := w.Body.String()
	assert.Contains(t, body, "Authenticated: test@example.com")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthenticateJWT_MissingToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler := AuthenticateJWT(dummyHandler)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "Missing token")
}

func TestAuthenticateJWT_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	w := httptest.NewRecorder()

	handler := AuthenticateJWT(dummyHandler)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

func TestAuthenticateJWT_TokenNotInDB(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer mockDB.Close()
	Db = mockDB

	tokenStr, err := generateValidToken("notfound@example.com")
	assert.NoError(t, err)

	mock.ExpectQuery(`SELECT token FROM user_tokens WHERE email = \$1`).
		WithArgs("notfound@example.com").
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()

	handler := AuthenticateJWT(dummyHandler)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, w.Body.String(), "Invalid token")
}
