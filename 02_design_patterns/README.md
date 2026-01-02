# Design Patterns in Golang for LLD Interviews

## 🎯 Why Learn Design Patterns?

Design patterns are **proven solutions to common problems**. In interviews:
1. They show you've worked on real systems
2. They demonstrate you can communicate design using standard terms
3. They help you solve problems faster

## 🧠 Interviewer's Perspective

Interviewers DON'T want you to:
- Force patterns where they don't fit
- Over-engineer simple problems
- Memorize patterns without understanding

Interviewers DO want you to:
- Know when a pattern is applicable
- Explain WHY you chose a pattern
- Implement it cleanly in Go

## 📚 Patterns Covered

### Creational Patterns (How objects are created)
1. **Singleton** - One instance only
2. **Factory** - Object creation logic
3. **Builder** - Complex object construction

### Structural Patterns (How objects are composed)
4. **Adapter** - Make incompatible interfaces work together
5. **Decorator** - Add behavior dynamically

### Behavioral Patterns (How objects communicate)
6. **Strategy** - Interchangeable algorithms
7. **Observer** - Event notification
8. **State** - State-dependent behavior

## 🔑 Most Important for Interviews

1. **Strategy** - Used in almost every LLD problem
2. **Factory** - Object creation flexibility
3. **Observer** - Event-driven systems
4. **State** - Vending machine, elevators, etc.
5. **Singleton** - Configuration, connection pools

## ⚠️ Go-Specific Notes

Go doesn't have traditional OOP (classes, inheritance), so patterns look different:
- Use **interfaces** instead of abstract classes
- Use **embedding** instead of inheritance
- Use **package-level variables** for singleton
- Use **closures** for simple strategies

## 📁 Files in This Module

```
02_design_patterns/
├── README.md
├── 01_singleton.go       # One instance pattern
├── 02_factory.go         # Object creation
├── 03_builder.go         # Complex construction
├── 04_strategy.go        # Interchangeable algorithms
├── 05_observer.go        # Event notification
├── 06_state.go           # State machine pattern
├── 07_decorator.go       # Add behavior dynamically
└── 08_adapter.go         # Interface adaptation
```

