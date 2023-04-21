package models

import (
	"fmt"
	"time"
	"workScheduler/internal/configuration"
)

type WorkItem struct {
	Deadline         time.Time `bson:"deadline,omitempty" json:"deadline"`
	DurationMinutes  int32     `bson:"durationMinutes,omitempty" json:"durationMinutes"`
	Id               string    `bson:"id,omitempty" json:"-"`
	WorkId           string    `bson:"workId,omitempty" json:"workId,omitempty"`
	Priority         string    `bson:"priority,omitempty" json:"priority,omitempty"`
	StartDate        time.Time `bson:"startDate,omitempty" json:"startDate,omitempty"`
	Status           string    `bson:"status,omitempty" json:"status,omitempty"`
	WorkType         string    `bson:"workType,omitempty" json:"workType"`
	Zones            []string  `bson:"zones,omitempty" json:"zones"`
	CompressionRate  float32   `bson:"compressionRate,omitempty" json:"compressionRate,omitempty"`
	InitialDuration  int32     `bson:"initialDuration,omitempty" json:"initialDuration,omitempty"`
	InitialStartDate time.Time `bson:"initialStartDate,omitempty" json:"initialStartDate,omitempty"`
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

func (w *WorkItem) CompressFromEnd(maxCompressionRate float32, endTime time.Time, minDuration int32) bool {
	compressionRate := w.CompressionRate
	durationMinutes := w.DurationMinutes
	initialDuration := w.InitialDuration
	endDate := w.EndTime()
	for endDate.Unix() >= endTime.Unix() && compressionRate > maxCompressionRate && durationMinutes >= minDuration {
		compressionRate = compressionRate - 0.01
		durationMinutes = int32(float32(initialDuration) * float32(compressionRate-0.01))
		deltaTime := (w.DurationMinutes - durationMinutes) * -1
		endDate = endDate.Add(time.Duration(deltaTime) * time.Minute)
	}
	if compressionRate <= maxCompressionRate || durationMinutes < minDuration {
		return false
	} else {
		w.DurationMinutes = durationMinutes
		w.CompressionRate = compressionRate
		return true
	}
}

func (w *WorkItem) CompressFromStart(maxCompressionRate float32, startTime time.Time, minDuration int32) bool {
	compressionRate := w.CompressionRate
	durationMinutes := w.DurationMinutes
	initialDuration := w.InitialDuration
	startDate := w.StartDate
	for startDate.Unix() < startTime.Unix() && compressionRate > maxCompressionRate && durationMinutes >= minDuration {
		compressionRate = compressionRate - 0.01
		durationMinutes = int32(float32(initialDuration) * float32(compressionRate-0.01))
		deltaTime := w.DurationMinutes - durationMinutes
		startDate.Add(time.Duration(deltaTime) * time.Minute)
	}
	if compressionRate <= maxCompressionRate || durationMinutes < minDuration {
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

func (w *WorkItem) SetNextPossibleStartDateInInterval(dateVariant time.Time, intervals []configuration.Window) bool {
	startHour := dateVariant.Hour()
	endHour := dateVariant.Add(time.Duration(w.DurationMinutes) * time.Minute).Hour()

	if endHour < startHour {
		endHour = 23
	}

	inInterval := false

	for _, interval := range intervals {
		if startHour >= int(interval.StartHour) && endHour < (int(interval.EndHour)) {
			inInterval = true
		}
	}

	fmt.Println(inInterval)

	if inInterval == true {
		w.StartDate = dateVariant
		return true
	}

	durationInHours := uint32(endHour - startHour)
	newStartHour := uint32(0)

	for _, interval := range intervals {
		deltaIntervalHours := interval.EndHour - interval.StartHour

		if durationInHours < deltaIntervalHours {
			if newStartHour <= interval.StartHour {
				newStartHour = interval.StartHour
				inInterval = true
			}
		}
	}

	if inInterval == true {
		deltaHour := int(newStartHour) - dateVariant.Hour()
		if deltaHour < 0 {
			deltaHour *= -1
		}
		dateVariant = dateVariant.Add(time.Duration(deltaHour) * time.Hour)

		w.StartDate = dateVariant
		return true
	} else {
		return false
	}
}
