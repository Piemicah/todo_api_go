package repository

import (
	"context"
	"time"
	"todo_api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTodo(pool *pgxpool.Pool,title string,completed bool)(*models.Todo,error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx,cancel = context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()

	var query string = `
		INSERT INTO todos (title,completed)
		VALUES ($1,$2)
		RETURNING id, title, completed, created_at, updated_at
	`

	var todo models.Todo
	var err error = pool.QueryRow(ctx,query,title,completed).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil{
		return nil,err
	}
	return &todo, nil
}

func GetAllTodos(pool *pgxpool.Pool) ([]models.Todo,error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx,cancel = context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()

	var query = `
	SELECT * FROM todos
	ORDER BY created_at DESC
	`

	
	var rows, err = pool.Query(ctx,query)

	if err != nil{
		return nil, err
	}
	defer rows.Close()
	
	var todos []models.Todo = []models.Todo{}
	
	for rows.Next(){
		var todo models.Todo
		
		err = rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)

		if err != nil{
			return nil, err
		}
		todos = append(todos, todo)

	}

	if err = rows.Err(); err !=nil{
		return nil,err
	}

	return todos,nil
		
}