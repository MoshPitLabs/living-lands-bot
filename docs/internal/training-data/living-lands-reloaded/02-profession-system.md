# Living Lands Reloaded - Profession System

## Overview

The profession system turns your playstyle into character progression. Instead of generic leveling, you gain XP and unlock abilities in **five distinct professions** based on how you actually play the game.

All players progress in all five professions simultaneously, but at different rates depending on their activities.

---

## The Five Professions

### ⚔️ Combat
**XP Source:** Killing hostile mobs and players (PvP if enabled)

**Philosophy:** Rewards fighters and defenders

**Progression:**
- **Level 15 (Tier 1):** +15% Combat XP gain
- **Level 45 (Tier 2):** +15 max hunger capacity
- **Level 100 (Tier 3):** **Adrenaline Rush** - +10% movement speed for 5 seconds after every kill

---

### ⛏️ Mining
**XP Source:** Mining ore blocks (coal, iron, gold, diamond, etc.)

**Philosophy:** Rewards resource gatherers and miners

**Progression:**
- **Level 15 (Tier 1):** +15% Mining XP gain
- **Level 45 (Tier 2):** +10 max thirst capacity
- **Level 100 (Tier 3):** **Ore Sense** - +10% ore drop rate (more ores from each block)

---

### 🪓 Logging
**XP Source:** Chopping down tree logs

**Philosophy:** Rewards woodcutters and lumber workers

**Progression:**
- **Level 15 (Tier 1):** +15% Logging XP gain
- **Level 45 (Tier 2):** +10 max energy capacity
- **Level 100 (Tier 3):** **Timber!** - +25% extra logs from each tree block

---

### 🏗️ Building
**XP Source:** Placing blocks in the world

**Philosophy:** Rewards builders and architects

**Progression:**
- **Level 15 (Tier 1):** +15% Building XP gain
- **Level 45 (Tier 2):** +15 max stamina capacity (API pending - feature coming soon)
- **Level 100 (Tier 3):** **Efficient Architect** - 12% chance blocks aren't consumed when placed

---

### 🌾 Gathering
**XP Source:** Picking up items from the ground

**Philosophy:** Rewards explorers and foragers

**Progression:**
- **Level 15 (Tier 1):** +15% Gathering XP gain
- **Level 45 (Tier 2):** +4 hunger and +4 thirst restored whenever you pick up food
- **Level 100 (Tier 3):** **Survivalist** - -15% metabolism depletion on ALL stats (hunger/thirst/energy)

---

## Level Progression

### XP Requirements
Each profession has **100 levels** with progressively increasing XP requirements.

**Approximate XP needed:**
- Level 1 → 15: Fast progression (200-500 XP per level)
- Level 15 → 45: Moderate pace (500-1500 XP per level)
- Level 45 → 100: Slow grind (1500-5000+ XP per level)

Total time to max a profession: **30-50 hours of focused gameplay** (varies by XP rate config)

### XP Sources Table

| Profession | XP Source | Base XP | Notes |
|------------|-----------|---------|-------|
| Combat | Kill mob | 10-50 | Varies by mob difficulty |
| Combat | Kill player | 100 | PvP only (if enabled) |
| Mining | Mine coal ore | 5 | Common ores |
| Mining | Mine iron ore | 10 | Uncommon ores |
| Mining | Mine diamond ore | 50 | Rare ores |
| Logging | Chop log block | 3 | All wood types |
| Building | Place any block | 1 | All placeable blocks |
| Gathering | Pick up any item | 1 | All item pickups |

**Note:** XP values are configurable in `professions.yml`

---

## Tier Abilities Explained

### Tier 1 (Level 15) - XP Boost

**Effect:** +15% XP gain in that profession

**How it Works:**
- Once you reach level 15 in a profession, you earn XP 15% faster
- Example: Mining normally gives 10 XP per iron ore → With Tier 1, you get 11.5 XP
- **This does NOT apply retroactively** - only to XP earned after reaching level 15

**Strategy:** Unlock Tier 1 quickly in your favorite profession to speed up later progression

---

### Tier 2 (Level 45) - Stat Boosts

**Effect:** Permanent increases to metabolism stats or stamina

**Abilities:**
- **Combat:** +15 max hunger (100 → 115)
- **Mining:** +10 max thirst (100 → 110)
- **Logging:** +10 max energy (100 → 110)
- **Gathering:** +4 hunger & +4 thirst on every food pickup
- **Building:** +15 max stamina (feature pending Hytale stamina API)

**How it Works:**
- These are **permanent passive effects**
- Stack with profession combinations (e.g., max Combat + max Mining = +15 hunger, +10 thirst)
- Gathering's Tier 2 is **active** - triggers when you pick up food items

**Strategy:** Higher max values = more buffer before penalties kick in

---

### Tier 3 (Level 100) - Special Abilities

**Effect:** Powerful unique abilities that define your playstyle

#### ⚔️ Adrenaline Rush (Combat)
- **Trigger:** On every mob/player kill
- **Effect:** +10% movement speed for 5 seconds
- **Stacks:** No (refresh duration on new kill)
- **Use Case:** Chaining kills in combat, quick escapes after fights

#### ⛏️ Ore Sense (Mining)
- **Trigger:** Passive (always active)
- **Effect:** +10% ore drop rate
- **How it Works:** 10% chance to get an extra ore from each ore block mined
- **Use Case:** More efficient resource gathering, better diamond yields

#### 🪓 Timber! (Logging)
- **Trigger:** Passive (always active)
- **Effect:** +25% extra logs from each tree block
- **How it Works:** 25% chance to get an extra log from each log block chopped
- **Use Case:** Faster wood farming, more building materials

#### 🏗️ Efficient Architect (Building)
- **Trigger:** On every block placement
- **Effect:** 12% chance the block isn't consumed from your inventory
- **How it Works:** Random chance - you still place the block, but it doesn't leave your inventory
- **Use Case:** Save materials on large builds, conserve rare blocks

#### 🌾 Survivalist (Gathering)
- **Trigger:** Passive (always active)
- **Effect:** -15% metabolism depletion on ALL stats
- **How it Works:** All hunger/thirst/energy drain rates reduced by 15%
- **Example:** Sprinting normally drains 0.8 hunger/min → With Survivalist, only 0.68/min
- **Use Case:** Easiest survival, less food/water pressure, longer exploration range

---

## Death Penalty System

### How It Works

When you die, you lose XP from your **2 highest professions**.

**Penalty Formula:**
```
Base Penalty: 10% of current level XP
Progressive Increase: +3% per death
Cap: 35% maximum penalty
```

**Example Progression:**
- **1st death:** Lose 10% XP from top 2 professions
- **2nd death:** Lose 13% XP (10% + 3%)
- **3rd death:** Lose 16% XP (10% + 6%)
- **4th death:** Lose 19% XP (10% + 9%)
- ...
- **9th+ death:** Lose 35% XP (capped)

### Penalty Decay

The progressive penalty decays over time to prevent permanent harsh penalties.

**Decay System:**
- **Decay Rate:** -1% per 30 minutes of playtime without dying
- **Example:** If you're at 25% penalty (6 deaths), it decays to 24% after 30 min, 23% after 1 hour, etc.
- **Resets to Base:** Eventually returns to 10% base penalty if you survive long enough

### Mercy System

**First-Time Mercy:**
- Your **very first death ever** has **no penalty** (tutorial grace period)
- Announced in chat: *"You've been granted mercy this time..."*

**Low-Level Mercy:**
- If you're below level 10 in a profession, the penalty is reduced by 50%
- Example: 10% penalty → 5% penalty for levels 1-9

### Which Professions Are Affected?

**Always the top 2 highest-level professions.**

**Example:**
- Combat: Level 87
- Mining: Level 72
- Logging: Level 45
- Building: Level 30
- Gathering: Level 19

**On death:** You lose XP from **Combat** and **Mining** only (the two highest).

**Strategy:** This encourages balanced progression - if you over-level one profession, it becomes your primary death risk.

---

## Configuration

### Config File Location
`plugins/LivingLandsReloaded/config/professions.yml`

### Customizable Settings

```yaml
professions:
  xp_rates:
    combat:
      mob_kill_base: 10      # Base XP per mob kill
      player_kill: 100       # XP per player kill (PvP)
    mining:
      coal_ore: 5
      iron_ore: 10
      gold_ore: 20
      diamond_ore: 50
    logging:
      log_block: 3           # XP per log chopped
    building:
      block_place: 1         # XP per block placed
    gathering:
      item_pickup: 1         # XP per item picked up
      
  death_penalty:
    enabled: true
    base_percentage: 10      # Starting penalty %
    progressive_increase: 3  # Added % per death
    maximum_penalty: 35      # Capped penalty %
    decay_rate: 1            # % decay per decay_interval
    decay_interval: 1800     # Seconds (1800 = 30 minutes)
    first_death_mercy: true  # No penalty on first ever death
    low_level_reduction: 0.5 # 50% reduction for levels 1-9
    
  tier_abilities:
    tier1_xp_boost: 0.15     # 15% XP boost at level 15
    tier2_stat_boosts:
      combat_hunger: 15
      mining_thirst: 10
      logging_energy: 10
      gathering_food_restore: 4
      building_stamina: 15   # Pending API
    tier3_abilities:
      adrenaline_rush:
        speed_boost: 0.10    # 10% speed
        duration: 5          # Seconds
      ore_sense:
        bonus_chance: 0.10   # 10% extra ore
      timber:
        bonus_chance: 0.25   # 25% extra logs
      efficient_architect:
        save_chance: 0.12    # 12% no-consume
      survivalist:
        metabolism_reduction: 0.15  # -15% drain
```

### Example: Double All XP Rates

```yaml
professions:
  xp_rates:
    combat:
      mob_kill_base: 20      # Doubled from 10
      player_kill: 200       # Doubled from 100
    mining:
      coal_ore: 10           # Doubled from 5
      iron_ore: 20           # Doubled from 10
      diamond_ore: 100       # Doubled from 50
    logging:
      log_block: 6           # Doubled from 3
    building:
      block_place: 2         # Doubled from 1
    gathering:
      item_pickup: 2         # Doubled from 1
```

### Example: Reduce Death Penalty Harshness

```yaml
professions:
  death_penalty:
    base_percentage: 5       # Reduced from 10
    progressive_increase: 2  # Reduced from 3
    maximum_penalty: 20      # Reduced from 35
    decay_rate: 2            # Faster decay (2% instead of 1%)
    decay_interval: 900      # Faster interval (15 min instead of 30 min)
```

### Example: Disable Death Penalty Completely

```yaml
professions:
  death_penalty:
    enabled: false
```

---

## Admin Commands

### Viewing Professions

```
/ll prof show <player>
```
**Example:** `/ll prof show MoshPit`

**Output:**
```
Combat: Level 87 (12,450 / 15,000 XP)
Mining: Level 72 (8,230 / 10,500 XP)
Logging: Level 45 (4,100 / 6,000 XP)
Building: Level 30 (2,300 / 4,000 XP)
Gathering: Level 19 (890 / 1,500 XP)
```

---

### Adding XP

```
/ll prof add <player> <profession> <amount>
```

**Example:** `/ll prof add MoshPit combat 500`  
**Result:** Adds 500 XP to player's Combat profession

---

### Setting Level Directly

```
/ll prof set <player> <profession> <level>
```

**Example:** `/ll prof set MoshPit mining 100`  
**Result:** Sets player's Mining profession to level 100 (maxed)

---

### Resetting a Profession

```
/ll prof reset <player> <profession>
```

**Example:** `/ll prof reset MoshPit gathering`  
**Result:** Resets Gathering profession to level 1, 0 XP

---

### Resetting ALL Professions

```
/ll prof reset <player> all
```

**Example:** `/ll prof reset MoshPit all`  
**Result:** Resets ALL five professions to level 1

---

## Player Commands

```
/ll professions
```
**Effect:** Toggles the professions panel display on/off

```
/ll progress
```
**Effect:** Toggles the compact profession progress panel (shows XP bars only)

---

## Frequently Asked Questions

### Can I reset my professions?

**Player:** No, players cannot reset their own professions.  
**Admin:** Yes, admins can use `/ll prof reset <player> <profession>`

### Do I lose profession abilities when I die?

**No.** You only lose XP, not levels or abilities.  
**Exception:** If XP loss causes you to drop below a tier threshold (e.g., level 100 → 99), you temporarily lose that tier's ability until you level back up.

### Can I level all professions to 100?

**Yes!** There's no restriction. You can max all five professions.

**Time Investment:** Expect 30-50 hours per profession = ~150-250 hours to max everything.

### Which profession should I prioritize?

**Depends on playstyle:**

- **Combat-focused player:** Combat → Logging (for energy) → Gathering (for survivalist)
- **Miner:** Mining → Logging (energy for mining) → Gathering (survivalist)
- **Builder:** Building → Mining (resources) → Logging (wood)
- **Explorer:** Gathering → Combat (for fights) → Logging (energy for travel)

**Best long-term combo:** Gathering (survivalist) + your main playstyle profession

### Does Survivalist stack with metabolism config changes?

**Yes!** The -15% reduction applies AFTER your config drain rates.

**Example:**
- Config: Sprinting drains 0.8 hunger/min
- With Survivalist: 0.8 × 0.85 = 0.68 hunger/min
- If you also reduce config to 0.4: 0.4 × 0.85 = 0.34 hunger/min

They multiply, not add.

---

## Progression Tips

### Early Game (Levels 1-15)
- **Goal:** Reach Tier 1 in your main profession for +15% XP boost
- **Strategy:** Focus on one profession - don't spread XP too thin
- **Best starter:** Gathering (easiest to level early, Survivalist is very strong)

### Mid Game (Levels 15-45)
- **Goal:** Unlock Tier 2 stat boosts in 2-3 professions
- **Strategy:** Balance progression - get Combat + your main profession to 45
- **Priority:** Combat (+15 hunger) or Gathering (+4 on food pickup) for survival

### Late Game (Levels 45-100)
- **Goal:** Max out 1-2 professions for Tier 3 abilities
- **Strategy:** Grind your main profession first, then diversify
- **Best Tier 3:** Survivalist (Gathering) is universally useful

### Death Risk Management
- Keep professions roughly balanced to spread death penalty risk
- If one profession is way ahead, deaths hurt it more
- Farm XP cautiously when close to level thresholds (14, 44, 99)

---

## Summary

The profession system rewards how you play:

- **5 professions:** Combat, Mining, Logging, Building, Gathering
- **100 levels each** with XP from normal gameplay
- **3 tiers of abilities:** XP boost (15), stat boost (45), special ability (100)
- **Death penalty:** Lose 10-35% XP from top 2 professions on death
- **Fully configurable:** XP rates, penalties, ability values

**Most important decision:** Which profession to max first?  
**Best all-around choice:** Gathering (Survivalist -15% metabolism drain)

**Hot Reload:** `/ll reload professions`
