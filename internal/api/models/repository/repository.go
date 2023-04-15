package repository

import (
	"time"
	"workScheduler/internal/api/models"
)

type ReadRepository interface {
	GetById(id string) (*models.WorkItem, error)
	List(from time.Time, to time.Time, zones []string) *models.WorkItem
}

type WriteRepository interface {
	Add(work *models.WorkItem) (*models.WorkItem, error)
	Update(work *models.WorkItem) *models.WorkItem
}
