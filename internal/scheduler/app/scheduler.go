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
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

const (
	PriorityCritical  = "critical"
	PriorityRegular   = "regular"
	WorkTypeAutomatic = "automatic"
	WorkTypeManual    = "manual"
	StatusInProgress  = "in_progress"
	StatusPlanned     = "planned"
	Statuscanceled    = "canceled"
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

func (sch *Scheduler) MoveWork(wis []*models.WorkItem) (schedule []*models.WorkItem, userMustApprove bool, err error) {
	sort.Slice(wis, func(i, j int) bool {
		return wis[i].StartDate.Before(wis[j].StartDate)
	})
	from := wis[0].StartDate.Add(time.Minute * time.Duration(-1*Max(sch.Config.MaxWorkDurationMinutes.Automatic, sch.Config.MaxWorkDurationMinutes.Manual)))
	to := wis[len(wis)-1].StartDate.Add(24 * time.Hour * time.Duration(sch.Config.MaxDeadlineDays)).Add(time.Minute * time.Duration(-1*(wis[len(wis)-1].DurationMinutes)))

	// текущее расписание
	allZonesSchedule, err := sch.getAllZonesSchedule(from, to)
	if err != nil {
		return
	}
	// not inclide current saved wi in schedule on move
	for z, sched := range allZonesSchedule.scheduleByZones {
		for i, w := range sched {
			if w.Work.WorkId == wis[0].WorkId {
				sched[i] = sched[len(sched)-1]
				allZonesSchedule.scheduleByZones[z] = sched[:len(sched)-1]
				break
			}
		}
	}
	//Критичные работы помещать в расписание вне очереди, принудительно отменяя обычные ручные работы и сжимая или перенося работы обычные автоматические.
	//При получении заявок на ручные работы отдавать им приоритет, отменяя работы автоматического типа.
	for _, wi := range wis {
		newSchedule, mustApprove, zoneErr := sch.chekScheduleChange(allZonesSchedule, wi, true, (wi.WorkType == WorkTypeManual), (wi.Priority == PriorityCritical))
		if zoneErr != nil {
			userMustApprove = mustApprove
			err = zoneErr
		}
		if len(newSchedule) != 0 {
			if len(newSchedule) > 1 {
				userMustApprove = true
			}
			schedule = append(schedule, newSchedule...)
		}
	}
	return
}

func (sch *Scheduler) ScheduleWork(wi *models.WorkItem) (schedule []*models.WorkItem, userMustApprove bool, err error) {
	userMustApprove = false
	from := wi.StartDate.Add(time.Minute * time.Duration(-1*Max(sch.Config.MaxWorkDurationMinutes.Automatic, sch.Config.MaxWorkDurationMinutes.Manual)))
	to := wi.StartDate.Add(24 * time.Hour * time.Duration(sch.Config.MaxDeadlineDays)).Add(time.Minute * time.Duration(-1*wi.DurationMinutes))

	// текущее расписание
	allZonesSchedule, err := sch.getAllZonesSchedule(from, to)
	if err != nil {
		return
	}
	//Критичные работы помещать в расписание вне очереди, принудительно отменяя обычные ручные работы и сжимая или перенося работы обычные автоматические.
	//При получении заявок на ручные работы отдавать им приоритет, отменяя работы автоматического типа.
	newSchedule, mustApprove, zoneErr := sch.chekScheduleChange(allZonesSchedule, wi, true, (wi.WorkType == WorkTypeManual), (wi.Priority == PriorityCritical))
	if zoneErr != nil {
		if len(newSchedule) == 0 {
			err = errors.Errorf("Unable to schedule new work: %s", zoneErr)
		} else {
			err = zoneErr
		}
	}
	if len(newSchedule) != 0 {
		userMustApprove = mustApprove
		if len(newSchedule) > 1 {
			userMustApprove = true
		}
		// schedule = append(schedule, mergeWiZones(wiChanges)...)
		schedule = append(schedule, newSchedule...)
	}
	return
}

func (sch *Scheduler) ProlongateWorkById(wis []*models.WorkItem) (schedule []*models.WorkItem, userMustApprove bool, err error) {
	userMustApprove = true
	sort.Slice(wis, func(i, j int) bool {
		return wis[i].StartDate.Before(wis[j].StartDate)
	})
	from := wis[0].StartDate.Add(time.Minute * time.Duration(-1*Max(sch.Config.MaxWorkDurationMinutes.Automatic, sch.Config.MaxWorkDurationMinutes.Manual)))
	to := wis[len(wis)-1].StartDate.Add(24 * time.Hour * time.Duration(sch.Config.MaxDeadlineDays)).Add(time.Minute * time.Duration(-1*(wis[len(wis)-1].DurationMinutes)))

	// текущее расписание
	allZonesSchedule, err := sch.getAllZonesSchedule(from, to)
	if err != nil {
		return
	}

	for _, wi := range wis {
		newSchedule, userApprove, zoneErr := sch.chekScheduleChange(allZonesSchedule, wi, true, true, false)
		if zoneErr != nil {
			userMustApprove = false
			err = zoneErr
		}

		if len(newSchedule) != 0 {
			userMustApprove = userApprove
			if len(newSchedule) > 1 {
				userMustApprove = true
			}
			// schedule = append(schedule, mergeWiZones(wiChanges)...)
			schedule = append(schedule, newSchedule...)
		}
	}
	return
}

func (sch *Scheduler) chekScheduleChange(zonesSchedule Schedule, wi *models.WorkItem, move bool, cancelAuto bool, cancelManual bool) (schedule []*models.WorkItem, userMustApprove bool, err error) {
	hasFreeWindow := false
	workItemInterval, err := getWorkInterval(wi)
	if err != nil {
		return
	}
	// wiChanges = make(map[string]*models.WorkItem)
	plannedWI := []*IntervalWork{}
	intervalAvailabilityForPlanned := [][]string{}

	sort.Slice(wi.Zones, func(i, j int) bool {
		return wi.Zones[i] < wi.Zones[j]
	})
	for zi, z := range wi.Zones {
		scheduleWIZone := []*models.WorkItem{}
		wiCopyForThisZ := *wi
		wiCopyForThisZ.Zones = []string{z}

		// проверить black и white листы
		available, zoneErr := sch.checkZoneLists(z, &wiCopyForThisZ)
		// попали в black лист или интервал заведомо больше, чем есть в white листе
		if zoneErr != nil {
			err = zoneErr
			return
		}
		// не попали по времени в white лист
		var minStartDate = wiCopyForThisZ.StartDate
		if !available {
			date, winErr := sch.getNearestZoneWindowStart(z, &wiCopyForThisZ)
			if winErr != nil {
				err = winErr
				return
			}
			minStartDate = date
			userMustApprove = true
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

		// проверяем по всем зонам работ, нет ли работ в это время или пытаемся сдвинуть планируемые работы по всем зонам сразу (для 1 зоны)
		if zi == 0 && len(wi.Zones) > 1 {
			startTime, _, moveErr := sch.moveToNextAvailable(zonesSchedule.scheduleByZones, "", wi)
			if moveErr == nil {
				if wi.StartDate != startTime {
					userMustApprove = true
				}
				change := wi
				change.StartDate = startTime
				change.Status = StatusPlanned
				schedule = append(schedule, change)
				return
			}
		}

		zoneScheduleByZone, ok := zonesSchedule.scheduleByZones[z]
		sugestedSchedule := zonesSchedule.scheduleByZones
		if ok {
			// для своей зоны - варианты сдвигов в рамках зоны педелах max_deadline_days, с учетом min_avialable_zones;
			// добавляем все ранее просчитанное по прошлым зонам
			if len(plannedWI) > 0 {
				for _, p := range plannedWI {
					for _, z := range p.Work.Zones {
						sugestedSchedule[z] = append(sugestedSchedule[z], p)
					}
				}
			}
			sugestedWi := *wi
			sugestedWi.Zones = []string{z}

			// если не смогли запланировать работы одновременно
			// eсли ранее планировали работы в другой зоне, пытаемся запланировать туда же, поощряем интервал с большей доступностью
			moreAvailable := -1
			for i, interv := range intervalAvailabilityForPlanned {
				if len(interv) > moreAvailable && slices.Contains(interv, z) {
					moreAvailable = i
				}
			}
			if moreAvailable > -1 {
				tryWi := *plannedWI[moreAvailable].Work
				tryWi.Zones = append(tryWi.Zones, z)
				sort.Slice(tryWi.Zones, func(i, j int) bool {
					return tryWi.Zones[i] < tryWi.Zones[j]
				})
				tryWi.Status = StatusPlanned
				span, _ := getWorkInterval(&tryWi)
				startTime, _, moveErr := sch.moveToNextAvailable(zonesSchedule.scheduleByZones, z, &tryWi)
				if moveErr == nil && startTime == tryWi.StartDate {
					trySchedule := zonesSchedule.scheduleByZones
					newIntervalWork := IntervalWork{
						Work: &tryWi,
						Span: span,
					}
					trySchedule[z] = append(trySchedule[z], &newIntervalWork)
					ok, _ := sch.checkMinAvailableZones(trySchedule, workItemInterval)
					if ok {
						plannedWI[moreAvailable].Work.Zones = tryWi.Zones
						sugestedSchedule[z] = append(sugestedSchedule[z], &newIntervalWork)
						userMustApprove = wi.StartDate != tryWi.StartDate
						continue
					}
				}
			}

			// проверить, нет ли работ в это время в каждой, если нет -> сдвиг, отмена и т.п. предложения
			startTime, availableInZones, moveErr := sch.moveToNextAvailable(sugestedSchedule, z, &sugestedWi)
			if moveErr == nil {
				change := &sugestedWi
				change.StartDate = startTime
				change.Zones = []string{z}
				change.Status = StatusPlanned
				scheduleWIZone = append(scheduleWIZone, change)

				span, _ := getWorkInterval(change)
				newIntervalWork := IntervalWork{
					Work: change,
					Span: span,
				}
				plannedWI = append(plannedWI, &newIntervalWork)
				intervalAvailabilityForPlanned = append(intervalAvailabilityForPlanned, availableInZones)
				sugestedSchedule[z] = append(sugestedSchedule[z], &newIntervalWork)
				schedule = append(schedule, scheduleWIZone...)
				userMustApprove = wi.StartDate != startTime
				continue
			}
		}

		// запланировать на попозже не получилось, надо двигать другие
		zoneWorkItemInterval, _ := interval.New(minStartDate, minStartDate.Add(time.Duration(wi.DurationMinutes)*time.Minute))
		for zoneWorkItemInterval.End().Before(wiCopyForThisZ.Deadline) {
			if ok {
				hasFreeWindow = checkZoneAvailabe(zoneScheduleByZone, zoneWorkItemInterval)
				if !hasFreeWindow {
					// варианты сдвигов и отмен других тасок в расписании, если это разрешено
					changes, moveErr := moveOrCancelOthers(zoneScheduleByZone, zoneWorkItemInterval, move, cancelAuto, cancelManual)
					if moveErr != nil || len(changes) == 0 {
						err = moveErr
						return
					} else {
						hasFreeWindow = true
						scheduleWIZone = append(scheduleWIZone, changes...)
						schedule = append(schedule, scheduleWIZone...)
						userMustApprove = true
					}
					return
				}
			}

			sugestedSchedule[z] = merge(sugestedSchedule[z], scheduleWIZone)
			//min_avialable_zones вполняется -> 201 планируем
			ok, _ := sch.checkMinAvailableZones(sugestedSchedule, workItemInterval)
			if ok {
				wiCopyForThisZ.StartDate = zoneWorkItemInterval.Start()
				userMustApprove = true
				schedule = append(schedule, scheduleWIZone...)
				break
			} else { //min_avialable_zones не вполняется -> пробуем сдвинуть еще дальше
				zoneWorkItemInterval, _ = interval.New(
					zoneWorkItemInterval.End(),
					zoneWorkItemInterval.End().Add(time.Duration(wiCopyForThisZ.DurationMinutes)*time.Minute))
				scheduleWIZone = []*models.WorkItem{}
			}
		}
		if !hasFreeWindow {
			err = fmt.Errorf("unable to schedule work: interval already occupied and unable to move to any time before deadline")
			return
		}
		schedule = append(schedule, scheduleWIZone...)
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
		sort.Slice(w.Zones, func(i, j int) bool {
			return w.Zones[i] < w.Zones[j]
		})
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

func (sch *Scheduler) checkMinAvailableZones(allZonesSchedule map[string][]*IntervalWork, workItemInterval *interval.Span) (ok bool, availableInZones []string) {
	availableCount := 0
	for z := range sch.Config.WhiteList {
		if checkZoneAvailabe(allZonesSchedule[z], *workItemInterval) {
			availableCount++
			availableInZones = append(availableInZones, z)
		}
	}
	ok = availableCount >= int(sch.Config.MinAvialableZones)
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

func (sch *Scheduler) moveToNextAvailable(sugestedAllZonesSchedule map[string][]*IntervalWork, currentZone string, wi *models.WorkItem) (startTime time.Time, availableInZones []string, err error) {
	if currentZone != "" {
		_, ok := sugestedAllZonesSchedule[currentZone]
		if !ok {
			err = fmt.Errorf("moveToNextAvailable error currentZone %v not found in sugestedAllZonesSchedule", currentZone)
			return
		}
	}
	var resultZoneSchedule []*IntervalWork
	sugestedStartTime := wi.StartDate
	sugestedStartTimeWithPause := sugestedStartTime
	sugestedEndTime := sugestedStartTime.Add(time.Duration(wi.DurationMinutes) * time.Minute)
	sugestedEndTimeWithPause := sugestedEndTime
	for sugestedEndTime.Before(wi.Deadline) {
		moved := false
		// проверяем интервал работ на предмет пересечения с работами в зоне
		var maxpause int32
		checkInterval, _ := interval.New(sugestedStartTimeWithPause, sugestedEndTimeWithPause)
		for z, zoneSchedule := range sugestedAllZonesSchedule {
			if (currentZone == "" || currentZone == z) && len(zoneSchedule) > 0 {
				// сортируем по startDate
				sort.Slice(zoneSchedule, func(i, j int) bool {
					return zoneSchedule[i].Work.StartDate.Before(zoneSchedule[j].Work.StartDate)
				})
				resultZoneSchedule = append(resultZoneSchedule, zoneSchedule...)
				for _, zi := range zoneSchedule {
					pause := sch.Config.PausesMinutes[z]
					if pause > maxpause {
						maxpause = pause
					}
					zoneCheckInterval, _ := interval.New(
						checkInterval.Start().Add(time.Duration(-1*pause)*time.Minute),
						checkInterval.End().Add(time.Duration(pause)*time.Minute))
					if zi.Span.IsIntersection(zoneCheckInterval) && zi.Span.Start() != zoneCheckInterval.End() && zi.Span.End() != zoneCheckInterval.Start() {
						sugestedStartTime = zi.Span.End()
						sugestedStartTimeWithPause = zi.Span.End().Add(time.Duration(1*pause) * time.Minute)
						/// TODO поиск начала интервала сободного времени
						sugestedEndTime = sugestedStartTime.Add(time.Duration(wi.DurationMinutes) * time.Minute)
						sugestedEndTimeWithPause = sugestedStartTimeWithPause.Add(time.Duration(wi.DurationMinutes) * time.Minute)
						checkInterval, _ = interval.New(sugestedStartTimeWithPause, sugestedEndTimeWithPause)
						moved = true
						break
					}
				}
				if moved {
					break
				}
			}
			hasFreeWindow := checkZoneAvailabe(resultZoneSchedule, checkInterval)
			if hasFreeWindow {
				resultSchedule := copyIntervalWork(sugestedAllZonesSchedule)
				sort.Slice(wi.Zones, func(i, j int) bool {
					return wi.Zones[i] < wi.Zones[j]
				})
				for _, z := range wi.Zones {
					resultWi := *wi
					resultWi.StartDate = checkInterval.Start()
					resultSchedule[z] = append(resultSchedule[z], &IntervalWork{Work: &resultWi, Span: &checkInterval})
				}
				minIntervalOk, zones := sch.checkMinAvailableZones(resultSchedule, &checkInterval)
				if minIntervalOk {
					startTime = sugestedStartTime
					availableInZones = zones
					return
				}
			}
		}
		if moved {
			continue
		}

		sugestedStartTimeWithPause = checkInterval.End().Add(time.Duration(maxpause) * time.Minute)
		sugestedEndTimeWithPause = sugestedStartTimeWithPause.Add(time.Duration(wi.DurationMinutes) * time.Minute)
		sugestedEndTime = sugestedStartTimeWithPause.Add(time.Duration(wi.DurationMinutes) * time.Minute)
	}
	err = fmt.Errorf("unable to move work to another time before deadline")
	return
}

func moveOrCancelOthers(zoneSchedule []*IntervalWork, checkInterval interval.Span, move bool, cancelAuto bool, cancelManual bool) (changes []*models.WorkItem, err error) {
	// TODO Pauses
	for _, interv := range zoneSchedule {
		if interv.Span.IsIntersection(checkInterval) {
			if (cancelManual && interv.Work.WorkType == WorkTypeManual) || (cancelAuto && interv.Work.WorkType == WorkTypeAutomatic) {
				if interv.Work.Status == StatusInProgress {
					err = fmt.Errorf("unable to cancel in progress work %v", interv.Work.WorkId)
					return
				} else {
					interv.Work.Status = Statuscanceled
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

	// TODO Pauses
	for _, inter := range zoneSchedule {
		if last_interval.Start().Before(inter.Span.Start()) && last_interval.End().After(inter.Span.Start()) {
			inter.Work.StartDate = last_interval.End()
			inter.Span, err = getWorkInterval(inter.Work)
			if err != nil {
				return
			}
			last_interval = *inter.Span
			inter.Work.Status = StatusPlanned
			changes = append(changes, inter.Work)
		} else {
			break
		}
	}
	return
}

func (sch *Scheduler) checkZoneLists(zone string, wi *models.WorkItem) (availavle bool, err error) {
	// проверяем, если зона в блеклисте && работы != критичные -> 500 возвращаем полную невозможность - err
	if slices.Contains(sch.Config.BlackList, zone) && wi.Priority != string(PriorityCritical) {
		err = fmt.Errorf("zone %v is in black list, unable to Schedule work with non-critical priority", zone)
		return
	}
	// проверяем, если зона в вайт листе && работы не в окне -> 500 возвращаем невозможность c вариантами сдвига
	windows, ok := sch.Config.WhiteList[zone]
	if !ok {
		err = fmt.Errorf("zone %v not found in zone white-list", zone)
		return
	} else {
		workIntervals := []configuration.Window{}
		endDate := wi.StartDate.Add(time.Duration(wi.DurationMinutes) * time.Minute)
		if wi.DurationMinutes >= 60*24 {
			for _, window := range windows {
				if window.EndHourDuration-window.StartHourDuration < 24*time.Hour {
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
			workIntervals = append(workIntervals, configuration.Window{
				StartHourDuration: time.Duration(wi.StartDate.Hour()) * time.Hour,
				EndHourDuration:   24 * time.Hour,
			})
			workIntervals = append(workIntervals, configuration.Window{
				StartHourDuration: 0,
				EndHourDuration:   time.Duration(endHour) * time.Hour,
			})
		}
		if endHour != 0 {
			workIntervals = append(workIntervals, configuration.Window{
				StartHourDuration: time.Duration(wi.StartDate.Hour()) * time.Hour,
				EndHourDuration:   time.Duration(endHour) * time.Hour,
			})
		}

		for _, workInterval := range workIntervals {
			isInWindow := false
			for _, window := range windows {
				// без учета пауз
				if workInterval.EndHourDuration <= window.StartHourDuration { //раньше текущего интервала
					continue
				}
				if workInterval.EndHourDuration > window.StartHourDuration {
					if workInterval.StartHourDuration < window.StartHourDuration { //наползает в начале
						break
					}
					if workInterval.StartHourDuration >= window.StartHourDuration && workInterval.EndHourDuration <= window.EndHourDuration { //целиком в текущем интервале
						isInWindow = true
						break
					}
					if workInterval.StartHourDuration >= window.StartHourDuration && workInterval.EndHourDuration > window.EndHourDuration { //наползает в конце
						break
					}
				}
				if workInterval.StartHourDuration < window.EndHourDuration && workInterval.EndHourDuration > window.EndHourDuration { //больше интервала с обеих сторон
					break
				}
				continue
			}
			if !isInWindow {
				return
			}
		}
		availavle = true
	}
	return
}

func (sch *Scheduler) getNearestZoneWindowStart(zone string, wi *models.WorkItem) (start time.Time, err error) {
	windows, ok := sch.Config.WhiteList[zone]
	if !ok {
		err = fmt.Errorf("zone %v not found in zone white-list", zone)
		return
	} else {
		day := time.Date(wi.StartDate.Year(), wi.StartDate.Month(), wi.StartDate.Day(), 0, 0, 0, 0, time.UTC)
		sugestions := []time.Time{}
		for _, w := range windows {
			if time.Duration(wi.StartDate.Hour())*time.Hour < w.StartHourDuration {
				sugestions = append(sugestions, day.Add(w.StartHourDuration))
				continue
			}
			if time.Duration(wi.StartDate.Hour())*time.Hour > w.EndHourDuration {
				sugestions = append(sugestions, day.Add(w.StartHourDuration).AddDate(0, 0, 1))
				continue
			}
		}
		if len(sugestions) > 0 {
			sort.Slice(sugestions, func(i, j int) bool {
				return sugestions[i].Before(sugestions[j])
			})
			start = sugestions[0]
		}
	}
	return
}

// func mergeWiZones(wiChanges map[string]*models.WorkItem) []*models.WorkItem {
// 	result := []*models.WorkItem{}
// 	for z, wiCh := range wiChanges {
// 		if len(result) == 0 {
// 			result = append(result, wiCh)
// 		} else {
// 			appended := false
// 			for _, res := range result {
// 				if res.StartDate == wiCh.StartDate {
// 					appended = true
// 					res.Zones = append(res.Zones, z)
// 					break
// 				}
// 			}
// 			if !appended {
// 				result = append(result, wiCh)
// 			}
// 		}
// 	}
// 	return result
// }

func copyIntervalWork(m map[string][]*IntervalWork) (copy map[string][]*IntervalWork) {
	copy = make(map[string][]*IntervalWork)
	for n, i := range m {
		copy[n] = i
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
func MaxTime(x time.Time, y time.Time) time.Time {
	if x.Before(y) {
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
