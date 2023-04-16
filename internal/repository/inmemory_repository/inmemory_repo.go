package inmemoryrepository

import (
	"context"
	"fmt"
	"sync"
	"time"
	"workScheduler/internal/repository"
	"workScheduler/internal/scheduler/models"

	"github.com/gofrs/uuid"
)

type InMemoryRepository struct {
	Data map[string]*models.WorkItem
	Mu   *sync.Mutex
}

func NewInmemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		Data: make(map[string]*models.WorkItem),
		Mu:   &sync.Mutex{},
	}
}

func inArray(arr []string, i []string) bool {
	if len(arr) == 0 {
		return true
	}
	for _, v := range arr {
		for _, v2 := range i {
			if v == v2 {
				return true
			}
		}
	}
	return false
}

func (inm *InMemoryRepository) GetById(ctx context.Context, id string) (*models.WorkItem, error) {
	inm.Mu.Lock()
	defer inm.Mu.Unlock()

	work, ok := inm.Data[id]
	if !ok {
		return nil, repository.NewErrorNotFound(fmt.Sprintf("Work with id %s not found", id))
	}

	return work, nil
}

func (inm *InMemoryRepository) List(ctx context.Context, from time.Time, to time.Time, zones []string, statuses []string) ([]*models.WorkItem, error) {
	inm.Mu.Lock()
	defer inm.Mu.Unlock()

	works := []*models.WorkItem{}

	for _, work := range inm.Data {
		if work.StartDate.Unix() >= from.Unix() && work.StartDate.Unix() < to.Unix() && inArray(zones, work.Zones) && inArray(statuses, []string{work.Status}) {
			works = append(works, work)
		}
	}
	return works, nil
}

func (inm *InMemoryRepository) Add(ctx context.Context, work *models.WorkItem) (*models.WorkItem, error) {
	inm.Mu.Lock()
	defer inm.Mu.Unlock()

	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	work.WorkId = uuid.String()
	inm.Data[uuid.String()] = work

	return work, nil
}

func (inm *InMemoryRepository) Update(ctx context.Context, work *models.WorkItem) (*models.WorkItem, error) {
	inm.Mu.Lock()
	defer inm.Mu.Unlock()

	if _, ok := inm.Data[work.WorkId]; !ok {
		return nil, repository.NewErrorNotFound(fmt.Sprintf("Work with id %s not found", work.WorkId))
	}
	inm.Data[work.WorkId] = work
	return work, nil
}
