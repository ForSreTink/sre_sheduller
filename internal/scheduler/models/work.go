package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkItem struct {
	Deadline           time.Time          `bson:"deadline,omitempty" json:"deadline"`
	DurationMinutes    int32              `bson:"durationMinutes,omitempty" json:"durationMinutes"`
	Id                 primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	WorkId             string             `bson:"workId,omitempty" json:"workId"`
	Priority           string             `bson:"priority,omitempty" json:"priority"`
	StartDate          time.Time          `bson:"startDate,omitempty" json:"startDate"`
	Status             string             `bson:"status,omitempty" json:"status"`
	WorkType           string             `bson:"workType,omitempty" json:"workType"`
	Zones              []string           `bson:"zones,omitempty" json:"zones"`
	CompressionRate    float32            `bson:"compressionRate,omitempty" json:"compressionRate"`
	MaxCompressionRate float32            `bson:"maxCompressionRate,omitempty" json:"maxCompressionRate"`
	InitialDuration    int32              `bson:"initialDuration,omitempty" json:"-"`
	InitialStartDate   time.Time          `bson:"initialStartDate,omitempty" json:"-"`
}

func (w *WorkItem) EndTime() time.Time {
	return w.StartDate.Add(time.Duration(w.DurationMinutes) * time.Minute)
}

func (w WorkItem) String() string {
	return fmt.Sprintf(
		`WorkId: %s,
StartDate: %s,
Duration: %d,
Priority: %s,
WorkType: %s,
CompressionRate: %f,
Zones: %v,
Status: %s
`, w.WorkId, w.StartDate, w.DurationMinutes, w.Priority, w.WorkType, w.CompressionRate, w.Zones, w.Status)
}

func (w *WorkItem) CompressFromEnd(endTime time.Time, minDuration int32) bool {
	compressionRate := w.CompressionRate
	durationMinutes := w.DurationMinutes
	initialDuration := w.InitialDuration
	endDate := w.EndTime()
	for endDate.Unix() >= endTime.Unix() && compressionRate > w.MaxCompressionRate && durationMinutes >= minDuration {
		compressionRate = compressionRate - 0.01
		durationMinutes = int32(float32(initialDuration) * float32(compressionRate-0.01))
		deltaTime := (w.DurationMinutes - durationMinutes) * -1
		endDate = endDate.Add(time.Duration(deltaTime) * time.Minute)
	}
	if compressionRate < w.MaxCompressionRate || durationMinutes < minDuration {
		return false
	} else {
		w.DurationMinutes = durationMinutes
		w.CompressionRate = compressionRate
		return true
	}
}

func (w *WorkItem) CompressFromStart(startTime time.Time, minDuration int32) bool {
	compressionRate := w.CompressionRate
	durationMinutes := w.DurationMinutes
	initialDuration := w.InitialDuration
	startDate := w.StartDate
	for startDate.Unix() < startTime.Unix() && compressionRate > w.MaxCompressionRate && durationMinutes >= minDuration {
		compressionRate = compressionRate - 0.01
		durationMinutes = int32(float32(initialDuration) * float32(compressionRate-0.01))
		deltaTime := w.DurationMinutes - durationMinutes
		startDate.Add(time.Duration(deltaTime) * time.Minute)
	}
	if compressionRate <= w.MaxCompressionRate || durationMinutes < minDuration {
		return false
	} else {
		w.DurationMinutes = durationMinutes
		w.CompressionRate = compressionRate
		w.StartDate = startDate
		return true
	}
}

func (w *WorkItem) Uncompress() {
	w.CompressionRate = 0
	w.StartDate = w.InitialStartDate
	w.DurationMinutes = w.InitialDuration
}
