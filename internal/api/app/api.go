package api

import (
	"encoding/json"
	"errors"
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

	if work.Priority != "regular" && work.Priority != "critical" {
		errStr += "Unknown priority; "
	}
	if work.WorkType != "automatic" && work.WorkType != "manual" {
		errStr += "Unknown worktype; "
	}

	delta := work.Deadline.Sub(ts)
	if inArray(work.Zones, a.Config.Data.BlackList) && work.Priority != "critical" {
		errStr += "Can't shedule work for zone in blacklist , excepted critical work; "
	}
	if !inWhiteList(a.Config.Data.WhiteList, work.Zones) && work.Priority != "critical" {
		errStr += "Can't shedule work for zone not in whitelist, excepted critical work; "
	}
	if delta.Hours()/24 > 4 {
		errStr += "Deadline can't be greater then now for 4 week; "
	}

	delta = work.StartDate.Sub(ts)
	if delta > 0 {
		errStr += "Start Date can't be in past; "
	}
	if work.DurationMinutes < 5 && work.WorkType == "automatic" {
		errStr += "Automatic work duration can't be lower then 5 minutes; "
	}
	if work.DurationMinutes < 30 && work.WorkType == "manual" {
		errStr += "Manual work duration can't be lower then 30 minutes; "
	}
	if work.DurationMinutes > 360 && work.Priority != "critical" {
		errStr += "Work max duration can't be greater then 360 minutes, excepted critical work; "
	}
	if work.StartDate.Minute()/5 != 0 && work.WorkType == "manual" {
		errStr += "Manual work started time must be multiple by 5 minutes; "
	}
	if work.StartDate.Minute()/1 != 0 && work.WorkType == "automatic" {
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
	// TODO: код проверки условий для добавления

	works, unexepted, err := a.Scheduller.ScheduleWork(work)
	if err != nil {
		if unexepted {
			a.writeInternalError(w, "Unexpected error", err.Error(), []*models.WorkItem{})
		}
		a.writeInternalError(w, "Unable to shedule", err.Error(), works)
		return
	}

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

	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
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

	// TODO: код для определения возможности переноса

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

	// TODO: код для определения возможности переноса

	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []*models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}
