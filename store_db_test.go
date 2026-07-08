package stocks

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const testSchema = `
CREATE TABLE IF NOT EXISTS portfolios (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS holdings (
    id TEXT PRIMARY KEY,
    portfolio_id TEXT NOT NULL,
    symbol TEXT NOT NULL,
    shares REAL NOT NULL DEFAULT 0,
    cost_basis REAL NOT NULL DEFAULT 0,
    allocation_pct REAL DEFAULT 0,
    sort_position INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (portfolio_id) REFERENCES portfolios(id) ON DELETE CASCADE,
    UNIQUE(portfolio_id, symbol)
);
`

func newTestStore(t *testing.T) *Store {
	t.Helper()
	f, err := os.CreateTemp("", "stocks-test-*.sqlitedb")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()
	t.Cleanup(func() { os.Remove(path) })

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec(testSchema); err != nil {
		t.Fatal(err)
	}

	return NewStore(db)
}

func TestListAllPortfolios(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.CreatePortfolio(ctx, "user-aaa", "Tech Stocks")
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.CreatePortfolio(ctx, "user-bbb", "Dividend Portfolio")
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.CreatePortfolio(ctx, "user-ccc", "Growth Fund")
	if err != nil {
		t.Fatal(err)
	}

	all, err := store.ListAllPortfolios(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(all) != 3 {
		t.Fatalf("expected 3 portfolios, got %d", len(all))
	}

	names := make(map[string]bool)
	for _, p := range all {
		names[p.Name] = true
	}
	for _, want := range []string{"Tech Stocks", "Dividend Portfolio", "Growth Fund"} {
		if !names[want] {
			t.Errorf("expected portfolio %q in results", want)
		}
	}
}

func TestListAllPortfolios_Empty(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	all, err := store.ListAllPortfolios(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(all) != 0 {
		t.Errorf("expected 0 portfolios, got %d", len(all))
	}
}
