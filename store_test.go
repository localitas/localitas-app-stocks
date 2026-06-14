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

func TestSimulationHolding_Calculation(t *testing.T) {
	invested := 1000.0
	startPrice := 100.0
	endPrice := 150.0
	shares := invested / startPrice
	value := shares * endPrice
	gain := value - invested
	gainPct := (gain / invested) * 100

	sh := SimulationHolding{
		Symbol:     "TEST",
		Invested:   invested,
		StartPrice: startPrice,
		EndPrice:   endPrice,
		Shares:     shares,
		Value:      value,
		Gain:       gain,
		GainPct:    gainPct,
	}

	if sh.Value != 1500 {
		t.Errorf("Value = %f, want 1500", sh.Value)
	}
	if sh.Gain != 500 {
		t.Errorf("Gain = %f, want 500", sh.Gain)
	}
	if sh.GainPct != 50 {
		t.Errorf("GainPct = %f, want 50", sh.GainPct)
	}
}
