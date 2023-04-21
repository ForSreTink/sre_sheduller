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

type Schedule struct {
	scheduleByZones map[string][]*IntervalWork
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

	from := wi.StartDate.Add(time.Minute * time.Duration(-1*Max(sch.Config.MaxWorkDurationMinutes.Automatic, sch.Config.MaxWorkDurationMinutes.Manual)))
	to := wi.StartDate.Add(24 * time.Hour * time.Duration(sch.Config.MaxDeadlineDays)).Add(time.Minute * time.Duration(-1*wi.DurationMinutes))

	// текущее расписание
	allZonesSchedule, err := sch.getAllZonesSchedule(from, to)
	if err != nil {
		return
	}
	// not inclide current saved wi in schedule on move
	for z, sched := range allZonesSchedule.scheduleByZones {
		for i, w := range sched {
			if w.Work.WorkId == wi.WorkId {
				sched[i] = sched[len(sched)-1]
				allZonesSchedule.scheduleByZones[z] = sched[:len(sched)-1]
				break
			}
		}
	}
	//Критичные работы помещать в расписание вне очереди, принудительно отменяя обычные ручные работы и сжимая или перенося работы обычные автоматические.
	//При получении заявок на ручные работы отдавать им приоритет, отменяя работы автоматического типа.
	newSchedule, wiChanges, zoneErr := sch.chekScheduleChange(allZonesSchedule, wi, true, (wi.WorkType == WorkTypeManual), (wi.Priority == PriorityCritical))
	if zoneErr != nil {
		errorIsUnexpected = false
		err = zoneErr
	}
	if len(newSchedule) == 0 {
		schedule = append(schedule, wi)
	} else {
		schedule = append(schedule, mergeWiZones(wiChanges)...)
		schedule = append(schedule, newSchedule...)
	}
	return
}

func (sch *Scheduler) ScheduleWork(wi *models.WorkItem) (schedule []*models.WorkItem, errorIsUnexpected bool, err error) {
	errorIsUnexpected = true
	from := wi.StartDate.Add(time.Minute * time.Duration(-1*Max(sch.Config.MaxWorkDurationMinutes.Automatic, sch.Config.MaxWorkDurationMinutes.Manual)))
	to := wi.StartDate.Add(24 * time.Hour * time.Duration(sch.Config.MaxDeadlineDays)).Add(time.Minute * time.Duration(-1*wi.DurationMinutes))

	// текущее расписание
	allZonesSchedule, err := sch.getAllZonesSchedule(from, to)
	if err != nil {
		return
	}
	//Критичные работы помещать в расписание вне очереди, принудительно отменяя обычные ручные работы и сжимая или перенося работы обычные автоматические.
	//При получении заявок на ручные работы отдавать им приоритет, отменяя работы автоматического типа.
	newSchedule, wiChanges, zoneErr := sch.chekScheduleChange(allZonesSchedule, wi, true, (wi.WorkType == WorkTypeManual), (wi.Priority == PriorityCritical))
	if zoneErr != nil {
		errorIsUnexpected = false
		err = zoneErr
	}
	if len(newSchedule) == 0 {
		schedule = append(schedule, wi)
	} else {
		schedule = append(schedule, mergeWiZones(wiChanges)...)
		schedule = append(schedule, newSchedule...)
	}
	return
}

func (sch *Scheduler) ProlongateWorkById(wi *models.WorkItem) (schedule []*models.WorkItem, errorIsUnexpected bool, err error) {
	errorIsUnexpected = true
	from := wi.StartDate.Add(time.Minute * time.Duration(-1*Max(sch.Config.MaxWorkDurationMinutes.Automatic, sch.Config.MaxWorkDurationMinutes.Manual)))
	to := wi.StartDate.Add(24 * time.Hour * time.Duration(sch.Config.MaxDeadlineDays)).Add(time.Minute * time.Duration(-1*wi.DurationMinutes))

	// текущее расписание
	allZonesSchedule, err := sch.getAllZonesSchedule(from, to)
	if err != nil {
		return
	}
	newSchedule, _, zoneErr := sch.chekScheduleChange(allZonesSchedule, wi, true, true, false)
	if zoneErr != nil {
		errorIsUnexpected = false
		err = zoneErr
	}

	//todo - check if we have problems in different zones
	schedule = append(schedule, wi)
	schedule = append(schedule, newSchedule...)
	return
}

func (sch *Scheduler) chekScheduleChange(zonesSchedule Schedule, wi *models.WorkItem, move bool, cancelAuto bool, cancelManual bool) (schedule []*models.WorkItem, wiChanges map[string]*models.WorkItem, err error) {
	hasFreeWindow := false
	workItemInterval, err := getWorkInterval(wi)
	if err != nil {
		return
	}
	wiChanges = make(map[string]*models.WorkItem)

	for _, z := range wi.Zones {
		scheduleWIZone := []*models.WorkItem{}
		wiCopyForThisZ := *wi
		wiCopyForThisZ.Zones = []string{z}

		// проверить black и white листы
		_, zoneErr := sch.checkZoneLists(z, wi)
		if zoneErr != nil {
			err = zoneErr
			return
		}
		// если пришло изменение работ, то надо удалить ранее существовавший айтем из списка сравнения
		tmp := zonesSchedule.scheduleByZones[z]
		removed_cnt := 0
		for i, s := range zonesSchedule.scheduleByZones[z] {
			if s.Work.WorkId == wiCopyForThisZ.WorkId {
				tmp[i-removed_cnt] = tmp[len(tmp)-1]
				tmp = tmp[:len(tmp)-1]
				removed_cnt++
			}
		}
		zonesSchedule.scheduleByZones[z] = tmp
		// проверить, нет ли работ в это время, если нет -> сдвиг, отмена и т.п. предложения
		zoneScheduleByZone, ok := zonesSchedule.scheduleByZones[z]

		zoneWorkItemInterval := workItemInterval
		for zoneWorkItemInterval.End().Before(wiCopyForThisZ.Deadline) {
			if ok {
				hasFreeWindow = checkZoneAvailabe(zoneScheduleByZone, *zoneWorkItemInterval)
				if !hasFreeWindow {
					//	todo по каждой зоне в течение дня считаем варианты: для своей зоны - варианты сдвигов в рамках зоны педелах max_deadline_days, с учетом min_avialable_zones;
					changes, moveErr := moveZoneAvailabe(zoneScheduleByZone, *zoneWorkItemInterval, move, cancelAuto, cancelManual)
					if moveErr != nil || len(changes) == 0 {
						err = moveErr
						return
					} else {
						hasFreeWindow = true
						scheduleWIZone = append(scheduleWIZone, changes...)
					}
					return
				}
			}
			//min_avialable_zones вполняется -> 201 планируем
			available_count := 0
			for zwl := range sch.Config.WhiteList {
				if checkZoneAvailabe(merge(zonesSchedule.scheduleByZones[zwl], scheduleWIZone), *workItemInterval) {
					available_count++
				}
			}
			if available_count > int(sch.Config.MinAvialableZones) {
				wiCopyForThisZ.StartDate = zoneWorkItemInterval.Start()
				wiChanges[z] = &wiCopyForThisZ
				schedule = append(schedule, scheduleWIZone...)
				break
			} else { //min_avialable_zones не вполняется -> пробуем сдвинуть еще дальше
				nextInterval, _ := interval.New(
					zoneWorkItemInterval.End(),
					zoneWorkItemInterval.End().Add(time.Duration(wiCopyForThisZ.DurationMinutes)*time.Minute))
				zoneWorkItemInterval = &nextInterval
				scheduleWIZone = []*models.WorkItem{}
			}
		}
		if !hasFreeWindow {
			err = fmt.Errorf("unable to schedule work: interval alredy occupied an unable to move to any time before deadline")
			return
		}
	}
	return
}

func merge(source []*IntervalWork, new []*models.WorkItem) []*IntervalWork {
	for _, n := range new {
		found := false
		for i, j := range source {
			if j.Work.WorkId == n.WorkId {
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

func (sch *Scheduler) getAllZonesSchedule(from time.Time, to time.Time) (zoneSchedules Schedule, err error) {
	statuses := []string{StatusPlanned, StatusInProgress}
	works, err := sch.Repository.List(sch.ctx, from, to, []string{}, statuses)
	if err != nil {
		return
	}
	sort.Slice(works, func(i, j int) bool {
		return works[i].StartDate.Before(works[j].StartDate)
	})

	zoneSchedules = Schedule{
		scheduleByZones: make(map[string][]*IntervalWork),
	}
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
			if value, ok := zoneSchedules.scheduleByZones[z]; !ok {
				zoneSchedules.scheduleByZones[z] = []*IntervalWork{&iw}
			} else {
				zoneSchedules.scheduleByZones[z] = append(value, &iw)
			}
		}
	}
	return
}

// func (s *Schedule) getWindowWorks(searchInterval interval.Span) (windows map[string][]*IntervalWork) {
// 	windows = make(map[string][]*IntervalWork)
// 	for z, works := range s.scheduleByZones {
// 		for _, w := range works {
// 			if w.Span.IsIntersection(searchInterval) {
// 				windows[z] = append(windows[z], w)
// 			}
// 			if w.Span.Start().After(searchInterval.End()) {
// 				break
// 			}
// 		}
// 	}
// 	return
// }

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
					err = fmt.Errorf("unable to cancel in progress work %v", interv.Work.WorkId)
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
	// alredy sorted??

	// sort.Slice(zoneSchedule, func(i, j int) bool {
	// 	return zoneSchedule[i].Work.StartDate.Before(zoneSchedule[j].Work.StartDate)
	// })
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

func mergeWiZones(wiChanges map[string]*models.WorkItem) []*models.WorkItem {
	result := []*models.WorkItem{}
	for z, wiCh := range wiChanges {
		if len(result) == 0 {
			result = append(result, wiCh)
		} else {
			appended := false
			for _, res := range result {
				if res.StartDate == wiCh.StartDate {
					appended = true
					res.Zones = append(res.Zones, z)
					break
				}
			}
			if !appended {
				result = append(result, wiCh)
			}
		}
	}
	return result
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
