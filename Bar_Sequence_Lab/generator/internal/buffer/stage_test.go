package buffer

import (
	"context"
	"testing"
	"time"

	"bar_sequence_lab/generator/internal/types"
)

func TestPushPopPreservesOrder(t *testing.T) {
	s := NewStage("A1", 8)
	ctx := context.Background()
	for i := 1; i <= 3; i++ {
		obs := types.Observation{Symbol: "AAPL", GeneratorSequenceNo: uint64(i), ReceivedTime: time.Now().UTC()}
		if err := s.Push(ctx, obs); err != nil {
			t.Fatal(err)
		}
	}
	for i := 1; i <= 3; i++ {
		obs, ok, err := s.Pop(ctx)
		if err != nil || !ok || obs.GeneratorSequenceNo != uint64(i) {
			t.Fatalf("got %+v ok=%v err=%v", obs, ok, err)
		}
	}
}

func TestFullIsCriticalNotSilentDrop(t *testing.T) {
	s := NewStage("A1", 1)
	ctx := context.Background()
	if err := s.Push(ctx, types.Observation{Symbol: "AAPL", ReceivedTime: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	err := s.Push(ctx, types.Observation{Symbol: "AAPL", ReceivedTime: time.Now().UTC()})
	if err == nil {
		t.Fatal("expected backpressure error")
	}
	snap := s.Snapshot()
	if snap.Depth != 1 {
		t.Fatalf("depth=%d, observation was dropped", snap.Depth)
	}
}
