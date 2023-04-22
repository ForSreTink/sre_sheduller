package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"workScheduler/internal/configuration"
	"workScheduler/internal/repository"
	"workScheduler/internal/scheduler/app"
	"workScheduler/internal/scheduler/models"

	"github.com/google/uuid"
)

type Api struct {
	RepoData   repository.ReadWriteRepository
	Scheduller *app.Scheduler
	Config     *configuration.Configurator
}

func NewApi(repo repository.ReadWriteRepository, scheduler *app.Scheduler, config *configuration.Configurator) *Api {
	return &Api{
		RepoData:   repo,
		Scheduller: scheduler,
		Config:     config,
	}
}

type ErrorStruct struct {
	Alternative []*models.WorkItem `json:"alternative,omitempty"`
	ErrorCode   string             `json:"errorCode,omitempty"`
	Message     string             `json:"message,omitempty"`
}

func inArray(arr []string, i []string) bool {
	for _, v := range arr {
		for _, v2 := range i {
			if v == v2 {
				return true
			}
		}
	}
	return false
}

func inWhiteList(whitelist map[string][]configuration.Window, zones []string) bool {
	for k := range whitelist {
		for _, v := range zones {
			if k == v {
				return true
			}
		}
	}
	return false
}

func (a *Api) validateAddWork(work *models.WorkItem) error {
	ts := time.Now().UTC()
	errStr := ""
	fmt.Println(work)
	if work.Priority != "regular" && work.Priority != "critical" {
		errStr += "Unknown priority; "
	}
	if work.WorkType != "automatic" && work.WorkType != "manual" {
		errStr += "Unknown worktype; "
	}

	if work.WorkType != "manual" && work.Priority == "critical" {
		errStr += "Only manual works may has cretical priority; "
	}

	delta := work.Deadline.Sub(ts)
	if delta.Hours() <= 0 || int32(delta.Hours()/24) > a.Config.Data.MaxDeadlineDays {
		errStr += "Deadline can't be greater then now for 4 week; "
	}
	if inArray(work.Zones, a.Config.Data.BlackList) && work.Priority != "critical" {
		errStr += "Can't schedule work for zone in blacklist , excepted critical work; "
	}
	if inArray(a.Config.Data.BlackList, work.Zones) && work.Priority != "critical" {
		errStr += "Can't schedule work for zone in blacklist, excepted critical work; "
	}
	if !inWhiteList(a.Config.Data.WhiteList, work.Zones) && work.Priority != "critical" {
		errStr += "Can't schedule work for zone not in whitelist, excepted critical work; "
	}

	if work.StartDate.UTC().Unix() <= ts.Unix() {
		errStr += "Start Date can't be in past; "
	}
	if work.DurationMinutes < a.Scheduller.Config.MinWorkDurationMinutes.Automatic && work.WorkType == "automatic" {
		errStr += fmt.Sprintf("Automatic work duration can't be lower then %d minutes; ", a.Scheduller.Config.MinWorkDurationMinutes.Automatic)
	}
	if work.DurationMinutes < a.Scheduller.Config.MinWorkDurationMinutes.Manual && work.WorkType == "manual" {
		errStr += fmt.Sprintf("Manual work duration can't be lower then %d minutes; ", a.Scheduller.Config.MinWorkDurationMinutes.Manual)
	}
	if work.DurationMinutes > a.Scheduller.Config.MaxWorkDurationMinutes.Manual && work.Priority != "critical" {
		errStr += fmt.Sprintf("Manual work max duration can't be greater then %d minutes, excepted critical work; ", a.Scheduller.Config.MaxWorkDurationMinutes.Manual)
	}
	if work.DurationMinutes > a.Scheduller.Config.MaxWorkDurationMinutes.Automatic && work.Priority != "critical" {
		errStr += fmt.Sprintf("Automatic work max duration can't be greater then %d minutes, excepted critical work; ", a.Scheduller.Config.MaxWorkDurationMinutes.Automatic)
	}
	if work.StartDate.Minute()%5 != 0 && work.WorkType == "manual" {
		errStr += "Manual work started time must be multiple by 5 minutes; "
	}
	if work.StartDate.Second() != 0 && work.StartDate.Nanosecond() != 0 {
		errStr += "Work started time must be multiple by 1 minutes; "
	}

	if work.WorkType == "automatic" && work.MaxCompressionRate < 0.0 || work.MaxCompressionRate > 1 {
		errStr += "MaxCompressionRate must be float value between 0 and 1;"
	}
	if errStr != "" {
		return errors.New(errStr)
	} else {
		return nil
	}
}

func (a *Api) writeError(w http.ResponseWriter, status int, code string, message error, alternative []*models.WorkItem) {
	w.WriteHeader(status)
	if message == nil {
		message = errors.New("")
	}
	err := ErrorStruct{
		Message:     message.Error(),
		ErrorCode:   code,
		Alternative: alternative,
	}
	errBytes, e := json.Marshal(err)
	if e != nil {
		w.Write([]byte(e.Error()))
	} else {
		w.Write(errBytes)
	}
}

func (a *Api) Getschedule(w http.ResponseWriter, r *http.Request, params GetscheduleParams) {
	defer r.Body.Close()

	var statuses []string
	var zones []string

	if params.FromDate == nil || params.ToDate == nil {
		a.writeError(w, http.StatusBadRequest, "Bad request", errors.New("FromDate and StartDate can't be null"), []*models.WorkItem{})
		return
	}

	if params.FromDate.Unix() >= params.ToDate.Unix() {
		a.writeError(w, http.StatusBadRequest, "Bad request", errors.New("FromDate must be before StartDate"), []*models.WorkItem{})
		return
	}

	if params.Statuses != nil {
		for _, s := range *params.Statuses {
			statuses = append(statuses, string(s))
		}
	}

	if params.Zones != nil {
		for _, s := range *params.Zones {
			zones = append(zones, string(s))
		}
	}

	works, err := a.RepoData.List(r.Context(), *params.FromDate, *params.ToDate, zones, statuses)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}

	works_b, err := json.Marshal(works)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(works_b)
}

func (a *Api) AddWork(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	work := &models.WorkItem{}

	err := json.NewDecoder(r.Body).Decode(work)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, "Bad request", err, []*models.WorkItem{})
		return
	}

	if err := a.validateAddWork(work); err != nil {
		a.writeError(w, http.StatusBadRequest, "Bad request", err, []*models.WorkItem{})
		return
	}

	work.WorkId = uuid.New().String()
	work.InitialDuration = work.DurationMinutes
	work.InitialStartDate = work.StartDate
	work.CompressionRate = 1
	if work.WorkType == "manual" {
		work.MaxCompressionRate = 1
	}

	works, needUserApprove, err := a.Scheduller.ScheduleWork(work)
	if needUserApprove {
		a.writeError(w, http.StatusInternalServerError, "Unable to schedule", err, works)
		return
	}
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal server error", err, []*models.WorkItem{})
		return
	}

	added_work := []*models.WorkItem{}

	for _, work := range works {
		if work.Id.IsZero() {
			work, err := a.RepoData.Add(r.Context(), work)
			if err != nil {
				a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
				return
			}
			added_work = append(added_work, work)
		} else {
			_, err := a.RepoData.Update(r.Context(), work)
			if err != nil {
				a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
				return
			}
		}
	}
	work_b, err := json.Marshal(added_work)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) GetWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()

	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}
	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) CancelWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()

	works, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}

	for idx := range works {
		works[idx].Status = "canceled"
		works[idx], err = a.RepoData.Update(r.Context(), works[idx])
		if err != nil {
			a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
			return
		}
	}

	work_b, err := json.Marshal(works)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) MoveWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()
	works, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}

	w_b := &models.WorkItem{}
	err = json.NewDecoder(r.Body).Decode(w_b)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, "Bad request", err, []*models.WorkItem{})
		return
	}

	planned := false

	for _, work := range works {
		if work.Status == "planned" {
			planned = true
		}
		work.StartDate = w_b.StartDate

		if w_b.DurationMinutes != 0 {
			work.DurationMinutes = w_b.DurationMinutes
		}

		if err := a.validateAddWork(work); err != nil {
			a.writeError(w, http.StatusBadRequest, "Bad request", err, []*models.WorkItem{})
			return
		}
	}

	if !planned {
		a.writeError(w, http.StatusBadRequest, "Bad request", errors.New("Can't move work with status != planned"), []*models.WorkItem{})
		return
	}

	works, needUserApprove, err := a.Scheduller.MoveWork(works)
	if err != nil {
		if needUserApprove {
			a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		}
		a.writeError(w, http.StatusInternalServerError, "Unable to schedule", err, works)
		return
	}

	for _, work := range works {
		_, err = a.RepoData.Update(r.Context(), work)
		if err != nil {
			a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
			return
		}
	}

	work_b, err := json.Marshal(works)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) ProlongateWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()

	w_b := &models.WorkItem{}
	err := json.NewDecoder(r.Body).Decode(w_b)
	if err != nil {
		a.writeError(w, http.StatusBadRequest, "Bad request", err, []*models.WorkItem{})
		return
	}

	works, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}

	if len(works) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	in_progress := false

	for _, work := range works {
		if work.Status == "in_progress" {
			in_progress = true
		}

		if work.WorkType != "manual" {
			a.writeError(w, http.StatusBadRequest, "Bad request", errors.New("Only manual work may be prolongate"), []*models.WorkItem{})
			return
		}

		work.DurationMinutes = w_b.DurationMinutes
		if err := a.validateAddWork(work); err != nil {
			a.writeError(w, http.StatusBadRequest, "Bad request", err, []*models.WorkItem{})
			return
		}
	}

	if !in_progress {
		a.writeError(w, http.StatusBadRequest, "Bad request", errors.New("Can't prolongate work with status != in_progress"), []*models.WorkItem{})
		return
	}

	works, needUserApprove, err := a.Scheduller.ProlongateWorkById(works)
	if needUserApprove {
		a.writeError(w, http.StatusInternalServerError, "Unable to schedule", err, works)
		return
	}
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}

	updatedWorks := []*models.WorkItem{}

	for _, wks := range works {
		wk, err := a.RepoData.Update(r.Context(), wks)
		if err != nil {
			a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
			return
		}
		updatedWorks = append(updatedWorks, wk)
	}

	works_b, err := json.Marshal(updatedWorks)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, "Internal error", err, []*models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(works_b)
}
