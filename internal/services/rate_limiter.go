package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lua script for atomic INCR and EXPIRE
const rateLimitScript = `
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('TTL', KEYS[1])
return {count, ttl}
`

// RateLimiter provides per-user rate limiting using Redis.
type RateLimiter struct {
	client         *redis.Client
	requestsPerMin int
	logger         *slog.Logger
	script         *redis.Script
}

// NewRateLimiter initializes a rate limiter with Redis client.
func NewRateLimiter(redisClient *redis.Client, requestsPerMin int, logger *slog.Logger) *RateLimiter {
	return &RateLimiter{
		client:         redisClient,
		requestsPerMin: requestsPerMin,
		logger:         logger,
		script:         redis.NewScript(rateLimitScript),
	}
}

// IsAllowed checks if a user is within their rate limit.
// Returns true if allowed, false if rate limit exceeded.
// Also returns the number of remaining requests and time until reset.
func (r *RateLimiter) IsAllowed(ctx context.Context, userID string) (bool, int, time.Duration, error) {
	key := fmt.Sprintf("rate_limit:%s", userID)

	// Execute atomic INCR + EXPIRE via Lua script
	result, err := r.script.Run(ctx, r.client, []string{key}, 60).Result()
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to execute rate limit script: %w", err)
	}

	// Parse result: [count, ttl]
	values, ok := result.([]interface{})
	if !ok || len(values) != 2 {
		return false, 0, 0, fmt.Errorf("unexpected rate limit script result: %v", result)
	}

	count, ok := values[0].(int64)
	if !ok {
		return false, 0, 0, fmt.Errorf("unexpected count type: %T", values[0])
	}

	ttlSeconds, ok := values[1].(int64)
	if !ok {
		return false, 0, 0, fmt.Errorf("unexpected ttl type: %T", values[1])
	}

	ttl := time.Duration(ttlSeconds) * time.Second
	if ttl < 0 {
		ttl = time.Minute
	}

	allowed := count <= int64(r.requestsPerMin)
	remaining := int(r.requestsPerMin) - int(count)
	if remaining < 0 {
		remaining = 0
	}

	if !allowed {
		r.logger.Warn("rate limit exceeded",
			"user_id", userID,
			"requests", count,
			"limit", r.requestsPerMin,
		)
	}

	return allowed, remaining, ttl, nil
}

// Reset clears the rate limit counter for a user (useful for testing).
func (r *RateLimiter) Reset(ctx context.Context, userID string) error {
	key := fmt.Sprintf("rate_limit:%s", userID)
	return r.client.Del(ctx, key).Err()
}

// GetCount returns the current request count for a user.
func (r *RateLimiter) GetCount(ctx context.Context, userID string) (int, error) {
	key := fmt.Sprintf("rate_limit:%s", userID)

	count, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil // Key doesn't exist
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get rate limit count: %w", err)
	}

	return count, nil
}
