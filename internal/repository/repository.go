package repository

import (
	"context"
	"time"
	"workScheduler/internal/scheduler/models"
)

type ErrorNotFound struct {
	text string
}

func (e *ErrorNotFound) Error() string {
	return e.text
}

func NewErrorNotFound(text string) *ErrorNotFound {
	return &ErrorNotFound{
		text: text,
	}
}

type ReadWriteRepository interface {
	ReadRepository
	WriteRepository
}

type ReadRepository interface {
	GetById(ctx context.Context, id string) ([]*models.WorkItem, error)
	List(ctx context.Context, from time.Time, to time.Time, zones []string, statuses []string) ([]*models.WorkItem, error)
}

type WriteRepository interface {
	Add(ctx context.Context, work *models.WorkItem) (*models.WorkItem, error)
	Update(ctx context.Context, work *models.WorkItem) (*models.WorkItem, error)
}
