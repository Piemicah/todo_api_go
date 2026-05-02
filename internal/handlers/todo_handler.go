package handlers

import (
	"net/http"
	"strconv"
	"todo_api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateTodoInput struct {
	Title     string `json:"title" binding:"required"`
	Completed bool   `json:"completed"`
}

type UpdateTodoInput struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

func CreateTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
return func(ctx *gin.Context) {
	userIDInterface, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID := userIDInterface.(string)
	
	var input CreateTodoInput
	if err:=ctx.ShouldBindJSON(&input); err != nil{
		ctx.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
		return 
	}

	todo,err := repository.CreateTodo(pool,input.Title,input.Completed,userID)

	if err != nil{
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
	}
	ctx.JSON(http.StatusCreated,todo)
}
}

func GetAllTodosHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
	userIDInterface, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID := userIDInterface.(string)

	todos,err := repository.GetAllTodos(pool,userID)

	if err != nil{
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
	}
	ctx.JSON(http.StatusCreated,todos)
}
}

func AdminGetAllTodosHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
	
	

	todos,err := repository.AdminGetAllTodos(pool)

	if err != nil{
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
	}
	ctx.JSON(http.StatusCreated,todos)
}
}

func GetTodoByIdHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return  func(ctx *gin.Context) {

		userIDInterface, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}

	userID := userIDInterface.(string)

		idStr := ctx.Param("id")
		id,err := strconv.Atoi(idStr)
		
		if err != nil{
			ctx.JSON(http.StatusBadRequest,gin.H{"error":"Invalid Todo ID"})
			return 
		}

		todo,err := repository.GetTodoById(pool,id,userID)

		if err != nil{
			if err == pgx.ErrNoRows{
				ctx.JSON(http.StatusNotFound,gin.H{"error":"Todo not found"})
				return 
			}
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
	}
	ctx.JSON(http.StatusCreated,todo)


	}
}

func UpdateTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return  func(ctx *gin.Context) {

		userIDInterface, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}

		userID := userIDInterface.(string)

		idStr := ctx.Param("id")
		id,err := strconv.Atoi(idStr)
		if err != nil{
			ctx.JSON(http.StatusBadRequest,gin.H{"error":"Invalid Todo ID"})
		}
	var input UpdateTodoInput
	if err:=ctx.ShouldBindJSON(&input); err != nil{
		ctx.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
		return 
	}

	if input.Title==nil && input.Completed==nil{
		ctx.JSON(http.StatusBadRequest,gin.H{"error":"At least one field must be provided"})
		return 
	}

	existingTodo,err := repository.GetTodoById(pool,id,userID)

	
		if err != nil{
			if err == pgx.ErrNoRows{
				ctx.JSON(http.StatusNotFound,gin.H{"error":"Todo not found"})
				return 
			}
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
		return 
	}

	title  := existingTodo.Title
	if input.Title != nil{
		title = *input.Title
	}

	completed := existingTodo.Completed

	if input.Completed != nil{
		completed = *input.Completed
	}

		todo,err := repository.UpdateTodo(pool,id,title,completed,userID)

		if err != nil{
			if err == pgx.ErrNoRows{
				ctx.JSON(http.StatusNotFound,gin.H{"error":"Todo not found"})
				return 
			}
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
		return 
	}
	ctx.JSON(http.StatusOK,todo)

	}
}

func DeleteTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc{
	return func(ctx *gin.Context) {

		userIDInterface, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
			return
		}

		userID := userIDInterface.(string)

		idStr := ctx.Param("id")
		id,err := strconv.Atoi(idStr)
		
		if err != nil{
			ctx.JSON(http.StatusBadRequest,gin.H{"error":"Invalid Todo ID"})
			return 
		}
		
		err = repository.DeleteTodo(pool,id,userID)
		
		if err != nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return 
		}

		ctx.JSON(http.StatusOK,gin.H{"status":"successfully deleted!"})

	}
}