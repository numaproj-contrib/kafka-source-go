package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigParser_UnParseThenParse(t *testing.T) {
	var parsers = []Parser{
		&YAMLConfigParser{},
	}
	for _, parser := range parsers {
		testConfig := &Config{
			Brokers:           []string{"kafka-broker:9092"},
			Topic:             "test-topic",
			ConsumerGroupName: "test-consumer-group",
		}
		configStr, err := parser.UnParse(testConfig)
		assert.NoError(t, err)
		println(configStr)
		config, err := parser.Parse(configStr)
		assert.NoError(t, err)
		assert.Equal(t, testConfig, config)
	}
}

func TestConfigParser_ParseErrScenarios(t *testing.T) {
	var parsers = []Parser{
		&YAMLConfigParser{},
	}
	for _, parser := range parsers {
		_, err := parser.Parse("invalid config string")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "failed to parse config string"))
		_, err = parser.UnParse(nil)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "config cannot be nil"))
	}
}
