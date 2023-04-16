package models

import "time"

type WorkItem struct {
	Deadline        time.Time `bson:"deadline,omitempty" json:"deadline"`
	DurationMinutes int32     `bson:"durationMinutes,omitempty" json:"durationMinutes"`
	Id              string    `bson:"id,omitempty" json:"-"`
	WorkId          string    `bson:"workId,omitempty" json:"workId,omitempty"`
	Priority        string    `bson:"priority,omitempty" json:"priority,omitempty"`
	StartDate       time.Time `bson:"startDate,omitempty" json:"startDate,omitempty"`
	Status          string    `bson:"status,omitempty" json:"status,omitempty"`
	WorkType        string    `bson:"workType,omitempty" json:"workType"`
	Zones           []string  `bson:"zones,omitempty" json:"zones"`
	CompressionRate float64   `bson:"compressionRate,omitempty" json:"compressionRate,omitempty"`
}
