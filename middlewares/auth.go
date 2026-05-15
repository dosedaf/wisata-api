package middlewares

import (
	"net/http"
	"strings"
	"time"
	"wisata-api/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var JWT_SECRET = []byte("secret af")

func GenerateToken(userID uint, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Expire 1 hari
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWT_SECRET)
}

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse("Akses ditolak, token tidak ada", nil))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return JWT_SECRET, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse("Token tidak valid", nil))
			c.Abort()
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		c.Set("user_id", uint(claims["user_id"].(float64)))
		c.Set("role", claims["role"].(string))
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "ADMIN" {
			c.JSON(http.StatusForbidden, utils.ErrorResponse("Akses ditolak, butuh hak akses Admin", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}
