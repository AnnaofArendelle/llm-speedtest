package core

import (
	"context"
	"sort"
	"sync"
	"time"
)

// InflightItem describes a request that is currently in progress.
type InflightItem struct {
	ID    string    `json:"id"`
	Stage string    `json:"stage"`
	Round int       `json:"round,omitempty"`
	Since time.Time `json:"since"`
}

type inflightEntry struct {
	cancel context.CancelFunc
	item   InflightItem
}

// Skipper lets a user skip individual models while a run is in progress.
// Engines wrap every request context with Watch; Skip cancels the in-flight
// request of that model (and only that one) and remembers the model so that
// later rounds, samples and stages leave it out. A nil *Skipper is a valid
// no-op, so engines can always call its methods.
type Skipper struct {
	mu       sync.Mutex
	inflight map[string]*inflightEntry
	skipped  map[string]bool
	now      func() time.Time
}

// NewSkipper creates an empty skipper.
func NewSkipper() *Skipper {
	return &Skipper{inflight: map[string]*inflightEntry{}, skipped: map[string]bool{}, now: time.Now}
}

// Watch derives a context for one request of model id. The returned release
// function must be called when the request has finished. If the model was
// already skipped the returned context is cancelled immediately.
func (s *Skipper) Watch(ctx context.Context, stage, id string, round int) (context.Context, func()) {
	if s == nil {
		return ctx, func() {}
	}
	cctx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	if s.skipped[id] {
		cancel()
	}
	entry := &inflightEntry{cancel: cancel, item: InflightItem{ID: id, Stage: stage, Round: round, Since: s.now()}}
	s.inflight[id] = entry
	s.mu.Unlock()
	return cctx, func() {
		s.mu.Lock()
		if cur := s.inflight[id]; cur == entry {
			delete(s.inflight, id)
		}
		s.mu.Unlock()
		cancel()
	}
}

// Skip marks the model as skipped and cancels its in-flight request, if any.
// It reports whether the model was known (in flight or already skipped).
func (s *Skipper) Skip(id string) bool {
	if s == nil || id == "" {
		return false
	}
	s.mu.Lock()
	known := s.skipped[id]
	s.skipped[id] = true
	entry := s.inflight[id]
	s.mu.Unlock()
	if entry != nil {
		entry.cancel()
		return true
	}
	return known
}

// SkipAll skips every model that currently has a request in flight and
// returns their ids.
func (s *Skipper) SkipAll() []string {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	ids := make([]string, 0, len(s.inflight))
	for id := range s.inflight {
		ids = append(ids, id)
	}
	s.mu.Unlock()
	sort.Strings(ids)
	for _, id := range ids {
		s.Skip(id)
	}
	return ids
}

// IsSkipped reports whether the model has been skipped.
func (s *Skipper) IsSkipped(id string) bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.skipped[id]
}

// Skipped returns the ids of all skipped models, sorted.
func (s *Skipper) Skipped() []string {
	if s == nil {
		return []string{}
	}
	s.mu.Lock()
	out := make([]string, 0, len(s.skipped))
	for id := range s.skipped {
		out = append(out, id)
	}
	s.mu.Unlock()
	sort.Strings(out)
	return out
}

// Inflight returns the requests currently in progress, oldest first.
func (s *Skipper) Inflight() []InflightItem {
	if s == nil {
		return []InflightItem{}
	}
	s.mu.Lock()
	out := make([]InflightItem, 0, len(s.inflight))
	for _, e := range s.inflight {
		out = append(out, e.item)
	}
	s.mu.Unlock()
	sort.Slice(out, func(i, j int) bool {
		if !out[i].Since.Equal(out[j].Since) {
			return out[i].Since.Before(out[j].Since)
		}
		return out[i].ID < out[j].ID
	})
	return out
}
