package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

const Interval1Min = "1Min"

type Observation struct {
	CollectionRunID      string    `json:"collection_run_id" bson:"collection_run_id"`
	PartitionID          string    `json:"partition_id" bson:"partition_id"`
	Symbol               string    `json:"symbol" bson:"symbol"`
	GeneratorSequenceNo  uint64    `json:"generator_sequence_no" bson:"generator_sequence_no"`
	SourceEventTime      time.Time `json:"source_event_time" bson:"source_event_time"`
	SourceTimestampText  string    `json:"source_timestamp_text" bson:"source_timestamp_text"`
	ReceivedTime         time.Time `json:"received_time" bson:"received_time"`
	PersistedTime        time.Time `json:"persisted_time,omitempty" bson:"persisted_time,omitempty"`
	Interval             string    `json:"interval" bson:"interval"`
	Open                 float64   `json:"open" bson:"open"`
	High                 float64   `json:"high" bson:"high"`
	Low                  float64   `json:"low" bson:"low"`
	Close                float64   `json:"close" bson:"close"`
	Volume               uint64    `json:"volume" bson:"volume"`
	EventCount           uint32    `json:"event_count,omitempty" bson:"event_count,omitempty"`
	SourceID             string    `json:"source_id" bson:"source_id"`
	AlpacaMessageType    string    `json:"alpaca_message_type" bson:"alpaca_message_type"`
	PayloadHash          string    `json:"payload_hash" bson:"payload_hash"`
	SourceTimeRegression bool      `json:"source_time_regression" bson:"source_time_regression"`
	DuplicateArrival     bool      `json:"duplicate_arrival" bson:"duplicate_arrival"`
}

func (o Observation) DuplicateKey() string {
	return fmt.Sprintf("%s|%s|%s|%s", o.Symbol, o.SourceTimestampText, o.AlpacaMessageType, o.PayloadHash)
}

func HashPayload(parts ...string) string {
	sum := sha256.New()
	for i, part := range parts {
		if i > 0 {
			_, _ = sum.Write([]byte{0})
		}
		_, _ = sum.Write([]byte(part))
	}
	return hex.EncodeToString(sum.Sum(nil))
}

func HashRaw(raw json.RawMessage) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
