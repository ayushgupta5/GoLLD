package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// URL SHORTENER - Low Level Design
// ============================================================
//
// What is a URL Shortener?
// - A service that converts long URLs into short, easy-to-share links
// - Example: "https://example.com/very/long/path" → "https://short.ly/abc123"
//
// Key Concepts Covered:
// 1. Base62 Encoding - Convert numbers to short alphanumeric strings
// 2. Counter-based ID Generation - Ensures unique short codes
// 3. Analytics - Track how many times each link is clicked
// 4. URL Expiration - Links can have a time-to-live (TTL)
// 5. Thread Safety - Using mutexes for concurrent access
//
// ============================================================

// ========== CONSTANTS ==========

const (
	// Base62Chars contains all characters used for encoding short URLs
	// Using 62 characters (0-9, A-Z, a-z) gives us many possible combinations
	// With 7 characters: 62^7 = 3.5 trillion unique codes!
	Base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	// ShortCodeLength is the number of characters in generated short codes
	ShortCodeLength = 7

	// DefaultBaseDomain is used when no custom domain is provided
	DefaultBaseDomain = "https://short.url"

	// Constraints for custom short codes
	MinCustomCodeLength = 3
	MaxCustomCodeLength = 20
)

// ========== URL ENTRY ==========
// URLEntry represents a single shortened URL with all its metadata.
// Think of it as a "record" that stores everything about one short link.

type URLEntry struct {
	ShortCode   string     // The short code (e.g., "abc123")
	OriginalURL string     // The full/original URL that this code points to
	CreatedAt   time.Time  // When this short URL was created
	ExpiresAt   time.Time  // When this short URL will expire (zero means never)
	CreatedBy   string     // ID of the user who created this short URL
	IsCustom    bool       // True if user chose their own custom code
	ClickCount  int64      // How many times this short URL has been accessed
	LastAccess  time.Time  // When was this URL last accessed
	IsActive    bool       // False if the URL has been deleted/deactivated
	mutex       sync.Mutex // Protects concurrent access to mutable fields
}

// IsExpired checks if this short URL has passed its expiration time.
// Returns false if no expiration was set (ExpiresAt is zero).
func (entry *URLEntry) IsExpired() bool {
	// If expiration time was never set, the URL never expires
	if entry.ExpiresAt.IsZero() {
		return false
	}
	// Check if current time is after the expiration time
	return time.Now().After(entry.ExpiresAt)
}

// IncrementClicks safely increases the click count by 1.
// Uses atomic operation to be thread-safe without heavy locking.
func (entry *URLEntry) IncrementClicks() {
	// atomic.AddInt64 is thread-safe - multiple goroutines can call this safely
	atomic.AddInt64(&entry.ClickCount, 1)

	// Update last access time (requires mutex since time.Time isn't atomic)
	entry.mutex.Lock()
	entry.LastAccess = time.Now()
	entry.mutex.Unlock()
}

// GetClickCount safely retrieves the current click count.
// Uses atomic load to ensure we get a consistent value.
func (entry *URLEntry) GetClickCount() int64 {
	return atomic.LoadInt64(&entry.ClickCount)
}

// ========== CLICK ANALYTICS ==========
// ClickEvent records details about each time a short URL is accessed.
// This helps track usage patterns and provides insights.

type ClickEvent struct {
	ShortCode string    // Which short URL was clicked
	Timestamp time.Time // When the click happened
	IPAddress string    // IP address of the visitor (for geo-location)
	UserAgent string    // Browser/device info
	Referer   string    // Where the click came from (e.g., Twitter, email)
}

// Analytics stores and manages all click events.
// This is a simple in-memory implementation (production would use a database).
type Analytics struct {
	clickEvents []ClickEvent // List of all click events
	mutex       sync.Mutex   // Protects concurrent access
}

// NewAnalytics creates a new Analytics tracker.
func NewAnalytics() *Analytics {
	return &Analytics{
		clickEvents: make([]ClickEvent, 0), // Start with empty slice
	}
}

// RecordClick adds a new click event to the analytics.
func (analytics *Analytics) RecordClick(shortCode, ipAddress, userAgent, referer string) {
	analytics.mutex.Lock()
	defer analytics.mutex.Unlock()

	newClick := ClickEvent{
		ShortCode: shortCode,
		Timestamp: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referer:   referer,
	}
	analytics.clickEvents = append(analytics.clickEvents, newClick)
}

// GetClickCountByCode returns how many times a specific short URL was clicked.
func (analytics *Analytics) GetClickCountByCode(shortCode string) int {
	analytics.mutex.Lock()
	defer analytics.mutex.Unlock()

	count := 0
	for _, clickEvent := range analytics.clickEvents {
		if clickEvent.ShortCode == shortCode {
			count++
		}
	}
	return count
}

// ========== URL SHORTENER SERVICE ==========
// URLShortener is the main service that handles all URL shortening operations.
// It manages creating, resolving, and tracking short URLs.

type URLShortener struct {
	baseDomain       string               // Base domain for short URLs (e.g., "https://short.ly")
	urlDatabase      map[string]*URLEntry // Maps: shortCode -> URLEntry
	reverseLookup    map[string]string    // Maps: originalURL -> shortCode (for deduplication)
	idCounter        uint64               // Auto-incrementing counter for unique ID generation
	analyticsTracker *Analytics           // Tracks click events
	mutex            sync.RWMutex         // Read-Write mutex for thread-safe access
}

// NewURLShortener creates a new URL shortener service with the given domain.
// If no domain is provided, it uses the default domain.
func NewURLShortener(domain string) *URLShortener {
	if domain == "" {
		domain = DefaultBaseDomain
	}
	return &URLShortener{
		baseDomain:       domain,
		urlDatabase:      make(map[string]*URLEntry),
		reverseLookup:    make(map[string]string),
		analyticsTracker: NewAnalytics(),
	}
}

// encodeBase62 converts a number to a Base62 string.
// Base62 uses 0-9, A-Z, a-z (62 characters) to create short, URL-safe strings.
//
// Why Base62?
// - URLs are case-sensitive, so we can use both uppercase and lowercase letters
// - 62 chars means 62^7 = 3.5 trillion combinations with just 7 characters!
// - No special characters that need URL encoding
func (shortener *URLShortener) encodeBase62(number uint64) string {
	// Handle zero case
	if number == 0 {
		return string(Base62Chars[0])
	}

	// Build the encoded string by repeatedly dividing by 62
	// Similar to converting decimal to any other base
	result := ""
	for number > 0 {
		remainder := number % 62                         // Get the current digit (0-61)
		result = string(Base62Chars[remainder]) + result // Prepend the character
		number /= 62                                     // Move to next digit
	}

	// Pad with leading zeros to ensure minimum length for consistency
	for len(result) < ShortCodeLength {
		result = string(Base62Chars[0]) + result
	}

	return result
}

// generateUniqueShortCode creates a new unique short code using counter-based generation.
// This approach guarantees uniqueness because the counter always increases.
func (shortener *URLShortener) generateUniqueShortCode() string {
	// Atomically increment and get the new counter value
	// atomic.AddUint64 is thread-safe, so multiple goroutines won't get same ID
	newID := atomic.AddUint64(&shortener.idCounter, 1)
	return shortener.encodeBase62(newID)
}

// Shorten creates a short URL from a long URL.
// Parameters:
//   - originalURL: The long URL to shorten
//   - userID: ID of the user creating the short URL
//   - ttlDays: Time-to-live in days (0 means never expires)
//
// Returns:
//   - The complete short URL (e.g., "https://short.ly/abc123")
//   - An error if the URL is empty
func (shortener *URLShortener) Shorten(originalURL string, userID string, ttlDays int) (string, error) {
	// Validate input
	if originalURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	shortener.mutex.Lock()
	defer shortener.mutex.Unlock()

	// Check if this URL was already shortened (deduplication)
	// This prevents creating multiple short codes for the same URL
	if existingCode, alreadyExists := shortener.reverseLookup[originalURL]; alreadyExists {
		existingEntry := shortener.urlDatabase[existingCode]
		// Only return existing code if it's still active and not expired
		if existingEntry.IsActive && !existingEntry.IsExpired() {
			return shortener.baseDomain + "/" + existingCode, nil
		}
	}

	// Generate a new unique short code
	shortCode := shortener.generateUniqueShortCode()

	// Handle collision (very unlikely with counter-based generation, but good practice)
	for _, codeExists := shortener.urlDatabase[shortCode]; codeExists; {
		shortCode = shortener.generateUniqueShortCode()
	}

	// Create the URL entry with all metadata
	newEntry := &URLEntry{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
		CreatedBy:   userID,
		IsActive:    true,
	}

	// Set expiration if TTL was specified
	if ttlDays > 0 {
		newEntry.ExpiresAt = time.Now().AddDate(0, 0, ttlDays)
	}

	// Store in both maps
	shortener.urlDatabase[shortCode] = newEntry
	shortener.reverseLookup[originalURL] = shortCode

	return shortener.baseDomain + "/" + shortCode, nil
}

// ShortenCustom creates a short URL with a user-chosen custom code.
// This allows users to create memorable/branded short links.
// Parameters:
//   - originalURL: The long URL to shorten
//   - customCode: User's desired custom code (e.g., "mylink")
//   - userID: ID of the user creating the short URL
//
// Returns:
//   - The complete short URL (e.g., "https://short.ly/mylink")
//   - An error if validation fails or code is already taken
func (shortener *URLShortener) ShortenCustom(originalURL, customCode, userID string) (string, error) {
	// Validate inputs
	if originalURL == "" || customCode == "" {
		return "", fmt.Errorf("URL and custom code cannot be empty")
	}

	// Validate custom code length
	if len(customCode) < MinCustomCodeLength || len(customCode) > MaxCustomCodeLength {
		return "", fmt.Errorf("custom code must be %d-%d characters", MinCustomCodeLength, MaxCustomCodeLength)
	}

	shortener.mutex.Lock()
	defer shortener.mutex.Unlock()

	// Check if custom code is already taken
	if _, codeExists := shortener.urlDatabase[customCode]; codeExists {
		return "", fmt.Errorf("custom code '%s' already taken", customCode)
	}

	// Create the URL entry with custom code
	newEntry := &URLEntry{
		ShortCode:   customCode,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
		CreatedBy:   userID,
		IsCustom:    true, // Mark as custom code
		IsActive:    true,
	}

	// Store in both maps
	shortener.urlDatabase[customCode] = newEntry
	shortener.reverseLookup[originalURL] = customCode

	return shortener.baseDomain + "/" + customCode, nil
}

// Resolve converts a short code back to the original URL.
// This is called when someone clicks on a short link.
// Also records analytics for tracking click counts.
func (shortener *URLShortener) Resolve(shortCode string) (string, error) {
	// Use read lock for better concurrency (multiple readers allowed)
	shortener.mutex.RLock()
	urlEntry, exists := shortener.urlDatabase[shortCode]
	shortener.mutex.RUnlock()

	// Check if the short code exists
	if !exists {
		return "", fmt.Errorf("short URL not found")
	}

	// Check if the URL is still active (not deleted)
	if !urlEntry.IsActive {
		return "", fmt.Errorf("short URL is inactive")
	}

	// Check if the URL has expired
	if urlEntry.IsExpired() {
		return "", fmt.Errorf("short URL has expired")
	}

	// Record this click for analytics
	urlEntry.IncrementClicks()
	shortener.analyticsTracker.RecordClick(shortCode, "", "", "")

	return urlEntry.OriginalURL, nil
}

// Delete deactivates a short URL (soft delete).
// The entry still exists in the database but is marked as inactive.
// This allows us to keep analytics data while preventing future access.
func (shortener *URLShortener) Delete(shortCode string) error {
	shortener.mutex.Lock()
	defer shortener.mutex.Unlock()

	urlEntry, exists := shortener.urlDatabase[shortCode]
	if !exists {
		return fmt.Errorf("short URL not found")
	}

	// Soft delete - mark as inactive instead of removing
	urlEntry.IsActive = false
	return nil
}

// GetStats returns the URLEntry for a given short code.
// This provides access to all metadata including click count, creation time, etc.
func (shortener *URLShortener) GetStats(shortCode string) (*URLEntry, error) {
	shortener.mutex.RLock()
	defer shortener.mutex.RUnlock()

	urlEntry, exists := shortener.urlDatabase[shortCode]
	if !exists {
		return nil, fmt.Errorf("short URL not found")
	}

	return urlEntry, nil
}

// ListAll returns all URL entries stored in the service.
// Useful for admin dashboards or debugging.
func (shortener *URLShortener) ListAll() []*URLEntry {
	shortener.mutex.RLock()
	defer shortener.mutex.RUnlock()

	// Create a slice with capacity equal to the number of entries
	allEntries := make([]*URLEntry, 0, len(shortener.urlDatabase))
	for _, urlEntry := range shortener.urlDatabase {
		allEntries = append(allEntries, urlEntry)
	}
	return allEntries
}

// PrintStats prints detailed statistics for a URL entry in a formatted display.
func (shortener *URLShortener) PrintStats(shortCode string) {
	entry, err := shortener.GetStats(shortCode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf(`
╔════════════════════════════════════════════════╗
║              📊 URL STATISTICS                 ║
╠════════════════════════════════════════════════╣
  Short Code: %s
  Full URL:   %s/%s
  Original:   %s
  
  Created:    %s
  Expires:    %s
  Custom:     %v
  Active:     %v
  
  Total Clicks: %d
  Last Access:  %s
╚════════════════════════════════════════════════╝
`,
		entry.ShortCode,
		shortener.baseDomain, entry.ShortCode,
		entry.OriginalURL,
		entry.CreatedAt.Format("Jan 02, 2006 15:04"),
		formatExpiry(entry.ExpiresAt),
		entry.IsCustom,
		entry.IsActive,
		entry.GetClickCount(),
		formatLastAccess(entry.LastAccess),
	)
}

// formatExpiry formats the expiration time for display.
// Returns "Never" if no expiration was set.
func formatExpiry(expirationTime time.Time) string {
	if expirationTime.IsZero() {
		return "Never"
	}
	return expirationTime.Format("Jan 02, 2006")
}

// formatLastAccess formats the last access time for display.
// Returns "Never" if the URL was never accessed.
func formatLastAccess(lastAccessTime time.Time) string {
	if lastAccessTime.IsZero() {
		return "Never"
	}
	return lastAccessTime.Format("Jan 02, 2006 15:04")
}

// ========== MAIN ==========

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("        🔗 URL SHORTENER SERVICE")
	fmt.Println("═══════════════════════════════════════════")

	shortener := NewURLShortener("https://short.ly")

	// Create short URLs
	fmt.Println("\n📝 Creating Short URLs...")
	fmt.Println("─────────────────────────────────────────")

	url1, _ := shortener.Shorten("https://www.example.com/very/long/path/to/page?with=query&params=123", "user1", 0)
	fmt.Printf("✅ %s\n   → Original: https://www.example.com/very/long/...\n", url1)

	url2, _ := shortener.Shorten("https://github.com/golang/go/wiki/CodeReviewComments", "user1", 30)
	fmt.Printf("✅ %s (expires in 30 days)\n   → Original: https://github.com/golang/...\n", url2)

	url3, _ := shortener.Shorten("https://docs.google.com/document/d/abc123/edit", "user2", 0)
	fmt.Printf("✅ %s\n   → Original: https://docs.google.com/...\n", url3)

	// Custom short URL
	url4, err := shortener.ShortenCustom("https://myportfolio.com/projects", "mywork", "user1")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("✅ %s (custom)\n   → Original: https://myportfolio.com/projects\n", url4)
	}

	// Try duplicate custom code
	_, err = shortener.ShortenCustom("https://other.com", "mywork", "user2")
	if err != nil {
		fmt.Printf("❌ Custom code error: %v\n", err)
	}

	// Resolve URLs
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("🔍 Resolving Short URLs...")

	codes := []string{"0000001", "0000002", "mywork", "invalid"}
	for _, code := range codes {
		original, err := shortener.Resolve(code)
		if err != nil {
			fmt.Printf("  ❌ %s: %v\n", code, err)
		} else {
			displayURL := original
			if len(original) > 50 {
				displayURL = original[:50] + "..."
			}
			fmt.Printf("  ✅ %s → %s\n", code, displayURL)
		}
	}

	// Simulate multiple clicks
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("👆 Simulating clicks...")

	// Simulate 5 additional clicks on first URL (we already resolved once above)
	for i := 0; i < 5; i++ {
		_, _ = shortener.Resolve("0000001") // Ignoring errors for simulation
	}
	// Simulate 12 additional clicks on custom URL
	for i := 0; i < 12; i++ {
		_, _ = shortener.Resolve("mywork") // Ignoring errors for simulation
	}
	// Simulate 2 additional clicks on second URL
	_, _ = shortener.Resolve("0000002")
	_, _ = shortener.Resolve("0000002")

	// Show statistics
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("📊 URL Statistics...")

	shortener.PrintStats("0000001")
	shortener.PrintStats("mywork")

	// List all URLs
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("📋 All Short URLs:")
	for _, entry := range shortener.ListAll() {
		status := "🟢"
		if !entry.IsActive || entry.IsExpired() {
			status = "🔴"
		}
		// Truncate long URLs for display
		displayURL := entry.OriginalURL
		if len(displayURL) > 40 {
			displayURL = displayURL[:40] + "..."
		}
		fmt.Printf("  %s %s → %s (clicks: %d)\n",
			status, entry.ShortCode, displayURL, entry.GetClickCount())
	}

	// Delete a URL
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("🗑️  Deleting short URL...")
	err = shortener.Delete("0000001")
	if err != nil {
		fmt.Printf("  Error deleting: %v\n", err)
	} else {
		fmt.Println("  Successfully deleted 0000001")
	}

	_, err = shortener.Resolve("0000001")
	fmt.Printf("  Resolve deleted URL: %v\n", err)

	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  KEY DESIGN DECISIONS:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. Counter-based ID (guaranteed unique)")
	fmt.Println("  2. Base62 encoding for short codes")
	fmt.Println("  3. Custom aliases supported")
	fmt.Println("  4. Click tracking & analytics")
	fmt.Println("  5. TTL/expiration support")
	fmt.Println("═══════════════════════════════════════════")
}
