package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// BOOKMYSHOW - Movie Ticket Booking System
// ============================================================
//
// This is a Low-Level Design (LLD) implementation of a movie ticket
// booking system similar to BookMyShow.
//
// KEY CONCEPTS DEMONSTRATED:
// 1. Entity Modeling - Movies, Seats, Shows, Theatres, Bookings
// 2. Concurrency Safety - Using mutexes for thread-safe operations
// 3. Enum Pattern in Go - Using iota for seat types and booking status
// 4. Service Layer - BookingService manages all business logic
//
// FLOW:
// User -> Selects Movie -> Selects Show -> Selects Seats -> Creates Booking -> Confirms/Cancels
//
// ============================================================

// ==================== ENUMS (Constants) ====================
// Go doesn't have built-in enums, so we use the iota pattern.
// iota starts at 0 and increments by 1 for each constant.

// SeatType represents the category of a seat (affects pricing)
type SeatType int

const (
	SeatTypeRegular SeatType = iota // 0 - Basic seats (cheapest)
	SeatTypePremium                 // 1 - Better seats (mid-price)
	SeatTypeVIP                     // 2 - Best seats (most expensive)
)

// String converts SeatType to a human-readable string
func (seatType SeatType) String() string {
	// Using an array literal indexed by the enum value
	seatTypeNames := [...]string{"Regular", "Premium", "VIP"}
	return seatTypeNames[seatType]
}

// GetPrice returns the price for this seat type
func (seatType SeatType) GetPrice() float64 {
	// Prices correspond to seat types: Regular=150, Premium=250, VIP=400
	seatTypePrices := [...]float64{150.0, 250.0, 400.0}
	return seatTypePrices[seatType]
}

// BookingStatus represents the current state of a booking
// Follows a state machine pattern: Pending -> Confirmed OR Pending -> Cancelled
type BookingStatus int

const (
	BookingStatusPending   BookingStatus = iota // 0 - Booking created, awaiting payment
	BookingStatusConfirmed                      // 1 - Payment received, booking active
	BookingStatusCancelled                      // 2 - Booking was cancelled
)

// String converts BookingStatus to a human-readable string
func (status BookingStatus) String() string {
	statusNames := [...]string{"Pending", "Confirmed", "Cancelled"}
	return statusNames[status]
}

// ==================== MOVIE ====================
// Movie represents a film that can be shown at theatres.
// This is an entity that holds all details about a movie.

type Movie struct {
	id          string    // Unique identifier for the movie
	title       string    // Name of the movie
	description string    // Brief plot summary
	duration    int       // Length in minutes
	genre       string    // Category (Action, Comedy, etc.)
	language    string    // Audio language
	releaseDate time.Time // When the movie was released
}

// NewMovie creates a new Movie instance with the provided details.
// This is the constructor pattern in Go - preferred over direct struct creation.
func NewMovie(id, title, genre, language string, durationMinutes int) *Movie {
	return &Movie{
		id:       id,
		title:    title,
		genre:    genre,
		language: language,
		duration: durationMinutes,
	}
}

// Getter methods - provide controlled access to private fields
// This is Go's way of encapsulation (lowercase fields are private)
func (movie *Movie) GetID() string    { return movie.id }
func (movie *Movie) GetTitle() string { return movie.title }

// String implements the Stringer interface for easy printing
func (movie *Movie) String() string {
	return fmt.Sprintf("%s (%s, %dmin)", movie.title, movie.language, movie.duration)
}

// ==================== SEAT ====================
// Seat represents an individual seat in a screen/auditorium.
// Each seat has a unique position (row + number) and a type that determines pricing.

type Seat struct {
	id       string   // Unique identifier (e.g., "A1", "B5")
	row      string   // Row letter (e.g., "A", "B", "C")
	number   int      // Seat number within the row
	seatType SeatType // Category of seat (Regular, Premium, VIP)
}

// NewSeat creates a new Seat with auto-generated ID.
// The ID is formed by combining row letter and seat number.
func NewSeat(row string, seatNumber int, seatType SeatType) *Seat {
	return &Seat{
		id:       fmt.Sprintf("%s%d", row, seatNumber), // e.g., "A1", "B5"
		row:      row,
		number:   seatNumber,
		seatType: seatType,
	}
}

// Getter methods for seat properties
func (seat *Seat) GetID() string     { return seat.id }
func (seat *Seat) GetType() SeatType { return seat.seatType }
func (seat *Seat) GetPrice() float64 { return seat.seatType.GetPrice() }

// String provides a readable representation of the seat
func (seat *Seat) String() string {
	return fmt.Sprintf("%s (%s)", seat.id, seat.seatType)
}

// ==================== SCREEN ====================
// Screen (Auditorium) represents a single screening room in a theatre.
// It contains multiple seats organized in rows.
// Seat types are automatically assigned based on row position:
// - First 2 rows: VIP (best view, premium pricing)
// - Next 2 rows: Premium (good view, mid-tier pricing)
// - Remaining rows: Regular (standard seating)

type Screen struct {
	id       string  // Unique identifier for the screen
	name     string  // Display name (e.g., "Screen 1", "IMAX")
	seats    []*Seat // All seats in this screen
	capacity int     // Total number of seats
}

// NewScreen creates a screen with seats auto-generated based on rows and seats per row.
// Seat types are assigned automatically based on row position (front rows get VIP).
func NewScreen(id, name string, rowLabels []string, seatsPerRow int) *Screen {
	screen := &Screen{
		id:    id,
		name:  name,
		seats: make([]*Seat, 0),
	}

	// Create seats for each row
	// Row index determines seat type: 0-1 = VIP, 2-3 = Premium, 4+ = Regular
	for rowIndex, rowLabel := range rowLabels {
		// Determine seat type based on row position
		seatType := SeatTypeRegular // Default to regular

		if rowIndex < 2 {
			seatType = SeatTypeVIP // Front rows (0, 1) are VIP
		} else if rowIndex < 4 {
			seatType = SeatTypePremium // Middle rows (2, 3) are Premium
		}
		// Remaining rows stay as Regular

		// Create all seats in this row
		for seatNumber := 1; seatNumber <= seatsPerRow; seatNumber++ {
			newSeat := NewSeat(rowLabel, seatNumber, seatType)
			screen.seats = append(screen.seats, newSeat)
		}
	}

	screen.capacity = len(screen.seats)
	return screen
}

// Getter methods for screen properties
func (screen *Screen) GetSeats() []*Seat { return screen.seats }
func (screen *Screen) GetCapacity() int  { return screen.capacity }
func (screen *Screen) GetName() string   { return screen.name }

// ==================== SHOW ====================
// Show represents a specific screening of a movie at a particular time.
// It tracks which seats have been booked and manages seat availability.
// The Show is the entity where seat booking state is managed.
//
// IMPORTANT: Show uses mutex for thread-safety because multiple users
// might try to book the same seats simultaneously.

type Show struct {
	id          string          // Unique identifier for this show
	movie       *Movie          // The movie being screened
	screen      *Screen         // The screen/auditorium for this show
	startTime   time.Time       // When the show starts
	endTime     time.Time       // When the show ends (auto-calculated)
	bookedSeats map[string]bool // Tracks booked seats (seatID -> true if booked)
	mutex       sync.RWMutex    // Protects bookedSeats from concurrent access
}

// NewShow creates a new show for a movie on a screen at a specific time.
// The end time is automatically calculated based on movie duration.
func NewShow(id string, movie *Movie, screen *Screen, startTime time.Time) *Show {
	return &Show{
		id:          id,
		movie:       movie,
		screen:      screen,
		startTime:   startTime,
		endTime:     startTime.Add(time.Duration(movie.duration) * time.Minute),
		bookedSeats: make(map[string]bool),
	}
}

// Getter methods for show properties
func (show *Show) GetID() string           { return show.id }
func (show *Show) GetMovie() *Movie        { return show.movie }
func (show *Show) GetScreen() *Screen      { return show.screen }
func (show *Show) GetStartTime() time.Time { return show.startTime }

// GetAvailableSeats returns a list of seats that have not been booked yet.
// Uses RLock (read lock) because we're only reading the bookedSeats map.
func (show *Show) GetAvailableSeats() []*Seat {
	show.mutex.RLock()
	defer show.mutex.RUnlock()

	availableSeats := make([]*Seat, 0)
	for _, seat := range show.screen.GetSeats() {
		isBooked := show.bookedSeats[seat.GetID()]
		if !isBooked {
			availableSeats = append(availableSeats, seat)
		}
	}
	return availableSeats
}

// IsSeatAvailable checks if a specific seat is available for booking.
func (show *Show) IsSeatAvailable(seatID string) bool {
	show.mutex.RLock()
	defer show.mutex.RUnlock()
	return !show.bookedSeats[seatID]
}

// BookSeats attempts to book the given seats atomically (all-or-nothing).
// This is thread-safe and will fail if any seat is already booked.
// Returns an error if any seat is unavailable, otherwise marks all seats as booked.
func (show *Show) BookSeats(seatIDs []string) error {
	show.mutex.Lock()
	defer show.mutex.Unlock()

	// STEP 1: Validate all seats are available (check before booking)
	// This ensures atomic booking - either all seats are booked or none
	for _, seatID := range seatIDs {
		if show.bookedSeats[seatID] {
			return fmt.Errorf("seat %s is already booked", seatID)
		}
	}

	// STEP 2: Book all seats (only reached if all seats are available)
	for _, seatID := range seatIDs {
		show.bookedSeats[seatID] = true
	}

	return nil
}

// ReleaseSeats releases previously booked seats (used when cancelling a booking).
// This allows the seats to be booked again by other users.
func (show *Show) ReleaseSeats(seatIDs []string) {
	show.mutex.Lock()
	defer show.mutex.Unlock()

	for _, seatID := range seatIDs {
		delete(show.bookedSeats, seatID)
	}
}

// String provides a summary of the show for display purposes.
func (show *Show) String() string {
	availableCount := len(show.GetAvailableSeats())
	totalCapacity := show.screen.GetCapacity()

	return fmt.Sprintf("%s | %s | %s | Available: %d/%d",
		show.movie.GetTitle(),
		show.screen.name,
		show.startTime.Format("15:04"),
		availableCount,
		totalCapacity)
}

// ==================== THEATRE ====================
// Theatre represents a cinema location with multiple screens.
// Each theatre can host multiple shows across its screens.

type Theatre struct {
	id      string    // Unique identifier for the theatre
	name    string    // Display name (e.g., "PVR Cinemas")
	city    string    // City where theatre is located
	address string    // Full address
	screens []*Screen // List of screens/auditoriums in this theatre
	shows   []*Show   // All shows scheduled at this theatre
	mutex   sync.RWMutex
}

// NewTheatre creates a new theatre in a given city.
func NewTheatre(id, name, city, address string) *Theatre {
	return &Theatre{
		id:      id,
		name:    name,
		city:    city,
		address: address,
		screens: make([]*Screen, 0),
		shows:   make([]*Show, 0),
	}
}

// AddScreen adds a new screen/auditorium to the theatre.
func (theatre *Theatre) AddScreen(screen *Screen) {
	theatre.screens = append(theatre.screens, screen)
}

// AddShow schedules a new show at this theatre.
func (theatre *Theatre) AddShow(show *Show) {
	theatre.mutex.Lock()
	defer theatre.mutex.Unlock()
	theatre.shows = append(theatre.shows, show)
}

// GetShows returns all shows scheduled at this theatre.
func (theatre *Theatre) GetShows() []*Show {
	theatre.mutex.RLock()
	defer theatre.mutex.RUnlock()
	return theatre.shows
}

// GetShowsForMovie returns all shows for a specific movie at this theatre.
func (theatre *Theatre) GetShowsForMovie(movieID string) []*Show {
	theatre.mutex.RLock()
	defer theatre.mutex.RUnlock()

	matchingShows := make([]*Show, 0)
	for _, show := range theatre.shows {
		if show.GetMovie().GetID() == movieID {
			matchingShows = append(matchingShows, show)
		}
	}
	return matchingShows
}

// String provides a readable representation of the theatre.
func (theatre *Theatre) String() string {
	return fmt.Sprintf("%s, %s", theatre.name, theatre.city)
}

// ==================== USER ====================
// User represents a customer who can book tickets.

type User struct {
	id    string // Unique identifier
	name  string // Full name
	email string // Email address
	phone string // Phone number
}

// NewUser creates a new user with the given details.
func NewUser(id, name, email, phone string) *User {
	return &User{
		id:    id,
		name:  name,
		email: email,
		phone: phone,
	}
}

// Getter methods for user properties
func (user *User) GetID() string    { return user.id }
func (user *User) GetName() string  { return user.name }
func (user *User) GetEmail() string { return user.email }

// ==================== BOOKING ====================
// Booking represents a user's reservation for seats at a show.
// It tracks the booking status through its lifecycle:
// Pending (created) -> Confirmed (paid) OR Cancelled

type Booking struct {
	id          string        // Unique booking identifier (e.g., "BKG-1")
	user        *User         // The user who made the booking
	show        *Show         // The show for this booking
	seats       []*Seat       // List of reserved seats
	totalAmount float64       // Total price (sum of all seat prices)
	status      BookingStatus // Current booking status
	createdAt   time.Time     // When the booking was created
	mutex       sync.Mutex    // Protects status changes
}

// bookingCounter is used to generate unique booking IDs.
// We use atomic operations for thread-safe incrementing.
var bookingCounter int64

// NewBooking creates a new booking for a user with the selected seats.
// The total amount is automatically calculated based on seat prices.
func NewBooking(user *User, show *Show, seats []*Seat) *Booking {
	// Atomically increment the counter to get a unique ID
	newID := atomic.AddInt64(&bookingCounter, 1)

	// Calculate total price from all selected seats
	var totalAmount float64
	for _, seat := range seats {
		totalAmount += seat.GetPrice()
	}

	return &Booking{
		id:          fmt.Sprintf("BKG-%d", newID),
		user:        user,
		show:        show,
		seats:       seats,
		totalAmount: totalAmount,
		status:      BookingStatusPending, // All bookings start as Pending
		createdAt:   time.Now(),
	}
}

// Getter methods for booking properties
func (booking *Booking) GetID() string            { return booking.id }
func (booking *Booking) GetStatus() BookingStatus { return booking.status }
func (booking *Booking) GetAmount() float64       { return booking.totalAmount }
func (booking *Booking) GetUser() *User           { return booking.user }
func (booking *Booking) GetShow() *Show           { return booking.show }
func (booking *Booking) GetSeats() []*Seat        { return booking.seats }

// Confirm marks the booking as confirmed (typically after payment).
func (booking *Booking) Confirm() {
	booking.mutex.Lock()
	defer booking.mutex.Unlock()
	booking.status = BookingStatusConfirmed
}

// Cancel marks the booking as cancelled and releases the reserved seats.
// Released seats become available for other users to book.
func (booking *Booking) Cancel() {
	booking.mutex.Lock()
	defer booking.mutex.Unlock()

	booking.status = BookingStatusCancelled

	// Collect seat IDs to release
	seatIDsToRelease := make([]string, len(booking.seats))
	for i, seat := range booking.seats {
		seatIDsToRelease[i] = seat.GetID()
	}

	// Release the seats back to the show (makes them available again)
	booking.show.ReleaseSeats(seatIDsToRelease)
}

// PrintTicket displays a formatted ticket for the booking.
func (booking *Booking) PrintTicket() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║          🎬 MOVIE TICKET 🎬            ║")
	fmt.Println("╠════════════════════════════════════════╣")
	fmt.Printf("║ Booking ID: %-26s ║\n", booking.id)
	fmt.Printf("║ Movie: %-31s ║\n", booking.show.GetMovie().GetTitle())
	fmt.Printf("║ Time: %-32s ║\n", booking.show.GetStartTime().Format("02 Jan, 15:04"))
	fmt.Printf("║ Screen: %-30s ║\n", booking.show.GetScreen().GetName())

	// Build seats string
	seatsStr := ""
	for i, seat := range booking.seats {
		if i > 0 {
			seatsStr += ", "
		}
		seatsStr += seat.GetID()
	}
	fmt.Printf("║ Seats: %-31s ║\n", seatsStr)

	fmt.Printf("║ Amount: ₹%-28.2f ║\n", booking.totalAmount)
	fmt.Printf("║ Status: %-30s ║\n", booking.status)
	fmt.Println("╚════════════════════════════════════════╝")
}

// ==================== BOOKING SERVICE ====================
// BookingService is the main service layer that coordinates all booking operations.
// It acts as a facade, providing a simplified interface for:
// - Managing movies and theatres
// - Finding shows
// - Creating and managing bookings
//
// This follows the Service Layer pattern, centralizing business logic.

type BookingService struct {
	theatres map[string]*Theatre // theatreID -> Theatre
	movies   map[string]*Movie   // movieID -> Movie
	bookings map[string]*Booking // bookingID -> Booking
	mutex    sync.RWMutex        // Protects maps from concurrent access
}

// NewBookingService creates a new booking service instance.
func NewBookingService() *BookingService {
	return &BookingService{
		theatres: make(map[string]*Theatre),
		movies:   make(map[string]*Movie),
		bookings: make(map[string]*Booking),
	}
}

// AddMovie registers a new movie in the system.
func (service *BookingService) AddMovie(movie *Movie) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.movies[movie.GetID()] = movie
}

// AddTheatre registers a new theatre in the system.
func (service *BookingService) AddTheatre(theatre *Theatre) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.theatres[theatre.id] = theatre
}

// GetMovies returns all registered movies.
func (service *BookingService) GetMovies() []*Movie {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	movies := make([]*Movie, 0, len(service.movies))
	for _, movie := range service.movies {
		movies = append(movies, movie)
	}
	return movies
}

// GetTheatresInCity returns all theatres in a specific city.
func (service *BookingService) GetTheatresInCity(city string) []*Theatre {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	theatresInCity := make([]*Theatre, 0)
	for _, theatre := range service.theatres {
		if theatre.city == city {
			theatresInCity = append(theatresInCity, theatre)
		}
	}
	return theatresInCity
}

// GetShowsForMovie finds all shows for a specific movie in a city.
func (service *BookingService) GetShowsForMovie(movieID, city string) []*Show {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	allShows := make([]*Show, 0)
	for _, theatre := range service.theatres {
		if theatre.city == city {
			theatreShows := theatre.GetShowsForMovie(movieID)
			allShows = append(allShows, theatreShows...)
		}
	}
	return allShows
}

// BookTickets creates a booking for the specified seats at a show.
// This is the main booking flow:
// 1. Attempt to reserve the seats (atomic operation)
// 2. Find the actual seat objects
// 3. Create and store the booking
//
// Returns the booking on success, or an error if seats are unavailable.
func (service *BookingService) BookTickets(user *User, show *Show, seatIDs []string) (*Booking, error) {
	// Step 1: Try to book seats (this is the critical section)
	// If any seat is already booked, this will fail and return an error
	if err := show.BookSeats(seatIDs); err != nil {
		return nil, err
	}

	// Step 2: Find the actual seat objects for the booked seat IDs
	// We need the seat objects to calculate price and store in booking
	seatsByID := make(map[string]*Seat)
	for _, seat := range show.GetScreen().GetSeats() {
		seatsByID[seat.GetID()] = seat
	}

	bookedSeats := make([]*Seat, 0, len(seatIDs))
	for _, seatID := range seatIDs {
		if seat, found := seatsByID[seatID]; found {
			bookedSeats = append(bookedSeats, seat)
		}
	}

	// Step 3: Create the booking and store it
	booking := NewBooking(user, show, bookedSeats)

	service.mutex.Lock()
	service.bookings[booking.GetID()] = booking
	service.mutex.Unlock()

	return booking, nil
}

// ConfirmBooking marks a pending booking as confirmed.
// This is typically called after successful payment.
func (service *BookingService) ConfirmBooking(bookingID string) error {
	service.mutex.RLock()
	booking, exists := service.bookings[bookingID]
	service.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("booking not found: %s", bookingID)
	}

	booking.Confirm()
	return nil
}

// CancelBooking cancels a booking and releases the seats.
// The released seats become available for other users.
func (service *BookingService) CancelBooking(bookingID string) error {
	service.mutex.RLock()
	booking, exists := service.bookings[bookingID]
	service.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("booking not found: %s", bookingID)
	}

	booking.Cancel()
	return nil
}

// GetBooking retrieves a booking by its ID.
func (service *BookingService) GetBooking(bookingID string) (*Booking, error) {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	booking, exists := service.bookings[bookingID]
	if !exists {
		return nil, fmt.Errorf("booking not found: %s", bookingID)
	}
	return booking, nil
}

// ========== MAIN ==========

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("    🎬 BOOKMYSHOW - Ticket Booking System")
	fmt.Println("═══════════════════════════════════════════")

	// Initialize service
	service := NewBookingService()

	// Add movies
	movie1 := NewMovie("M1", "Avengers: Endgame", "Action", "English", 180)
	movie2 := NewMovie("M2", "Inception", "Sci-Fi", "English", 150)
	service.AddMovie(movie1)
	service.AddMovie(movie2)

	// Create theatre with screens
	theatre := NewTheatre("T1", "PVR Cinemas", "Mumbai", "Phoenix Mall")
	screen1 := NewScreen("S1", "Screen 1", []string{"A", "B", "C", "D", "E", "F"}, 10)
	screen2 := NewScreen("S2", "Screen 2", []string{"A", "B", "C", "D"}, 8)
	theatre.AddScreen(screen1)
	theatre.AddScreen(screen2)

	// Add shows
	today := time.Now()
	show1 := NewShow("SH1", movie1, screen1, time.Date(today.Year(), today.Month(), today.Day(), 14, 0, 0, 0, time.Local))
	show2 := NewShow("SH2", movie1, screen1, time.Date(today.Year(), today.Month(), today.Day(), 18, 0, 0, 0, time.Local))
	show3 := NewShow("SH3", movie2, screen2, time.Date(today.Year(), today.Month(), today.Day(), 16, 0, 0, 0, time.Local))
	theatre.AddShow(show1)
	theatre.AddShow(show2)
	theatre.AddShow(show3)

	service.AddTheatre(theatre)

	// Display available shows
	fmt.Println("\n📽️  Now Showing in Mumbai:")
	fmt.Println("─────────────────────────────────────────")
	for _, show := range theatre.GetShows() {
		fmt.Printf("  • %s\n", show)
	}

	// Create user
	user := NewUser("U1", "John Doe", "john@email.com", "9876543210")

	// Book tickets
	fmt.Println("\n🎫 Booking Tickets...")
	fmt.Println("─────────────────────────────────────────")

	booking1, err := service.BookTickets(user, show1, []string{"A1", "A2", "A3"})
	if err != nil {
		fmt.Printf("Booking failed: %v\n", err)
	} else {
		fmt.Printf("✅ Booking created: %s\n", booking1.GetID())

		// Simulate payment and confirm
		if confirmErr := service.ConfirmBooking(booking1.GetID()); confirmErr != nil {
			fmt.Printf("Failed to confirm booking: %v\n", confirmErr)
		}
		booking1.PrintTicket()
	}

	// Show updated availability
	fmt.Println("\n📊 Updated Availability:")
	fmt.Printf("  %s\n", show1)

	// Try to book same seats again (should fail)
	fmt.Println("\n⚠️  Trying to book same seats again...")
	_, err = service.BookTickets(user, show1, []string{"A1", "A2"})
	if err != nil {
		fmt.Printf("  ❌ Expected error: %v\n", err)
	}

	// Book different seats
	fmt.Println("\n🎫 Booking different seats...")
	booking2, err := service.BookTickets(user, show1, []string{"B5", "B6"})
	if err != nil {
		fmt.Printf("Booking failed: %v\n", err)
	} else {
		if confirmErr := service.ConfirmBooking(booking2.GetID()); confirmErr != nil {
			fmt.Printf("Failed to confirm booking: %v\n", confirmErr)
		}
		booking2.PrintTicket()
	}

	// Cancel a booking
	fmt.Println("\n❌ Cancelling first booking...")
	if cancelErr := service.CancelBooking(booking1.GetID()); cancelErr != nil {
		fmt.Printf("Failed to cancel booking: %v\n", cancelErr)
	} else {
		fmt.Printf("  Booking %s cancelled\n", booking1.GetID())
	}
	fmt.Printf("  Updated: %s\n", show1)

	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  KEY DESIGN DECISIONS:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. Show owns seat booking state")
	fmt.Println("  2. Mutex for concurrent booking safety")
	fmt.Println("  3. Booking status flow: Pending→Confirmed")
	fmt.Println("  4. Seat release on cancellation")
	fmt.Println("═══════════════════════════════════════════")
}
