-- Migration: Change discord_id from BIGINT to VARCHAR(20)
-- Reason: discordgo v0.29.0 provides Discord IDs as strings
--         Storing as int64 and converting is fragile and error-prone
--
-- This migration safely converts the existing BIGINT data to VARCHAR

-- Step 1: Create new column with VARCHAR type
ALTER TABLE users ADD COLUMN discord_id_new VARCHAR(20);

-- Step 2: Copy and convert existing data from BIGINT to string
UPDATE users SET discord_id_new = CAST(discord_id AS VARCHAR(20)) WHERE discord_id IS NOT NULL;

-- Step 3: Make sure all values were converted (no NULLs unless original was NULL)
-- No NULLs should exist since discord_id was NOT NULL

-- Step 4: Drop the old unique constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_discord_id_key;

-- Step 5: Drop the old BIGINT column
ALTER TABLE users DROP COLUMN discord_id;

-- Step 6: Rename the new column to the original name
ALTER TABLE users RENAME COLUMN discord_id_new TO discord_id;

-- Step 7: Add back the unique constraint
ALTER TABLE users ADD CONSTRAINT users_discord_id_unique UNIQUE (discord_id);

-- Step 8: Add NOT NULL constraint (since original was NOT NULL)
ALTER TABLE users ALTER COLUMN discord_id SET NOT NULL;

-- Verify: All discord_id values should now be strings and not NULL
-- SELECT discord_id, typeof(discord_id) FROM users LIMIT 5;
