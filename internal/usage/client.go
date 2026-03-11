package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	defaultBaseURL = "https://api.anthropic.com"
	usagePath      = "/api/oauth/usage"
	defaultTTL     = 60 * time.Second
	requestTimeout = 5 * time.Second
	betaHeader     = "oauth-2025-04-20"
)

// Window holds a single usage window's utilization percentage and reset time.
type Window struct {
	Utilization float64    // 0-100 percentage
	ResetsAt    *time.Time // nil if not parseable
}

// Data holds the three usage windows returned by the Anthropic usage API.
type Data struct {
	FiveHour     *Window
	SevenDay     *Window
	SevenDayOpus *Window // nil for non-Opus tiers
}

// Client fetches usage data from the Anthropic API with in-memory caching.
type Client struct {
	token   string
	baseURL string
	ttl     time.Duration

	mu       sync.Mutex
	cached   *Data
	cachedAt time.Time
	lastGood *Data
}

// NewClient creates a new Client. If baseURL is empty, the production
// Anthropic API base URL is used. Tests may pass a local httptest URL.
func NewClient(token, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{token: token, baseURL: baseURL, ttl: defaultTTL}
}

// SetTTL overrides the cache TTL. Primarily used in tests to bypass caching.
func (c *Client) SetTTL(d time.Duration) {
	c.mu.Lock()
	c.ttl = d
	c.mu.Unlock()
}

// Fetch returns the current usage data, whether the result is stale, and any
// error. Results are cached for the configured TTL (default 60s). On HTTP or
// parse errors, the last successful response is returned with stale=true; if
// there is no previous good response, the error is returned instead.
func (c *Client) Fetch(ctx context.Context) (*Data, bool, error) {
	c.mu.Lock()
	if c.cached != nil && time.Since(c.cachedAt) < c.ttl {
		cached := c.cached
		c.mu.Unlock()
		return cached, false, nil
	}
	c.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+usagePath, nil)
	if err != nil {
		return c.stale(fmt.Errorf("building request: %w", err))
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("anthropic-beta", betaHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return c.stale(fmt.Errorf("fetching usage: %w", err))
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return c.stale(fmt.Errorf("usage API returned HTTP %d", resp.StatusCode))
	}

	var raw struct {
		FiveHour     *rawWindow `json:"five_hour"`
		SevenDay     *rawWindow `json:"seven_day"`
		SevenDayOpus *rawWindow `json:"seven_day_opus"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return c.stale(fmt.Errorf("decoding response: %w", err))
	}

	data := &Data{
		FiveHour:     parseWindow(raw.FiveHour),
		SevenDay:     parseWindow(raw.SevenDay),
		SevenDayOpus: parseWindow(raw.SevenDayOpus),
	}

	c.mu.Lock()
	c.cached = data
	c.cachedAt = time.Now()
	c.lastGood = data
	c.mu.Unlock()

	return data, false, nil
}

type rawWindow struct {
	Utilization float64 `json:"utilization"`
	ResetsAtRaw *string `json:"resets_at"`
}

func parseWindow(r *rawWindow) *Window {
	if r == nil {
		return nil
	}
	w := &Window{Utilization: r.Utilization}
	if r.ResetsAtRaw != nil {
		if t, err := time.Parse(time.RFC3339Nano, *r.ResetsAtRaw); err == nil {
			w.ResetsAt = &t
		}
	}
	return w
}

func (c *Client) stale(err error) (*Data, bool, error) {
	c.mu.Lock()
	last := c.lastGood
	c.mu.Unlock()
	if last != nil {
		return last, true, nil
	}
	return nil, false, err
}
