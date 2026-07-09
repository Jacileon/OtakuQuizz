-- Rate limit table for anon messages
CREATE TABLE IF NOT EXISTS anon_rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_anon_rate_limits_key ON anon_rate_limits(key);

ALTER TABLE anon_rate_limits ENABLE ROW LEVEL SECURITY;
CREATE POLICY "anon_rate_limits_full_access" ON anon_rate_limits FOR ALL USING (true) WITH CHECK (true);
