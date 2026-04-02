package model

import (
	"time"

	"github.com/google/uuid"
)

/*
ENUM(

	todo
	in_progress
	complete

)
*/
type TaskStatus string

type Task struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	DueDate     time.Time  `json:"dueDate"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type CreateTaskRequest struct {
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	DueDate     time.Time  `json:"dueDate"`
}

type UpdateStatusRequest struct {
	Status TaskStatus `json:"status"`
}
