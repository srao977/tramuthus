// Package config tests rejection of invalid experiment identity before side effects.
package config

import "testing"

func TestSeriesSizeValidation(t *testing.T) {
	for _, value := range []string{"0", "-1"} {
		if _, err := Parse([]string{"-series-size", value}); err == nil {
			t.Fatalf("series-size %s should fail", value)
		}
	}
	cfg, err := Parse([]string{"-series-size", "1"})
	if err != nil || cfg.SeriesSize != 1 {
		t.Fatalf("positive series size failed: %+v, %v", cfg, err)
	}
}

func TestDifferentCollectionsRequired(t *testing.T) {
	_, err := Parse([]string{"-series-size", "6", "-raw-collection", "same", "-phase-collection", "same"})
	if err == nil {
		t.Fatal("matching raw and phase collections must fail")
	}
}
