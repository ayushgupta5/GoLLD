# LRU Cache - Low Level Design

## 🎯 Problem Statement

Design a Least Recently Used (LRU) Cache that:
1. Has a fixed capacity
2. Get operation in O(1)
3. Put operation in O(1)
4. Evicts least recently used item when full

## 🧠 Interviewer's Mindset

This tests:
1. **Data Structure Knowledge** - HashMap + Doubly Linked List
2. **Time Complexity** - Must be O(1) for both operations
3. **Edge Cases** - Empty cache, capacity 0, update existing key

## 💡 Key Insight

Use TWO data structures:
- **HashMap**: For O(1) lookup by key
- **Doubly Linked List**: For O(1) removal and insertion (maintain order)

## 📋 Operations

- `Get(key)`: Return value, move to front (most recent)
- `Put(key, value)`: Add/update, move to front, evict if needed

