package repository

import (
	"time"
	"workScheduler/internal/api/models"
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

type ReadRepository interface {
	GetById(id string) (*models.WorkItem, error)
	List(from time.Time, to time.Time, zones []string, statuses []string) ([]*models.WorkItem, error)
}

type WriteRepository interface {
	Add(work *models.WorkItem) (*models.WorkItem, error)
	Update(work *models.WorkItem) *models.WorkItem
}
