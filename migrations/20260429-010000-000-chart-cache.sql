CREATE TABLE IF NOT EXISTS chart_cache (
    symbol TEXT NOT NULL,
    interval TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    open REAL NOT NULL DEFAULT 0,
    high REAL NOT NULL DEFAULT 0,
    low REAL NOT NULL DEFAULT 0,
    close REAL NOT NULL DEFAULT 0,
    volume INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (symbol, interval, timestamp)
);

CREATE INDEX IF NOT EXISTS idx_chart_cache_symbol ON chart_cache(symbol, interval);
