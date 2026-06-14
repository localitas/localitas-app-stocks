CREATE TABLE IF NOT EXISTS financials_cache (
    symbol TEXT NOT NULL,
    period_type TEXT NOT NULL,
    end_date TEXT NOT NULL,
    total_revenue INTEGER NOT NULL DEFAULT 0,
    net_income INTEGER NOT NULL DEFAULT 0,
    gross_profit INTEGER NOT NULL DEFAULT 0,
    operating_income INTEGER NOT NULL DEFAULT 0,
    eps_estimate REAL NOT NULL DEFAULT 0,
    eps_actual REAL NOT NULL DEFAULT 0,
    eps_surprise_pct REAL NOT NULL DEFAULT 0,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (symbol, period_type, end_date)
);
