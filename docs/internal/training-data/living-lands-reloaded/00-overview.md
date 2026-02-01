# Living Lands Reloaded - Overview

Living Lands Reloaded is a Hytale server modification that makes survival feel personal through metabolism tracking and profession progression.

Every trip has a cost, every fight drains you, and every run home becomes a decision. 
As you play, your craft turns into identity: you grow into a fighter, miner, builder, logger, or gatherer, unlocking lasting perks along the way.

---

**IMPORTANT COMPATIBILITY WARNING:**

THIS MOD IS CURRENTLY NOT TESTED FOR COMPATIBILITY WITH OTHER MODS THAT HAVE A CUSTOM UI OR MANIPULATE PLAYER STATS! USE AT YOUR OWN RISK!

---

## Core Features

### 🧭 RPG Survival: Needs That Create Stakes

Three stats tracked and shown on the HUD:

- **Hunger** - food pressure (includes damage at low hunger)
- **Thirst** - travel pressure (stamina pressure at low thirst)
- **Energy** - effort pressure (speed debuffs at low energy, speed buff at high energy)

Drain is **activity-based** (sprinting/combat/travel), and foods restore different needs.

### 🏅 Professions: Your Character Sheet In Motion

Gain XP from normal play:

- **Combat** (kills), **Mining** (ores), **Logging** (logs), **Building** (placing blocks), **Gathering** (item pickups)

Abilities unlock at:

- **Level 15 (Tier 1):** +15% XP in that profession
- **Level 45 (Tier 2):** Combat +15 max hunger, Mining +10 max thirst, Logging +10 max energy, Gathering +4 hunger/thirst on food pickup, Building +15 max stamina (API pending)
- **Level 100 (Tier 3):** Adrenaline Rush (+10% speed on kill), Ore Sense (+10% ore drops), Timber! (+25% extra logs), Efficient Architect (12% no-consume on place), Survivalist (-15% metabolism depletion)

Death penalty: progressive XP loss on your **2 highest professions** (base **10%**, +**3%** per death, capped at **35%**), with decay and mercy.

---

## 🛠️ Built For Server Owners (Not Just Players)

- **Drop-in setup** - install the jar, boot the server, configs generate automatically
- **Hot reload** - `/ll reload [module]`
- **Per-world metabolism rules** - optional world overrides for metabolism config
- **Configurable** - XP rates, drain rates, penalties, announcements

---

## 📢 Server Announcements (NEW IN 1.3.0)

- **MOTD** + **welcome messages** (first-time vs returning)
- **Recurring announcements** (configurable intervals)
- **Admin broadcast** (`/ll broadcast <message>`)
- Placeholders: `{player_name}`, `{server_name}`, `{join_count}`
- Color codes: `&a`, `&6`, etc.

Note: per-world overrides are supported in config; world-specific routing is being finished.

---

## ⌨️ Commands

### Player Commands:
- `/ll stats` - toggle metabolism HUD panel
- `/ll buffs` - toggle buffs display
- `/ll debuffs` - toggle debuffs display
- `/ll professions` - toggle professions panel
- `/ll progress` - toggle compact professions progress panel

### Admin Commands:
- `/ll reload [module]` - reload configuration files (operator only)
- `/ll broadcast <message>` - broadcast message to all players (operator only)
- `/ll prof set/add/reset/show` - manage player professions (operator only)

---

## 🗺️ Roadmap

**Next Updates:**
- **Land Claims** (planned) - protect builds, trust friends, per-world management
- **Advanced mechanics** (planned) - high-stat buffs, seasonal variation, custom food effects, poison and dangerous consumables

---

## 📦 Installation

### For Players:
Join a server running Living Lands Reloaded. The HUD appears automatically.

### For Server Owners:

1. Download `livinglands-reloaded-1.3.0.jar`
2. Place it in your server `plugins/` folder
3. Start the server
4. Configs generate in `LivingLandsReloaded/config/`

---

## 💬 Community

- Bugs and issues: [GitHub](https://github.com/MoshPitCodes/living-lands-reloaded)
- Suggestions and discussion: [Discord](https://discord.gg/8jgMj9GPsq)
- If you like the mod: reviews help more than you think

Built by **MoshPitCodes**.

**Current Version:** 1.3.0
**License:** Apache 2.0
**Source Code:** Available on GitHub

---

☕ **Support Development**

Enjoying Living Lands Reloaded? Consider supporting development: [Ko-fi](https://ko-fi.com/moshpitplays)
