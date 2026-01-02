package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// RATE LIMITER - Low Level Design Implementation
// ============================================================================
//
// What is a Rate Limiter?
// -----------------------
// A rate limiter controls the number of requests a user/client can make to a
// service within a specific time window. It helps:
// - Prevent abuse and denial-of-service attacks
// - Ensure fair resource usage among users
// - Protect servers from being overwhelmed
//
// This file implements 4 popular rate limiting algorithms:
// 1. Token Bucket     - Allows burst traffic, most widely used
// 2. Sliding Window   - Smooth limiting, no boundary issues
// 3. Fixed Window     - Simple, but has boundary problems
// 4. Leaky Bucket     - Processes requests at constant rate
//
// Design Pattern Used: Strategy Pattern
// - RateLimiter interface defines the contract
// - Each algorithm implements this interface
// - API Gateway uses any limiter via the interface
//
// ============================================================================

// ============================================================================
// SECTION 1: RATE LIMITER INTERFACE (Strategy Pattern)
// ============================================================================

// RateLimiter defines the contract for all rate limiting algorithms.
// Any rate limiter must implement these two methods.
type RateLimiter interface {
	// Allow checks if a request from the given userID should be permitted.
	// Returns true if allowed, false if rate limited.
	Allow(userID string) bool

	// GetName returns the name of the algorithm for logging purposes.
	GetName() string
}

// ============================================================================
// SECTION 2: TOKEN BUCKET ALGORITHM
// ============================================================================
//
// How Token Bucket Works:
// -----------------------
// Imagine a bucket that holds tokens. Each request costs 1 token.
// - Bucket has a maximum capacity (e.g., 10 tokens)
// - Tokens are added at a fixed rate (e.g., 2 tokens per second)
// - When a request arrives, if there's a token, consume it and allow
// - If bucket is empty, reject the request
// - Allows burst traffic (up to bucket capacity) followed by steady rate
//
// Example: capacity=5, refillRate=2/sec
// - User can make 5 quick requests (burst)
// - Then must wait for tokens to refill (2 per second)
//
// Pros: Allows controlled bursts, smooth refill
// Cons: Slightly more complex than fixed window
//
// ============================================================================

// TokenBucket represents a single user's token bucket.
type TokenBucket struct {
	maxCapacity     int           // Maximum tokens the bucket can hold
	currentTokens   int           // Current number of tokens available
	tokensPerRefill int           // How many tokens to add per refill interval
	refillInterval  time.Duration // How often tokens are refilled
	lastRefillTime  time.Time     // When tokens were last refilled
	mutex           sync.Mutex    // Protects concurrent access to this bucket
}

// NewTokenBucket creates a new token bucket with the specified configuration.
func NewTokenBucket(maxCapacity, tokensPerRefill int, refillInterval time.Duration) *TokenBucket {
	return &TokenBucket{
		maxCapacity:     maxCapacity,
		currentTokens:   maxCapacity, // Start with a full bucket
		tokensPerRefill: tokensPerRefill,
		refillInterval:  refillInterval,
		lastRefillTime:  time.Now(),
	}
}

// refillTokens adds tokens based on elapsed time since last refill.
// This is called internally before checking/consuming tokens.
func (bucket *TokenBucket) refillTokens() {
	currentTime := time.Now()
	timeSinceLastRefill := currentTime.Sub(bucket.lastRefillTime)

	// Calculate how many refill intervals have passed
	intervalsPassed := int(timeSinceLastRefill / bucket.refillInterval)
	tokensToAdd := intervalsPassed * bucket.tokensPerRefill

	if tokensToAdd > 0 {
		// Add tokens but don't exceed capacity
		bucket.currentTokens = min(bucket.maxCapacity, bucket.currentTokens+tokensToAdd)
		bucket.lastRefillTime = currentTime
	}
}

// TryConsume attempts to consume one token. Returns true if successful.
func (bucket *TokenBucket) TryConsume() bool {
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	// First, refill tokens based on elapsed time
	bucket.refillTokens()

	// Check if we have tokens available
	if bucket.currentTokens > 0 {
		bucket.currentTokens--
		return true
	}
	return false
}

// GetAvailableTokens returns the current number of available tokens.
func (bucket *TokenBucket) GetAvailableTokens() int {
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()
	bucket.refillTokens()
	return bucket.currentTokens
}

// TokenBucketRateLimiter manages token buckets for multiple users.
type TokenBucketRateLimiter struct {
	userBuckets     map[string]*TokenBucket // Map of userID -> their bucket
	maxCapacity     int                     // Bucket capacity for new users
	tokensPerRefill int                     // Refill rate for new users
	refillInterval  time.Duration           // Refill interval for new users
	mutex           sync.RWMutex            // Protects the userBuckets map
}

// NewTokenBucketRateLimiter creates a new token bucket rate limiter.
func NewTokenBucketRateLimiter(maxCapacity, tokensPerRefill int, refillInterval time.Duration) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		userBuckets:     make(map[string]*TokenBucket),
		maxCapacity:     maxCapacity,
		tokensPerRefill: tokensPerRefill,
		refillInterval:  refillInterval,
	}
}

// getOrCreateBucket retrieves an existing bucket or creates a new one for the user.
// Uses double-checked locking pattern for thread safety and efficiency.
func (limiter *TokenBucketRateLimiter) getOrCreateBucket(userID string) *TokenBucket {
	// First, try to read with a read lock (allows concurrent reads)
	limiter.mutex.RLock()
	bucket, exists := limiter.userBuckets[userID]
	limiter.mutex.RUnlock()

	if exists {
		return bucket
	}

	// Bucket doesn't exist, need to create it with write lock
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	// Double-check: another goroutine might have created it while we waited
	if bucket, exists = limiter.userBuckets[userID]; exists {
		return bucket
	}

	// Create new bucket for this user
	bucket = NewTokenBucket(limiter.maxCapacity, limiter.tokensPerRefill, limiter.refillInterval)
	limiter.userBuckets[userID] = bucket
	return bucket
}

// Allow checks if a request from userID should be permitted.
// Implements the RateLimiter interface.
func (limiter *TokenBucketRateLimiter) Allow(userID string) bool {
	bucket := limiter.getOrCreateBucket(userID)
	return bucket.TryConsume()
}

// GetName returns the algorithm name.
func (limiter *TokenBucketRateLimiter) GetName() string {
	return "Token Bucket"
}

// ============================================================================
// SECTION 3: SLIDING WINDOW ALGORITHM
// ============================================================================
//
// How Sliding Window Works:
// -------------------------
// Track the timestamp of each request within a sliding time window.
// - Window slides with current time (not fixed boundaries)
// - Count requests in the last N seconds (where N is window size)
// - If count < limit, allow request; otherwise reject
//
// Example: 3 requests per 2 seconds
// - At time 0s: Request 1 ✅ (count: 1)
// - At time 0.5s: Request 2 ✅ (count: 2)
// - At time 1s: Request 3 ✅ (count: 3)
// - At time 1.5s: Request 4 ❌ (count: 3, limit reached)
// - At time 2.1s: Request 5 ✅ (Request 1 is now outside window)
//
// Pros: Smooth rate limiting, no boundary spike issues
// Cons: Higher memory usage (stores all request timestamps)
//
// ============================================================================

// SlidingWindowRecord stores request timestamps for one user.
type SlidingWindowRecord struct {
	requestTimestamps []time.Time // List of timestamps of recent requests
	mutex             sync.Mutex  // Protects concurrent access
}

// SlidingWindowRateLimiter implements sliding window rate limiting.
type SlidingWindowRateLimiter struct {
	userWindows    map[string]*SlidingWindowRecord // Map of userID -> their record
	maxRequests    int                             // Maximum requests allowed per window
	windowDuration time.Duration                   // Size of the sliding window
	mutex          sync.RWMutex                    // Protects the userWindows map
}

// NewSlidingWindowRateLimiter creates a new sliding window rate limiter.
func NewSlidingWindowRateLimiter(maxRequests int, windowDuration time.Duration) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		userWindows:    make(map[string]*SlidingWindowRecord),
		maxRequests:    maxRequests,
		windowDuration: windowDuration,
	}
}

// getOrCreateWindow retrieves or creates a sliding window record for a user.
func (limiter *SlidingWindowRateLimiter) getOrCreateWindow(userID string) *SlidingWindowRecord {
	limiter.mutex.RLock()
	window, exists := limiter.userWindows[userID]
	limiter.mutex.RUnlock()

	if exists {
		return window
	}

	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	// Double-check after acquiring write lock
	if window, exists = limiter.userWindows[userID]; exists {
		return window
	}

	window = &SlidingWindowRecord{
		requestTimestamps: make([]time.Time, 0),
	}
	limiter.userWindows[userID] = window
	return window
}

// Allow checks if a request from userID should be permitted.
func (limiter *SlidingWindowRateLimiter) Allow(userID string) bool {
	window := limiter.getOrCreateWindow(userID)
	window.mutex.Lock()
	defer window.mutex.Unlock()

	currentTime := time.Now()
	windowStartTime := currentTime.Add(-limiter.windowDuration)

	// Remove timestamps that are outside the current window (expired requests)
	validTimestamps := make([]time.Time, 0, len(window.requestTimestamps))
	for _, timestamp := range window.requestTimestamps {
		if timestamp.After(windowStartTime) {
			validTimestamps = append(validTimestamps, timestamp)
		}
	}
	window.requestTimestamps = validTimestamps

	// Check if we're under the limit
	if len(window.requestTimestamps) < limiter.maxRequests {
		window.requestTimestamps = append(window.requestTimestamps, currentTime)
		return true
	}
	return false
}

// GetName returns the algorithm name.
func (limiter *SlidingWindowRateLimiter) GetName() string {
	return "Sliding Window"
}

// ============================================================================
// SECTION 4: FIXED WINDOW ALGORITHM
// ============================================================================
//
// How Fixed Window Works:
// -----------------------
// Divide time into fixed windows (e.g., every minute) and count requests.
// - Reset counter at the start of each new window
// - Simple to implement with low memory overhead
//
// Example: 4 requests per 3-second window
// - Window 1 (0s-3s): Can make up to 4 requests
// - Window 2 (3s-6s): Counter resets, another 4 requests allowed
//
// IMPORTANT: Boundary Problem!
// - User makes 4 requests at 2.9s (end of window 1)
// - User makes 4 requests at 3.1s (start of window 2)
// - Result: 8 requests in 0.2 seconds! This exceeds the intended rate.
//
// Pros: Very simple, low memory usage
// Cons: Vulnerable to boundary spikes (traffic bursts at window edges)
//
// ============================================================================

// FixedWindowRecord stores request count for one user's current window.
type FixedWindowRecord struct {
	requestCount    int        // Number of requests in current window
	windowStartTime time.Time  // When the current window started
	mutex           sync.Mutex // Protects concurrent access
}

// FixedWindowRateLimiter implements fixed window rate limiting.
type FixedWindowRateLimiter struct {
	userWindows    map[string]*FixedWindowRecord // Map of userID -> their record
	maxRequests    int                           // Maximum requests per window
	windowDuration time.Duration                 // Duration of each window
	mutex          sync.RWMutex                  // Protects the userWindows map
}

// NewFixedWindowRateLimiter creates a new fixed window rate limiter.
func NewFixedWindowRateLimiter(maxRequests int, windowDuration time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		userWindows:    make(map[string]*FixedWindowRecord),
		maxRequests:    maxRequests,
		windowDuration: windowDuration,
	}
}

// getOrCreateWindow retrieves or creates a fixed window record for a user.
func (limiter *FixedWindowRateLimiter) getOrCreateWindow(userID string) *FixedWindowRecord {
	limiter.mutex.RLock()
	window, exists := limiter.userWindows[userID]
	limiter.mutex.RUnlock()

	if exists {
		return window
	}

	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	// Double-check after acquiring write lock
	if window, exists = limiter.userWindows[userID]; exists {
		return window
	}

	window = &FixedWindowRecord{
		windowStartTime: time.Now(),
	}
	limiter.userWindows[userID] = window
	return window
}

// Allow checks if a request from userID should be permitted.
func (limiter *FixedWindowRateLimiter) Allow(userID string) bool {
	window := limiter.getOrCreateWindow(userID)
	window.mutex.Lock()
	defer window.mutex.Unlock()

	currentTime := time.Now()

	// Check if we've moved to a new window
	timeSinceWindowStart := currentTime.Sub(window.windowStartTime)
	if timeSinceWindowStart >= limiter.windowDuration {
		// Start a new window: reset counter and update start time
		window.requestCount = 0
		window.windowStartTime = currentTime
	}

	// Check if we're under the limit
	if window.requestCount < limiter.maxRequests {
		window.requestCount++
		return true
	}
	return false
}

// GetName returns the algorithm name.
func (limiter *FixedWindowRateLimiter) GetName() string {
	return "Fixed Window"
}

// ============================================================================
// SECTION 5: LEAKY BUCKET ALGORITHM
// ============================================================================
//
// How Leaky Bucket Works:
// -----------------------
// Imagine a bucket with a hole at the bottom. Water (requests) drip out at
// a constant rate.
// - Incoming requests fill the bucket
// - Requests "leak" (are processed) at a constant rate
// - If bucket overflows, new requests are rejected
//
// Key Difference from Token Bucket:
// - Token Bucket: Tokens accumulate, allowing bursts
// - Leaky Bucket: Requests leak at constant rate, smoothing traffic
//
// Example: capacity=3, leakRate=1 per 500ms
// - Requests fill the bucket (up to 3)
// - Every 500ms, one "leaks out" (processed/removed)
// - If bucket is full, reject new requests until some leak out
//
// Pros: Smooths traffic, predictable output rate
// Cons: No bursting allowed, may delay requests
//
// ============================================================================

// LeakyBucketRecord represents one user's leaky bucket.
type LeakyBucketRecord struct {
	currentQueueSize int           // Current number of requests in the bucket
	maxCapacity      int           // Maximum requests the bucket can hold
	lastLeakTime     time.Time     // When requests last "leaked" out
	leakInterval     time.Duration // Time between each leak
	mutex            sync.Mutex    // Protects concurrent access
}

// LeakyBucketRateLimiter implements leaky bucket rate limiting.
type LeakyBucketRateLimiter struct {
	userBuckets  map[string]*LeakyBucketRecord // Map of userID -> their bucket
	maxCapacity  int                           // Bucket capacity for new users
	leakInterval time.Duration                 // Leak interval for new users
	mutex        sync.RWMutex                  // Protects the userBuckets map
}

// NewLeakyBucketRateLimiter creates a new leaky bucket rate limiter.
func NewLeakyBucketRateLimiter(maxCapacity int, leakInterval time.Duration) *LeakyBucketRateLimiter {
	return &LeakyBucketRateLimiter{
		userBuckets:  make(map[string]*LeakyBucketRecord),
		maxCapacity:  maxCapacity,
		leakInterval: leakInterval,
	}
}

// getOrCreateBucket retrieves or creates a leaky bucket for a user.
func (limiter *LeakyBucketRateLimiter) getOrCreateBucket(userID string) *LeakyBucketRecord {
	limiter.mutex.RLock()
	bucket, exists := limiter.userBuckets[userID]
	limiter.mutex.RUnlock()

	if exists {
		return bucket
	}

	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	// Double-check after acquiring write lock
	if bucket, exists = limiter.userBuckets[userID]; exists {
		return bucket
	}

	bucket = &LeakyBucketRecord{
		maxCapacity:  limiter.maxCapacity,
		leakInterval: limiter.leakInterval,
		lastLeakTime: time.Now(),
	}
	limiter.userBuckets[userID] = bucket
	return bucket
}

// Allow checks if a request from userID should be permitted.
func (limiter *LeakyBucketRateLimiter) Allow(userID string) bool {
	bucket := limiter.getOrCreateBucket(userID)
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	currentTime := time.Now()

	// Calculate how many requests have "leaked" out since last check
	timeSinceLastLeak := currentTime.Sub(bucket.lastLeakTime)
	leakedCount := int(timeSinceLastLeak / bucket.leakInterval)

	if leakedCount > 0 {
		// Remove leaked requests from the queue (but don't go below 0)
		bucket.currentQueueSize = max(0, bucket.currentQueueSize-leakedCount)
		bucket.lastLeakTime = currentTime
	}

	// Try to add the new request to the bucket
	if bucket.currentQueueSize < bucket.maxCapacity {
		bucket.currentQueueSize++
		return true
	}
	return false
}

// GetName returns the algorithm name.
func (limiter *LeakyBucketRateLimiter) GetName() string {
	return "Leaky Bucket"
}

// ============================================================================
// SECTION 6: API GATEWAY (Client that uses Rate Limiter)
// ============================================================================
//
// The API Gateway is a common component that sits between clients and backend
// services. It uses a rate limiter to protect the backend.
//
// This demonstrates the Strategy Pattern:
// - APIGateway depends on the RateLimiter interface, not concrete implementations
// - We can swap different rate limiting algorithms without changing APIGateway
//
// ============================================================================

// APIGateway handles incoming requests and applies rate limiting.
type APIGateway struct {
	rateLimiter RateLimiter // The rate limiting strategy (can be any algorithm)
}

// NewAPIGateway creates a new API Gateway with the specified rate limiter.
func NewAPIGateway(rateLimiter RateLimiter) *APIGateway {
	return &APIGateway{
		rateLimiter: rateLimiter,
	}
}

// HandleRequest processes an incoming request from a user.
// It first checks if the request is allowed by the rate limiter.
func (gateway *APIGateway) HandleRequest(userID string, endpoint string) {
	if gateway.rateLimiter.Allow(userID) {
		fmt.Printf("✅ [%s] Request ALLOWED for %s: %s\n",
			gateway.rateLimiter.GetName(), userID, endpoint)
	} else {
		fmt.Printf("❌ [%s] Request REJECTED for %s: %s (rate limited)\n",
			gateway.rateLimiter.GetName(), userID, endpoint)
	}
}

// ============================================================================
// SECTION 7: MAIN FUNCTION (Demo)
// ============================================================================

func main() {
	printSeparator()
	fmt.Println("        RATE LIMITER - Low Level Design Demo")
	printSeparator()

	// ----------------------------------------
	// Demo 1: Token Bucket Rate Limiter
	// ----------------------------------------
	fmt.Println("\n📊 Demo 1: TOKEN BUCKET LIMITER")
	fmt.Println("   Configuration: 5 tokens capacity, refill 2 tokens per second")
	fmt.Println("   Allows burst traffic up to bucket capacity")
	printLine()

	tokenBucketLimiter := NewTokenBucketRateLimiter(
		5,           // maxCapacity: bucket can hold 5 tokens
		2,           // tokensPerRefill: add 2 tokens per interval
		time.Second, // refillInterval: refill every 1 second
	)
	gateway1 := NewAPIGateway(tokenBucketLimiter)

	// Simulate burst of 7 rapid requests (only 5 should succeed)
	fmt.Println("\n   Sending burst of 7 rapid requests...")
	for i := 1; i <= 7; i++ {
		gateway1.HandleRequest("user1", fmt.Sprintf("/api/resource/%d", i))
	}

	// Wait for tokens to refill
	fmt.Println("\n   ⏳ Waiting 2 seconds for tokens to refill...")
	time.Sleep(2 * time.Second)

	// Try more requests after refill
	fmt.Println("\n   Sending 3 more requests after refill...")
	for i := 1; i <= 3; i++ {
		gateway1.HandleRequest("user1", fmt.Sprintf("/api/resource/%d", i))
	}

	// ----------------------------------------
	// Demo 2: Sliding Window Rate Limiter
	// ----------------------------------------
	fmt.Println("\n📊 Demo 2: SLIDING WINDOW LIMITER")
	fmt.Println("   Configuration: 3 requests per 2 seconds")
	fmt.Println("   Smooth limiting with no boundary issues")
	printLine()

	slidingWindowLimiter := NewSlidingWindowRateLimiter(
		3,             // maxRequests: allow 3 requests
		2*time.Second, // windowDuration: within 2 seconds
	)
	gateway2 := NewAPIGateway(slidingWindowLimiter)

	// Send requests with small delays
	fmt.Println("\n   Sending 5 requests with 300ms delays...")
	for i := 1; i <= 5; i++ {
		gateway2.HandleRequest("user2", fmt.Sprintf("/api/data/%d", i))
		time.Sleep(300 * time.Millisecond)
	}

	// ----------------------------------------
	// Demo 3: Fixed Window Rate Limiter
	// ----------------------------------------
	fmt.Println("\n📊 Demo 3: FIXED WINDOW LIMITER")
	fmt.Println("   Configuration: 4 requests per 3-second window")
	fmt.Println("   Simple but has boundary spike vulnerability")
	printLine()

	fixedWindowLimiter := NewFixedWindowRateLimiter(
		4,             // maxRequests: allow 4 requests
		3*time.Second, // windowDuration: per 3-second window
	)
	gateway3 := NewAPIGateway(fixedWindowLimiter)

	// Send rapid burst of requests
	fmt.Println("\n   Sending 6 rapid requests...")
	for i := 1; i <= 6; i++ {
		gateway3.HandleRequest("user3", fmt.Sprintf("/api/item/%d", i))
	}

	// ----------------------------------------
	// Demo 4: Leaky Bucket Rate Limiter
	// ----------------------------------------
	fmt.Println("\n📊 Demo 4: LEAKY BUCKET LIMITER")
	fmt.Println("   Configuration: 3 capacity, leak 1 request per 500ms")
	fmt.Println("   Smooths traffic to constant rate")
	printLine()

	leakyBucketLimiter := NewLeakyBucketRateLimiter(
		3,                    // maxCapacity: bucket holds 3 requests
		500*time.Millisecond, // leakInterval: process 1 request every 500ms
	)
	gateway4 := NewAPIGateway(leakyBucketLimiter)

	// Send rapid burst of requests
	fmt.Println("\n   Sending 5 rapid requests...")
	for i := 1; i <= 5; i++ {
		gateway4.HandleRequest("user4", fmt.Sprintf("/api/stream/%d", i))
	}

	// Wait for requests to leak out
	fmt.Println("\n   ⏳ Waiting 1 second for requests to leak out...")
	time.Sleep(1 * time.Second)

	// Try more requests after some leaked out
	fmt.Println("\n   Sending 3 more requests after leak...")
	for i := 1; i <= 3; i++ {
		gateway4.HandleRequest("user4", fmt.Sprintf("/api/stream/%d", i))
	}

	// ----------------------------------------
	// Summary: Algorithm Comparison
	// ----------------------------------------
	printSeparator()
	fmt.Println("        ALGORITHM COMPARISON SUMMARY")
	printSeparator()
	fmt.Println()
	fmt.Println("  ┌─────────────────┬──────────────────────────────────────────┐")
	fmt.Println("  │ Algorithm       │ Characteristics                          │")
	fmt.Println("  ├─────────────────┼──────────────────────────────────────────┤")
	fmt.Println("  │ Token Bucket    │ Allows bursts, smooth refill, most used  │")
	fmt.Println("  │ Sliding Window  │ Smooth limiting, no boundary issues      │")
	fmt.Println("  │ Fixed Window    │ Simple & fast, but has boundary problem  │")
	fmt.Println("  │ Leaky Bucket    │ Constant output rate, smooths traffic    │")
	fmt.Println("  └─────────────────┴──────────────────────────────────────────┘")
	fmt.Println()
	printSeparator()
}

// ============================================================================
// SECTION 8: HELPER FUNCTIONS
// ============================================================================

// printSeparator prints a visual separator line for better readability.
func printSeparator() {
	fmt.Println("═══════════════════════════════════════════════════════════════")
}

// printLine prints a thin line for section separation.
func printLine() {
	fmt.Println("───────────────────────────────────────────────────────────────")
}
