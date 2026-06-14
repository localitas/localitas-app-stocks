CREATE TABLE IF NOT EXISTS etf_holdings_cache (
    etf_symbol TEXT NOT NULL,
    symbol TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    weight REAL NOT NULL DEFAULT 0,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (etf_symbol, symbol)
);

CREATE TABLE IF NOT EXISTS analyst_cache (
    symbol TEXT PRIMARY KEY,
    low REAL NOT NULL DEFAULT 0,
    mean REAL NOT NULL DEFAULT 0,
    median REAL NOT NULL DEFAULT 0,
    high REAL NOT NULL DEFAULT 0,
    num_analysts INTEGER NOT NULL DEFAULT 0,
    recommendation TEXT NOT NULL DEFAULT '',
    updated_at INTEGER NOT NULL
);
