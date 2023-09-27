package utils

import (
	"fmt"
	"os"

	"github.com/IBM/sarama"

	"github.com/numaproj-contrib/kafka-source-go/pkg/config"
)

// GetSASL is a utility function to get sarama.Config.Net.SASL
func GetSASL(saslConfig *config.SASL, reader VolumeReader) (*struct {
	Enable                   bool
	Mechanism                sarama.SASLMechanism
	Version                  int16
	Handshake                bool
	AuthIdentity             string
	User                     string
	Password                 string
	SCRAMAuthzID             string
	SCRAMClientGeneratorFunc func() sarama.SCRAMClient
	TokenProvider            sarama.AccessTokenProvider
	GSSAPI                   sarama.GSSAPIConfig
}, error) {
	c := sarama.NewConfig()
	switch *saslConfig.Mechanism {
	case config.SASLTypeGSSAPI:
		if gssapi := saslConfig.GSSAPI; gssapi != nil {
			c.Net.SASL.Enable = true
			c.Net.SASL.Mechanism = sarama.SASLTypeGSSAPI
			if gssapi, err := GetGSSAPIConfig(gssapi, reader); err != nil {
				return nil, fmt.Errorf("error loading gssapi config, %w", err)
			} else {
				c.Net.SASL.GSSAPI = *gssapi
			}
		}
	case config.SASLTypePlaintext:
		if plain := saslConfig.Plain; plain != nil {
			c.Net.SASL.Enable = true
			c.Net.SASL.Mechanism = sarama.SASLTypePlaintext
			if plain.UserSecret != nil {
				user, err := reader.GetSecretFromVolume(plain.UserSecret)
				if err != nil {
					return nil, err
				} else {
					c.Net.SASL.User = user
				}
			}
			if plain.PasswordSecret != nil {
				password, err := reader.GetSecretFromVolume(plain.PasswordSecret)
				if err != nil {
					return nil, err
				} else {
					c.Net.SASL.Password = password
				}
			}
			c.Net.SASL.Handshake = plain.Handshake
		}
	default:
		return nil, fmt.Errorf("SASL mechanism not supported: %s", *saslConfig.Mechanism)
	}
	return &c.Net.SASL, nil
}

// GetGSSAPIConfig A utility function to get sasl.gssapi.Config
func GetGSSAPIConfig(gssapiConfig *config.GSSAPI, reader VolumeReader) (*sarama.GSSAPIConfig, error) {
	if gssapiConfig == nil {
		return nil, nil
	}

	c := &sarama.GSSAPIConfig{
		ServiceName: gssapiConfig.ServiceName,
		Realm:       gssapiConfig.Realm,
	}

	switch *gssapiConfig.AuthType {
	case config.KRB5UserAuth:
		c.AuthType = sarama.KRB5_USER_AUTH
	case config.KRB5KeytabAuth:
		c.AuthType = sarama.KRB5_KEYTAB_AUTH
	default:
		return nil, fmt.Errorf("failed to parse GSSAPI AuthType %v. Must be one of the following: ['KRB5_USER_AUTH', 'KRB5_KEYTAB_AUTH']", gssapiConfig.AuthType)
	}

	if us := gssapiConfig.UsernameSecret; us != nil {
		username, err := reader.GetSecretFromVolume(us)
		if err != nil {
			return nil, err
		} else {
			c.Username = username
		}
	}

	if ks := gssapiConfig.KeytabSecret; ks != nil {
		keyTabPath, err := reader.GetSecretVolumePath(ks)
		if err != nil {
			return nil, err
		}
		if len(keyTabPath) > 0 {
			_, err := os.ReadFile(keyTabPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read keytab file %s, %w", keyTabPath, err)
			}
		}
		c.KeyTabPath = keyTabPath
	}

	if kcs := gssapiConfig.KerberosConfigSecret; kcs != nil {
		kerberosConfigPath, err := reader.GetSecretVolumePath(kcs)
		if err != nil {
			return nil, err
		}
		if len(kerberosConfigPath) > 0 {
			_, err := os.ReadFile(kerberosConfigPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read kerberos config file %s, %w", kerberosConfigPath, err)
			}
		}
		c.KerberosConfigPath = kerberosConfigPath
	}

	if ps := gssapiConfig.PasswordSecret; ps != nil {
		password, err := reader.GetSecretFromVolume(ps)
		if err != nil {
			return nil, err
		} else {
			c.Password = password
		}
	}
	return c, nil
}
