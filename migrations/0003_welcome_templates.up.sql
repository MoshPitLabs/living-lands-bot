CREATE TABLE IF NOT EXISTS welcome_templates (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    weight INTEGER DEFAULT 1,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed lore-friendly welcome messages
INSERT INTO welcome_templates (message, weight, active) VALUES
    ('Welcome, traveler {username}! The lands of Orbis await your arrival.', 1, true),
    ('Greetings, {username}! The spirits whisper of your coming to these Living Lands.', 1, true),
    ('Hail, {username}! Another soul ventures into the realm of mystery and adventure.', 1, true),
    ('The Elder Sage senses a new presence... Welcome, {username}, to our humble gathering.', 2, true),
    ('A wanderer approaches! Welcome, {username}. May your journey here be filled with wonder.', 1, true)
ON CONFLICT DO NOTHING;
