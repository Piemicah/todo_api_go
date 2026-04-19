package models

import "time"

type Todo struct {
	ID          int    `json:"id" db:"id"`
	Title       string `json:"title" db:"title"`
	Completed   bool   `json:"completed" db:"completed"`
	CreatedAt time.Time `json:"completed_at" db:"completed_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}