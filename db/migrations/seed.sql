-- Clear previous data (optional for dev)
DELETE FROM targeting_rules;
DELETE FROM campaigns;

-- Seed campaigns as per assignment
INSERT INTO campaigns (cid, name, img, cta, status) VALUES
('spotify', 'Spotify - Music for everyone', 'https://somelink', 'Download', 'ACTIVE'),
('duolingo', 'Duolingo: Best way to learn', 'https://somelink2', 'Install', 'ACTIVE'),
('subwaysurfer', 'Subway Surfer', 'https://somelink3', 'Play', 'ACTIVE');

-- Seed targeting rules as per assignment
INSERT INTO targeting_rules (cid, include_country, exclude_country, include_os, exclude_os, include_app, exclude_app) VALUES
('spotify', ARRAY['us', 'canada'], NULL, NULL, NULL, NULL, NULL),
('duolingo', NULL, ARRAY['us'], ARRAY['android', 'ios'], NULL, NULL, NULL),
('subwaysurfer', NULL, NULL, ARRAY['android'], NULL, ARRAY['com.gametion.ludokinggame'], NULL);
