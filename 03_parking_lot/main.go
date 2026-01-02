package main

import (
	"fmt"
	"time"
)

// ============================================================
// PARKING LOT SYSTEM - LOW LEVEL DESIGN DEMO
// ============================================================
//
// This file demonstrates a complete parking lot management system.
// It covers key design patterns and SOLID principles.
//
// KEY CONCEPTS COVERED:
// - Interface-based design (Vehicle interface)
// - Strategy Pattern (FeeCalculator, PaymentMethod)
// - Single Responsibility Principle (each struct has one job)
// - Composition (ParkingLot contains Floors, Floor contains Spots)
//
// Run: go run .
// ============================================================

// ============================================================
// SECTION 1: VEHICLE TYPES AND SPOT SIZES
// ============================================================

// VehicleType represents the type of vehicle (Motorcycle, Car, Truck)
// Using iota for automatic enumeration (0, 1, 2)
type VehicleType int

const (
	VehicleTypeMotorcycle VehicleType = iota // 0 - Smallest vehicle
	VehicleTypeCar                           // 1 - Medium vehicle
	VehicleTypeTruck                         // 2 - Largest vehicle
)

// String converts VehicleType to a human-readable string
func (vehicleType VehicleType) String() string {
	switch vehicleType {
	case VehicleTypeMotorcycle:
		return "Motorcycle"
	case VehicleTypeCar:
		return "Car"
	case VehicleTypeTruck:
		return "Truck"
	default:
		return "Unknown"
	}
}

// SpotSize represents the size of a parking spot
// Larger spots can accommodate smaller vehicles
type SpotSize int

const (
	SpotSizeSmall  SpotSize = iota // 0 - For motorcycles
	SpotSizeMedium                 // 1 - For cars
	SpotSizeLarge                  // 2 - For trucks
)

// String converts SpotSize to a human-readable string
func (spotSize SpotSize) String() string {
	switch spotSize {
	case SpotSizeSmall:
		return "Small"
	case SpotSizeMedium:
		return "Medium"
	case SpotSizeLarge:
		return "Large"
	default:
		return "Unknown"
	}
}

// CanFit checks if this spot size can accommodate a vehicle of given size
// Rule: A larger spot can fit a smaller vehicle (e.g., truck spot can fit a car)
func (spotSize SpotSize) CanFit(requiredSize SpotSize) bool {
	return spotSize >= requiredSize
}

// ============================================================
// SECTION 2: VEHICLE INTERFACE AND IMPLEMENTATIONS
// ============================================================

// Vehicle is the interface that all vehicle types must implement
// This allows the parking lot to work with any type of vehicle
type Vehicle interface {
	GetType() VehicleType          // Returns the type of vehicle
	GetLicensePlate() string       // Returns the unique license plate
	GetRequiredSpotSize() SpotSize // Returns the minimum spot size needed
}

// -------------------- Motorcycle --------------------

// Motorcycle represents a two-wheeler vehicle
type Motorcycle struct {
	licensePlate string // Unique identifier for the motorcycle
}

// NewMotorcycle creates a new Motorcycle instance
func NewMotorcycle(licensePlate string) *Motorcycle {
	return &Motorcycle{licensePlate: licensePlate}
}

// GetType returns VehicleTypeMotorcycle
func (motorcycle *Motorcycle) GetType() VehicleType {
	return VehicleTypeMotorcycle
}

// GetLicensePlate returns the motorcycle's license plate
func (motorcycle *Motorcycle) GetLicensePlate() string {
	return motorcycle.licensePlate
}

// GetRequiredSpotSize returns SpotSizeSmall (motorcycles need small spots)
func (motorcycle *Motorcycle) GetRequiredSpotSize() SpotSize {
	return SpotSizeSmall
}

// -------------------- Car --------------------

// Car represents a standard four-wheeler vehicle
type Car struct {
	licensePlate string // Unique identifier for the car
}

// NewCar creates a new Car instance
func NewCar(licensePlate string) *Car {
	return &Car{licensePlate: licensePlate}
}

// GetType returns VehicleTypeCar
func (car *Car) GetType() VehicleType {
	return VehicleTypeCar
}

// GetLicensePlate returns the car's license plate
func (car *Car) GetLicensePlate() string {
	return car.licensePlate
}

// GetRequiredSpotSize returns SpotSizeMedium (cars need medium spots)
func (car *Car) GetRequiredSpotSize() SpotSize {
	return SpotSizeMedium
}

// -------------------- Truck --------------------

// Truck represents a large commercial vehicle
type Truck struct {
	licensePlate string // Unique identifier for the truck
}

// NewTruck creates a new Truck instance
func NewTruck(licensePlate string) *Truck {
	return &Truck{licensePlate: licensePlate}
}

// GetType returns VehicleTypeTruck
func (truck *Truck) GetType() VehicleType {
	return VehicleTypeTruck
}

// GetLicensePlate returns the truck's license plate
func (truck *Truck) GetLicensePlate() string {
	return truck.licensePlate
}

// GetRequiredSpotSize returns SpotSizeLarge (trucks need large spots)
func (truck *Truck) GetRequiredSpotSize() SpotSize {
	return SpotSizeLarge
}

// ============================================================
// SECTION 3: PARKING SPOT
// ============================================================

// ParkingSpot represents a single parking space in the lot
type ParkingSpot struct {
	spotID        string   // Unique ID like "F1-S1" (Floor 1, Spot 1)
	floorNumber   int      // Which floor this spot is on
	spotNumber    int      // Spot number on this floor
	size          SpotSize // Size of this spot (small/medium/large)
	parkedVehicle Vehicle  // Currently parked vehicle (nil if empty)
}

// NewParkingSpot creates a new parking spot with given parameters
func NewParkingSpot(floorNumber, spotNumber int, size SpotSize) *ParkingSpot {
	return &ParkingSpot{
		spotID:        fmt.Sprintf("F%d-S%d", floorNumber, spotNumber),
		floorNumber:   floorNumber,
		spotNumber:    spotNumber,
		size:          size,
		parkedVehicle: nil, // Initially empty
	}
}

// GetID returns the unique spot identifier
func (spot *ParkingSpot) GetID() string {
	return spot.spotID
}

// GetFloorNumber returns which floor this spot is on
func (spot *ParkingSpot) GetFloorNumber() int {
	return spot.floorNumber
}

// GetSize returns the size of this parking spot
func (spot *ParkingSpot) GetSize() SpotSize {
	return spot.size
}

// IsAvailable checks if the spot is empty (no vehicle parked)
func (spot *ParkingSpot) IsAvailable() bool {
	return spot.parkedVehicle == nil
}

// GetVehicle returns the currently parked vehicle (nil if empty)
func (spot *ParkingSpot) GetVehicle() Vehicle {
	return spot.parkedVehicle
}

// CanPark checks if a given vehicle can park in this spot
// Conditions: spot must be empty AND spot size must fit the vehicle
func (spot *ParkingSpot) CanPark(vehicle Vehicle) bool {
	isSpotEmpty := spot.parkedVehicle == nil
	canFitVehicle := spot.size.CanFit(vehicle.GetRequiredSpotSize())
	return isSpotEmpty && canFitVehicle
}

// Park places a vehicle in this spot
// Returns an error if the vehicle cannot be parked here
func (spot *ParkingSpot) Park(vehicle Vehicle) error {
	if !spot.CanPark(vehicle) {
		return fmt.Errorf("cannot park vehicle in spot %s: spot is occupied or too small", spot.spotID)
	}
	spot.parkedVehicle = vehicle
	return nil
}

// Unpark removes and returns the vehicle from this spot
// Returns the vehicle that was parked (nil if spot was already empty)
func (spot *ParkingSpot) Unpark() Vehicle {
	removedVehicle := spot.parkedVehicle
	spot.parkedVehicle = nil
	return removedVehicle
}

// ============================================================
// SECTION 4: FLOOR
// ============================================================

// Floor represents one level of the parking lot
// Each floor contains multiple parking spots of different sizes
type Floor struct {
	floorNumber int            // Floor number (1, 2, 3, etc.)
	spots       []*ParkingSpot // All parking spots on this floor
}

// NewFloor creates a new floor with specified number of spots for each size
// Parameters:
//   - floorNumber: which floor (1, 2, 3, etc.)
//   - smallSpotCount: number of small spots (for motorcycles)
//   - mediumSpotCount: number of medium spots (for cars)
//   - largeSpotCount: number of large spots (for trucks)
func NewFloor(floorNumber, smallSpotCount, mediumSpotCount, largeSpotCount int) *Floor {
	floor := &Floor{
		floorNumber: floorNumber,
		spots:       make([]*ParkingSpot, 0),
	}

	currentSpotNumber := 1

	// Create small spots first
	for i := 0; i < smallSpotCount; i++ {
		newSpot := NewParkingSpot(floorNumber, currentSpotNumber, SpotSizeSmall)
		floor.spots = append(floor.spots, newSpot)
		currentSpotNumber++
	}

	// Create medium spots
	for i := 0; i < mediumSpotCount; i++ {
		newSpot := NewParkingSpot(floorNumber, currentSpotNumber, SpotSizeMedium)
		floor.spots = append(floor.spots, newSpot)
		currentSpotNumber++
	}

	// Create large spots
	for i := 0; i < largeSpotCount; i++ {
		newSpot := NewParkingSpot(floorNumber, currentSpotNumber, SpotSizeLarge)
		floor.spots = append(floor.spots, newSpot)
		currentSpotNumber++
	}

	return floor
}

// FindAvailableSpot finds a suitable parking spot for the given vehicle
// Strategy: First try to find exact size match, then try larger spots
// This optimization prevents wasting large spots on small vehicles
func (floor *Floor) FindAvailableSpot(vehicle Vehicle) *ParkingSpot {
	requiredSize := vehicle.GetRequiredSpotSize()

	// First pass: Look for exact size match (best fit)
	for _, spot := range floor.spots {
		if spot.CanPark(vehicle) && spot.GetSize() == requiredSize {
			return spot
		}
	}

	// Second pass: Look for any spot that can fit (larger spot is okay)
	for _, spot := range floor.spots {
		if spot.CanPark(vehicle) {
			return spot
		}
	}

	// No suitable spot found on this floor
	return nil
}

// GetAvailableSpotCount returns the count of available spots of a specific size
func (floor *Floor) GetAvailableSpotCount(spotSize SpotSize) int {
	availableCount := 0
	for _, spot := range floor.spots {
		if spot.IsAvailable() && spot.GetSize() == spotSize {
			availableCount++
		}
	}
	return availableCount
}

// ============================================================
// SECTION 5: PARKING TICKET
// ============================================================

// Ticket represents a parking ticket issued when a vehicle enters
type Ticket struct {
	ticketID     string       // Unique ticket ID like "TKT-1"
	vehiclePlate string       // License plate of the parked vehicle
	vehicleType  VehicleType  // Type of vehicle
	assignedSpot *ParkingSpot // Which spot the vehicle is parked in
	entryTime    time.Time    // When the vehicle entered
	exitTime     time.Time    // When the vehicle exited (zero if still parked)
	amountPaid   float64      // Amount paid (0 if not paid yet)
	isPaid       bool         // Whether payment has been made
}

// ticketCounter is used to generate unique ticket IDs
// Note: In production, use a proper ID generator or database sequence
var ticketCounter int = 0

// NewTicket creates a new parking ticket for a vehicle
func NewTicket(vehicle Vehicle, spot *ParkingSpot) *Ticket {
	ticketCounter++
	return &Ticket{
		ticketID:     fmt.Sprintf("TKT-%d", ticketCounter),
		vehiclePlate: vehicle.GetLicensePlate(),
		vehicleType:  vehicle.GetType(),
		assignedSpot: spot,
		entryTime:    time.Now(),
		// exitTime, amountPaid, isPaid are zero/false by default
	}
}

// GetParkingDurationHours calculates how long the vehicle has been parked
// Returns at least 1 hour (minimum billing)
func (ticket *Ticket) GetParkingDurationHours() int {
	var parkingDuration time.Duration

	// If vehicle hasn't exited yet, calculate duration from entry until now
	if ticket.exitTime.IsZero() {
		parkingDuration = time.Since(ticket.entryTime)
	} else {
		// Vehicle has exited, use the recorded exit time
		parkingDuration = ticket.exitTime.Sub(ticket.entryTime)
	}

	// Convert to hours (minimum 1 hour billing)
	hours := int(parkingDuration.Hours())
	if hours < 1 {
		hours = 1 // Minimum charge is 1 hour
	}
	return hours
}

// RecordExit marks the exit time when vehicle leaves
func (ticket *Ticket) RecordExit() {
	ticket.exitTime = time.Now()
}

// RecordPayment marks the ticket as paid with the given amount
func (ticket *Ticket) RecordPayment(amount float64) {
	ticket.amountPaid = amount
	ticket.isPaid = true
}

// ============================================================
// SECTION 6: FEE CALCULATOR (Strategy Pattern)
// ============================================================
// The Strategy Pattern allows us to define different fee calculation
// algorithms and switch between them at runtime. This makes the system
// flexible - we can easily add new pricing strategies without changing
// existing code.

// FeeCalculator is the interface for fee calculation strategies
type FeeCalculator interface {
	CalculateFee(ticket *Ticket) float64
}

// HourlyRateCalculator calculates fee based on hourly rates per vehicle type
type HourlyRateCalculator struct {
	hourlyRates map[VehicleType]float64 // Rate per hour for each vehicle type
}

// NewHourlyRateCalculator creates a calculator with default hourly rates
// Rates: Motorcycle=$1/hr, Car=$2/hr, Truck=$3/hr
func NewHourlyRateCalculator() *HourlyRateCalculator {
	return &HourlyRateCalculator{
		hourlyRates: map[VehicleType]float64{
			VehicleTypeMotorcycle: 1.0, // $1 per hour
			VehicleTypeCar:        2.0, // $2 per hour
			VehicleTypeTruck:      3.0, // $3 per hour
		},
	}
}

// CalculateFee calculates the total fee based on duration and vehicle type
func (calculator *HourlyRateCalculator) CalculateFee(ticket *Ticket) float64 {
	parkingHours := ticket.GetParkingDurationHours()
	hourlyRate := calculator.hourlyRates[ticket.vehicleType]
	totalFee := float64(parkingHours) * hourlyRate
	return totalFee
}

// ============================================================
// SECTION 7: PAYMENT METHOD (Strategy Pattern)
// ============================================================
// Another use of Strategy Pattern for handling different payment methods.
// This allows easy addition of new payment options (UPI, Wallet, etc.)

// PaymentMethod is the interface for different payment options
type PaymentMethod interface {
	ProcessPayment(amount float64) error
}

// CashPayment handles cash payments
type CashPayment struct{}

// ProcessPayment processes a cash payment
func (payment *CashPayment) ProcessPayment(amount float64) error {
	fmt.Printf("  [Cash Payment] Amount received: $%.2f\n", amount)
	return nil
}

// CardPayment handles credit/debit card payments
type CardPayment struct {
	cardNumber string // Full card number (would be encrypted in production)
}

// NewCardPayment creates a new card payment with the given card number
func NewCardPayment(cardNumber string) *CardPayment {
	return &CardPayment{cardNumber: cardNumber}
}

// ProcessPayment processes a card payment
// Shows only last 4 digits for security
func (payment *CardPayment) ProcessPayment(amount float64) error {
	// Validate card number length to avoid panic
	if len(payment.cardNumber) < 4 {
		return fmt.Errorf("invalid card number")
	}

	// Show only last 4 digits for security
	lastFourDigits := payment.cardNumber[len(payment.cardNumber)-4:]
	fmt.Printf("  [Card Payment] Amount charged: $%.2f (Card: ****%s)\n", amount, lastFourDigits)
	return nil
}

// ============================================================
// SECTION 8: PARKING LOT (Main Controller)
// ============================================================

// ParkingLot is the main class that manages the entire parking system
type ParkingLot struct {
	name          string             // Name of the parking lot
	floors        []*Floor           // All floors in the parking lot
	activeTickets map[string]*Ticket // Maps license plate -> active ticket
	feeCalculator FeeCalculator      // Strategy for calculating fees
}

// FloorConfig defines the configuration for one floor
// [0] = small spots, [1] = medium spots, [2] = large spots
type FloorConfig [3]int

// NewParkingLot creates a new parking lot with the given configuration
// Parameters:
//   - name: Name of the parking lot
//   - floorsConfig: Array of FloorConfig, one for each floor
//     Each FloorConfig is [smallSpots, mediumSpots, largeSpots]
func NewParkingLot(name string, floorsConfig []FloorConfig) *ParkingLot {
	parkingLot := &ParkingLot{
		name:          name,
		floors:        make([]*Floor, 0),
		activeTickets: make(map[string]*Ticket),
		feeCalculator: NewHourlyRateCalculator(), // Default fee calculator
	}

	// Create floors based on configuration
	for floorIndex, config := range floorsConfig {
		floorNumber := floorIndex + 1 // Floors are 1-indexed
		smallSpots := config[0]
		mediumSpots := config[1]
		largeSpots := config[2]

		newFloor := NewFloor(floorNumber, smallSpots, mediumSpots, largeSpots)
		parkingLot.floors = append(parkingLot.floors, newFloor)
	}

	return parkingLot
}

// ParkVehicle parks a vehicle and returns a ticket
// Returns an error if:
//   - Vehicle is already parked
//   - No suitable spot is available
func (lot *ParkingLot) ParkVehicle(vehicle Vehicle) (*Ticket, error) {
	licensePlate := vehicle.GetLicensePlate()

	// Check if this vehicle is already parked
	if _, alreadyParked := lot.activeTickets[licensePlate]; alreadyParked {
		return nil, fmt.Errorf("vehicle %s is already parked in the lot", licensePlate)
	}

	// Find an available spot across all floors
	var availableSpot *ParkingSpot
	for _, floor := range lot.floors {
		availableSpot = floor.FindAvailableSpot(vehicle)
		if availableSpot != nil {
			break // Found a spot, stop searching
		}
	}

	// No spot found
	if availableSpot == nil {
		return nil, fmt.Errorf("no parking spot available for %s", vehicle.GetType())
	}

	// Park the vehicle in the found spot
	if err := availableSpot.Park(vehicle); err != nil {
		return nil, err
	}

	// Create and store the ticket
	ticket := NewTicket(vehicle, availableSpot)
	lot.activeTickets[licensePlate] = ticket

	fmt.Printf("  [PARKED] %s (%s) -> Spot %s\n",
		licensePlate, vehicle.GetType(), availableSpot.GetID())

	return ticket, nil
}

// UnparkVehicle removes a vehicle, calculates fee, processes payment
// Returns the completed ticket or an error if vehicle not found
func (lot *ParkingLot) UnparkVehicle(licensePlate string, paymentMethod PaymentMethod) (*Ticket, error) {
	// Find the ticket for this vehicle
	ticket, exists := lot.activeTickets[licensePlate]
	if !exists {
		return nil, fmt.Errorf("vehicle %s is not found in the parking lot", licensePlate)
	}

	// Record exit time and calculate fee
	ticket.RecordExit()
	parkingFee := lot.feeCalculator.CalculateFee(ticket)

	// Process payment
	if err := paymentMethod.ProcessPayment(parkingFee); err != nil {
		return nil, fmt.Errorf("payment failed: %v", err)
	}
	ticket.RecordPayment(parkingFee)

	// Free up the parking spot
	ticket.assignedSpot.Unpark()

	// Remove from active tickets
	delete(lot.activeTickets, licensePlate)

	fmt.Printf("  [EXITED] %s - Total Paid: $%.2f\n", licensePlate, parkingFee)

	return ticket, nil
}

// DisplayAvailability shows the current availability of parking spots
func (lot *ParkingLot) DisplayAvailability() {
	fmt.Println()
	fmt.Println("+----------------------------------------------------+")
	fmt.Printf("|  %-48s  |\n", lot.name+" - AVAILABILITY")
	fmt.Println("+----------------------------------------------------+")

	for _, floor := range lot.floors {
		smallAvailable := floor.GetAvailableSpotCount(SpotSizeSmall)
		mediumAvailable := floor.GetAvailableSpotCount(SpotSizeMedium)
		largeAvailable := floor.GetAvailableSpotCount(SpotSizeLarge)

		fmt.Printf("|  Floor %d: Motorcycle: %2d  Car: %2d  Truck: %2d       |\n",
			floor.floorNumber, smallAvailable, mediumAvailable, largeAvailable)
	}
	fmt.Println("+----------------------------------------------------+")
}

// ============================================================
// SECTION 9: MAIN FUNCTION - DEMO
// ============================================================

func main() {
	fmt.Println("=================================================")
	fmt.Println("     PARKING LOT SYSTEM - LOW LEVEL DESIGN DEMO")
	fmt.Println("=================================================")
	fmt.Println()

	// ----- Step 1: Create the Parking Lot -----
	// Configuration: 2 floors
	// Each floor has: 5 small spots, 10 medium spots, 3 large spots
	parkingLotConfig := []FloorConfig{
		{5, 10, 3}, // Floor 1: 5 small, 10 medium, 3 large
		{5, 10, 3}, // Floor 2: 5 small, 10 medium, 3 large
	}

	parkingLot := NewParkingLot("City Center Parking", parkingLotConfig)

	// Show initial state
	fmt.Println(">>> Initial Parking Lot State:")
	parkingLot.DisplayAvailability()

	// ----- Step 2: Park Some Vehicles -----
	fmt.Println("\n>>> Parking Vehicles...")

	motorcycle1 := NewMotorcycle("BIKE-001")
	if _, err := parkingLot.ParkVehicle(motorcycle1); err != nil {
		fmt.Printf("  [ERROR] %v\n", err)
	}

	car1 := NewCar("CAR-1234")
	if _, err := parkingLot.ParkVehicle(car1); err != nil {
		fmt.Printf("  [ERROR] %v\n", err)
	}

	car2 := NewCar("CAR-5678")
	if _, err := parkingLot.ParkVehicle(car2); err != nil {
		fmt.Printf("  [ERROR] %v\n", err)
	}

	truck1 := NewTruck("TRUCK-01")
	if _, err := parkingLot.ParkVehicle(truck1); err != nil {
		fmt.Printf("  [ERROR] %v\n", err)
	}

	// Show state after parking
	fmt.Println("\n>>> After Parking 4 Vehicles:")
	parkingLot.DisplayAvailability()

	// ----- Step 3: Try Parking Same Vehicle Again (Error Case) -----
	fmt.Println("\n>>> Testing: Try to park same vehicle again...")
	_, err := parkingLot.ParkVehicle(car1)
	if err != nil {
		fmt.Printf("  [ERROR] %v\n", err)
	}

	// ----- Step 4: Exit Some Vehicles (Process Payments) -----
	fmt.Println("\n>>> Vehicles Exiting...")

	// Car exits with cash payment
	if _, err := parkingLot.UnparkVehicle("CAR-1234", &CashPayment{}); err != nil {
		fmt.Printf("  [ERROR] %v\n", err)
	}

	// Truck exits with card payment
	if _, err := parkingLot.UnparkVehicle("TRUCK-01", NewCardPayment("4111222233334444")); err != nil {
		fmt.Printf("  [ERROR] %v\n", err)
	}

	// Show state after exits
	fmt.Println("\n>>> After 2 Vehicles Exited:")
	parkingLot.DisplayAvailability()

	// ----- Step 5: More Vehicles Arrive -----
	fmt.Println("\n>>> More Vehicles Arriving...")

	if _, err := parkingLot.ParkVehicle(NewCar("CAR-9999")); err != nil {
		fmt.Printf("  [ERROR] %v\n", err)
	}
	if _, err := parkingLot.ParkVehicle(NewMotorcycle("BIKE-002")); err != nil {
		fmt.Printf("  [ERROR] %v\n", err)
	}
	if _, err := parkingLot.ParkVehicle(NewTruck("TRUCK-02")); err != nil {
		fmt.Printf("  [ERROR] %v\n", err)
	}

	// Show final state
	fmt.Println("\n>>> Final Parking Lot State:")
	parkingLot.DisplayAvailability()

	// ----- Summary of Design Decisions -----
	fmt.Println()
	fmt.Println("=================================================")
	fmt.Println("  KEY DESIGN PATTERNS & PRINCIPLES USED:")
	fmt.Println("=================================================")
	fmt.Println("  1. Interface (Vehicle) - Open/Closed Principle")
	fmt.Println("     -> Easy to add new vehicle types without changes")
	fmt.Println()
	fmt.Println("  2. Strategy Pattern (FeeCalculator)")
	fmt.Println("     -> Flexible fee calculation algorithms")
	fmt.Println()
	fmt.Println("  3. Strategy Pattern (PaymentMethod)")
	fmt.Println("     -> Multiple payment options (Cash, Card, etc.)")
	fmt.Println()
	fmt.Println("  4. Single Responsibility Principle")
	fmt.Println("     -> Each struct has one well-defined job")
	fmt.Println()
	fmt.Println("  5. Composition over Inheritance")
	fmt.Println("     -> ParkingLot contains Floors contains Spots")
	fmt.Println("=================================================")
}
