package app

import (
	"context"
	"fmt"
	"sort"
	"time"

	"workScheduler/internal/configuration"
	"workScheduler/internal/repository"
	"workScheduler/internal/scheduler/models"

	interval "github.com/go-follow/time-interval"
	"golang.org/x/exp/slices"
)

const (
	PriorityCritical  = "critical"
	PriorityRegular   = "regular"
	WorkTypeAutomatic = "automatic"
	WorkTypeManual    = "manual"
	StatusInProgress  = "in_progress"
	StatusPlanned     = "planned"
	StatusCancelled   = "cancelled"
)

type Scheduler struct {
	Config     *configuration.Config
	Repository repository.ReadRepository
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

func NewScheduler(ctx context.Context, repository repository.ReadRepository, config *configuration.Configurator) (scheduler *Scheduler) {
	scheduler = &Scheduler{
		Repository: repository,
		Config:     config.Data,
		ctx:        ctx,
	}
	return
}

func (sch *Scheduler) MoveWork(wi *models.WorkItem) (schedule []*models.WorkItem, errorIsUnexpected bool, err error) {
	errorIsUnexpected = true
	statuses := []string{StatusPlanned, StatusInProgress}
	from := wi.StartDate.Add(time.Minute * time.Duration(-1*Max(sch.Config.MaxWorkDurationMinutes.Automatic, sch.Config.MaxWorkDurationMinutes.Manual)))
	to := wi.StartDate.Add(24 * time.Hour * time.Duration(sch.Config.MaxDeadlineDays)).Add(time.Minute * time.Duration(-1*wi.DurationMinutes))

	// текущее расписание
	zoneSchedule, err := sch.getZoneSchedule(from, to, wi.Zones, statuses)
	if err != nil {
		return
	}
	// not inclide current saved wi in schedule on move
	for z, sched := range zoneSchedule {
		for i, w := range sched {
			if w.Work.Id == wi.Id {
				sched[i] = sched[len(sched)-1]
				zoneSchedule[z] = sched[:len(sched)-1]
				break
			}
		}
	}
	//Критичные работы помещать в расписание вне очереди, принудительно отменяя обычные ручные работы и сжимая или перенося работы обычные автоматические.
	//При получении заявок на ручные работы отдавать им приоритет, отменяя работы автоматического типа.
	schedule, zoneErr := sch.chekScheduleChange(zoneSchedule, wi, true, (wi.WorkType == WorkTypeManual), (wi.Priority == PriorityCritical))
	if zoneErr != nil {
		errorIsUnexpected = false
		err = zoneErr
	}
	return
}

func (sch *Scheduler) ScheduleWork(wi *models.WorkItem) (schedule []*models.WorkItem, errorIsUnexpected bool, err error) {
	errorIsUnexpected = true
	statuses := []string{StatusPlanned, StatusInProgress}
	from := wi.StartDate.Add(time.Minute * time.Duration(-1*Max(sch.Config.MaxWorkDurationMinutes.Automatic, sch.Config.MaxWorkDurationMinutes.Manual)))
	to := wi.StartDate.Add(24 * time.Hour * time.Duration(sch.Config.MaxDeadlineDays)).Add(time.Minute * time.Duration(-1*wi.DurationMinutes))

	// текущее расписание
	zoneSchedule, err := sch.getZoneSchedule(from, to, wi.Zones, statuses)
	if err != nil {
		return
	}
	//Критичные работы помещать в расписание вне очереди, принудительно отменяя обычные ручные работы и сжимая или перенося работы обычные автоматические.
	//При получении заявок на ручные работы отдавать им приоритет, отменяя работы автоматического типа.
	schedule, zoneErr := sch.chekScheduleChange(zoneSchedule, wi, true, (wi.WorkType == WorkTypeManual), (wi.Priority == PriorityCritical))
	if zoneErr != nil {
		errorIsUnexpected = false
		err = zoneErr
	}
	return
}

func (sch *Scheduler) ProlongateWorkById(wi *models.WorkItem) (schedule []*models.WorkItem, errorIsUnexpected bool, err error) {
	errorIsUnexpected = true
	statuses := []string{StatusPlanned, StatusInProgress}
	from := wi.StartDate.Add(time.Minute * time.Duration(-1*Max(sch.Config.MaxWorkDurationMinutes.Automatic, sch.Config.MaxWorkDurationMinutes.Manual)))
	to := wi.StartDate.Add(24 * time.Hour * time.Duration(sch.Config.MaxDeadlineDays)).Add(time.Minute * time.Duration(-1*wi.DurationMinutes))

	// текущее расписание
	zoneSchedule, err := sch.getZoneSchedule(from, to, wi.Zones, statuses)
	if err != nil {
		return
	}
	newSchedule, zoneErr := sch.chekScheduleChange(zoneSchedule, wi, true, true, false)
	if zoneErr != nil {
		errorIsUnexpected = false
		err = zoneErr
	}
	schedule = append(schedule, wi)
	schedule = append(schedule, newSchedule...)
	return
}

func (sch *Scheduler) chekScheduleChange(zoneSchedule map[string][]*IntervalWork, wi *models.WorkItem, move bool, cancelAuto bool, cancelManual bool) (schedule []*models.WorkItem, err error) {
	hasFreeWindow := false
	workItemInterval, err := getWorkInterval(wi)
	if err != nil {
		return
	}

	scheduleRyZone := make(map[string][]*models.WorkItem)
	for _, z := range wi.Zones {
		// проверить black и white листы
		_, zoneErr := sch.checkZoneLists(z, wi)
		if zoneErr != nil {
			err = zoneErr
			return
		}
		// если пришло изменение работ, то надо удалить ранее сущестовавший айтем из списка сравнения
		tmp := zoneSchedule[z]
		for i, s := range zoneSchedule[z] {
			if s.Work.Id == wi.Id {
				tmp[i] = tmp[len(tmp)-1]
				tmp = tmp[:len(tmp)-1]
			}
		}
		zoneSchedule[z] = tmp
		// проверить, нет ли работ в это время, если нет -> проверить min_avialable_zones
		zoneScheduleByZone, ok := zoneSchedule[z]
		if ok {
			hasFreeWindow = checkZoneAvailabe(zoneScheduleByZone, *workItemInterval)
		}
		if !hasFreeWindow {
			//	todo по каждой зоне в течение дня считаем варианты: для своей зоны - варианты сдвигов в рамках зоны педелах max_deadline_days, с учетом min_avialable_zones;
			changes, moveErr := moveZoneAvailabe(zoneScheduleByZone, *workItemInterval, move, cancelAuto, cancelManual)
			if moveErr != nil || len(changes) == 0 {
				err = moveErr
				return
			} else {
				hasFreeWindow = true
				scheduleRyZone[z] = append(scheduleRyZone[z], wi)
				scheduleRyZone[z] = append(scheduleRyZone[z], changes...)
			}
			return
		}
	}
	if hasFreeWindow {
		//min_avialable_zones вполняется -> 201 планируем
		available_count := 0
		for z := range sch.Config.WhiteList {
			if checkZoneAvailabe(merge(zoneSchedule[z], scheduleRyZone[z]), *workItemInterval) {
				available_count++
			}
		}
		if available_count > int(sch.Config.MinAvialableZones) {
			for _, i := range scheduleRyZone {
				schedule = append(schedule, i...)
			}

		} else {
			err = fmt.Errorf("unable to schedule work: should keep min available zones = %v", sch.Config.MinAvialableZones)
			return
		}
	} else {
		err = fmt.Errorf("unable to schedule work: interval alredy occupied")
		return
	}
	return
}

func merge(source []*IntervalWork, new []*models.WorkItem) []*IntervalWork {
	for _, n := range new {
		found := false
		for i, j := range source {
			if j.Work.Id == n.Id {
				found = true
				source[i].Work = n
			}
		}
		if !found {
			span, _ := getWorkInterval(n)
			iw := IntervalWork{
				Work: n,
				Span: span,
			}
			source = append(source, &iw)
		}
	}
	return source
}

func (sch *Scheduler) getZoneSchedule(from time.Time, to time.Time, zones []string, statuses []string) (zoneSchedules map[string][]*IntervalWork, err error) {
	works, err := sch.Repository.List(sch.ctx, from, to, zones, statuses)
	if err != nil {
		return
	}
	zoneSchedules = make(map[string][]*IntervalWork)
	for _, w := range works {
		interval, intErr := getWorkInterval(w)
		if intErr != nil {
			err = intErr
			return
		}
		iw := IntervalWork{
			Span: interval,
			Work: w,
		}
		for _, z := range w.Zones {
			if value, ok := zoneSchedules[z]; !ok {
				zoneSchedules[z] = []*IntervalWork{&iw}
			} else {
				zoneSchedules[z] = append(value, &iw)
			}
		}
	}
	return
}

func checkZoneAvailabe(zoneSchedule []*IntervalWork, checkInterval interval.Span) (available bool) {
	available = true
	for _, interv := range zoneSchedule {
		if interv.Span.IsIntersection(checkInterval) {
			available = false
			//todo считать сдвиги по зонам
			break
		}
	}
	return
}

func moveZoneAvailabe(zoneSchedule []*IntervalWork, checkInterval interval.Span, move bool, cancelAuto bool, cancelManual bool) (changes []*models.WorkItem, err error) {
	for _, interv := range zoneSchedule {
		if interv.Span.IsIntersection(checkInterval) {
			if (cancelManual && interv.Work.WorkType == WorkTypeManual) || (cancelAuto && interv.Work.WorkType == WorkTypeAutomatic) {
				if interv.Work.Status == StatusInProgress {
					err = fmt.Errorf("unable to cancel in progress work %v", interv.Work.Id)
					return
				} else {
					interv.Work.Status = StatusCancelled
					changes = append(changes, interv.Work)
					break
					// todo переносить или сжимать автоматические вместо отмены
				}
			}
			if move && interv.Work.WorkType == WorkTypeAutomatic {
				changes, err = createMovement(zoneSchedule, checkInterval)
			}
		}
	}
	if len(changes) == 0 {
		err = fmt.Errorf("unable to cancel or move work")
	}
	return
}

func createMovement(zoneSchedule []*IntervalWork, checkInterval interval.Span) (changes []*models.WorkItem, err error) {
	sort.Slice(zoneSchedule, func(i, j int) bool {
		return zoneSchedule[i].Work.StartDate.Before(zoneSchedule[j].Work.StartDate)
	})
	last_interval := checkInterval
	for _, inter := range zoneSchedule {
		if last_interval.Start().Before(inter.Span.Start()) && last_interval.End().After(inter.Span.Start()) {
			inter.Work.StartDate = last_interval.End()
			inter.Span, err = getWorkInterval(inter.Work)
			if err != nil {
				return
			}
			last_interval = *inter.Span
			changes = append(changes, inter.Work)
		} else {
			break
		}
	}
	return
}

func (sch *Scheduler) checkZoneLists(zone string, wi *models.WorkItem) (availavle bool, err error) {
	// проверяем, если зона в блеклисте && работы != критичные -> 500 возвращаем полную невозможность
	if slices.Contains(sch.Config.BlackList, zone) && wi.Priority != string(PriorityCritical) {
		err = fmt.Errorf("zone %v is in black list, unable to Schedule work with non-critical priority", zone)
		return
	}
	// проверяем, если зона в вайт листе && работы не в окне -> 500 возвращаем невозможность c вариантами сдвига
	windows, ok := sch.Config.WhiteList[zone]
	if ok {
		workIntervals := []configuration.Window{}
		endDate := wi.StartDate.Add(time.Duration(wi.DurationMinutes) * time.Minute)
		if wi.DurationMinutes >= 60*24 {
			for _, window := range windows {
				if window.EndHour-window.StartHour < 24 {
					err = fmt.Errorf("work duration %v is longer than zone white-list windows", wi.DurationMinutes)
					return
				}
			}
		}

		endHour := endDate.Hour()
		if endDate.Minute() != 0 {
			endHour++
		}
		if endDate.Day() > wi.StartDate.Day() {
			workIntervals = append(workIntervals, configuration.Window{StartHour: uint32(wi.StartDate.Hour()), EndHour: 24})
			workIntervals = append(workIntervals, configuration.Window{StartHour: 0, EndHour: uint32(endHour)})
		}
		if endHour != 0 {
			workIntervals = append(workIntervals, configuration.Window{StartHour: uint32(wi.StartDate.Hour()), EndHour: uint32(endHour)})
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
				err = fmt.Errorf("zone %v is in white list, but work time is not in zone white-list window", zone)
				//todo - варианты сдвига
				return
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
