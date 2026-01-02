package main

import (
	"fmt"
	"sync"
)

// ============================================================================
// LRU CACHE - Complete Low-Level Design Implementation
// ============================================================================
//
// WHAT IS LRU CACHE?
// -------------------
// LRU (Least Recently Used) Cache is a data structure that stores a limited
// number of items. When the cache is full and we need to add a new item,
// we remove the item that was accessed least recently.
//
// REAL-WORLD EXAMPLES:
// - Browser cache (stores recently visited web pages)
// - CPU cache (stores recently accessed memory)
// - Database query cache (stores recent query results)
//
// WHY THIS DATA STRUCTURE?
// ------------------------
// We need TWO data structures working together:
//
// 1. HashMap (map in Go): Gives us O(1) lookup by key
//    - Without this, we'd need O(n) to find an item
//
// 2. Doubly Linked List: Gives us O(1) reordering
//    - Maintains items in order of recent access
//    - Most recently used at HEAD, least recently used at TAIL
//    - Without this, moving items would be O(n)
//
// TIME COMPLEXITY: O(1) for both Get and Put operations
// SPACE COMPLEXITY: O(capacity) - we store at most 'capacity' items
//
// ============================================================================

// ========================== DOUBLY LINKED LIST NODE ==========================

// CacheNode represents a single entry in our cache.
// It holds the key-value pair and pointers to neighboring nodes.
//
// Why store the key in the node?
// - When we evict the LRU node, we need its key to remove it from the HashMap
type CacheNode struct {
	key      int        // The cache key (needed for HashMap removal during eviction)
	value    int        // The cached value
	previous *CacheNode // Pointer to the previous node (towards HEAD)
	next     *CacheNode // Pointer to the next node (towards TAIL)
}

// NewCacheNode creates a new cache node with the given key and value.
func NewCacheNode(key, value int) *CacheNode {
	return &CacheNode{
		key:   key,
		value: value,
		// previous and next are nil by default (Go zero values)
	}
}

// ============================ DOUBLY LINKED LIST =============================

// DoublyLinkedList maintains the access order of cache entries.
//
// Structure visualization:
//
//	HEAD (dummy) <-> Node1 <-> Node2 <-> Node3 <-> TAIL (dummy)
//	                  ^                    ^
//	           Most Recent            Least Recent
//	              (MRU)                  (LRU)
//
// Why use dummy head and tail?
// - Eliminates null checks when adding/removing nodes
// - Simplifies edge cases (empty list, single element)
// - The actual data nodes are always BETWEEN head and tail
type DoublyLinkedList struct {
	head *CacheNode // Dummy head node (never contains real data)
	tail *CacheNode // Dummy tail node (never contains real data)
	size int        // Current number of data nodes (excludes dummy nodes)
}

// NewDoublyLinkedList creates an empty doubly linked list with dummy head and tail.
func NewDoublyLinkedList() *DoublyLinkedList {
	list := &DoublyLinkedList{
		head: NewCacheNode(0, 0), // Dummy head (key/value don't matter)
		tail: NewCacheNode(0, 0), // Dummy tail (key/value don't matter)
		size: 0,
	}

	// Connect head and tail to each other (empty list)
	// HEAD <-> TAIL
	list.head.next = list.tail
	list.tail.previous = list.head

	return list
}

// AddToFront inserts a node right after the head (most recently used position).
//
// Before: HEAD <-> A <-> B <-> TAIL
// After adding X: HEAD <-> X <-> A <-> B <-> TAIL
//
// Time Complexity: O(1)
func (list *DoublyLinkedList) AddToFront(node *CacheNode) {
	// Step 1: Set the new node's pointers
	node.previous = list.head  // New node's previous points to HEAD
	node.next = list.head.next // New node's next points to what was after HEAD

	// Step 2: Update existing connections
	list.head.next.previous = node // The old first node now points back to new node
	list.head.next = node          // HEAD now points forward to new node

	list.size++
}

// RemoveNode removes a node from its current position in the list.
//
// Before: A <-> X <-> B (removing X)
// After:  A <-> B
//
// Time Complexity: O(1)
func (list *DoublyLinkedList) RemoveNode(node *CacheNode) {
	// Skip the node by connecting its neighbors to each other
	node.previous.next = node.next     // Previous node skips over 'node' to next
	node.next.previous = node.previous // Next node skips back over 'node' to previous

	list.size--
}

// RemoveLRUNode removes and returns the least recently used node (right before tail).
//
// Before: HEAD <-> A <-> B <-> LRU <-> TAIL
// After:  HEAD <-> A <-> B <-> TAIL (LRU is removed and returned)
//
// Time Complexity: O(1)
func (list *DoublyLinkedList) RemoveLRUNode() *CacheNode {
	// If list is empty, nothing to remove
	if list.size == 0 {
		return nil
	}

	// The LRU node is always right before the TAIL
	lruNode := list.tail.previous
	list.RemoveNode(lruNode)

	return lruNode
}

// MoveToFront moves an existing node to the front (marks it as most recently used).
//
// Before: HEAD <-> A <-> X <-> B <-> TAIL (X is somewhere in middle)
// After:  HEAD <-> X <-> A <-> B <-> TAIL (X is now at front)
//
// Time Complexity: O(1)
func (list *DoublyLinkedList) MoveToFront(node *CacheNode) {
	// First remove from current position, then add to front
	list.RemoveNode(node)
	list.AddToFront(node)
}

// GetSize returns the current number of nodes in the list (excluding dummy nodes).
func (list *DoublyLinkedList) GetSize() int {
	return list.size
}

// IsEmpty returns true if the list has no data nodes.
func (list *DoublyLinkedList) IsEmpty() bool {
	return list.size == 0
}

// ================================ LRU CACHE ==================================

// LRUCache implements a thread-safe Least Recently Used cache.
//
// How it works:
// 1. HashMap (cache): Stores key -> node mapping for O(1) lookup
// 2. DoublyLinkedList (accessOrder): Maintains access order for O(1) eviction
//
// On Get(key):
// - Look up the node in HashMap
// - Move the node to front of list (mark as recently used)
// - Return the value
//
// On Put(key, value):
// - If key exists: Update value and move to front
// - If key doesn't exist:
//   - If cache is full: Evict LRU node (from tail) and remove from HashMap
//   - Create new node, add to front of list and HashMap
type LRUCache struct {
	capacity    int                // Maximum number of items the cache can hold
	cache       map[int]*CacheNode // HashMap: key -> node (for O(1) lookup)
	accessOrder *DoublyLinkedList  // Doubly Linked List: maintains access order
	mutex       sync.RWMutex       // Read-Write mutex for thread safety
}

// NewLRUCache creates a new LRU cache with the specified maximum capacity.
//
// Example:
//
//	cache := NewLRUCache(3)  // Cache can hold at most 3 items
func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 1 // Ensure minimum capacity of 1
	}

	return &LRUCache{
		capacity:    capacity,
		cache:       make(map[int]*CacheNode),
		accessOrder: NewDoublyLinkedList(),
	}
}

// Get retrieves the value associated with the given key.
//
// Returns:
// - (value, true) if the key exists in the cache
// - (0, false) if the key does not exist
//
// Side effect: If found, the item is marked as recently used (moved to front)
//
// Time Complexity: O(1)
//
// Example:
//
//	value, found := cache.Get(1)
//	if found {
//	    fmt.Println("Value:", value)
//	} else {
//	    fmt.Println("Key not found")
//	}
func (lru *LRUCache) Get(key int) (int, bool) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	// Look up the key in our HashMap
	node, exists := lru.cache[key]

	if !exists {
		// Key not found - cache miss
		return 0, false
	}

	// Key found - cache hit!
	// Move this node to front (mark as most recently used)
	lru.accessOrder.MoveToFront(node)

	return node.value, true
}

// Put adds a new key-value pair to the cache or updates an existing key.
//
// If the cache is at capacity and the key is new:
// - The least recently used item is automatically evicted
//
// Time Complexity: O(1)
//
// Example:
//
//	cache.Put(1, 100)  // Add key=1, value=100
//	cache.Put(1, 200)  // Update key=1 to value=200
func (lru *LRUCache) Put(key, value int) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	// Check if key already exists
	if existingNode, exists := lru.cache[key]; exists {
		// Key exists - update the value and move to front
		existingNode.value = value
		lru.accessOrder.MoveToFront(existingNode)
		return
	}

	// Key doesn't exist - we need to add a new entry

	// First, check if we need to evict (cache is full)
	if lru.accessOrder.GetSize() >= lru.capacity {
		// Remove the least recently used node (at the tail)
		evictedNode := lru.accessOrder.RemoveLRUNode()
		if evictedNode != nil {
			// Also remove from HashMap using the key stored in the node
			delete(lru.cache, evictedNode.key)
		}
	}

	// Create new node and add to cache
	newNode := NewCacheNode(key, value)
	lru.accessOrder.AddToFront(newNode) // Add to front of list (most recent)
	lru.cache[key] = newNode            // Add to HashMap for O(1) lookup
}

// Delete removes a key from the cache.
//
// Returns:
// - true if the key was found and removed
// - false if the key was not in the cache
//
// Time Complexity: O(1)
//
// Example:
//
//	wasDeleted := cache.Delete(1)
func (lru *LRUCache) Delete(key int) bool {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	node, exists := lru.cache[key]
	if !exists {
		return false // Key not found
	}

	// Remove from both data structures
	lru.accessOrder.RemoveNode(node)
	delete(lru.cache, key)

	return true
}

// Size returns the current number of items in the cache.
func (lru *LRUCache) Size() int {
	lru.mutex.RLock()
	defer lru.mutex.RUnlock()

	return lru.accessOrder.GetSize()
}

// Capacity returns the maximum number of items the cache can hold.
func (lru *LRUCache) Capacity() int {
	return lru.capacity
}

// Contains checks if a key exists in the cache WITHOUT marking it as recently used.
// Use this when you want to check existence without affecting the LRU order.
func (lru *LRUCache) Contains(key int) bool {
	lru.mutex.RLock()
	defer lru.mutex.RUnlock()

	_, exists := lru.cache[key]
	return exists
}

// Clear removes all items from the cache.
func (lru *LRUCache) Clear() {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	// Reset both data structures
	lru.cache = make(map[int]*CacheNode)
	lru.accessOrder = NewDoublyLinkedList()
}

// PrintCache displays the cache contents for debugging.
// Items are shown from most recently used (left) to least recently used (right).
func (lru *LRUCache) PrintCache() {
	lru.mutex.RLock()
	defer lru.mutex.RUnlock()

	fmt.Printf("Cache [%d/%d items]: ", lru.accessOrder.GetSize(), lru.capacity)

	// Traverse from head to tail (MRU to LRU)
	currentNode := lru.accessOrder.head.next
	for currentNode != lru.accessOrder.tail {
		fmt.Printf("(%d=%d) ", currentNode.key, currentNode.value)
		currentNode = currentNode.next
	}

	if lru.accessOrder.IsEmpty() {
		fmt.Print("(empty)")
	}
	fmt.Println()
}

// ========================= GENERIC LRU CACHE (BONUS) =========================
// This version works with any key type (string, int, etc.) and any value type.
// Uses Go 1.18+ generics for type safety.

// GenericCacheNode is a generic version of CacheNode for any key-value types.
type GenericCacheNode[K comparable, V any] struct {
	key      K
	value    V
	previous *GenericCacheNode[K, V]
	next     *GenericCacheNode[K, V]
}

// GenericLRUCache is a type-safe LRU cache that works with any comparable key
// and any value type.
//
// Example usage:
//
//	// String keys, int values
//	cache1 := NewGenericLRUCache[string, int](100)
//
//	// Int keys, struct values
//	cache2 := NewGenericLRUCache[int, User](50)
type GenericLRUCache[K comparable, V any] struct {
	capacity int
	cache    map[K]*GenericCacheNode[K, V]
	head     *GenericCacheNode[K, V] // Dummy head
	tail     *GenericCacheNode[K, V] // Dummy tail
	mutex    sync.RWMutex
}

// NewGenericLRUCache creates a new generic LRU cache with the specified capacity.
func NewGenericLRUCache[K comparable, V any](capacity int) *GenericLRUCache[K, V] {
	if capacity <= 0 {
		capacity = 1
	}

	cache := &GenericLRUCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]*GenericCacheNode[K, V]),
		head:     &GenericCacheNode[K, V]{}, // Dummy head
		tail:     &GenericCacheNode[K, V]{}, // Dummy tail
	}

	// Connect dummy head and tail
	cache.head.next = cache.tail
	cache.tail.previous = cache.head

	return cache
}

// addToFront adds a node right after the head (internal helper method).
func (cache *GenericLRUCache[K, V]) addToFront(node *GenericCacheNode[K, V]) {
	node.previous = cache.head
	node.next = cache.head.next
	cache.head.next.previous = node
	cache.head.next = node
}

// removeNode removes a node from its current position (internal helper method).
func (cache *GenericLRUCache[K, V]) removeNode(node *GenericCacheNode[K, V]) {
	node.previous.next = node.next
	node.next.previous = node.previous
}

// Get retrieves the value for a key from the generic cache.
// Returns the value and true if found, or zero value and false if not found.
func (cache *GenericLRUCache[K, V]) Get(key K) (V, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	node, exists := cache.cache[key]
	if !exists {
		var zeroValue V // Return zero value for type V
		return zeroValue, false
	}

	// Move to front (mark as recently used)
	cache.removeNode(node)
	cache.addToFront(node)

	return node.value, true
}

// Put adds or updates a key-value pair in the generic cache.
func (cache *GenericLRUCache[K, V]) Put(key K, value V) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// Check if key already exists
	if existingNode, exists := cache.cache[key]; exists {
		existingNode.value = value
		cache.removeNode(existingNode)
		cache.addToFront(existingNode)
		return
	}

	// Evict if at capacity
	if len(cache.cache) >= cache.capacity {
		lruNode := cache.tail.previous
		cache.removeNode(lruNode)
		delete(cache.cache, lruNode.key)
	}

	// Add new node
	newNode := &GenericCacheNode[K, V]{key: key, value: value}
	cache.addToFront(newNode)
	cache.cache[key] = newNode
}

// Size returns the current number of items in the generic cache.
func (cache *GenericLRUCache[K, V]) Size() int {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return len(cache.cache)
}

// ================================== MAIN =====================================

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║              LRU CACHE - Low Level Design Demo                ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")

	// Create an LRU cache that can hold maximum 3 items
	cache := NewLRUCache(3)

	fmt.Println("\n📋 DEMONSTRATION OF LRU CACHE OPERATIONS")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	// Operation 1: Add items to the cache
	fmt.Println("\n▶ Step 1: Adding items to cache")
	fmt.Println("   cache.Put(1, 100)")
	cache.Put(1, 100)
	cache.PrintCache()

	fmt.Println("   cache.Put(2, 200)")
	cache.Put(2, 200)
	cache.PrintCache()

	fmt.Println("   cache.Put(3, 300)")
	cache.Put(3, 300)
	cache.PrintCache()

	// Operation 2: Access an item (moves it to front)
	fmt.Println("\n▶ Step 2: Accessing key 1 (moves it to front as MRU)")
	fmt.Println("   cache.Get(1)")
	if value, found := cache.Get(1); found {
		fmt.Printf("   → Found value: %d\n", value)
	}
	cache.PrintCache()
	fmt.Println("   Notice: Key 1 is now at the front (most recently used)")

	// Operation 3: Add new item when cache is full (triggers eviction)
	fmt.Println("\n▶ Step 3: Adding key 4 when cache is full")
	fmt.Println("   cache.Put(4, 400)")
	fmt.Println("   → Cache is full, so LRU item (key 2) will be evicted")
	cache.Put(4, 400)
	cache.PrintCache()

	// Operation 4: Try to access evicted item
	fmt.Println("\n▶ Step 4: Trying to access evicted key 2")
	fmt.Println("   cache.Get(2)")
	if _, found := cache.Get(2); !found {
		fmt.Println("   → Key 2 not found (was evicted) ✓")
	}

	// Operation 5: Update existing key
	fmt.Println("\n▶ Step 5: Updating existing key 3")
	fmt.Println("   cache.Put(3, 333)")
	cache.Put(3, 333)
	cache.PrintCache()
	fmt.Println("   Notice: Key 3's value updated and moved to front")

	// Operation 6: Another eviction
	fmt.Println("\n▶ Step 6: Adding key 5 (triggers another eviction)")
	fmt.Println("   cache.Put(5, 500)")
	cache.Put(5, 500)
	cache.PrintCache()

	// Operation 7: Delete a key
	fmt.Println("\n▶ Step 7: Deleting key 3")
	fmt.Println("   cache.Delete(3)")
	deleted := cache.Delete(3)
	fmt.Printf("   → Deleted: %v\n", deleted)
	cache.PrintCache()

	// ========== GENERIC CACHE DEMO ==========
	fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║            GENERIC LRU CACHE (String Keys Demo)               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")

	// Create a string->string cache with capacity 2
	stringCache := NewGenericLRUCache[string, string](2)

	fmt.Println("\n   stringCache.Put(\"name\", \"Alice\")")
	stringCache.Put("name", "Alice")

	fmt.Println("   stringCache.Put(\"city\", \"New York\")")
	stringCache.Put("city", "New York")

	if value, found := stringCache.Get("name"); found {
		fmt.Printf("   stringCache.Get(\"name\") → \"%s\"\n", value)
	}

	fmt.Println("\n   stringCache.Put(\"country\", \"USA\")  // This evicts \"city\"")
	stringCache.Put("country", "USA")

	if _, found := stringCache.Get("city"); !found {
		fmt.Println("   stringCache.Get(\"city\") → Not found (was evicted) ✓")
	}

	// ========== KEY DESIGN SUMMARY ==========
	fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    KEY DESIGN DECISIONS                       ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════╣")
	fmt.Println("║  1. HashMap (map)       → O(1) key lookup                     ║")
	fmt.Println("║  2. Doubly Linked List  → O(1) reordering/eviction            ║")
	fmt.Println("║  3. Dummy head/tail     → Simplifies edge cases               ║")
	fmt.Println("║  4. sync.RWMutex        → Thread-safe concurrent access       ║")
	fmt.Println("║  5. Key stored in node  → Enables HashMap cleanup on eviction ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")

	fmt.Println("\n✅ LRU Cache demonstration complete!")
}
