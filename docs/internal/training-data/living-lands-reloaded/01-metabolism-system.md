# Living Lands Reloaded - Metabolism System

## Overview

The metabolism system adds survival pressure through three interconnected stats: **Hunger**, **Thirst**, and **Energy**. Each stat drains based on player activity and has unique effects when low or high.

All three stats are displayed on the in-game HUD and can be toggled with `/ll stats`.

---

## The Three Stats

### 🍖 Hunger
**Purpose:** Food pressure

**Mechanics:**
- Drains from sprinting, combat, and general activity
- Different foods restore hunger by different amounts
- Low hunger prevents natural health regeneration

**Effects:**
- **Below 30:** Natural health regeneration stops
- **Below 10:** Player takes 0.5 damage per second
- **At 0:** Continuous damage until hunger is restored

**How to Restore:**
- Eat food items
- Different foods restore different amounts
- Gathering profession (Tier 2) grants +4 hunger on food pickup

---

### 💧 Thirst
**Purpose:** Travel pressure

**Mechanics:**
- Drains from walking, sprinting, and traveling
- Drains faster in hot biomes
- Affects stamina regeneration when low

**Effects:**
- **Below 20:** Stamina regeneration reduced by 50%
- **Below 10:** Cannot sprint
- **At 0:** Severe stamina penalties

**How to Restore:**
- Drink water bottles
- Consume juices, soups, and other liquids
- Gathering profession (Tier 2) grants +4 thirst on food pickup
- Mining profession (Tier 2) grants +10 max thirst capacity

---

### ⚡ Energy
**Purpose:** Effort pressure

**Mechanics:**
- Drains from combat, mining, logging, and sprinting
- Drains the fastest of the three stats during intensive activities
- Provides movement speed buffs/debuffs based on level

**Effects:**
- **Above 80:** +15% movement speed (buff)
- **Below 20:** -25% movement speed (debuff)
- **Below 10:** Mining speed reduced by 30%
- **At 0:** Severe penalties to all actions

**How to Restore:**
- Coffee and energy foods
- Slow passive regeneration while idle/resting
- Logging profession (Tier 2) grants +10 max energy capacity

---

## Configuration

### Config File Location

**Server Path:** `plugins/LivingLandsReloaded/config/metabolism.yml`

This file auto-generates on first server start with default values.

### Default Configuration Structure

```yaml
metabolism:
  enabled: true  # Enable/disable the entire metabolism system
  
  hunger:
    max_value: 100
    starting_value: 100
    drain_rates:
      idle: 0.1          # Hunger lost per minute while standing still
      walking: 0.3       # Per minute while walking
      sprinting: 0.8     # Per minute while sprinting
      combat: 1.2        # Per minute during combat
      mining: 0.4        # Per minute while mining
    effects:
      damage_threshold: 10           # Start taking damage below this value
      damage_amount: 0.5             # Damage per tick when below threshold
      damage_interval: 20            # Ticks between damage (20 = 1 second)
      regen_stop_threshold: 30       # Stop health regen below this value
      
  thirst:
    max_value: 100
    starting_value: 100
    drain_rates:
      idle: 0.15         # Thirst lost per minute while standing still
      walking: 0.4       # Per minute while walking
      sprinting: 1.0     # Per minute while sprinting
      combat: 0.5        # Per minute during combat
      hot_biome_multiplier: 1.5  # Multiply drain rates in hot biomes (desert, savanna)
    effects:
      stamina_penalty_threshold: 20  # Stamina penalty below this value
      stamina_penalty_amount: 0.5    # 50% reduction to stamina regen
      sprint_block_threshold: 10     # Cannot sprint below this value
      
  energy:
    max_value: 100
    starting_value: 100
    drain_rates:
      idle: 0.05         # Energy lost per minute while standing still
      walking: 0.2       # Per minute while walking
      sprinting: 1.5     # Per minute while sprinting
      combat: 2.0        # Per minute during combat
      mining: 1.8        # Per minute while mining blocks
      logging: 1.6       # Per minute while chopping trees
    effects:
      speed_buff_threshold: 80       # Movement speed buff above this value
      speed_buff_amount: 0.15        # +15% movement speed when buffed
      speed_debuff_threshold: 20     # Movement speed debuff below this value
      speed_debuff_amount: 0.25      # -25% movement speed when debuffed
      mining_penalty_threshold: 10   # Mining speed penalty below this value
      mining_penalty_amount: 0.30    # -30% mining speed when penalized
```

---

## How to Customize Metabolism Rates

### Question: "How can I tailor the metabolism depletion rates to my liking?"

**Answer:** Edit the `metabolism.yml` configuration file and adjust the `drain_rates` values for each stat.

**Step-by-Step Guide:**

1. **Locate the config file:**
   - Navigate to `plugins/LivingLandsReloaded/config/metabolism.yml`

2. **Open the file in a text editor**

3. **Find the stat you want to adjust** (hunger, thirst, or energy)

4. **Modify the drain_rates values:**
   - **Higher values** = faster drain (harder survival)
   - **Lower values** = slower drain (easier survival)

5. **Save the file**

6. **Reload the config in-game:**
   ```
   /ll reload metabolism
   ```

7. **Test and adjust** as needed

---

## Customization Examples

### Example 1: Make Survival Easier (50% Slower Drain)

```yaml
metabolism:
  hunger:
    drain_rates:
      idle: 0.05         # Changed from 0.1
      walking: 0.15      # Changed from 0.3
      sprinting: 0.4     # Changed from 0.8
      combat: 0.6        # Changed from 1.2
      mining: 0.2        # Changed from 0.4
      
  thirst:
    drain_rates:
      idle: 0.075        # Changed from 0.15
      walking: 0.2       # Changed from 0.4
      sprinting: 0.5     # Changed from 1.0
      combat: 0.25       # Changed from 0.5
      
  energy:
    drain_rates:
      idle: 0.025        # Changed from 0.05
      walking: 0.1       # Changed from 0.2
      sprinting: 0.75    # Changed from 1.5
      combat: 1.0        # Changed from 2.0
      mining: 0.9        # Changed from 1.8
      logging: 0.8       # Changed from 1.6
```

---

### Example 2: Hardcore Mode (Double Drain + Harsher Penalties)

```yaml
metabolism:
  hunger:
    drain_rates:
      idle: 0.2          # Doubled
      walking: 0.6       # Doubled
      sprinting: 1.6     # Doubled
      combat: 2.4        # Doubled
      mining: 0.8        # Doubled
    effects:
      damage_threshold: 20   # Take damage earlier (changed from 10)
      damage_amount: 1.0     # Double damage (changed from 0.5)
      regen_stop_threshold: 50  # Stop regen earlier (changed from 30)
      
  thirst:
    drain_rates:
      idle: 0.3          # Doubled
      walking: 0.8       # Doubled
      sprinting: 2.0     # Doubled
      combat: 1.0        # Doubled
    effects:
      stamina_penalty_threshold: 40  # Penalty starts earlier (changed from 20)
      sprint_block_threshold: 20     # Can't sprint earlier (changed from 10)
      
  energy:
    drain_rates:
      idle: 0.1          # Doubled
      walking: 0.4       # Doubled
      sprinting: 3.0     # Doubled
      combat: 4.0        # Doubled
      mining: 3.6        # Doubled
      logging: 3.2       # Doubled
```

---

### Example 3: Creative Mode (No Metabolism)

If you want to disable metabolism in a specific world (like a creative world):

```yaml
world_overrides:
  world_creative:  # Replace with your world name
    metabolism:
      enabled: false
```

---

### Example 4: Desert Survival Challenge

Make thirst drain much faster in desert biomes:

```yaml
world_overrides:
  world_desert:  # Replace with your world name
    metabolism:
      thirst:
        drain_rates:
          hot_biome_multiplier: 3.0  # Triple drain in hot biomes (changed from 1.5)
        max_value: 150               # Higher max capacity to compensate
```

---

### Example 5: Combat-Focused Survival

Make combat drain energy and hunger significantly, but reduce travel/mining drain:

```yaml
metabolism:
  hunger:
    drain_rates:
      idle: 0.05
      walking: 0.1
      sprinting: 0.4
      combat: 3.0        # Tripled combat drain
      mining: 0.2
      
  energy:
    drain_rates:
      idle: 0.05
      walking: 0.1
      sprinting: 0.8
      combat: 5.0        # Significantly increased combat drain
      mining: 0.9        # Reduced mining drain
      logging: 0.8       # Reduced logging drain
```

---

### Example 6: Adjust Only Sprinting Drain

If you only want to make sprinting less punishing:

```yaml
metabolism:
  hunger:
    drain_rates:
      sprinting: 0.4     # Changed from 0.8 (50% reduction)
      
  thirst:
    drain_rates:
      sprinting: 0.5     # Changed from 1.0 (50% reduction)
      
  energy:
    drain_rates:
      sprinting: 0.75    # Changed from 1.5 (50% reduction)
```

---

## Per-World Overrides

You can override metabolism settings for specific worlds. This is useful for:
- Disabling metabolism in creative/minigame worlds
- Creating challenge worlds with harsher survival
- Setting different drain rates for different dimensions

### Structure:

```yaml
world_overrides:
  world_name_here:  # Must match your world folder name EXACTLY
    metabolism:
      enabled: true   # Set to false to disable in this world
      hunger:
        # World-specific hunger settings
      thirst:
        # World-specific thirst settings
      energy:
        # World-specific energy settings
```

### Example: Different Settings for Main World vs Nether

```yaml
world_overrides:
  world:  # Main world
    metabolism:
      enabled: true
      # Uses default values
      
  world_nether:  # Nether dimension
    metabolism:
      enabled: true
      thirst:
        drain_rates:
          hot_biome_multiplier: 2.5  # Extra thirst drain in nether
      energy:
        drain_rates:
          combat: 3.0  # More dangerous combat in nether
```

---

## Hot Reload

After editing `metabolism.yml`, you **DO NOT need to restart the server**.

Simply run:
```
/ll reload metabolism
```

Changes apply immediately to all players. Players will see updated drain rates and effects instantly.

---

## Profession Interactions

Some professions modify metabolism mechanics:

### Tier 2 Abilities (Level 45):
- **Combat Profession:** +15 max hunger capacity
- **Mining Profession:** +10 max thirst capacity
- **Logging Profession:** +10 max energy capacity
- **Gathering Profession:** +4 hunger and +4 thirst on every food pickup

### Tier 3 Abilities (Level 100):
- **Survivalist (Gathering):** -15% metabolism depletion across ALL stats
  - This is a permanent passive effect
  - Effectively reduces all drain rates by 15%
  - Stacks with your config settings

**Example:** If sprinting normally drains 0.8 hunger/minute, a max-level Gatherer with Survivalist will only lose 0.68 hunger/minute (0.8 × 0.85).

---

## Admin Commands

- `/ll stats` - Toggle metabolism HUD display
- `/ll reload metabolism` - Reload metabolism configuration without restarting server
- `/ll debug metabolism <player>` - View detailed metabolism stats for a player (debug tool)

---

## Troubleshooting

### "My metabolism isn't draining at all"

**Check:**
1. Is `metabolism.enabled` set to `true` in config?
2. Are you in a world with world overrides that disable it?
3. Have you reloaded the config after changes? (`/ll reload metabolism`)

### "Drain rates feel too fast/slow"

**Solution:**
1. Check your `drain_rates` values in config
2. Remember: values are **per minute**, not per second
3. Test with different activities (idle vs sprinting vs combat)
4. Adjust in small increments (e.g., change 0.8 to 0.6, not 0.8 to 0.1)

### "Hot reload isn't working"

**Solution:**
1. Make sure you're using `/ll reload metabolism` (not `/reload`)
2. Check server console for error messages
3. Verify the YAML syntax is correct (use a YAML validator)
4. If still not working, restart the server

### "World overrides aren't applying"

**Check:**
1. World name matches **exactly** (case-sensitive)
2. YAML indentation is correct (use spaces, not tabs)
3. You've reloaded the config after adding overrides
4. World name matches the folder name in your server directory

---

## Quick Reference: Default Drain Rates

| Activity  | Hunger/min | Thirst/min | Energy/min |
|-----------|------------|------------|------------|
| Idle      | 0.1        | 0.15       | 0.05       |
| Walking   | 0.3        | 0.4        | 0.2        |
| Sprinting | 0.8        | 1.0        | 1.5        |
| Combat    | 1.2        | 0.5        | 2.0        |
| Mining    | 0.4        | -          | 1.8        |
| Logging   | -          | -          | 1.6        |

**Note:** Thirst has additional multiplier (1.5x) in hot biomes by default.

---

## Summary

**To customize metabolism depletion rates:**

1. Open `plugins/LivingLandsReloaded/config/metabolism.yml`
2. Adjust the `drain_rates` values under `hunger`, `thirst`, or `energy`
3. Save the file
4. Run `/ll reload metabolism` in-game
5. Test and repeat as needed

**Higher values = faster drain = harder survival**  
**Lower values = slower drain = easier survival**

You can also adjust penalties, buff/debuff thresholds, and create per-world overrides for different gameplay experiences.
