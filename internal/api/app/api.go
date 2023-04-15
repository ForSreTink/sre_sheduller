package api

import (
	"encoding/json"
	"net/http"
	"workScheduler/internal/repository"
	"workScheduler/internal/scheduler/models"
)

type Api struct {
	RepoData repository.ReadWriteRepository
}

func NewApi(repo repository.ReadWriteRepository) *Api {
	return &Api{
		RepoData: repo,
	}
}

type ErrorStruct struct {
	Alternative []string `json:"alternative,omitempty"`
	ErrorCode   uint32   `json:"errorCode,omitempty"`
	Message     string   `json:"message,omitempty"`
}

func (a *Api) writeBadRequestError(w http.ResponseWriter, code uint32, message string) {
	w.WriteHeader(http.StatusBadRequest)
	err := ErrorStruct{
		Message:     message,
		ErrorCode:   code,
		Alternative: []string{},
	}
	errBytes, e := json.Marshal(err)
	if e != nil {
		w.Write([]byte(e.Error()))
	} else {
		w.Write(errBytes)
	}
}
func (a *Api) writeInternalError(w http.ResponseWriter, code uint32, message string, alternative []string) {
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
		a.writeBadRequestError(w, 400, err.Error())
	}

	// TODO: код проверки условий для добавления

	work, err = a.RepoData.Add(r.Context(), work)
	if err != nil {
		a.writeInternalError(w, 500, err.Error(), []string{})
		return
	}
	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, 500, err.Error(), []string{})
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(work_b)
}

func (a *Api) GetWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()

	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, 500, err.Error(), []string{})
		return
	}
	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, 500, err.Error(), []string{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) CancelWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()

	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, 500, err.Error(), []string{})
		return
	}

	work.Status = "cancelled"

	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, 500, err.Error(), []string{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) MoveWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()
	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, 500, err.Error(), []string{})
		return
	}

	// TODO: код для определения возможности переноса

	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, 500, err.Error(), []string{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}

func (a *Api) ProlongateWorkById(w http.ResponseWriter, r *http.Request, workId string) {
	defer r.Body.Close()
	work, err := a.RepoData.GetById(r.Context(), workId)
	if err != nil {
		a.writeInternalError(w, 500, err.Error(), []string{})
		return
	}

	// TODO: код для определения возможности переноса

	work_b, err := json.Marshal(work)
	if err != nil {
		a.writeInternalError(w, 500, err.Error(), []string{})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(work_b)
}
