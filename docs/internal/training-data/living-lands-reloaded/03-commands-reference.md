# Living Lands Reloaded - Commands Reference

Complete reference for all `/ll` commands in Living Lands Reloaded.

---

## Table of Contents

- [Player Commands](#player-commands)
- [Admin Commands](#admin-commands)
- [Command Permissions](#command-permissions)
- [Common Usage Examples](#common-usage-examples)

---

## Player Commands

These commands are available to all players.

### `/ll stats`

**Purpose:** Toggle the metabolism HUD panel

**Usage:** `/ll stats`

**Effect:**
- Shows/hides the metabolism HUD panel
- Panel displays: Hunger, Thirst, Energy levels
- Settings are saved per-player

**Example:**
```
/ll stats
```
**Output:** `Metabolism HUD toggled OFF` or `Metabolism HUD toggled ON`

---

### `/ll buffs`

**Purpose:** Toggle the active buffs display

**Usage:** `/ll buffs`

**Effect:**
- Shows/hides active buffs on your HUD
- Displays profession buffs (e.g., Adrenaline Rush, Ore Sense)
- Displays metabolism buffs (e.g., high energy speed boost)

**Example:**
```
/ll buffs
```
**Output:** `Buffs display toggled OFF` or `Buffs display toggled ON`

---

### `/ll debuffs`

**Purpose:** Toggle the active debuffs display

**Usage:** `/ll debuffs`

**Effect:**
- Shows/hides active debuffs on your HUD
- Displays metabolism debuffs (e.g., low hunger damage, low energy speed penalty)
- Helps you track what's hurting your character

**Example:**
```
/ll debuffs
```
**Output:** `Debuffs display toggled OFF` or `Debuffs display toggled ON`

---

### `/ll professions`

**Purpose:** Toggle the professions panel

**Usage:** `/ll professions`

**Effect:**
- Shows/hides the full professions panel
- Displays all 5 professions with levels and XP progress
- Shows unlocked tier abilities

**Example:**
```
/ll professions
```
**Output:** `Professions panel toggled OFF` or `Professions panel toggled ON`

---

### `/ll progress`

**Purpose:** Toggle the compact professions progress panel

**Usage:** `/ll progress`

**Effect:**
- Shows/hides a smaller, compact XP progress display
- Only shows XP bars without full details
- Less screen space than full professions panel

**Example:**
```
/ll progress
```
**Output:** `Progress panel toggled OFF` or `Progress panel toggled ON`

---

## Admin Commands

These commands require operator permissions.

### `/ll reload [module]`

**Purpose:** Reload configuration files without restarting the server

**Permission:** `livinglands.admin.reload`

**Usage:** `/ll reload <module>`

**Modules:**
- `metabolism` - Reload metabolism config
- `professions` - Reload professions config
- `announcer` - Reload announcer/MOTD config
- `all` - Reload all configs

**Examples:**
```
/ll reload metabolism
```
**Output:** `Metabolism configuration reloaded successfully`

```
/ll reload all
```
**Output:** `All configurations reloaded successfully`

**When to Use:**
- After editing any `.yml` config file
- To apply changes without server restart
- Testing config changes live

**Note:** Players will see updated values immediately (e.g., new drain rates, XP rates)

---

### `/ll broadcast <message>`

**Purpose:** Broadcast a message to all online players

**Permission:** `livinglands.admin.broadcast`

**Usage:** `/ll broadcast <message>`

**Features:**
- Sends message to every player on the server
- Supports color codes (`&a`, `&6`, `&c`, etc.)
- Supports placeholders: `{player_name}`, `{server_name}`

**Examples:**
```
/ll broadcast Server restart in 5 minutes!
```
**Output (to all players):** `[Server] Server restart in 5 minutes!`

```
/ll broadcast &6Special Event: &aDouble XP for 1 hour!
```
**Output (to all players):** `[Server] Special Event: Double XP for 1 hour!` (with colors)

**Color Codes:**
- `&0` = Black
- `&1` = Dark Blue
- `&2` = Dark Green
- `&3` = Dark Aqua
- `&4` = Dark Red
- `&5` = Dark Purple
- `&6` = Gold
- `&7` = Gray
- `&8` = Dark Gray
- `&9` = Blue
- `&a` = Green
- `&b` = Aqua
- `&c` = Red
- `&d` = Light Purple
- `&e` = Yellow
- `&f` = White
- `&l` = Bold
- `&o` = Italic
- `&r` = Reset

---

### `/ll prof show <player>`

**Purpose:** View a player's profession levels and XP

**Permission:** `livinglands.admin.professions`

**Usage:** `/ll prof show <player>`

**Example:**
```
/ll prof show MoshPit
```

**Output:**
```
=== MoshPit's Professions ===
Combat: Level 87 (12,450 / 15,000 XP) [Tier 3: Adrenaline Rush]
Mining: Level 72 (8,230 / 10,500 XP) [Tier 2: +10 Thirst]
Logging: Level 45 (4,100 / 6,000 XP) [Tier 2: +10 Energy]
Building: Level 30 (2,300 / 4,000 XP) [Tier 1: +15% XP]
Gathering: Level 19 (890 / 1,500 XP) [Tier 1: +15% XP]
```

**Use Cases:**
- Check player progression
- Verify XP gain after config changes
- Help troubleshoot profession issues

---

### `/ll prof add <player> <profession> <amount>`

**Purpose:** Add XP to a player's profession

**Permission:** `livinglands.admin.professions`

**Usage:** `/ll prof add <player> <profession> <amount>`

**Professions:** `combat`, `mining`, `logging`, `building`, `gathering`

**Examples:**
```
/ll prof add MoshPit combat 500
```
**Output:** `Added 500 XP to MoshPit's Combat profession (now Level 88)`

```
/ll prof add Steve mining 10000
```
**Output:** `Added 10,000 XP to Steve's Mining profession (now Level 65)`

**Use Cases:**
- Reward players for events
- Compensate for lost XP due to bugs
- Testing profession abilities
- Admin bonus for good behavior

**Note:** Adding XP can cause level-ups and unlock new tier abilities

---

### `/ll prof set <player> <profession> <level>`

**Purpose:** Set a player's profession to a specific level

**Permission:** `livinglands.admin.professions`

**Usage:** `/ll prof set <player> <profession> <level>`

**Level Range:** 1-100

**Examples:**
```
/ll prof set MoshPit gathering 100
```
**Output:** `Set MoshPit's Gathering profession to Level 100 (Tier 3: Survivalist unlocked)`

```
/ll prof set Alex combat 15
```
**Output:** `Set Alex's Combat profession to Level 15 (Tier 1: +15% XP unlocked)`

**Use Cases:**
- Fast-track players to test high-level abilities
- Reset a profession to a specific level after issues
- Admin privileges for trusted players
- Testing tier abilities

**Warning:** This sets the level directly and resets XP progress to 0 for that level

---

### `/ll prof reset <player> <profession>`

**Purpose:** Reset a specific profession to Level 1

**Permission:** `livinglands.admin.professions`

**Usage:** `/ll prof reset <player> <profession>`

**Examples:**
```
/ll prof reset MoshPit combat
```
**Output:** `Reset MoshPit's Combat profession to Level 1 (0 XP)`

```
/ll prof reset Steve mining
```
**Output:** `Reset Steve's Mining profession to Level 1 (0 XP)`

**Use Cases:**
- Player requested a fresh start
- Fix progression bugs
- Punishment for rule violations
- Testing progression from scratch

**Warning:** This is irreversible and removes all tier abilities for that profession

---

### `/ll prof reset <player> all`

**Purpose:** Reset ALL of a player's professions to Level 1

**Permission:** `livinglands.admin.professions`

**Usage:** `/ll prof reset <player> all`

**Example:**
```
/ll prof reset MoshPit all
```

**Output:**
```
Reset ALL professions for MoshPit:
- Combat: Level 1 (0 XP)
- Mining: Level 1 (0 XP)
- Logging: Level 1 (0 XP)
- Building: Level 1 (0 XP)
- Gathering: Level 1 (0 XP)
```

**Use Cases:**
- Player requested complete restart
- Switching to new season/wipe
- Severe rule violation
- Testing fresh player experience

**Warning:** This is IRREVERSIBLE and removes ALL profession progress

---

## Command Permissions

### Permission Nodes

```yaml
permissions:
  # Player commands (default: true for all players)
  livinglands.player.stats: true
  livinglands.player.buffs: true
  livinglands.player.debuffs: true
  livinglands.player.professions: true
  livinglands.player.progress: true
  
  # Admin commands (default: operator only)
  livinglands.admin.reload: op
  livinglands.admin.broadcast: op
  livinglands.admin.professions: op
  livinglands.admin.debug: op
```

### Example: Grant Reload Permission to Moderators

If using a permissions plugin (e.g., LuckPerms):

```
/lp group moderator permission set livinglands.admin.reload true
```

Now moderators can reload configs without full operator permissions.

---

## Common Usage Examples

### Scenario 1: Player Wants to Hide HUD

**Player Issue:** "The HUD is blocking my view"

**Solution:**
```
/ll stats
/ll professions
```

**Result:** Metabolism and profession panels hidden, clean screen

---

### Scenario 2: Testing Config Changes

**Admin Task:** You edited `metabolism.yml` to reduce drain rates

**Solution:**
```
/ll reload metabolism
```

**Result:** New drain rates apply instantly, no server restart needed

**Verification:**
```
/ll debug metabolism <player>
```
(Check if new drain rates are active)

---

### Scenario 3: Event Reward

**Admin Task:** Give all players 1000 XP in Combat for event participation

**Solution:**
```
/ll prof add Player1 combat 1000
/ll prof add Player2 combat 1000
/ll prof add Player3 combat 1000
```

**Alternative:** Use a script to loop through online players (requires plugin/scripting)

---

### Scenario 4: Player Lost XP Due to Bug

**Player Report:** "I died and lost 50% XP instead of 10%!"

**Admin Response:**
1. Check their professions:
```
/ll prof show PlayerName
```

2. Add back the lost XP:
```
/ll prof add PlayerName combat 5000
```

3. Investigate the bug and check death penalty config

---

### Scenario 5: Testing Tier 3 Abilities

**Admin Task:** Test the Survivalist ability (Gathering Level 100)

**Solution:**
```
/ll prof set TestPlayer gathering 100
```

**Result:** Player instantly gets Tier 3 Survivalist (-15% metabolism drain)

**Testing:** Monitor metabolism drain rates to verify the effect works

---

### Scenario 6: Season Wipe

**Admin Task:** Reset all player professions for new season

**Solution:**
```
/ll prof reset Player1 all
/ll prof reset Player2 all
/ll prof reset Player3 all
```

**Alternative:** Use database script to wipe all profession data at once (advanced)

---

### Scenario 7: Broadcast Server Event

**Admin Task:** Announce double XP event with colors

**Solution:**
```
/ll broadcast &6&l[EVENT] &aDouble XP Weekend! &eAll professions earn 2x XP for 48 hours!
```

**Result:** Colorful announcement to all players:
`[EVENT] Double XP Weekend! All professions earn 2x XP for 48 hours!`

---

## Command Aliases

Living Lands Reloaded supports these command aliases:

- `/ll` - Main command (recommended)
- `/livinglands` - Full name (verbose)
- `/llr` - Short form (quick typing)

**All aliases work identically:**
```
/ll stats
/livinglands stats
/llr stats
```

---

## Tab Completion

All commands support tab completion for easier use:

**Examples:**
```
/ll prof <TAB>
→ show, add, set, reset

/ll prof add MoshPit <TAB>
→ combat, mining, logging, building, gathering

/ll reload <TAB>
→ metabolism, professions, announcer, all
```

---

## Error Messages

### Common Errors and Solutions

**Error:** `Unknown command. Type "/help" for help.`  
**Cause:** Mod not installed or not loaded  
**Solution:** Verify Living Lands Reloaded is in `plugins/` folder and server started successfully

---

**Error:** `You do not have permission to use this command.`  
**Cause:** Player lacks required permission  
**Solution:** Grant permission via permissions plugin or make player operator

---

**Error:** `Player not found: <name>`  
**Cause:** Player is offline or name is misspelled  
**Solution:** Verify player is online (for some commands) or check spelling

---

**Error:** `Invalid profession: <name>`  
**Cause:** Profession name misspelled  
**Solution:** Use exact names: `combat`, `mining`, `logging`, `building`, `gathering`

---

**Error:** `Invalid level: must be between 1 and 100`  
**Cause:** Level number out of range  
**Solution:** Use a number between 1-100

---

**Error:** `Failed to reload configuration: <reason>`  
**Cause:** YAML syntax error in config file  
**Solution:** Fix the YAML syntax error (check indentation, colons, spacing)

---

## Best Practices

### For Players

1. **Customize your HUD:** Toggle panels you don't use to reduce clutter
2. **Use `/ll progress`:** Compact panel if you only want XP tracking
3. **Check buffs/debuffs:** Know what's affecting you in combat
4. **Don't spam commands:** Panel toggles are instant, no need to repeat

### For Admins

1. **Always reload after config changes:** Use `/ll reload <module>` instead of restarting
2. **Test changes on yourself first:** Use `/ll prof set` to test abilities before affecting players
3. **Document XP rewards:** Keep track of event XP bonuses for fairness
4. **Use `/ll prof show` for support:** Check player progression before making changes
5. **Backup before mass resets:** Use `/ll prof reset all` carefully - it's irreversible

### For Server Owners

1. **Set up permissions:** Grant moderators `/ll reload` and `/ll broadcast` only
2. **Limit profession admin commands:** Only trusted admins should use `/ll prof set/reset`
3. **Configure tab completion:** Ensure tab completion works for easier admin workflow
4. **Log commands:** Enable command logging to track admin actions

---

## Summary

**Player Commands (5 total):**
- `/ll stats` - Toggle metabolism HUD
- `/ll buffs` - Toggle buffs display
- `/ll debuffs` - Toggle debuffs display
- `/ll professions` - Toggle full professions panel
- `/ll progress` - Toggle compact XP progress panel

**Admin Commands (6 total):**
- `/ll reload [module]` - Reload configs
- `/ll broadcast <message>` - Broadcast to all players
- `/ll prof show <player>` - View professions
- `/ll prof add <player> <profession> <amount>` - Add XP
- `/ll prof set <player> <profession> <level>` - Set level
- `/ll prof reset <player> <profession|all>` - Reset profession(s)

**Most Used:**
- Players: `/ll stats`, `/ll professions`
- Admins: `/ll reload all`, `/ll prof show`

**Hot Reload:** `/ll reload metabolism` or `/ll reload professions` after config edits
