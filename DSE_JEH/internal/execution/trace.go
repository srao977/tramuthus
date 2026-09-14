package execution

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RunType string

const (
	RunTypeA RunType = "RUN_A"
	RunTypeB RunType = "RUN_B"
)

var runTypePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]*$`)

type TraceRecord struct {
	PipelineRunID     string         `bson:"pipeline_run_id"`
	RunType           RunType        `bson:"run_type"`
	CollectionRunID   string         `bson:"collection_run_id"`
	Symbol            string         `bson:"symbol"`
	SequenceNo        int64          `bson:"sequence_no"`
	Stage             string         `bson:"stage"`
	StageOrder        int32          `bson:"stage_order"`
	EventTime         time.Time      `bson:"event_time"`
	EmitType          *string        `bson:"emit_type"`
	Input             map[string]any `bson:"input,omitempty"`
	Output            map[string]any `bson:"output,omitempty"`
	CapitalAllocation map[string]any `bson:"capital_allocation,omitempty"`
}

func ValidateRunType(runType RunType) error {
	if !runTypePattern.MatchString(string(runType)) {
		return fmt.Errorf("run_type must match %s", runTypePattern.String())
	}
	return nil
}

func (record TraceRecord) ValidateRunType() error { return ValidateRunType(record.RunType) }

type TraceWriter interface {
	Write(context.Context, TraceRecord) (bson.ObjectID, error)
	Close(context.Context) error
}
