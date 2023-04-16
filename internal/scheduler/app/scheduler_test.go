package app

import (
	"context"
	"fmt"
	"testing"
	"time"
	"workScheduler/internal/configuration"
	"workScheduler/internal/repository"
	"workScheduler/internal/scheduler/models"
)

type RepositoryMock struct {
	GetByIdResult *models.WorkItem
	ListResult    []*models.WorkItem
}

var _ repository.ReadRepository = (*RepositoryMock)(nil)

func (r RepositoryMock) GetById(ctx context.Context, id string) (mod *models.WorkItem, err error) {
	mod = r.GetByIdResult
	if r.GetByIdResult == nil {
		err = fmt.Errorf("test error from RepositoryMock")
	}
	return
}
func (r RepositoryMock) List(ctx context.Context, from time.Time, to time.Time, zones []string, statuses []string) (mod []*models.WorkItem, err error) {
	mod = r.ListResult
	return
}

func TestScheduleWorkSuccees(t *testing.T) {

	t.Run("succees ScheduleWork", func(t *testing.T) {

		testTime := time.Now()
		expectedInDb := models.WorkItem{
			Zones:           []string{"zone1"},
			StartDate:       testTime.Add(time.Duration(48) * time.Hour),
			DurationMinutes: 30,
			Id:              "testId",
			Priority:        "critical",
		}
		testItem := models.WorkItem{
			Zones:           []string{"zone1"},
			StartDate:       testTime.Round(time.Duration(24) * time.Hour).Add(time.Duration(7) * time.Hour),
			DurationMinutes: 30,
			Id:              "testId2",
			Priority:        "critical",
		}

		rep := RepositoryMock{
			ListResult: []*models.WorkItem{&expectedInDb},
		}
		ctx := context.Background()
		c := configuration.NewConfigurator(ctx, "../../../config.yml")
		c.Run()

		scheduler := NewScheduler(ctx, rep, c)
		result, _, err := scheduler.ScheduleWork(&testItem)
		if err != nil {
			t.Errorf("unexpected error %v, %v", err, result)
		}

	})
}
