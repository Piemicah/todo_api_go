package repository

import (
	"context"
	"fmt"
	"time"
	"todo_api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTodo(pool *pgxpool.Pool,title string,completed bool,userID string)(*models.Todo,error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx,cancel = context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()

	var query string = `
		INSERT INTO todos (title,completed,user_id)
		VALUES ($1,$2,$3)
		RETURNING id, title, completed, created_at, updated_at, user_id
	`

	var todo models.Todo
	var err error = pool.QueryRow(ctx,query,title,completed,userID).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.UserID,
	)

	if err != nil{
		return nil,err
	}
	return &todo, nil
}

func GetAllTodos(pool *pgxpool.Pool,userID string) ([]models.Todo,error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx,cancel = context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()

	var query = `
	SELECT * FROM todos
	WHERE user_id = $1
	ORDER BY created_at DESC
	`

	
	var rows, err = pool.Query(ctx,query,userID)

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
			&todo.UserID,
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


func AdminGetAllTodos(pool *pgxpool.Pool) ([]models.Todo,error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx,cancel = context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()

	var query = `SELECT * FROM todos ORDER BY created_at DESC
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
			&todo.UserID,
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

func GetTodoById(pool *pgxpool.Pool,id int,userID string) (*models.Todo,error) {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx,cancel = context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()

	var query string = `
	SELECT * FROM todos
	WHERE id=$1 AND user_id=$2
	`
	var todo models.Todo

	
	 var err error = pool.QueryRow(ctx,query,id,userID).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.UserID,
	)

	if err != nil{
		return nil,err
	}

	return &todo,nil
}

func UpdateTodo(pool *pgxpool.Pool,id int,title string,completed bool,userID string) (*models.Todo,error){
	var ctx context.Context
	var cancel context.CancelFunc
	ctx,cancel = context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()

	var query string = `
	UPDATE todos
	SET title=$1, completed=$2, updated_at = CURRENT_TIMESTAMP
	WHERE id=$3 AND user_id=$4
	RETURNING id, title, completed, created_at, updated_at, user_id
	`
	var todo models.Todo

	
	 var err error = pool.QueryRow(ctx,query,title,completed,id,userID).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
		&todo.UserID,
	)

	if err != nil{
		return nil,err
	}

	return &todo,nil
}

func DeleteTodo(pool *pgxpool.Pool,id int,userID string) error{
	var ctx context.Context
	var cancel context.CancelFunc
	ctx,cancel = context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()

	var query string = `
	DELETE FROM todos WHERE id=$1 AND user_id=$2
	`
	tag,err := pool.Exec(ctx,query,id,userID)

	if err != nil{
		return err
	}
	if tag.RowsAffected()==0{
		return  fmt.Errorf("todo with id %d not found",id)
	}

	return nil
}