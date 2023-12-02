package middlewares

import (
	"daijai/token"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		token.ExtractToken(c, &tokenString)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("SECRET")), nil
		})

		if err != nil || !token.Valid {
			log.Printf("Invalid or expired token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Check if any of the user's roles match the required roles
		userRole, ok := claims["role"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid roles in token"})
			c.Abort()
			return
		}

		roleMatched := false
		for _, role := range roles {
			if userRole == role {
				roleMatched = true
				break
			}
		}

		if !roleMatched {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Insufficient role privileges"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// func AuthMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		config := config.GetConfig()
// 		reqKey := c.Request.Header.Get("X-Auth-Key")
// 		reqSecret := c.Request.Header.Get("X-Auth-Secret")

// 		var key string
// 		var secret string
// 		if key = config.GetString("http.auth.key"); len(strings.TrimSpace(key)) == 0 {
// 			c.AbortWithStatus(500)
// 		}
// 		if secret = config.GetString("http.auth.secret"); len(strings.TrimSpace(secret)) == 0 {
// 			c.AbortWithStatus(401)
// 		}
// 		if key != reqKey || secret != reqSecret {
// 			c.AbortWithStatus(401)
// 			return
// 		}
// 		c.Next()
// 	}
// }
