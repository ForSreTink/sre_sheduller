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
	ts := time.Now()
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
		errStr += "Can't shedule work for zone in blacklist , excepted critical work; "
	}
	if !inWhiteList(a.Config.Data.WhiteList, work.Zones) && work.Priority != "critical" {
		errStr += "Can't shedule work for zone not in whitelist, excepted critical work; "
	}

	delta = work.StartDate.Sub(ts)
	if delta <= 0 {
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
	if work.StartDate.Minute()%1 != 0 && work.WorkType == "automatic" {
		errStr += "Automatic work started time must be multiple by 1 minutes; "
	}

	if errStr != "" {
		return errors.New(errStr)
	} else {
		return nil
	}
}

func (a *Api) writeBadRequestError(w http.ResponseWriter, code string, message string) {
	w.WriteHeader(http.StatusBadRequest)
	err := ErrorStruct{
		Message:     message,
		ErrorCode:   code,
		Alternative: []*models.WorkItem{},
	}
	errBytes, e := json.Marshal(err)
	if e != nil {
		w.Write([]byte(e.Error()))
	} else {
		w.Write(errBytes)
	}
}

func (a *Api) writeInternalError(w http.ResponseWriter, code string, message string, alternative []*models.WorkItem) {
	w.WriteHeader(http.StatusInternalServerError)
	err := ErrorStruct{
		Message:     message,
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
		a.writeBadRequestError(w, "Bad request", "FromDate and StartDate can't be null")
		return
	}

	if params.FromDate.Unix() >= params.ToDate.Unix() {
		a.writeBadRequestError(w, "Bad request", "FromDate must be before StartDate")
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
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}

	works_b, err := json.Marshal(works)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
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
		a.writeBadRequestError(w, "Bad request", err.Error())
		return
	}

	if err := a.validateAddWork(work); err != nil {
		a.writeBadRequestError(w, "Bad request", err.Error())
		return
	}

	works, unexepted, err := a.Scheduller.ScheduleWork(work)
	if err != nil {
		if unexepted {
			a.writeInternalError(w, "Unexpected error", err.Error(), []*models.WorkItem{})
		}
		a.writeInternalError(w, "Unable to shedule", err.Error(), works)
		return
	}

	work.Status = "planned"
	work, err = a.RepoData.Add(r.Context(), work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}
	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(work_b)
}

func (a *Api) GetWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()

	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}
	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) CancelWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()

	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}

	work.Status = "cancelled"

	work, err = a.RepoData.Update(r.Context(), work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}

	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		fmt.Println("AAAAAAAAAA")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) MoveWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()
	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}

	if work.Status != "planned" {
		a.writeBadRequestError(w, "Bad request", "Can't move work with status != planned")
		return
	}

	w_b := &models.WorkItem{}
	err = json.NewDecoder(r.Body).Decode(w_b)
	if err != nil {
		a.writeBadRequestError(w, "Bad request", err.Error())
		return
	}

	work.StartDate = w_b.StartDate

	if w_b.DurationMinutes != 0 {
		work.DurationMinutes = w_b.DurationMinutes
	}

	if err := a.validateAddWork(work); err != nil {
		a.writeBadRequestError(w, "Bad request", err.Error())
		return
	}

	works, unexepted, err := a.Scheduller.MoveWork(work)
	if err != nil {
		if unexepted {
			a.writeInternalError(w, "Unexpected error", err.Error(), []*models.WorkItem{})
		}
		a.writeInternalError(w, "Unable to shedule", err.Error(), works)
		return
	}

	work, err = a.RepoData.Update(r.Context(), work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}

	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) ProlongateWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()
	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}

	if work.Status != "in_progress" {
		a.writeBadRequestError(w, "Bad request", "Can't prolongate work with status != in_progress")
		return
	}

	if work.WorkType != "manual" {
		a.writeBadRequestError(w, "Bad request", "Only manual work may be prolongate")
		return
	}

	w_b := &models.WorkItem{}
	err = json.NewDecoder(r.Body).Decode(w_b)
	if err != nil {
		a.writeBadRequestError(w, "Bad request", err.Error())
		return
	}

	work.DurationMinutes = w_b.DurationMinutes
	if err := a.validateAddWork(work); err != nil {
		a.writeBadRequestError(w, "Bad request", err.Error())
		return
	}

	works, unexepted, err := a.Scheduller.ProlongateWorkById(work)
	if err != nil {
		if unexepted {
			a.writeInternalError(w, "Unexpected error", err.Error(), []*models.WorkItem{})
		}
		a.writeInternalError(w, "Unable to shedule", err.Error(), works)
		return
	}

	updatedWorks := []*models.WorkItem{}

	for _, wks := range works {
		wk, err := a.RepoData.Update(r.Context(), wks)
		if err != nil {
			a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
			return
		}
		updatedWorks = append(updatedWorks, wk)
	}

	works_b, err := json.Marshal(updatedWorks)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(works_b)
}
