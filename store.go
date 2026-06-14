package stocks

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

const DatabaseName = "stocks"

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

func OpenStore(coreURL, dbID, token string) (*Store, error) {
	dsn := fmt.Sprintf("%s?database_id=%s&token=%s", coreURL, dbID, token)
	db, err := sql.Open("localitas", dsn)
	if err != nil {
		return nil, err
	}
	return NewStore(db), nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) CreatePortfolio(ctx context.Context, userID, name string) (*Portfolio, error) {
	id := newID()
	now := time.Now().UTC().Unix()
	_, err := s.db.ExecContext(ctx, "INSERT INTO portfolios (id, user_id, name, sort_order, created_at, updated_at) VALUES (?, ?, ?, 0, ?, ?)", id, userID, name, now, now)
	if err != nil {
		return nil, err
	}
	return &Portfolio{ID: id, Name: name, CreatedAt: time.Unix(now, 0), UpdatedAt: time.Unix(now, 0)}, nil
}

func (s *Store) ListPortfolios(ctx context.Context, userID string) ([]*Portfolio, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, sort_order, created_at, updated_at FROM portfolios WHERE user_id = ? ORDER BY sort_order, name", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*Portfolio, 0)
	for rows.Next() {
		var p Portfolio
		var createdAt, updatedAt int64
		if err := rows.Scan(&p.ID, &p.Name, &p.SortOrder, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		p.CreatedAt = time.Unix(createdAt, 0)
		p.UpdatedAt = time.Unix(updatedAt, 0)
		out = append(out, &p)
	}
	return out, nil
}

func (s *Store) UpdatePortfolio(ctx context.Context, id, name string) error {
	now := time.Now().UTC().Unix()
	_, err := s.db.ExecContext(ctx, "UPDATE portfolios SET name=?, updated_at=? WHERE id=?", name, now, id)
	return err
}

func (s *Store) DeletePortfolio(ctx context.Context, id string) error {
	s.db.ExecContext(ctx, "DELETE FROM holdings WHERE portfolio_id = ?", id)
	_, err := s.db.ExecContext(ctx, "DELETE FROM portfolios WHERE id = ?", id)
	return err
}

func (s *Store) AddHolding(ctx context.Context, portfolioID, symbol string) (*Holding, error) {
	id := newID()
	now := time.Now().UTC().Unix()
	sortPos := time.Now().UTC().UnixNano()
	_, err := s.db.ExecContext(ctx, `INSERT INTO holdings (id, portfolio_id, symbol, shares, cost_basis, allocation_pct, sort_position, created_at, updated_at) VALUES (?, ?, ?, 0, 0, 0, ?, ?, ?)
		ON CONFLICT(portfolio_id, symbol) DO NOTHING`,
		id, portfolioID, symbol, sortPos, now, now)
	if err != nil {
		return nil, err
	}
	return &Holding{ID: id, PortfolioID: portfolioID, Symbol: symbol, CreatedAt: time.Unix(now, 0), UpdatedAt: time.Unix(now, 0)}, nil
}

func (s *Store) ListHoldings(ctx context.Context, portfolioID string) ([]*Holding, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, portfolio_id, symbol, COALESCE(allocation_pct,0), COALESCE(sort_position,0), created_at, updated_at FROM holdings WHERE portfolio_id = ? ORDER BY sort_position ASC, symbol ASC", portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*Holding, 0)
	for rows.Next() {
		var h Holding
		var createdAt, updatedAt int64
		if err := rows.Scan(&h.ID, &h.PortfolioID, &h.Symbol, &h.AllocationPct, &h.SortPosition, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		h.CreatedAt = time.Unix(createdAt, 0)
		h.UpdatedAt = time.Unix(updatedAt, 0)
		out = append(out, &h)
	}
	return out, nil
}

func (s *Store) UpdateSortPosition(ctx context.Context, id string, position int64) error {
	_, err := s.db.ExecContext(ctx, "UPDATE holdings SET sort_position=? WHERE id=?", position, id)
	return err
}

func (s *Store) UpdateHolding(ctx context.Context, id string, allocationPct float64) error {
	now := time.Now().UTC().Unix()
	_, err := s.db.ExecContext(ctx, "UPDATE holdings SET allocation_pct=?, updated_at=? WHERE id=?", allocationPct, now, id)
	return err
}

func (s *Store) DeleteHolding(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM holdings WHERE id = ?", id)
	return err
}

func (s *Store) GetAllSymbols(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT DISTINCT symbol FROM holdings ORDER BY symbol")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var symbols []string
	for rows.Next() {
		var s string
		rows.Scan(&s)
		symbols = append(symbols, s)
	}
	return symbols, nil
}

func (s *Store) GetCachedFinancials(ctx context.Context, symbol string) ([]FinancialStatement, int64, error) {
	var updatedAt int64
	s.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(updated_at),0) FROM financials_cache WHERE symbol = ?", symbol).Scan(&updatedAt)
	if updatedAt == 0 {
		return nil, 0, nil
	}
	rows, err := s.db.QueryContext(ctx, "SELECT symbol, period_type, end_date, total_revenue, net_income, gross_profit, operating_income, eps_estimate, eps_actual, eps_surprise_pct FROM financials_cache WHERE symbol = ? ORDER BY period_type, end_date DESC", symbol)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var stmts []FinancialStatement
	for rows.Next() {
		var f FinancialStatement
		rows.Scan(&f.Symbol, &f.PeriodType, &f.EndDate, &f.TotalRevenue, &f.NetIncome, &f.GrossProfit, &f.OperatingIncome, &f.EpsEstimate, &f.EpsActual, &f.EpsSurprisePct)
		stmts = append(stmts, f)
	}
	return stmts, updatedAt, nil
}

func (s *Store) SaveFinancials(ctx context.Context, stmts []FinancialStatement) {
	now := time.Now().UTC().Unix()
	for _, f := range stmts {
		s.db.ExecContext(ctx, "INSERT OR REPLACE INTO financials_cache (symbol, period_type, end_date, total_revenue, net_income, gross_profit, operating_income, eps_estimate, eps_actual, eps_surprise_pct, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			f.Symbol, f.PeriodType, f.EndDate, f.TotalRevenue, f.NetIncome, f.GrossProfit, f.OperatingIncome, f.EpsEstimate, f.EpsActual, f.EpsSurprisePct, now)
	}
}

func (s *Store) GetCachedETFHoldings(ctx context.Context, etfSymbol string) ([]ETFHolding, int64, error) {
	var updatedAt int64
	s.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(updated_at),0) FROM etf_holdings_cache WHERE etf_symbol = ?", etfSymbol).Scan(&updatedAt)
	if updatedAt == 0 {
		return nil, 0, nil
	}
	rows, err := s.db.QueryContext(ctx, "SELECT symbol, name, weight FROM etf_holdings_cache WHERE etf_symbol = ? ORDER BY weight DESC", etfSymbol)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var holdings []ETFHolding
	for rows.Next() {
		var h ETFHolding
		rows.Scan(&h.Symbol, &h.Name, &h.Weight)
		holdings = append(holdings, h)
	}
	return holdings, updatedAt, nil
}

func (s *Store) SaveETFHoldings(ctx context.Context, etfSymbol string, holdings []ETFHolding) {
	now := time.Now().UTC().Unix()
	s.db.ExecContext(ctx, "DELETE FROM etf_holdings_cache WHERE etf_symbol = ?", etfSymbol)
	for _, h := range holdings {
		s.db.ExecContext(ctx, "INSERT INTO etf_holdings_cache (etf_symbol, symbol, name, weight, updated_at) VALUES (?, ?, ?, ?, ?)",
			etfSymbol, h.Symbol, h.Name, h.Weight, now)
	}
}

func (s *Store) GetCachedAnalyst(ctx context.Context, symbol string) (*AnalystTargets, int64, error) {
	var t AnalystTargets
	var updatedAt int64
	err := s.db.QueryRowContext(ctx, "SELECT low, mean, median, high, num_analysts, recommendation, updated_at FROM analyst_cache WHERE symbol = ?", symbol).
		Scan(&t.Low, &t.Mean, &t.Median, &t.High, &t.NumberOfAnalysts, &t.Recommendation, &updatedAt)
	if err != nil {
		return nil, 0, err
	}
	return &t, updatedAt, nil
}

func (s *Store) SaveAnalyst(ctx context.Context, symbol string, t *AnalystTargets) {
	now := time.Now().UTC().Unix()
	s.db.ExecContext(ctx, "INSERT OR REPLACE INTO analyst_cache (symbol, low, mean, median, high, num_analysts, recommendation, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		symbol, t.Low, t.Mean, t.Median, t.High, t.NumberOfAnalysts, t.Recommendation, now)
}

func (s *Store) GetCachedChart(ctx context.Context, symbol, interval string) ([]ChartPoint, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT timestamp, open, high, low, close, volume FROM chart_cache WHERE symbol = ? AND interval = ? ORDER BY timestamp ASC",
		symbol, interval)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []ChartPoint
	for rows.Next() {
		var p ChartPoint
		rows.Scan(&p.Timestamp, &p.Open, &p.High, &p.Low, &p.Close, &p.Volume)
		points = append(points, p)
	}
	return points, nil
}

func (s *Store) GetLastCachedTimestamp(ctx context.Context, symbol, interval string) int64 {
	var ts int64
	s.db.QueryRowContext(ctx, "SELECT COALESCE(MAX(timestamp),0) FROM chart_cache WHERE symbol = ? AND interval = ?", symbol, interval).Scan(&ts)
	return ts
}

func (s *Store) SaveChartPoints(ctx context.Context, symbol, interval string, points []ChartPoint) error {
	for _, p := range points {
		s.db.ExecContext(ctx,
			"INSERT OR REPLACE INTO chart_cache (symbol, interval, timestamp, open, high, low, close, volume) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			symbol, interval, p.Timestamp, p.Open, p.High, p.Low, p.Close, p.Volume)
	}
	return nil
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
