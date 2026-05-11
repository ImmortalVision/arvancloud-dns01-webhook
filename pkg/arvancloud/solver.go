package arvancloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	v1alpha1 "github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type Solver struct {
	client     kubernetes.Interface
	httpClient *http.Client
}

func (s *Solver) Name() string {
	return "arvancloud"
}

func (s *Solver) Initialize(kubeClientConfig *rest.Config, _ <-chan struct{}) error {
	client, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}

	s.client = client
	s.httpClient = &http.Client{Timeout: 15 * time.Second}

	return nil
}

func (s *Solver) Present(ch *v1alpha1.ChallengeRequest) error {
	ctx := context.Background()

	cfg, err := loadConfig(ch)
	if err != nil {
		return err
	}

	apiKey, err := s.loadAPIKey(ctx, ch.ResourceNamespace, cfg.APIKeySecretRef)
	if err != nil {
		return err
	}

	zone := strings.TrimSuffix(cfg.Zone, ".")
	if zone == "" {
		zone = strings.TrimSuffix(ch.ResolvedZone, ".")
	}
	if zone == "" {
		return fmt.Errorf("zone is empty; provide config.zone or ensure challenge has resolvedZone")
	}

	name, err := recordNameFromFQDN(ch.ResolvedFQDN, zone)
	if err != nil {
		return err
	}

	c := newClient(cfg.APIEndpoint, normalizeAuthorizationHeader(apiKey), s.httpClient)

	records, err := c.ListTXTRecords(ctx, zone, name)
	if err != nil {
		return err
	}

	if hasTXT(records, name, ch.Key) {
		klog.Infof("TXT record already exists for %s in zone %s", name, zone)
		return nil
	}

	if err := c.CreateTXTRecord(ctx, zone, name, ch.Key, cfg.TTL); err != nil {
		return err
	}

	klog.Infof("created TXT record %s in zone %s", name, zone)
	return nil
}

func (s *Solver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	ctx := context.Background()

	cfg, err := loadConfig(ch)
	if err != nil {
		return err
	}

	apiKey, err := s.loadAPIKey(ctx, ch.ResourceNamespace, cfg.APIKeySecretRef)
	if err != nil {
		return err
	}

	zone := strings.TrimSuffix(cfg.Zone, ".")
	if zone == "" {
		zone = strings.TrimSuffix(ch.ResolvedZone, ".")
	}
	if zone == "" {
		return fmt.Errorf("zone is empty; provide config.zone or ensure challenge has resolvedZone")
	}

	name, err := recordNameFromFQDN(ch.ResolvedFQDN, zone)
	if err != nil {
		return err
	}

	c := newClient(cfg.APIEndpoint, normalizeAuthorizationHeader(apiKey), s.httpClient)

	records, err := c.ListTXTRecords(ctx, zone, name)
	if err != nil {
		return err
	}

	for _, record := range records {
		if !isTXT(record) || !sameRecordName(record.Name, name) {
			continue
		}

		text, ok := txtText(record)
		if !ok || text != ch.Key {
			continue
		}

		if err := c.DeleteRecord(ctx, zone, record.ID); err != nil {
			return err
		}
		klog.Infof("deleted TXT record %s (%s) in zone %s", name, record.ID, zone)
	}

	return nil
}

func loadConfig(ch *v1alpha1.ChallengeRequest) (*solverConfig, error) {
	if ch.Config == nil {
		return nil, fmt.Errorf("missing webhook solver config")
	}

	var cfg solverConfig
	if err := json.Unmarshal(ch.Config.Raw, &cfg); err != nil {
		return nil, fmt.Errorf("decode webhook solver config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (s *Solver) loadAPIKey(ctx context.Context, fallbackNamespace string, ref secretKeySelector) (string, error) {
	ns := ref.Namespace
	if ns == "" {
		ns = fallbackNamespace
	}

	secret, err := s.client.CoreV1().Secrets(ns).Get(ctx, ref.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("read secret %s/%s: %w", ns, ref.Name, err)
	}

	key, err := secretValue(secret, ref.Key)
	if err != nil {
		return "", fmt.Errorf("read secret key %s in %s/%s: %w", ref.Key, ns, ref.Name, err)
	}

	return key, nil
}

func secretValue(secret *corev1.Secret, key string) (string, error) {
	val, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("key not found")
	}

	trimmed := strings.TrimSpace(string(val))
	if trimmed == "" {
		return "", fmt.Errorf("key value is empty")
	}

	return trimmed, nil
}

func normalizeAuthorizationHeader(apiKey string) string {
	trimmed := strings.TrimSpace(apiKey)
	fields := strings.Fields(trimmed)
	if len(fields) >= 2 && strings.EqualFold(fields[0], "api") && strings.EqualFold(fields[1], "key") {
		token := strings.TrimSpace(strings.Join(fields[2:], " "))
		if token != "" {
			return "APIKEY " + token
		}
		return trimmed
	}
	if len(fields) >= 1 && strings.EqualFold(fields[0], "apikey") {
		token := strings.TrimSpace(strings.Join(fields[1:], " "))
		if token != "" {
			return "APIKEY " + token
		}
		return trimmed
	}
	return "APIKEY " + trimmed
}

func recordNameFromFQDN(resolvedFQDN, zone string) (string, error) {
	fqdn := strings.TrimSuffix(resolvedFQDN, ".")
	zone = strings.TrimSuffix(zone, ".")

	if fqdn == "" || zone == "" {
		return "", fmt.Errorf("invalid fqdn/zone for record name resolution")
	}

	fqdnLower := strings.ToLower(fqdn)
	zoneLower := strings.ToLower(zone)
	suffix := "." + zoneLower

	if fqdnLower == zoneLower {
		return "@", nil
	}
	if !strings.HasSuffix(fqdnLower, suffix) {
		return "", fmt.Errorf("resolvedFQDN %q does not belong to zone %q", resolvedFQDN, zone)
	}

	name := fqdn[:len(fqdn)-len(suffix)]
	if name == "" {
		return "@", nil
	}
	return name, nil
}

func isTXT(record dnsRecord) bool {
	return strings.EqualFold(record.Type, "txt")
}

func hasTXT(records []dnsRecord, name, value string) bool {
	for _, record := range records {
		if !isTXT(record) || !sameRecordName(record.Name, name) {
			continue
		}

		text, ok := txtText(record)
		if ok && text == value {
			return true
		}
	}

	return false
}

func sameRecordName(actual, expected string) bool {
	return strings.EqualFold(strings.TrimSuffix(actual, "."), strings.TrimSuffix(expected, "."))
}

func txtText(record dnsRecord) (string, bool) {
	var v txtValue
	if err := json.Unmarshal(record.Value, &v); err != nil {
		return "", false
	}

	if v.Text == "" {
		return "", false
	}

	return v.Text, true
}
