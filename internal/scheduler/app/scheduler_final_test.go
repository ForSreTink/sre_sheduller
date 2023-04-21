package app

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"
	"workScheduler/internal/configuration"
	"workScheduler/internal/scheduler/models"
)

const (
	finalTestConfigName = "../test_configs/scheduler_final_config.yml"
)

type TestEvent struct {
	Name             string
	ExpectedInDb     []*models.WorkItem
	NewWork          *models.WorkItem
	Action           string
	ActionTime       time.Time
	UserMustApprove  bool
	ExpectedVariants []*models.WorkItem
	ConfigChange     func(*configuration.Config)
}

func TestScheduleEvents(t *testing.T) {

	testEvents := []TestEvent{
		{
			Name:       "1. Заявка на проведение работ 1, 18 апреля 8:00",
			ActionTime: time.Date(2023, 04, 18, 8, 0, 0, 0, time.UTC), //18 апреля	8:00
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:           []string{"Zone_2", "Zone_4"},
				StartDate:       time.Date(2023, 04, 19, 0, 0, 0, 0, time.UTC), //19 апреля	00:00
				DurationMinutes: 60,
				Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
				WorkId:          "1",
				Priority:        "regular",
				WorkType:        "manual",
			},
		},
		{
			Name:       "2. Заявка на проведение работ 2, 18 апреля 8:20",
			ActionTime: time.Date(2023, 04, 18, 8, 20, 0, 0, time.UTC), //18 апреля	8:20
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:           []string{"Zone_1", "Zone_2"},
				StartDate:       time.Date(2023, 04, 19, 1, 0, 0, 0, time.UTC), //19 апреля	01:00
				DurationMinutes: 120,
				Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
				WorkId:          "2",
				Priority:        "regular",
				WorkType:        "manual",
			},
		},
		{
			Name:       "3. Заявка на проведение работ 3, 18 апреля 10:00",
			ActionTime: time.Date(2023, 04, 18, 10, 0, 0, 0, time.UTC), //18 апреля	10:00
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:           []string{"Zone_1", "Zone_4"},
				StartDate:       time.Date(2023, 04, 19, 0, 0, 0, 0, time.UTC), //19 апреля	00:00
				DurationMinutes: 30,
				Deadline:        time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 05:00
				WorkId:          "3",
				Priority:        "regular",
				WorkType:        "automatic",
			},
			UserMustApprove: true, // проводить работы в Zone_1 в 00:00 не разрешено - предлагаем боту слот в 3:00, на который он соглашается
			ExpectedVariants: []*models.WorkItem{
				{
					Zones:           []string{"Zone_1", "Zone_4"},
					StartDate:       time.Date(2023, 04, 19, 3, 0, 0, 0, time.UTC), //19 апреля	00:00
					DurationMinutes: 30,
					Deadline:        time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 05:00
					WorkId:          "3",
					Priority:        "regular",
					WorkType:        "automatic",
				},
			},
		},
		{
			Name:       "4. Заявка на проведение работ 4, 18 апреля 11:59",
			ActionTime: time.Date(2023, 04, 18, 11, 59, 0, 0, time.UTC), //18 апреля 11:59
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:           []string{"Zone_1", "Zone_2", "Zone_3", "Zone_4"},
				StartDate:       time.Date(2023, 04, 19, 1, 0, 0, 0, time.UTC), //19 апреля	01:00
				DurationMinutes: 30,
				Deadline:        time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 05:00
				WorkId:          "4",
				Priority:        "regular",
				WorkType:        "automatic",
			},
			UserMustApprove: true, //конфликт с работами с ID 2, плюс в Zone_4 в это время уже ничего нельзя проводить
			ExpectedVariants: []*models.WorkItem{ //предлагаем два свободных слота в 03:30 и 04:00, т.к. занять все зоны доступности сразу нельзя - бот соглашается
				{
					Zones:           []string{"Zone_1", "Zone_3"},
					StartDate:       time.Date(2023, 04, 19, 3, 30, 0, 0, time.UTC), //19 апреля	01:00
					DurationMinutes: 30,
					Deadline:        time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 05:00
					WorkId:          "4",
					Priority:        "regular",
					WorkType:        "automatic",
				},
				{
					Zones:           []string{"Zone_2", "Zone_4"},
					StartDate:       time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля	04:00
					DurationMinutes: 30,
					Deadline:        time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 05:00
					WorkId:          "4",
					Priority:        "regular",
					WorkType:        "automatic",
				},
			},
		},
		{
			Name:       "5. Заявка на проведение работ 5, 18 апреля 14:00",
			ActionTime: time.Date(2023, 04, 18, 14, 00, 0, 0, time.UTC), //18 апреля 14:00
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:           []string{"Zone_3", "Zone_4"},
				StartDate:       time.Date(2023, 04, 19, 2, 0, 0, 0, time.UTC), //19 апреля	02:00
				DurationMinutes: 90,
				Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
				WorkId:          "4",
				Priority:        "regular",
				WorkType:        "manual",
			},
		},
		{
			Name:       "6. Перенос работ 2, 18 апреля 15:00",
			ActionTime: time.Date(2023, 04, 18, 11, 59, 0, 0, time.UTC), //18 апреля 15:00
			Action:     "move",
			NewWork: &models.WorkItem{
				Zones:           []string{"Zone_1", "Zone_2"},
				StartDate:       time.Date(2023, 04, 19, 2, 0, 0, 0, time.UTC), //19 апреля	02:00
				DurationMinutes: 120,
				Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
				WorkId:          "2",
				Priority:        "regular",
				WorkType:        "manual",
			},
		},
		{
			Name:       "7. Заявка на проведение работ 7, 18 апреля 19:00",
			ActionTime: time.Date(2023, 04, 18, 19, 00, 0, 0, time.UTC), //18 апреля 19:00
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:           []string{"Zone_1", "Zone_2"},
				StartDate:       time.Date(2023, 04, 19, 2, 30, 0, 0, time.UTC), //19 апреля 02:30
				DurationMinutes: 60,
				Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
				WorkId:          "7",
				Priority:        "regular",
				WorkType:        "manual",
			},
		},
		{
			Name:       "8. Изменение конфига, 18 апреля 19:00",
			ActionTime: time.Date(2023, 04, 18, 22, 50, 0, 0, time.UTC), //18 апреля 22:50
			Action:     "config change",
			ConfigChange: func(c *configuration.Config) {
				c.WhiteList["Zone_1"][0].StartHour = 0
				c.WhiteList["Zone_1"] = append(c.WhiteList["Zone_1"], configuration.Window{StartHour: 23, EndHour: 24})
				c.WhiteList["Zone_3"][0].StartHour = 0
				c.WhiteList["Zone_3"] = append(c.WhiteList["Zone_3"], configuration.Window{StartHour: 23, EndHour: 24})
			},
		},
		{
			Name:       "9. Заявка на проведение работ 9, 18 апреля 23:00",
			ActionTime: time.Date(2023, 04, 18, 23, 00, 0, 0, time.UTC), //18 апреля 23:00
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:           []string{"Zone_1", "Zone_3"},
				StartDate:       time.Date(2023, 04, 18, 23, 30, 0, 0, time.UTC), //18 апреля 23:30
				DurationMinutes: 60,
				Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
				WorkId:          "9",
				Priority:        "regular",
				WorkType:        "manual",
			},
			UserMustApprove: true, //ожидаем разделение работ, т.к. есть пересечения с 1 и не выполняется min available zone
			ExpectedVariants: []*models.WorkItem{ //предлагаем два свободных слота в 03:30 и 04:00, т.к. занять все зоны доступности сразу нельзя - бот соглашается
				{
					Zones:           []string{"Zone_1"},
					StartDate:       time.Date(2023, 04, 18, 23, 30, 0, 0, time.UTC), //18 апреля 23:30
					DurationMinutes: 60,
					Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
					WorkId:          "9",
					Priority:        "regular",
					WorkType:        "manual",
				},
				{
					Zones:           []string{"Zone_3"},
					StartDate:       time.Date(2023, 04, 19, 0, 30, 0, 0, time.UTC), //18 апреля 23:30
					DurationMinutes: 60,
					Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
					WorkId:          "9",
					Priority:        "regular",
					WorkType:        "manual",
				},
			},
		},
		{
			Name:       "10. Продление работ 1, 19 апреля 00:47",
			ActionTime: time.Date(2023, 04, 18, 23, 00, 0, 0, time.UTC), //18 апреля 00:47
			Action:     "prolongate",
			NewWork: &models.WorkItem{
				Zones:           []string{"Zone_2", "Zone_4"},
				StartDate:       time.Date(2023, 04, 19, 0, 0, 0, 0, time.UTC), //19 апреля	00:00
				DurationMinutes: 120,
				Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
				WorkId:          "1",
				Priority:        "regular",
				WorkType:        "manual",
			},
			UserMustApprove: false, // ожидаем отмену работ 2. Не вернется UserMustApprove, потому что от пользователя не требуется подтверждений
			ExpectedVariants: []*models.WorkItem{
				{
					Zones:           []string{"Zone_2", "Zone_4"},
					StartDate:       time.Date(2023, 04, 19, 0, 0, 0, 0, time.UTC), //19 апреля	00:00
					DurationMinutes: 120,
					Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
					WorkId:          "1",
					Priority:        "regular",
					WorkType:        "manual",
				},
				{
					Zones:           []string{"Zone_1", "Zone_2"},
					StartDate:       time.Date(2023, 04, 19, 1, 0, 0, 0, time.UTC), //19 апреля	01:00
					DurationMinutes: 120,
					Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
					WorkId:          "2",
					Priority:        "regular",
					WorkType:        "manual",
					Status:          "cancelled",
				},
			},
		},
	}

	ctx := context.Background()
	c := configuration.NewConfigurator(ctx, finalTestConfigName)
	c.Run()
	time.Sleep(2 * time.Second)

	for i, e := range testEvents {
		t.Run(e.Name, func(t *testing.T) {

			rep := RepositoryMock{
				ListResult: e.ExpectedInDb,
			}
			scheduler := NewScheduler(ctx, rep, c)
			var result []*models.WorkItem
			//var errorIsUnexpected bool
			var err error
			var userMustApprove bool
			switch e.Action {
			case "add":
				result, userMustApprove, err = scheduler.ScheduleWork(e.NewWork)
			case "move":
				result, userMustApprove, err = scheduler.MoveWork(e.NewWork)
			case "prolongate":
				result, userMustApprove, err = scheduler.MoveWork(e.NewWork)
			case "config change":
				e.ConfigChange(c.Data)
			default:
				t.Fatalf("%s: unexpected event action\n", e.Name)
			}

			if !e.UserMustApprove && err != nil {
				t.Errorf("%s: unexpected error %v, %v\n", e.Name, err, result)
			}
			if e.UserMustApprove {
				if !userMustApprove {
					t.Errorf("%s: unexpected error %v, %v\n", e.Name, err, result)
				} else {
					fmt.Printf("%s: event processing return expected error %v\n", e.Name, err)
				}
				if len(e.ExpectedVariants) > 0 {
					if len(e.ExpectedVariants) != len(result) {
						t.Errorf("%s: event processing return unexpected variants count: %v, expected %v\n", e.Name, len(result), len(e.ExpectedVariants))
					} else {
						// sort results and expectations by start date
						sort.Slice(result, func(x, y int) bool {
							return result[x].StartDate.Before(result[y].StartDate)
						})
						sort.Slice(e.ExpectedVariants, func(x, y int) bool {
							return e.ExpectedVariants[x].StartDate.Before(e.ExpectedVariants[y].StartDate)
						})
						// check results
						for i, res := range result {
							CompareWorkItems(t, e.Name, res, e.ExpectedVariants[i])
						}
						fmt.Printf("%s: event processing return variants %v\n", e.Name, result)
					}

				}
			}
			if err == nil {
				if len(result) == 0 {
					t.Errorf("%s: event processing return unexpected results count: %v, expected >= 1\n", e.Name, len(result))
				} else {
					// successfully planned 1 or planned 1 + movement of others
					CompareWorkItems(t, e.Name, result[0], e.NewWork)
				}
			}

			fmt.Printf("%s: event processed successfully\n", e.Name)
			if i < len(testEvents)-1 {
				testEvents[i+1].ExpectedInDb = append(e.ExpectedInDb, result...)
			}
		})
	}
}

func CompareWorkItems(t *testing.T, testName string, result *models.WorkItem, expected *models.WorkItem) {
	if result.Id != expected.Id {
		t.Errorf("%s: unexpected result Id: %v, expected %v\n", testName, result, expected)
	}
	if result.StartDate != expected.StartDate {
		t.Errorf("%s: unexpected result StartDate: %v, expected %v\n", testName, result, expected)
	}

	if len(expected.Zones) != len(result.Zones) {
		t.Errorf("%s: unexpected zones count in result: %v, expected %v\n", testName, result, expected)
	} else {
		sort.Slice(result.Zones, func(x, y int) bool {
			return result.Zones[x] < result.Zones[y]
		})
		sort.Slice(expected.Zones, func(x, y int) bool {
			return expected.Zones[x] < expected.Zones[y]
		})

		for j, zone := range result.Zones {
			if zone != expected.Zones[j] {
				t.Errorf("%s: unexpected zone names in result: %v, expected %v\n", testName, result, expected)
				break
			}
		}
	}
}
