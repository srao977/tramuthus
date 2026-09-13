package evidence

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"

	dsejehv1 "tramuthus/dse-jeh-transsat-1/gen/dse_jeh/v1"
)

type PhaseWriter struct {
	mu     sync.Mutex
	file   *os.File
	writer *bufio.Writer
	digest hash.Hash
}

func NewPhaseWriter(path string) (*PhaseWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create phase evidence directory: %w", err)
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create phase evidence output: %w", err)
	}
	return &PhaseWriter{file: file, writer: bufio.NewWriter(file), digest: sha256.New()}, nil
}

func (writer *PhaseWriter) Write(value *dsejehv1.PhaseEvidence) error {
	encoded, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal phase evidence: %w", err)
	}
	encoded = append(encoded, '\n')
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if _, err := writer.digest.Write(encoded); err != nil {
		return fmt.Errorf("digest phase evidence: %w", err)
	}
	if _, err := writer.writer.Write(encoded); err != nil {
		return fmt.Errorf("write phase evidence: %w", err)
	}
	if err := writer.writer.Flush(); err != nil {
		return fmt.Errorf("flush phase evidence: %w", err)
	}
	return nil
}

func (writer *PhaseWriter) Digest() string {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	return hex.EncodeToString(writer.digest.Sum(nil))
}

func (writer *PhaseWriter) Close() error {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if err := writer.writer.Flush(); err != nil {
		_ = writer.file.Close()
		return fmt.Errorf("flush phase evidence: %w", err)
	}
	if err := writer.file.Close(); err != nil {
		return fmt.Errorf("close phase evidence: %w", err)
	}
	return nil
}
