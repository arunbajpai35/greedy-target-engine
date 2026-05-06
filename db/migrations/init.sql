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

CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_cid ON targeting_rules(cid);

-- GIN indexes power the @> (array contains) probes used by the matcher.
-- Without these, every match scans the whole rules table.
CREATE INDEX IF NOT EXISTS idx_targeting_rules_include_country ON targeting_rules USING GIN (include_country);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_exclude_country ON targeting_rules USING GIN (exclude_country);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_include_os      ON targeting_rules USING GIN (include_os);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_exclude_os      ON targeting_rules USING GIN (exclude_os);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_include_app     ON targeting_rules USING GIN (include_app);
CREATE INDEX IF NOT EXISTS idx_targeting_rules_exclude_app     ON targeting_rules USING GIN (exclude_app);

-- NOTIFY targeting_changes on any write to either table; the app uses this to
-- invalidate its in-memory cache.
CREATE OR REPLACE FUNCTION notify_targeting_change() RETURNS trigger AS $$
BEGIN
    PERFORM pg_notify('targeting_changes', TG_TABLE_NAME);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS targeting_rules_notify ON targeting_rules;
CREATE TRIGGER targeting_rules_notify
    AFTER INSERT OR UPDATE OR DELETE ON targeting_rules
    FOR EACH STATEMENT EXECUTE FUNCTION notify_targeting_change();

DROP TRIGGER IF EXISTS campaigns_notify ON campaigns;
CREATE TRIGGER campaigns_notify
    AFTER INSERT OR UPDATE OR DELETE ON campaigns
    FOR EACH STATEMENT EXECUTE FUNCTION notify_targeting_change();
