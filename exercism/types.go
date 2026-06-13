package exercism

import "strings"

// Track is the output record for a programming language track.
type Track struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	NumExercises int    `json:"num_exercises"`
	NumConcepts  int    `json:"num_concepts"`
	NumLearners  int    `json:"num_learners"`
	Tags         string `json:"tags"` // semicolon-joined
	URL          string `json:"url"`
}

// Exercise is the output record for a track exercise.
type Exercise struct {
	UUID       string `json:"uuid"`
	Slug       string `json:"slug"`
	Title      string `json:"title"`
	Difficulty string `json:"difficulty"`
	Blurb      string `json:"blurb"`
	External   bool   `json:"is_external"`
	Unlocked   bool   `json:"is_unlocked"`
	URL        string `json:"url"`
}

// Concept is the output record for a track concept.
type Concept struct {
	UUID  string `json:"uuid"`
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Blurb string `json:"blurb"`
	URL   string `json:"url"`
}

// ─── wire types ──────────────────────────────────────────────────────────────

type wireLinks struct {
	Self      string `json:"self"`
	Exercises string `json:"exercises"`
	Concepts  string `json:"concepts"`
}

type wireTrack struct {
	Slug         string    `json:"slug"`
	Title        string    `json:"title"`
	NumExercises int       `json:"num_exercises"`
	NumConcepts  int       `json:"num_concepts"`
	NumLearners  int       `json:"num_learners"`
	Tags         []string  `json:"tags"`
	Links        wireLinks `json:"links"`
}

type wireExerciseLinks struct {
	Self string `json:"self"`
}

type wireExercise struct {
	UUID       string            `json:"uuid"`
	Slug       string            `json:"slug"`
	Title      string            `json:"title"`
	Difficulty string            `json:"difficulty"`
	Blurb      string            `json:"blurb"`
	IsExternal bool              `json:"is_external"`
	IsUnlocked bool              `json:"is_unlocked"`
	Links      wireExerciseLinks `json:"links"`
}

type wireConceptLinks struct {
	Self string `json:"self"`
}

type wireConcept struct {
	UUID  string           `json:"uuid"`
	Slug  string           `json:"slug"`
	Name  string           `json:"name"`
	Blurb string           `json:"blurb"`
	Links wireConceptLinks `json:"links"`
}

// ─── converters ──────────────────────────────────────────────────────────────

func wireTrackToTrack(w wireTrack) Track {
	return Track{
		Slug:         w.Slug,
		Title:        w.Title,
		NumExercises: w.NumExercises,
		NumConcepts:  w.NumConcepts,
		NumLearners:  w.NumLearners,
		Tags:         strings.Join(w.Tags, ";"),
		URL:          w.Links.Self,
	}
}

func wireExerciseToExercise(w wireExercise) Exercise {
	return Exercise{
		UUID:       w.UUID,
		Slug:       w.Slug,
		Title:      w.Title,
		Difficulty: w.Difficulty,
		Blurb:      w.Blurb,
		External:   w.IsExternal,
		Unlocked:   w.IsUnlocked,
		URL:        w.Links.Self,
	}
}

func wireConceptToConcept(w wireConcept) Concept {
	return Concept{
		UUID:  w.UUID,
		Slug:  w.Slug,
		Name:  w.Name,
		Blurb: w.Blurb,
		URL:   w.Links.Self,
	}
}
