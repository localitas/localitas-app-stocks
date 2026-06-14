ALTER TABLE portfolios ADD COLUMN user_id TEXT NOT NULL DEFAULT '';
UPDATE portfolios SET user_id = '2b9af8b9-856a-4710-9ab6-58fe4eccdf24' WHERE user_id = '';
CREATE INDEX IF NOT EXISTS idx_portfolios_user ON portfolios(user_id);
