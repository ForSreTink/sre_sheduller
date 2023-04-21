package main

import (
	"encoding/json"
	"fmt"
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
				minute := Minute{Value: fmt.Sprintf("%d", j)}
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

func main() {
	ts := time.Now()
	tmp := NewTeplateData(ts)
	vs, _ := json.Marshal(tmp)
	fmt.Println(string(vs))
}
