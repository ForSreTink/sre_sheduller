package configuration

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Configurator struct {
	ConfigPath string
	Data       *Config
	Ctx        context.Context
	Mu         *sync.Mutex
	Started    bool
}

type Config struct {
	WhiteList               map[string][]Window  `yaml:"white_list"`
	BlackList               []string             `yaml:"black_list"`
	MinAvialableZones       int32                `yaml:"min_avialable_zones"`
	PausesMinutes           map[string]int32     `yaml:"pauses"`
	MinWorkDurationMinutes  WorkDurationSettings `yaml:"min_work_duration_minutes"`
	MaxWorkDurationMinutes  WorkDurationSettings `yaml:"max_work_duration_minutes"`
	MaxDeadlineDays         int32                `yaml:"max_deadline_days"`
	TimeCompressionPersents string               `yaml:"time_compression_persents"`
	TimeCompressionRate     float32
}

type Window struct {
	StartTime     uint32 `yaml:"start_hour"`
	DurationHours uint32 `yaml:"end_hour"`
}

type WorkDurationSettings struct {
	Automatic int32 `yaml:"automatic"`
	Manual    int32 `yaml:"manual"`
}

func NewConfigurator(ctx context.Context, filepath string) *Configurator {
	return &Configurator{
		ConfigPath: filepath,
		Data:       &Config{},
		Ctx:        ctx,
		Mu:         &sync.Mutex{},
		Started:    false,
	}
}

func (c *Configurator) Run() error {

	return nil
}

func (c *Configurator) mainProcess() {
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-c.Ctx.Done():
			return
		case <-ticker.C:
			conf, err := c.readConfig()
			if err != nil && !c.Started {
				log.Fatal(err)
			} else if err != nil {
				log.Print("WARNING: Given config is invalid, config update ignoring")
				continue
			}
			err = c.validateConfig(conf)
			if err != nil && !c.Started {
				log.Fatal(err)
			} else if err != nil {
				log.Print("WARNING: Given valid is invalid, config update ignoring")
				continue
			} else {
				c.Data = &conf
				c.Started = true
			}
		}
	}
}

func (c *Configurator) readConfig() (config Config, err error) {
	file, err := os.ReadFile(c.ConfigPath)
	if err != nil {
		return
	}
	config.WhiteList = make(map[string][]Window)
	config.PausesMinutes = make(map[string]int32)
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return
	}

	return
}

func (c *Configurator) validateConfig(conf Config) error {
	errStr := ""
	for _, v := range conf.WhiteList {
		if time.
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
}

//todo validate config??
