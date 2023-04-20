package configuration

import (
	"context"
	"testing"
	"time"
)

const (
	unitTestConfigName = "../scheduler/test_configs/scheduler_unit_config.yml"
)

func TestConfigSuccees(t *testing.T) {

	t.Run("succees read config", func(t *testing.T) {
		ctx := context.Background()
		configurator := NewConfigurator(ctx, unitTestConfigName)
		configurator.Run()
		time.Sleep(2 * time.Second)
		if configurator.Data.TimeCompressionRate != 0.90 {
			t.Errorf("wrong TimeCompressionRate in config file %s: want 0.60, got %v", unitTestConfigName, configurator.Data.TimeCompressionRate)
		}
		//todo check other fields automaitically
	})

}
