package handlers

import (
	"net/http"
	"strings"
	"time"
	"todo_api/internal/config"
	"todo_api/internal/models"
	"todo_api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct{
	Token string `json:"token"`
}

func CreateUserHandler(pool *pgxpool.Pool) gin.HandlerFunc{
	return func(ctx *gin.Context) {
		var registerRequest RegisterRequest

		if err := ctx.ShouldBindJSON(&registerRequest); err != nil{
			ctx.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return 
		}

		if len(registerRequest.Password) < 6{
			ctx.JSON(http.StatusBadRequest,gin.H{"error":"Password must be at least 6 characters long"})
			return 
		}

		hashedPassword,err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password),bcrypt.DefaultCost)

		if err != nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to hash password"+err.Error()})
			return 
		}

		user := models.User{
			Email: registerRequest.Email,
			Password: string(hashedPassword),
		}

		 createdUser,err := repository.CreateUser(pool,&user)

		 if err != nil{
			if strings.Contains(err.Error(),"duplicate") || strings.Contains(err.Error(),"unique"){
				ctx.JSON(http.StatusBadRequest,gin.H{"error":"email already registered"})
				return 
			}
			ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return 
		 }

		 ctx.JSON(http.StatusCreated,createdUser)

	}
}

func LoginHandler(pool *pgxpool.Pool,cfg *config.Config) gin.HandlerFunc{
	return func(ctx *gin.Context) {
		var loginRequest LoginRequest

		if err:= ctx.BindJSON(&loginRequest); err != nil{
			ctx.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return 
		}

		user,err := repository.GetUserByEmail(pool,loginRequest.Email)
		if err != nil{
			ctx.JSON(http.StatusUnauthorized,gin.H{"error":"invalid credentials"})
			return 
		}

		bcryptErr := bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(loginRequest.Password))

		if bcryptErr != nil{
			ctx.JSON(http.StatusUnauthorized,gin.H{"error":"invalid credentials"})
			return 
		}

		// map[string]interface
		// map[string]any
		claims := jwt.MapClaims{
			"user_id": user.ID,
			"email": user.Email,
			"exp": time.Now().Add(24 * time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)

		signedToken,err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to generate token"+err.Error()})
			return 
		}

		ctx.JSON(http.StatusOK,LoginResponse{Token: signedToken})

	}
}

func TestProtectedHandler() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		userID,exists := ctx.Get("user_id")

		if !exists{
			ctx.JSON(http.StatusInternalServerError,gin.H{"error":"user_id not found in the context"})
			return
		}

		ctx.JSON(http.StatusOK,gin.H{
			"message":"Protected route accessed successfully!",
			"user_id":userID,
		})
	}
}