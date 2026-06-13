package exercism_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/exercism-cli/exercism"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*exercism.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	cfg := exercism.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	return exercism.NewClient(cfg), srv
}

func TestGetSendsUserAgent(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(`{"tracks":[]}`))
	})
	defer srv.Close()

	_, err := c.Tracks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"tracks":[]}`))
	})
	defer srv.Close()

	start := time.Now()
	_, err := c.Tracks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestGetNotFound(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	_, err := c.GetTrack(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTracks(t *testing.T) {
	const fixture = `{
		"tracks": [
			{
				"slug": "go",
				"title": "Go",
				"num_exercises": 140,
				"num_concepts": 14,
				"num_learners": 45231,
				"tags": ["compiled", "static"],
				"links": {
					"self": "https://exercism.org/tracks/go",
					"exercises": "https://exercism.org/tracks/go/exercises",
					"concepts": "https://exercism.org/tracks/go/concepts"
				}
			}
		]
	}`

	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/tracks" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(fixture))
	})
	defer srv.Close()

	tracks, err := c.Tracks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 1 {
		t.Fatalf("got %d tracks, want 1", len(tracks))
	}
	tr := tracks[0]
	if tr.Slug != "go" {
		t.Errorf("slug = %q, want go", tr.Slug)
	}
	if tr.Title != "Go" {
		t.Errorf("title = %q, want Go", tr.Title)
	}
	if tr.NumExercises != 140 {
		t.Errorf("num_exercises = %d, want 140", tr.NumExercises)
	}
	if tr.Tags != "compiled;static" {
		t.Errorf("tags = %q, want compiled;static", tr.Tags)
	}
	if tr.URL != "https://exercism.org/tracks/go" {
		t.Errorf("url = %q", tr.URL)
	}
}

func TestGetTrack(t *testing.T) {
	const fixture = `{
		"track": {
			"slug": "go",
			"title": "Go",
			"num_exercises": 140,
			"num_concepts": 14,
			"links": {
				"self": "https://exercism.org/tracks/go",
				"exercises": "https://exercism.org/tracks/go/exercises",
				"concepts": "https://exercism.org/tracks/go/concepts"
			}
		}
	}`

	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/tracks/go" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(fixture))
	})
	defer srv.Close()

	tr, err := c.GetTrack(context.Background(), "go")
	if err != nil {
		t.Fatal(err)
	}
	if tr.Slug != "go" {
		t.Errorf("slug = %q, want go", tr.Slug)
	}
	if tr.NumConcepts != 14 {
		t.Errorf("num_concepts = %d, want 14", tr.NumConcepts)
	}
}

func TestExercises(t *testing.T) {
	const fixture = `{
		"exercises": [
			{
				"uuid": "6c88f46b-1234-5678-9012-abcdef012345",
				"slug": "hello-world",
				"title": "Hello, World!",
				"difficulty": "easy",
				"blurb": "The classical introductory exercise.",
				"is_external": false,
				"is_unlocked": true,
				"links": {
					"self": "https://exercism.org/tracks/go/exercises/hello-world"
				}
			}
		]
	}`

	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/tracks/go/exercises" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(fixture))
	})
	defer srv.Close()

	exercises, err := c.Exercises(context.Background(), "go")
	if err != nil {
		t.Fatal(err)
	}
	if len(exercises) != 1 {
		t.Fatalf("got %d exercises, want 1", len(exercises))
	}
	ex := exercises[0]
	if ex.Slug != "hello-world" {
		t.Errorf("slug = %q, want hello-world", ex.Slug)
	}
	if ex.Difficulty != "easy" {
		t.Errorf("difficulty = %q, want easy", ex.Difficulty)
	}
	if ex.URL != "https://exercism.org/tracks/go/exercises/hello-world" {
		t.Errorf("url = %q", ex.URL)
	}
}

func TestConcepts(t *testing.T) {
	const fixture = `{
		"concepts": [
			{
				"uuid": "a2b3c4d5-1234-5678-9012-abcdef012345",
				"slug": "basics",
				"name": "Basics",
				"blurb": "Go is a statically-typed compiled language.",
				"links": {
					"self": "https://exercism.org/tracks/go/concepts/basics"
				}
			}
		]
	}`

	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/tracks/go/concepts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(fixture))
	})
	defer srv.Close()

	concepts, err := c.Concepts(context.Background(), "go")
	if err != nil {
		t.Fatal(err)
	}
	if len(concepts) != 1 {
		t.Fatalf("got %d concepts, want 1", len(concepts))
	}
	cn := concepts[0]
	if cn.Slug != "basics" {
		t.Errorf("slug = %q, want basics", cn.Slug)
	}
	if cn.Name != "Basics" {
		t.Errorf("name = %q, want Basics", cn.Name)
	}
	if cn.URL != "https://exercism.org/tracks/go/concepts/basics" {
		t.Errorf("url = %q", cn.URL)
	}
}
