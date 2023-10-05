package main

import (
	"context"
	"fmt"
	"os"

	"github.com/numaproj/numaflow-go/pkg/sourcer"

	"github.com/numaproj-contrib/kafka-source-go/pkg/config"
	"github.com/numaproj-contrib/kafka-source-go/pkg/kafka"
	"github.com/numaproj-contrib/kafka-source-go/pkg/utils"
)

func main() {
	logger := utils.NewLogger()
	// Get the config file path and format from env vars
	var format string
	format, ok := os.LookupEnv("CONFIG_FORMAT")
	if !ok {
		logger.Info("CONFIG_FORMAT not set, defaulting to yaml")
		format = "yaml"
	}

	var c *config.Config
	var err error
	c, err = getConfigFromEnvVars(format)
	if err != nil {
		c, err = getConfigFromFile(format)
		if err != nil {
			logger.Panic("Failed to parse config file : ", err)
		} else {
			logger.Info("Successfully parsed config file")
		}
	} else {
		logger.Info("Successfully parsed config from env vars")
	}

	logger.Info("Starting Kafka source...")
	kafkaSrc, err := kafka.New(c)
	if err != nil {
		logger.Panic("Failed to create kafka source : ", err)
	}
	defer kafkaSrc.Close()
	err = sourcer.NewServer(kafkaSrc).Start(context.Background())
	if err != nil {
		logger.Panic("Failed to start source server : ", err)
	}
}

func getConfigFromFile(format string) (*config.Config, error) {
	if format == "yaml" {
		parser := &config.YAMLConfigParser{}
		content, err := os.ReadFile(fmt.Sprintf("%s/%s", utils.ConfigVolumePath, utils.ConfigFileName))
		if err != nil {
			return nil, err
		}
		return parser.Parse(string(content))
	} else {
		return nil, fmt.Errorf("invalid config format %s", format)
	}
}

func getConfigFromEnvVars(format string) (*config.Config, error) {
	var c string
	c, ok := os.LookupEnv("KAFKA_CONFIG")
	if !ok {
		return nil, fmt.Errorf("KAFKA_CONFIG environment variable is not set")
	}
	if format == "yaml" {
		parser := &config.YAMLConfigParser{}
		return parser.Parse(c)
	} else {
		return nil, fmt.Errorf("invalid config format %s", format)
	}
}
