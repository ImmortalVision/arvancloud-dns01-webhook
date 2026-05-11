package arvancloud

import (
	"encoding/json"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestRecordNameFromFQDN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fqdn    string
		zone    string
		want    string
		wantErr bool
	}{
		{name: "apex", fqdn: "example.com.", zone: "example.com", want: "@"},
		{name: "subdomain", fqdn: "_acme-challenge.example.com.", zone: "example.com", want: "_acme-challenge"},
		{name: "nested", fqdn: "_acme.foo.example.com.", zone: "example.com.", want: "_acme.foo"},
		{name: "case insensitive", fqdn: "_ACME.Example.com.", zone: "example.com", want: "_ACME"},
		{name: "zone mismatch", fqdn: "_acme.example.net.", zone: "example.com", wantErr: true},
		{name: "empty fqdn", fqdn: "", zone: "example.com", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := recordNameFromFQDN(tt.fqdn, tt.zone)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("recordNameFromFQDN() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeAuthorizationHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "my-token", want: "APIKEY my-token"},
		{in: "APIKEY abc", want: "APIKEY abc"},
		{in: "api key xyz", want: "APIKEY xyz"},
		{in: "API KEY xyz", want: "APIKEY xyz"},
		{in: "  APIKEY   xyz  ", want: "APIKEY xyz"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			if got := normalizeAuthorizationHeader(tt.in); got != tt.want {
				t.Fatalf("normalizeAuthorizationHeader() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSecretValue(t *testing.T) {
	t.Parallel()

	secret := &corev1.Secret{
		Data: map[string][]byte{
			"api-key": []byte("  value  "),
			"empty":   []byte("   "),
		},
	}

	if _, err := secretValue(secret, "missing"); err == nil {
		t.Fatalf("expected missing key error")
	}
	if _, err := secretValue(secret, "empty"); err == nil {
		t.Fatalf("expected empty value error")
	}

	got, err := secretValue(secret, "api-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "value" {
		t.Fatalf("secretValue() = %q, want %q", got, "value")
	}
}

func TestTXTRecordHelpers(t *testing.T) {
	t.Parallel()

	okValue, _ := json.Marshal(txtValue{Text: "token"})
	emptyValue, _ := json.Marshal(txtValue{Text: ""})

	records := []dnsRecord{
		{Type: "TXT", Name: "_acme-challenge", Value: okValue},
		{Type: "TXT", Name: "_acme-other", Value: emptyValue},
		{Type: "A", Name: "_acme-challenge"},
	}

	if !sameRecordName("_acme-challenge.", "_acme-challenge") {
		t.Fatalf("sameRecordName should ignore trailing dot")
	}
	if !hasTXT(records, "_acme-challenge", "token") {
		t.Fatalf("expected hasTXT to match token")
	}
	if hasTXT(records, "_acme-challenge", "different") {
		t.Fatalf("expected hasTXT to reject different token")
	}
}
