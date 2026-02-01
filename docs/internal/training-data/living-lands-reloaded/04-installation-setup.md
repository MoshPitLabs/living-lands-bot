# Living Lands Reloaded - Installation & Setup Guide

Complete guide for installing and configuring Living Lands Reloaded on your Hytale server.

---

## For Players

### Joining a Server Running Living Lands Reloaded

**Good news:** You don't need to install anything!

When you join a Hytale server running Living Lands Reloaded:

1. **The HUD appears automatically** - Metabolism stats (Hunger/Thirst/Energy) display on your screen
2. **Professions start tracking** - You begin earning XP in all five professions immediately
3. **No client-side mods required** - Everything runs server-side

**First Steps:**
```
/ll stats          # Toggle metabolism HUD if you want to hide it
/ll professions    # View your profession levels
/ll progress       # See a compact XP progress bar
```

That's it! You're ready to play.

---

## For Server Owners

### Prerequisites

**System Requirements:**
- **Hytale Server:** Version 1.0+ (adjust based on actual Hytale release requirements)
- **Java:** Java 21+ (recommended Java 25 for best performance)
- **Memory:** Minimum 2GB RAM allocated to server (4GB+ recommended)
- **Storage:** 50MB for mod + configs
- **Permissions:** Operator or permissions plugin (LuckPerms, etc.)

**Compatible With:**
- Hytale server plugins (API compatibility pending)
- Most world generation mods
- Permissions plugins (LuckPerms, PermissionsEx)

**NOT Compatible With:**
- Other custom UI mods (not tested, use at your own risk)
- Other player stat modification mods (not tested, may conflict)

---

## Installation Steps

### Step 1: Download the Mod

**Where to Download:**
- **CurseForge:** [Living Lands Reloaded](https://www.curseforge.com/hytale/mods/living-lands-reloaded)
- **GitHub Releases:** [https://github.com/MoshPitCodes/living-lands-reloaded/releases](https://github.com/MoshPitCodes/living-lands-reloaded/releases)

**Current Version:** 1.3.0  
**File Name:** `livinglands-reloaded-1.3.0.jar`

**Verify Download:**
- Check file size: ~2-5MB (varies by version)
- Verify file extension is `.jar`
- Ensure you downloaded the latest version

---

### Step 2: Install the JAR File

1. **Locate your server's plugins folder:**
   ```
   /path/to/your/server/plugins/
   ```

2. **Copy the JAR file:**
   - Place `livinglands-reloaded-1.3.0.jar` into the `plugins/` folder

3. **Verify placement:**
   ```
   plugins/
   ├── livinglands-reloaded-1.3.0.jar
   └── ... (other plugins)
   ```

**DO NOT:**
- Rename the JAR file
- Extract/unzip the JAR file
- Place it in any other folder

---

### Step 3: Start the Server

**First-Time Startup:**

1. Start your Hytale server:
   ```bash
   ./start.sh
   # or
   java -Xmx4G -Xms2G -jar hytale-server.jar
   ```

2. **Watch the console** for Living Lands Reloaded startup messages:
   ```
   [INFO] [LivingLandsReloaded] Enabling LivingLandsReloaded v1.3.0
   [INFO] [LivingLandsReloaded] Generating default configuration files...
   [INFO] [LivingLandsReloaded] Metabolism system initialized
   [INFO] [LivingLandsReloaded] Profession system initialized
   [INFO] [LivingLandsReloaded] Living Lands Reloaded enabled successfully!
   ```

3. **Config files auto-generate** in:
   ```
   plugins/LivingLandsReloaded/config/
   ├── metabolism.yml
   ├── professions.yml
   ├── announcer.yml
   └── settings.yml
   ```

**If startup fails:**
- Check console for error messages
- Verify Java version (Java 21+)
- Check file permissions (server must be able to write to `plugins/` folder)
- See [Troubleshooting](#troubleshooting) section

---

### Step 4: Verify Installation

**In-Game Verification:**

1. Join your server as an operator

2. Run these commands:
   ```
   /ll stats
   /ll professions
   /ll reload all
   ```

3. **Expected Results:**
   - Metabolism HUD appears on screen
   - Professions panel shows all 5 professions at Level 1
   - Reload command confirms configs loaded

4. **Test metabolism:**
   - Sprint around - watch Energy drain
   - Wait idle - watch Hunger/Thirst drain slowly
   - Eat food - watch Hunger restore

5. **Test professions:**
   - Kill a mob - Combat XP increases
   - Mine a block - Mining XP increases
   - Pick up an item - Gathering XP increases

**If verification fails:**
- See [Troubleshooting](#troubleshooting) section below

---

## Configuration

### Config File Locations

All config files are in:
```
plugins/LivingLandsReloaded/config/
```

**Files:**
- **`metabolism.yml`** - Hunger/Thirst/Energy drain rates, effects, per-world overrides
- **`professions.yml`** - XP rates, death penalty, tier abilities
- **`announcer.yml`** - MOTD, welcome messages, recurring announcements
- **`settings.yml`** - General mod settings, HUD defaults, logging

---

### Quick Configuration Guide

#### Adjust Metabolism Drain Rates

**File:** `metabolism.yml`

**Example: Make survival 50% easier**
```yaml
metabolism:
  hunger:
    drain_rates:
      sprinting: 0.4   # Changed from 0.8
  thirst:
    drain_rates:
      sprinting: 0.5   # Changed from 1.0
  energy:
    drain_rates:
      sprinting: 0.75  # Changed from 1.5
```

**Apply changes:**
```
/ll reload metabolism
```

---

#### Increase Profession XP Rates

**File:** `professions.yml`

**Example: Double XP gain**
```yaml
professions:
  xp_rates:
    combat:
      mob_kill_base: 20    # Doubled from 10
    mining:
      iron_ore: 20         # Doubled from 10
    logging:
      log_block: 6         # Doubled from 3
```

**Apply changes:**
```
/ll reload professions
```

---

#### Customize Welcome Messages

**File:** `announcer.yml`

**Example: Change MOTD**
```yaml
announcer:
  motd:
    enabled: true
    message: "&6Welcome to Our Server! &aPowered by Living Lands Reloaded"
  welcome:
    first_join:
      enabled: true
      message: "&aWelcome, {player_name}! This is your first adventure!"
```

**Apply changes:**
```
/ll reload announcer
```

---

## Per-World Configuration

You can override settings for specific worlds.

### Example: Disable Metabolism in Creative World

**File:** `metabolism.yml`

```yaml
world_overrides:
  world_creative:  # Must match your world folder name EXACTLY
    metabolism:
      enabled: false
```

### Example: Harsher Survival in Hardcore World

**File:** `metabolism.yml`

```yaml
world_overrides:
  world_hardcore:
    metabolism:
      hunger:
        drain_rates:
          sprinting: 1.6  # Double the default
      death_penalty:
        base_percentage: 20  # Double death penalty
```

**World Names:**
- Find your world folder names in: `server_root/world_name/`
- Examples: `world`, `world_nether`, `world_the_end`, `world_creative`
- Names are **case-sensitive**

---

## Troubleshooting

### Mod Won't Load

**Symptom:** No console messages from Living Lands Reloaded on server start

**Possible Causes:**
1. JAR file in wrong folder
2. JAR file corrupted during download
3. Incompatible Hytale server version

**Solutions:**
- Verify JAR is in `plugins/` folder
- Re-download the mod from official source
- Check mod version matches server version
- Review console for error messages

---

### Configs Not Generating

**Symptom:** `plugins/LivingLandsReloaded/config/` folder is empty

**Possible Causes:**
1. Server doesn't have write permissions
2. Mod failed to initialize
3. Config generation disabled

**Solutions:**
- Check folder permissions: `chmod -R 755 plugins/LivingLandsReloaded/`
- Review console errors during startup
- Delete the `LivingLandsReloaded/` folder and restart (configs will regenerate)

---

### Metabolism Not Draining

**Symptom:** Stats stay at 100%, never drain

**Possible Causes:**
1. Metabolism disabled in config
2. World override disabling it
3. Player has admin bypass

**Solutions:**
- Check `metabolism.yml`: Ensure `enabled: true`
- Check world overrides: Remove or fix world-specific disables
- Run `/ll debug metabolism <player>` to view current drain rates

---

### Professions Not Gaining XP

**Symptom:** Kill mobs, mine blocks, but XP stays at 0

**Possible Causes:**
1. XP rates set to 0 in config
2. Player data corrupted
3. Profession tracking disabled

**Solutions:**
- Check `professions.yml`: Verify XP rates > 0
- Run `/ll prof show <player>` to view current progression
- Reset player professions: `/ll prof reset <player> all` (WARNING: deletes all progress)

---

### Hot Reload Not Working

**Symptom:** `/ll reload` runs but changes don't apply

**Possible Causes:**
1. YAML syntax error in config file
2. Wrong module name
3. Cached values not clearing

**Solutions:**
- Validate YAML syntax: [YAML Validator](https://www.yamllint.com/)
- Use exact module names: `metabolism`, `professions`, `announcer`, `all`
- Restart server if reload fails

---

### Players Can't See HUD

**Symptom:** Metabolism/profession panels don't display

**Possible Causes:**
1. Player toggled panels off
2. UI rendering issue
3. Incompatible client-side mods

**Solutions:**
- Tell player to run: `/ll stats` and `/ll professions` (toggles them back on)
- Have player restart their client
- Check for conflicting UI mods on client side

---

### High Server Lag

**Symptom:** Server TPS drops significantly after installing mod

**Possible Causes:**
1. Metabolism ticking too frequently
2. Too many players with active professions
3. Database queries inefficient

**Solutions:**
- Increase tick intervals in `settings.yml` (reduce update frequency)
- Allocate more RAM to server
- Monitor console for performance warnings
- Report issue to developer with `/ll debug performance`

---

## Upgrading from Older Versions

### Upgrading from 1.2.x to 1.3.0

1. **Stop the server**

2. **Backup everything:**
   ```bash
   cp -r plugins/LivingLandsReloaded/ backup/
   ```

3. **Replace the JAR file:**
   ```bash
   rm plugins/livinglands-reloaded-1.2.x.jar
   cp livinglands-reloaded-1.3.0.jar plugins/
   ```

4. **Start the server**
   - New configs auto-generate
   - Old configs merge automatically

5. **Review new features:**
   - 1.3.0 added Announcer module
   - Check `announcer.yml` for new settings

6. **Test in-game:**
   ```
   /ll reload all
   ```

**Breaking Changes:**
- None in 1.3.0 (backward compatible)

**New Config Files:**
- `announcer.yml` (auto-generated)

---

### Migrating Player Data

Player profession data is stored in:
```
plugins/LivingLandsReloaded/data/players.db
```

**To migrate data to a new server:**

1. **On old server:**
   ```bash
   cp plugins/LivingLandsReloaded/data/players.db backup/
   ```

2. **On new server:**
   ```bash
   cp backup/players.db plugins/LivingLandsReloaded/data/
   ```

3. **Restart server**

**Warning:** Database format may change between major versions (1.x → 2.x). Always backup first.

---

## Uninstallation

### Complete Removal

1. **Stop the server**

2. **Remove the JAR file:**
   ```bash
   rm plugins/livinglands-reloaded-1.3.0.jar
   ```

3. **Optionally, delete config folder:**
   ```bash
   rm -rf plugins/LivingLandsReloaded/
   ```

4. **Start the server**

**Player Data:**
- Profession progress is stored in `LivingLandsReloaded/data/players.db`
- **If you delete this folder, ALL player progression is lost**
- Backup first if you might reinstall later

---

## Performance Optimization

### For Small Servers (1-10 players)

**Default settings are optimal** - no changes needed.

---

### For Medium Servers (10-50 players)

**Increase tick intervals to reduce CPU load:**

**File:** `settings.yml`

```yaml
performance:
  metabolism_tick_interval: 20   # Default: 10 (1 tick = 0.05 sec)
  profession_tick_interval: 40   # Default: 20
```

**Impact:** Slightly less precise drain/XP tracking, but better performance

---

### For Large Servers (50+ players)

**Aggressive optimization:**

```yaml
performance:
  metabolism_tick_interval: 40   # Reduced precision
  profession_tick_interval: 60   # Reduced precision
  database_batch_updates: true   # Batch writes to database
  cache_player_data: true        # Cache in memory
```

**Also consider:**
- Allocate more RAM: `-Xmx6G -Xms4G` (or higher)
- Use faster storage (SSD over HDD)
- Run database cleanup: `/ll admin cleanup-database`

---

## Backup Best Practices

### What to Backup

**Essential:**
- `plugins/LivingLandsReloaded/data/players.db` - Player profession data
- `plugins/LivingLandsReloaded/config/` - All config files

**Optional:**
- `livinglands-reloaded-1.3.0.jar` - The mod file itself (can re-download)

---

### Automated Backup Script

**Example backup script (Linux):**

```bash
#!/bin/bash
# backup-livinglands.sh

DATE=$(date +%Y-%m-%d_%H-%M-%S)
BACKUP_DIR="/path/to/backups/livinglands"

mkdir -p "$BACKUP_DIR"

cp -r plugins/LivingLandsReloaded/ "$BACKUP_DIR/backup_$DATE/"

echo "Living Lands Reloaded backed up to: $BACKUP_DIR/backup_$DATE/"
```

**Run daily via cron:**
```bash
crontab -e

# Add this line:
0 3 * * * /path/to/backup-livinglands.sh
```

(Runs every day at 3:00 AM)

---

## Getting Help

### Support Channels

**Bug Reports:**
- GitHub Issues: [https://github.com/MoshPitCodes/living-lands-reloaded/issues](https://github.com/MoshPitCodes/living-lands-reloaded/issues)

**Discord Community:**
- [https://discord.gg/8jgMj9GPsq](https://discord.gg/8jgMj9GPsq)

**Documentation:**
- This guide
- `README.md` in GitHub repo
- In-game: `/ll help`

---

### Before Asking for Help

1. **Check this guide** for common issues
2. **Read the error message** in server console
3. **Verify your setup:**
   - Correct Hytale server version
   - Mod file in `plugins/` folder
   - Java 21+ installed
4. **Try basic troubleshooting:**
   - Restart server
   - Reload configs: `/ll reload all`
   - Test with default configs

**When reporting bugs, include:**
- Mod version (e.g., 1.3.0)
- Hytale server version
- Full error message from console
- Steps to reproduce the issue
- Config files (if customized)

---

## Summary

**Quick Install:**
1. Download `livinglands-reloaded-1.3.0.jar`
2. Place in `plugins/` folder
3. Start server
4. Configs auto-generate in `plugins/LivingLandsReloaded/config/`

**Quick Config:**
1. Edit `metabolism.yml` or `professions.yml`
2. Run `/ll reload <module>`
3. Changes apply instantly

**Quick Troubleshooting:**
- Mod won't load? Check console errors
- Configs missing? Check folder permissions
- Hot reload fails? Validate YAML syntax

**Backup:** `players.db` + `config/` folder before major changes

**Support:** GitHub Issues or Discord for help

**Current Version:** 1.3.0 (stable)
