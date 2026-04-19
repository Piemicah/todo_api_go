package handlers

import (
	"net/http"
	"todo_api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateTodoInput struct {
	Title     string `json:"title" binding:"required"`
	Completed bool   `json:"completed"`
}

func CreateTodoHandler(pool *pgxpool.Pool) gin.HandlerFunc {
return func(ctx *gin.Context) {
	var input CreateTodoInput
	if err:=ctx.ShouldBindJSON(&input); err != nil{
		ctx.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
		return 
	}

	todo,err := repository.CreateTodo(pool,input.Title,input.Completed)

	if err != nil{
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
	}
	ctx.JSON(http.StatusCreated,todo)
}
}

func GetAllTodosHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
	

	todos,err := repository.GetAllTodos(pool)

	if err != nil{
		ctx.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
	}
	ctx.JSON(http.StatusCreated,todos)
}
}