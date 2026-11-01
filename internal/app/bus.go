package app

import (
	"sync"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
)

// Bus fans run events out to subscribers (SSE clients, CLI progress) with
// bounded buffers. Progress events are coalesced for slow consumers;
// milestone events (round, done, error) are retained in a short history so
// that reconnecting clients can catch up by sequence number.
type Bus struct {
	mu      sync.Mutex
	subs    map[string]map[*subscriber]struct{}
	seq     map[string]int64
	latest  map[string]core.Event
	history map[string][]core.Event
	maxHist int
}

type subscriber struct {
	ch      chan core.Event
	dropped int
}

// NewBus creates a bus.
func NewBus() *Bus {
	return &Bus{subs: map[string]map[*subscriber]struct{}{}, seq: map[string]int64{}, latest: map[string]core.Event{}, history: map[string][]core.Event{}, maxHist: 500}
}

// Publish delivers an event to the subscribers of a run.
func (b *Bus) Publish(runID string, ev core.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.seq[runID]++
	ev.Seq = b.seq[runID]
	ev.RunID = runID
	if ev.At.IsZero() {
		ev.At = time.Now()
	}
	if ev.Kind == core.EventProgress || ev.Kind == core.EventAttempt || ev.Kind == core.EventSample {
		b.latest[runID] = ev
	}
	switch ev.Kind {
	case core.EventRound, core.EventLevel, core.EventDone, core.EventError, core.EventLog, core.EventCooldown:
		h := b.history[runID]
		h = append(h, ev)
		if len(h) > b.maxHist {
			h = h[len(h)-b.maxHist:]
		}
		b.history[runID] = h
	}
	for s := range b.subs[runID] {
		select {
		case s.ch <- ev:
		default:
			// Full: drop the oldest buffered event to make room for the new one.
			select {
			case <-s.ch:
				s.dropped++
			default:
			}
			select {
			case s.ch <- ev:
			default:
				s.dropped++
			}
		}
	}
}

// Subscribe returns a channel of events for a run and an unsubscribe func.
func (b *Bus) Subscribe(runID string, buffer int) (<-chan core.Event, func()) {
	if buffer <= 0 {
		buffer = 256
	}
	s := &subscriber{ch: make(chan core.Event, buffer)}
	b.mu.Lock()
	if b.subs[runID] == nil {
		b.subs[runID] = map[*subscriber]struct{}{}
	}
	b.subs[runID][s] = struct{}{}
	b.mu.Unlock()
	var once sync.Once
	return s.ch, func() {
		once.Do(func() {
			b.mu.Lock()
			delete(b.subs[runID], s)
			if len(b.subs[runID]) == 0 {
				delete(b.subs, runID)
			}
			b.mu.Unlock()
		})
	}
}

// Latest returns the most recent progress-type event of a run.
func (b *Bus) Latest(runID string) (core.Event, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ev, ok := b.latest[runID]
	return ev, ok
}

// History returns retained milestone events with Seq > after.
func (b *Bus) History(runID string, after int64) []core.Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []core.Event
	for _, ev := range b.history[runID] {
		if ev.Seq > after {
			out = append(out, ev)
		}
	}
	return out
}

// Forget drops retained state for a finished run (subscribers are kept).
func (b *Bus) Forget(runID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.latest, runID)
	delete(b.history, runID)
	delete(b.seq, runID)
}
