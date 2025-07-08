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

	var err error
	dsn := "host=localhost user=eesa password=babygirl dbname=users sslmode=disable"
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	err = db.Ping()
	if err != nil {
		panic("Database unreachable: " + err.Error())
	}

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

	//credentials didn't match
	// if creds.Email != sampleEmail {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"Error": "Unauthorized"})
	// 	return
	// }

	expirationTime := time.Now().Add(5 * time.Minute)

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
