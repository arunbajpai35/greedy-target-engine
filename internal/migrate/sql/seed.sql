INSERT INTO campaigns (cid, name, img, cta, status) VALUES
('spotify', 'Spotify - Music for everyone', 'https://somelink', 'Download', 'ACTIVE'),
('duolingo', 'Duolingo: Best way to learn', 'https://somelink2', 'Install', 'ACTIVE'),
('subwaysurfer', 'Subway Surfer', 'https://somelink3', 'Play', 'ACTIVE')
ON CONFLICT (cid) DO NOTHING;

INSERT INTO targeting_rules (cid, include_country, exclude_country, include_os, exclude_os, include_app, exclude_app)
SELECT * FROM (VALUES
    ('spotify',      ARRAY['us', 'canada']::text[], NULL::text[], NULL::text[], NULL::text[], NULL::text[], NULL::text[]),
    ('duolingo',     NULL::text[], ARRAY['us']::text[], ARRAY['android', 'ios']::text[], NULL::text[], NULL::text[], NULL::text[]),
    ('subwaysurfer', NULL::text[], NULL::text[], ARRAY['android']::text[], NULL::text[], ARRAY['com.gametion.ludokinggame']::text[], NULL::text[])
) AS v(cid, include_country, exclude_country, include_os, exclude_os, include_app, exclude_app)
WHERE NOT EXISTS (SELECT 1 FROM targeting_rules tr WHERE tr.cid = v.cid);
