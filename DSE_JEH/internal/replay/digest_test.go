// Package replay tests deterministic evidence digesting for audit comparison.
// Inputs are replay results; output is a stable digest that changes with evidence.
// No runtime or scientific configuration is involved beyond the evidence itself.
// A digest comparison proves byte-level replay equality, not scientific validity.
package replay

import (
	"testing"

	"tramuthus/dse-jeh/internal/evidence"
)

func TestDigestIsStableAndEvidenceSensitive(t *testing.T) {
	result := evidence.RunResult{RankingMode: "PROVING_ONLY_ORDER", HopOnCandidates: []string{"AAPL"}}
	first, err := Digest(result)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Digest(result)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("stable input produced different digests: %s != %s", first, second)
	}
	result.HopOnCandidates = append(result.HopOnCandidates, "MSFT")
	changed, err := Digest(result)
	if err != nil {
		t.Fatal(err)
	}
	if changed == first {
		t.Fatal("changed evidence produced identical digest")
	}
}
