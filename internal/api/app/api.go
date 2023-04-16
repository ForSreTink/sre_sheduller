package api

import (
	"encoding/json"
	"net/http"
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
	Alternative []models.WorkItem `json:"alternative,omitempty"`
	ErrorCode   string            `json:"errorCode,omitempty"`
	Message     string            `json:"message,omitempty"`
}

func (a *Api) writeBadRequestError(w http.ResponseWriter, code string, message string) {
	w.WriteHeader(http.StatusBadRequest)
	err := ErrorStruct{
		Message:     message,
		ErrorCode:   code,
		Alternative: []models.WorkItem{},
	}
	errBytes, e := json.Marshal(err)
	if e != nil {
		w.Write([]byte(e.Error()))
	} else {
		w.Write(errBytes)
	}
}

func (a *Api) writeInternalError(w http.ResponseWriter, code string, message string, alternative []models.WorkItem) {
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
	}

	// TODO: код проверки условий для добавления

	works, unexepted, err := a.Scheduller.ScheduleWork(work)
	if err != nil {
		if unexepted {
			a.writeInternalError(w, "Unexpected error", err.Error(), []models.WorkItem{})
		}
		a.writeInternalError(w, "Unable to shedule", err.Error(), *works)
		return
	}

	work, err = a.RepoData.Add(r.Context(), work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []models.WorkItem{})
		return
	}
	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(work_b)
}

func (a *Api) GetWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()

	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []models.WorkItem{})
		return
	}
	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) CancelWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()

	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []models.WorkItem{})
		return
	}

	work.Status = "cancelled"

	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) MoveWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()
	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []models.WorkItem{})
		return
	}

	// TODO: код для определения возможности переноса

	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) ProlongateWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()
	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []models.WorkItem{})
		return
	}

	// TODO: код для определения возможности переноса

	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, "internal error", err.Error(), []models.WorkItem{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}
