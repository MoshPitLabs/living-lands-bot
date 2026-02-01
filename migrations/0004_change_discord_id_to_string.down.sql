-- Rollback: Revert discord_id from VARCHAR(20) back to BIGINT
-- This rollback assumes data can be converted back to BIGINT
-- WARNING: This will fail if any discord_id values exceed BIGINT range

-- Step 1: Create temporary BIGINT column
ALTER TABLE users ADD COLUMN discord_id_old BIGINT;

-- Step 2: Convert string values back to BIGINT
UPDATE users SET discord_id_old = CAST(discord_id AS BIGINT);

-- Step 3: Drop the unique constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_discord_id_unique;

-- Step 4: Drop the VARCHAR column
ALTER TABLE users DROP COLUMN discord_id;

-- Step 5: Rename the BIGINT column back to the original name
ALTER TABLE users RENAME COLUMN discord_id_old TO discord_id;

-- Step 6: Add back the unique constraint
ALTER TABLE users ADD CONSTRAINT users_discord_id_key UNIQUE (discord_id);

-- Step 7: Add NOT NULL constraint
ALTER TABLE users ALTER COLUMN discord_id SET NOT NULL;

-- Verify: All discord_id values should now be BIGINT and not NULL
-- SELECT discord_id, typeof(discord_id) FROM users LIMIT 5;
