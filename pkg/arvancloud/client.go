package arvancloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type client struct {
	baseURL    string
	authHeader string
	httpClient *http.Client
	sleep      func(time.Duration)
	realSleep  bool
}

const (
	maxRetries            = 3
	baseRetryDelay        = 200 * time.Millisecond
	maxRetryDelay         = 2 * time.Second
	retryJitterPercentage = 0.20
)

type dnsRecord struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

type listDNSRecordsResponse struct {
	Data []dnsRecord `json:"data"`
}

type createDNSRecordResponse struct {
	Data dnsRecord `json:"data"`
}

type txtValue struct {
	Text string `json:"text"`
}

type createTXTRequest struct {
	Type  string   `json:"type"`
	Name  string   `json:"name"`
	Cloud bool     `json:"cloud"`
	Value txtValue `json:"value"`
	TTL   int      `json:"ttl"`
}

func newClient(baseURL, authHeader string, httpClient *http.Client) *client {
	return &client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		authHeader: authHeader,
		httpClient: httpClient,
		sleep:      time.Sleep,
		realSleep:  true,
	}
}

func (c *client) CreateTXTRecord(ctx context.Context, domain, name, value string, ttl int) error {
	reqBody := createTXTRequest{
		Type:  "TXT",
		Name:  name,
		Cloud: false,
		Value: txtValue{Text: value},
		TTL:   ttl,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal create TXT payload: %w", err)
	}

	endpoint, err := c.endpoint("domains", domain, "dns-records")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build create TXT request: %w", err)
	}

	var resp createDNSRecordResponse
	if err := c.do(req, &resp); err != nil {
		return err
	}

	return nil
}

func (c *client) ListTXTRecords(ctx context.Context, domain, search string) ([]dnsRecord, error) {
	endpoint, err := c.endpoint("domains", domain, "dns-records")
	if err != nil {
		return nil, err
	}

	const perPage = 100
	all := make([]dnsRecord, 0, perPage)

	for page := 1; ; page++ {
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, fmt.Errorf("parse list endpoint: %w", err)
		}

		q := u.Query()
		q.Set("type", "txt")
		q.Set("page", fmt.Sprintf("%d", page))
		q.Set("per_page", fmt.Sprintf("%d", perPage))
		if search != "" {
			q.Set("search", search)
		}
		u.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("build list TXT request: %w", err)
		}

		var resp listDNSRecordsResponse
		if err := c.do(req, &resp); err != nil {
			return nil, err
		}

		all = append(all, resp.Data...)
		if len(resp.Data) < perPage {
			break
		}
	}

	return all, nil
}

func (c *client) DeleteRecord(ctx context.Context, domain, recordID string) error {
	endpoint, err := c.endpoint("domains", domain, "dns-records", recordID)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build delete DNS record request: %w", err)
	}

	if err := c.do(req, nil); err != nil {
		var apiErr *apiError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return nil
		}
		return err
	}

	return nil
}

func (c *client) do(req *http.Request, out any) error {
	for attempt := 0; ; attempt++ {
		attemptReq, err := cloneRequest(req)
		if err != nil {
			return err
		}

		body, _, err := c.doOnce(attemptReq)
		if err == nil {
			if out == nil || len(body) == 0 {
				return nil
			}
			if err := json.Unmarshal(body, out); err != nil {
				return fmt.Errorf("decode response body: %w", err)
			}
			return nil
		}

		var apiErr *apiError
		if !errors.As(err, &apiErr) || !shouldRetryStatus(apiErr.StatusCode) || attempt >= maxRetries {
			return err
		}

		delay := retryDelay(attempt, attemptReq.URL.Path)
		if err := c.waitRetry(req.Context(), delay); err != nil {
			return err
		}
	}
}

type apiError struct {
	Method     string
	Path       string
	StatusCode int
	Body       string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("arvan API %s %s failed with %d: %s", e.Method, e.Path, e.StatusCode, e.Body)
}

func (c *client) doOnce(req *http.Request) ([]byte, int, error) {
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Accept", "application/json")
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, &apiError{
			Method:     req.Method,
			Path:       req.URL.Path,
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(string(body)),
		}
	}

	return body, resp.StatusCode, nil
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	cloned := req.Clone(req.Context())
	if req.Body == nil {
		return cloned, nil
	}
	if req.GetBody == nil {
		return nil, fmt.Errorf("request body cannot be retried for %s %s", req.Method, req.URL.Path)
	}
	body, err := req.GetBody()
	if err != nil {
		return nil, fmt.Errorf("reset request body for %s %s: %w", req.Method, req.URL.Path, err)
	}
	cloned.Body = body
	return cloned, nil
}

func shouldRetryStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= http.StatusInternalServerError
}

func retryDelay(attempt int, path string) time.Duration {
	// Exponential backoff with deterministic jitter (+/-20%).
	delay := float64(baseRetryDelay) * math.Pow(2, float64(attempt))
	if delay > float64(maxRetryDelay) {
		delay = float64(maxRetryDelay)
	}

	jitterSign := 1.0
	if (attempt+len(path))%2 == 0 {
		jitterSign = -1.0
	}
	jitter := delay * retryJitterPercentage * jitterSign
	return time.Duration(delay + jitter)
}

func (c *client) waitRetry(ctx context.Context, d time.Duration) error {
	if d <= 0 || c.sleep == nil {
		return nil
	}

	// Tests can replace sleep with a no-op collector to avoid wall-clock waits.
	if !c.realSleep {
		c.sleep(d)
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return fmt.Errorf("request canceled: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}

func (c *client) endpoint(parts ...string) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base URL: %w", err)
	}

	allParts := append([]string{u.Path}, parts...)
	u.Path = path.Join(allParts...)

	return u.String(), nil
}
