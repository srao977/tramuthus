package buffer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"bar_sequence_lab/generator/internal/types"
)

type Snapshot struct {
	Name      string
	Depth     int
	Capacity  int
	HighWater int
	OldestAge time.Duration
}

type Stage struct {
	name     string
	capacity int
	ch       chan types.Observation

	mu        sync.Mutex
	closed    bool
	highWater int
	enqueued  uint64
	dequeued  uint64
	oldest    time.Time
}

func NewStage(name string, capacity int) *Stage {
	return &Stage{
		name:     name,
		capacity: capacity,
		ch:       make(chan types.Observation, capacity),
	}
}

func (s *Stage) Push(ctx context.Context, obs types.Observation) error {
	select {
	case s.ch <- obs:
		s.mu.Lock()
		s.enqueued++
		if len(s.ch) > s.highWater {
			s.highWater = len(s.ch)
		}
		if s.oldest.IsZero() {
			s.oldest = obs.ReceivedTime
		}
		s.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("critical backpressure: buffer %s full capacity=%d; refused valid observation symbol=%s (no silent drop)", s.name, s.capacity, obs.Symbol)
	}
}

func (s *Stage) Pop(ctx context.Context) (types.Observation, bool, error) {
	select {
	case obs, ok := <-s.ch:
		if !ok {
			return types.Observation{}, false, nil
		}
		s.mu.Lock()
		s.dequeued++
		if len(s.ch) == 0 {
			s.oldest = time.Time{}
		}
		s.mu.Unlock()
		return obs, true, nil
	case <-ctx.Done():
		return types.Observation{}, false, ctx.Err()
	}
}

func (s *Stage) TryPop() (types.Observation, bool) {
	select {
	case obs, ok := <-s.ch:
		if !ok {
			return types.Observation{}, false
		}
		s.noteDequeued()
		return obs, true
	default:
		return types.Observation{}, false
	}
}

func (s *Stage) noteDequeued() {
	s.mu.Lock()
	s.dequeued++
	if len(s.ch) == 0 {
		s.oldest = time.Time{}
	}
	s.mu.Unlock()
}

func (s *Stage) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.ch)
}

func (s *Stage) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	depth := len(s.ch)
	age := time.Duration(0)
	if !s.oldest.IsZero() {
		age = time.Since(s.oldest)
	}
	return Snapshot{
		Name:      s.name,
		Depth:     depth,
		Capacity:  s.capacity,
		HighWater: s.highWater,
		OldestAge: age,
	}
}

func (s *Stage) C() <-chan types.Observation {
	return s.ch
}
