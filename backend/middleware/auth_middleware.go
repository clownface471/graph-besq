package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthAndRoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token dibutuhkan"})
			c.Abort()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return []byte("secret_key_besq_2026"), nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			roleClaim, ok := claims["role"].(string)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Role invalid"})
				c.Abort()
				return
			}

			authorized := false
			for _, role := range allowedRoles {
				if role == roleClaim {
					authorized = true
					break
				}
			}

			if !authorized {
				c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
				c.Abort()
				return
			}

			c.Set("userRole", roleClaim)
			c.Set("userID", claims["user_id"])
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token invalid", "details": err.Error()})
			c.Abort()
		}
	}
}