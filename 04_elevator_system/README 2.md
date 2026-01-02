# Elevator System - Low Level Design

## 🎯 Problem Statement

Design an elevator system for a building that can:
1. Handle multiple elevators
2. Accept requests from floors (external) and inside elevator (internal)
3. Efficiently schedule elevator movements
4. Handle edge cases (full elevator, maintenance mode)

## 🧠 Interviewer's Mindset

This is a STATE-HEAVY problem. Interviewers evaluate:

1. **State Machine**: Can you model elevator states properly?
2. **Scheduling Algorithm**: How do you decide which elevator handles a request?
3. **Concurrency**: Multiple requests coming simultaneously
4. **Edge Cases**: Elevator full, emergency, maintenance

## ❓ Questions to Ask Interviewer

1. **Building Size**: How many floors? How many elevators?
2. **Elevator Types**: Passenger only? Freight? Service elevators?
3. **Capacity**: Max passengers per elevator?
4. **Features**: Express elevators? Floor restrictions?
5. **Request Handling**: FCFS? Nearest elevator? SCAN algorithm?

## 📋 Requirements

### Functional Requirements
- Multiple elevators in a building
- External requests (floor buttons: UP/DOWN)
- Internal requests (destination floor buttons inside elevator)
- Move elevators to requested floors
- Track elevator direction and current floor

### Non-Functional Requirements
- Efficient scheduling (minimize wait time)
- Thread-safe operations
- Extensible for new features

## 🎨 Design Approach

### Step 1: Identify States
```
Elevator States:
- IDLE       : Not moving, no requests
- MOVING_UP  : Moving upward
- MOVING_DOWN: Moving downward
- STOPPED    : Doors open, loading/unloading
- MAINTENANCE: Out of service
```

### Step 2: Identify Entities
- Building
- Elevator
- Floor
- Request (ExternalRequest, InternalRequest)
- ElevatorController (Scheduler)
- Display

### Step 3: Define Relationships
```
Building
    └── has many → Elevators
    └── has many → Floors
                      └── has → UP/DOWN buttons
    └── has one  → ElevatorController

ElevatorController
    └── schedules → Requests
    └── manages  → Elevators
```

### Step 4: State Pattern Application
- Each elevator state is a separate struct
- Transitions defined by state objects
- Clean separation of concerns

## 📁 Files Structure

```
04_elevator_system/
├── README.md
├── direction.go      # Direction enum
├── request.go        # Request types
├── elevator.go       # Elevator with states
├── floor.go          # Floor representation
├── controller.go     # Scheduling logic
├── building.go       # Main facade
└── main.go           # Demo
```

## 🔑 Key Design Patterns Used

1. **State Pattern**: Elevator states (Idle, Moving, Stopped)
2. **Strategy Pattern**: Scheduling algorithm
3. **Observer Pattern**: Notify floors when elevator arrives
4. **Singleton Pattern**: ElevatorController (optional)

## ⚡ Scheduling Algorithms

### 1. FCFS (First Come First Serve)
- Simple, fair
- Not efficient

### 2. SSTF (Shortest Seek Time First)
- Serve nearest floor first
- Better efficiency
- May cause starvation

### 3. SCAN (Elevator Algorithm)
- Go in one direction until end
- Then reverse
- Used in real elevators!

### 4. LOOK
- Like SCAN but doesn't go to end
- Reverses when no more requests in direction

## ❌ Common Mistakes

1. Not considering direction in scheduling
2. Forgetting to handle edge cases (empty requests, invalid floors)
3. Not making it thread-safe
4. Over-complicating the state machine
5. Not separating scheduling logic from elevator logic

