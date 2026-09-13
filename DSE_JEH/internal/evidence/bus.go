package evidence

import (
	"sync"
	"sync/atomic"
	"time"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

const SubscriberBuffer = 64

type Bus struct {
	mu          sync.RWMutex
	subscribers map[uint64]chan *dsejehv1.RuntimeEvidenceEnvelope
	nextID      uint64
	sequence    atomic.Uint64
}

func NewBus() *Bus {
	return &Bus{subscribers: make(map[uint64]chan *dsejehv1.RuntimeEvidenceEnvelope)}
}

func (bus *Bus) Publish(runtimeID string, envelope *dsejehv1.RuntimeEvidenceEnvelope) {
	if envelope == nil {
		return
	}
	envelope.PublicationId = ID("publication", runtimeID, time.Now().UTC().Format(time.RFC3339Nano))
	envelope.RuntimeId = runtimeID
	envelope.PublicationSequence = bus.sequence.Add(1)
	envelope.ProducedUnixMs = time.Now().UnixMilli()
	bus.mu.RLock()
	defer bus.mu.RUnlock()
	for _, subscriber := range bus.subscribers {
		select {
		case subscriber <- envelope:
		default:
		}
	}
}

func (bus *Bus) Subscribe() (<-chan *dsejehv1.RuntimeEvidenceEnvelope, func()) {
	bus.mu.Lock()
	bus.nextID++
	id := bus.nextID
	channel := make(chan *dsejehv1.RuntimeEvidenceEnvelope, SubscriberBuffer)
	bus.subscribers[id] = channel
	bus.mu.Unlock()
	return channel, func() {
		bus.mu.Lock()
		delete(bus.subscribers, id)
		bus.mu.Unlock()
	}
}
