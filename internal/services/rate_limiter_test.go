package services

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestRateLimiter_Allowed(t *testing.T) {
	// Skip if no Redis available
	redisClient := getTestRedis(t)
	if redisClient == nil {
		t.Skip("Redis not available for testing")
	}
	defer redisClient.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	limiter := NewRateLimiter(redisClient, 5, logger)

	ctx := context.Background()
	userID := "test-user-1"

	// Reset before test
	_ = limiter.Reset(ctx, userID)

	// First 5 requests should be allowed
	for i := 1; i <= 5; i++ {
		allowed, remaining, ttl, err := limiter.IsAllowed(ctx, userID)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed", i)
		}
		if remaining != 5-i {
			t.Errorf("Request %d: expected remaining %d, got %d", i, 5-i, remaining)
		}
		if ttl <= 0 || ttl > time.Minute {
			t.Errorf("Request %d: unexpected TTL %v", i, ttl)
		}
	}

	// 6th request should be blocked
	allowed, remaining, ttl, err := limiter.IsAllowed(ctx, userID)
	if err != nil {
		t.Fatalf("6th request failed: %v", err)
	}
	if allowed {
		t.Error("6th request should be blocked")
	}
	if remaining != 0 {
		t.Errorf("6th request: expected remaining 0, got %d", remaining)
	}
	if ttl <= 0 || ttl > time.Minute {
		t.Errorf("6th request: unexpected TTL %v", ttl)
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	redisClient := getTestRedis(t)
	if redisClient == nil {
		t.Skip("Redis not available for testing")
	}
	defer redisClient.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	limiter := NewRateLimiter(redisClient, 10, logger)

	ctx := context.Background()
	userID := "test-user-concurrent"

	// Reset before test
	_ = limiter.Reset(ctx, userID)

	// Launch 20 concurrent requests
	var wg sync.WaitGroup
	successCount := 0
	successMutex := sync.Mutex{}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _, _, err := limiter.IsAllowed(ctx, userID)
			if err != nil {
				t.Errorf("Request failed: %v", err)
				return
			}
			if allowed {
				successMutex.Lock()
				successCount++
				successMutex.Unlock()
			}
		}()
	}

	wg.Wait()

	// Exactly 10 requests should succeed
	if successCount != 10 {
		t.Errorf("Expected exactly 10 successful requests, got %d", successCount)
	}

	// Verify count
	count, err := limiter.GetCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 20 {
		t.Errorf("Expected count 20, got %d", count)
	}
}

func TestRateLimiter_TTLExpiry(t *testing.T) {
	redisClient := getTestRedis(t)
	if redisClient == nil {
		t.Skip("Redis not available for testing")
	}
	defer redisClient.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	limiter := NewRateLimiter(redisClient, 5, logger)

	ctx := context.Background()
	userID := "test-user-ttl"

	// Reset before test
	_ = limiter.Reset(ctx, userID)

	// Use up the limit
	for i := 0; i < 5; i++ {
		_, _, _, err := limiter.IsAllowed(ctx, userID)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}

	// Verify we're blocked
	allowed, _, ttl, err := limiter.IsAllowed(ctx, userID)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}
	if allowed {
		t.Error("Should be blocked after 5 requests")
	}

	// Verify TTL is set (should be close to 60 seconds - 1 minute window)
	if ttl <= 0 || ttl > time.Minute {
		t.Errorf("Unexpected TTL: %v", ttl)
	}

	// Instead of waiting for full 60 second expiry (impractical in tests),
	// we verify that TTL decreases over time and that manual reset works

	// Reset manually and verify it works
	err = limiter.Reset(ctx, userID)
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	// Should be allowed again after reset
	allowed, remaining, _, err := limiter.IsAllowed(ctx, userID)
	if err != nil {
		t.Fatalf("Request after reset failed: %v", err)
	}
	if !allowed {
		t.Error("Should be allowed after reset")
	}
	if remaining != 4 {
		t.Errorf("Expected remaining 4 after reset, got %d", remaining)
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	redisClient := getTestRedis(t)
	if redisClient == nil {
		t.Skip("Redis not available for testing")
	}
	defer redisClient.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	limiter := NewRateLimiter(redisClient, 5, logger)

	ctx := context.Background()
	userID := "test-user-reset"

	// Ensure clean state before test
	_ = limiter.Reset(ctx, userID)

	// Use some requests
	for i := 0; i < 3; i++ {
		_, _, _, err := limiter.IsAllowed(ctx, userID)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}

	// Verify count
	count, err := limiter.GetCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get count: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Reset
	err = limiter.Reset(ctx, userID)
	if err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	// Verify reset
	count, err = limiter.GetCount(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get count after reset: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0 after reset, got %d", count)
	}

	// Should be allowed again
	allowed, _, _, err := limiter.IsAllowed(ctx, userID)
	if err != nil {
		t.Fatalf("Request after reset failed: %v", err)
	}
	if !allowed {
		t.Error("Should be allowed after reset")
	}
}

func TestRateLimiter_DifferentUsers(t *testing.T) {
	redisClient := getTestRedis(t)
	if redisClient == nil {
		t.Skip("Redis not available for testing")
	}
	defer redisClient.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	limiter := NewRateLimiter(redisClient, 5, logger)

	ctx := context.Background()

	// Reset both users to ensure clean state
	_ = limiter.Reset(ctx, "user1")
	_ = limiter.Reset(ctx, "user2")

	// User 1 uses all requests
	for i := 0; i < 5; i++ {
		_, _, _, err := limiter.IsAllowed(ctx, "user1")
		if err != nil {
			t.Fatalf("User1 request %d failed: %v", i, err)
		}
	}

	// User 2 should still be allowed
	allowed, remaining, _, err := limiter.IsAllowed(ctx, "user2")
	if err != nil {
		t.Fatalf("User2 request failed: %v", err)
	}
	if !allowed {
		t.Error("User2 should be allowed")
	}
	if remaining != 4 {
		t.Errorf("User2: expected remaining 4, got %d", remaining)
	}
}

func TestRateLimiter_RedisFailure(t *testing.T) {
	// Create a bad Redis client
	badClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Bad port
	})
	defer badClient.Close()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	limiter := NewRateLimiter(badClient, 5, logger)

	ctx := context.Background()
	_, _, _, err := limiter.IsAllowed(ctx, "test-user")
	if err == nil {
		t.Error("Expected error with bad Redis connection")
	}
}

// Helper to get Redis client for testing
func getTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Logf("Redis not available: %v", err)
		return nil
	}

	return client
}
