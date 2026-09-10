package sequence

import (
	"sync"
	"time"

	"bar_sequence_lab/generator/internal/types"
)

type Authority struct {
	runID string

	mu          sync.Mutex
	next        map[string]uint64
	lastTime    map[string]time.Time
	seen        map[string]struct{}
	regressions uint64
	duplicates  uint64
}

func New(runID string) *Authority {
	return &Authority{
		runID:    runID,
		next:     map[string]uint64{},
		lastTime: map[string]time.Time{},
		seen:     map[string]struct{}{},
	}
}

func (a *Authority) Accept(obs types.Observation) types.Observation {
	a.mu.Lock()
	defer a.mu.Unlock()
	obs.CollectionRunID = a.runID
	seq := a.next[obs.Symbol] + 1
	a.next[obs.Symbol] = seq
	obs.GeneratorSequenceNo = seq

	if last, ok := a.lastTime[obs.Symbol]; ok && !obs.SourceEventTime.IsZero() && obs.SourceEventTime.Before(last) {
		obs.SourceTimeRegression = true
		a.regressions++
	}
	if !obs.SourceEventTime.IsZero() {
		a.lastTime[obs.Symbol] = obs.SourceEventTime
	}

	key := obs.DuplicateKey()
	if _, ok := a.seen[key]; ok {
		obs.DuplicateArrival = true
		a.duplicates++
	} else {
		a.seen[key] = struct{}{}
	}
	return obs
}

func (a *Authority) Stats() (regressions, duplicates uint64, latest map[string]uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	latest = make(map[string]uint64, len(a.next))
	for k, v := range a.next {
		latest[k] = v
	}
	return a.regressions, a.duplicates, latest
}
