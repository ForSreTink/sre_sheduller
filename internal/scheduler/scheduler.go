package scheduler

import (
	"context"
	"fmt"
	"time"
	api "workScheduler/internal/api/app"
	"workScheduler/internal/configuration"
	"workScheduler/internal/repository"
	"workScheduler/internal/scheduler/models"

	interval "github.com/go-follow/time-interval"
	"golang.org/x/exp/slices"
)

type Scheduler struct {
	Config     *configuration.Config
	Repository repository.ReadWriteRepository
	ctx        context.Context
}

type IntervalWork struct {
	Span *interval.Span
	Work *models.WorkItem
}

func getWorkInterval(wi *models.WorkItem) (span *interval.Span, err error) {
	workInt, err := interval.New(wi.StartDate, wi.StartDate.Add(time.Minute*time.Duration(wi.DurationMinutes)))
	span = &workInt
	return
}

func NewScheduler(ctx context.Context, repository repository.ReadWriteRepository, config *configuration.Config) (scheduler *Scheduler) {
	scheduler = &Scheduler{
		Repository: repository,
		Config:     config,
		ctx:        ctx,
	}
	return
}

func (sch *Scheduler) ScheduleWork(wi *models.WorkItem) (schedule *[]models.WorkItem, errorIsUnexpected bool, err error) {
	errorIsUnexpected = true
	from := wi.StartDate.Add(time.Minute * time.Duration(-1*Max(sch.Config.MaxWorkDurationMinutes.Automatic, sch.Config.MaxWorkDurationMinutes.Manual)))
	to := wi.StartDate.Add(24 * time.Hour * time.Duration(sch.Config.MaxDeadlineDays)).Add(time.Minute * time.Duration(-1*wi.DurationMinutes))

	// проверяем, если зона в блеклисте && работы != критичные -> 500 возвращаем полную невозможность
	for _, z := range wi.Zones {
		// проверяем, если зона в блеклисте && работы != критичные -> 500 возвращаем полную невозможность
		if slices.Contains(sch.Config.BlackList, z) && wi.Priority != string(api.Critical) {
			errorIsUnexpected = false
			err = fmt.Errorf("Zone %v is in black list, unable to shedule work with non-critical priority", z)
			return
		}
		// проверяем, если зона в вайт листе && работы не в окне -> 500 возвращаем невозможность c вариантами сдвига
		// workInt, intErr := getWorkInterval(wi)
		// if intErr != nil {
		// 	err = intErr
		// 	return
		// }
		windows, ok := sch.Config.WhiteList[z]
		if ok {
			workIntervals := []configuration.Window{}
			endDate := wi.StartDate.Add(time.Duration(wi.DurationMinutes) * time.Minute)
			if wi.DurationMinutes >= 60*24 {
				for _, window := range windows {
					if window.EndHour-window.StartHour < 24 {
						errorIsUnexpected = false
						err = fmt.Errorf("Work duration %v is longer than zone white-list windows", wi.DurationMinutes)
						return
					}
				}
			}
			if endDate.Day() > wi.StartDate.Day() {
				workIntervals = append(workIntervals, configuration.Window{StartHour: uint32(wi.StartDate.Hour()), EndHour: 24})
			}
			endHour := endDate.Hour()
			if endDate.Minute() != 0 {
				endHour++
			}
			if endHour != 0 {
				workIntervals = append(workIntervals, configuration.Window{StartHour: 0, EndHour: uint32(endHour)})
			}

			for _, workInterval := range workIntervals {
				isInWindow := false
				for _, window := range windows {
					// без учета пауз
					if workInterval.EndHour <= window.StartHour { //раньше текущего интервала
						continue
					}
					if workInterval.EndHour > window.StartHour {
						if workInterval.StartHour < window.StartHour { //наползает в начале
							break
						}
						if workInterval.StartHour >= window.StartHour && workInterval.EndHour <= window.EndHour { //целиком в текущем интервале
							isInWindow = true
							break
						}
						if workInterval.StartHour >= window.StartHour && workInterval.EndHour > window.EndHour { //наползает в конце
							break
						}
					}
					if workInterval.StartHour < window.EndHour && workInterval.EndHour > window.EndHour { //больше интервала с обеих сторон
						break
					}
					continue
				}
				if !isInWindow {
					errorIsUnexpected = false
					err = fmt.Errorf("Zone %v is in white list, but work time is not in zone white-list window", z)
					//todo - варианты сдвига
					return
				}
			}
		}
	}

	// текущее расписание
	zoneShedules, err := sch.getZoneSchedule(from, to, wi.Zones)
	if err != nil {
		return
	}
	// проверить, нет ли работ в это время, если нет ->
	//			min_avialable_zones вполняется -> 201 планируем
	//			min_avialable_zones не вполняется ->
	//					по каждой зоне в течение дня считаем варианты: для своей зоны - варианты сдвигов в рамках зоны педелах max_deadline_days, с учетом min_avialable_zones;

	return
}

func (sch *Scheduler) getZoneSchedule(from time.Time, to time.Time, zones []string) (zoneShedules map[string][]*IntervalWork, err error) {
	works, err := sch.Repository.List(sch.ctx, from, to, zones, []string{})
	if err != nil {
		return
	}
	zoneShedules = make(map[string][]*IntervalWork)
	for _, w := range works {
		for _, z := range w.Zones {
			interval, intErr := getWorkInterval(w)
			if intErr != nil {
				err = intErr
				return
			}
			iw := IntervalWork{
				Span: interval,
				Work: w,
			}
			if value, ok := zoneShedules[z]; !ok {
				zoneShedules[z] = []*IntervalWork{&iw}
			} else {
				zoneShedules[z] = append(value, &iw)
			}
		}
	}
	return

}
func Max(x int32, y int32) int32 {
	if x > y {
		return x
	} else {
		return y
	}
}

// type WorkInterval struct {
// 	Id              string
// 	Priority        int32
// 	DurationMinutes int32
// 	CompressionRate float64
// }
// func (wi WorkInterval) Weight() int64 {
// 	return int64(wi.DurationMinutes)
// }
// func (wi WorkInterval) Value() int64 {
// 	return int64(wi.Priority)
// }

// func main() {
// 	// Ввод:
// 	// Приоритет (ценность загруженные в массив v)
// 	// Веса предметов (загруженные в массив w)
// 	// Количество предметов (n)
// 	// Грузоподъемность (W)

// 	zonesCount := 3
// 	zonesSafeCount := 1
// 	deadline := 28 * 24 * 60
// 	zonesWeight := (zonesCount - zonesSafeCount) * deadline
// 	var all_works []knapsack.Packable
// 	var v []int
// 	var w []int

// 	id := 1
// 	for i := 0; i < 30; i++ {
// 		duration := rand.Intn(100)
// 		all_works = append(all_works, WorkInterval{
// 			Id:              int32(id),
// 			Priority:        int32(proirity_critical),
// 			DurationMinutes: int32(duration),
// 		})
// 		v = append(v, proirity_critical)
// 		w = append(w, duration)
// 		id++

// 		duration = rand.Intn(100)
// 		all_works = append(all_works, WorkInterval{
// 			Id:              int32(id),
// 			Priority:        int32(proirity_regular),
// 			DurationMinutes: int32(duration),
// 		})
// 		v = append(v, proirity_critical)
// 		w = append(w, duration)
// 		id++
// 	}
