package stocks

import (
	"testing"
)

func TestSearchSymbol(t *testing.T) {
	results, err := SearchSymbol("AAPL")
	if err != nil {
		t.Fatalf("SearchSymbol: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results for AAPL")
	}
	found := false
	for _, r := range results {
		if r.Symbol == "AAPL" {
			found = true
		}
	}
	if !found {
		t.Error("AAPL not found in search results")
	}
}

func TestFetchQuotes(t *testing.T) {
	quotes, err := FetchQuotes([]string{"AAPL"})
	if err != nil {
		t.Fatalf("FetchQuotes: %v", err)
	}
	if len(quotes) == 0 {
		t.Fatal("expected at least 1 quote")
	}
	if quotes[0].Price <= 0 {
		t.Errorf("AAPL price = %f, want > 0", quotes[0].Price)
	}
	if quotes[0].Symbol != "AAPL" {
		t.Errorf("symbol = %s, want AAPL", quotes[0].Symbol)
	}
}

func TestFetchChart(t *testing.T) {
	points, err := FetchChart("AAPL", "5d", "")
	if err != nil {
		t.Fatalf("FetchChart: %v", err)
	}
	if len(points) == 0 {
		t.Fatal("expected chart points")
	}
	if points[0].Close <= 0 {
		t.Errorf("first close = %f, want > 0", points[0].Close)
	}
}

func TestFetchETFHoldings(t *testing.T) {
	holdings, err := FetchETFHoldings("QQQ")
	if err != nil {
		t.Skipf("ETF holdings API may be gated: %v", err)
	}
	if len(holdings) == 0 {
		t.Skip("no ETF holdings returned")
	}
	if holdings[0].Symbol == "" {
		t.Error("first holding symbol is empty")
	}
	if holdings[0].Weight <= 0 {
		t.Errorf("first holding weight = %f, want > 0", holdings[0].Weight)
	}
}

func TestFetchAnalystTargets(t *testing.T) {
	targets, err := FetchAnalystTargets("AAPL")
	if err != nil {
		t.Skipf("analyst targets API may be gated: %v", err)
	}
	if targets == nil {
		t.Skip("no analyst targets returned")
	}
	if targets.Median <= 0 {
		t.Errorf("median target = %f, want > 0", targets.Median)
	}
	if targets.NumberOfAnalysts <= 0 {
		t.Errorf("number of analysts = %d, want > 0", targets.NumberOfAnalysts)
	}
}

func TestFetchQuotes_Empty(t *testing.T) {
	quotes, err := FetchQuotes(nil)
	if err != nil {
		t.Fatalf("FetchQuotes nil: %v", err)
	}
	if quotes != nil {
		t.Errorf("expected nil, got %d quotes", len(quotes))
	}
}

func TestFetchFinancials(t *testing.T) {
	stmts, err := FetchFinancials("AAPL")
	if err != nil {
		t.Skipf("financials API may be gated: %v", err)
	}
	if len(stmts) == 0 {
		t.Skip("no financial data returned")
	}
	var annual, quarterly, earnings int
	for _, s := range stmts {
		switch s.PeriodType {
		case "annual":
			annual++
		case "quarterly":
			quarterly++
		case "earnings":
			earnings++
		}
	}
	if annual == 0 {
		t.Error("expected annual statements")
	}
	if quarterly == 0 {
		t.Error("expected quarterly statements")
	}
	if earnings == 0 {
		t.Error("expected earnings history")
	}
}

func TestFetchQuotes_MultipleSymbols(t *testing.T) {
	quotes, err := FetchQuotes([]string{"AAPL", "MSFT", "GOOGL"})
	if err != nil {
		t.Fatalf("FetchQuotes: %v", err)
	}
	if len(quotes) < 3 {
		t.Errorf("expected 3 quotes, got %d", len(quotes))
	}
}
