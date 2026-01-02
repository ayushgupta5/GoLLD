# Parking Lot System - Low Level Design

## 🎯 Problem Statement

Design a parking lot system that can:
1. Park vehicles of different sizes (motorcycle, car, truck)
2. Handle multiple floors/levels
3. Track available spots
4. Calculate parking fees

## 🧠 Interviewer's Mindset

This is THE most asked LLD question! Interviewers evaluate:

1. **Requirement Clarification** - Do you ask the right questions?
2. **Entity Identification** - Can you identify key objects?
3. **Relationship Modeling** - How do entities interact?
4. **SOLID Principles** - Is your design extensible?
5. **Edge Cases** - What about full lots, invalid inputs?

## ❓ Questions to Ask Interviewer

Always ask these before designing:

1. **Vehicle Types**: What types of vehicles? (motorcycle, car, bus, truck?)
2. **Parking Spots**: Different spot sizes? Can small vehicles use large spots?
3. **Multiple Floors**: Single level or multiple floors?
4. **Entry/Exit**: Multiple entry/exit points?
5. **Payment**: Hourly rate? Different rates for different vehicles?
6. **Concurrency**: Multiple vehicles entering/exiting simultaneously?
7. **Features**: Reservations? Electric charging spots?

## 📋 Requirements (Simplified for Interview)

### Functional Requirements
- Support multiple vehicle types (Motorcycle, Car, Truck)
- Support different spot types (Small, Medium, Large)
- Multiple floors with multiple spots per floor
- Track which spots are available
- Park and unpark vehicles
- Calculate parking fee based on duration

### Non-Functional Requirements
- Thread-safe (concurrent access)
- Efficient spot finding
- Extensible design

## 🎨 Design Approach

### Step 1: Identify Entities (Nouns)
- ParkingLot
- Floor
- ParkingSpot
- Vehicle (Motorcycle, Car, Truck)
- Ticket
- Payment

### Step 2: Identify Actions (Verbs)
- Park vehicle
- Unpark vehicle
- Find available spot
- Calculate fee
- Process payment

### Step 3: Define Relationships
```
ParkingLot
    └── has many → Floors
                      └── has many → ParkingSpots
                                        └── can hold → Vehicle

Vehicle → gets → Ticket
Ticket → has → ParkingSpot, Entry Time, Exit Time
```

### Step 4: Apply SOLID
- **S**: Each class has one job (Spot manages spot, Vehicle represents vehicle)
- **O**: New vehicle types can be added without changing existing code
- **L**: All vehicles can be parked if they implement Vehicle interface
- **I**: Small, focused interfaces
- **D**: ParkingLot depends on interfaces, not concrete types

## 📁 Files Structure

```
03_parking_lot/
├── README.md           # This file
├── vehicle.go          # Vehicle types
├── parking_spot.go     # Parking spot types
├── floor.go            # Floor management
├── ticket.go           # Parking ticket
├── parking_lot.go      # Main parking lot logic
├── payment.go          # Payment strategies
└── main.go             # Demo
```

## 🔑 Key Interview Points

1. **Start Simple**: Don't over-engineer initially
2. **Extend Gradually**: Show how design can grow
3. **Think Aloud**: Explain your decisions
4. **Handle Edge Cases**: Full lot, invalid vehicle, etc.
5. **Consider Concurrency**: Mention mutex/locks for thread safety

## ❌ Common Mistakes

1. Jumping to code without clarifying requirements
2. Over-engineering (reservations, valet, etc. unless asked)
3. Not considering thread safety
4. Hardcoding vehicle/spot types (use enums/constants)
5. Not handling error cases

