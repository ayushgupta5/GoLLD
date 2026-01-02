package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// CAR RENTAL SYSTEM - Low Level Design
// ============================================================================
//
// This system demonstrates:
// - Entity Modeling (Vehicle, Customer, Reservation)
// - Reservation Lifecycle Management (Pending -> Confirmed -> PickedUp -> Returned)
// - Pricing Strategy with daily rates and extras
// - Location-based Fleet Management
// - Thread-safe operations using mutex locks
//
// ============================================================================

// ============================================================================
// SECTION 1: ENUMS (Type-safe constants)
// ============================================================================

// VehicleType represents different categories of vehicles available for rent.
// Using iota for auto-incrementing values (Bike=0, Car=1, SUV=2, etc.)
type VehicleType int

const (
	VehicleTypeBike   VehicleType = iota // 0 - Two-wheeler
	VehicleTypeCar                       // 1 - Standard sedan
	VehicleTypeSUV                       // 2 - Sport Utility Vehicle
	VehicleTypeLuxury                    // 3 - Premium vehicles
	VehicleTypeVan                       // 4 - Multi-passenger vehicle
)

// String returns a human-readable name for the vehicle type.
func (vehicleType VehicleType) String() string {
	names := [...]string{"Bike", "Car", "SUV", "Luxury", "Van"}
	if int(vehicleType) < len(names) {
		return names[vehicleType]
	}
	return "Unknown"
}

// DailyRate returns the base rental rate per day for each vehicle type.
func (vehicleType VehicleType) DailyRate() float64 {
	rates := [...]float64{15.0, 40.0, 60.0, 120.0, 80.0}
	if int(vehicleType) < len(rates) {
		return rates[vehicleType]
	}
	return 0.0
}

// VehicleStatus represents the current availability state of a vehicle.
type VehicleStatus int

const (
	VehicleStatusAvailable   VehicleStatus = iota // 0 - Ready for rental
	VehicleStatusRented                           // 1 - Currently rented out
	VehicleStatusMaintenance                      // 2 - Under maintenance
	VehicleStatusReserved                         // 3 - Reserved but not picked up yet
)

// String returns a human-readable name for the vehicle status.
func (status VehicleStatus) String() string {
	names := [...]string{"Available", "Rented", "Maintenance", "Reserved"}
	if int(status) < len(names) {
		return names[status]
	}
	return "Unknown"
}

// ReservationStatus represents the lifecycle state of a reservation.
type ReservationStatus int

const (
	ReservationStatusPending   ReservationStatus = iota // 0 - Newly created, awaiting confirmation
	ReservationStatusConfirmed                          // 1 - Confirmed, ready for pickup
	ReservationStatusPickedUp                           // 2 - Customer has the vehicle
	ReservationStatusReturned                           // 3 - Vehicle returned, rental complete
	ReservationStatusCancelled                          // 4 - Reservation was cancelled
)

// String returns a human-readable name for the reservation status.
func (status ReservationStatus) String() string {
	names := [...]string{"Pending", "Confirmed", "Picked Up", "Returned", "Cancelled"}
	if int(status) < len(names) {
		return names[status]
	}
	return "Unknown"
}

// ============================================================================
// SECTION 2: VEHICLE ENTITY
// ============================================================================

// Vehicle represents a rentable vehicle in the fleet.
// It contains all information about the vehicle and its current state.
type Vehicle struct {
	id           string        // Unique identifier for the vehicle
	licensePlate string        // License plate number (e.g., "ABC-123")
	make         string        // Manufacturer (e.g., "Toyota")
	model        string        // Model name (e.g., "Camry")
	year         int           // Manufacturing year
	vehicleType  VehicleType   // Category of vehicle
	status       VehicleStatus // Current availability status
	mileage      int           // Total miles driven (for tracking)
	fuelLevel    int           // Fuel percentage (0-100)
	dailyRate    float64       // Rental cost per day
	location     string        // Current location (e.g., "Airport")
	mutex        sync.Mutex    // Protects concurrent access to vehicle state
}

// NewVehicle creates and initializes a new Vehicle instance.
// The vehicle starts with status "Available" and full fuel tank.
func NewVehicle(id, licensePlate, make, model string, year int, vehicleType VehicleType, location string) *Vehicle {
	return &Vehicle{
		id:           id,
		licensePlate: licensePlate,
		make:         make,
		model:        model,
		year:         year,
		vehicleType:  vehicleType,
		status:       VehicleStatusAvailable,
		mileage:      0,
		fuelLevel:    100,
		dailyRate:    vehicleType.DailyRate(),
		location:     location,
	}
}

// Getter methods for Vehicle fields
// These provide controlled read access to private fields

func (vehicle *Vehicle) GetID() string           { return vehicle.id }
func (vehicle *Vehicle) GetLicensePlate() string { return vehicle.licensePlate }
func (vehicle *Vehicle) GetType() VehicleType    { return vehicle.vehicleType }
func (vehicle *Vehicle) GetDailyRate() float64   { return vehicle.dailyRate }
func (vehicle *Vehicle) GetLocation() string     { return vehicle.location }
func (vehicle *Vehicle) GetMake() string         { return vehicle.make }
func (vehicle *Vehicle) GetModel() string        { return vehicle.model }
func (vehicle *Vehicle) GetYear() int            { return vehicle.year }

// GetStatus returns the current status of the vehicle (thread-safe).
func (vehicle *Vehicle) GetStatus() VehicleStatus {
	vehicle.mutex.Lock()
	defer vehicle.mutex.Unlock()
	return vehicle.status
}

// SetStatus updates the vehicle status (thread-safe).
func (vehicle *Vehicle) SetStatus(newStatus VehicleStatus) {
	vehicle.mutex.Lock()
	defer vehicle.mutex.Unlock()
	vehicle.status = newStatus
}

// IsAvailable checks if the vehicle can be rented.
func (vehicle *Vehicle) IsAvailable() bool {
	return vehicle.GetStatus() == VehicleStatusAvailable
}

// String returns a formatted description of the vehicle.
func (vehicle *Vehicle) String() string {
	return fmt.Sprintf("%d %s %s (%s) - $%.2f/day - %s",
		vehicle.year, vehicle.make, vehicle.model,
		vehicle.vehicleType, vehicle.dailyRate, vehicle.status)
}

// ============================================================================
// SECTION 3: CUSTOMER ENTITY
// ============================================================================

// Customer represents a person who can rent vehicles.
type Customer struct {
	id            string         // Unique identifier
	name          string         // Full name
	email         string         // Contact email
	phone         string         // Contact phone number
	driverLicense string         // Driver's license number (required for rental)
	rentalHistory []*Reservation // History of past rentals
}

// NewCustomer creates and initializes a new Customer instance.
func NewCustomer(id, name, email, phone, driverLicense string) *Customer {
	return &Customer{
		id:            id,
		name:          name,
		email:         email,
		phone:         phone,
		driverLicense: driverLicense,
		rentalHistory: make([]*Reservation, 0),
	}
}

// Getter methods for Customer fields
func (customer *Customer) GetID() string            { return customer.id }
func (customer *Customer) GetName() string          { return customer.name }
func (customer *Customer) GetEmail() string         { return customer.email }
func (customer *Customer) GetPhone() string         { return customer.phone }
func (customer *Customer) GetDriverLicense() string { return customer.driverLicense }

// AddRentalToHistory adds a completed reservation to customer's history.
func (customer *Customer) AddRentalToHistory(reservation *Reservation) {
	customer.rentalHistory = append(customer.rentalHistory, reservation)
}

// ============================================================================
// SECTION 4: EXTRA (Add-on) ENTITY
// ============================================================================

// Extra represents an additional service/item that can be added to a rental.
// Examples: GPS Navigation, Child Seat, Insurance, etc.
type Extra struct {
	name       string  // Name of the extra service
	dailyPrice float64 // Cost per day for this extra
}

// NewExtra creates a new Extra instance.
func NewExtra(name string, dailyPrice float64) Extra {
	return Extra{
		name:       name,
		dailyPrice: dailyPrice,
	}
}

func (extra Extra) GetName() string        { return extra.name }
func (extra Extra) GetDailyPrice() float64 { return extra.dailyPrice }

// ============================================================================
// SECTION 5: RESERVATION ENTITY
// ============================================================================

// Reservation represents a vehicle booking made by a customer.
// It tracks the entire lifecycle from creation to completion.
type Reservation struct {
	id             string            // Unique identifier (e.g., "RES-1")
	customer       *Customer         // Customer who made the reservation
	vehicle        *Vehicle          // Reserved vehicle
	pickupDate     time.Time         // When the rental starts
	returnDate     time.Time         // When the rental ends
	pickupLocation string            // Where to pick up the vehicle
	returnLocation string            // Where to return the vehicle
	status         ReservationStatus // Current status of the reservation
	dailyRate      float64           // Base daily rate at time of booking
	totalAmount    float64           // Total cost including extras
	extras         []Extra           // Additional services added
	createdAt      time.Time         // When the reservation was created
	mutex          sync.Mutex        // Protects concurrent modifications
}

// reservationIDGenerator generates unique IDs for reservations.
// Using a struct with mutex for thread-safety.
type reservationIDGenerator struct {
	counter int
	mutex   sync.Mutex
}

var idGenerator = &reservationIDGenerator{counter: 0}

// NextID generates the next unique reservation ID.
func (gen *reservationIDGenerator) NextID() string {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()
	gen.counter++
	return fmt.Sprintf("RES-%d", gen.counter)
}

// NewReservation creates a new reservation for a customer and vehicle.
// It calculates the initial total based on the number of rental days.
func NewReservation(customer *Customer, vehicle *Vehicle, pickupDate, returnDate time.Time, location string) *Reservation {
	rentalDays := calculateRentalDays(pickupDate, returnDate)
	dailyRate := vehicle.GetDailyRate()

	return &Reservation{
		id:             idGenerator.NextID(),
		customer:       customer,
		vehicle:        vehicle,
		pickupDate:     pickupDate,
		returnDate:     returnDate,
		pickupLocation: location,
		returnLocation: location,
		status:         ReservationStatusPending,
		dailyRate:      dailyRate,
		totalAmount:    dailyRate * float64(rentalDays),
		extras:         make([]Extra, 0),
		createdAt:      time.Now(),
	}
}

// calculateRentalDays computes the number of days between two dates.
// Minimum rental period is 1 day.
func calculateRentalDays(startDate, endDate time.Time) int {
	hours := endDate.Sub(startDate).Hours()
	days := int(hours/24) + 1 // +1 because partial days count as full days
	if days < 1 {
		days = 1
	}
	return days
}

// Getter methods for Reservation
func (reservation *Reservation) GetID() string     { return reservation.id }
func (reservation *Reservation) GetTotal() float64 { return reservation.totalAmount }
func (reservation *Reservation) GetStatus() ReservationStatus {
	reservation.mutex.Lock()
	defer reservation.mutex.Unlock()
	return reservation.status
}

// GetRentalDays returns the total number of rental days.
func (reservation *Reservation) GetRentalDays() int {
	return calculateRentalDays(reservation.pickupDate, reservation.returnDate)
}

// AddExtra adds an optional service/item to the reservation.
// The extra's cost is added to the total for each rental day.
func (reservation *Reservation) AddExtra(name string, dailyPrice float64) {
	reservation.mutex.Lock()
	defer reservation.mutex.Unlock()

	extra := NewExtra(name, dailyPrice)
	reservation.extras = append(reservation.extras, extra)

	// Calculate extra cost: dailyPrice × number of rental days
	rentalDays := calculateRentalDays(reservation.pickupDate, reservation.returnDate)
	reservation.totalAmount += dailyPrice * float64(rentalDays)
}

// Confirm moves the reservation from Pending to Confirmed status.
// The vehicle is marked as Reserved to prevent double-booking.
func (reservation *Reservation) Confirm() error {
	reservation.mutex.Lock()
	defer reservation.mutex.Unlock()

	if reservation.status != ReservationStatusPending {
		return fmt.Errorf("cannot confirm: reservation is not in pending status (current: %s)", reservation.status)
	}

	reservation.status = ReservationStatusConfirmed
	reservation.vehicle.SetStatus(VehicleStatusReserved)
	return nil
}

// PickUp processes the vehicle pickup by the customer.
// Only confirmed reservations can be picked up.
func (reservation *Reservation) PickUp() error {
	reservation.mutex.Lock()
	defer reservation.mutex.Unlock()

	if reservation.status != ReservationStatusConfirmed {
		return fmt.Errorf("cannot pick up: reservation is not confirmed (current: %s)", reservation.status)
	}

	reservation.status = ReservationStatusPickedUp
	reservation.vehicle.SetStatus(VehicleStatusRented)
	return nil
}

// Return processes the vehicle return by the customer.
// Only picked-up reservations can be returned.
func (reservation *Reservation) Return() error {
	reservation.mutex.Lock()
	defer reservation.mutex.Unlock()

	if reservation.status != ReservationStatusPickedUp {
		return fmt.Errorf("cannot return: vehicle was not picked up (current: %s)", reservation.status)
	}

	reservation.status = ReservationStatusReturned
	reservation.vehicle.SetStatus(VehicleStatusAvailable)

	// Add to customer's rental history
	reservation.customer.AddRentalToHistory(reservation)
	return nil
}

// Cancel cancels the reservation if the vehicle hasn't been picked up.
func (reservation *Reservation) Cancel() error {
	reservation.mutex.Lock()
	defer reservation.mutex.Unlock()

	if reservation.status == ReservationStatusPickedUp {
		return fmt.Errorf("cannot cancel: vehicle has already been picked up")
	}

	if reservation.status == ReservationStatusReturned {
		return fmt.Errorf("cannot cancel: rental has already been completed")
	}

	if reservation.status == ReservationStatusCancelled {
		return fmt.Errorf("reservation is already cancelled")
	}

	reservation.status = ReservationStatusCancelled
	reservation.vehicle.SetStatus(VehicleStatusAvailable)
	return nil
}

// PrintReceipt displays a formatted receipt for the reservation.
func (reservation *Reservation) PrintReceipt() {
	rentalDays := reservation.GetRentalDays()
	baseCharge := reservation.dailyRate * float64(rentalDays)

	fmt.Printf(`
╔════════════════════════════════════════════════╗
║           🚗 RENTAL RECEIPT                    ║
╠════════════════════════════════════════════════╣
  Reservation: %s
  Status: %s
  
  Customer: %s
  License: %s
  
  Vehicle: %d %s %s
  Type: %s
  Plate: %s
  
  Pickup:  %s at %s
  Return:  %s at %s
  Days: %d
  
  ────────────────────────────────
  CHARGES:
  Daily Rate: $%.2f x %d days = $%.2f
`,
		reservation.id,
		reservation.status,
		reservation.customer.GetName(),
		reservation.customer.GetDriverLicense(),
		reservation.vehicle.GetYear(),
		reservation.vehicle.GetMake(),
		reservation.vehicle.GetModel(),
		reservation.vehicle.GetType(),
		reservation.vehicle.GetLicensePlate(),
		reservation.pickupDate.Format("Jan 02"),
		reservation.pickupLocation,
		reservation.returnDate.Format("Jan 02"),
		reservation.returnLocation,
		rentalDays,
		reservation.dailyRate,
		rentalDays,
		baseCharge)

	// Print each extra service
	for _, extra := range reservation.extras {
		extraTotal := extra.GetDailyPrice() * float64(rentalDays)
		fmt.Printf("  %s: $%.2f x %d days = $%.2f\n",
			extra.GetName(), extra.GetDailyPrice(), rentalDays, extraTotal)
	}

	fmt.Printf(`  ────────────────────────────────
  TOTAL: $%.2f
╚════════════════════════════════════════════════╝
`, reservation.totalAmount)
}

// ============================================================================
// SECTION 6: RENTAL SERVICE (Main Business Logic)
// ============================================================================

// RentalService is the central service that manages the car rental operations.
// It coordinates vehicles, customers, and reservations.
type RentalService struct {
	vehicles     map[string]*Vehicle     // All vehicles in the fleet (key: vehicle ID)
	customers    map[string]*Customer    // All registered customers (key: customer ID)
	reservations map[string]*Reservation // All reservations (key: reservation ID)
	locations    []string                // Available pickup/return locations
	mutex        sync.RWMutex            // Read-write lock for thread-safe operations
}

// NewRentalService creates and initializes a new RentalService.
func NewRentalService() *RentalService {
	return &RentalService{
		vehicles:     make(map[string]*Vehicle),
		customers:    make(map[string]*Customer),
		reservations: make(map[string]*Reservation),
		locations:    []string{"Airport", "Downtown", "Mall"},
	}
}

// AddVehicle adds a vehicle to the fleet.
func (service *RentalService) AddVehicle(vehicle *Vehicle) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.vehicles[vehicle.GetID()] = vehicle
}

// RegisterCustomer adds a new customer to the system.
func (service *RentalService) RegisterCustomer(customer *Customer) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.customers[customer.GetID()] = customer
}

// GetAvailableVehiclesByType returns all available vehicles of a specific type at a location.
func (service *RentalService) GetAvailableVehiclesByType(vehicleType VehicleType, location string) []*Vehicle {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	availableVehicles := make([]*Vehicle, 0)
	for _, vehicle := range service.vehicles {
		if vehicle.IsAvailable() &&
			vehicle.GetType() == vehicleType &&
			vehicle.GetLocation() == location {
			availableVehicles = append(availableVehicles, vehicle)
		}
	}
	return availableVehicles
}

// GetAllAvailableVehicles returns all available vehicles at a specific location.
func (service *RentalService) GetAllAvailableVehicles(location string) []*Vehicle {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	availableVehicles := make([]*Vehicle, 0)
	for _, vehicle := range service.vehicles {
		if vehicle.IsAvailable() && vehicle.GetLocation() == location {
			availableVehicles = append(availableVehicles, vehicle)
		}
	}
	return availableVehicles
}

// CreateReservation creates a new reservation for a customer and vehicle.
// Returns an error if the customer or vehicle doesn't exist, or if the vehicle isn't available.
func (service *RentalService) CreateReservation(customerID, vehicleID string, pickupDate, returnDate time.Time) (*Reservation, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	// Validate customer exists
	customer, customerExists := service.customers[customerID]
	if !customerExists {
		return nil, fmt.Errorf("customer with ID '%s' not found", customerID)
	}

	// Validate vehicle exists
	vehicle, vehicleExists := service.vehicles[vehicleID]
	if !vehicleExists {
		return nil, fmt.Errorf("vehicle with ID '%s' not found", vehicleID)
	}

	// Validate vehicle availability
	if !vehicle.IsAvailable() {
		return nil, fmt.Errorf("vehicle '%s' is not available (status: %s)", vehicleID, vehicle.GetStatus())
	}

	// Validate dates
	if returnDate.Before(pickupDate) {
		return nil, fmt.Errorf("return date cannot be before pickup date")
	}

	// Create and store the reservation
	reservation := NewReservation(customer, vehicle, pickupDate, returnDate, vehicle.GetLocation())
	service.reservations[reservation.GetID()] = reservation

	return reservation, nil
}

// ConfirmReservation confirms a pending reservation.
func (service *RentalService) ConfirmReservation(reservationID string) error {
	service.mutex.RLock()
	reservation, exists := service.reservations[reservationID]
	service.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("reservation with ID '%s' not found", reservationID)
	}

	return reservation.Confirm()
}

// PickUpVehicle processes the vehicle pickup for a reservation.
func (service *RentalService) PickUpVehicle(reservationID string) error {
	service.mutex.RLock()
	reservation, exists := service.reservations[reservationID]
	service.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("reservation with ID '%s' not found", reservationID)
	}

	return reservation.PickUp()
}

// ReturnVehicle processes the vehicle return for a reservation.
func (service *RentalService) ReturnVehicle(reservationID string) error {
	service.mutex.RLock()
	reservation, exists := service.reservations[reservationID]
	service.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("reservation with ID '%s' not found", reservationID)
	}

	return reservation.Return()
}

// CancelReservation cancels an existing reservation.
func (service *RentalService) CancelReservation(reservationID string) error {
	service.mutex.RLock()
	reservation, exists := service.reservations[reservationID]
	service.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("reservation with ID '%s' not found", reservationID)
	}

	return reservation.Cancel()
}

// ShowFleetStatus displays the current status of all vehicles in the fleet.
func (service *RentalService) ShowFleetStatus() {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	fmt.Println("\n🚗 Fleet Status:")
	fmt.Println("─────────────────────────────────────────")

	for _, vehicle := range service.vehicles {
		// Use green circle for available, red for unavailable
		statusIcon := "🟢"
		if vehicle.GetStatus() != VehicleStatusAvailable {
			statusIcon = "🔴"
		}
		fmt.Printf("  %s [%s] %s\n", statusIcon, vehicle.GetLocation(), vehicle)
	}
}

// GetLocations returns all available pickup/return locations.
func (service *RentalService) GetLocations() []string {
	return service.locations
}

// ============================================================================
// SECTION 7: MAIN - DEMONSTRATION
// ============================================================================

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("         🚗 CAR RENTAL SYSTEM")
	fmt.Println("═══════════════════════════════════════════")

	// Initialize the rental service
	rentalService := NewRentalService()

	// =========================================
	// STEP 1: Add vehicles to the fleet
	// =========================================
	fmt.Println("\n📦 Adding vehicles to fleet...")

	rentalService.AddVehicle(NewVehicle("V001", "ABC-123", "Toyota", "Camry", 2023, VehicleTypeCar, "Airport"))
	rentalService.AddVehicle(NewVehicle("V002", "XYZ-789", "Honda", "CR-V", 2023, VehicleTypeSUV, "Airport"))
	rentalService.AddVehicle(NewVehicle("V003", "DEF-456", "BMW", "5 Series", 2024, VehicleTypeLuxury, "Downtown"))
	rentalService.AddVehicle(NewVehicle("V004", "GHI-321", "Ford", "Explorer", 2022, VehicleTypeSUV, "Airport"))
	rentalService.AddVehicle(NewVehicle("V005", "JKL-654", "Toyota", "Sienna", 2023, VehicleTypeVan, "Mall"))

	fmt.Println("✅ 5 vehicles added to fleet")

	// =========================================
	// STEP 2: Register customers
	// =========================================
	fmt.Println("\n👥 Registering customers...")

	customer1 := NewCustomer("C001", "John Doe", "john@email.com", "555-0101", "DL-12345")
	customer2 := NewCustomer("C002", "Jane Smith", "jane@email.com", "555-0102", "DL-67890")

	rentalService.RegisterCustomer(customer1)
	rentalService.RegisterCustomer(customer2)

	fmt.Printf("✅ Registered: %s\n", customer1.GetName())
	fmt.Printf("✅ Registered: %s\n", customer2.GetName())

	// =========================================
	// STEP 3: Display fleet status
	// =========================================
	rentalService.ShowFleetStatus()

	// =========================================
	// STEP 4: Search for available vehicles
	// =========================================
	fmt.Println("\n🔍 Available at Airport:")
	fmt.Println("─────────────────────────────────────────")

	availableVehicles := rentalService.GetAllAvailableVehicles("Airport")
	for _, vehicle := range availableVehicles {
		fmt.Printf("  • %s\n", vehicle)
	}

	// =========================================
	// STEP 5: Create a reservation
	// =========================================
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("📝 Creating Reservation...")

	pickupDate := time.Now()
	returnDate := pickupDate.Add(3 * 24 * time.Hour) // 3-day rental

	reservation, err := rentalService.CreateReservation("C001", "V002", pickupDate, returnDate)
	if err != nil {
		fmt.Printf("❌ Error creating reservation: %v\n", err)
		return
	}
	fmt.Printf("✅ Reservation created: %s\n", reservation.GetID())

	// =========================================
	// STEP 6: Add extras to the reservation
	// =========================================
	fmt.Println("\n🎁 Adding extras...")

	reservation.AddExtra("GPS Navigation", 5.00)
	reservation.AddExtra("Child Seat", 8.00)
	reservation.AddExtra("Insurance", 15.00)

	fmt.Println("✅ Extras added: GPS Navigation, Child Seat, Insurance")

	// =========================================
	// STEP 7: Confirm and pickup the vehicle
	// =========================================
	err = rentalService.ConfirmReservation(reservation.GetID())
	if err != nil {
		fmt.Printf("❌ Error confirming reservation: %v\n", err)
		return
	}
	fmt.Println("✅ Reservation confirmed")

	err = rentalService.PickUpVehicle(reservation.GetID())
	if err != nil {
		fmt.Printf("❌ Error picking up vehicle: %v\n", err)
		return
	}
	fmt.Printf("✅ Vehicle picked up by %s\n", customer1.GetName())

	// Show updated fleet status
	rentalService.ShowFleetStatus()

	// =========================================
	// STEP 8: Return the vehicle
	// =========================================
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("🔑 Returning vehicle...")

	err = rentalService.ReturnVehicle(reservation.GetID())
	if err != nil {
		fmt.Printf("❌ Error returning vehicle: %v\n", err)
		return
	}
	fmt.Println("✅ Vehicle returned")

	// =========================================
	// STEP 9: Print the final receipt
	// =========================================
	reservation.PrintReceipt()

	// Show final fleet status
	rentalService.ShowFleetStatus()

	// =========================================
	// SUMMARY: Key Design Decisions
	// =========================================
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  KEY DESIGN DECISIONS:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. Vehicle types with different daily rates")
	fmt.Println("  2. Reservation lifecycle: Pending → Confirmed → PickedUp → Returned")
	fmt.Println("  3. Extras (GPS, Insurance, etc.) added dynamically")
	fmt.Println("  4. Location-based fleet management")
	fmt.Println("  5. Thread-safe operations using mutex locks")
	fmt.Println("  6. Clean separation of entities and service layer")
	fmt.Println("═══════════════════════════════════════════")
}
