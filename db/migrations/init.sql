-- campaigns table
CREATE TABLE IF NOT EXISTS campaigns (
    cid TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    img TEXT,
    cta TEXT,
    status TEXT CHECK (status IN ('ACTIVE', 'PAUSED')) NOT NULL
);

-- targeting_rules table
CREATE TABLE IF NOT EXISTS targeting_rules (
    id SERIAL PRIMARY KEY,
    cid TEXT REFERENCES campaigns(cid),
    app TEXT,
    country TEXT,
    os TEXT
);
