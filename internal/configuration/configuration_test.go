package configuration

import (
	"testing"
)

const (
	configFile = "../../config.yml"
)

func TestConfigSuccees(t *testing.T) {

	t.Run("succees read config", func(t *testing.T) {
		config, err := ReadConfig(configFile)
		if err != nil {
			t.Errorf("unable to read config file %s: %v", configFile, err)
		}
		if config.TimeCompressionRate != 0.60 {
			t.Errorf("wrong TimeCompressionRate in config file %s: want 0.60, got %v", configFile, config.TimeCompressionRate)
		}
		//todo check other fields automaitically
	})

}
