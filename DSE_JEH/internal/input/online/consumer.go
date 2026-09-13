package online

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	finfeedsatv1 "fin_feedsat_1/gen/fin_feedsat/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/analytical"
	"tramuthus/dse-jeh-transsat-1/internal/evidence"
)

type Mapper struct {
	mu        sync.Mutex
	sequences map[string]uint64
}

func NewMapper() *Mapper {
	return &Mapper{sequences: make(map[string]uint64)}
}

func (mapper *Mapper) Map(bar *finfeedsatv1.Bar) *dsejehv1.BarEvent {
	mapper.mu.Lock()
	symbol := strings.ToUpper(strings.TrimSpace(bar.GetSymbol()))
	mapper.sequences[symbol]++
	sequence := mapper.sequences[symbol]
	mapper.mu.Unlock()
	return MapBar(bar, sequence)
}

func MapBar(bar *finfeedsatv1.Bar, sequence uint64) *dsejehv1.BarEvent {
	symbol := strings.ToUpper(strings.TrimSpace(bar.GetSymbol()))
	sourceObservationID := bar.GetMarketSnapshotId()
	if sourceObservationID == "" {
		sourceObservationID = evidence.ID("online-observation", bar.GetSource(), symbol, strconv.FormatInt(bar.GetIntervalStartUnixMs(), 10))
	}
	volume := bar.GetVolume()
	eventCount := bar.GetEventCount()
	return &dsejehv1.BarEvent{
		EventId:             evidence.ID("bar-event", sourceObservationID),
		EntityId:            symbol,
		Symbol:              symbol,
		InstrumentId:        bar.GetInstrumentId(),
		Interval:            bar.GetInterval(),
		IntervalStartUnixMs: bar.GetIntervalStartUnixMs(),
		IntervalEndUnixMs:   bar.GetIntervalEndUnixMs(),
		Open:                bar.GetOpen(),
		High:                bar.GetHigh(),
		Low:                 bar.GetLow(),
		Close:               bar.GetClose(),
		Volume:              &volume,
		EventCount:          &eventCount,
		Provenance: &dsejehv1.SourceProvenance{
			SourceMode:          dsejehv1.RuntimeMode_RUNTIME_MODE_ONLINE,
			SourceId:            bar.GetSource(),
			SourceObservationId: sourceObservationID,
			EntitySequence:      sequence,
			EntitySequenceScope: analytical.Scope("online:"+bar.GetSource(), symbol),
			SourceTimestamp:     bar.GetSourceTimestamp(),
			SourceEventUnixMs:   bar.GetIntervalStartUnixMs(),
			ReceivedUnixMs:      bar.GetReceiptUnixMs(),
			IsFinal:             bar.GetIsFinal(),
			IsBackfilled:        bar.GetIsBackfilled(),
		},
	}
}

type Consumer struct {
	address       string
	symbols       []string
	maxBars       uint32
	finalizedOnly bool
	mapper        *Mapper
}

func New(address string, symbols []string, maxBars uint32, finalizedOnly bool) *Consumer {
	return &Consumer{address: address, symbols: symbols, maxBars: maxBars, finalizedOnly: finalizedOnly, mapper: NewMapper()}
}

func (consumer *Consumer) Stream(ctx context.Context, emit func(*dsejehv1.BarEvent) error) error {
	connection, err := grpc.NewClient(consumer.address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create Fin_FeedSat_1 client: %w", err)
	}
	defer connection.Close()
	client := finfeedsatv1.NewIngestionServiceClient(connection)
	backoff := 200 * time.Millisecond
	for attempt := 0; attempt < 5; attempt++ {
		stream, streamErr := client.StreamBars(ctx, &finfeedsatv1.StreamBarsRequest{
			Symbols:       consumer.symbols,
			MaxBars:       consumer.maxBars,
			FinalizedOnly: consumer.finalizedOnly,
		})
		if streamErr == nil {
			for {
				bar, receiveErr := stream.Recv()
				if receiveErr == nil {
					if err := emit(consumer.mapper.Map(bar)); err != nil {
						return err
					}
					continue
				}
				if receiveErr == io.EOF {
					return nil
				}
				streamErr = receiveErr
				break
			}
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if attempt == 4 {
			return fmt.Errorf("Fin_FeedSat_1 stream failed after 5 attempts: %w", streamErr)
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		backoff = min(backoff*2, 3*time.Second)
	}
	return nil
}
