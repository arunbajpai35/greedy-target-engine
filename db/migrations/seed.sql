-- Clear previous data (optional for dev)
DELETE FROM targeting_rules;
DELETE FROM campaigns;

-- Seed campaigns
INSERT INTO campaigns (cid, name, img, cta, status) VALUES
('spotify', 'Spotify - Music for everyone', 'https://somelink', 'Download', 'ACTIVE'),
('subwaysurfer', 'Subway Surfer', 'https://somelink3', 'Play', 'ACTIVE');

-- Seed targeting rules
INSERT INTO targeting_rules (cid, country, os, app) VALUES
('spotify', ARRAY['us', 'canada'], NULL, NULL),
('subwaysurfer', NULL, ARRAY['android'], ARRAY['com.gametion.ludokinggame']);
