# Critical Fixes Implementation Plan

## Overview
This document outlines the critical issues found during code review and the implementation plan to fix them.

## Phase 1: Immediate Fixes (Do First)
These are critical issues affecting production stability and correctness.

### Issue 1: Fix race condition in RAGService
**File:** `internal/services/rag.go`
**Problem:** The `collectionID` field is accessed concurrently without synchronization. Multiple goroutines may call `ensureCollection()` simultaneously, leading to race conditions.
**Solution:** Add `sync.RWMutex` to protect `collectionID` field
**Implementation:**
- Add mutex field to RAGService struct
- Lock reads/writes to collectionID
- Ensure quick release of locks

### Issue 2: Add rate limiting to /ask command
**File:** `internal/bot/commands.go`
**Problem:** The /ask command doesn't implement rate limiting, allowing users to spam LLM requests and exhaust resources.
**Solution:** Check rate limiter before processing ask command
**Implementation:**
- Get user ID from interaction
- Call RateLimiter.IsAllowed() before defer
- Return 429-like ephemeral message if rate limited
- Log rate limit events

### Issue 3: Fix context timeout hierarchy
**File:** `internal/bot/commands.go`
**Problem:** RAG sub-context (5s) is created with parent context timeout instead of background, causing premature cancellation if parent has earlier deadline.
**Solution:** Ensure proper context hierarchy with fallback timeout
**Implementation:**
- Set RAG timeout correctly from background if needed
- Log timeout violations
- Document timeout strategy clearly

### Issue 4: Close database connections in CLI commands
**File:** `cmd/bot/main.go`
**Problem:** `handleMigrate()` and `handleIndexDocs()` don't close the database connection after use, causing resource leaks.
**Solution:** Properly close database connection
**Implementation:**
- Get sql.DB from GORM
- Defer Close() in both CLI functions

### Issue 5: Add nil checks for Discord user fields
**File:** `internal/bot/commands.go`
**Problem:** Code accesses `i.Member.User.Username` without checking if User is nil, and `i.User.Username` without nil check on i.User.
**Solution:** Add defensive nil checks
**Implementation:**
- Check i.Member != nil before accessing i.Member.User
- Check i.Member.User != nil before accessing fields
- Check i.User != nil before accessing fields
- Use fallback username if all are nil

### Issue 6: Improve Ollama error messages with response body
**File:** `pkg/ollama/client.go`
**Problem:** When Ollama returns errors, the response body is ignored, making debugging difficult.
**Solution:** Read and include response body in error messages
**Implementation:**
- Read response body on error status codes
- Include in error message for debugging
- Handle very large response bodies (truncate to 500 chars)

## Phase 2: High Priority Fixes (This Week)
These improve data integrity and security.

### Issue 7: Change Discord IDs from int64 to string
**File:** `internal/database/models/user.go` and related
**Problem:** Discord IDs are stored as int64, but discordgo v0.29.0 provides them as strings. Conversion is fragile and error-prone.
**Solution:** Change database schema and model to use VARCHAR(20)
**Implementation:**
- Create migration file: `migrations/002_change_discord_id_to_string.sql`
- Update User model to use string for DiscordID
- Update DiscordUsername handling (it's already string)
- Update all code that reads/writes DiscordID
- Update queries that filter by DiscordID

### Issue 8: Add prompt injection sanitization
**File:** `internal/services/llm.go`
**Problem:** User input is directly included in the prompt without sanitization, allowing prompt injection attacks.
**Solution:** Sanitize user input before including in prompt
**Implementation:**
- Create sanitization function
- Escape special characters and prompt delimiters
- Apply in buildPrompt() method
- Add tests for injection attempts

### Issue 9: Fix resource cleanup on startup errors
**File:** `cmd/bot/main.go`
**Problem:** If initialization fails (e.g., RAG service fails), resources like database connections and Redis aren't closed.
**Solution:** Use defer for cleanup or proper error handling
**Implementation:**
- Cleanup DB on initialization errors
- Cleanup Redis on initialization errors
- Log cleanup errors
- Ensure graceful failure

### Issue 10: Fix UTF-8 truncation using runes
**File:** `internal/services/rag.go`
**Problem:** `truncateString()` truncates by byte count, which can split multi-byte UTF-8 characters, corrupting text.
**Solution:** Truncate by rune count to preserve UTF-8 integrity
**Implementation:**
- Convert string to runes
- Truncate rune slice
- Convert back to string
- Add tests for multi-byte characters

### Issue 11: Add config validation with Validate() method
**File:** `internal/config/config.go`
**Problem:** Invalid configurations can cause cryptic failures deep in the application.
**Solution:** Add comprehensive config validation on load
**Implementation:**
- Add Validate() method to Config
- Check port ranges
- Check durations/timeouts are reasonable
- Check model names aren't empty
- Call Validate() in Load()
- Return clear error messages

## Phase 3: Enhancements (Nice to Have)
These improve resilience and observability.

### Issue 12: Add circuit breaker for Ollama
**File:** `internal/services/llm.go` or new `pkg/circuit/`
**Problem:** Transient Ollama failures cause immediate failures instead of graceful degradation.
**Solution:** Add circuit breaker pattern
**Implementation:**
- Use github.com/sony/gobreaker
- Wrap Ollama client calls
- Define failure threshold and recovery timeout
- Return fallback message when circuit is open
- Add metrics for circuit state changes

### Issue 13: Add Prometheus metrics (if time permits)
**File:** TBD
**Problem:** No observability into bot performance, error rates, latency
**Solution:** Add Prometheus metrics
**Implementation:**
- Track LLM generation latency (histogram)
- Track RAG query latency (histogram)
- Track rate limit events (counter)
- Track Discord API errors (counter)
- Export metrics on /metrics endpoint

## Testing Requirements

### Unit Tests
- Race condition fixes (run with `-race` flag)
- Rate limiter behavior at limits
- Config validation edge cases
- UTF-8 truncation with multi-byte characters
- Prompt injection sanitization

### Integration Tests
- RAG service concurrent access
- Database connection cleanup
- Config loading with various inputs
- Ollama error handling with actual error responses

### Manual Testing
- /ask command rate limiting behavior
- /link command with nil user fields
- Discord timeout behavior (simulate slow Ollama)
- Database migration (test rollback)

## Database Migration Strategy

### Migration File: `002_change_discord_id_to_string.sql`
1. Create new column `discord_id_string` as VARCHAR(20)
2. Copy/convert data from `discord_id` to `discord_id_string`
3. Drop old `discord_id` unique index
4. Rename column
5. Recreate unique index on new column
6. Ensure NOT NULL constraint

### Rollback Procedure
Same steps in reverse (test locally first!)

## Risk Mitigation

1. **Race conditions:** Test with `-race` flag
2. **Config validation:** Test with invalid configs
3. **UTF-8 handling:** Test with various Unicode strings
4. **Database migration:** Test migration and rollback locally
5. **Discord ID changes:** Gradual deployment, monitoring logs

## Metrics

- Estimated Phase 1 effort: 4 hours
- Estimated Phase 2 effort: 6 hours
- Estimated Phase 3 effort: 4 hours (if included)
- Total: 14 hours (Phase 1+2)

## Success Criteria

✅ All code compiles with no warnings
✅ All unit tests pass
✅ Integration tests pass
✅ No race conditions detected (go run -race)
✅ Bot handles errors gracefully
✅ Database migration works and is reversible
✅ Rate limiting prevents abuse
✅ Config validation prevents misconfiguration
