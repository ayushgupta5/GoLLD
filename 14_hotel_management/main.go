package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// HOTEL MANAGEMENT SYSTEM - Low Level Design
// ============================================================================
//
// This system demonstrates:
// - Entity Modeling (Guest, Room, Booking, Hotel)
// - State Management (Room status, Booking lifecycle)
// - Business Logic (Check-in, Check-out, Billing)
// - Thread-safe operations using mutex locks
//
// ============================================================================

// ============================================================================
// SECTION 1: ROOM TYPE ENUM
// ============================================================================

// RoomType represents different categories of hotel rooms.
// Using iota for auto-incrementing values (Standard=0, Deluxe=1, etc.)
type RoomType int

const (
	RoomTypeStandard     RoomType = iota // 0 - Basic room with essential amenities
	RoomTypeDeluxe                       // 1 - Upgraded room with extra amenities
	RoomTypeSuite                        // 2 - Large room with living area
	RoomTypePresidential                 // 3 - Top-tier luxury room
)

// String returns a human-readable name for the room type.
func (roomType RoomType) String() string {
	names := [...]string{"Standard", "Deluxe", "Suite", "Presidential"}
	if int(roomType) < len(names) {
		return names[roomType]
	}
	return "Unknown"
}

// BasePrice returns the nightly rate for each room type.
func (roomType RoomType) BasePrice() float64 {
	prices := [...]float64{100.0, 150.0, 250.0, 500.0}
	if int(roomType) < len(prices) {
		return prices[roomType]
	}
	return 0.0
}

// ============================================================================
// SECTION 2: ROOM STATUS ENUM
// ============================================================================

// RoomStatus represents the current availability state of a room.
type RoomStatus int

const (
	RoomStatusAvailable   RoomStatus = iota // 0 - Room is ready for booking
	RoomStatusOccupied                      // 1 - Guest is currently staying
	RoomStatusMaintenance                   // 2 - Under repair/maintenance
	RoomStatusCleaning                      // 3 - Being cleaned after checkout
)

// String returns a human-readable name for the room status.
func (status RoomStatus) String() string {
	names := [...]string{"Available", "Occupied", "Maintenance", "Cleaning"}
	if int(status) < len(names) {
		return names[status]
	}
	return "Unknown"
}

// ============================================================================
// SECTION 3: BOOKING STATUS ENUM
// ============================================================================

// BookingStatus represents the lifecycle state of a booking.
type BookingStatus int

const (
	BookingStatusPending    BookingStatus = iota // 0 - Booking created, awaiting confirmation
	BookingStatusConfirmed                       // 1 - Booking confirmed, room reserved
	BookingStatusCheckedIn                       // 2 - Guest has checked in
	BookingStatusCheckedOut                      // 3 - Guest has checked out
	BookingStatusCancelled                       // 4 - Booking was cancelled
)

// String returns a human-readable name for the booking status.
func (status BookingStatus) String() string {
	names := [...]string{"Pending", "Confirmed", "Checked-In", "Checked-Out", "Cancelled"}
	if int(status) < len(names) {
		return names[status]
	}
	return "Unknown"
}

// ============================================================================
// SECTION 4: GUEST ENTITY
// ============================================================================

// Guest represents a person who books a room at the hotel.
type Guest struct {
	id           string // Unique identifier (e.g., "G001")
	name         string // Full name
	email        string // Contact email
	phone        string // Contact phone number
	identityCard string // Government ID number (for verification)
	address      string // Home address
}

// NewGuest creates and initializes a new Guest instance.
func NewGuest(id, name, email, phone string) *Guest {
	return &Guest{
		id:           id,
		name:         name,
		email:        email,
		phone:        phone,
		identityCard: "",
		address:      "",
	}
}

// Getter methods for Guest fields
func (guest *Guest) GetID() string    { return guest.id }
func (guest *Guest) GetName() string  { return guest.name }
func (guest *Guest) GetEmail() string { return guest.email }
func (guest *Guest) GetPhone() string { return guest.phone }

// SetIdentityCard sets the guest's ID card number (for verification at check-in).
func (guest *Guest) SetIdentityCard(idCard string) {
	guest.identityCard = idCard
}

// SetAddress sets the guest's home address.
func (guest *Guest) SetAddress(address string) {
	guest.address = address
}

// ============================================================================
// SECTION 5: ROOM ENTITY
// ============================================================================

// Room represents a hotel room that can be booked by guests.
type Room struct {
	number        string     // Room number (e.g., "101", "201")
	floor         int        // Floor number
	roomType      RoomType   // Type of room (Standard, Deluxe, etc.)
	status        RoomStatus // Current availability status
	pricePerNight float64    // Cost per night in dollars
	amenities     []string   // List of amenities (WiFi, TV, etc.)
	mutex         sync.Mutex // Protects concurrent access to room state
}

// NewRoom creates a new Room with appropriate amenities based on room type.
// Amenities are automatically assigned based on the room type.
func NewRoom(roomNumber string, floor int, roomType RoomType) *Room {
	// Base amenities for all rooms
	amenities := []string{"WiFi", "TV", "AC"}

	// Add amenities based on room type (cumulative)
	if roomType >= RoomTypeDeluxe {
		amenities = append(amenities, "Mini Bar", "Room Service")
	}
	if roomType >= RoomTypeSuite {
		amenities = append(amenities, "Living Room", "Jacuzzi")
	}
	if roomType == RoomTypePresidential {
		amenities = append(amenities, "Butler Service", "Private Pool")
	}

	return &Room{
		number:        roomNumber,
		floor:         floor,
		roomType:      roomType,
		status:        RoomStatusAvailable,
		pricePerNight: roomType.BasePrice(),
		amenities:     amenities,
	}
}

// Getter methods for Room fields
func (room *Room) GetNumber() string      { return room.number }
func (room *Room) GetFloor() int          { return room.floor }
func (room *Room) GetType() RoomType      { return room.roomType }
func (room *Room) GetPrice() float64      { return room.pricePerNight }
func (room *Room) GetAmenities() []string { return room.amenities }

// GetStatus returns the current status of the room (thread-safe).
func (room *Room) GetStatus() RoomStatus {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	return room.status
}

// SetStatus updates the room status (thread-safe).
func (room *Room) SetStatus(newStatus RoomStatus) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	room.status = newStatus
}

// IsAvailable checks if the room can be booked.
func (room *Room) IsAvailable() bool {
	return room.GetStatus() == RoomStatusAvailable
}

// String returns a formatted description of the room.
func (room *Room) String() string {
	return fmt.Sprintf("Room %s (%s) - $%.2f/night - %s",
		room.number, room.roomType, room.pricePerNight, room.status)
}

// ============================================================================
// SECTION 6: SERVICE ENTITY
// ============================================================================

// Service represents an additional service consumed by a guest during their stay.
// Examples: Room Service, Spa Treatment, Mini Bar, Laundry, etc.
type Service struct {
	name      string    // Name of the service
	price     float64   // Cost of the service
	timestamp time.Time // When the service was ordered
}

// NewService creates a new Service instance.
func NewService(name string, price float64) Service {
	return Service{
		name:      name,
		price:     price,
		timestamp: time.Now(),
	}
}

func (service Service) GetName() string   { return service.name }
func (service Service) GetPrice() float64 { return service.price }

// ============================================================================
// SECTION 7: BOOKING ENTITY
// ============================================================================

// bookingIDGenerator generates unique IDs for bookings (thread-safe).
type bookingIDGenerator struct {
	counter int
	mutex   sync.Mutex
}

var bookingIDGen = &bookingIDGenerator{counter: 0}

// NextID generates the next unique booking ID.
func (gen *bookingIDGenerator) NextID() string {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()
	gen.counter++
	return fmt.Sprintf("BK-%d", gen.counter)
}

// Booking represents a room reservation made by a guest.
// It tracks the entire stay lifecycle from creation to checkout.
type Booking struct {
	id           string        // Unique identifier (e.g., "BK-1")
	guest        *Guest        // Guest who made the booking
	room         *Room         // Room that was booked
	checkInDate  time.Time     // Scheduled check-in date
	checkOutDate time.Time     // Scheduled check-out date
	status       BookingStatus // Current status of the booking
	totalAmount  float64       // Total bill amount (room + services)
	services     []Service     // Additional services consumed
	createdAt    time.Time     // When the booking was created
	mutex        sync.Mutex    // Protects concurrent modifications
}

// NewBooking creates a new booking for a guest and room.
// The total amount is initially calculated based on room rate and number of nights.
func NewBooking(guest *Guest, room *Room, checkInDate, checkOutDate time.Time) *Booking {
	numberOfNights := calculateNights(checkInDate, checkOutDate)
	roomTotal := room.GetPrice() * float64(numberOfNights)

	return &Booking{
		id:           bookingIDGen.NextID(),
		guest:        guest,
		room:         room,
		checkInDate:  checkInDate,
		checkOutDate: checkOutDate,
		status:       BookingStatusPending,
		totalAmount:  roomTotal,
		services:     make([]Service, 0),
		createdAt:    time.Now(),
	}
}

// calculateNights computes the number of nights between two dates.
// Minimum stay is 1 night.
func calculateNights(checkIn, checkOut time.Time) int {
	hours := checkOut.Sub(checkIn).Hours()
	nights := int(hours / 24)
	if nights < 1 {
		nights = 1
	}
	return nights
}

// Getter methods for Booking
func (booking *Booking) GetID() string              { return booking.id }
func (booking *Booking) GetGuest() *Guest           { return booking.guest }
func (booking *Booking) GetRoom() *Room             { return booking.room }
func (booking *Booking) GetTotal() float64          { return booking.totalAmount }
func (booking *Booking) GetCheckInDate() time.Time  { return booking.checkInDate }
func (booking *Booking) GetCheckOutDate() time.Time { return booking.checkOutDate }

// GetStatus returns the current booking status (thread-safe).
func (booking *Booking) GetStatus() BookingStatus {
	booking.mutex.Lock()
	defer booking.mutex.Unlock()
	return booking.status
}

// GetNights returns the number of nights for this booking.
func (booking *Booking) GetNights() int {
	return calculateNights(booking.checkInDate, booking.checkOutDate)
}

// Confirm changes the booking status from Pending to Confirmed.
func (booking *Booking) Confirm() error {
	booking.mutex.Lock()
	defer booking.mutex.Unlock()

	if booking.status != BookingStatusPending {
		return fmt.Errorf("cannot confirm: booking is not in pending status (current: %s)", booking.status)
	}

	booking.status = BookingStatusConfirmed
	return nil
}

// CheckIn processes guest check-in.
// Only confirmed bookings can be checked in.
func (booking *Booking) CheckIn() error {
	booking.mutex.Lock()
	defer booking.mutex.Unlock()

	if booking.status != BookingStatusConfirmed {
		return fmt.Errorf("cannot check in: booking must be confirmed first (current: %s)", booking.status)
	}

	booking.status = BookingStatusCheckedIn
	booking.room.SetStatus(RoomStatusOccupied)
	return nil
}

// CheckOut processes guest checkout.
// Only checked-in guests can check out.
func (booking *Booking) CheckOut() error {
	booking.mutex.Lock()
	defer booking.mutex.Unlock()

	if booking.status != BookingStatusCheckedIn {
		return fmt.Errorf("cannot check out: guest has not checked in (current: %s)", booking.status)
	}

	booking.status = BookingStatusCheckedOut
	booking.room.SetStatus(RoomStatusCleaning) // Room needs cleaning after checkout
	return nil
}

// Cancel cancels the booking if the guest hasn't checked in yet.
func (booking *Booking) Cancel() error {
	booking.mutex.Lock()
	defer booking.mutex.Unlock()

	if booking.status == BookingStatusCheckedIn {
		return fmt.Errorf("cannot cancel: guest has already checked in")
	}

	if booking.status == BookingStatusCheckedOut {
		return fmt.Errorf("cannot cancel: booking has already been completed")
	}

	if booking.status == BookingStatusCancelled {
		return fmt.Errorf("booking is already cancelled")
	}

	booking.status = BookingStatusCancelled
	return nil
}

// AddService adds an additional service to the booking and updates the total.
func (booking *Booking) AddService(serviceName string, price float64) {
	booking.mutex.Lock()
	defer booking.mutex.Unlock()

	service := NewService(serviceName, price)
	booking.services = append(booking.services, service)
	booking.totalAmount += price
}

// GenerateBill creates a formatted invoice for the booking.
func (booking *Booking) GenerateBill() string {
	numberOfNights := booking.GetNights()
	roomCharge := booking.room.GetPrice() * float64(numberOfNights)

	bill := fmt.Sprintf(`
╔════════════════════════════════════════════════╗
║              🏨 HOTEL INVOICE                  ║
╠════════════════════════════════════════════════╣
  Booking ID: %s
  Guest: %s
  Room: %s (%s)
  
  Check-in:  %s
  Check-out: %s
  Nights: %d
  
  ─────────────────────────────────────
  CHARGES:
  Room (%d nights × $%.2f): $%.2f
`,
		booking.id,
		booking.guest.GetName(),
		booking.room.GetNumber(),
		booking.room.GetType(),
		booking.checkInDate.Format("Jan 02, 2006"),
		booking.checkOutDate.Format("Jan 02, 2006"),
		numberOfNights,
		numberOfNights,
		booking.room.GetPrice(),
		roomCharge,
	)

	// Add each service to the bill
	for _, service := range booking.services {
		bill += fmt.Sprintf("  %s: $%.2f\n", service.GetName(), service.GetPrice())
	}

	bill += fmt.Sprintf(`  ─────────────────────────────────────
  TOTAL: $%.2f
╚════════════════════════════════════════════════╝
`, booking.totalAmount)

	return bill
}

// ============================================================================
// SECTION 8: HOTEL (Main Service)
// ============================================================================

// Hotel is the central service that manages rooms, guests, and bookings.
type Hotel struct {
	name     string              // Hotel name
	address  string              // Hotel address
	rooms    map[string]*Room    // All rooms (key: room number)
	bookings map[string]*Booking // All bookings (key: booking ID)
	guests   map[string]*Guest   // All registered guests (key: guest ID)
	mutex    sync.RWMutex        // Read-write lock for thread-safe operations
}

// NewHotel creates and initializes a new Hotel instance.
func NewHotel(name, address string) *Hotel {
	return &Hotel{
		name:     name,
		address:  address,
		rooms:    make(map[string]*Room),
		bookings: make(map[string]*Booking),
		guests:   make(map[string]*Guest),
	}
}

// Getter methods for Hotel
func (hotel *Hotel) GetName() string    { return hotel.name }
func (hotel *Hotel) GetAddress() string { return hotel.address }

// AddRoom adds a room to the hotel's inventory.
func (hotel *Hotel) AddRoom(room *Room) {
	hotel.mutex.Lock()
	defer hotel.mutex.Unlock()
	hotel.rooms[room.GetNumber()] = room
}

// RegisterGuest adds a guest to the hotel's system.
func (hotel *Hotel) RegisterGuest(guest *Guest) {
	hotel.mutex.Lock()
	defer hotel.mutex.Unlock()
	hotel.guests[guest.GetID()] = guest
}

// GetAvailableRoomsByType returns all available rooms of a specific type.
func (hotel *Hotel) GetAvailableRoomsByType(roomType RoomType) []*Room {
	hotel.mutex.RLock()
	defer hotel.mutex.RUnlock()

	availableRooms := make([]*Room, 0)
	for _, room := range hotel.rooms {
		if room.IsAvailable() && room.GetType() == roomType {
			availableRooms = append(availableRooms, room)
		}
	}
	return availableRooms
}

// GetAllAvailableRooms returns all available rooms regardless of type.
func (hotel *Hotel) GetAllAvailableRooms() []*Room {
	hotel.mutex.RLock()
	defer hotel.mutex.RUnlock()

	availableRooms := make([]*Room, 0)
	for _, room := range hotel.rooms {
		if room.IsAvailable() {
			availableRooms = append(availableRooms, room)
		}
	}
	return availableRooms
}

// CreateBooking creates a new booking for a guest.
// Returns an error if the guest or room doesn't exist, or if the room isn't available.
func (hotel *Hotel) CreateBooking(guestID, roomNumber string, checkIn, checkOut time.Time) (*Booking, error) {
	hotel.mutex.Lock()
	defer hotel.mutex.Unlock()

	// Validate guest exists
	guest, guestExists := hotel.guests[guestID]
	if !guestExists {
		return nil, fmt.Errorf("guest with ID '%s' not found", guestID)
	}

	// Validate room exists
	room, roomExists := hotel.rooms[roomNumber]
	if !roomExists {
		return nil, fmt.Errorf("room '%s' not found", roomNumber)
	}

	// Validate room is available
	if !room.IsAvailable() {
		return nil, fmt.Errorf("room '%s' is not available (status: %s)", roomNumber, room.GetStatus())
	}

	// Validate dates
	if checkOut.Before(checkIn) {
		return nil, fmt.Errorf("check-out date cannot be before check-in date")
	}

	// Create and store the booking
	booking := NewBooking(guest, room, checkIn, checkOut)
	hotel.bookings[booking.GetID()] = booking

	return booking, nil
}

// ConfirmBooking confirms a pending booking.
func (hotel *Hotel) ConfirmBooking(bookingID string) error {
	hotel.mutex.RLock()
	booking, exists := hotel.bookings[bookingID]
	hotel.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("booking with ID '%s' not found", bookingID)
	}

	return booking.Confirm()
}

// CheckIn processes guest check-in for a booking.
func (hotel *Hotel) CheckIn(bookingID string) error {
	hotel.mutex.RLock()
	booking, exists := hotel.bookings[bookingID]
	hotel.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("booking with ID '%s' not found", bookingID)
	}

	return booking.CheckIn()
}

// CheckOut processes guest checkout for a booking.
// Returns the booking so the final bill can be generated.
func (hotel *Hotel) CheckOut(bookingID string) (*Booking, error) {
	hotel.mutex.RLock()
	booking, exists := hotel.bookings[bookingID]
	hotel.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("booking with ID '%s' not found", bookingID)
	}

	err := booking.CheckOut()
	if err != nil {
		return nil, err
	}

	return booking, nil
}

// CancelBooking cancels an existing booking.
func (hotel *Hotel) CancelBooking(bookingID string) error {
	hotel.mutex.RLock()
	booking, exists := hotel.bookings[bookingID]
	hotel.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("booking with ID '%s' not found", bookingID)
	}

	return booking.Cancel()
}

// DisplayRoomStatus shows the current status of all rooms in the hotel.
func (hotel *Hotel) DisplayRoomStatus() {
	hotel.mutex.RLock()
	defer hotel.mutex.RUnlock()

	fmt.Printf("\n🏨 %s - Room Status\n", hotel.name)
	fmt.Println("─────────────────────────────────────────")

	for _, room := range hotel.rooms {
		// Use green circle for available, red for unavailable
		statusIcon := "🟢"
		if room.GetStatus() != RoomStatusAvailable {
			statusIcon = "🔴"
		}
		fmt.Printf("  %s %s\n", statusIcon, room)
	}
}

// ============================================================================
// SECTION 9: MAIN - DEMONSTRATION
// ============================================================================

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("       🏨 HOTEL MANAGEMENT SYSTEM")
	fmt.Println("═══════════════════════════════════════════")

	// =========================================
	// STEP 1: Create the hotel
	// =========================================
	hotel := NewHotel("Grand Plaza Hotel", "123 Main Street")

	// =========================================
	// STEP 2: Add rooms to the hotel
	// =========================================
	fmt.Println("\n📦 Setting up hotel rooms...")

	hotel.AddRoom(NewRoom("101", 1, RoomTypeStandard))
	hotel.AddRoom(NewRoom("102", 1, RoomTypeStandard))
	hotel.AddRoom(NewRoom("201", 2, RoomTypeDeluxe))
	hotel.AddRoom(NewRoom("202", 2, RoomTypeDeluxe))
	hotel.AddRoom(NewRoom("301", 3, RoomTypeSuite))
	hotel.AddRoom(NewRoom("401", 4, RoomTypePresidential))

	fmt.Println("✅ 6 rooms added to hotel")

	// =========================================
	// STEP 3: Register guests
	// =========================================
	fmt.Println("\n👥 Registering guests...")

	guest1 := NewGuest("G001", "John Smith", "john@email.com", "555-0101")
	guest2 := NewGuest("G002", "Jane Doe", "jane@email.com", "555-0102")

	hotel.RegisterGuest(guest1)
	hotel.RegisterGuest(guest2)

	fmt.Printf("✅ Registered: %s\n", guest1.GetName())
	fmt.Printf("✅ Registered: %s\n", guest2.GetName())

	// =========================================
	// STEP 4: Display room status
	// =========================================
	hotel.DisplayRoomStatus()

	// =========================================
	// STEP 5: Search for available rooms
	// =========================================
	fmt.Println("\n📋 Available Deluxe Rooms:")
	fmt.Println("─────────────────────────────────────────")

	deluxeRooms := hotel.GetAvailableRoomsByType(RoomTypeDeluxe)
	for _, room := range deluxeRooms {
		fmt.Printf("  • %s\n", room)
	}

	// =========================================
	// STEP 6: Create bookings
	// =========================================
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("📝 Creating Bookings...")

	checkInDate := time.Now()
	checkOutDate := checkInDate.Add(3 * 24 * time.Hour) // 3-night stay

	booking1, err := hotel.CreateBooking("G001", "201", checkInDate, checkOutDate)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Booking created: %s for %s\n", booking1.GetID(), guest1.GetName())
	}

	booking2, err := hotel.CreateBooking("G002", "301", checkInDate, checkOutDate)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Booking created: %s for %s\n", booking2.GetID(), guest2.GetName())
	}

	// =========================================
	// STEP 7: Confirm booking and check-in
	// =========================================
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("🔑 Check-in Process...")

	err = hotel.ConfirmBooking(booking1.GetID())
	if err != nil {
		fmt.Printf("❌ Error confirming: %v\n", err)
	} else {
		fmt.Printf("✅ Booking %s confirmed\n", booking1.GetID())
	}

	err = hotel.CheckIn(booking1.GetID())
	if err != nil {
		fmt.Printf("❌ Error checking in: %v\n", err)
	} else {
		fmt.Printf("✅ %s checked into Room %s\n", guest1.GetName(), booking1.GetRoom().GetNumber())
	}

	// =========================================
	// STEP 8: Add services during stay
	// =========================================
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("🍽️  Adding Services...")

	booking1.AddService("Room Service - Dinner", 45.00)
	booking1.AddService("Mini Bar", 30.00)
	booking1.AddService("Spa Treatment", 120.00)

	fmt.Println("✅ Services added:")
	fmt.Println("   • Room Service - Dinner: $45.00")
	fmt.Println("   • Mini Bar: $30.00")
	fmt.Println("   • Spa Treatment: $120.00")

	// =========================================
	// STEP 9: Display room status after check-in
	// =========================================
	hotel.DisplayRoomStatus()

	// =========================================
	// STEP 10: Process checkout
	// =========================================
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("🚪 Check-out Process...")

	completedBooking, err := hotel.CheckOut(booking1.GetID())
	if err != nil {
		fmt.Printf("❌ Error checking out: %v\n", err)
	} else {
		fmt.Printf("✅ %s checked out from Room %s\n",
			completedBooking.GetGuest().GetName(),
			completedBooking.GetRoom().GetNumber())
	}

	// =========================================
	// STEP 11: Generate and print the bill
	// =========================================
	fmt.Println(booking1.GenerateBill())

	// =========================================
	// SUMMARY: Key Design Decisions
	// =========================================
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  KEY DESIGN DECISIONS:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. Room has lifecycle: Available → Occupied → Cleaning")
	fmt.Println("  2. Booking lifecycle: Pending → Confirmed → CheckedIn → CheckedOut")
	fmt.Println("  3. Services added dynamically during stay")
	fmt.Println("  4. Bill generated at checkout with itemized charges")
	fmt.Println("  5. Thread-safe operations using mutex locks")
	fmt.Println("  6. Clean separation of entities and service layer")
	fmt.Println("═══════════════════════════════════════════")
}
