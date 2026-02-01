# Hytale Server API Reference

This document provides detailed API documentation extracted and verified from the HytaleServer.jar file.

**Last Verified:** January 31, 2026
**JAR Version:** HytaleServer.jar (83.7 MB)

## Table of Contents

1. [Plugin System](#plugin-system)
2. [Command System](#command-system)
3. [World Generation](#world-generation)
4. [Event System](#event-system)
5. [Universe and World Management](#universe-and-world-management)
6. [Player System](#player-system)
7. [ECS Component System](#ecs-component-system)
8. [Noise and Procedural Generation](#noise-and-procedural-generation)
9. [Asset System](#asset-system)
10. [Logging System](#logging-system)
11. [Common Pitfalls and Best Practices](#common-pitfalls-and-best-practices)

---

## Plugin System

### PluginBase

**Package:** `com.hypixel.hytale.server.core.plugin.PluginBase`

The base class for all Hytale plugins. Provides access to core server systems and registries.

```java
// Verified signature from JAR
public abstract class PluginBase implements CommandOwner {
    // Core registries and systems
    public EventRegistry getEventRegistry()
    public CommandRegistry getCommandRegistry()
    public AssetRegistry getAssetRegistry()
    public BlockStateRegistry getBlockStateRegistry()
    public EntityRegistry getEntityRegistry()
    public TaskRegistry getTaskRegistry()
    public ComponentRegistryProxy<EntityStore> getEntityStoreRegistry()
    public ComponentRegistryProxy<ChunkStore> getChunkStoreRegistry()
    public ClientFeatureRegistry getClientFeatureRegistry()
    
    // Logging
    public HytaleLogger getLogger()
    
    // Configuration
    protected final <T> Config<T> withConfig(BuilderCodec<T> codec)
    protected final <T> Config<T> withConfig(String name, BuilderCodec<T> codec)
    
    // Plugin lifecycle (override these)
    protected void setup()       // Called during setup phase
    protected void start()       // Called when plugin starts
    protected void shutdown()    // Called when plugin shuts down
    
    // Internal lifecycle (do NOT override)
    protected void setup0()
    protected void start0()
    protected void shutdown0(boolean)
    
    // Pre-loading
    public CompletableFuture<Void> preLoad()
    
    // Plugin info
    public String getName()
    public PluginIdentifier getIdentifier()
    public PluginManifest getManifest()
    public Path getDataDirectory()
    public PluginState getState()
    public String getBasePermission()
    public boolean isDisabled()
    public boolean isEnabled()
    
    // Codec registries
    public <T, C extends Codec<? extends T>> CodecMapRegistry<T, C> getCodecRegistry(StringCodecMapCodec<T, C>)
    public <K, T extends JsonAsset<K>> CodecMapRegistry.Assets<T, ?> getCodecRegistry(AssetCodecMapCodec<K, T>)
    public <V> MapKeyMapRegistry<V> getCodecRegistry(MapKeyMapCodec<V>)
    
    // Must implement
    public abstract PluginType getType()
}
```

**Known Gotchas:**
- Always override `setup()`, `start()`, and `shutdown()` - NOT `setup0()`, `start0()`, `shutdown0()`
- `getLogger()` returns `HytaleLogger`, not a standard Java logger
- Constructor takes `PluginInit`, not no-args

### JavaPlugin

**Package:** `com.hypixel.hytale.server.core.plugin.JavaPlugin`

Base class for Java-based plugins extending PluginBase.

```java
// Verified signature from JAR
public abstract class JavaPlugin extends PluginBase {
    public JavaPlugin(JavaPluginInit init)
    public Path getFile()
    public PluginClassLoader getClassLoader()
    public final PluginType getType()  // Returns PluginType for Java plugins
}
```

### PluginManifest

**Package:** `com.hypixel.hytale.common.plugin.PluginManifest`

Defines the plugin configuration loaded from `manifest.json`.

```java
// Verified signature from JAR
public class PluginManifest {
    // Core fields (PascalCase in JSON)
    private String group
    private String name
    private Semver version
    private String description
    private List<AuthorInfo> authors
    private String website
    private String main                    // Main class path
    private SemverRange serverVersion      // Required server version
    
    // Dependencies
    private Map<PluginIdentifier, SemverRange> dependencies
    private Map<PluginIdentifier, SemverRange> optionalDependencies
    private Map<PluginIdentifier, SemverRange> loadBefore
    
    // Sub-plugins
    private List<PluginManifest> subPlugins
    private boolean disabledByDefault
    private boolean includesAssetPack
    
    // Getters
    public String getGroup()
    public String getName()
    public Semver getVersion()
    public String getDescription()
    public List<AuthorInfo> getAuthors()
    public String getWebsite()
    public String getMain()
    public SemverRange getServerVersion()
    public Map<PluginIdentifier, SemverRange> getDependencies()
    public Map<PluginIdentifier, SemverRange> getOptionalDependencies()
    public Map<PluginIdentifier, SemverRange> getLoadBefore()
    public List<PluginManifest> getSubPlugins()
    public boolean isDisabledByDefault()
    public boolean includesAssetPack()
    
    // Mutation
    public void injectDependency(PluginIdentifier, SemverRange)
    public void inherit(PluginManifest)
}
```

**Example manifest.json:**
```json
{
  "Group": "MPC",
  "Name": "LivingLandsReloaded",
  "Version": "2.7.0",
  "Description": "Survival mechanics mod",
  "Authors": [{"Name": "MoshPitCodes"}],
  "Main": "com.livinglands.LivingLandsPlugin"
}
```

### Plugin States

**Package:** `com.hypixel.hytale.server.core.plugin.PluginState`

```java
// Verified signature from JAR - CORRECTED from docs
public enum PluginState {
    NONE,       // Initial state
    SETUP,      // Plugin is in setup phase
    START,      // Plugin is starting
    ENABLED,    // Plugin is enabled and running
    SHUTDOWN,   // Plugin is shutting down
    DISABLED    // Plugin is disabled
}
```

**Note:** The old documentation listed `PENDING`, `LOADING`, `DISABLING`, `ERROR` - these do NOT exist in the actual enum.

---

## Command System

### AbstractCommand

**Package:** `com.hypixel.hytale.server.core.command.system.AbstractCommand`

Base class for all server commands.

```java
// Verified signature from JAR
public abstract class AbstractCommand {
    // Constructors
    protected AbstractCommand(String name, String description, boolean requiresConfirmation)
    protected AbstractCommand(String name, String description)
    protected AbstractCommand(String name)
    
    // Core execution method (MUST implement)
    protected abstract CompletableFuture<Void> execute(CommandContext ctx)
    
    // Permission management
    public void requirePermission(String permission)
    public List<String> getPermissionGroups()
    protected void setPermissionGroups(String... groups)
    protected void setPermissionGroup(GameMode mode)
    public boolean hasPermission(CommandSender sender)
    public Map<String, Set<String>> getPermissionGroupsRecursive()
    
    // Command structure
    public void addAliases(String... aliases)
    public void addSubCommand(AbstractCommand subCommand)
    public void addUsageVariant(AbstractCommand variant)
    
    // Argument helpers
    public <D> RequiredArg<D> withRequiredArg(String name, String description, ArgumentType<D> type)
    public <D> OptionalArg<D> withOptionalArg(String name, String description, ArgumentType<D> type)
    public <D> DefaultArg<D> withDefaultArg(String name, String description, ArgumentType<D> type, D defaultValue, String defaultDisplay)
    public FlagArg withFlagArg(String name, String description)
    
    // List argument variants
    public <D> RequiredArg<List<D>> withListRequiredArg(String name, String description, ArgumentType<D> type)
    public <D> DefaultArg<List<D>> withListDefaultArg(String name, String description, ArgumentType<D> type, List<D> defaultValue, String defaultDisplay)
    public <D> OptionalArg<List<D>> withListOptionalArg(String name, String description, ArgumentType<D> type)
    
    // Wrapped argument variants
    public <W extends WrappedArg<D>, D> W withRequiredArg(String name, String description, ArgWrapper<W, D>)
    public <W extends WrappedArg<D>, D> W withDefaultArg(String name, String description, ArgWrapper<W, D>, D defaultValue, String defaultDisplay)
    public <W extends WrappedArg<D>, D> W withOptionalArg(String name, String description, ArgWrapper<W, D>)
    
    // Configuration
    public void setOwner(CommandOwner owner)
    protected void setUnavailableInSingleplayer(boolean unavailable)
    public void setAllowsExtraArguments(boolean allows)
    
    // Getters
    public String getName()
    public String getDescription()
    public String getFullyQualifiedName()
    public Set<String> getAliases()
    public Map<String, AbstractCommand> getSubCommands()
    public CommandOwner getOwner()
    public String getPermission()
    public List<RequiredArg<?>> getRequiredArguments()
    public boolean isVariant()
    public boolean hasBeenRegistered()
    
    // Usage
    public Message getUsageString(CommandSender sender)
    public Message getUsageShort(CommandSender sender, boolean)
    
    // Registration
    public void completeRegistration() throws GeneralCommandException
}
```

**Known Gotchas:**
- Third constructor parameter is `requiresConfirmation`, NOT `requiresOp`
- `execute()` returns `CompletableFuture<Void>`, always return `CompletableFuture.completedFuture(null)` for sync commands

### CommandContext

**Package:** `com.hypixel.hytale.server.core.command.system.CommandContext`

Context passed to command execution containing parsed arguments and sender info.

```java
// Verified signature from JAR
public final class CommandContext {
    // Constructor
    public CommandContext(AbstractCommand calledCommand, CommandSender sender, String inputString)
    
    // Argument retrieval
    public <DataType> DataType get(Argument<?, DataType> arg)
    public String[] getInput(Argument<?, ?> arg)
    public boolean provided(Argument<?, ?> arg)
    public String getInputString()
    
    // Sender info
    public CommandSender sender()
    public boolean isPlayer()
    public <T extends CommandSender> T senderAs(Class<T> type)
    public Ref<EntityStore> senderAsPlayerRef()
    
    // Messaging
    public void sendMessage(Message message)
    
    // Command info
    public AbstractCommand getCalledCommand()
}
```

**Example Command Implementation:**
```kotlin
class MyCommand : AbstractCommand("mycommand", "Does something") {
    
    private val nameArg = withRequiredArg("name", "Player name", ArgTypes.STRING)
    
    init {
        requirePermission("myplugin.mycommand")
    }
    
    override fun execute(ctx: CommandContext): CompletableFuture<Void> {
        val name = ctx.get(nameArg)
        ctx.sendMessage(Message.raw("Hello, $name!"))
        return CompletableFuture.completedFuture(null)
    }
}
```

### Argument Types

**Package:** `com.hypixel.hytale.server.core.command.system.arguments.types.ArgTypes`

```java
// Verified signature from JAR
public final class ArgTypes {
    // Primitives
    public static final SingleArgumentType<Boolean> BOOLEAN
    public static final SingleArgumentType<Integer> INTEGER
    public static final SingleArgumentType<String> STRING
    public static final SingleArgumentType<Float> FLOAT
    public static final SingleArgumentType<Double> DOUBLE
    public static final SingleArgumentType<UUID> UUID
    public static final SingleArgumentType<Integer> COLOR
    
    // Player-related
    public static final SingleArgumentType<UUID> PLAYER_UUID
    public static final SingleArgumentType<PlayerRef> PLAYER_REF
    public static final SingleArgumentType<CompletableFuture<ProfileServiceClient.PublicGameProfile>> GAME_PROFILE_LOOKUP_ASYNC
    public static final SingleArgumentType<ProfileServiceClient.PublicGameProfile> GAME_PROFILE_LOOKUP
    
    // World/Position
    public static final SingleArgumentType<World> WORLD
    public static final SingleArgumentType<Coord> RELATIVE_DOUBLE_COORD
    public static final SingleArgumentType<IntCoord> RELATIVE_INT_COORD
    public static final SingleArgumentType<RelativeInteger> RELATIVE_INTEGER
    public static final SingleArgumentType<RelativeFloat> RELATIVE_FLOAT
    public static final ArgumentType<Vector2i> VECTOR2I
    public static final ArgumentType<Vector3i> VECTOR3I
    public static final ArgumentType<RelativeVector3i> RELATIVE_VECTOR3I
    public static final ArgumentType<RelativeIntPosition> RELATIVE_BLOCK_POSITION
    public static final ArgumentType<RelativeDoublePosition> RELATIVE_POSITION
    public static final ArgumentType<RelativeChunkPosition> RELATIVE_CHUNK_POSITION
    public static final ArgumentType<Vector3f> ROTATION
    
    // Assets
    public static final SingleArgumentType<ModelAsset> MODEL_ASSET
    public static final SingleArgumentType<Weather> WEATHER_ASSET
    public static final SingleArgumentType<Interaction> INTERACTION_ASSET
    public static final SingleArgumentType<RootInteraction> ROOT_INTERACTION_ASSET
    public static final SingleArgumentType<EntityEffect> EFFECT_ASSET
    public static final SingleArgumentType<Environment> ENVIRONMENT_ASSET
    public static final SingleArgumentType<Item> ITEM_ASSET
    public static final SingleArgumentType<BlockType> BLOCK_TYPE_ASSET
    public static final SingleArgumentType<ParticleSystem> PARTICLE_SYSTEM
    public static final SingleArgumentType<SoundEvent> SOUND_EVENT_ASSET
    public static final SingleArgumentType<AmbienceFX> AMBIENCE_FX_ASSET
    public static final SingleArgumentType<SoundCategory> SOUND_CATEGORY
    
    // Game
    public static final SingleArgumentType<GameMode> GAME_MODE
    public static final SingleArgumentType<Integer> TICK_RATE
    public static final ArgumentType<BlockPattern> BLOCK_PATTERN
    public static final ArgumentType<BlockMask> BLOCK_MASK
    public static final SingleArgumentType<String> BLOCK_TYPE_KEY
    public static final ArgumentType<Integer> BLOCK_ID
    
    // Entity
    public static final ArgWrapper<EntityWrappedArg, UUID> ENTITY_ID
    
    // Ranges
    public static final ArgumentType<Pair<Integer, Integer>> INT_RANGE
    public static final ArgumentType<RelativeIntegerRange> RELATIVE_INT_RANGE
    
    // Enum helper
    public static <E extends Enum<E>> SingleArgumentType<E> forEnum(String name, Class<E> enumClass)
}
```

**Note:** Old docs listed `ArgumentType<T>` for most types, but actual JAR uses `SingleArgumentType<T>` for single-value types.

---

## World Generation

### IWorldGen

**Package:** `com.hypixel.hytale.server.core.universe.world.worldgen.IWorldGen`

Core interface for world generation implementations.

```java
// Verified from JAR - interface exists
public interface IWorldGen {
    // Chunk generation
    CompletableFuture<GeneratedChunk> generate(
        int chunkX, 
        long seed, 
        int chunkZ, 
        LongPredicate shouldCancel
    )
    
    // Spawn point management
    Transform[] getSpawnPoints(int count)
    default ISpawnProvider getDefaultSpawnProvider(int count)
    
    // Performance tracking
    WorldGenTimingsCollector getTimings()
    
    // Lifecycle
    default void shutdown()
}
```

### IWorldGenProvider

**Package:** `com.hypixel.hytale.server.core.universe.world.worldgen.provider.IWorldGenProvider`

Provider interface for creating IWorldGen instances.

```java
public interface IWorldGenProvider {
    static final BuilderCodecMapCodec<IWorldGenProvider> CODEC
    
    IWorldGen getGenerator() throws WorldGenLoadException
}
```

**Built-in Providers:**
- `FlatWorldGenProvider` - Flat world generation
- `VoidWorldGenProvider` - Empty void world
- `DummyWorldGenProvider` - Placeholder/no-op generator

---

## Event System

### EventRegistry

**Package:** `com.hypixel.hytale.event.EventRegistry`

Central registry for event handlers.

```java
// Verified signature from JAR
public class EventRegistry extends Registry<EventRegistration<?, ?>> implements IEventRegistry {
    // Constructor
    public EventRegistry(List<BooleanConsumer> shutdownTasks, BooleanSupplier, String, IEventRegistry parent)
    
    // Synchronous event registration (Void key)
    public <EventType extends IBaseEvent<Void>> EventRegistration<Void, EventType> register(
        Class<? super EventType> eventClass,
        Consumer<EventType> handler
    )
    
    public <EventType extends IBaseEvent<Void>> EventRegistration<Void, EventType> register(
        EventPriority priority,
        Class<? super EventType> eventClass,
        Consumer<EventType> handler
    )
    
    public <EventType extends IBaseEvent<Void>> EventRegistration<Void, EventType> register(
        short priority,  // Raw priority value
        Class<? super EventType> eventClass,
        Consumer<EventType> handler
    )
    
    // Keyed event registration
    public <KeyType, EventType extends IBaseEvent<KeyType>> EventRegistration<KeyType, EventType> register(
        Class<? super EventType> eventClass,
        KeyType key,
        Consumer<EventType> handler
    )
    
    public <KeyType, EventType extends IBaseEvent<KeyType>> EventRegistration<KeyType, EventType> register(
        EventPriority priority,
        Class<? super EventType> eventClass,
        KeyType key,
        Consumer<EventType> handler
    )
    
    // Asynchronous event registration
    public <EventType extends IAsyncEvent<Void>> EventRegistration<Void, EventType> registerAsync(
        Class<? super EventType> eventClass,
        Function<CompletableFuture<EventType>, CompletableFuture<EventType>> handler
    )
    
    public <EventType extends IAsyncEvent<Void>> EventRegistration<Void, EventType> registerAsync(
        EventPriority priority,
        Class<? super EventType> eventClass,
        Function<CompletableFuture<EventType>, CompletableFuture<EventType>> handler
    )
    
    // Global handlers (receive all events of type regardless of key)
    public <KeyType, EventType extends IBaseEvent<KeyType>> EventRegistration<KeyType, EventType> registerGlobal(
        Class<? super EventType> eventClass,
        Consumer<EventType> handler
    )
    
    public <KeyType, EventType extends IBaseEvent<KeyType>> EventRegistration<KeyType, EventType> registerGlobal(
        EventPriority priority,
        Class<? super EventType> eventClass,
        Consumer<EventType> handler
    )
    
    // Async global
    public <KeyType, EventType extends IAsyncEvent<KeyType>> EventRegistration<KeyType, EventType> registerAsyncGlobal(
        Class<? super EventType> eventClass,
        Function<CompletableFuture<EventType>, CompletableFuture<EventType>> handler
    )
    
    // Unhandled event handlers (called when no other handler matched)
    public <KeyType, EventType extends IBaseEvent<KeyType>> EventRegistration<KeyType, EventType> registerUnhandled(
        Class<? super EventType> eventClass,
        Consumer<EventType> handler
    )
    
    public <KeyType, EventType extends IAsyncEvent<KeyType>> EventRegistration<KeyType, EventType> registerAsyncUnhandled(
        Class<? super EventType> eventClass,
        Function<CompletableFuture<EventType>, CompletableFuture<EventType>> handler
    )
}
```

### EventPriority

**Package:** `com.hypixel.hytale.event.EventPriority`

```java
// Verified signature from JAR
public enum EventPriority {
    FIRST,   // Highest priority, executes first (value accessible via getValue())
    EARLY,   // Early execution
    NORMAL,  // Default priority
    LATE,    // Late execution
    LAST     // Lowest priority, executes last
    
    // Methods
    public short getValue()
}
```

### Event Interfaces

**Package:** `com.hypixel.hytale.event`

```java
public interface IBaseEvent<KeyType> {
    // Base marker interface for all events
}

public interface IEvent<KeyType> extends IBaseEvent<KeyType> {
    // Standard synchronous event
}

public interface IAsyncEvent<KeyType> extends IBaseEvent<KeyType> {
    // Asynchronous event
}

public interface ICancellable {
    boolean isCancelled()
    void setCancelled(boolean cancelled)
}
```

### Common Player Events

**Package:** `com.hypixel.hytale.server.core.event.events.player`

```java
// Base class for player events with ECS Ref
public abstract class PlayerEvent<KeyType> implements IEvent<KeyType> {
    public PlayerEvent(Ref<EntityStore> playerRef, Player player)
    public Ref<EntityStore> getPlayerRef()  // ECS reference
    public Player getPlayer()               // Player entity component
}

// Base class for player events with PlayerRef
public abstract class PlayerRefEvent<KeyType> implements IEvent<KeyType> {
    public PlayerRefEvent(PlayerRef playerRef)
    public PlayerRef getPlayerRef()  // Universe PlayerRef
}

// Specific events
public class PlayerReadyEvent extends PlayerEvent<String> {
    public PlayerReadyEvent(Ref<EntityStore>, Player, int readyId)
    public int getReadyId()
}

public class PlayerDisconnectEvent extends PlayerRefEvent<Void> {
    public PlayerDisconnectEvent(PlayerRef playerRef)
    public PacketHandler.DisconnectReason getDisconnectReason()
}

public class PlayerConnectEvent extends PlayerRefEvent<Void> { ... }
public class PlayerChatEvent extends PlayerEvent<Void> implements ICancellable { ... }
public class AddPlayerToWorldEvent extends PlayerRefEvent<Void> { ... }
public class DrainPlayerFromWorldEvent extends PlayerRefEvent<Void> { ... }
```

### Common ECS Events

**Package:** `com.hypixel.hytale.server.core.event.events.ecs`

```java
// Block events
public class BreakBlockEvent implements IEvent<Void>, ICancellable { ... }
public class PlaceBlockEvent implements IEvent<Void>, ICancellable { ... }
public class UseBlockEvent { 
    public class Pre implements IEvent<Void>, ICancellable { ... }
    public class Post implements IEvent<Void> { ... }
}
public class DamageBlockEvent implements IEvent<Void> { ... }

// Item events
public class DropItemEvent {
    public class Drop implements IEvent<Void> { ... }
    public class PlayerRequest implements IEvent<Void>, ICancellable { ... }
}
public class InteractivelyPickupItemEvent implements IEvent<Void>, ICancellable { ... }

// Crafting
public class CraftRecipeEvent {
    public class Pre implements IEvent<Void>, ICancellable { ... }
    public class Post implements IEvent<Void> { ... }
}

// Game mode
public class ChangeGameModeEvent implements IEvent<Void> { ... }
```

---

## Universe and World Management

### Universe

**Package:** `com.hypixel.hytale.server.core.universe.Universe`

The central universe manager - singleton instance managing all worlds and players.

```java
// Verified signature from JAR
public class Universe extends JavaPlugin implements IMessageReceiver, MetricProvider {
    // Singleton access
    public static Universe get()
    
    // World management
    public CompletableFuture<World> addWorld(String name)
    public CompletableFuture<World> addWorld(String name, String type, String generator)
    public CompletableFuture<World> makeWorld(String name, Path path, WorldConfig config)
    public CompletableFuture<World> makeWorld(String name, Path path, WorldConfig config, boolean)
    public CompletableFuture<World> loadWorld(String name)
    public World getWorld(String name)
    public World getWorld(UUID uuid)
    public World getDefaultWorld()
    public boolean removeWorld(String name)
    public void removeWorldExceptionally(String name)
    public Map<String, World> getWorlds()
    public boolean isWorldLoadable(String name)
    
    // Player management
    public List<PlayerRef> getPlayers()
    public PlayerRef getPlayer(UUID uuid)
    public PlayerRef getPlayer(String name, NameMatching matching)
    public PlayerRef getPlayer(String name, Comparator<String>, BiPredicate<String, String>)
    public PlayerRef getPlayerByUsername(String username, NameMatching matching)
    public PlayerRef getPlayerByUsername(String username, Comparator<String>, BiPredicate<String, String>)
    public int getPlayerCount()
    public CompletableFuture<PlayerRef> addPlayer(
        Channel channel, 
        String name, 
        ProtocolVersion version,
        UUID uuid, 
        String username, 
        PlayerAuthentication auth,
        int entityId, 
        PlayerSkin skin
    )
    public void removePlayer(PlayerRef player)
    public CompletableFuture<PlayerRef> resetPlayer(PlayerRef player)
    public CompletableFuture<PlayerRef> resetPlayer(PlayerRef player, Holder<EntityStore>)
    public CompletableFuture<PlayerRef> resetPlayer(PlayerRef player, Holder<EntityStore>, World, Transform)
    
    // Universe lifecycle
    public CompletableFuture<Void> runBackup()
    public void disconnectAllPLayers()  // Note: typo in actual method name
    public void shutdownAllWorlds()
    public CompletableFuture<Void> getUniverseReady()
    
    // Messaging
    public void sendMessage(Message message)
    public void broadcastPacket(Packet packet)
    public void broadcastPacketNoCache(Packet packet)
    public void broadcastPacket(Packet... packets)
    
    // Storage
    public PlayerStorage getPlayerStorage()
    public void setPlayerStorage(PlayerStorage storage)
    public WorldConfigProvider getWorldConfigProvider()
    public ComponentType<EntityStore, PlayerRef> getPlayerRefComponentType()
    
    // Paths
    public Path getPath()
    public static Path getWorldGenPath()
}
```

**Known Gotchas:**
- `disconnectAllPLayers()` has a typo (capital L in PLayers) - this is the actual method name
- Always access via `Universe.get()` singleton

### World

**Package:** `com.hypixel.hytale.server.core.universe.world.World`

Individual world instance with chunk and entity management.

```java
// Verified signature from JAR
public class World extends TickingThread implements Executor, ChunkAccessor<WorldChunk>, IWorldChunks, IMessageReceiver {
    // Constants
    public static final float SAVE_INTERVAL
    public static final String DEFAULT = "world"
    
    // Constructor
    public World(String name, Path savePath, WorldConfig worldConfig) throws IOException
    
    // Lifecycle
    public CompletableFuture<World> init()
    public void stopIndividualWorld()
    public void validateDeleteOnRemove()
    
    // World properties
    public String getName()
    public boolean isAlive()
    public WorldConfig getWorldConfig()
    public DeathConfig getDeathConfig()
    public GameplayConfig getGameplayConfig()
    public int getDaytimeDurationSeconds()
    public int getNighttimeDurationSeconds()
    public boolean isTicking()
    public void setTicking(boolean ticking)
    public boolean isPaused()
    public void setPaused(boolean paused)
    public long getTick()
    public HytaleLogger getLogger()
    public Path getSavePath()
    
    // Chunk management
    public WorldChunk loadChunkIfInMemory(long chunkPos)
    public WorldChunk getChunkIfInMemory(long chunkPos)
    public WorldChunk getChunkIfLoaded(long chunkPos)
    public WorldChunk getChunkIfNonTicking(long chunkPos)
    public CompletableFuture<WorldChunk> getChunkAsync(long chunkPos)
    public CompletableFuture<WorldChunk> getNonTickingChunkAsync(long chunkPos)
    
    // Player management
    public List<Player> getPlayers()
    public int getPlayerCount()
    public Collection<PlayerRef> getPlayerRefs()
    public void trackPlayerRef(PlayerRef playerRef)
    public void untrackPlayerRef(PlayerRef playerRef)
    public CompletableFuture<PlayerRef> addPlayer(PlayerRef playerRef)
    public CompletableFuture<PlayerRef> addPlayer(PlayerRef playerRef, Transform transform)
    public CompletableFuture<PlayerRef> addPlayer(PlayerRef playerRef, Transform transform, Boolean, Boolean)
    public CompletableFuture<Void> drainPlayersTo(World targetWorld)
    
    // Entity management
    public Entity getEntity(UUID uuid)
    public Ref<EntityStore> getEntityRef(UUID uuid)
    public <T extends Entity> T spawnEntity(T entity, Vector3d pos, Vector3f rotation)
    public <T extends Entity> T addEntity(T entity, Vector3d pos, Vector3f rotation, AddReason reason)
    
    // Messaging
    public void sendMessage(Message message)
    
    // Thread execution - CRITICAL FOR ECS ACCESS
    public void execute(Runnable runnable)
    public void consumeTaskQueue()
    
    // Stores
    public ChunkStore getChunkStore()
    public EntityStore getEntityStore()
    public ChunkLightingManager getChunkLighting()
    public WorldMapManager getWorldMapManager()
    public WorldPathConfig getWorldPathConfig()
    public WorldNotificationHandler getNotificationHandler()
    public EventRegistry getEventRegistry()
    
    // Features
    public Map<ClientFeature, Boolean> getFeatures()
    public boolean isFeatureEnabled(ClientFeature feature)
    public void registerFeature(ClientFeature feature, boolean enabled)
    public void broadcastFeatures()
    
    // Ticking
    public void setTps(int tps)
}
```

**CRITICAL: World Thread Safety**

All ECS operations MUST be executed on the world thread:

```kotlin
// CORRECT - runs on world thread
world.execute {
    val player = store.getComponent(ref, Player.getComponentType())
    player.sendMessage(Message.raw("Hello"))
}

// WRONG - may cause race conditions
val player = store.getComponent(ref, Player.getComponentType())  // Don't do this!
```

### WorldConfig

**Package:** `com.hypixel.hytale.server.core.universe.world.WorldConfig`

```java
// Verified signature from JAR (partial - most important fields)
public class WorldConfig {
    // Identity
    private UUID uuid
    private String displayName
    
    // Generation
    private long seed
    private ISpawnProvider spawnProvider
    private IWorldGenProvider worldGenProvider
    private IWorldMapProvider worldMapProvider
    private IChunkStorageProvider chunkStorageProvider
    private ChunkConfig chunkConfig
    
    // Gameplay
    private boolean isTicking
    private boolean isBlockTicking
    private boolean isPvpEnabled
    private boolean isFallDamageEnabled
    private boolean isGameTimePaused
    private Instant gameTime
    private String forcedWeather
    private GameMode gameMode
    private boolean isSpawningNPC
    private boolean isAllNPCFrozen
    
    // Storage
    private boolean isSavingPlayers
    private boolean canSaveChunks
    private boolean saveNewChunks
    private boolean canUnloadChunks
    private boolean deleteOnUniverseStart
    private boolean deleteOnRemove
    
    // Plugin requirements
    private Map<PluginIdentifier, SemverRange> requiredPlugins
    
    // Plugin-specific configuration
    protected MapKeyMapCodec.TypeMap<Object> pluginConfig
    
    // Getters/setters for all fields...
}
```

---

## Player System

### PlayerRef

**Package:** `com.hypixel.hytale.server.core.universe.PlayerRef`

Reference to a player that persists across world transfers.

```java
// Verified signature from JAR
public class PlayerRef implements Component<EntityStore>, MetricProvider, IMessageReceiver {
    // Static factory
    public static ComponentType<EntityStore, PlayerRef> getComponentType()
    
    // Constructor
    public PlayerRef(Holder<EntityStore> holder, UUID uuid, String username, String language, PacketHandler packetHandler, ChunkTracker chunkTracker)
    
    // Identity
    public UUID getUuid()
    public String getUsername()
    public String getLanguage()
    public void setLanguage(String language)
    
    // ECS integration
    public Ref<EntityStore> addToStore(Store<EntityStore> store)
    public void addedToStore(Ref<EntityStore> ref)
    public Holder<EntityStore> removeFromStore()
    public boolean isValid()
    public Ref<EntityStore> getReference()
    public Holder<EntityStore> getHolder()
    public void replaceHolder(Holder<EntityStore> holder)
    
    // Component access
    public <T extends Component<EntityStore>> T getComponent(ComponentType<EntityStore, T> type)
    
    // Position
    public Transform getTransform()
    public UUID getWorldUuid()
    public Vector3f getHeadRotation()
    public void updatePosition(World world, Transform transform, Vector3f headRotation)
    
    // Network
    public PacketHandler getPacketHandler()
    public ChunkTracker getChunkTracker()
    public HiddenPlayersManager getHiddenPlayersManager()
    
    // Messaging
    public void sendMessage(Message message)
    
    // Server transfer
    public void referToServer(String host, int port)
    public void referToServer(String host, int port, byte[] data)
    
    // Clone
    public Component<EntityStore> clone()
}
```

**Known Gotchas:**
- `getUuid()` is the correct method, NOT `uuid` property (Kotlin property access works)
- `getWorldUuid()` returns the UUID of the world the player is in
- To get the World object: `Universe.get().getWorld(playerRef.getWorldUuid())`

### Player (Entity Component)

**Package:** `com.hypixel.hytale.server.core.entity.entities.Player`

The Player entity component with gameplay state.

```java
// Verified signature from JAR (partial - key methods)
public class Player extends LivingEntity implements CommandSender, PermissionHolder, MetricProvider {
    // Static factory
    public static ComponentType<EntityStore, Player> getComponentType()
    
    // Initialization
    public void init(UUID uuid, PlayerRef playerRef)
    public void copyFrom(Player other)
    
    // Managers
    public WindowManager getWindowManager()
    public PageManager getPageManager()
    public HudManager getHudManager()
    public HotbarManager getHotbarManager()
    public WorldMapTracker getWorldMapTracker()
    
    // Player data
    public PlayerConfigData getPlayerConfigData()
    public void markNeedsSave()
    
    // Game mode
    public GameMode getGameMode()  // Inherited from LivingEntity or defined here
    
    // Messaging
    public void sendMessage(Message message)
    
    // Permissions
    public boolean hasPermission(String permission)
    public boolean hasPermission(String permission, boolean defaultValue)
    
    // Network
    public PacketHandler getPlayerConnection()
    
    // Spawn state
    public boolean isFirstSpawn()
    public void setFirstSpawn(boolean firstSpawn)
    public boolean hasSpawnProtection()
    
    // View
    public int getClientViewRadius()
    public void setClientViewRadius(int radius)
    public int getViewRadius()
    
    // Client ready state
    public void startClientReadyTimeout()
    public void handleClientReady(boolean ready)
    public boolean isWaitingForClientReady()
    
    // Inventory
    public void sendInventory()
    public Inventory setInventory(Inventory inventory)
}
```

---

## ECS Component System

### Ref<T>

**Package:** `com.hypixel.hytale.component.Ref`

Reference wrapper for ECS entities.

```java
// Verified signature from JAR
public class Ref<ECS_TYPE> {
    public static final Ref<?>[] EMPTY_ARRAY
    
    // Constructors
    public Ref(Store<ECS_TYPE> store)
    public Ref(Store<ECS_TYPE> store, int index)
    
    // Access
    public Store<ECS_TYPE> getStore()
    public int getIndex()
    
    // Validity
    public boolean isValid()
    public void validate()
    
    // Equality
    public boolean equals(Object obj)
    public boolean equals(Ref<ECS_TYPE> other)
    public int hashCode()
}
```

**Note:** Unlike old docs, `Ref` does NOT have `getId()` or `get()` or `remove()` methods. Use `getIndex()` and work through the Store.

### Store<T>

**Package:** `com.hypixel.hytale.component.Store`

ECS store containing entities and their components.

```java
// Verified signature from JAR (key methods)
public class Store<ECS_TYPE> implements ComponentAccessor<ECS_TYPE> {
    // Component access
    public <T extends Component<ECS_TYPE>> T getComponent(Ref<ECS_TYPE> ref, ComponentType<ECS_TYPE, T> type)
    public <T extends Component<ECS_TYPE>> T addComponent(Ref<ECS_TYPE> ref, ComponentType<ECS_TYPE, T> type)
    public <T extends Component<ECS_TYPE>> void addComponent(Ref<ECS_TYPE> ref, ComponentType<ECS_TYPE, T> type, T component)
    public <T extends Component<ECS_TYPE>> void replaceComponent(Ref<ECS_TYPE> ref, ComponentType<ECS_TYPE, T> type, T component)
    public <T extends Component<ECS_TYPE>> void putComponent(Ref<ECS_TYPE> ref, ComponentType<ECS_TYPE, T> type, T component)
    public <T extends Component<ECS_TYPE>> void removeComponent(Ref<ECS_TYPE> ref, ComponentType<ECS_TYPE, T> type)
    public <T extends Component<ECS_TYPE>> void tryRemoveComponent(Ref<ECS_TYPE> ref, ComponentType<ECS_TYPE, T> type)
    public <T extends Component<ECS_TYPE>> boolean removeComponentIfExists(Ref<ECS_TYPE> ref, ComponentType<ECS_TYPE, T> type)
    public <T extends Component<ECS_TYPE>> T ensureAndGetComponent(Ref<ECS_TYPE> ref, ComponentType<ECS_TYPE, T> type)
    
    // Entity management
    public Ref<ECS_TYPE> addEntity(Archetype<ECS_TYPE> archetype, AddReason reason)
    public Ref<ECS_TYPE> addEntity(Holder<ECS_TYPE> holder, AddReason reason)
    public Holder<ECS_TYPE> removeEntity(Ref<ECS_TYPE> ref, RemoveReason reason)
    public Holder<ECS_TYPE> copyEntity(Ref<ECS_TYPE> ref)
    public Archetype<ECS_TYPE> getArchetype(Ref<ECS_TYPE> ref)
    
    // Resource access
    public <T extends Resource<ECS_TYPE>> T getResource(ResourceType<ECS_TYPE, T> type)
    public <T extends Resource<ECS_TYPE>> void replaceResource(ResourceType<ECS_TYPE, T> type, T resource)
    
    // Iteration
    public void forEachChunk(BiConsumer<ArchetypeChunk<ECS_TYPE>, CommandBuffer<ECS_TYPE>> consumer)
    public void forEachChunk(Query<ECS_TYPE> query, BiConsumer<ArchetypeChunk<ECS_TYPE>, CommandBuffer<ECS_TYPE>> consumer)
    public void forEachEntityParallel(IntBiObjectConsumer<ArchetypeChunk<ECS_TYPE>, CommandBuffer<ECS_TYPE>> consumer)
    
    // Events
    public <Event extends EcsEvent> void invoke(Ref<ECS_TYPE> ref, Event event)
    public <Event extends EcsEvent> void invoke(Event event)
    
    // Ticking
    public void tick(float delta)
    public void pausedTick(float delta)
    
    // State
    public int getEntityCount()
    public boolean isShutdown()
    public boolean isProcessing()
    public void assertThread()
    public boolean isInThread()
    
    // Registry
    public ComponentRegistry<ECS_TYPE> getRegistry()
    public ECS_TYPE getExternalData()
}
```

**ECS Usage Pattern:**
```kotlin
// Always access ECS on the world thread
world.execute {
    val store = world.entityStore
    val player = store.getComponent(playerRef, Player.getComponentType())
    
    // Do something with player
    player.sendMessage(Message.raw("Hello!"))
}
```

---

## Noise and Procedural Generation

### NoiseFunction

**Package:** `com.hypixel.hytale.procedurallib.NoiseFunction`

Core interface for noise functions, combining 2D and 3D capabilities.

```java
// Verified signature from JAR
public interface NoiseFunction extends NoiseFunction2d, NoiseFunction3d {
    // Inherits methods from both interfaces
}
```

### NoiseFunction2d

**Package:** `com.hypixel.hytale.procedurallib.NoiseFunction2d`

```java
public interface NoiseFunction2d {
    double get(int seed, int octaves, double x, double z)
}
```

### NoiseFunction3d

**Package:** `com.hypixel.hytale.procedurallib.NoiseFunction3d`

```java
public interface NoiseFunction3d {
    double get(int seed, int octaves, double x, double y, double z)
}
```

### Built-in Noise Implementations

#### SimplexNoise

**Package:** `com.hypixel.hytale.procedurallib.logic.SimplexNoise`

```java
// Verified signature from JAR
public class SimplexNoise implements NoiseFunction {
    public static final SimplexNoise INSTANCE
    
    // Private constructor - use INSTANCE
    private SimplexNoise()
    
    public double get(int seed, int octaves, double x, double z)
    public double get(int seed, int octaves, double x, double y, double z)
}
```

#### PerlinNoise

**Package:** `com.hypixel.hytale.procedurallib.logic.PerlinNoise`

```java
// Verified signature from JAR
public class PerlinNoise implements NoiseFunction {
    public PerlinNoise(GeneralNoise.InterpolationFunction interpolation)
    
    public double get(int seed, int octaves, double x, double z)
    public double get(int seed, int octaves, double x, double y, double z)
    
    public GeneralNoise.InterpolationFunction getInterpolationFunction()
}
```

#### CellNoise (Voronoi/Cellular)

**Package:** `com.hypixel.hytale.procedurallib.logic.CellNoise`

```java
// Verified signature from JAR
public class CellNoise implements NoiseFunction {
    public CellNoise(
        CellDistanceFunction distanceFunction,
        PointEvaluator pointEvaluator,
        CellFunction cellFunction,
        NoiseProperty noiseLookup
    )
    
    public double get(int seed, int octaves, double x, double z)
    public double get(int seed, int octaves, double x, double y, double z)
    
    public CellDistanceFunction getDistanceFunction()
    public CellFunction getCellFunction()
    public NoiseProperty getNoiseLookup()
}
```

**CellMode enum:** (nested in CellNoise)
```java
public enum CellMode {
    CELL_VALUE,     // Returns cell value
    NOISE_LOOKUP,   // Uses noise lookup
    DISTANCE,       // Returns distance to nearest cell
    DISTANCE_2      // Returns distance to 2nd nearest cell
}
```

---

## Asset System

### AssetRegistry

**Package:** `com.hypixel.hytale.server.core.plugin.registry.AssetRegistry`

Registry for plugin-specific assets.

```java
// Actual signature - simpler than old docs suggested
public class AssetRegistry {
    public void register(Asset asset)
    public void unregister(String id)
    public <T extends Asset> T get(String id, Class<T> type)
    public boolean has(String id)
    public void registerAll(Collection<Asset> assets)
    public void clear()
}
```

---

## Logging System

### HytaleLogger

**Package:** `com.hypixel.hytale.logger.HytaleLogger`

```java
// Verified signature from JAR
public class HytaleLogger extends AbstractLogger<HytaleLogger.Api> {
    // Static accessors
    public static HytaleLogger getLogger()
    public static HytaleLogger forEnclosingClass()
    public static HytaleLogger forEnclosingClassFull()
    public static HytaleLogger get(String name)
    
    // Logging
    public HytaleLogger.Api at(Level level)
    
    // Configuration
    public String getName()
    public Level getLevel()
    public void setLevel(Level level)
    public HytaleLogger getSubLogger(String name)
    
    // Sentry integration
    public void setSentryClient(IScopes scopes)
    public void setPropagatesSentryToParent(boolean propagates)
}
```

**Usage:**
```kotlin
class MyPlugin(init: JavaPluginInit) : JavaPlugin(init) {
    override fun setup() {
        logger.atInfo().log("Plugin setting up")
        logger.atFine().log("Debug message")
        logger.atSevere().withCause(exception).log("Error occurred")
    }
}
```

---

## Common Pitfalls and Best Practices

### 1. World Thread Safety

**CRITICAL:** All ECS operations must occur on the World thread.

```kotlin
// CORRECT
world.execute {
    val store = world.entityStore
    val player = store.getComponent(ref, Player.getComponentType())
    // Safe to access player here
}

// WRONG - Race condition!
val player = world.entityStore.getComponent(ref, Player.getComponentType())
```

### 2. Event Registration with Error Handling

Always wrap event handlers in try/catch:

```kotlin
eventRegistry.register(PlayerReadyEvent::class.java) { event ->
    try {
        // Handle event
        val player = event.player
        player.sendMessage(Message.raw("Welcome!"))
    } catch (e: Exception) {
        logger.atSevere().withCause(e).log("Error handling PlayerReadyEvent")
    }
}
```

### 3. PlayerRef vs Player vs Ref<EntityStore>

- **PlayerRef** (`com.hypixel.hytale.server.core.universe.PlayerRef`) - Persistent reference to a player, survives world transfers
- **Player** (`com.hypixel.hytale.server.core.entity.entities.Player`) - ECS component with player data
- **Ref<EntityStore>** - ECS entity reference, may be invalidated on world transfer

```kotlin
// From PlayerReadyEvent
val ref: Ref<EntityStore> = event.playerRef  // ECS reference
val player: Player = event.player            // Player component

// From PlayerRef
val playerRef: PlayerRef = Universe.get().getPlayer(uuid)
val ecsRef: Ref<EntityStore> = playerRef.reference
val world: World = Universe.get().getWorld(playerRef.worldUuid)

// Getting Player from Ref (must be on world thread)
world.execute {
    val player = world.entityStore.getComponent(ref, Player.getComponentType())
}
```

### 4. Getting World from PlayerRef

PlayerRef does NOT have a `world` property. Access world via UUID:

```kotlin
val playerRef: PlayerRef = ...
val worldUuid: UUID = playerRef.worldUuid
val world: World = Universe.get().getWorld(worldUuid)
```

### 5. Command Return Type

Commands must return `CompletableFuture<Void>`:

```kotlin
override fun execute(ctx: CommandContext): CompletableFuture<Void> {
    // Synchronous command
    ctx.sendMessage(Message.raw("Done!"))
    return CompletableFuture.completedFuture(null)
}

override fun execute(ctx: CommandContext): CompletableFuture<Void> {
    // Async command
    return someAsyncOperation().thenApply { result ->
        ctx.sendMessage(Message.raw("Result: $result"))
        null  // Must return null for Void
    }
}
```

### 6. Message Creation

Use `Message.raw()` for plain text, `Message.translation()` for localized text:

```kotlin
// Plain text
ctx.sendMessage(Message.raw("Hello, world!"))

// With color (using chain methods)
ctx.sendMessage(Message.raw("Error!").color("#FF0000"))

// Translated (requires client-side translation files)
ctx.sendMessage(Message.translation("myplugin.welcome"))

// With parameters
ctx.sendMessage(Message.translation("myplugin.greeting").param("name", playerName))
```

### 7. PluginState Enum Values

The actual enum values are different from what some old docs say:

```java
// CORRECT (verified)
NONE, SETUP, START, ENABLED, SHUTDOWN, DISABLED

// WRONG (old docs)
PENDING, LOADING, DISABLING, ERROR  // These do NOT exist!
```

### 8. PageManager and HudManager Access

Access through Player component, not directly:

```kotlin
world.execute {
    val player = store.getComponent(ref, Player.getComponentType())
    
    // HUD management
    val hudManager = player.hudManager
    
    // Page management (for full-screen UI)
    val pageManager = player.pageManager
}
```

---

*Generated and verified from HytaleServer.jar analysis - January 31, 2026*
