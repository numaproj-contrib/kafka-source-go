package utils

import (
	"fmt"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// TODO - Unit Test this

// VolumeReader is an interface that defines methods to interact with secret volumes in a Kubernetes environment.
type VolumeReader interface {
	GetSecretFromVolume(selector *corev1.SecretKeySelector) (string, error)
	GetSecretVolumePath(selector *corev1.SecretKeySelector) (string, error)
}

// KafkaVolumeReader is a utility struct for reading secret volumes.
type KafkaVolumeReader struct {
	secretPath string
}

// NewKafkaVolumeReader creates a new KafkaVolumeReader with the specified secret path.
func NewKafkaVolumeReader(secretPath string) *KafkaVolumeReader {
	return &KafkaVolumeReader{
		secretPath: secretPath,
	}
}

// GetSecretFromVolume retrieves the value from a mounted secret volume. It trims any newline suffix from the value.
func (nvr *KafkaVolumeReader) GetSecretFromVolume(selector *corev1.SecretKeySelector) (string, error) {
	if selector == nil {
		return "", fmt.Errorf("secret key selector is nil")
	}
	filePath, err := nvr.GetSecretVolumePath(selector)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read secret (name: %s, key: %s): %w", selector.Name, selector.Key, err)
	}
	return strings.TrimSuffix(string(data), "\n"), nil
}

// GetSecretVolumePath constructs and returns the path of a mounted secret based on the secret key selector.
func (nvr *KafkaVolumeReader) GetSecretVolumePath(selector *corev1.SecretKeySelector) (string, error) {
	if selector == nil {
		return "", fmt.Errorf("secret key selector is nil")
	}
	return fmt.Sprintf("%s/%s/%s", nvr.secretPath, selector.Name, selector.Key), nil
}
