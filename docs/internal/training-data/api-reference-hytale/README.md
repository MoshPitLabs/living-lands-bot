# Hytale API Documentation

This directory contains comprehensive API documentation for the Hytale server modding ecosystem, extracted and **verified** from the official Hytale server libraries.

**Last Verified:** January 31, 2026
**Source JAR:** `libs/Server/HytaleServer.jar` (83.7 MB)

## Documentation Files

### 1. [Server API Reference](01-server-api-reference.md)
Complete Java/Kotlin API reference covering:
- Plugin system (PluginBase, JavaPlugin, PluginManifest)
- Command system (AbstractCommand, CommandContext, ArgTypes)
- Event system (EventRegistry, EventPriority, common events)
- Universe and World management
- Player system (PlayerRef, Player component)
- ECS Component system (Ref, Store, ComponentType)
- Noise and procedural generation
- Logging system (HytaleLogger)
- **Common pitfalls and best practices**

### 2. [Asset System Reference](02-asset-system-reference.md)
Asset organization and structure covering:
- Server-side assets (HytaleGenerator, BlockTypeList, Audio)
- Common assets (textures, cosmetics)
- Asset loading and registration
- Asset types reference
- Asset validation and hot-reloading

### 3. [World Generation Reference](03-world-generation-reference.md)
World Generation V2 system covering:
- Biome definitions and structure
- Density functions and noise types
- World structures and biome selection
- Assignments (prop placement)
- Material providers
- Complete node type reference

## Verification Status

All signatures in `01-server-api-reference.md` have been verified against the actual JAR using `javap`:

| Class | Verified | Notes |
|-------|----------|-------|
| PluginBase | YES | Constructor takes PluginInit |
| JavaPlugin | YES | Constructor takes JavaPluginInit |
| PluginState | YES | **CORRECTED** - enum values differ from old docs |
| AbstractCommand | YES | Third param is `requiresConfirmation`, not `requiresOp` |
| CommandContext | YES | All methods verified |
| ArgTypes | YES | Uses `SingleArgumentType<T>` not `ArgumentType<T>` |
| EventRegistry | YES | Full method list verified |
| EventPriority | YES | Enum values confirmed |
| Universe | YES | Note: `disconnectAllPLayers()` typo is real |
| World | YES | Thread safety patterns documented |
| PlayerRef | YES | `getUuid()` method, `worldUuid` field |
| Player | YES | Key managers documented |
| Ref | YES | **CORRECTED** - no `getId()`, `get()`, `remove()` methods |
| Store | YES | Component access patterns documented |
| Message | YES | Factory methods verified |
| HytaleLogger | YES | Flogger-based API |

## Source Files

The documentation was generated from analysis of:

- `libs/Server/HytaleServer.jar` - Server API classes (verified via `javap`)
- `libs/Server/Assets.zip` - Asset bundles
- World generation JSON configs (extracted from assets)

## Usage

These documents are reference materials for:

1. **Plugin Developers** - Implementing server mods
2. **World Generator Authors** - Creating custom terrain
3. **Asset Pack Creators** - Building custom content

## Key Concepts

### Plugin Architecture
- Plugins extend `JavaPlugin` (which extends `PluginBase`)
- Lifecycle: `setup()` -> `start()` -> `shutdown()`
- Commands extend `AbstractCommand`
- Events use `EventRegistry` with priority levels
- **Always use `HytaleLogger` from `getLogger()`, not standard Java loggers**

### World Thread Safety (CRITICAL)
All ECS operations must occur on the world thread:
```kotlin
world.execute {
    val store = world.entityStore
    val player = store.getComponent(ref, Player.getComponentType())
    // Safe to access player here
}
```

### PlayerRef vs Player vs Ref
- **PlayerRef** - Universe-level reference, survives world transfers
- **Player** - ECS component with gameplay data
- **Ref<EntityStore>** - ECS reference, may be invalidated on world transfer

### World Generation V2
- JSON-driven node-based system
- Density functions generate 3D terrain
- Biomes define materials and props
- Assignments place props via patterns and scanners

### Asset System
- Server assets in `Server/` directory
- PascalCase IDs with namespace prefix
- Hot-reloading in development mode
- Schema validation on load

## Common Corrections from Old Docs

1. **PluginState enum** - Values are `NONE, SETUP, START, ENABLED, SHUTDOWN, DISABLED` (NOT `PENDING, LOADING, DISABLING, ERROR`)
2. **AbstractCommand constructor** - Third param is `requiresConfirmation`, not `requiresOp`
3. **ArgTypes fields** - Most are `SingleArgumentType<T>`, not `ArgumentType<T>`
4. **Ref class** - Does NOT have `getId()`, `get()`, or `remove()` methods
5. **Universe.disconnectAllPLayers()** - This typo (capital L) is the actual method name
6. **PlayerRef.world** - Does NOT exist; use `getWorldUuid()` then `Universe.get().getWorld(uuid)`

## Verification Commands

To verify any class signature:
```bash
# From project root
javap -p -cp libs/Server/HytaleServer.jar <fully.qualified.ClassName>

# Example
javap -p -cp libs/Server/HytaleServer.jar com.hypixel.hytale.server.core.plugin.PluginBase
```

To list classes in a package:
```bash
jar tf libs/Server/HytaleServer.jar | grep "^com/hypixel/hytale/server/core/plugin/" | grep "\.class$"
```

## Related Resources

- Project AGENTS.md - Project-specific guidelines
- `libs/Server/HytaleServer.jar` - Source JAR for verification
- Community modding resources (verify against JAR before trusting)

---

*Generated and verified: January 31, 2026*
