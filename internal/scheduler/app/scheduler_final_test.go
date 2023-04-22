package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
	"workScheduler/internal/configuration"
	"workScheduler/internal/scheduler/models"
)

const (
	finalTestConfigName = "../test_configs/scheduler_final_config.yml"
)

func TestScheduleFinalEvents(t *testing.T) {

	testEvents := []TestEvent{
		{
			Name:       "1. Заявка на проведение работ 1, 18 апреля 8:00",
			ActionTime: time.Date(2023, 04, 18, 8, 0, 0, 0, time.UTC), //18 апреля	8:00
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:              []string{"Zone_2", "Zone_4"},
				StartDate:          time.Date(2023, 04, 19, 0, 0, 0, 0, time.UTC), //19 апреля	00:00
				DurationMinutes:    60,
				Deadline:           time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
				WorkId:             "1",
				Priority:           "regular",
				WorkType:           "automatic",
				MaxCompressionRate: 0.50,
			},
		},
		{
			Name:       "2. Заявка на проведение работ 2, 18 апреля 8:20",
			ActionTime: time.Date(2023, 04, 18, 8, 20, 0, 0, time.UTC), //18 апреля	8:20
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:              []string{"Zone_1", "Zone_2"},
				StartDate:          time.Date(2023, 04, 19, 1, 0, 0, 0, time.UTC), //19 апреля	01:00
				DurationMinutes:    30,
				Deadline:           time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
				WorkId:             "2",
				Priority:           "regular",
				WorkType:           "automatic",
				MaxCompressionRate: 0,
			},
		},
		{
			Name:       "3. Заявка на проведение работ 3, 18 апреля 10:00",
			ActionTime: time.Date(2023, 04, 18, 10, 0, 0, 0, time.UTC), //18 апреля	10:00
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:              []string{"Zone_1", "Zone_4"},
				StartDate:          time.Date(2023, 04, 19, 0, 0, 0, 0, time.UTC), //19 апреля	00:00
				DurationMinutes:    30,
				Deadline:           time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 05:00
				WorkId:             "3",
				Priority:           "regular",
				WorkType:           "automatic",
				MaxCompressionRate: 0,
			},
			UserMustApprove: "true", // проводить работы в Zone_1 в 00:00 не разрешено - предлагаем боту слот в 3:00, на который он соглашается
			ExpectedVariants: []*models.WorkItem{
				{
					Zones:              []string{"Zone_1", "Zone_4"},
					StartDate:          time.Date(2023, 04, 19, 3, 0, 0, 0, time.UTC), //19 апреля	00:00
					DurationMinutes:    30,
					Deadline:           time.Date(2023, 04, 19, 5, 30, 0, 0, time.UTC), //19 апреля 05:30
					WorkId:             "3",
					Priority:           "regular",
					WorkType:           "automatic",
					MaxCompressionRate: 0,
				},
			},
		},
		{
			Name:       "4. Заявка на проведение работ 4, 18 апреля 11:59",
			ActionTime: time.Date(2023, 04, 18, 11, 59, 0, 0, time.UTC), //18 апреля 11:59
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:              []string{"Zone_1", "Zone_2", "Zone_3", "Zone_4"},
				StartDate:          time.Date(2023, 04, 19, 1, 0, 0, 0, time.UTC), //19 апреля	01:00
				DurationMinutes:    30,
				Deadline:           time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 04:00
				WorkId:             "4",
				Priority:           "regular",
				WorkType:           "automatic",
				MaxCompressionRate: 0,
			},
			UserMustApprove: "true", //конфликт с работами с ID 2, плюс в Zone_4 в это время уже ничего нельзя проводить
			ExpectedVariants: []*models.WorkItem{ //тоже нормальные варианты
				{
					Zones:           []string{"Zone_1", "Zone_2", "Zone_3"},
					StartDate:       time.Date(2023, 04, 19, 3, 30, 0, 0, time.UTC), //19 апреля 01:00
					DurationMinutes: 30,
					Deadline:        time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 05:00
					WorkId:          "4",
					Priority:        "regular",
					WorkType:        "automatic",
					Status:          "planned",
				},
				{
					Zones:           []string{"Zone_4"},
					StartDate:       time.Date(2023, 04, 19, 1, 0, 0, 0, time.UTC), //19 апреля	04:00
					DurationMinutes: 30,
					Deadline:        time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 05:00
					WorkId:          "4",
					Priority:        "regular",
					WorkType:        "automatic",
					Status:          "planned",
				},
			},
			AppendToExpectedInDb: []*models.WorkItem{ //предлагаем два свободных слота в 03:30 и 04:00, т.к. занять все зоны доступности сразу нельзя - бот соглашается
				{
					Zones:           []string{"Zone_1", "Zone_3"},
					StartDate:       time.Date(2023, 04, 19, 3, 30, 0, 0, time.UTC), //19 апреля	01:00
					DurationMinutes: 30,
					Deadline:        time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 05:00
					WorkId:          "4",
					Priority:        "regular",
					WorkType:        "automatic",
					Status:          "planned",
				},
				{
					Zones:           []string{"Zone_2", "Zone_4"},
					StartDate:       time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля	04:00
					DurationMinutes: 30,
					Deadline:        time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 05:00
					WorkId:          "4",
					Priority:        "regular",
					WorkType:        "automatic",
					Status:          "planned",
				},
			},
		},
		{
			Name:       "5. Заявка на проведение работ 5, 18 апреля 14:00",
			ActionTime: time.Date(2023, 04, 18, 14, 00, 0, 0, time.UTC), //18 апреля 14:00
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:              []string{"Zone_3", "Zone_4"},
				StartDate:          time.Date(2023, 04, 19, 2, 0, 0, 0, time.UTC), //19 апреля	02:00
				DurationMinutes:    90,
				Deadline:           time.Date(2023, 04, 19, 4, 30, 0, 0, time.UTC), //19 апреля 04:30
				WorkId:             "5",
				Priority:           "regular",
				WorkType:           "automatic",
				MaxCompressionRate: 0.5,
			},
			UserMustApprove: "true", //конфликт с работами с ID 3 и 4, нужен сдвиг
			ExpectedVariants: []*models.WorkItem{ //тоже нормальные варианты
				{
					Zones:           []string{"Zone_3"},
					StartDate:       time.Date(2023, 04, 19, 2, 0, 0, 0, time.UTC), //19 апреля	02:00
					DurationMinutes: 90,
					Deadline:        time.Date(2023, 04, 19, 4, 30, 0, 0, time.UTC), //19 апреля 04:30
					WorkId:          "5",
					Priority:        "regular",
					WorkType:        "automatic",
				},
				{
					Zones:           []string{"Zone_4"},
					StartDate:       time.Date(2023, 04, 19, 3, 0, 0, 0, time.UTC), //19 апреля	02:00
					DurationMinutes: 90,
					Deadline:        time.Date(2023, 04, 19, 4, 30, 0, 0, time.UTC), //19 апреля 04:30
					WorkId:          "5",
					Priority:        "regular",
					WorkType:        "automatic",
				},
				{
					Zones:           []string{"Zone_4"},
					StartDate:       time.Date(2023, 04, 19, 4, 30, 0, 0, time.UTC), //19 апреля 03:30 -> 04:30
					DurationMinutes: 30,
					Deadline:        time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 05:00
					WorkId:          "4",
					Priority:        "regular",
					WorkType:        "automatic",
				},
				{
					Zones:           []string{"Zone_4"},
					StartDate:       time.Date(2023, 04, 19, 5, 0, 0, 0, time.UTC), //19 апреля 03:00 -> 05:00
					DurationMinutes: 30,
					Deadline:        time.Date(2023, 04, 19, 5, 30, 0, 0, time.UTC), //19 апреля 05:30
					WorkId:          "3",
					Priority:        "regular",
					WorkType:        "automatic",
				},
			},
		},
		{
			//Эту заявку система будет обязана отклонить, т.к. слотов для таких длинных работ не остается из-за конфликта с ID 5, которые были согласованы ранее переноса!
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
				WorkType:        "automatic",
			},
			ExpectedErrorContains: "unable to schedule work",
		},
		{
			// Эта заявка также не может быть удовлетворена из-за конфликта, слоты в 4:30 и 5:00 - занимать сразу в двух зонах нельзя,
			// передвигать 4 и 3 на более позднее время - тоже нельзя из-за дедлайна
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
			ExpectedErrorContains: "unable to schedule work",
		},
		{
			// ??? Нельзя провести работы 3 раньше - хоть расписание теперь это позволяет, алгоритм должен быть
			// устойчивым и должен стараться вносить минимальные изменения в расписание!
			Name:       "8. Изменение конфига, 18 апреля 19:00",
			ActionTime: time.Date(2023, 04, 18, 22, 50, 0, 0, time.UTC), //18 апреля 22:50
			Action:     "config change",
			ConfigChange: func(c *configuration.Config) {
				c.BlackList = append(c.BlackList, "Zone_3")
			},
		},
		{
			// За счет предыдущего изменения конфига можем предложить последовательное проведение работы в двух разных зонах доступности - клиент соглашается.
			Name:       "9. Заявка на проведение работ 9, 18 апреля 23:00",
			ActionTime: time.Date(2023, 04, 18, 23, 00, 0, 0, time.UTC), //18 апреля 23:00
			Action:     "add",
			NewWork: &models.WorkItem{
				Zones:              []string{"Zone_1", "Zone_3"},
				StartDate:          time.Date(2023, 04, 18, 23, 30, 0, 0, time.UTC), //18 апреля 23:30
				DurationMinutes:    60,
				Deadline:           time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
				WorkId:             "9",
				Priority:           "regular",
				WorkType:           "automatic",
				MaxCompressionRate: 0.5,
			},
			UserMustApprove: "true", //ожидаем разделение работ, т.к. есть пересечения с 1 и не выполняется min available zone
			ExpectedVariants: []*models.WorkItem{ //предлагаем два свободных слота в 03:30 и 04:00, т.к. занять все зоны доступности сразу нельзя - бот соглашается
				{
					Zones:           []string{"Zone_1"},
					StartDate:       time.Date(2023, 04, 18, 23, 30, 0, 0, time.UTC), //18 апреля 23:30
					DurationMinutes: 60,
					Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
					WorkId:          "9",
					Priority:        "regular",
					WorkType:        "automatic",
				},
				{
					Zones:           []string{"Zone_3"},
					StartDate:       time.Date(2023, 04, 19, 0, 30, 0, 0, time.UTC), //18 апреля 23:30
					DurationMinutes: 60,
					Deadline:        time.Date(2023, 04, 19, 4, 0, 0, 0, time.UTC), //19 апреля 04:00
					WorkId:          "9",
					Priority:        "regular",
					WorkType:        "automatic",
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
			if e.NewWork != nil {
				e.NewWork.InitialDuration = e.NewWork.DurationMinutes
				e.NewWork.InitialStartDate = e.NewWork.StartDate
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
				result, userMustApprove, err = scheduler.MoveWork([]*models.WorkItem{e.NewWork})
			case "prolongate":
				result, userMustApprove, err = scheduler.MoveWork([]*models.WorkItem{e.NewWork})
			case "config change":
				e.ConfigChange(c.Data)
			default:
				t.Fatalf("%s: unexpected event action\n", e.Name)
			}

			if e.UserMustApprove == "false" && err != nil && e.ExpectedErrorContains == "" {
				t.Errorf("%s: unexpected error %v, result: %v\n", e.Name, err, result)
			}

			if e.ExpectedErrorContains != "" {
				if err == nil {
					t.Errorf("%s: expected error with message [%v], got no errors \n", e.Name, e.ExpectedErrorContains)
				}
				if err != nil && !strings.Contains(err.Error(), e.ExpectedErrorContains) {
					t.Errorf("%s: expected error with message [%v], got [%v]\n", e.Name, e.ExpectedErrorContains, err.Error())
				}
				fmt.Printf("%s: event processing return expected error [%v]\n", e.Name, err)
			}

			if e.UserMustApprove == "true" {
				if !userMustApprove {
					t.Errorf("%s: expected userMustApprove %v, err: %v\n", e.Name, userMustApprove, err)
				} else {
					fmt.Printf("%s: event processing return userMustApprove\n", e.Name)
				}
				if len(e.ExpectedVariants) > 0 {
					// fmt.Printf("%s: event processing return variants:  %v", e.Name, result)
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
						// fmt.Printf("%s: event processing return variants %v\n", e.Name, result)
					}

				}
			}
			if err == nil {
				if e.Action != "config change" {
					if len(result) == 0 {
						t.Errorf("%s: event processing return unexpected results count: %v, expected >= 1\n", e.Name, len(result))
					} else if len(e.ExpectedVariants) > 0 {
						// successfully planned 1 or planned 1 + movement of others
						CompareWorkItems(t, e.Name, result[0], e.ExpectedVariants[0])
					}
				}
			}

			if t.Failed() {
				fmt.Printf("%s: event processed with test errors\n", e.Name)
			} else {
				fmt.Printf("%s: event processed successfully\n", e.Name)
			}

			if i < len(testEvents)-1 {
				if len(testEvents[i+1].ExpectedInDb) == 0 {
					if len(e.AppendToExpectedInDb) > 0 {
						testEvents[i+1].ExpectedInDb = append(e.ExpectedInDb, e.AppendToExpectedInDb...)
					} else {
						testEvents[i+1].ExpectedInDb = append(e.ExpectedInDb, result...)
					}
				}

			}
		})
	}
}
