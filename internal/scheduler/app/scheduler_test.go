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

const (
	unitTestConfigName = "../test_configs/scheduler_unit_config.yml"
)

type RepositoryMock struct {
	GetByIdResult []*models.WorkItem
	ListResult    []*models.WorkItem
}

var _ repository.ReadRepository = (*RepositoryMock)(nil)

func (r RepositoryMock) GetById(ctx context.Context, id string) (mod []*models.WorkItem, err error) {
	mod = r.GetByIdResult
	if len(r.GetByIdResult) == 0 {
		err = fmt.Errorf("test error from RepositoryMock")
	}
	return
}
func (r RepositoryMock) List(ctx context.Context, from time.Time, to time.Time, zones []string, statuses []string) (mod []*models.WorkItem, err error) {
	mod = r.ListResult
	return
}

func TestScheduleWorkSuccees(t *testing.T) {

	t.Run("succees schedule work", func(t *testing.T) {

		testTime := time.Now().Round(time.Hour * 24)
		expectedInDb := []*models.WorkItem{
			{
				Zones:           []string{"zone1"},
				StartDate:       testTime.Add(time.Duration(7) * time.Hour),
				Deadline:        testTime.Add(time.Duration(24) * time.Hour),
				DurationMinutes: 30,
				WorkId:          "testId",
				Priority:        "critical",
			},
			{
				Zones:           []string{"zone3"},
				StartDate:       testTime.Add(time.Duration(8) * time.Hour),
				Deadline:        testTime.Add(time.Duration(24) * time.Hour),
				DurationMinutes: 30,
				WorkId:          "testId",
				Priority:        "critical",
			},
		}
		testItem := models.WorkItem{
			Zones:           []string{"zone1", "zone2"},
			StartDate:       testTime.Add(time.Duration(8) * time.Hour),
			DurationMinutes: 30,
			WorkId:          "",
			Priority:        "critical",
		}

		rep := RepositoryMock{
			ListResult: expectedInDb,
		}
		ctx := context.Background()
		c := configuration.NewConfigurator(ctx, unitTestConfigName)
		c.Run()
		time.Sleep(2 * time.Second)

		scheduler := NewScheduler(ctx, rep, c)
		result, _, err := scheduler.ScheduleWork(&testItem)
		if err != nil {
			t.Errorf("unexpected error %v, %v", err, result)
		}

	})
}

func TestDublicateScheduleWorkError(t *testing.T) {

	t.Run("error duplicate schedule work", func(t *testing.T) {

		testTime := time.Now().Round(time.Hour * 24)
		expectedInDb := models.WorkItem{
			Zones:           []string{"zone1"},
			StartDate:       testTime.Add(time.Duration(12) * time.Hour),
			DurationMinutes: 50,
			WorkId:          "testId",
			Priority:        "regular",
			WorkType:        "manual",
			Deadline:        testTime.Add(time.Duration(480) * time.Hour),
		}
		testItem := expectedInDb
		testItem.WorkId = ""
		rep := RepositoryMock{
			ListResult: []*models.WorkItem{&expectedInDb},
		}
		ctx := context.Background()
		c := configuration.NewConfigurator(ctx, "../../../config.yml")
		c.Run()
		time.Sleep(2 * time.Second)

		scheduler := NewScheduler(ctx, rep, c)
		result, _, err := scheduler.ScheduleWork(&testItem)
		if err == nil {
			t.Errorf("Expect return error but got result: %v", result)
		}

	})
}

func TestProlongateWorkSuccees(t *testing.T) {

	t.Run("succees prolongate work", func(t *testing.T) {

		testTime := time.Now().Round(time.Hour * 24)
		expectedInDb := models.WorkItem{
			Zones:           []string{"zone1"},
			StartDate:       testTime.Add(time.Duration(12) * time.Hour),
			DurationMinutes: 30,
			WorkId:          "testId",
			Priority:        "regular",
			WorkType:        "manual",
			Deadline:        testTime.Add(time.Duration(480) * time.Hour),
			Status:          "in_progress",
		}
		testItem := expectedInDb
		testItem.DurationMinutes = 120
		rep := RepositoryMock{
			ListResult: []*models.WorkItem{&expectedInDb},
		}
		ctx := context.Background()
		c := configuration.NewConfigurator(ctx, "../../../config.yml")
		c.Run()
		time.Sleep(2 * time.Second)

		scheduler := NewScheduler(ctx, rep, c)
		result, _, err := scheduler.ProlongateWorkById([]*models.WorkItem{&testItem})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(result) == 0 {
			t.Errorf("Expect non-zero works count in result: %v", err)
		} else {
			if result[0].DurationMinutes != testItem.DurationMinutes {
				t.Errorf("Unexpected work DurationMinutes got: %v, want %v", result[0].DurationMinutes, testItem.DurationMinutes)
			}
		}
	})
}
