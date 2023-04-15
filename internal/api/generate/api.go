package api

import "net/http"

type Api struct {
}

func NewApi() *Api {
	return &Api{}
}

func (a *Api) GetShedule(w http.ResponseWriter, r *http.Request, params GetSheduleParams) {

}

func (a *Api) AddWork(w http.ResponseWriter, r *http.Request) {

}
func (a *Api) GetWorkById(w http.ResponseWriter, r *http.Request, workId string) {

}
func (a *Api) CancelWorkById(w http.ResponseWriter, r *http.Request, workId string) {

}

func (a *Api) MoveWorkById(w http.ResponseWriter, r *http.Request, workId string) {

}

func (a *Api) ProlongateWorkById(w http.ResponseWriter, r *http.Request, workId string) {

}
