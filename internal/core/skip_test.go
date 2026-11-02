package core

import (
	"context"
	"testing"
	"time"
)

func TestSkipper(t *testing.T) {
	var nilSkipper *Skipper
	ctx, release := nilSkipper.Watch(context.Background(), "scan", "m", 1)
	release()
	if ctx.Err() != nil || nilSkipper.Skip("m") || nilSkipper.IsSkipped("m") || len(nilSkipper.Inflight()) != 0 || len(nilSkipper.SkipAll()) != 0 {
		t.Fatal("nil skipper must be a no-op")
	}

	s := NewSkipper()
	base := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	s.now = func() time.Time { base = base.Add(time.Second); return base }
	ctxA, relA := s.Watch(context.Background(), "scan", "a", 1)
	ctxB, relB := s.Watch(context.Background(), "scan", "b", 1)
	if got := s.Inflight(); len(got) != 2 || got[0].ID != "a" || got[1].ID != "b" || got[0].Stage != "scan" {
		t.Fatalf("inflight = %+v", got)
	}
	if !s.Skip("a") {
		t.Fatal("skip of an in-flight model must report true")
	}
	if ctxA.Err() == nil {
		t.Fatal("skipped request must be cancelled")
	}
	if ctxB.Err() != nil {
		t.Fatal("other requests must not be affected")
	}
	if !s.IsSkipped("a") || s.IsSkipped("b") {
		t.Fatal("skipped set wrong")
	}
	relA()
	if got := s.Inflight(); len(got) != 1 || got[0].ID != "b" {
		t.Fatalf("inflight after release = %+v", got)
	}
	// A later request for a skipped model starts cancelled.
	ctxA2, relA2 := s.Watch(context.Background(), "latency", "a", 0)
	if ctxA2.Err() == nil {
		t.Fatal("watch of a skipped model must return a cancelled context")
	}
	relA2()
	if s.Skip("unknown") {
		t.Fatal("skip of an unknown model reports false")
	}
	if !s.Skip("a") {
		t.Fatal("skip of an already skipped model reports true")
	}
	if ids := s.SkipAll(); len(ids) != 1 || ids[0] != "b" || ctxB.Err() == nil {
		t.Fatalf("SkipAll = %v", ids)
	}
	relB()
	if got := s.Skipped(); len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "unknown" {
		t.Fatalf("Skipped = %v", got)
	}
	if len(s.Inflight()) != 0 {
		t.Fatal("inflight must be empty after release")
	}
}
