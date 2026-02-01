CREATE TABLE IF NOT EXISTS channel_routes (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(50) UNIQUE NOT NULL,
    channel_id VARCHAR(64) NOT NULL,
    description TEXT,
    emoji VARCHAR(10),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed data for common channels
INSERT INTO channel_routes (keyword, channel_id, description, emoji) VALUES
    ('bugs', '0', 'Report bugs and issues', '🐛'),
    ('changelog', '0', 'View update notes', '📋'),
    ('wiki', '0', 'Documentation and guides', '📚'),
    ('support', '0', 'Get help from the community', '💬')
ON CONFLICT DO NOTHING;
