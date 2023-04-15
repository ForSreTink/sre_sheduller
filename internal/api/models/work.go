package models

import "time"

type WorkItem struct {
	Deadline        time.Time `bson:"deadline,omitempty"`
	DurationMinutes int32     `bson:"durationMinutes,omitempty"`
	Id              string    `bson:"workId,omitempty"`
	Priority        string    `bson:"priority,omitempty"`
	StartDate       time.Time `bson:"startDate,omitempty"`
	Status          string    `bson:"status,omitempty"`
	WorkType        string    `bson:"workType,omitempty"`
	Zone            string    `bson:"zone,omitempty"`
	CompressionRate float64   `bson:"compressionRate,omitempty"`
}
