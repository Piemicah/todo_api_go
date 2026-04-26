package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"todo_api/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc{
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")

		if authHeader == ""{
			ctx.JSON(http.StatusUnauthorized,gin.H{"error":"Authorization header required"})
			ctx.Abort()
			return 
		}

		tokenString := strings.TrimPrefix(authHeader,"Bearer ")

		if tokenString == "" || tokenString == authHeader{
			ctx.JSON(http.StatusUnauthorized,gin.H{"error":"Invalid Authorization header format"})
			ctx.Abort()
			return
		}

		token,err := jwt.Parse(tokenString,func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg(){
				return nil,fmt.Errorf("Unexpected signing method: %v",t.Header["alg"])
			}
			return  []byte(cfg.JWTSecret),nil
		})
       
		if err != nil || !token.Valid{
			ctx.JSON(http.StatusUnauthorized,gin.H{"error":"Invalid or expired token"})
			ctx.Abort()
			return 
		}

		claims,ok := token.Claims.(jwt.MapClaims)

		if !ok{
			ctx.JSON(http.StatusUnauthorized,gin.H{"error":"Invalid token claims"})
			ctx.Abort()
			return 
		}
		
		userID,ok := claims["user_id"].(string)

		if !ok{
			ctx.JSON(http.StatusUnauthorized,gin.H{"error":"Invalid token claims"})
			ctx.Abort()
			return 
		}

		if exp,ok:=claims["exp"].(float64); ok{
			expirationTime := time.Unix(int64(exp),0)
			if time.Now().After(expirationTime){
			ctx.JSON(http.StatusUnauthorized,gin.H{"error":"token has expired"})
			ctx.Abort()
			return 
		}
		}

		ctx.Set("user_id",userID)
		ctx.Next()
	}
}