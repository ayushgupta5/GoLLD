# Logger System - Low Level Design

## 🎯 Problem Statement

Design a logging framework that:
1. Supports multiple log levels (DEBUG, INFO, WARN, ERROR)
2. Supports multiple output destinations (Console, File, Database)
3. Is thread-safe
4. Supports log formatting

## 🧠 Key Patterns

- **Singleton**: Logger instance
- **Chain of Responsibility**: Log level filtering
- **Strategy**: Different output handlers
- **Builder**: Log message construction

