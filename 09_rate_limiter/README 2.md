# Rate Limiter - Low Level Design

## 🎯 Problem Statement

Design a Rate Limiter that:
1. Limits requests per user/API
2. Supports different algorithms
3. Thread-safe for concurrent requests

## 🧠 Common Algorithms

1. **Token Bucket** - Most common, allows bursts
2. **Sliding Window** - Smooth rate limiting
3. **Fixed Window** - Simple but has edge issues
4. **Leaky Bucket** - Constant rate output

## 📋 Use Cases

- API rate limiting (100 req/min per user)
- DDoS protection
- Resource allocation

