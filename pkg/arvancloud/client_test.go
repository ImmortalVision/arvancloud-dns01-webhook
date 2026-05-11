package arvancloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCreateTXTRecord(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/cdn/4.0/domains/example.com/dns-records" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "APIKEY token" {
			t.Fatalf("authorization = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %q", got)
		}

		var req createTXTRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Name != "_acme-challenge" || req.Value.Text != "challenge-token" || req.TTL != 120 {
			t.Fatalf("unexpected request payload: %+v", req)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"id":"1","name":"_acme-challenge","type":"TXT","value":{"text":"challenge-token"}}}`))
	}))
	defer srv.Close()

	c := newClient(srv.URL+"/cdn/4.0", "APIKEY token", srv.Client())
	c.realSleep = false
	c.sleep = func(time.Duration) {}

	if err := c.CreateTXTRecord(context.Background(), "example.com", "_acme-challenge", "challenge-token", 120); err != nil {
		t.Fatalf("CreateTXTRecord() error = %v", err)
	}
}

func TestListTXTRecordsPaginationAndSearch(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}

		q := r.URL.Query()
		if q.Get("type") != "txt" {
			t.Fatalf("type query = %q", q.Get("type"))
		}
		if q.Get("per_page") != "100" {
			t.Fatalf("per_page = %q", q.Get("per_page"))
		}
		if q.Get("search") != "_acme-challenge" {
			t.Fatalf("search query = %q", q.Get("search"))
		}

		page, _ := strconv.Atoi(q.Get("page"))
		if page == 1 {
			data := make([]dnsRecord, 100)
			for i := range data {
				data[i] = dnsRecord{ID: strconv.Itoa(i), Name: "_acme-challenge", Type: "TXT", Value: json.RawMessage(`{"text":"token"}`)}
			}
			_ = json.NewEncoder(w).Encode(listDNSRecordsResponse{Data: data})
			return
		}

		_ = json.NewEncoder(w).Encode(listDNSRecordsResponse{
			Data: []dnsRecord{
				{ID: "100", Name: "_acme-challenge", Type: "TXT", Value: json.RawMessage(`{"text":"token-2"}`)},
			},
		})
	}))
	defer srv.Close()

	c := newClient(srv.URL, "APIKEY token", srv.Client())
	c.realSleep = false
	c.sleep = func(time.Duration) {}

	got, err := c.ListTXTRecords(context.Background(), "example.com", "_acme-challenge")
	if err != nil {
		t.Fatalf("ListTXTRecords() error = %v", err)
	}
	if len(got) != 101 {
		t.Fatalf("ListTXTRecords() len = %d, want 101", len(got))
	}
}

func TestDeleteRecordIgnoresNotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer srv.Close()

	c := newClient(srv.URL, "APIKEY token", srv.Client())
	c.realSleep = false
	c.sleep = func(time.Duration) {}

	if err := c.DeleteRecord(context.Background(), "example.com", "missing-id"); err != nil {
		t.Fatalf("DeleteRecord() error = %v", err)
	}
}

func TestDoRetriesOn429AndSucceeds(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		call := atomic.AddInt32(&calls, 1)
		if call < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"rate limit"}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	var slept []time.Duration
	c := newClient(srv.URL, "APIKEY token", srv.Client())
	c.realSleep = false
	c.sleep = func(d time.Duration) { slept = append(slept, d) }

	_, err := c.ListTXTRecords(context.Background(), "example.com", "")
	if err != nil {
		t.Fatalf("ListTXTRecords() error = %v", err)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
	if len(slept) != 2 {
		t.Fatalf("sleep count = %d, want 2", len(slept))
	}
}

func TestDoStopsAfterMaxRetries(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`boom`))
	}))
	defer srv.Close()

	c := newClient(srv.URL, "APIKEY token", srv.Client())
	c.realSleep = false
	c.sleep = func(time.Duration) {}

	_, err := c.ListTXTRecords(context.Background(), "example.com", "")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed with 500") {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&calls) != maxRetries+1 {
		t.Fatalf("calls = %d, want %d", calls, maxRetries+1)
	}
}

func TestDoContextCancellationDuringRetry(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`retry later`))
	}))
	defer srv.Close()

	c := newClient(srv.URL, "APIKEY token", srv.Client())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := c.ListTXTRecords(ctx, "example.com", "")
	if err == nil {
		t.Fatalf("expected cancellation error")
	}
	if !strings.Contains(err.Error(), "request canceled") {
		t.Fatalf("unexpected error: %v", err)
	}
}
