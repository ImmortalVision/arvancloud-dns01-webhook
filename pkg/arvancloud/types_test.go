package arvancloud

import "testing"

func TestSolverConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     solverConfig
		wantErr bool
	}{
		{
			name: "valid with defaults",
			cfg: solverConfig{
				APIKeySecretRef: secretKeySelector{Name: "sec", Key: "api-key"},
			},
		},
		{
			name: "missing secret name",
			cfg: solverConfig{
				APIKeySecretRef: secretKeySelector{Key: "api-key"},
			},
			wantErr: true,
		},
		{
			name: "missing secret key",
			cfg: solverConfig{
				APIKeySecretRef: secretKeySelector{Name: "sec"},
			},
			wantErr: true,
		},
		{
			name: "invalid ttl",
			cfg: solverConfig{
				APIKeySecretRef: secretKeySelector{Name: "sec", Key: "api-key"},
				TTL:             121,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.cfg.TTL != defaultTTL {
				t.Fatalf("default TTL = %d, want %d", tt.cfg.TTL, defaultTTL)
			}
			if tt.cfg.APIEndpoint != defaultAPIEndpoint {
				t.Fatalf("default API endpoint = %q, want %q", tt.cfg.APIEndpoint, defaultAPIEndpoint)
			}
		})
	}
}
