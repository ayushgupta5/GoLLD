# LLD Interview Questions & Answers

## 🎯 How to Approach LLD Interviews

### Step 1: Clarify Requirements (2-3 mins)
- Ask about scope (what features?)
- Ask about constraints (scale, users)
- Ask about edge cases
- Write down requirements

### Step 2: Identify Entities (3-5 mins)
- List all nouns → These become classes/structs
- List all verbs → These become methods
- Identify relationships between entities

### Step 3: Design Core Classes (5-10 mins)
- Start with the main entity
- Define attributes and methods
- Use interfaces for flexibility
- Apply SOLID principles

### Step 4: Handle Edge Cases (2-3 mins)
- What if input is invalid?
- What about concurrency?
- Error handling

### Step 5: Write Code (15-20 mins)
- Start with interfaces
- Implement core classes
- Add business logic
- Keep it simple!

---

## 📝 Common Interview Questions

### SOLID Principles

**Q: Explain Single Responsibility Principle with an example.**
```
A: A class should have only one reason to change.
Example: UserService should only handle user business logic,
not email sending (EmailService) or database operations (UserRepository).
```

**Q: How does Open/Closed Principle help in real projects?**
```
A: You can add new features without modifying existing tested code.
Example: Adding new PaymentMethod by implementing PaymentProcessor interface,
without changing existing CreditCard or PayPal code.
```

**Q: What's the difference between LSP and ISP?**
```
A: LSP ensures subtypes can replace parent types correctly.
   ISP ensures interfaces aren't too fat.
   
LSP violation: Square extends Rectangle but behaves differently
ISP violation: Robot forced to implement eat() because Worker interface has it
```

### Design Patterns

**Q: When would you use Factory vs Builder pattern?**
```
A: Factory: When you need to create one of several related objects
   Builder: When object has many optional parameters or complex construction

Factory example: PaymentProcessorFactory.create("paypal")
Builder example: User.Builder().Name("John").Age(25).Email("...").Build()
```

**Q: Explain Strategy pattern with a real example.**
```
A: Strategy allows swapping algorithms at runtime.
Example: Sorting - can switch between QuickSort and MergeSort
Example: Payments - can switch between CreditCard and PayPal
The context (ShoppingCart) doesn't care which strategy is used.
```

**Q: When is Singleton appropriate? What are its drawbacks?**
```
A: Appropriate for: Logger, Config, Connection Pool, Cache

Drawbacks:
- Hard to test (global state)
- Hidden dependencies
- Concurrency issues if not implemented correctly
- Tight coupling
```

**Q: How does Observer pattern help in event-driven systems?**
```
A: Decouples publishers from subscribers.
- Publishers don't know who's listening
- Easy to add new subscribers
- Subscribers can subscribe/unsubscribe dynamically

Example: Stock price alerts - Investors subscribe to price changes
```

### System Design

**Q: Design a Parking Lot - what classes would you need?**
```
A: Core Classes:
- ParkingLot (main facade)
- Floor (contains spots)
- ParkingSpot (small, medium, large)
- Vehicle (interface) → Car, Motorcycle, Truck
- Ticket (tracks parking session)
- Payment (Strategy for fee calculation)

Key decisions:
- Vehicle interface for extensibility
- Strategy pattern for pricing
- Thread-safe spot allocation
```

**Q: How would you handle concurrent access in Parking Lot?**
```
A: 
1. Use mutex/RWMutex for shared resources
2. Lock when finding spot, parking, unparking
3. Use RLock for read operations (checking availability)
4. Consider using channels for Go-idiomatic approach
```

**Q: Design an Elevator System - what's the key challenge?**
```
A: Key challenge is SCHEDULING algorithm.

Options:
1. FCFS - Simple but inefficient
2. SSTF - Nearest first, may cause starvation
3. SCAN - Go in one direction, then reverse (used in real elevators)

Key classes:
- Building (facade)
- ElevatorController (scheduler)
- Elevator (state machine)
- Request (floor + direction)
```

**Q: What pattern would you use for Elevator states?**
```
A: State Pattern

States: IDLE, MOVING_UP, MOVING_DOWN, STOPPED, MAINTENANCE
Each state handles operations differently:
- IDLE can accept requests
- MOVING can add stops in same direction
- MAINTENANCE rejects all requests
```

---

## 🎯 Problem-Specific Questions

### Parking Lot
1. How would you handle multiple entry/exit gates?
2. How to implement spot reservation?
3. How to add electric vehicle charging spots?
4. How to handle lost tickets?

### Elevator System
1. How would you prioritize emergency stops?
2. How to handle VIP/express elevators?
3. How to optimize for rush hour?
4. How to handle elevator capacity?

### Snake & Ladder
1. How to add power-ups?
2. How to support multiplayer online?
3. How to add undo functionality?
4. How to persist game state?

---

## ❌ Common Mistakes to Avoid

1. **Jumping to code without clarifying requirements**
   - Always spend 2-3 mins asking questions

2. **Over-engineering**
   - Don't add features not asked for
   - KISS (Keep It Simple, Stupid)

3. **Ignoring SOLID**
   - Giant classes doing everything
   - Hardcoded dependencies

4. **Not considering concurrency**
   - Always mention thread safety
   - Use mutex where needed

5. **Not handling errors**
   - Always return errors
   - Validate inputs

6. **Not explaining decisions**
   - Say WHY you're doing something
   - Trade-offs matter!

---

## 💡 Pro Tips

1. **Think aloud** - Interviewers want to see your thought process
2. **Start simple** - Get basic design right, then extend
3. **Use interfaces** - Shows you understand abstraction
4. **Mention trade-offs** - Shows maturity
5. **Ask clarifying questions** - Shows real-world thinking
6. **Draw diagrams** - Visual helps both you and interviewer
7. **Know patterns** - But don't force them where not needed

