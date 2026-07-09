package stocks

import (
	"testing"
)

func TestNewID(t *testing.T) {
	id1 := newID()
	id2 := newID()
	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
	if len(id1) != 32 {
		t.Errorf("expected 32 char hex, got %d", len(id1))
	}
}

func TestSimulationHolding_Fields(t *testing.T) {
	sh := SimulationHolding{
		Symbol:           "TEST",
		AllocationPct:    50.0,
		Invested:         1000.0,
		AnnualizedReturn: 15.5,
		CurrentPrice:     150.0,
	}

	if sh.Invested != 1000 {
		t.Errorf("Invested = %f, want 1000", sh.Invested)
	}
	if sh.AnnualizedReturn != 15.5 {
		t.Errorf("AnnualizedReturn = %f, want 15.5", sh.AnnualizedReturn)
	}
}

func TestProjectionSnapshot(t *testing.T) {
	p := ProjectionSnapshot{
		Label:          "5 Years",
		Years:          5,
		ProjectedValue: 20113.57,
		ProjectedGain:  10113.57,
		ProjectedPct:   101.14,
	}

	if p.Years != 5 {
		t.Errorf("Years = %f, want 5", p.Years)
	}
	if p.ProjectedGain < 0 {
		t.Error("ProjectedGain should be positive for positive return")
	}
}
