package configuration

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
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
	TimeCompressionPercents string               `yaml:"time_compression_percents"`
	TimeCompressionRate     float32
}

type Window struct {
	StartHour uint32 `yaml:"start_hour"`
	EndHour   uint32 `yaml:"end_hour"`
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

func (c *Configurator) Run() {
	c.updateConfig()
	go c.mainProcess()
}

func (c *Configurator) mainProcess() {
	path := filepath.Dir(c.ConfigPath)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	err = watcher.Add(path)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) {
				if event.Name == c.ConfigPath {
					c.updateConfig()
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("WARNING: %s", err)
		case <-c.Ctx.Done():
			return
		}
	}
}

func (c *Configurator) updateConfig() {
	log.Println("Config processing...")
	c.Mu.Lock()
	defer c.Mu.Unlock()
	conf, err := c.readConfig()
	if err != nil && !c.Started {
		log.Fatal(err)
	} else if err != nil {
		log.Printf("WARNING: Given config is invalid, config update ignoring: %s", err)
		return
	}
	err = c.validateConfig(conf)
	if err != nil && !c.Started {
		log.Fatal(err)
	} else if err != nil {
		log.Printf("WARNING: Given config is invalid, config update ignoring: %s", err)
		return
	} else {
		c.Data = &conf
		c.Started = true
		log.Println("Configuration updated succesfully!")
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
	ts := time.Now()
	zones := make(map[string]time.Time)
	for name, zone := range conf.WhiteList {
		zones[name] = ts
		for _, interval := range zone {
			if interval.StartHour >= 24 {
				errStr += "start_hour must be uint in range from 0 to 23; "
			}
			if interval.EndHour >= 24 {
				errStr += "EndHour must be uint in range from 1 to 23; "
			}
			if interval.StartHour >= interval.EndHour {
				errStr += "EndHour must greater then StartHour.; "
			}
		}
	}
	for _, name := range conf.BlackList {
		if _, ok := zones[name]; ok {
			errStr += "Zone from black list in white list; "
		} else {
			zones[name] = ts
		}
	}

	if conf.MinAvialableZones > int32(len(conf.WhiteList)-2) || conf.MinAvialableZones < 2 {
		errStr += "MinAvialableZones must be greater then zone count, min 2; "
	}

	for k, v := range conf.PausesMinutes {
		if _, ok := zones[k]; !ok {
			errStr += fmt.Sprintf("%s zone in pauses not found in zone (black/white list); ", k)
		}
		if v < 0 || v > 60 {
			errStr += "Pauses value must be in range from 0 to 60"
		}
	}

	if conf.MinWorkDurationMinutes.Automatic < 5 {
		errStr += "min_work_duration_minutes duration value for automatic works must be greater then 5; "
	}

	if conf.MinWorkDurationMinutes.Manual < 30 {
		errStr += "min_work_duration_minutes duration value for manual works must be greater then 30;"
	}

	if conf.MinWorkDurationMinutes.Automatic > conf.MaxWorkDurationMinutes.Automatic {
		errStr += "min_work_duration_minutes duration value for automatic can't be greater then max_work_duration_minutes;"
	}

	if conf.MinWorkDurationMinutes.Manual > conf.MaxWorkDurationMinutes.Manual {
		errStr += "min_work_duration_minutes duration value for manual can't be greater then max_work_duration_minutes;"
	}

	if conf.MaxDeadlineDays <= 0 {
		errStr += "max_deadline_days duration value must be greater then 0;"
	}

	var validTimeCompressionPercents = regexp.MustCompile(`^(?P<num>[0-9]{1,2})%$`)
	if len(conf.TimeCompressionPercents) != 0 {
		if matches := validTimeCompressionPercents.FindStringSubmatch(conf.TimeCompressionPercents); len(matches) > 0 {
			subgroup := validTimeCompressionPercents.SubexpIndex("num")
			numValue, _ := strconv.ParseFloat(matches[subgroup], 32)
			conf.TimeCompressionRate = float32(numValue) / float32(100)
		} else {
			errStr += fmt.Sprintf("invalid time_compression_persents value in config: [%s]", conf.TimeCompressionPercents)
		}
	}

	if errStr != "" {
		return errors.New(errStr)
	} else {
		return nil
	}
}
