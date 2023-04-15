package inmemoryrepository

import (
	"fmt"
	"sync"
	"time"
	"workScheduler/internal/api/models"
	"workScheduler/internal/api/models/repository"

	"github.com/gofrs/uuid"
)

type InMemoryRepository struct {
	Data map[string]*models.WorkItem
	Mu   *sync.Mutex
}

func inArray(arr []string, i string) bool {
	for _, v := range arr {
		if i == v {
			return true
		}
	}
	return false
}

func (inm *InMemoryRepository) GetById(id string) (*models.WorkItem, error) {
	inm.Mu.Lock()
	defer inm.Mu.Unlock()

	work, ok := inm.Data[id]
	if !ok {
		return nil, repository.NewErrorNotFound(fmt.Sprintf("Work with id %s not found", id))
	}

	return work, nil
}

func (inm *InMemoryRepository) List(from time.Time, to time.Time, zones []string) ([]*models.WorkItem, error) {
	inm.Mu.Lock()
	defer inm.Mu.Unlock()

	works := []*models.WorkItem{}

	if len(zones) == 0 {
		for _, work := range inm.Data {
			if work.StartDate.Unix() >= from.Unix() && work.StartDate.Unix() < to.Unix() {
				works = append(works, work)
			}
		}
	} else {
		for _, work := range inm.Data {
			if work.StartDate.Unix() >= from.Unix() && work.StartDate.Unix() < to.Unix() && inArray(zones, work.Zone) {
				works = append(works, work)
			}
		}
	}
	return works, nil
}

func (inm *InMemoryRepository) Add(work *models.WorkItem) (*models.WorkItem, error) {
	inm.Mu.Lock()
	defer inm.Mu.Unlock()

	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	work.Id = uuid.String()
	inm.Data[uuid.String()] = work

	return work, nil
}

func (inm *InMemoryRepository) Update(work *models.WorkItem) (*models.WorkItem, error) {
	inm.Mu.Lock()
	defer inm.Mu.Unlock()

	if _, ok := inm.Data[work.Id]; !ok {
		return nil, repository.NewErrorNotFound(fmt.Sprintf("Work with id %s not found", work.Id))
	}
	inm.Data[work.Id] = work
	return work, nil
}
