# SOLID Principles in Golang

## 🎯 What are SOLID Principles?

SOLID is an acronym for **5 design principles** that help us write:
- Clean code
- Maintainable code
- Extensible code

Think of SOLID as **rules that prevent your code from becoming a mess** as it grows.

## 🧠 Interviewer's Perspective

Interviewers check SOLID because:
1. It shows you understand **good design**
2. It proves you can write **production-quality code**
3. It demonstrates you think about **future changes**

**What interviewers look for:**
- Can you explain WHY a principle matters?
- Can you identify violations?
- Can you refactor bad code to follow SOLID?

---

## 📘 S - Single Responsibility Principle (SRP)

### Simple Explanation
> "A struct should have only ONE reason to change"

Think of it like a **job title**:
- A Chef cooks food (one job)
- A Chef shouldn't also be the accountant, waiter, and cleaner

### Real-World Analogy
Imagine your phone's camera app:
- Bad: Camera app that also manages contacts and plays music
- Good: Camera app only handles photos/videos

### Why It Matters
- Easier to test (one thing to test)
- Easier to understand (one purpose)
- Easier to change (change won't break unrelated features)

---

## 📘 O - Open/Closed Principle (OCP)

### Simple Explanation
> "Open for extension, closed for modification"

You should be able to **add new features WITHOUT changing existing code**.

### Real-World Analogy
Think of a **power strip**:
- You can plug in new devices (extend)
- You don't need to rewire the strip (no modification)

### Why It Matters
- Existing code is tested and working
- Modifying it can introduce bugs
- Extensions are safer

---

## 📘 L - Liskov Substitution Principle (LSP)

### Simple Explanation
> "If S is a subtype of T, you should be able to use S wherever T is expected"

In Go terms: **If a struct implements an interface, it should work correctly when used through that interface**.

### Real-World Analogy
If you order a "vehicle" to transport goods:
- A truck works ✓
- A car works ✓
- A toy car doesn't work ✗ (violates the "vehicle" contract)

### Why It Matters
- Ensures interfaces are implemented correctly
- Prevents surprises when using polymorphism

---

## 📘 I - Interface Segregation Principle (ISP)

### Simple Explanation
> "Don't force structs to implement methods they don't need"

Create **small, focused interfaces** instead of big, fat ones.

### Real-World Analogy
Job applications:
- Bad: One form asking for cooking skills, programming skills, and flying experience
- Good: Different forms for different job types

### Why It Matters
- Simpler interfaces are easier to implement
- Classes aren't burdened with unused methods
- **Go naturally encourages this** (small interfaces are idiomatic)

---

## 📘 D - Dependency Inversion Principle (DIP)

### Simple Explanation
> "Depend on abstractions (interfaces), not concrete implementations"

High-level modules shouldn't depend on low-level modules. Both should depend on interfaces.

### Real-World Analogy
Your laptop charger:
- You plug into a wall socket (interface)
- You don't care HOW electricity is generated (concrete implementation)
- Could be solar, nuclear, or coal - your laptop doesn't care

### Why It Matters
- Easy to swap implementations
- Easy to test (mock interfaces)
- Loosely coupled code

---

## 🔑 Key Points for Interviews

1. **SRP**: One struct, one job
2. **OCP**: Add features by adding code, not changing code
3. **LSP**: Subtypes must honor the parent contract
4. **ISP**: Small interfaces > big interfaces
5. **DIP**: Depend on interfaces, not concrete types

## ❌ Common Mistakes

1. Confusing SRP with "one method per struct"
2. Thinking OCP means never modify code
3. Creating interfaces before you need them
4. Making every field an interface (over-engineering)
5. Not being able to explain WHY you used a principle

