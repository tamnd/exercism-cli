// Package exercism is the library behind the exercism command line:
// the HTTP client, request shaping, and the typed data models for exercism.org.
//
// All endpoints are open and require no API key. The Client paces requests,
// retries 429/5xx with linear backoff, and returns typed record structs.
package exercism

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// DefaultUserAgent identifies the client to exercism.org.
const DefaultUserAgent = "exercism/dev (+https://github.com/tamnd/exercism-cli)"

// ErrNotFound is returned when the API returns a 404 for a resource.
var ErrNotFound = errors.New("not found")

// Config holds constructor parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://exercism.org",
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Retries:   3,
		Timeout:   15 * time.Second,
	}
}

// Client talks to exercism.org over HTTP.
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
	rate       time.Duration
	retries    int
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client configured from cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		userAgent:  cfg.UserAgent,
		rate:       cfg.Rate,
		retries:    cfg.Retries,
	}
}

// get fetches a URL with pacing and retries.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, false, ErrNotFound
	}
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rate <= 0 {
		return
	}
	if wait := c.rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// getJSON fetches and JSON-decodes into v.
func (c *Client) getJSON(ctx context.Context, rawURL string, v any) error {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("decode %s: %w", rawURL, err)
	}
	return nil
}

// ─── API methods ─────────────────────────────────────────────────────────────

type tracksResp struct {
	Tracks []wireTrack `json:"tracks"`
}

// Tracks returns all programming tracks.
func (c *Client) Tracks(ctx context.Context) ([]Track, error) {
	u := c.baseURL + "/api/v2/tracks"
	var resp tracksResp
	if err := c.getJSON(ctx, u, &resp); err != nil {
		return nil, err
	}
	out := make([]Track, len(resp.Tracks))
	for i, wt := range resp.Tracks {
		out[i] = wireTrackToTrack(wt)
	}
	return out, nil
}

type trackResp struct {
	Track wireTrack `json:"track"`
}

// GetTrack returns detail for a single track. Returns ErrNotFound when the
// slug does not exist.
func (c *Client) GetTrack(ctx context.Context, slug string) (Track, error) {
	u := c.baseURL + "/api/v2/tracks/" + slug
	var resp trackResp
	if err := c.getJSON(ctx, u, &resp); err != nil {
		return Track{}, err
	}
	return wireTrackToTrack(resp.Track), nil
}

type exercisesResp struct {
	Exercises []wireExercise `json:"exercises"`
}

// Exercises returns all exercises in the named track.
func (c *Client) Exercises(ctx context.Context, slug string) ([]Exercise, error) {
	u := c.baseURL + "/api/v2/tracks/" + slug + "/exercises"
	var resp exercisesResp
	if err := c.getJSON(ctx, u, &resp); err != nil {
		return nil, err
	}
	out := make([]Exercise, len(resp.Exercises))
	for i, we := range resp.Exercises {
		out[i] = wireExerciseToExercise(we)
	}
	return out, nil
}

type conceptsResp struct {
	Concepts []wireConcept `json:"concepts"`
}

// Concepts returns all concepts in the named track.
func (c *Client) Concepts(ctx context.Context, slug string) ([]Concept, error) {
	u := c.baseURL + "/api/v2/tracks/" + slug + "/concepts"
	var resp conceptsResp
	if err := c.getJSON(ctx, u, &resp); err != nil {
		return nil, err
	}
	out := make([]Concept, len(resp.Concepts))
	for i, wc := range resp.Concepts {
		out[i] = wireConceptToConcept(wc)
	}
	return out, nil
}
