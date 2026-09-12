// Package replay drives admitted phase evidence through the operational proving strategy path.
// Inputs are already validated, entity-ordered CSV phase records; output is deterministic run evidence.
// There are no scientific parameters here: the strategy version is owned by evidence/strategy.
// Replay does not calculate phase, synchronize entities, rank scientifically, or calculate P&L.
package replay

import (
	"fmt"

	"tramuthus/dse-jeh/internal/evidence"
	"tramuthus/dse-jeh/internal/strategy"
)

// Run applies every admitted row to the same strategy engine used by the proving command.
func Run(inputs []evidence.PhaseEvidence) (evidence.RunResult, error) {
	engine := strategy.NewEngine()
	for _, input := range inputs {
		if err := engine.Apply(input); err != nil {
			return evidence.RunResult{}, fmt.Errorf("apply %s sequence %d: %w", input.Symbol, input.GeneratorSequence, err)
		}
	}
	return engine.Result(inputs), nil
}
