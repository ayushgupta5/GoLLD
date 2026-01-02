package main

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// ============================================================
// ELEVATOR SYSTEM - Low Level Design Implementation
// ============================================================
//
// This implementation demonstrates several important design patterns:
// - State Pattern: Elevator has different states (Idle, Moving, Stopped, Maintenance)
// - Strategy Pattern: Different algorithms for selecting elevators
// - Facade Pattern: Building class provides a simple interface
// - SCAN Algorithm: Efficient floor-serving algorithm (like disk scheduling)
//
// Key Components:
// 1. Direction - Which way the elevator is moving
// 2. ElevatorState - Current operational state of elevator
// 3. Elevator - Individual elevator car with its own logic
// 4. SchedulingStrategy - Interface for elevator selection algorithms
// 5. ElevatorController - Manages all elevators and coordinates requests
// 6. Building - Main entry point that users interact with

// ============================================================
// CONSTANTS - Configuration values for the elevator system
// ============================================================

const (
	// Timing constants for simulation
	FloorTravelTime = 200 * time.Millisecond // Time to travel one floor
	DoorOpenTime    = 500 * time.Millisecond // Time doors stay open

	// Scheduling strategy bonuses/penalties
	SameDirectionBonus       = 10 // Bonus score for elevator going same direction
	OppositeDirectionPenalty = 20 // Penalty for elevator going opposite direction
)

// ============================================================
// DIRECTION - Represents elevator movement direction
// ============================================================

// Direction represents which way an elevator is moving
type Direction int

const (
	DirectionIdle Direction = iota // Elevator is stationary
	DirectionUp                    // Elevator is moving up
	DirectionDown                  // Elevator is moving down
)

// String returns a human-readable representation of the direction
func (d Direction) String() string {
	switch d {
	case DirectionUp:
		return "UP ⬆️"
	case DirectionDown:
		return "DOWN ⬇️"
	default:
		return "IDLE ⏸️"
	}
}

// ============================================================
// ELEVATOR STATE - Represents operational state of an elevator
// ============================================================

// ElevatorState represents what the elevator is currently doing
type ElevatorState int

const (
	StateIdle        ElevatorState = iota // Elevator is idle, waiting for requests
	StateMoving                           // Elevator is moving between floors
	StateStopped                          // Elevator stopped at a floor (doors open)
	StateMaintenance                      // Elevator is under maintenance
)

// String returns a human-readable representation of the state
func (s ElevatorState) String() string {
	switch s {
	case StateIdle:
		return "IDLE"
	case StateMoving:
		return "MOVING"
	case StateStopped:
		return "STOPPED"
	case StateMaintenance:
		return "MAINTENANCE"
	default:
		return "UNKNOWN"
	}
}

// ============================================================
// ELEVATOR - Represents a single elevator car
// ============================================================

// Elevator represents a single elevator car in the building
type Elevator struct {
	id              int           // Unique identifier for this elevator
	currentFloor    int           // The floor the elevator is currently on
	direction       Direction     // Current movement direction
	state           ElevatorState // Current operational state
	pendingRequests []int         // List of floors this elevator needs to visit
	maxCapacity     int           // Maximum number of people
	currentLoad     int           // Current number of people
	minFloor        int           // Lowest floor this elevator serves
	maxFloor        int           // Highest floor this elevator serves
	mutex           sync.Mutex    // Protects concurrent access to elevator state
}

// NewElevator creates a new elevator with the given configuration
func NewElevator(id, minFloor, maxFloor, capacity int) *Elevator {
	return &Elevator{
		id:              id,
		currentFloor:    minFloor, // Start at the lowest floor
		direction:       DirectionIdle,
		state:           StateIdle,
		pendingRequests: make([]int, 0),
		maxCapacity:     capacity,
		currentLoad:     0,
		minFloor:        minFloor,
		maxFloor:        maxFloor,
	}
}

// GetID returns the elevator's unique identifier
func (e *Elevator) GetID() int {
	return e.id
}

// GetCurrentFloor returns the floor the elevator is currently on (thread-safe)
func (e *Elevator) GetCurrentFloor() int {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.currentFloor
}

// GetDirection returns the current movement direction (thread-safe)
func (e *Elevator) GetDirection() Direction {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.direction
}

// GetState returns the current operational state (thread-safe)
func (e *Elevator) GetState() ElevatorState {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.state
}

// IsAvailable checks if the elevator can accept new requests
// An elevator is available if it's not under maintenance and has capacity
func (e *Elevator) IsAvailable() bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.state != StateMaintenance && e.currentLoad < e.maxCapacity
}

// AddFloorRequest adds a floor to the elevator's request queue
// Duplicate requests are ignored to avoid unnecessary stops
func (e *Elevator) AddFloorRequest(floor int) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Check if this floor is already in the queue (avoid duplicates)
	for _, existingFloor := range e.pendingRequests {
		if existingFloor == floor {
			return // Floor already requested, skip
		}
	}

	e.pendingRequests = append(e.pendingRequests, floor)
	fmt.Printf("  📝 Elevator %d: Added floor %d to queue\n", e.id, floor)
}

// ProcessAllRequests processes all pending floor requests using SCAN algorithm
// The SCAN algorithm serves floors in one direction, then reverses (like an elevator!)
func (e *Elevator) ProcessAllRequests() {
	for {
		e.mutex.Lock()

		// No more requests - go back to idle state
		if len(e.pendingRequests) == 0 {
			e.state = StateIdle
			e.direction = DirectionIdle
			e.mutex.Unlock()
			return
		}

		// Sort requests to minimize direction changes (SCAN algorithm)
		e.sortRequestsForOptimalPath()

		// Get the next floor to visit (first in sorted queue)
		nextFloor := e.pendingRequests[0]
		e.pendingRequests = e.pendingRequests[1:] // Remove from queue

		// Determine which direction we need to go
		if nextFloor > e.currentFloor {
			e.direction = DirectionUp
		} else if nextFloor < e.currentFloor {
			e.direction = DirectionDown
		}
		// If nextFloor == currentFloor, keep current direction

		e.state = StateMoving
		e.mutex.Unlock()

		// Move to the next floor (this takes time)
		e.moveToFloor(nextFloor)
	}
}

// sortRequestsForOptimalPath sorts the pending requests to minimize direction changes
// This implements the SCAN (elevator) algorithm:
// - If going UP: serve all floors above first (ascending), then floors below (descending)
// - If going DOWN: serve all floors below first (descending), then floors above (ascending)
func (e *Elevator) sortRequestsForOptimalPath() {
	var floorsAbove, floorsBelow []int

	// Separate requests into floors above and below current position
	for _, floor := range e.pendingRequests {
		if floor >= e.currentFloor {
			floorsAbove = append(floorsAbove, floor)
		} else {
			floorsBelow = append(floorsBelow, floor)
		}
	}

	// Sort floors above in ascending order (1, 2, 3, ...)
	sort.Ints(floorsAbove)

	// Sort floors below in descending order (5, 4, 3, ...)
	sort.Sort(sort.Reverse(sort.IntSlice(floorsBelow)))

	// Combine based on current direction
	if e.direction == DirectionUp || e.direction == DirectionIdle {
		// Going UP: serve above first, then below
		e.pendingRequests = make([]int, 0, len(floorsAbove)+len(floorsBelow))
		e.pendingRequests = append(e.pendingRequests, floorsAbove...)
		e.pendingRequests = append(e.pendingRequests, floorsBelow...)
	} else {
		// Going DOWN: serve below first, then above
		e.pendingRequests = make([]int, 0, len(floorsAbove)+len(floorsBelow))
		e.pendingRequests = append(e.pendingRequests, floorsBelow...)
		e.pendingRequests = append(e.pendingRequests, floorsAbove...)
	}
}

// moveToFloor moves the elevator from current floor to the target floor
// This simulates the physical movement floor by floor
func (e *Elevator) moveToFloor(targetFloor int) {
	e.mutex.Lock()
	startFloor := e.currentFloor
	e.mutex.Unlock()

	// Determine direction of movement
	step := 1 // Going up
	if targetFloor < startFloor {
		step = -1 // Going down
	}

	// Simulate moving through each floor one at a time
	for currentPosition := startFloor; currentPosition != targetFloor; currentPosition += step {
		time.Sleep(FloorTravelTime) // Simulate travel time between floors

		e.mutex.Lock()
		e.currentFloor = currentPosition + step
		e.mutex.Unlock()
	}

	// Arrived at destination floor
	e.mutex.Lock()
	e.state = StateStopped
	e.currentFloor = targetFloor
	e.mutex.Unlock()

	fmt.Printf("  🔔 Elevator %d arrived at floor %d\n", e.id, targetFloor)

	// Simulate doors opening and closing
	time.Sleep(DoorOpenTime)
}

// SetMaintenanceMode enables or disables maintenance mode for this elevator
func (e *Elevator) SetMaintenanceMode(enabled bool) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if enabled {
		e.state = StateMaintenance
	} else {
		e.state = StateIdle
	}
}

// String returns a human-readable status of the elevator
func (e *Elevator) String() string {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	return fmt.Sprintf("Elevator %d: Floor %d, %s, %s, Load: %d/%d",
		e.id, e.currentFloor, e.direction, e.state, e.currentLoad, e.maxCapacity)
}

// ============================================================
// SCHEDULING STRATEGY - Interface for elevator selection algorithms
// ============================================================

// SchedulingStrategy defines how to select an elevator for a floor request
// This is the Strategy Pattern - different algorithms can be swapped at runtime
type SchedulingStrategy interface {
	// SelectElevator chooses the best elevator for the given floor and direction
	SelectElevator(elevators []*Elevator, requestedFloor int, requestedDirection Direction) *Elevator
	// GetName returns the name of this strategy for display purposes
	GetName() string
}

// ============================================================
// NEAREST ELEVATOR STRATEGY - Selects closest available elevator
// ============================================================

// NearestElevatorStrategy selects the elevator that is closest to the requested floor
// It also considers the direction to prefer elevators already going the same way
type NearestElevatorStrategy struct{}

// SelectElevator finds the best elevator based on distance and direction
func (s *NearestElevatorStrategy) SelectElevator(elevators []*Elevator, requestedFloor int, requestedDirection Direction) *Elevator {
	var bestElevator *Elevator
	bestScore := math.MaxInt // Lower score is better

	for _, elevator := range elevators {
		// Skip elevators that aren't available
		if !elevator.IsAvailable() {
			continue
		}

		// Calculate base distance (number of floors away)
		currentFloor := elevator.GetCurrentFloor()
		score := absoluteValue(requestedFloor - currentFloor)

		// Adjust score based on elevator's current direction
		elevatorDirection := elevator.GetDirection()
		if elevatorDirection != DirectionIdle {
			// Give bonus if elevator is going same direction and will pass the floor
			if requestedDirection == DirectionUp && elevatorDirection == DirectionUp && currentFloor <= requestedFloor {
				score -= SameDirectionBonus // Bonus: elevator will naturally pass this floor
			} else if requestedDirection == DirectionDown && elevatorDirection == DirectionDown && currentFloor >= requestedFloor {
				score -= SameDirectionBonus // Bonus: elevator will naturally pass this floor
			} else {
				score += OppositeDirectionPenalty // Penalty: elevator going wrong way
			}
		}

		// Keep track of the best (lowest score) elevator
		if score < bestScore {
			bestScore = score
			bestElevator = elevator
		}
	}

	return bestElevator
}

// GetName returns the strategy name
func (s *NearestElevatorStrategy) GetName() string {
	return "Nearest Elevator Strategy"
}

// ============================================================
// ROUND ROBIN STRATEGY - Distributes requests evenly among elevators
// ============================================================

// RoundRobinStrategy distributes requests evenly by cycling through elevators
// Also known as FCFS (First Come First Serve) in some contexts
type RoundRobinStrategy struct {
	lastUsedIndex int // Tracks which elevator was last assigned
}

// SelectElevator picks the next available elevator in round-robin order
func (s *RoundRobinStrategy) SelectElevator(elevators []*Elevator, requestedFloor int, requestedDirection Direction) *Elevator {
	elevatorCount := len(elevators)

	// Try each elevator starting from the one after last used
	for i := 0; i < elevatorCount; i++ {
		index := (s.lastUsedIndex + i + 1) % elevatorCount
		if elevators[index].IsAvailable() {
			s.lastUsedIndex = index
			return elevators[index]
		}
	}

	return nil // No available elevator found
}

// GetName returns the strategy name
func (s *RoundRobinStrategy) GetName() string {
	return "Round Robin Strategy"
}

// ============================================================
// ELEVATOR CONTROLLER - Manages all elevators in the system
// ============================================================

// ElevatorController coordinates all elevators and handles floor requests
type ElevatorController struct {
	elevators []*Elevator        // All elevators managed by this controller
	strategy  SchedulingStrategy // Algorithm for selecting elevators
	mutex     sync.Mutex         // Protects concurrent access
}

// NewElevatorController creates a new controller with the given elevators and strategy
func NewElevatorController(elevators []*Elevator, strategy SchedulingStrategy) *ElevatorController {
	return &ElevatorController{
		elevators: elevators,
		strategy:  strategy,
	}
}

// HandleExternalRequest handles a request from a floor button (outside elevator)
// Returns the assigned elevator or an error if none available
func (c *ElevatorController) HandleExternalRequest(floor int, direction Direction) (*Elevator, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Use the scheduling strategy to select the best elevator
	selectedElevator := c.strategy.SelectElevator(c.elevators, floor, direction)
	if selectedElevator == nil {
		return nil, fmt.Errorf("no elevator available")
	}

	fmt.Printf("📍 Floor %d requested %s - Assigned to Elevator %d\n",
		floor, direction, selectedElevator.GetID())

	// Add this floor to the elevator's request queue
	selectedElevator.AddFloorRequest(floor)

	// Start processing requests in a background goroutine
	go selectedElevator.ProcessAllRequests()

	return selectedElevator, nil
}

// HandleInternalRequest handles a request from inside an elevator (floor button press)
func (c *ElevatorController) HandleInternalRequest(elevatorID, floor int) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Find the elevator with the given ID
	for _, elevator := range c.elevators {
		if elevator.GetID() == elevatorID {
			fmt.Printf("📍 Elevator %d: Floor %d button pressed\n", elevatorID, floor)
			elevator.AddFloorRequest(floor)
			go elevator.ProcessAllRequests()
			return nil
		}
	}

	return fmt.Errorf("elevator %d not found", elevatorID)
}

// GetSystemStatus returns a formatted string showing all elevator statuses
func (c *ElevatorController) GetSystemStatus() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	status := "\n╔══════════════════════════════════════════╗\n"
	status += "║           ELEVATOR STATUS                ║\n"
	status += "╠══════════════════════════════════════════╣\n"

	for _, elevator := range c.elevators {
		status += fmt.Sprintf("║ %s\n", elevator.String())
	}
	status += "╚══════════════════════════════════════════╝\n"

	return status
}

// SetSchedulingStrategy changes the elevator selection algorithm at runtime
func (c *ElevatorController) SetSchedulingStrategy(strategy SchedulingStrategy) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.strategy = strategy
}

// ============================================================
// BUILDING - Main facade for the elevator system
// ============================================================

// Building represents a building with elevators
// This is the main entry point for users of the elevator system
type Building struct {
	name       string              // Name of the building
	minFloor   int                 // Lowest floor number
	maxFloor   int                 // Highest floor number
	controller *ElevatorController // Controller managing all elevators
}

// NewBuilding creates a new building with the specified configuration
func NewBuilding(name string, minFloor, maxFloor, numberOfElevators, elevatorCapacity int) *Building {
	// Create all elevators
	elevators := make([]*Elevator, numberOfElevators)
	for i := 0; i < numberOfElevators; i++ {
		elevators[i] = NewElevator(i+1, minFloor, maxFloor, elevatorCapacity)
	}

	// Create the controller with default strategy (nearest elevator)
	controller := NewElevatorController(elevators, &NearestElevatorStrategy{})

	return &Building{
		name:       name,
		minFloor:   minFloor,
		maxFloor:   maxFloor,
		controller: controller,
	}
}

// CallElevator requests an elevator to a specific floor
// This is used when someone presses the up/down button on a floor
func (b *Building) CallElevator(floor int, direction Direction) (*Elevator, error) {
	// Validate the floor number
	if floor < b.minFloor || floor > b.maxFloor {
		return nil, fmt.Errorf("invalid floor %d (must be between %d and %d)", floor, b.minFloor, b.maxFloor)
	}
	return b.controller.HandleExternalRequest(floor, direction)
}

// SelectFloor is used when someone inside an elevator presses a floor button
func (b *Building) SelectFloor(elevatorID, floor int) error {
	// Validate the floor number
	if floor < b.minFloor || floor > b.maxFloor {
		return fmt.Errorf("invalid floor %d (must be between %d and %d)", floor, b.minFloor, b.maxFloor)
	}
	return b.controller.HandleInternalRequest(elevatorID, floor)
}

// GetStatus returns the current status of all elevators
func (b *Building) GetStatus() string {
	return b.controller.GetSystemStatus()
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// absoluteValue returns the absolute value of an integer
func absoluteValue(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ============================================================
// MAIN FUNCTION - Demonstrates the elevator system
// ============================================================

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("       ELEVATOR SYSTEM - LLD DEMO")
	fmt.Println("═══════════════════════════════════════════")

	// Create a building with:
	// - Floors 0 (ground) to 10
	// - 3 elevators
	// - Each elevator can hold 10 people
	building := NewBuilding("Tech Tower", 0, 10, 3, 10)

	// Display initial status
	fmt.Println(building.GetStatus())

	// ========== SCENARIO 1: Multiple floor requests ==========
	fmt.Println("\n📌 SCENARIO 1: Multiple floor requests")
	fmt.Println("─────────────────────────────────────────")

	// Person on floor 5 wants to go down
	_, _ = building.CallElevator(5, DirectionDown)
	time.Sleep(100 * time.Millisecond)

	// Person on floor 8 wants to go down
	_, _ = building.CallElevator(8, DirectionDown)
	time.Sleep(100 * time.Millisecond)

	// Person on floor 2 wants to go up
	_, _ = building.CallElevator(2, DirectionUp)
	time.Sleep(100 * time.Millisecond)

	// Wait for elevators to process requests
	time.Sleep(2 * time.Second)
	fmt.Println(building.GetStatus())

	// ========== SCENARIO 2: Internal floor selection ==========
	fmt.Println("\n📌 SCENARIO 2: Internal floor selection")
	fmt.Println("─────────────────────────────────────────")

	// Person inside elevator 1 presses button for floor 10
	_ = building.SelectFloor(1, 10)

	// Person inside elevator 2 presses button for floor 0 (ground)
	_ = building.SelectFloor(2, 0)

	time.Sleep(3 * time.Second)
	fmt.Println(building.GetStatus())

	// ========== SCENARIO 3: Rush hour simulation ==========
	fmt.Println("\n📌 SCENARIO 3: Rush hour simulation")
	fmt.Println("─────────────────────────────────────────")

	// Multiple people calling elevators from different floors
	for floor := 1; floor <= 5; floor++ {
		_, _ = building.CallElevator(floor, DirectionUp)
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(4 * time.Second)
	fmt.Println(building.GetStatus())

	// ========== Summary of Design Patterns ==========
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  KEY DESIGN PATTERNS USED:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. State Pattern  - Elevator states (Idle, Moving, Stopped, Maintenance)")
	fmt.Println("  2. Strategy Pattern - Scheduling algorithms (Nearest, RoundRobin)")
	fmt.Println("  3. Facade Pattern - Building provides simple interface")
	fmt.Println("  4. SCAN Algorithm - Efficient floor serving (elevator algorithm)")
	fmt.Println("═══════════════════════════════════════════")
}
