package runtime

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/encoding/protojson"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
	"tramuthus/dse-jeh-transsat-1/internal/adapter/mongo"
	"tramuthus/dse-jeh-transsat-1/internal/admission"
	"tramuthus/dse-jeh-transsat-1/internal/analytical"
	"tramuthus/dse-jeh-transsat-1/internal/config"
	"tramuthus/dse-jeh-transsat-1/internal/input/offline"
	"tramuthus/dse-jeh-transsat-1/internal/input/online"
	"tramuthus/dse-jeh-transsat-1/internal/validation/phasecompare"
)

type Stats struct {
	Read         uint64 `json:"read"`
	Admitted     uint64 `json:"admitted"`
	Initializing uint64 `json:"initializing"`
	Observable   uint64 `json:"observable"`
	Invalid      uint64 `json:"invalid"`
	Duplicate    uint64 `json:"duplicate"`
	Conflict     uint64 `json:"conflict"`
	Gap          uint64 `json:"gap"`
	OutOfOrder   uint64 `json:"out_of_order"`
	Rejected     uint64 `json:"rejected"`
	Digest       string `json:"digest"`
	OutputPath   string `json:"output_path"`
}

type Engine struct {
	admitter    *admission.Admitter
	coordinator *analytical.Coordinator
	writer      *bufio.Writer
	file        *os.File
	digest      hash.Hash
	stats       Stats
}

func Run(ctx context.Context, cfg config.Config) (Stats, error) {
	engine, err := newEngine(cfg.OutputPath)
	if err != nil {
		return Stats{}, err
	}
	defer engine.close()

	switch cfg.Mode {
	case dsejehv1.RuntimeMode_RUNTIME_MODE_OFFLINE:
		repository, err := mongo.Open(ctx, cfg.MongoURI, cfg.MongoDatabase, cfg.MongoCollection)
		if err != nil {
			return Stats{}, err
		}
		defer repository.Close(context.Background())
		events, err := offline.New(repository, cfg.CollectionRunID).Events(ctx)
		if err != nil {
			return Stats{}, err
		}
		for _, event := range events {
			if err := engine.process(event); err != nil {
				return Stats{}, err
			}
		}
	case dsejehv1.RuntimeMode_RUNTIME_MODE_ONLINE:
		consumer := online.New(cfg.GRPCAddress, cfg.Symbols, cfg.MaxBars, cfg.FinalizedOnly)
		if err := consumer.Stream(ctx, engine.process); err != nil {
			return Stats{}, err
		}
	default:
		return Stats{}, fmt.Errorf("unsupported runtime mode %s", cfg.Mode)
	}
	if err := engine.writer.Flush(); err != nil {
		return Stats{}, fmt.Errorf("flush evidence: %w", err)
	}
	engine.stats.Digest = hex.EncodeToString(engine.digest.Sum(nil))
	if cfg.ReferenceCSV != "" {
		report, err := phasecompare.Compare(cfg.ReferenceCSV, cfg.OutputPath, 1e-9)
		if err != nil {
			return Stats{}, err
		}
		if err := phasecompare.WriteReport(cfg.ComparisonReport, report); err != nil {
			return Stats{}, err
		}
	}
	return engine.stats, nil
}

func newEngine(outputPath string) (*Engine, error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("create evidence output: %w", err)
	}
	return &Engine{
		admitter:    admission.New(),
		coordinator: analytical.NewCoordinator(),
		writer:      bufio.NewWriter(file),
		file:        file,
		digest:      sha256.New(),
		stats:       Stats{OutputPath: outputPath},
	}, nil
}

func (engine *Engine) process(event *dsejehv1.BarEvent) error {
	engine.stats.Read++
	admissionEvidence := engine.admitter.Admit(event)
	switch admissionEvidence.GetStatus() {
	case dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_ADMITTED:
		engine.stats.Admitted++
	case dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_DUPLICATE:
		engine.stats.Duplicate++
	case dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_CONFLICT:
		engine.stats.Conflict++
	case dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_GAP:
		engine.stats.Gap++
	case dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_OUT_OF_ORDER:
		engine.stats.OutOfOrder++
	case dsejehv1.BarAdmissionStatus_BAR_ADMISSION_STATUS_INVALID:
		engine.stats.Invalid++
	default:
		engine.stats.Rejected++
	}
	phaseEvidence := engine.coordinator.Process(event, admissionEvidence)
	if phaseEvidence == nil {
		return nil
	}
	switch phaseEvidence.GetStatus() {
	case dsejehv1.PhaseStatus_PHASE_STATUS_INITIALIZING:
		engine.stats.Initializing++
	case dsejehv1.PhaseStatus_PHASE_STATUS_OBSERVABLE:
		engine.stats.Observable++
	case dsejehv1.PhaseStatus_PHASE_STATUS_INVALID:
		engine.stats.Invalid++
	}
	encoded, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(phaseEvidence)
	if err != nil {
		return fmt.Errorf("marshal phase evidence: %w", err)
	}
	encoded = append(encoded, '\n')
	_, _ = engine.digest.Write(encoded)
	if _, err := engine.writer.Write(encoded); err != nil {
		return fmt.Errorf("write phase evidence: %w", err)
	}
	return nil
}

func (engine *Engine) close() {
	if engine == nil {
		return
	}
	_ = engine.writer.Flush()
	_ = engine.file.Close()
}
