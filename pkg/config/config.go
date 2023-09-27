package config

import corev1 "k8s.io/api/core/v1"

type Config struct {
	Brokers           []string `json:"brokers,omitempty" protobuf:"bytes,1,rep,name=brokers"`
	Topic             string   `json:"topic" protobuf:"bytes,2,opt,name=topic"`
	ConsumerGroupName string   `json:"consumerGroup,omitempty" protobuf:"bytes,3,opt,name=consumerGroup"`
	// TLS user to configure TLS connection for kafka broker
	// TLS.enable=true default for TLS.
	// +optional
	TLS *TLS `json:"tls" protobuf:"bytes,4,opt,name=tls"`
	// +optional
	Config string `json:"config,omitempty" protobuf:"bytes,5,opt,name=config"`
	// SASL user to configure SASL connection for kafka broker
	// SASL.enable=true default for SASL.
	// +optional
	SASL *SASL `json:"sasl" protobuf:"bytes,6,opt,name=sasl"`
}

type TLS struct {
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty" protobuf:"bytes,1,opt,name=insecureSkipVerify"`
	// CACertSecret refers to the secret that contains the CA cert
	// +optional
	CACertSecret *corev1.SecretKeySelector `json:"caCertSecret,omitempty" protobuf:"bytes,2,opt,name=caCertSecret"`
	// CertSecret refers to the secret that contains the cert
	// +optional
	CertSecret *corev1.SecretKeySelector `json:"clientCertSecret,omitempty" protobuf:"bytes,3,opt,name=certSecret"`
	// KeySecret refers to the secret that contains the key
	// +optional
	KeySecret *corev1.SecretKeySelector `json:"clientKeySecret,omitempty" protobuf:"bytes,4,opt,name=keySecret"`
}

type SASL struct {
	// SASL mechanism to use
	Mechanism *SASLType `json:"mechanism" protobuf:"bytes,1,opt,name=mechanism,casttype=SASLType"`
	// GSSAPI contains the kerberos config
	// +optional
	GSSAPI *GSSAPI `json:"gssapi" protobuf:"bytes,2,opt,name=gssapi"`
	// SASLPlain contains the sasl plain config
	// +optional
	Plain *SASLPlain `json:"plain" protobuf:"bytes,3,opt,name=plain"`
}

// SASLType describes the SASL type
type SASLType string

const (
	// SASLTypeOAuth represents the SASL/OAUTHBEARER mechanism (Kafka 2.0.0+)
	// SASLTypeOAuth = "OAUTHBEARER"
	SASLTypeOAuth SASLType = "OAUTHBEARER"
	// SASLTypePlaintext represents the SASL/PLAIN mechanism
	// SASLTypePlaintext = "PLAIN"
	SASLTypePlaintext SASLType = "PLAIN"
	// SASLTypeSCRAMSHA256 represents the SCRAM-SHA-256 mechanism.
	// SASLTypeSCRAMSHA256 = "SCRAM-SHA-256"
	SASLTypeSCRAMSHA256 SASLType = "SCRAM-SHA-256"
	// SASLTypeSCRAMSHA512 represents the SCRAM-SHA-512 mechanism.
	// SASLTypeSCRAMSHA512 = "SCRAM-SHA-512"
	SASLTypeSCRAMSHA512 SASLType = "SCRAM-SHA-512"
	// SASLTypeGSSAPI represents the GSSAPI mechanism
	// SASLTypeGSSAPI      = "GSSAPI"
	SASLTypeGSSAPI SASLType = "GSSAPI"
)

// GSSAPI represents a SASL GSSAPI config
type GSSAPI struct {
	ServiceName string `json:"serviceName" protobuf:"bytes,1,opt,name=serviceName"`
	Realm       string `json:"realm" protobuf:"bytes,2,opt,name=realm"`
	// UsernameSecret refers to the secret that contains the username
	UsernameSecret *corev1.SecretKeySelector `json:"usernameSecret" protobuf:"bytes,3,opt,name=usernameSecret"`
	// valid inputs - KRB5_USER_AUTH, KRB5_KEYTAB_AUTH
	AuthType *KRB5AuthType `json:"authType" protobuf:"bytes,4,opt,name=authType,casttype=KRB5AuthType"`
	// PasswordSecret refers to the secret that contains the password
	// +optional
	PasswordSecret *corev1.SecretKeySelector `json:"passwordSecret,omitempty" protobuf:"bytes,5,opt,name=passwordSecret"`
	// KeytabSecret refers to the secret that contains the keytab
	// +optional
	KeytabSecret *corev1.SecretKeySelector `json:"keytabSecret,omitempty" protobuf:"bytes,6,opt,name=keytabSecret"`
	// KerberosConfigSecret refers to the secret that contains the kerberos config
	// +optional
	KerberosConfigSecret *corev1.SecretKeySelector `json:"kerberosConfigSecret,omitempty" protobuf:"bytes,7,opt,name=kerberosConfigSecret"`
}

// KRB5AuthType describes the kerberos auth type
// +enum
type KRB5AuthType string

const (
	// KRB5UserAuth represents the password method
	// KRB5UserAuth = "KRB5_USER_AUTH" = 1
	KRB5UserAuth KRB5AuthType = "KRB5_USER_AUTH"
	// KRB5KeytabAuth represents the password method
	// KRB5KeytabAuth = "KRB5_KEYTAB_AUTH" = 2
	KRB5KeytabAuth KRB5AuthType = "KRB5_KEYTAB_AUTH"
)

type SASLPlain struct {
	// UserSecret refers to the secret that contains the user
	UserSecret *corev1.SecretKeySelector `json:"userSecret" protobuf:"bytes,1,opt,name=userSecret"`
	// PasswordSecret refers to the secret that contains the password
	// +optional
	PasswordSecret *corev1.SecretKeySelector `json:"passwordSecret" protobuf:"bytes,2,opt,name=passwordSecret"`
	Handshake      bool                      `json:"handshake" protobuf:"bytes,3,opt,name=handshake"`
}
