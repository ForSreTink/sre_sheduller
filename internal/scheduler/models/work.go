package models

import "time"

type WorkItem struct {
	Deadline        time.Time `bson:"deadline,omitempty" json:"deadline"`
	DurationMinutes int32     `bson:"durationMinutes,omitempty" json:"durationMinutes"`
	Id              string    `bson:"workId,omitempty" json:"Id"`
	Priority        string    `bson:"priority,omitempty" json:"priority"`
	StartDate       time.Time `bson:"startDate,omitempty" json:"startDate"`
	Status          string    `bson:"status,omitempty" json:"status"`
	WorkType        string    `bson:"workType,omitempty" json:"workType"`
	Zone            string    `bson:"zone,omitempty" json:"zone"`
	CompressionRate float64   `bson:"compressionRate,omitempty" json:"compressionRate"`
}
