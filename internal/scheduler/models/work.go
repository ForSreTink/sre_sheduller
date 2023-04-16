package models

import "time"

type WorkItem struct {
	Deadline        time.Time `bson:"deadline,omitempty" json:"deadline,omitempty"`
	DurationMinutes int32     `bson:"durationMinutes,omitempty" json:"durationMinutes,omitempty"`
	Id              string    `bson:"workId,omitempty" json:"Id"`
	Priority        string    `bson:"priority,omitempty" json:"priority"`
	StartDate       time.Time `bson:"startDate,omitempty" json:"startDate,omitempty"`
	Status          string    `bson:"status,omitempty" json:"status"`
	WorkType        string    `bson:"workType,omitempty" json:"workType,omitempty"`
	Zones           []string  `bson:"zones,omitempty" json:"zone,omitempty"`
	CompressionRate float64   `bson:"compressionRate,omitempty" json:"compressionRate"`
}
