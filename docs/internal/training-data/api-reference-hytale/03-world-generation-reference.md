# Hytale World Generation Reference

This document provides documentation for the Hytale World Generation system based on the HytaleGenerator JSON configurations and verified server-side APIs from `libs/Server/HytaleServer.jar`.

**Last Verified:** January 31, 2026
**Source JAR:** `libs/Server/HytaleServer.jar`

## Table of Contents

1. [World Generation Overview](#world-generation-overview)
2. [Biome Definitions](#biome-definitions)
3. [Density Functions](#density-functions)
4. [World Structures](#world-structures)
5. [Assignments (Props)](#assignments-props)
6. [Material Providers](#material-providers)
7. [Noise Types](#noise-types)
8. [Node Types](#node-types)

9. [Server WorldGen APIs (Verified)](#server-worldgen-apis-verified)

---

## World Generation Overview

Hytale uses a node-based procedural generation system. Worlds are generated through:

1. **World Structure** - Defines biome layout and global parameters
2. **Biomes** - Define terrain shape, materials, and props
3. **Density Functions** - Generate 3D density fields for terrain
4. **Assignments** - Place props (trees, rocks, etc.) in the world

### Generation Pipeline

```
World Structure → Biome Selection → Density Generation → Material Application → Prop Placement
```

---

## Server WorldGen APIs (Verified)

This section documents the server-side Java APIs that drive chunk generation.

### IWorldGen

**Package:** `com.hypixel.hytale.server.core.universe.world.worldgen.IWorldGen`

```java
// Verified signature from JAR
public interface IWorldGen {
    WorldGenTimingsCollector getTimings();

    CompletableFuture<GeneratedChunk> generate(
        int chunkX,
        long seed,
        int chunkZ,
        int ???,
        LongPredicate shouldCancel
    );

    Transform[] getSpawnPoints(int count);
    default ISpawnProvider getDefaultSpawnProvider(int count);
    default void shutdown();
}
```

**Note:** `generate(...)` has a 4th `int` parameter in the JAR in addition to `(chunkX, seed, chunkZ, ...)`. The exact meaning is not documented here yet; treat it as required.

### GeneratedChunk

**Package:** `com.hypixel.hytale.server.core.universe.world.worldgen.GeneratedChunk`

```java
// Verified signature from JAR (partial)
public class GeneratedChunk {
    public GeneratedChunk();

    public GeneratedBlockChunk getBlockChunk();
    public GeneratedBlockStateChunk getBlockStateChunk();
    public GeneratedEntityChunk getEntityChunk();
    public Holder<ChunkStore>[] getSections();

    public Holder<ChunkStore> toWorldChunk(World);
    public Holder<ChunkStore> toHolder(World);

    public static Holder<ChunkStore>[] makeSections();
}
```

### IWorldGenProvider

**Package:** `com.hypixel.hytale.server.core.universe.world.worldgen.provider.IWorldGenProvider`

```java
// Verified signature from JAR
public interface IWorldGenProvider {
    public static final BuilderCodecMapCodec<IWorldGenProvider> CODEC;

    IWorldGen getGenerator() throws WorldGenLoadException;
}
```

### Built-in Providers

**Package:** `com.hypixel.hytale.server.core.universe.world.worldgen.provider`

- `FlatWorldGenProvider`
- `VoidWorldGenProvider`
- `DummyWorldGenProvider`

### WorldGenTimingsCollector

**Package:** `com.hypixel.hytale.server.core.universe.world.worldgen.WorldGenTimingsCollector`

Provides server timings for various worldgen phases.

```java
// Verified signature from JAR (partial)
public class WorldGenTimingsCollector {
    public double reportChunk(long nanos);
    public double reportZoneBiomeResult(long nanos);
    public double reportPrepare(long nanos);
    public double reportBlocksGeneration(long nanos);
    public double reportCaveGeneration(long nanos);
    public double reportPrefabGeneration(long nanos);

    public long getChunkCounter();
    public double getChunkTime();
    public int getQueueLength();
    public int getGeneratingCount();
}
```

### WorldGen Commands

The server ships built-in worldgen commands, useful for understanding reload/benchmark behaviors.

**Package:** `com.hypixel.hytale.server.core.command.commands.world.worldgen`

- `WorldGenCommand` extends `AbstractCommandCollection`
- `WorldGenReloadCommand` extends `AbstractAsyncWorldCommand` and exposes a `--clear` flag

---

---

## Biome Definitions

### Biome Structure

**Location:** `HytaleGenerator/Biomes/*.json`

```json
{
  "$Title": "[ROOT] Biome",
  "$Position": {"$x": -4228, "$y": -4226},
  "Name": "BiomeName",
  
  // Terrain generation
  "Terrain": {
    "Type": "DAOTerrain",
    "Density": { /* Density nodes */ }
  },
  
  // Material assignment
  "MaterialProvider": {
    "Type": "Solidity",
    "Solid": { /* Material for solid blocks */ },
    "Empty": { /* Material for empty blocks */ }
  },
  
  // Prop placement
  "Props": [ /* Assignment nodes */ ],
  
  // Environment settings
  "EnvironmentProvider": {
    "Type": "Constant",
    "Environment": "Env_Zone1_Plains"
  },
  
  // Visual tint
  "TintProvider": {
    "Type": "Constant",
    "Color": "#5b9e28"
  }
}
```

### Basic Biome Example

```json
{
  "$Title": "[ROOT] Biome",
  "$Position": {"$x": -4228, "$y": -4226},
  "Name": "Basic",
  
  "Terrain": {
    "$Position": {"$x": -3366, "$y": -4882},
    "Type": "DAOTerrain",
    "Density": {
      "$Position": {"$x": -3160, "$y": -4896},
      "Type": "Sum",
      "Skip": false,
      "Inputs": [
        {
          "$Position": {"$x": -2522, "$y": -5102},
          "Type": "SimplexNoise2D",
          "Skip": false,
          "Lacunarity": 10,
          "Persistence": 0.05,
          "Octaves": 1,
          "Scale": 150,
          "Seed": "A"
        },
        {
          "$Position": {"$x": -2526, "$y": -4727},
          "Type": "CurveMapper",
          "Skip": false,
          "Curve": {
            "Type": "Manual",
            "Points": [
              {"In": 0, "Out": 1},
              {"In": 50, "Out": -1}
            ]
          },
          "Inputs": [{
            "Type": "BaseHeight",
            "Skip": false,
            "BaseHeightName": "Base",
            "Distance": true
          }]
        }
      ]
    }
  },
  
  "MaterialProvider": {
    "Type": "Solidity",
    "Solid": {
      "Type": "Queue",
      "Queue": [{
        "Type": "Constant",
        "Material": {"Solid": "Rock_Stone"}
      }]
    },
    "Empty": {
      "Type": "Queue",
      "Queue": [{
        "$Comment": "REQUIRED",
        "Type": "Constant",
        "Material": {"Solid": "Empty"}
      }]
    }
  },
  
  "EnvironmentProvider": {
    "Type": "Constant",
    "Environment": "Env_Zone1_Plains"
  },
  
  "TintProvider": {
    "Type": "Constant",
    "Color": "#5b9e28"
  }
}
```

### Floating Islands Biome

**Location:** `HytaleGenerator/Biomes/FloatingIslands_Primary.json`

Complex biome using:
- Multiple noise layers
- Curve mapping for island shapes
- Blending between density fields
- Exported noise for reuse

---

## Density Functions

### Density Node Types

Density functions generate 3D scalar fields used for terrain generation.

#### SimplexNoise2D

2D Simplex noise (heightmap-style generation).

```json
{
  "Type": "SimplexNoise2D",
  "Skip": false,
  "Lacunarity": 2,        // Frequency multiplier per octave
  "Persistence": 0.5,     // Amplitude multiplier per octave
  "Octaves": 1,           // Number of noise layers
  "Scale": 150,           // Spatial scale (larger = smoother)
  "Seed": "A"             // Seed string (deterministic)
}
```

#### SimplexNoise3D

3D Simplex noise (true 3D density).

```json
{
  "Type": "SimplexNoise3D",
  "Skip": false,
  "Lacunarity": 2,
  "Persistence": 0.5,
  "Octaves": 2,
  "Scale": 50,
  "Seed": "B"
}
```

#### CellNoise (Voronoi)

Cellular/Voronoi noise for distinct cell patterns.

```json
{
  "Type": "PositionsCellNoise",
  "Skip": false,
  "MaxDistance": 460,
  "Positions": {
    "Type": "Mesh2D",
    "ExportAs": "World-Biome-Cells",
    "Skip": false,
    "PointsY": 0,
    "PointGenerator": {
      "Type": "Mesh",
      "Jitter": 0.2,          // Cell position randomization (0-1)
      "ScaleX": 350,          // Cell spacing X
      "ScaleY": 350,          // Cell spacing Y
      "ScaleZ": 350,          // Cell spacing Z
      "Seed": "World-Biome-Cells"
    }
  },
  "ReturnType": {
    "Type": "CellValue",
    "Density": { /* Noise for cell values */ },
    "DefaultValue": 0
  },
  "DistanceFunction": {
    "Type": "Euclidean"  // or "Manhattan", "Chebyshev"
  }
}
```

#### Mathematical Operations

**Sum:** Add multiple inputs
```json
{
  "Type": "Sum",
  "Skip": false,
  "Inputs": [ /* density nodes */ ]
}
```

**Min/Max:** Minimum or maximum of inputs
```json
{
  "Type": "Min",  // or "Max"
  "Skip": false,
  "Inputs": [ /* density nodes */ ]
}
```

**SmoothMin:** Smooth minimum with range
```json
{
  "Type": "SmoothMin",
  "Skip": false,
  "Range": 0.5,
  "Inputs": [ /* density nodes */ ]
}
```

**Constant:** Fixed value
```json
{
  "Type": "Constant",
  "Skip": false,
  "Value": -0.3
}
```

**Inverter:** Invert sign
```json
{
  "Type": "Inverter",
  "Skip": false,
  "Inputs": [ /* density node */ ]
}
```

**Abs:** Absolute value
```json
{
  "Type": "Abs",
  "Skip": false,
  "Inputs": [ /* density node */ ]
}
```

**Pow:** Power/exponent
```json
{
  "Type": "Pow",
  "Skip": false,
  "Exponent": 3,
  "Inputs": [ /* density node */ ]
}
```

#### Curve Mapping

**CurveMapper:** Remap values through a curve
```json
{
  "Type": "CurveMapper",
  "Skip": false,
  "Curve": {
    "Type": "Manual",
    "Points": [
      {"In": 0, "Out": 1},
      {"In": 50, "Out": -1}
    ]
  },
  "Inputs": [ /* density node */ ]
}
```

**Normalizer:** Scale values from one range to another
```json
{
  "Type": "Normalizer",
  "Skip": false,
  "FromMin": -1,
  "FromMax": 1,
  "ToMin": -0.05,
  "ToMax": 0.05,
  "Inputs": [ /* density node */ ]
}
```

#### Utility Nodes

**Cache:** Cache results for performance
```json
{
  "Type": "Cache",
  "Skip": false,
  "Capacity": 3,
  "Inputs": [ /* density node */ ]
}
```

**YOverride:** Force Y coordinate
```json
{
  "Type": "YOverride",
  "Skip": false,
  "Value": 0,
  "Inputs": [ /* density node */ ]
}
```

**Scale:** Scale coordinates
```json
{
  "Type": "Scale",
  "Skip": false,
  "ScaleX": 1,
  "ScaleY": 1,
  "ScaleZ": 1,
  "Inputs": [ /* density node */ ]
}
```

**Mix:** Blend between inputs
```json
{
  "Type": "Mix",
  "Skip": false,
  "Inputs": [ /* density nodes */ ]
}
```

**Clamp:** Clamp to range
```json
{
  "Type": "Clamp",
  "Skip": false,
  "WallA": -1,
  "WallB": 1,
  "Inputs": [ /* density node */ ]
}
```

**FastGradientWarp:** Domain warping
```json
{
  "Type": "FastGradientWarp",
  "Skip": false,
  "WarpScale": 100,
  "WarpPersistence": 0.2,
  "WarpLacunarity": 2,
  "WarpOctaves": 2,
  "WarpFactor": 100,
  "Seed": "A",
  "Inputs": [ /* density node */ ]
}
```

#### Import/Export

**Exported:** Export density for reuse
```json
{
  "Type": "Exported",
  "ExportAs": "Biome-Map",
  "SingleInstance": true,
  "Skip": false,
  "Inputs": [ /* density node */ ]
}
```

**Imported:** Import previously exported density
```json
{
  "Type": "Imported",
  "Skip": false,
  "Name": "Biome-Map"
}
```

#### Base Heights

**BaseHeight:** Reference named height field
```json
{
  "Type": "BaseHeight",
  "Skip": false,
  "BaseHeightName": "Base",  // Name defined in WorldStructure
  "Distance": true           // Calculate distance from height
}
```

**Distance:** Distance-based falloff
```json
{
  "Type": "Distance",
  "Skip": false,
  "Curve": {
    "Type": "Manual",
    "Points": [
      {"In": 0, "Out": -1},
      {"In": 1000, "Out": 0}
    ]
  }
}
```

### Complete Density Example

From `HytaleGenerator/Density/Map_Default.json`:

```json
{
  "Type": "Exported",
  "ExportAs": "Biome-Map",
  "SingleInstance": true,
  "Inputs": [{
    "Type": "YOverride",
    "Value": 0,
    "Inputs": [{
      "Type": "Cache",
      "Capacity": 1,
      "Inputs": [{
        "Type": "Mix",
        "Inputs": [
          // Biome tiles using CellNoise
          {
            "Type": "FastGradientWarp",
            "WarpScale": 100,
            "WarpPersistence": 0.2,
            "WarpLacunarity": 2,
            "WarpOctaves": 2,
            "WarpFactor": 100,
            "Seed": "A",
            "Inputs": [{
              "Type": "PositionsCellNoise",
              "MaxDistance": 460,
              "Positions": {
                "Type": "Mesh2D",
                "ExportAs": "World-Biome-Cells",
                "PointGenerator": {
                  "Type": "Mesh",
                  "Jitter": 0.2,
                  "ScaleX": 350,
                  "ScaleY": 350,
                  "ScaleZ": 350,
                  "Seed": "World-Biome-Cells"
                }
              },
              "ReturnType": {
                "Type": "CellValue",
                "Density": {
                  "Type": "SimplexNoise2D",
                  "Lacunarity": 2,
                  "Persistence": 0.5,
                  "Octaves": 1,
                  "Scale": 50,
                  "Seed": "World-Biome-Cells_Random"
                }
              },
              "DistanceFunction": {"Type": "Euclidean"}
            }]
          },
          // River density
          {"Type": "Constant", "Value": -0.3},
          // Continent mask
          {"Type": "Clamp", "WallA": 0, "WallB": 1, "Inputs": [{
            "Type": "Normalizer",
            "FromMin": 0, "FromMax": 0, "ToMin": 0, "ToMax": 1,
            "Inputs": [{"Type": "Imported", "Name": "World-Continent-Map"}]
          }]}
        ]
      }]
    }]
  }]
}
```

---

## World Structures

### World Structure Format

**Location:** `HytaleGenerator/WorldStructures/*.json`

Defines how biomes are arranged in the world.

```json
{
  "Type": "NoiseRange",           // Selection type
  "Biomes": [                     // Biome definitions
    {
      "Biome": "FloatingIslands_Primary",
      "Min": -0.25,
      "Max": 0.25
    },
    {
      "Biome": "FloatingIslands_Satellite",
      "Min": 0.25,
      "Max": 0.45
    }
  ],
  "DefaultBiome": "Basic",        // Fallback biome
  "DefaultTransitionDistance": 32,// Blending distance
  "MaxBiomeEdgeDistance": 32,     // Max edge transition
  "Density": {                    // Biome selection density
    "Type": "Imported",
    "Name": "Biome-Map"
  },
  "ContentFields": [              // Named height references
    {"Type": "BaseHeight", "Name": "Base", "Y": 100},
    {"Type": "BaseHeight", "Name": "Water", "Y": 100},
    {"Type": "BaseHeight", "Name": "Bedrock", "Y": 0}
  ]
}
```

### World Structure Types

**NoiseRange:** Select biome based on noise value ranges
```json
{
  "Type": "NoiseRange",
  "Biomes": [
    {"Biome": "BiomeName", "Min": -1.0, "Max": -0.5},
    {"Biome": "OtherBiome", "Min": -0.5, "Max": 0.5}
  ]
}
```

---

## Assignments (Props)

### Assignment Structure

**Location:** `HytaleGenerator/Assignments/*/*.json`

Define how props (trees, rocks, vegetation) are placed in biomes.

```json
{
  "$Title": "[ROOT] Weighted Assignments",
  "Type": "FieldFunction",
  "ExportAs": "AssignmentName",
  "FieldFunction": { /* Noise for selection */ },
  "Delimiters": [ /* Value ranges with assignments */ ]
}
```

### Assignment Types

**FieldFunction:** Noise-based assignment selection
```json
{
  "Type": "FieldFunction",
  "ExportAs": "Forest_Trees",
  "FieldFunction": {
    "Type": "SimplexNoise2D",
    "Lacunarity": 2,
    "Persistence": 0.5,
    "Octaves": 1,
    "Scale": 120,
    "Seed": "1235"
  },
  "Delimiters": [{
    "Min": -1,
    "Max": -0.8,
    "Assignments": { /* Assignment definition */ }
  }]
}
```

### Prop Types

**Constant:** Fixed prop
```json
{
  "Type": "Constant",
  "Prop": { /* prop definition */ }
}
```

**Weighted:** Random selection by weight
```json
{
  "Type": "Weighted",
  "SkipChance": 0.97,
  "Seed": "A",
  "WeightedAssignments": [
    {"Weight": 20, "Assignments": { /* prop */ }},
    {"Weight": 40, "Assignments": { /* prop */ }},
    {"Weight": 40, "Assignments": { /* prop */ }}
  ]
}
```

**Prefab:** Place a prefab file
```json
{
  "Type": "Prefab",
  "Skip": false,
  "WeightedPrefabPaths": [
    {"Path": "Trees/Fir/Stage_00", "Weight": 1},
    {"Path": "Trees/Fir/Stage_01", "Weight": 2}
  ],
  "LegacyPath": false,
  "LoadEntities": true,
  "Directionality": { /* placement rules */ },
  "Scanner": { /* placement scanning */ },
  "MoldingDirection": "NONE",
  "MoldingChildren": false
}
```

**Column:** Vertical stack of blocks
```json
{
  "Type": "Column",
  "Skip": false,
  "ColumnBlocks": [
    {"Y": 0, "Material": {"Solid": "Wood_Sticks"}}
  ],
  "Directionality": { /* direction rules */ },
  "Scanner": { /* placement scanning */ }
}
```

**Cluster:** Group of props
```json
{
  "Type": "Cluster",
  "Skip": false,
  "Range": 10,
  "Seed": "A",
  "DistanceCurve": {
    "Type": "Manual",
    "Points": [
      {"In": 2, "Out": 0.5},
      {"In": 5, "Out": 0}
    ]
  },
  "WeightedProps": [ /* weighted columns */ ],
  "Pattern": { /* surface matching */ },
  "Scanner": { /* placement scanning */ }
}
```

**Box:** Rectangular volume
```json
{
  "Type": "Box",
  "Skip": false,
  "BoxBlockType": "BoxBlockType",
  "Range": {"X": 2, "Y": 2, "Z": 2},
  "Material": {"Solid": "Soil_Dirt"},
  "Pattern": { /* surface matching */ },
  "Scanner": { /* placement scanning */ }
}
```

**Union:** Combine multiple props
```json
{
  "Type": "Union",
  "Skip": false,
  "Props": [ /* prop definitions */ ]
}
```

### Surface Matching (Patterns)

**Floor:** Match surface from below
```json
{
  "Type": "Floor",
  "Skip": false,
  "Origin": { /* block to find */ },
  "Floor": { /* block to place on */ }
}
```

**BlockSet:** Match any block in a set
```json
{
  "Type": "BlockSet",
  "Skip": false,
  "BlockSet": {
    "Inclusive": true,
    "Materials": [
      {"Solid": "Soil_Mud"},
      {"Solid": "Soil_Dirt_Cold"}
    ]
  }
}
```

**BlockType:** Match specific block
```json
{
  "Type": "BlockType",
  "Skip": false,
  "Material": {"Solid": "Empty"}
}
```

**And:** Combine patterns
```json
{
  "Type": "And",
  "ExportAs": "PatternName",
  "Skip": false,
  "Patterns": [ /* pattern definitions */ ]
}
```

**Offset:** Offset pattern check
```json
{
  "Type": "Offset",
  "Skip": false,
  "Pattern": { /* pattern to check */ },
  "Offset": {"X": 0, "Y": 5, "Z": 0}
}
```

### Placement Scanning

**ColumnLinear:** Scan vertically
```json
{
  "Type": "ColumnLinear",
  "Skip": false,
  "MaxY": 320,
  "MinY": 0,
  "RelativeToPosition": false,
  "BaseHeightName": "Water",
  "TopDownOrder": true,
  "ResultCap": 1
}
```

### Directionality

**Random:** Random rotation
```json
{
  "Type": "Random",
  "Seed": "A",
  "Pattern": { /* surface pattern */ }
}
```

**Static:** Fixed rotation
```json
{
  "Type": "Static",
  "Rotation": 0,
  "Pattern": { /* surface pattern */ }
}
```

---

## Material Providers

### Solidity Provider

**Type:** `Solidity`

Different materials for solid vs empty space:

```json
{
  "Type": "Solidity",
  "Solid": {
    "Type": "Queue",
    "Queue": [{
      "Type": "Constant",
      "Material": {"Solid": "Rock_Stone"}
    }]
  },
  "Empty": {
    "$Comment": "REQUIRED",
    "Type": "Queue",
    "Queue": [{
      "Type": "Constant",
      "Material": {"Solid": "Empty"}
    }]
  }
}
```

### Material Queue

Process materials in order (first match wins):

```json
{
  "Type": "Queue",
  "Queue": [
    {"Type": "Constant", "Material": {"Solid": "Soil_Grass"}},
    {"Type": "Constant", "Material": {"Solid": "Soil_Dirt"}},
    {"Type": "Constant", "Material": {"Solid": "Rock_Stone"}}
  ]
}
```

---

## Noise Types

### Summary Table

| Type | Dimensions | Use Case |
|------|------------|----------|
| `SimplexNoise2D` | 2D | Heightmaps, biome maps |
| `SimplexNoise3D` | 3D | Caves, overhangs, true 3D terrain |
| `CellNoise` | 2D/3D | Biome cells, distinct regions |
| `ValueNoise` | 2D/3D | Simple smooth noise |
| `PerlinNoise` | 2D/3D | Classic Perlin noise |

### Noise Parameters

**Lacunarity:** Frequency multiplier per octave
- Higher = more detail at each octave
- Typical: 2.0

**Persistence:** Amplitude multiplier per octave
- Higher = more influence from higher octaves
- Typical: 0.5

**Octaves:** Number of noise layers
- More = more detail but slower
- Typical: 1-4

**Scale:** Spatial scale
- Higher = smoother, lower frequency
- Typical: 50-1000

---

## Node Types

### Complete Node Type Reference

#### Generation Nodes

| Type | Description |
|------|-------------|
| `DAOTerrain` | Density-based terrain generator |
| `SimplexNoise2D` | 2D Simplex noise |
| `SimplexNoise3D` | 3D Simplex noise |
| `CellNoise` | Cellular/Voronoi noise |
| `PerlinNoise` | Perlin noise |
| `ValueNoise` | Value noise |

#### Mathematical Nodes

| Type | Description |
|------|-------------|
| `Sum` | Add inputs |
| `Min` | Minimum of inputs |
| `Max` | Maximum of inputs |
| `SmoothMin` | Smooth minimum |
| `Constant` | Fixed value |
| `Inverter` | Negate value |
| `Abs` | Absolute value |
| `Pow` | Power function |
| `Normalizer` | Range remapping |

#### Control Flow Nodes

| Type | Description |
|------|-------------|
| `CurveMapper` | Curve-based remapping |
| `Clamp` | Clamp to range |
| `Mix` | Blend inputs |
| `Cache` | Cache results |
| `YOverride` | Force Y coordinate |
| `Scale` | Scale coordinates |

#### Position/Point Nodes

| Type | Description |
|------|-------------|
| `Mesh` | Grid point generator |
| `Mesh2D` | 2D grid points |
| `MeshPointGenerator` | Configurable grid |
| `BaseHeight` | Reference named height |
| `Distance` | Distance-based falloff |

#### Import/Export Nodes

| Type | Description |
|------|-------------|
| `Exported` | Export for reuse |
| `Imported` | Import exported value |

#### Warping Nodes

| Type | Description |
|------|-------------|
| `FastGradientWarp` | Domain warping |
| `CoordinateRotator` | Rotate coordinates |
| `CoordinateRandomizer` | Randomize coordinates |

#### Prop/Assignment Nodes

| Type | Description |
|------|-------------|
| `Constant` | Fixed assignment |
| `Weighted` | Random weighted selection |
| `Prefab` | Prefab placement |
| `Column` | Vertical block stack |
| `Cluster` | Prop cluster |
| `Box` | Rectangular volume |
| `Union` | Combine props |

#### Pattern/Scanner Nodes

| Type | Description |
|------|-------------|
| `Floor` | Surface matching |
| `BlockSet` | Block set matching |
| `BlockType` | Single block matching |
| `And` | Combine patterns |
| `Offset` | Offset pattern |
| `ColumnLinear` | Vertical scan |
| `Random` | Random direction |
| `Static` | Fixed direction |

#### Material Nodes

| Type | Description |
|------|-------------|
| `Solidity` | Solid/empty materials |
| `Queue` | Ordered material list |
| `BlockSet` | Block set material |
| `BlockType` | Single block material |

---

## Best Practices

1. **Performance:**
   - Use `Cache` nodes for expensive calculations
   - Export reused densities with `Exported`/`Imported`
   - Limit octaves (1-3 typical, 4+ for detail areas)

2. **Biome Design:**
   - Keep density ranges consistent (-1 to 1 typical)
   - Use `BaseHeight` for sea level references
   - Design for smooth biome transitions

3. **Props:**
   - Use `SkipChance` to control density
   - Combine small props into clusters
   - Use appropriate scanners for each prop type

4. **Organization:**
   - Group related assignments by biome
   - Export commonly used patterns
   - Use descriptive node names

5. **Testing:**
   - Test with different seeds
   - Verify edge cases (min/max values)
   - Check performance with profiling

---

*Generated from HytaleGenerator JSON analysis*
