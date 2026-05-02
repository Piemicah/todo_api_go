package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"todo_api/internal/config"
	"todo_api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const accessCookieName = "access_token"

// AuthMiddleware reads the JWT from the HttpOnly cookie (not the Authorization
// header). It still accepts a Bearer token as a fallback so tools like curl and
// Postman keep working during development.
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString, err := ctx.Cookie(accessCookieName)

		// Fallback: accept Authorization: Bearer <token> header when no cookie.
		if err != nil || tokenString == "" {
			authHeader := ctx.GetHeader("Authorization")
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" || tokenString == authHeader {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
				ctx.Abort()
				return
			}
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			ctx.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims: missing user_id"})
			ctx.Abort()
			return
		}

		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().After(time.Unix(int64(exp), 0)) {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token has expired"})
				ctx.Abort()
				return
			}
		}

		roleStr, _ := claims["role"].(string)

		ctx.Set("user_id", userID)
		ctx.Set("role", roleStr)
		ctx.Next()
	}
}

// RequireRole aborts the request if the authenticated user's role is not in
// the allowed set. Must be placed after AuthMiddleware in the chain.
//
//	admin := router.Group("/admin")
//	admin.Use(middleware.AuthMiddleware(cfg), middleware.RequireRole(models.RoleAdmin))
func RequireRole(allowed ...models.Role) gin.HandlerFunc {
	allowedSet := make(map[models.Role]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}

	return func(ctx *gin.Context) {
		roleVal, exists := ctx.Get("role")
		if !exists {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "role not found in context"})
			ctx.Abort()
			return
		}

		role := models.Role(fmt.Sprintf("%v", roleVal))
		if _, ok := allowedSet[role]; !ok {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}