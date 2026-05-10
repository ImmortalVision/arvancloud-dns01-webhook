package arvancloud

import "fmt"

const (
	defaultAPIEndpoint = "https://napi.arvancloud.ir/cdn/4.0"
	defaultTTL         = 120
)

var allowedTTLs = map[int]struct{}{
	120: {}, 180: {}, 300: {}, 600: {}, 900: {}, 1800: {},
	3600: {}, 7200: {}, 18000: {}, 43200: {}, 86400: {},
	172800: {}, 432000: {},
}

type solverConfig struct {
	APIKeySecretRef secretKeySelector `json:"apiKeySecretRef"`
	APIEndpoint     string            `json:"apiEndpoint,omitempty"`
	Zone            string            `json:"zone,omitempty"`
	TTL             int               `json:"ttl,omitempty"`
}

type secretKeySelector struct {
	Name      string `json:"name"`
	Key       string `json:"key"`
	Namespace string `json:"namespace,omitempty"`
}

func (c *solverConfig) validate() error {
	if c.APIKeySecretRef.Name == "" {
		return fmt.Errorf("config.apiKeySecretRef.name is required")
	}
	if c.APIKeySecretRef.Key == "" {
		return fmt.Errorf("config.apiKeySecretRef.key is required")
	}
	if c.TTL == 0 {
		c.TTL = defaultTTL
	}
	if _, ok := allowedTTLs[c.TTL]; !ok {
		return fmt.Errorf("config.ttl=%d is invalid for ArvanCloud API", c.TTL)
	}
	if c.APIEndpoint == "" {
		c.APIEndpoint = defaultAPIEndpoint
	}
	return nil
}
