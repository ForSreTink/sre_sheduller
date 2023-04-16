package configuration

import (
	"context"
	"testing"
	"time"
)

const (
	configFile = "../../config.yml"
)

func TestConfigSuccees(t *testing.T) {

	t.Run("succees read config", func(t *testing.T) {
		ctx := context.Background()
		configurator := NewConfigurator(ctx, configFile)
		configurator.Run()
		time.Sleep(2 * time.Second)
		if configurator.Data.TimeCompressionRate != 0.60 {
			t.Errorf("wrong TimeCompressionRate in config file %s: want 0.60, got %v", configFile, configurator.Data.TimeCompressionRate)
		}
		//todo check other fields automaitically
	})

}
