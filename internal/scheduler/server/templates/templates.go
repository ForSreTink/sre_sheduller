package templates

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
	"workScheduler/internal/repository"
)

type Template struct {
	Data repository.ReadRepository
}

type Minute struct {
	Value string
}

type Hour struct {
	Minutes []Minute
}

type Date struct {
	Hours []Hour
}

type TemplateData struct {
	WorkId string
	Date   map[string][]Hour
}

func NewTemplate(data repository.ReadRepository) *Template {
	return &Template{
		Data: data,
	}
}

func NewTeplateData(ts time.Time) TemplateData {
	tmpl := TemplateData{}
	days := []string{
		ts.AddDate(0, 0, -2).Format(time.DateOnly),
		ts.AddDate(0, 0, -1).Format(time.DateOnly),
		ts.Format(time.DateOnly),
		ts.AddDate(0, 0, 1).Format(time.DateOnly),
		ts.AddDate(0, 0, 2).Format(time.DateOnly),
	}

	tmpl.Date = make(map[string][]Hour)

	for _, day := range days {
		hours := []Hour{}
		for i := 0; i < 24; i++ {
			minutes := []Minute{}
			for j := 0; j < 60; j++ {
				minute := Minute{Value: "#000000"}
				minutes = append(minutes, minute)
			}
			hour := Hour{
				Minutes: minutes,
			}
			hours = append(hours, hour)
		}
		tmpl.Date[day] = hours
	}

	return tmpl
}

func (tmpl *Template) Generate(w http.ResponseWriter, r *http.Request) {
	ts := time.Now()
	startDate := ts.AddDate(0, 0, 2)
	endDate := ts.AddDate(0, 0, -2)
	tmpls := map[string][]TemplateData{}

	works, err := tmpl.Data.List(r.Context(), startDate, endDate, []string{}, []string{})
	if err != nil {
		log.Printf("ERROR: Can't generate template for '/' request, %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	for _, work := range works {
		for _, zone := range work.Zones {
			startDate := work.StartDate
			duration := work.DurationMinutes

			for i := duration; i > 0; i-- {
				day := startDate.Format(time.DateOnly)
				hour := startDate.Hour()
				minute := startDate.Minute()
				tmp := NewTeplateData(ts)
				if work.Status == "canceled" {
					tmp.Date[day][hour].Minutes[minute].Value = "#DCDCDC"
				} else if work.Status == "in_progress" {
					tmp.Date[day][hour].Minutes[minute].Value = "#7B68EE"
				} else if work.Status == "planned" {
					tmp.Date[day][hour].Minutes[minute].Value = "#00FA9A"
				} else {
					tmp.Date[day][hour].Minutes[minute].Value = "#32CD32"
				}
				tmpls[zone] = append(tmpls[zone], tmp)
				startDate = startDate.Add(1 * time.Minute)
			}
		}
		tmpl := template.Must(template.ParseFiles("template.html"))
		err := tmpl.Execute(w, tmpls)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		mrshl, err := json.Marshal(tmpl)
		fmt.Println(string(mrshl))
	}
}
