package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

var db *sql.DB

var jwtKey = []byte("your_secret_key")

type UserCredentials struct {
	Email string `json:"email"`
}

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func main() {

	//database
	var err error
	dsn := "host=localhost user=postgres password=your_password dbname=auth_ sslmode=disable"
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	err = db.Ping()
	if err != nil {
		panic("Database unreachable: " + err.Error())
	}

	//
	router := gin.Default()

	router.POST("/login", login)
	router.GET("/welcome", authenticateJWT(), welcome)

	router.Run(":8080")

}
func login(c *gin.Context) {

	var creds UserCredentials
	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid request"})
		return
	}
	///////////////////////
	var userExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", creds.Email).Scan(&userExists)

	if err != nil || !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	//////////////////////
	//credentials didn't match
	// if creds.Email != sampleEmail {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"Error": "Unauthorized"})
	// 	return
	// }

	expirationTime := time.Now().Add(15 * time.Minute)

	claims := &Claims{
		Email: creds.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenstr, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "Could not generate token"})
		return
	}

	_, err = db.Exec(
		"INSERT INTO user_tokens (email, token, expires_at) VALUES ($1, $2, $3)",
		creds.Email,
		tokenstr,
		expirationTime,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not store token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenstr})

}

func authenticateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "Missing token"})
			return
		}

		tokenstr := strings.TrimPrefix(authHeader, "Bearer ")
		tokenstr = strings.TrimSpace(tokenstr)

		// Check if token exists and isn't revoked in database
		var isRevoked bool
		var expiresAt time.Time
		err := db.QueryRow(
			"SELECT is_revoked, expires_at FROM user_tokens WHERE token = $1",
			tokenstr,
		).Scan(&isRevoked, &expiresAt)

		if err != nil || isRevoked || time.Now().After(expiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		/////////////////////////
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenstr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("email", claims.Email)
		c.Next()
	}
}
func welcome(c *gin.Context) {
	email := c.GetString("email")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Welcome %s!", email)})
}
