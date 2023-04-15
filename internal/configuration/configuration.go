package configuration

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	WhiteList               []ZoneWindow         `yaml:"white_list"`
	BlackList               []ZoneWindow         `yaml:"black_list"`
	MinAvialableZones       int32                `yaml:"min_avialable_zones"`
	Pauses                  []Pause              `yaml:"pauses"`
	MinWorkDurationMinutes  WorkDurationSettings `yaml:"min_work_duration_minutes"`
	MaxWorkDurationMinutes  WorkDurationSettings `yaml:"max_work_duration_minutes"`
	MaxDeadlineDays         int32                `yaml:"max_deadline_days"`
	TimeCompressionPersents string               `yaml:"time_compression_persents"`
	TimeCompressionRate     float32
}

type ZoneWindow struct {
	ZoneId  string   `yaml:"zone"`
	Windows []Window `yaml:"windows"`
}

type Pause struct {
	ZoneId       string `yaml:"zone"`
	PauseMinutes int32  `yaml:"pause_minutes"`
}

type Window struct {
	StartTime     time.Time `yaml:"start_time"`
	DurationHours int32     `yaml:"duration"`
}

type WorkDurationSettings struct {
	Automatic int32 `yaml:"automatic"`
	Manual    int32 `yaml:"manual"`
}

func ReadConfig(configFile string) (config Config, err error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return
	}

	var validTimeCompressionPersents = regexp.MustCompile(`^(?P<num>[0-9]{1,2})%$`)
	if len(config.TimeCompressionPersents) != 0 {
		if matches := validTimeCompressionPersents.FindStringSubmatch(config.TimeCompressionPersents); len(matches) > 0 {
			subgroup := validTimeCompressionPersents.SubexpIndex("num")
			numValue, _ := strconv.ParseFloat(matches[subgroup], 32)
			config.TimeCompressionRate = float32(numValue) / float32(100)
		} else {
			err = fmt.Errorf("invalid time_compression_persents value in config: [%s]", config.TimeCompressionPersents)
			return
		}
	}

	return
}

//todo validate config??
