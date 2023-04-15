package inmemoryrepository

import (
	"sync"
	"time"
	"workScheduler/internal/api/models"

	"github.com/gofrs/uuid"
)

type InMemoryRepository struct {
	Data map[string]*models.WorkItem
	Mu   *sync.Mutex
}

func (inm *InMemoryRepository) GetById(id string) (*models.WorkItem, error) {

}

func (inm *InMemoryRepository) List(from time.Time, to time.Time, zones []string) *models.WorkItem

func (inm *InMemoryRepository) Add(work *models.WorkItem) (*models.WorkItem, error) {
	inm.Mu.Lock()
	defer inm.Mu.Unlock()

	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	work.Id = uuid
	inm.Data[uuid] = work

	return work, nil
}

func (inm *InMemoryRepository) Update(work *models.WorkItem) *models.WorkItem
