# Hytale Asset System Reference

This document describes the asset system structure extracted from the Assets.zip archive.

## Table of Contents

1. [Asset System Overview](#asset-system-overview)
2. [Server Assets](#server-assets)
3. [Common Assets](#common-assets)
4. [Block Type Lists](#block-type-lists)
5. [Asset Organization](#asset-organization)

---

## Asset System Overview

The Hytale asset system is organized into two main categories:

1. **Common Assets** (`Common/`) - Client-side assets (textures, models, sounds)
2. **Server Assets** (`Server/`) - Server-side configuration files (JSON)

### Asset Loading

Assets are loaded automatically by the server when:
- The server starts up
- An asset pack is registered via `AssetPackRegisterEvent`
- Hot-reloading is triggered (in development mode)

---

## Server Assets

### HytaleGenerator

**Location:** `Server/HytaleGenerator/`

The HytaleGenerator directory contains the procedural world generation configuration files.

#### Directory Structure

```
Server/HytaleGenerator/
├── Assignments/          # Prop placement rules
│   ├── Boreal1/
│   ├── Desert1/
│   ├── Experimental/
│   ├── Generative/
│   ├── Plains1/
│   ├── Taiga1/
│   └── Volcanic1/
├── Biomes/               # Biome definitions
│   ├── Boreal1/
│   ├── Default_Flat/
│   ├── Default_Void/
│   ├── Desert1/
│   ├── Experimental/
│   ├── Generative/
│   ├── Ocean1/
│   ├── Plains1/
│   ├── Taiga1/
│   ├── Volcanic1/
│   ├── Basic.json
│   ├── FloatingIslands_Primary.json
│   ├── FloatingIslands_Satellite.json
│   └── Void.json
├── BlockMasks/           # Block filtering masks
├── Density/              # Density field definitions
├── Graphs/               # Graph-based structures
├── Settings/             # Generator settings
└── WorldStructures/      # World structure definitions
```

#### Settings (Settings.json)

**Location:** `Server/HytaleGenerator/Settings/Settings.json`

```json
{
  "StatsCheckpoints": [1, 100, 500, 1000],
  "CustomConcurrency": -1,
  "BufferCapacityFactor": 0.1,
  "TargetViewDistance": 512,
  "TargetPlayerCount": 3
}
```

**Fields:**
- `StatsCheckpoints` - Performance measurement checkpoints
- `CustomConcurrency` - Custom thread concurrency (-1 = auto)
- `BufferCapacityFactor` - Buffer sizing factor
- `TargetViewDistance` - Target view distance in blocks
- `TargetPlayerCount` - Target concurrent player count

---

### BlockTypeList

**Location:** `Server/BlockTypeList/`

JSON files defining categorized block type groups for use in generation and tools.

#### Available Lists

| File | Description |
|------|-------------|
| `AllScatter.json` | All blocks suitable for scattering |
| `Empty.json` | Empty block reference |
| `Gravel.json` | Gravel variants |
| `Ores.json` | Ore blocks |
| `PlantsAndTrees.json` | Plant and tree blocks |
| `PlantScatter.json` | Plants for scattering |
| `Rock.json` | Rock and stone variants |
| `Snow.json` | Snow blocks |
| `Soils.json` | Soil and dirt variants |
| `TreeLeaves.json` | Tree leaf blocks |

#### Format

```json
{
  "Blocks": [
    "BlockType_String_ID_1",
    "BlockType_String_ID_2",
    "..."
  ]
}
```

**Example - Soils.json:**
```json
{
  "Blocks": [
    "Soil_Dirt",
    "Soil_Dirt_Burnt",
    "Soil_Dirt_Cold",
    "Soil_Dirt_Dry",
    "Soil_Dirt_Poisoned",
    "Soil_Grass",
    "Soil_Grass_Burnt",
    "Soil_Grass_Cold",
    "Soil_Grass_Deep",
    "Soil_Grass_Dry",
    "Soil_Grass_Full",
    "Soil_Grass_Sunny",
    "Soil_Grass_Wet"
  ]
}
```

---

### Audio Configuration

**Location:** `Server/Audio/`

#### AmbienceFX

**Location:** `Server/Audio/AmbienceFX/`

Ambient sound effect configurations.

```
Server/Audio/AmbienceFX/
├── AmbFX_Placeholder.json
├── AmbFX_Void.json
└── Ambience/
    ├── Global/
    │   ├── Cave/
    │   ├── Dungeon/
    │   ├── Mage_Tower/
    │   ├── Mineshaft/
    │   ├── Underwater/
    │   └── Weather/
    └── Unique/
        └── Dread_Wade/
```

**Example Structure:**
```json
{
  "Type": "AmbienceFX",
  "Id": "AmbFX_Caves",
  "AmbientBed": {
    "Sounds": [...],
    "Volume": 0.5,
    "Loop": true
  },
  "Conditions": {
    "Biomes": [...],
    "HeightRange": [0, 64],
    "TimeOfDay": "Any"
  }
}
```

---

## Common Assets

### Block Textures

**Location:** `Common/BlockTextures/`

Block texture assets organized by material type.

```
Common/BlockTextures/
├── Bone_Side.png
├── Bone_Top.png
├── Calcite.png
├── Calcite_Brick_*.png
├── Clay_*.png
├── Cloth_*.png
├── Cracks/              # Block breaking overlay textures
│   ├── T_Crack_Dirt_*.png
│   ├── T_Crack_Generic_*.png
│   ├── T_Crack_Stone_*.png
│   └── T_Crack_Wood_*.png
├── Deco_*.png
├── EditorBlock*.png
└── Fluid_*.png
```

---

### Cosmetics

**Location:** `Common/Cosmetics/CharacterCreator/`

Character customization assets and configurations.

```
Common/Cosmetics/CharacterCreator/
├── BodyCharacteristics.json
├── Capes.json
├── EarAccessory.json
├── Ears.json
├── Emotes.json
├── EyeColors.json
├── Eyebrows.json
├── Eyes.json
├── FaceAccessory.json
├── Faces.json
├── FacialHair.json
├── GenericColors.json
├── Gloves.json
├── GradientSets.json
├── HairColors.json
├── HaircutFallbacks.json
├── Haircuts.json
├── HeadAccessory.json
├── Mouths.json
├── Overpants.json
├── Overtops.json
├── Pants.json
├── Shoes.json
├── SkinFeatures.json
├── Tags.json
└── Undertops.json
```

---

## Asset Loading and Registration

### Asset Pack Structure

A complete asset pack for a mod should be organized as:

```
MyMod.jar/
├── manifest.json
└── Server/
    ├── HytaleGenerator/
    │   ├── Biomes/
    │   ├── Assignments/
    │   ├── Density/
    │   └── Settings/
    ├── BlockTypeList/
    └── Audio/
        └── AmbienceFX/
```

### Asset Registration Events

**Package:** `com.hypixel.hytale.server.core.asset`

```java
// Verified signatures from JAR
public class AssetPackRegisterEvent implements IEvent<Void> {
    public AssetPackRegisterEvent(AssetPack);
    public AssetPack getAssetPack();
}

public class AssetPackUnregisterEvent implements IEvent<Void> {
    public AssetPackUnregisterEvent(AssetPack);
    public AssetPack getAssetPack();
}
```

### Loading Assets Programmatically

```kotlin
class MyPlugin : JavaPlugin(init) {
    
    override fun setup() {
        // LoadAssetEvent exists, but it does not expose per-asset IDs.
        eventRegistry.register(LoadAssetEvent::class.java) { event ->
            logger.atInfo().log("Asset load bootStart=%d reasons=%s", event.bootStart, event.reasons)
        }
        
        // Access plugin asset registry
        val assetRegistry = assetRegistry
    }
}
```

---

## Asset Types Reference

### Block-Related Asset Types

| Asset Type | Description | File Extension |
|------------|-------------|----------------|
| BlockType | Block definition | .json |
| BlockSet | Collection of blocks | .json |
| BlockSoundSet | Block sound mappings | .json |
| BlockParticleSet | Block particle effects | .json |
| BlockBreakingDecal | Breaking animation | .json |
| BlockHitbox | Collision bounds | .json |

### World Generation Asset Types

| Asset Type | Description | Location |
|------------|-------------|----------|
| Biome | Biome configuration | HytaleGenerator/Biomes/ |
| Density | Density field | HytaleGenerator/Density/ |
| Assignment | Prop placement rules | HytaleGenerator/Assignments/ |
| BlockMask | Block filtering | HytaleGenerator/BlockMasks/ |
| Graph | Structure graphs | HytaleGenerator/Graphs/ |

### Audio Asset Types

| Asset Type | Description | Location |
|------------|-------------|----------|
| SoundEvent | Sound definitions | Server/Audio/SoundEvents/ |
| AmbienceFX | Ambient effects | Server/Audio/AmbienceFX/ |
| AudioCategory | Audio mixing | Server/Audio/Categories/ |

### Gameplay Asset Types

| Asset Type | Description |
|------------|-------------|
| CraftingBench | Crafting stations |
| ProcessingBench | Processing stations |
| Droplist | Loot tables |
| Particle | Particle effects |
| Zone | World zones |
| Weather | Weather configs |
| Attitude | NPC attitudes |

---

## Asset Schema

### Common Asset Properties

All assets share common properties:

```json
{
  "$Title": "[ROOT] Asset Name",
  "$Position": {
    "$x": -4228,
    "$y": -4226
  },
  "$WorkspaceID": "HytaleGenerator - Type",
  "$Groups": [...]
}
```

**Editor Metadata Fields:**
- `$Title` - Display title for editor
- `$Position` - Editor node position (x, y coordinates)
- `$WorkspaceID` - Workspace identifier
- `$Groups` - Visual grouping for editor

### Asset Identifiers

Assets are referenced by string identifiers:

```
namespace:asset_name
```

**Examples:**
- `hytale:Basic` - Basic biome
- `hytale:Soil_Dirt` - Dirt block
- `mpc:sky_islands` - Custom mod biome

---

## Asset Validation

### JSON Schema Validation

Assets are validated against schemas on load.

```java
// Verified signature from JAR (note: validates String values/paths, not JsonObject directly)
public class CommonAssetValidator implements Validator<String> {
    public void accept(String, ValidationResults);
    public void updateSchema(SchemaContext, Schema);
}
```

### Common Validation Rules

1. **Required Fields:** All required fields must be present
2. **Type Checking:** Field types must match schema
3. **Reference Validation:** Referenced assets must exist
4. **Range Validation:** Numeric values must be in valid ranges
5. **Circular Dependency:** No circular references allowed

---

## Asset Hot-Reloading

### Development Mode

In development mode, assets can be hot-reloaded. The server exposes a monitor event carrying created/modified/removed path lists.

```kotlin
eventRegistry.register(CommonAssetMonitorEvent::class.java) { event ->
    val pack = event.assetPack
    val createdOrModified = event.createdOrModifiedFilesToLoad
    val removed = event.removedFilesToUnload
    logger.atInfo().log(
        "Asset monitor pack=%s +%d -%d",
        pack,
        createdOrModified.size,
        removed.size
    )
}
```

### Asset Monitor Events

**Package:** `com.hypixel.hytale.server.core.asset.common.events`

```java
// Verified signature from JAR
public class CommonAssetMonitorEvent extends AssetMonitorEvent<Void> {
    public CommonAssetMonitorEvent(String, List<Path>, List<Path>, List<Path>, List<Path>);
}

// Base event carries the data
public abstract class AssetMonitorEvent<T> implements IEvent<T> {
    public String getAssetPack();
    public List<Path> getCreatedOrModifiedFilesToLoad();
    public List<Path> getRemovedFilesToUnload();
    public List<Path> getCreatedOrModifiedDirectories();
    public List<Path> getRemovedFilesAndDirectories();
}
```

---

## Best Practices

1. **Organization:** Group related assets in subdirectories
2. **Naming:** Use consistent naming conventions (PascalCase for IDs)
3. **References:** Always use fully qualified IDs (namespace:name)
4. **Validation:** Test assets in development mode before deployment
5. **Versioning:** Include asset pack version in manifest
6. **Optimization:** Minimize asset file sizes for client distribution

---

*Generated from Assets.zip analysis*
