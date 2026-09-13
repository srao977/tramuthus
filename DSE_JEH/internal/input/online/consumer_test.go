package online

import (
	"context"
	"net"
	"testing"

	finfeedsatv1 "fin_feedsat_1/gen/fin_feedsat/v1"

	"google.golang.org/grpc"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

type ingestionServer struct {
	finfeedsatv1.UnimplementedIngestionServiceServer
}

func (ingestionServer) StreamBars(request *finfeedsatv1.StreamBarsRequest, stream grpc.ServerStreamingServer[finfeedsatv1.Bar]) error {
	for index := uint32(0); index < request.GetMaxBars(); index++ {
		if err := stream.Send(&finfeedsatv1.Bar{
			Symbol: "AAPL", High: 101 + float64(index), Low: 99 + float64(index),
			Open: 100 + float64(index), Close: 100.5 + float64(index),
			Interval: "1Min", IntervalStartUnixMs: int64(index+1) * 60_000,
			IntervalEndUnixMs: int64(index+2) * 60_000, Source: "test",
			MarketSnapshotId: string(rune('a' + index)), IsFinal: true,
		}); err != nil {
			return err
		}
	}
	return nil
}

func TestConsumerUsesGeneratedStreamBarsContract(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	finfeedsatv1.RegisterIngestionServiceServer(server, ingestionServer{})
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})

	var sequences []uint64
	err = New(listener.Addr().String(), []string{"AAPL"}, 3, true).Stream(context.Background(), func(event *dsejehv1.BarEvent) error {
		sequences = append(sequences, event.GetProvenance().GetEntitySequence())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sequences) != 3 || sequences[0] != 1 || sequences[2] != 3 {
		t.Fatalf("sequences = %v; want [1 2 3]", sequences)
	}
}
