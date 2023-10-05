package utils

const (
	// SecretVolumePath is the path of ALL the mounted secrets.
	SecretVolumePath = "/etc/secrets"
	// ConfigVolumePath is the path of the mounted Kafka config file.
	ConfigVolumePath = "/etc/config"
	// ConfigFileName is the name of the mounted Kafka config file.
	ConfigFileName = "kafka-config.yaml"
)
