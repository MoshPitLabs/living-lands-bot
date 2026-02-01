# Implementation Plan: Code Review Critical Fixes

**Project:** Living Lands Discord Bot  
**Created:** 2026-02-01  
**Target Version:** v0.2.0-fixes  
**Status:** Ready for Implementation

---

## Executive Summary

This document provides a detailed implementation plan for fixing 12 critical and high-priority issues identified in the code review. The fixes are organized into three phases based on priority and dependencies.

**Estimated Total Effort:** 8-12 hours  
**Risk Level:** Medium (database migration required)

---

## Phase 1: Critical Fixes (Must Do Immediately)

### 1.1 Race Condition in RAGService

| Attribute | Value |
|-----------|-------|
| **Priority** | CRITICAL |
| **File(s)** | `internal/services/rag.go` |
| **Risk** | Medium - Core service change |
| **Estimated Time** | 30 minutes |
| **Dependencies** | None |

**Issue:** The `collectionID` field is accessed concurrently in `ensureCollection()` without synchronization. Multiple goroutines calling `Query()` simultaneously could cause a data race.

**Current Code (lines 362-366):**
```go
func (s *RAGService) ensureCollection(ctx context.Context) error {
    // Return early if we've already cached the collection ID
    if s.collectionID != "" {
        return nil
    }
    // ...
```

**Implementation Approach:**

1. Add `sync.RWMutex` field to `RAGService` struct
2. Use double-checked locking pattern:
   - Acquire read lock, check if set, release
   - If not set, acquire write lock, re-check, then set
3. All reads of `collectionID` must hold at least read lock

**Code Changes:**

```go
// In RAGService struct (line 23-32)
type RAGService struct {
    chromaURL          string
    ollamaClient       *ollama.Client
    httpClient         *http.Client
    embedModel         string
    logger             *slog.Logger
    collectionID       string
    collectionName     string
    relevanceThreshold float32
    mu                 sync.RWMutex // ADD THIS
}

// In ensureCollection (line 362)
func (s *RAGService) ensureCollection(ctx context.Context) error {
    // Fast path: read lock check
    s.mu.RLock()
    if s.collectionID != "" {
        s.mu.RUnlock()
        return nil
    }
    s.mu.RUnlock()

    // Slow path: write lock with double-check
    s.mu.Lock()
    defer s.mu.Unlock()

    // Double-check after acquiring write lock
    if s.collectionID != "" {
        return nil
    }

    // ... rest of method (lines 368-408)
}

// Also update createCollection (line 462) to be called within lock
// s.collectionID = collID assignment is already within the locked ensureCollection
```

**Testing Requirements:**
- Add test: `TestRAGService_ConcurrentQuery` - spawn 50 goroutines querying simultaneously
- Add test: `TestRAGService_RaceCondition` - use `-race` flag
- Run existing `rag_test.go` and `rag_integration_test.go`

**Gotchas:**
- Ensure mutex is not held across HTTP calls (would block other goroutines)
- The lock only protects `collectionID` field, not the HTTP client

---

### 1.2 Missing Rate Limiting Enforcement

| Attribute | Value |
|-----------|-------|
| **Priority** | CRITICAL |
| **File(s)** | `internal/bot/commands.go`, `internal/bot/bot.go` |
| **Risk** | Low - Adding logic only |
| **Estimated Time** | 45 minutes |
| **Dependencies** | None |

**Issue:** The `RateLimiter` service exists and is initialized but never called in `handleAskCommand`. Users can spam the `/ask` command without restriction.

**Implementation Approach:**

1. Add `limiter *services.RateLimiter` to `CommandHandlers` struct
2. Pass rate limiter from `bot.go` to `CommandHandlers`
3. Check rate limit at start of `handleAskCommand`
4. Return user-friendly message when rate limited

**Code Changes:**

```go
// internal/bot/commands.go - Update struct (line 14-19)
type CommandHandlers struct {
    account *services.AccountService
    rag     *services.RAGService
    llm     *services.LLMService
    limiter *services.RateLimiter // ADD THIS
    logger  *slog.Logger
}

// Update constructor (line 21-28)
func NewCommandHandlers(
    account *services.AccountService,
    rag *services.RAGService,
    llm *services.LLMService,
    limiter *services.RateLimiter, // ADD THIS
    logger *slog.Logger,
) *CommandHandlers {
    return &CommandHandlers{
        account: account,
        rag:     rag,
        llm:     llm,
        limiter: limiter, // ADD THIS
        logger:  logger,
    }
}

// In handleAskCommand, after startTime declaration (line 185-200)
func (h *CommandHandlers) handleAskCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
    startTime := time.Now()

    // Extract user ID safely for rate limiting
    var userID string
    if i.Member != nil && i.Member.User != nil {
        userID = i.Member.User.ID
    } else if i.User != nil {
        userID = i.User.ID
    } else {
        // Cannot identify user - reject request
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Unable to identify user. Please try again.",
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
        return
    }

    // Check rate limit
    ctx := context.Background()
    allowed, remaining, ttl, err := h.limiter.IsAllowed(ctx, userID)
    if err != nil {
        h.logger.Error("rate limit check failed", "error", err, "user_id", userID)
        // Fail open - allow request if rate limiter is unavailable
    } else if !allowed {
        h.logger.Warn("user rate limited", "user_id", userID, "ttl_seconds", ttl.Seconds())
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf(
                    "Whoa there, traveler! The ancient scrolls need rest. "+
                    "You can ask again in %d seconds. (Limit: %d questions per minute)",
                    int(ttl.Seconds()), remaining+1,
                ),
                Flags: discordgo.MessageFlagsEphemeral,
            },
        })
        return
    }

    // ... rest of handleAskCommand
```

**Also update bot.go (line 32):**
```go
handlers := NewCommandHandlers(account, rag, llm, limiter, logger)
```

**Testing Requirements:**
- Add test: `TestHandleAskCommand_RateLimited`
- Add test: `TestHandleAskCommand_RateLimiterFailure` (fail-open behavior)
- Manual test: Spam `/ask` in Discord, verify rate limiting

---

### 1.3 Context Timeout Hierarchy Violation

| Attribute | Value |
|-----------|-------|
| **Priority** | CRITICAL |
| **File(s)** | `internal/bot/commands.go` |
| **Risk** | Low |
| **Estimated Time** | 20 minutes |
| **Dependencies** | None |

**Issue:** Child context for RAG is created with 5-second timeout from parent, but parent may have less time remaining. This could cause child context to outlive parent.

**Current Code (lines 250-258):**
```go
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()

// ...

ragCtx, ragCancel := context.WithTimeout(ctx, 5*time.Second)
defer ragCancel()
```

**Implementation Approach:**

Check parent context's deadline and use the minimum of the desired timeout and remaining parent time.

**Code Changes:**

```go
// internal/bot/commands.go - Add helper function after imports
func contextWithSafeTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
    if deadline, ok := parent.Deadline(); ok {
        remaining := time.Until(deadline)
        if remaining < timeout {
            timeout = remaining
        }
        // Ensure we don't create a context with zero/negative timeout
        if timeout <= 0 {
            // Return a canceled context immediately
            ctx, cancel := context.WithCancel(parent)
            cancel()
            return ctx, cancel
        }
    }
    return context.WithTimeout(parent, timeout)
}

// Update line 257
ragCtx, ragCancel := contextWithSafeTimeout(ctx, 5*time.Second)
defer ragCancel()
```

**Testing Requirements:**
- Add test: `TestContextWithSafeTimeout_RespectParentDeadline`
- Add test: `TestContextWithSafeTimeout_ParentExpired`

---

### 1.4 Database Connection Leaks

| Attribute | Value |
|-----------|-------|
| **Priority** | CRITICAL |
| **File(s)** | `cmd/bot/main.go` |
| **Risk** | Low |
| **Estimated Time** | 15 minutes |
| **Dependencies** | None |

**Issue:** In `handleMigrate` and `handleIndexDocs`, database connections are opened but never closed before `os.Exit()`.

**Implementation Approach:**

Add `defer db.Close()` after successful `database.Open()` calls.

**Code Changes:**

```go
// handleMigrate (line 51-64)
func handleMigrate(cfg *config.Config, logger *slog.Logger) {
    db, err := database.Open(cfg)
    if err != nil {
        logger.Error("db open failed", "error", err)
        os.Exit(1)
    }
    defer db.Close() // ADD THIS

    if err := database.RunMigrations(db, "migrations"); err != nil {
        logger.Error("migrations failed", "error", err)
        os.Exit(1)
    }

    logger.Info("migrations complete")
}

// handleIndexDocs (line 82-120)
func handleIndexDocs(cfg *config.Config, logger *slog.Logger) {
    // ... flag parsing ...

    db, err := database.Open(cfg)
    if err != nil {
        logger.Error("db open failed", "error", err)
        os.Exit(1)
    }
    defer db.Close() // ADD THIS

    // ... rest of function
    // Remove unused db reference at end (line 120)
}
```

**Note:** Check if `database.DB` has a `Close()` method. If using GORM, you may need:
```go
sqlDB, _ := db.Gorm.DB()
defer sqlDB.Close()
```

**Testing Requirements:**
- Verify connection pool metrics after repeated CLI operations
- Integration test: Run migrate 10 times, check PostgreSQL connection count

---

### 1.5 Nil Pointer Risk in Discord Interactions

| Attribute | Value |
|-----------|-------|
| **Priority** | CRITICAL |
| **File(s)** | `internal/bot/commands.go` |
| **Risk** | Low |
| **Estimated Time** | 20 minutes |
| **Dependencies** | None |

**Issue:** In `handleLinkCommand`, `i.Member.User.ID` is accessed without checking if `i.Member.User` is nil. In DMs, `i.Member` is nil but `i.User` is set.

**Current Code (lines 101-112):**
```go
func (h *CommandHandlers) handleLinkCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
    var discordID int64
    var username string

    if i.Member != nil {
        fmt.Sscanf(i.Member.User.ID, "%d", &discordID) // DANGER: i.Member.User could be nil
        username = i.Member.User.Username
    } else if i.User != nil {
        // ...
    }
```

**Implementation Approach:**

Add nil checks for nested User field and use a helper function for consistent user extraction.

**Code Changes:**

```go
// Add helper function at package level (after imports)
// getUserFromInteraction safely extracts user info from an interaction.
// Returns userID (as string), username, and success boolean.
func getUserFromInteraction(i *discordgo.InteractionCreate) (string, string, bool) {
    // Guild context: Member is set
    if i.Member != nil && i.Member.User != nil {
        return i.Member.User.ID, i.Member.User.Username, true
    }
    // DM context: User is set directly
    if i.User != nil {
        return i.User.ID, i.User.Username, true
    }
    return "", "", false
}

// Update handleLinkCommand (line 101-112)
func (h *CommandHandlers) handleLinkCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID, username, ok := getUserFromInteraction(i)
    if !ok {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Unable to identify your account. Please try again.",
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
        return
    }

    var discordID int64
    fmt.Sscanf(userID, "%d", &discordID)

    code, err := h.account.GenerateVerificationCode(discordID, username)
    // ... rest of function
}
```

**Also update the user extraction in handleAskCommand (lines 305-310):**
```go
// Replace:
var username string
if i.Member != nil && i.Member.User != nil {
    username = i.Member.User.Username
} else if i.User != nil {
    username = i.User.Username
}

// With:
_, username, _ := getUserFromInteraction(i)
```

**Testing Requirements:**
- Add test: `TestGetUserFromInteraction_GuildContext`
- Add test: `TestGetUserFromInteraction_DMContext`
- Add test: `TestGetUserFromInteraction_NilMember`

---

### 1.6 Poor Ollama Error Handling

| Attribute | Value |
|-----------|-------|
| **Priority** | CRITICAL |
| **File(s)** | `pkg/ollama/client.go` |
| **Risk** | Low |
| **Estimated Time** | 20 minutes |
| **Dependencies** | None |

**Issue:** When Ollama returns a non-200 status, the error message only includes the status code, not the response body which often contains useful debugging information.

**Current Code (lines 93-94, 130-131):**
```go
if resp.StatusCode != http.StatusOK {
    return nil, fmt.Errorf("ollama returned %d", resp.StatusCode)
}
```

**Implementation Approach:**

Read and include response body in error messages (with size limit to avoid huge error messages).

**Code Changes:**

```go
// pkg/ollama/client.go - Add helper function
func formatHTTPError(resp *http.Response) error {
    // Read limited body for error message
    const maxBodySize = 512
    body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
    if err != nil {
        return fmt.Errorf("ollama returned %d (could not read body: %v)", resp.StatusCode, err)
    }
    
    bodyStr := strings.TrimSpace(string(body))
    if len(bodyStr) > 0 {
        return fmt.Errorf("ollama returned %d: %s", resp.StatusCode, bodyStr)
    }
    return fmt.Errorf("ollama returned %d", resp.StatusCode)
}

// Add import at top
import (
    // ... existing imports
    "io"
    "strings"
)

// Update Generate method (line 93-94)
if resp.StatusCode != http.StatusOK {
    return nil, formatHTTPError(resp)
}

// Update Embed method (line 130-131)
if resp.StatusCode != http.StatusOK {
    return nil, formatHTTPError(resp)
}
```

**Testing Requirements:**
- Add test: `TestClient_Generate_ErrorResponse`
- Add test: `TestFormatHTTPError_WithBody`
- Add test: `TestFormatHTTPError_EmptyBody`

---

## Phase 2: High Priority (This Week)

### 2.1 Discord ID Type Issue

| Attribute | Value |
|-----------|-------|
| **Priority** | HIGH |
| **File(s)** | Multiple (see below) |
| **Risk** | HIGH - Database migration required |
| **Estimated Time** | 2 hours |
| **Dependencies** | Phase 1 completion |

**Issue:** Discord IDs are stored as `int64` (BIGINT) but can exceed 64-bit signed integer range. Discord IDs are "snowflakes" - 64-bit unsigned integers. Some already exceed `9223372036854775807` (max int64).

**Files to Modify:**
1. `internal/database/models/user.go`
2. `internal/services/account.go`
3. `internal/bot/commands.go`
4. `migrations/0004_discord_id_to_string.up.sql` (NEW)
5. `migrations/0004_discord_id_to_string.down.sql` (NEW)

**Implementation Approach:**

1. Create database migration to change `discord_id` from BIGINT to VARCHAR
2. Update Go model to use `string` type
3. Update all code that parses/uses Discord IDs
4. Remove `fmt.Sscanf` conversion, use string directly

**Database Migration:**

```sql
-- migrations/0004_discord_id_to_string.up.sql
-- Change discord_id from BIGINT to VARCHAR(20) to handle snowflake IDs safely
-- Discord snowflakes are 64-bit unsigned integers, max 18446744073709551615 (20 digits)

-- Step 1: Add new column
ALTER TABLE users ADD COLUMN discord_id_new VARCHAR(20);

-- Step 2: Migrate data (cast BIGINT to VARCHAR)
UPDATE users SET discord_id_new = CAST(discord_id AS VARCHAR(20));

-- Step 3: Drop old column
ALTER TABLE users DROP COLUMN discord_id;

-- Step 4: Rename new column
ALTER TABLE users RENAME COLUMN discord_id_new TO discord_id;

-- Step 5: Add constraints back
ALTER TABLE users ALTER COLUMN discord_id SET NOT NULL;
CREATE UNIQUE INDEX idx_users_discord_id ON users (discord_id);
```

```sql
-- migrations/0004_discord_id_to_string.down.sql
-- WARNING: This may fail if any discord_id values exceed BIGINT range
ALTER TABLE users ADD COLUMN discord_id_old BIGINT;
UPDATE users SET discord_id_old = CAST(discord_id AS BIGINT);
ALTER TABLE users DROP COLUMN discord_id;
ALTER TABLE users RENAME COLUMN discord_id_old TO discord_id;
ALTER TABLE users ALTER COLUMN discord_id SET NOT NULL;
CREATE UNIQUE INDEX idx_users_discord_id ON users (discord_id);
```

**Code Changes:**

```go
// internal/database/models/user.go
type User struct {
    ID               uint   `gorm:"primaryKey"`
    DiscordID        string `gorm:"uniqueIndex;not null;size:20"` // CHANGED from int64
    DiscordUsername  string `gorm:"not null"`
    // ... rest unchanged
}

// internal/services/account.go
// Change GenerateVerificationCode signature (line 31)
func (s *AccountService) GenerateVerificationCode(discordID string, discordUsername string) (string, error) {
    // Update query (line 40)
    err := s.db.Where("discord_id = ?", discordID).
        Assign(user).
        FirstOrCreate(user).Error
    // ... rest unchanged
}

// internal/bot/commands.go - handleLinkCommand
func (h *CommandHandlers) handleLinkCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
    userID, username, ok := getUserFromInteraction(i)
    if !ok {
        // ... error handling
        return
    }

    // REMOVE: fmt.Sscanf conversion - use string directly
    code, err := h.account.GenerateVerificationCode(userID, username)
    // ...
}
```

**Backward Compatibility:**
- Existing data will be migrated automatically
- No API changes for external consumers (Hytale mod)
- GORM will handle the type change automatically

**Testing Requirements:**
- Test migration on copy of production database
- Add test: `TestGenerateVerificationCode_LargeDiscordID`
- Verify existing records are preserved after migration

**Risks:**
- Migration must run during maintenance window
- Rollback could fail if new IDs exceed int64 range

---

### 2.2 Prompt Injection Risk

| Attribute | Value |
|-----------|-------|
| **Priority** | HIGH |
| **File(s)** | `internal/services/llm.go` |
| **Risk** | Medium - Security fix |
| **Estimated Time** | 45 minutes |
| **Dependencies** | None |

**Issue:** User input is directly concatenated into prompts without sanitization. A malicious user could inject prompt manipulation.

**Current Code (lines 350-371):**
```go
func (s *LLMService) buildPrompt(userMessage string, ragContext []string, mode ResponseMode) string {
    var prompt strings.Builder
    // ...
    prompt.WriteString(fmt.Sprintf("User: %s\nAssistant:", userMessage))
    return prompt.String()
}
```

**Implementation Approach:**

1. Create sanitization function to neutralize prompt injection patterns
2. Apply to user messages and RAG context
3. Add logging for potential injection attempts

**Code Changes:**

```go
// internal/services/llm.go - Add sanitization function

// sanitizePromptInput removes or escapes patterns that could manipulate LLM behavior.
// This is defense-in-depth against prompt injection attacks.
func sanitizePromptInput(input string) string {
    if input == "" {
        return input
    }

    // List of patterns that might manipulate the conversation
    injectionPatterns := []string{
        "System:",
        "SYSTEM:",
        "system:",
        "Assistant:",
        "ASSISTANT:",
        "assistant:",
        "User:",
        "USER:",
        "user:",
        "```system",
        "```assistant",
        "<|system|>",
        "<|user|>",
        "<|assistant|>",
        "[INST]",
        "[/INST]",
        "<<SYS>>",
        "<</SYS>>",
        "### Instruction:",
        "### Response:",
        "Ignore previous instructions",
        "ignore previous instructions",
        "Ignore all previous",
        "ignore all previous",
        "Disregard your instructions",
        "disregard your instructions",
    }

    result := input
    for _, pattern := range injectionPatterns {
        // Replace with escaped version (wrap in brackets to neutralize)
        result = strings.ReplaceAll(result, pattern, "["+pattern+"]")
    }

    // Also limit length to prevent context exhaustion
    const maxInputLength = 2000
    if len(result) > maxInputLength {
        result = result[:maxInputLength]
    }

    return result
}

// detectPotentialInjection logs suspicious patterns for monitoring
func (s *LLMService) detectPotentialInjection(input string) bool {
    lowerInput := strings.ToLower(input)
    suspiciousPatterns := []string{
        "ignore",
        "disregard",
        "system:",
        "assistant:",
        "pretend",
        "roleplay as",
        "you are now",
        "act as",
    }

    for _, pattern := range suspiciousPatterns {
        if strings.Contains(lowerInput, pattern) {
            s.logger.Warn("potential prompt injection detected",
                "input_preview", input[:min(len(input), 100)],
                "pattern", pattern,
            )
            return true
        }
    }
    return false
}

// Update buildPrompt to use sanitization (line 349)
func (s *LLMService) buildPrompt(userMessage string, ragContext []string, mode ResponseMode) string {
    var prompt strings.Builder

    // Sanitize user input
    cleanMessage := sanitizePromptInput(userMessage)
    s.detectPotentialInjection(userMessage) // Log but don't block

    if mode == ModeDeep && len(ragContext) > 0 {
        prompt.WriteString("Relevant documentation (use only if it answers the question):\n")
        for i, ctx := range ragContext {
            // Also sanitize RAG context (defense in depth)
            truncated := sanitizePromptInput(ctx)
            if len(truncated) > 500 {
                truncated = truncated[:500] + "..."
            }
            prompt.WriteString(fmt.Sprintf("%d. %s\n", i+1, truncated))
        }
        prompt.WriteString("\n---\n\n")
    }

    prompt.WriteString(fmt.Sprintf("User: %s\nAssistant:", cleanMessage))
    return prompt.String()
}
```

**Testing Requirements:**
- Add test: `TestSanitizePromptInput_InjectionPatterns`
- Add test: `TestSanitizePromptInput_LengthLimit`
- Add test: `TestDetectPotentialInjection`
- Manual test: Try injection attacks via `/ask` command

---

### 2.3 Resource Cleanup on Startup Failure

| Attribute | Value |
|-----------|-------|
| **Priority** | HIGH |
| **File(s)** | `cmd/bot/main.go` |
| **Risk** | Low |
| **Estimated Time** | 30 minutes |
| **Dependencies** | None |

**Issue:** In `startBot()`, if initialization fails after Redis/DB are connected, they're not properly cleaned up before `os.Exit()`.

**Current Code (lines 131-143):**
```go
if err := redisClient.Ping(pingCtx).Err(); err != nil {
    logger.Error("redis connection failed", "error", err)
    os.Exit(1) // db is open but not closed!
}
```

**Implementation Approach:**

Use a cleanup pattern with named returns or defer chain.

**Code Changes:**

```go
// cmd/bot/main.go - Refactor startBot()
func startBot(cfg *config.Config, logger *slog.Logger) {
    // Track resources for cleanup
    var cleanupFuncs []func()
    cleanup := func() {
        for i := len(cleanupFuncs) - 1; i >= 0; i-- {
            cleanupFuncs[i]()
        }
    }

    // Open database
    db, err := database.Open(cfg)
    if err != nil {
        logger.Error("db open failed", "error", err)
        os.Exit(1)
    }
    cleanupFuncs = append(cleanupFuncs, func() {
        sqlDB, _ := db.Gorm.DB()
        if sqlDB != nil {
            sqlDB.Close()
        }
    })

    // Initialize Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr: cfg.Redis.Addr,
    })
    cleanupFuncs = append(cleanupFuncs, func() {
        redisClient.Close()
    })

    // Test Redis connection
    pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := redisClient.Ping(pingCtx).Err(); err != nil {
        logger.Error("redis connection failed", "error", err)
        cleanup()
        os.Exit(1)
    }
    logger.Info("redis client initialized", "url", cfg.Redis.URL)

    // ... rest of initialization with similar pattern

    // On graceful shutdown
    defer cleanup()

    // ... signal handling and main loop
}
```

**Testing Requirements:**
- Manual test: Start bot with invalid Redis URL, verify no connection leaks
- Check resource monitoring during startup failures

---

### 2.4 UTF-8 String Truncation Bug

| Attribute | Value |
|-----------|-------|
| **Priority** | HIGH |
| **File(s)** | `internal/services/llm.go`, `internal/services/rag.go` |
| **Risk** | Low |
| **Estimated Time** | 20 minutes |
| **Dependencies** | None |

**Issue:** String truncation using byte slicing `s[:500]` can split multi-byte UTF-8 characters, causing corrupted output.

**Current Code (llm.go lines 358-361, rag.go lines 200-205):**
```go
if len(truncated) > 500 {
    truncated = truncated[:500] + "..."
}
```

**Implementation Approach:**

Use rune-aware truncation that respects UTF-8 boundaries.

**Code Changes:**

```go
// internal/services/llm.go - Add helper function (can also go in internal/utils/)
// truncateRunes safely truncates a string to maxRunes runes without splitting multi-byte characters.
func truncateRunes(s string, maxRunes int) string {
    if maxRunes <= 0 {
        return ""
    }

    runes := []rune(s)
    if len(runes) <= maxRunes {
        return s
    }

    return string(runes[:maxRunes]) + "..."
}

// Update buildPrompt (line 359)
truncated := truncateRunes(sanitizePromptInput(ctx), 500)
// Remove the separate length check since truncateRunes handles it

// internal/services/rag.go - Update truncateString (line 199)
func truncateString(s string, maxRunes int) string {
    runes := []rune(s)
    if len(runes) <= maxRunes {
        return s
    }
    return string(runes[:maxRunes]) + "..."
}
```

**Testing Requirements:**
- Add test: `TestTruncateRunes_MultiByte` - test with Japanese, emoji, Cyrillic
- Add test: `TestTruncateRunes_ASCII`
- Add test: `TestTruncateRunes_EdgeCases`

---

## Phase 3: Enhancements (Next Sprint)

### 3.1 Missing Config Validation

| Attribute | Value |
|-----------|-------|
| **Priority** | MEDIUM |
| **File(s)** | `internal/config/config.go` |
| **Risk** | Low |
| **Estimated Time** | 45 minutes |
| **Dependencies** | None |

**Issue:** Config struct uses `required` tags but lacks semantic validation (e.g., port ranges, URL formats, positive integers).

**Implementation Approach:**

Add a `Validate()` method with comprehensive checks.

**Code Changes:**

```go
// internal/config/config.go - Add after Load()

import (
    "errors"
    "net/url"
    // ... existing imports
)

// Validate performs semantic validation on the configuration.
func (c *Config) Validate() error {
    var errs []error

    // Discord validation
    if len(c.Discord.Token) < 50 {
        errs = append(errs, errors.New("DISCORD_TOKEN appears invalid (too short)"))
    }
    if c.Discord.GuildID != "" && len(c.Discord.GuildID) < 17 {
        errs = append(errs, errors.New("DISCORD_GUILD_ID appears invalid (too short)"))
    }

    // Database validation
    if c.Database.Port < 1 || c.Database.Port > 65535 {
        errs = append(errs, fmt.Errorf("DB_PORT must be 1-65535, got %d", c.Database.Port))
    }

    // URL validation
    for _, urlCfg := range []struct {
        name string
        val  string
    }{
        {"OLLAMA_URL", c.Ollama.URL},
        {"CHROMA_URL", c.Chroma.URL},
    } {
        if _, err := url.Parse(urlCfg.val); err != nil {
            errs = append(errs, fmt.Errorf("%s is not a valid URL: %w", urlCfg.name, err))
        }
    }

    // Numeric range validation
    if c.Bot.RateLimitPerMin < 1 || c.Bot.RateLimitPerMin > 100 {
        errs = append(errs, fmt.Errorf("RATE_LIMIT_PER_MINUTE must be 1-100, got %d", c.Bot.RateLimitPerMin))
    }
    if c.Ollama.RequestTimeout < 10 || c.Ollama.RequestTimeout > 300 {
        errs = append(errs, fmt.Errorf("OLLAMA_TIMEOUT must be 10-300, got %d", c.Ollama.RequestTimeout))
    }

    // LLM token validation
    if c.LLM.FastMaxTokens < 10 {
        errs = append(errs, errors.New("LLM_FAST_MAX_TOKENS must be >= 10"))
    }
    if c.LLM.StandardMaxTokens < c.LLM.FastMaxTokens {
        errs = append(errs, errors.New("LLM_STANDARD_MAX_TOKENS must be >= LLM_FAST_MAX_TOKENS"))
    }
    if c.LLM.DeepMaxTokens < c.LLM.StandardMaxTokens {
        errs = append(errs, errors.New("LLM_DEEP_MAX_TOKENS must be >= LLM_STANDARD_MAX_TOKENS"))
    }

    // Temperature validation (0.0 to 2.0 is typical range)
    for _, t := range []float64{c.LLM.FastTemperature, c.LLM.StandardTemperature, c.LLM.DeepTemperature} {
        if t < 0.0 || t > 2.0 {
            errs = append(errs, fmt.Errorf("LLM temperature must be 0.0-2.0, got %f", t))
        }
    }

    // File existence validation
    if _, err := os.Stat(c.Bot.PersonalityFile); os.IsNotExist(err) {
        errs = append(errs, fmt.Errorf("PERSONALITY_FILE does not exist: %s", c.Bot.PersonalityFile))
    }

    if len(errs) > 0 {
        return errors.Join(errs...)
    }
    return nil
}

// Update Load() to call Validate()
func Load() (*Config, error) {
    var cfg Config
    if err := envconfig.Process("", &cfg); err != nil {
        return nil, err
    }

    // ... existing post-processing

    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("configuration validation failed: %w", err)
    }

    return &cfg, nil
}
```

**Testing Requirements:**
- Add test: `TestConfig_Validate_HappyPath`
- Add test: `TestConfig_Validate_InvalidPort`
- Add test: `TestConfig_Validate_InvalidURL`
- Add test: `TestConfig_Validate_InvalidTokens`

---

### 3.2 Add Circuit Breaker for Ollama

| Attribute | Value |
|-----------|-------|
| **Priority** | MEDIUM |
| **File(s)** | `pkg/ollama/client.go`, `internal/services/llm.go` |
| **Risk** | Medium - Adds dependency |
| **Estimated Time** | 1.5 hours |
| **Dependencies** | None |

**Issue:** When Ollama is down, every request fails immediately but still tries to connect. This wastes resources and could cascade failures.

**Implementation Approach:**

Wrap Ollama client with `github.com/sony/gobreaker` circuit breaker.

**Code Changes:**

```go
// pkg/ollama/client.go

import (
    "github.com/sony/gobreaker"
    // ... existing imports
)

type Client struct {
    baseURL    string
    httpClient *http.Client
    breaker    *gobreaker.CircuitBreaker
}

func NewClientWithTimeout(baseURL string, timeout time.Duration) *Client {
    // Configure circuit breaker
    // Opens after 5 consecutive failures, half-opens after 30 seconds
    cbSettings := gobreaker.Settings{
        Name:        "ollama",
        MaxRequests: 3,                // Allow 3 requests in half-open state
        Interval:    60 * time.Second, // Clear counts after 60s
        Timeout:     30 * time.Second, // Try again after 30s in open state
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            // Open circuit after 5 consecutive failures
            return counts.ConsecutiveFailures >= 5
        },
        OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
            // This will be logged via the calling service
        },
    }

    return &Client{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: timeout,
        },
        breaker: gobreaker.NewCircuitBreaker(cbSettings),
    }
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
    result, err := c.breaker.Execute(func() (interface{}, error) {
        return c.doGenerate(ctx, req)
    })
    if err != nil {
        return nil, err
    }
    return result.(*GenerateResponse), nil
}

// doGenerate is the actual implementation (move existing code here)
func (c *Client) doGenerate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
    req.Stream = false
    // ... existing implementation
}

// Similar wrapper for Embed method
func (c *Client) Embed(ctx context.Context, model, text string) ([]float32, error) {
    result, err := c.breaker.Execute(func() (interface{}, error) {
        return c.doEmbed(ctx, model, text)
    })
    if err != nil {
        return nil, err
    }
    return result.([]float32), nil
}

// GetBreakerState returns the current circuit breaker state for monitoring
func (c *Client) GetBreakerState() gobreaker.State {
    return c.breaker.State()
}
```

**Add dependency:**
```bash
go get github.com/sony/gobreaker@v0.6.0
```

**Testing Requirements:**
- Add test: `TestClient_CircuitBreaker_Opens`
- Add test: `TestClient_CircuitBreaker_HalfOpen`
- Add test: `TestClient_CircuitBreaker_Closes`
- Integration test: Kill Ollama during test, verify circuit opens

---

## Implementation Order & Dependencies

```
Phase 1 (Critical - Day 1):
├── 1.1 Race Condition (no deps)
├── 1.2 Rate Limiting (no deps)
├── 1.3 Context Timeout (no deps)
├── 1.4 DB Connection Leaks (no deps)
├── 1.5 Nil Pointer Risk (no deps)
└── 1.6 Ollama Error Handling (no deps)

Phase 2 (High - Days 2-3):
├── 2.1 Discord ID Type (after Phase 1, requires maintenance window)
├── 2.2 Prompt Injection (no deps)
├── 2.3 Resource Cleanup (no deps)
└── 2.4 UTF-8 Truncation (no deps)

Phase 3 (Enhancements - Next Sprint):
├── 3.1 Config Validation (no deps)
└── 3.2 Circuit Breaker (no deps, adds go.mod dependency)
```

---

## Testing Strategy Summary

### Unit Tests to Add

| Test File | Tests to Add |
|-----------|-------------|
| `internal/services/rag_test.go` | `TestRAGService_ConcurrentQuery`, `TestRAGService_RaceCondition` |
| `internal/bot/commands_test.go` (NEW) | `TestHandleAskCommand_RateLimited`, `TestGetUserFromInteraction_*` |
| `internal/services/llm_test.go` | `TestSanitizePromptInput_*`, `TestTruncateRunes_*` |
| `internal/config/config_test.go` (NEW) | `TestConfig_Validate_*` |
| `pkg/ollama/client_test.go` (NEW) | `TestFormatHTTPError_*`, `TestClient_CircuitBreaker_*` |

### Integration Tests

1. Run full test suite with `-race` flag after Phase 1
2. Test database migration on staging DB
3. Load test `/ask` command with rate limiting

### Manual Testing Checklist

- [ ] Verify rate limiting in Discord (spam `/ask` 10 times)
- [ ] Test `/link` command in DM (without guild context)
- [ ] Test prompt injection via `/ask`
- [ ] Verify graceful shutdown (Ctrl+C during active request)
- [ ] Test CLI commands (`migrate`, `index-docs`) for connection leaks
- [ ] Test with non-ASCII usernames (Unicode)

---

## Risk Assessment

| Fix | Risk Level | Mitigation |
|-----|------------|------------|
| Discord ID Migration | **HIGH** | Test on staging first, schedule maintenance window |
| Race Condition Fix | **MEDIUM** | Thorough testing with `-race`, load testing |
| Circuit Breaker | **MEDIUM** | Feature flag, gradual rollout |
| Rate Limiting | **LOW** | Fail-open on Redis errors |
| All Others | **LOW** | Standard testing |

---

## Rollback Plan

### Phase 1 Rollback
- All changes are additive, can be reverted via git
- No database changes

### Phase 2 Rollback
- **Discord ID Migration**: Down migration available, but may fail if new IDs stored
- Backup database before migration
- Keep old code branch for 1 week

### Phase 3 Rollback
- Circuit breaker: Can be disabled via config flag
- Config validation: Remove call to `Validate()` in `Load()`

---

## Linear Issues to Create

After approval, create the following Linear issues:

1. `[CRITICAL] Fix race condition in RAGService` - Phase 1.1
2. `[CRITICAL] Enforce rate limiting on /ask command` - Phase 1.2
3. `[CRITICAL] Fix context timeout hierarchy` - Phase 1.3
4. `[CRITICAL] Fix database connection leaks in CLI` - Phase 1.4
5. `[CRITICAL] Add nil checks for Discord interactions` - Phase 1.5
6. `[CRITICAL] Improve Ollama error messages` - Phase 1.6
7. `[HIGH] Migrate Discord ID to string type` - Phase 2.1
8. `[HIGH] Add prompt injection protection` - Phase 2.2
9. `[HIGH] Fix resource cleanup on startup failure` - Phase 2.3
10. `[HIGH] Fix UTF-8 string truncation` - Phase 2.4
11. `[MEDIUM] Add config validation` - Phase 3.1
12. `[MEDIUM] Add circuit breaker for Ollama` - Phase 3.2

---

## Appendix: File Change Summary

| File | Phase | Changes |
|------|-------|---------|
| `internal/services/rag.go` | 1.1, 2.4 | Add mutex, fix truncation |
| `internal/bot/commands.go` | 1.2, 1.3, 1.5 | Rate limiting, context fix, nil checks |
| `internal/bot/bot.go` | 1.2 | Pass rate limiter to handlers |
| `cmd/bot/main.go` | 1.4, 2.3 | Add defer close, cleanup pattern |
| `pkg/ollama/client.go` | 1.6, 3.2 | Error handling, circuit breaker |
| `internal/services/llm.go` | 2.2, 2.4 | Sanitization, truncation |
| `internal/database/models/user.go` | 2.1 | Change DiscordID type |
| `internal/services/account.go` | 2.1 | Update parameter type |
| `internal/config/config.go` | 3.1 | Add Validate() method |
| `migrations/0004_*.sql` (NEW) | 2.1 | Discord ID migration |

---

**Document Version:** 1.0  
**Author:** Code Review Implementation Planning  
**Status:** Ready for Review  
**Next Action:** Create Linear issues and begin Phase 1
