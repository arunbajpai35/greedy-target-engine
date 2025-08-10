-- campaigns table
CREATE TABLE IF NOT EXISTS campaigns (
    cid TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    img TEXT,
    cta TEXT,
    status TEXT CHECK (status IN ('ACTIVE', 'INACTIVE')) NOT NULL
);

-- targeting_rules table with proper array support
CREATE TABLE IF NOT EXISTS targeting_rules (
    id SERIAL PRIMARY KEY,
    cid TEXT REFERENCES campaigns(cid) ON DELETE CASCADE,
    include_country TEXT[],
    exclude_country TEXT[],
    include_os TEXT[],
    exclude_os TEXT[],
    include_app TEXT[],
    exclude_app TEXT[]
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_cid ON targeting_rules(cid);
