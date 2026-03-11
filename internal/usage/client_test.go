package usage_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Curt-Park/claudeview/internal/usage"
)

func makeServer(t *testing.T, fiveHourPct, sevenDayPct float64) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]any{
			"five_hour": map[string]any{
				"utilization": fiveHourPct,
				"resets_at":   time.Now().Add(2 * time.Hour).Format(time.RFC3339Nano),
			},
			"seven_day": map[string]any{
				"utilization": sevenDayPct,
				"resets_at":   time.Now().Add(5 * 24 * time.Hour).Format(time.RFC3339Nano),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
}

func TestClientFetch(t *testing.T) {
	srv := makeServer(t, 42.5, 18.0)
	defer srv.Close()

	c := usage.NewClient("test-token", srv.URL)
	data, stale, err := c.Fetch(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stale {
		t.Error("expected fresh result on first fetch")
	}
	if data.FiveHour == nil || data.FiveHour.Utilization != 42.5 {
		t.Errorf("unexpected FiveHour: %v", data.FiveHour)
	}
	if data.SevenDay == nil || data.SevenDay.Utilization != 18.0 {
		t.Errorf("unexpected SevenDay: %v", data.SevenDay)
	}
}

func TestClientFetchCaches(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		payload := map[string]any{
			"five_hour": map[string]any{"utilization": 10.0, "resets_at": time.Now().Add(time.Hour).Format(time.RFC3339Nano)},
			"seven_day": map[string]any{"utilization": 5.0, "resets_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339Nano)},
		}
		json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	c := usage.NewClient("tok", srv.URL)
	c.Fetch(t.Context())
	c.Fetch(t.Context()) // should hit cache, not server

	if calls != 1 {
		t.Errorf("expected 1 HTTP call, got %d (caching not working)", calls)
	}
}

func TestClientFetchHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := usage.NewClient("tok", srv.URL)
	_, _, err := c.Fetch(t.Context())
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestClientFetchStaleOnError(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			payload := map[string]any{
				"five_hour": map[string]any{"utilization": 50.0, "resets_at": time.Now().Add(time.Hour).Format(time.RFC3339Nano)},
				"seven_day": map[string]any{"utilization": 25.0, "resets_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339Nano)},
			}
			json.NewEncoder(w).Encode(payload)
		} else {
			http.Error(w, "error", http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	c := usage.NewClient("tok", srv.URL)
	c.SetTTL(0) // bypass cache so second call hits server
	c.Fetch(t.Context())
	data, stale, err := c.Fetch(t.Context())
	if err != nil {
		t.Fatalf("expected stale fallback, got error: %v", err)
	}
	if !stale {
		t.Error("expected stale=true on second failing fetch")
	}
	if data == nil || data.FiveHour == nil {
		t.Error("expected stale data to be non-nil")
	}
}
