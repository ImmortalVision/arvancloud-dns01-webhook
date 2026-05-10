package arvancloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type client struct {
	baseURL    string
	authHeader string
	httpClient *http.Client
}

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
		if strings.Contains(strings.ToLower(err.Error()), "404") {
			return nil
		}
		return err
	}

	return nil
}

func (c *client) do(req *http.Request, out any) error {
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Accept", "application/json")
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("arvan API %s %s failed with %d: %s", req.Method, req.URL.Path, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if out == nil || len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
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
