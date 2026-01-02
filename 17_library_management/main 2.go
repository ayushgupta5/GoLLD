package main

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// LIBRARY MANAGEMENT SYSTEM - Low Level Design
// ============================================================
//
// This system demonstrates a typical library management scenario with:
// - Books and their physical copies
// - Members who can borrow books
// - Loan tracking with due dates and fines
// - Search functionality by title, author, or ISBN
// - Reservation system for unavailable books
//
// Key Concepts for Beginners:
// 1. Entity separation (Book vs BookCopy)
// 2. State management using enums
// 3. Thread safety with mutexes
// 4. Fine calculation for overdue books

// =====================================================
// ENUMS - Status types for books and members
// =====================================================

// BookStatus represents the current state of a book copy
type BookStatus int

const (
	BookStatusAvailable BookStatus = iota // Book is available for borrowing
	BookStatusLoaned                      // Book is currently loaned out
	BookStatusReserved                    // Book is reserved by a member
	BookStatusLost                        // Book is marked as lost
)

// String returns a human-readable representation of BookStatus
func (status BookStatus) String() string {
	statusNames := [...]string{"Available", "Loaned", "Reserved", "Lost"}
	if status < 0 || int(status) >= len(statusNames) {
		return "Unknown"
	}
	return statusNames[status]
}

// MemberStatus represents the current state of a library member
type MemberStatus int

const (
	MemberStatusActive  MemberStatus = iota // Member can borrow books
	MemberStatusBlocked                     // Member is blocked (e.g., unpaid fines)
	MemberStatusClosed                      // Membership is closed
)

// String returns a human-readable representation of MemberStatus
func (status MemberStatus) String() string {
	statusNames := [...]string{"Active", "Blocked", "Closed"}
	if status < 0 || int(status) >= len(statusNames) {
		return "Unknown"
	}
	return statusNames[status]
}

// =====================================================
// BOOK - Represents a book title in the library
// =====================================================

// Book represents a book title with all its metadata.
// A single Book can have multiple physical copies (BookCopy).
// For example, "Clean Code" might have 3 physical copies.
type Book struct {
	isbn        string      // Unique identifier for the book
	title       string      // Title of the book
	author      string      // Author name
	publisher   string      // Publisher name
	totalCopies int         // Total number of physical copies
	copies      []*BookCopy // List of physical copies
}

// NewBook creates a new book with the specified number of copies.
// Each copy gets a unique ID in the format "ISBN-copyNumber".
func NewBook(isbn, title, author, publisher string, totalCopies int) *Book {
	book := &Book{
		isbn:        isbn,
		title:       title,
		author:      author,
		publisher:   publisher,
		totalCopies: totalCopies,
		copies:      make([]*BookCopy, 0, totalCopies), // Pre-allocate capacity
	}

	// Create physical copies of this book
	for copyNumber := 1; copyNumber <= totalCopies; copyNumber++ {
		bookCopy := &BookCopy{
			id:     fmt.Sprintf("%s-%d", isbn, copyNumber),
			book:   book,
			status: BookStatusAvailable,
		}
		book.copies = append(book.copies, bookCopy)
	}
	return book
}

// GetAvailableCopy returns the first available copy of this book.
// Returns nil if no copies are available.
func (book *Book) GetAvailableCopy() *BookCopy {
	for _, bookCopy := range book.copies {
		if bookCopy.status == BookStatusAvailable {
			return bookCopy
		}
	}
	return nil
}

// GetAvailableCount returns the number of copies currently available.
func (book *Book) GetAvailableCount() int {
	availableCount := 0
	for _, bookCopy := range book.copies {
		if bookCopy.status == BookStatusAvailable {
			availableCount++
		}
	}
	return availableCount
}

// String returns a formatted string representation of the book.
func (book *Book) String() string {
	return fmt.Sprintf("%s by %s (ISBN: %s) - %d/%d available",
		book.title, book.author, book.isbn, book.GetAvailableCount(), book.totalCopies)
}

// =====================================================
// BOOK COPY - Represents a physical copy of a book
// =====================================================

// BookCopy represents a single physical copy of a book.
// While "Clean Code" is a Book, each physical copy on the shelf is a BookCopy.
type BookCopy struct {
	id       string     // Unique identifier for this copy
	book     *Book      // Reference to the parent Book
	status   BookStatus // Current status of this copy
	loanedTo *Member    // Member who has this copy (if loaned)
	dueDate  time.Time  // When this copy is due back (if loaned)
}

// IsAvailable checks if this copy is available for borrowing.
func (bookCopy *BookCopy) IsAvailable() bool {
	return bookCopy.status == BookStatusAvailable
}

// =====================================================
// MEMBER - Represents a library member
// =====================================================

// Member represents a person registered with the library.
// Members can borrow books, have fines, and be in different statuses.
type Member struct {
	id            string       // Unique member ID
	name          string       // Full name
	email         string       // Email address
	phone         string       // Phone number
	status        MemberStatus // Current membership status
	borrowedBooks []*BookLoan  // List of current loans
	maxBooks      int          // Maximum books allowed to borrow
	fineAmount    float64      // Outstanding fine amount
	mu            sync.Mutex   // Mutex for thread-safe operations
}

// NewMember creates a new library member with default settings.
func NewMember(id, name, email, phone string) *Member {
	return &Member{
		id:            id,
		name:          name,
		email:         email,
		phone:         phone,
		status:        MemberStatusActive,
		borrowedBooks: make([]*BookLoan, 0),
		maxBooks:      5, // Default borrowing limit
		fineAmount:    0,
	}
}

// CanBorrow checks if the member is eligible to borrow books.
// A member can borrow if:
// - Their status is Active
// - They haven't reached their borrowing limit
// - They have no outstanding fines
func (member *Member) CanBorrow() bool {
	member.mu.Lock()
	defer member.mu.Unlock()

	isActive := member.status == MemberStatusActive
	hasCapacity := len(member.borrowedBooks) < member.maxBooks
	hasNoFines := member.fineAmount == 0

	return isActive && hasCapacity && hasNoFines
}

// GetBorrowedCount returns the number of books currently borrowed.
func (member *Member) GetBorrowedCount() int {
	member.mu.Lock()
	defer member.mu.Unlock()
	return len(member.borrowedBooks)
}

// AddLoan adds a new loan to the member's borrowed books list.
func (member *Member) AddLoan(loan *BookLoan) {
	member.mu.Lock()
	defer member.mu.Unlock()
	member.borrowedBooks = append(member.borrowedBooks, loan)
}

// RemoveLoan removes a loan from the member's borrowed books list.
func (member *Member) RemoveLoan(loanID string) {
	member.mu.Lock()
	defer member.mu.Unlock()

	for index, loan := range member.borrowedBooks {
		if loan.id == loanID {
			// Remove the loan by creating a new slice without it
			member.borrowedBooks = append(
				member.borrowedBooks[:index],
				member.borrowedBooks[index+1:]...,
			)
			return
		}
	}
}

// AddFine adds a fine amount to the member's account.
func (member *Member) AddFine(amount float64) {
	member.mu.Lock()
	defer member.mu.Unlock()
	member.fineAmount += amount
}

// PayFine reduces the member's fine by the payment amount.
func (member *Member) PayFine(amount float64) {
	member.mu.Lock()
	defer member.mu.Unlock()
	member.fineAmount -= amount
	if member.fineAmount < 0 {
		member.fineAmount = 0 // Prevent negative fines
	}
}

// GetFineAmount returns the current outstanding fine amount.
func (member *Member) GetFineAmount() float64 {
	member.mu.Lock()
	defer member.mu.Unlock()
	return member.fineAmount
}

// =====================================================
// BOOK LOAN - Represents a borrowing transaction
// =====================================================

// BookLoan represents a single borrowing transaction.
// It tracks when a book was borrowed, when it's due, and any fines.
type BookLoan struct {
	id         string    // Unique loan ID (e.g., "LOAN-1")
	bookCopy   *BookCopy // The physical copy that was borrowed
	member     *Member   // The member who borrowed the book
	issueDate  time.Time // When the book was borrowed
	dueDate    time.Time // When the book should be returned
	returnDate time.Time // When the book was actually returned (zero if not returned)
	isReturned bool      // Whether the book has been returned
	fine       float64   // Fine amount (if any)
}

// loanCounter is a thread-safe counter for generating unique loan IDs.
var loanCounter int64

// NewBookLoan creates a new loan record.
func NewBookLoan(bookCopy *BookCopy, member *Member, loanDays int) *BookLoan {
	// Generate unique loan ID using atomic increment for thread safety
	loanID := atomic.AddInt64(&loanCounter, 1)
	currentTime := time.Now()

	return &BookLoan{
		id:         fmt.Sprintf("LOAN-%d", loanID),
		bookCopy:   bookCopy,
		member:     member,
		issueDate:  currentTime,
		dueDate:    currentTime.AddDate(0, 0, loanDays), // Add loanDays to current date
		isReturned: false,
		fine:       0,
	}
}

// IsOverdue checks if the book is past its due date.
// Returns false if the book has already been returned.
func (loan *BookLoan) IsOverdue() bool {
	if loan.isReturned {
		return false
	}
	return time.Now().After(loan.dueDate)
}

// CalculateFine calculates the fine based on how many days overdue.
// Returns 0 if the book is not overdue.
func (loan *BookLoan) CalculateFine(finePerDay float64) float64 {
	if !loan.IsOverdue() {
		return 0
	}
	// Calculate number of overdue days
	overdueDuration := time.Since(loan.dueDate)
	overdueDays := int(overdueDuration.Hours() / 24)
	return float64(overdueDays) * finePerDay
}

// GetDaysUntilDue returns the number of days until the book is due.
// Returns negative number if overdue.
func (loan *BookLoan) GetDaysUntilDue() int {
	if loan.isReturned {
		return 0
	}
	duration := time.Until(loan.dueDate)
	return int(duration.Hours() / 24)
}

// =====================================================
// RESERVATION - Represents a book reservation
// =====================================================

// Reservation represents a member's request to reserve a book
// that is currently not available. When a copy becomes available,
// the reservation can be fulfilled.
type Reservation struct {
	id          string    // Unique reservation ID (e.g., "RES-1")
	book        *Book     // The book being reserved
	member      *Member   // The member who made the reservation
	createdAt   time.Time // When the reservation was made
	expiresAt   time.Time // When the reservation expires
	isFulfilled bool      // Whether the reservation has been fulfilled
}

// reservationCounter is a thread-safe counter for generating unique reservation IDs.
var reservationCounter int64

// NewReservation creates a new book reservation.
// Reservations expire after 7 days if not fulfilled.
func NewReservation(book *Book, member *Member) *Reservation {
	reservationID := atomic.AddInt64(&reservationCounter, 1)
	currentTime := time.Now()

	return &Reservation{
		id:          fmt.Sprintf("RES-%d", reservationID),
		book:        book,
		member:      member,
		createdAt:   currentTime,
		expiresAt:   currentTime.AddDate(0, 0, 7), // 7 days to pickup
		isFulfilled: false,
	}
}

// IsExpired checks if the reservation has expired.
func (reservation *Reservation) IsExpired() bool {
	return time.Now().After(reservation.expiresAt) && !reservation.isFulfilled
}

// IsValid checks if the reservation is still valid (not expired and not fulfilled).
func (reservation *Reservation) IsValid() bool {
	return !reservation.isFulfilled && !reservation.IsExpired()
}

// =====================================================
// LIBRARY - Main system that manages books and members
// =====================================================

// Library represents the main library system.
// It manages books, members, loans, and reservations.
type Library struct {
	name         string               // Name of the library
	books        map[string]*Book     // ISBN -> Book mapping
	members      map[string]*Member   // MemberID -> Member mapping
	loans        map[string]*BookLoan // LoanID -> BookLoan mapping
	reservations []*Reservation       // List of all reservations
	finePerDay   float64              // Fine amount per day for overdue books
	loanDays     int                  // Default number of days for a loan
	mu           sync.RWMutex         // Mutex for thread-safe operations
}

// NewLibrary creates a new library with default settings.
func NewLibrary(name string) *Library {
	return &Library{
		name:         name,
		books:        make(map[string]*Book),
		members:      make(map[string]*Member),
		loans:        make(map[string]*BookLoan),
		reservations: make([]*Reservation, 0),
		finePerDay:   1.0, // $1 per day
		loanDays:     14,  // 2 weeks
	}
}

// AddBook adds a new book to the library catalog.
func (lib *Library) AddBook(book *Book) {
	lib.mu.Lock()
	defer lib.mu.Unlock()
	lib.books[book.isbn] = book
}

// RemoveBook removes a book from the library catalog.
func (lib *Library) RemoveBook(isbn string) error {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	book, exists := lib.books[isbn]
	if !exists {
		return fmt.Errorf("book not found: %s", isbn)
	}

	// Check if any copies are currently loaned
	for _, bookCopy := range book.copies {
		if bookCopy.status == BookStatusLoaned {
			return fmt.Errorf("cannot remove book: some copies are currently on loan")
		}
	}

	delete(lib.books, isbn)
	return nil
}

// RegisterMember registers a new member in the library.
func (lib *Library) RegisterMember(member *Member) {
	lib.mu.Lock()
	defer lib.mu.Unlock()
	lib.members[member.id] = member
}

// GetMember retrieves a member by their ID.
func (lib *Library) GetMember(memberID string) (*Member, bool) {
	lib.mu.RLock()
	defer lib.mu.RUnlock()
	member, exists := lib.members[memberID]
	return member, exists
}

// SearchByTitle searches for books by title (case-insensitive partial match).
func (lib *Library) SearchByTitle(title string) []*Book {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	results := make([]*Book, 0)
	searchTerm := strings.ToLower(title)

	for _, book := range lib.books {
		if strings.Contains(strings.ToLower(book.title), searchTerm) {
			results = append(results, book)
		}
	}
	return results
}

// SearchByAuthor searches for books by author (case-insensitive partial match).
func (lib *Library) SearchByAuthor(author string) []*Book {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	results := make([]*Book, 0)
	searchTerm := strings.ToLower(author)

	for _, book := range lib.books {
		if strings.Contains(strings.ToLower(book.author), searchTerm) {
			results = append(results, book)
		}
	}
	return results
}

// GetBookByISBN retrieves a book by its ISBN.
func (lib *Library) GetBookByISBN(isbn string) *Book {
	lib.mu.RLock()
	defer lib.mu.RUnlock()
	return lib.books[isbn]
}

// IssueBook issues a book copy to a member.
// Returns the loan record if successful, or an error if:
// - Member doesn't exist
// - Member can't borrow (limit reached or has fines)
// - Book doesn't exist
// - No copies are available
func (lib *Library) IssueBook(memberID, isbn string) (*BookLoan, error) {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	// Step 1: Validate the member exists
	member, exists := lib.members[memberID]
	if !exists {
		return nil, fmt.Errorf("member not found: %s", memberID)
	}

	// Step 2: Check if the member can borrow
	if !member.CanBorrow() {
		return nil, fmt.Errorf("member cannot borrow (limit reached or has fines)")
	}

	// Step 3: Validate the book exists
	book, exists := lib.books[isbn]
	if !exists {
		return nil, fmt.Errorf("book not found: %s", isbn)
	}

	// Step 4: Find an available copy
	bookCopy := book.GetAvailableCopy()
	if bookCopy == nil {
		return nil, fmt.Errorf("no copies available for: %s", book.title)
	}

	// Step 5: Create the loan record
	loan := NewBookLoan(bookCopy, member, lib.loanDays)

	// Step 6: Update the book copy status
	bookCopy.status = BookStatusLoaned
	bookCopy.loanedTo = member
	bookCopy.dueDate = loan.dueDate

	// Step 7: Add the loan to the member and library
	member.AddLoan(loan)
	lib.loans[loan.id] = loan

	return loan, nil
}

// ReturnBook processes a book return and calculates any fines.
// Returns the fine amount if successful, or an error if:
// - Loan doesn't exist
// - Book was already returned
func (lib *Library) ReturnBook(loanID string) (float64, error) {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	// Step 1: Find the loan
	loan, exists := lib.loans[loanID]
	if !exists {
		return 0, fmt.Errorf("loan not found: %s", loanID)
	}

	// Step 2: Check if already returned
	if loan.isReturned {
		return 0, fmt.Errorf("book already returned")
	}

	// Step 3: Calculate fine for overdue returns
	fine := loan.CalculateFine(lib.finePerDay)
	loan.fine = fine
	loan.isReturned = true
	loan.returnDate = time.Now()

	// Step 4: Update book copy status (make it available again)
	loan.bookCopy.status = BookStatusAvailable
	loan.bookCopy.loanedTo = nil
	loan.bookCopy.dueDate = time.Time{} // Reset due date

	// Step 5: Update member's records
	loan.member.RemoveLoan(loanID)
	if fine > 0 {
		loan.member.AddFine(fine)
	}

	// Step 6: Check for pending reservations (notify if needed)
	lib.checkPendingReservations(loan.bookCopy.book)

	return fine, nil
}

// checkPendingReservations checks if there are pending reservations for a book.
// This is called when a book is returned to potentially fulfill a reservation.
func (lib *Library) checkPendingReservations(book *Book) {
	for _, reservation := range lib.reservations {
		if reservation.book.isbn == book.isbn && reservation.IsValid() {
			// In a real system, we would notify the member here
			fmt.Printf("📧 Notification: %s, '%s' is now available for pickup!\n",
				reservation.member.name, book.title)
			break // Only notify the first valid reservation
		}
	}
}

// ReserveBook creates a reservation for a book that's not currently available.
// Returns an error if:
// - Member doesn't exist
// - Book doesn't exist
// - Book is already available (no need to reserve)
func (lib *Library) ReserveBook(memberID, isbn string) (*Reservation, error) {
	lib.mu.Lock()
	defer lib.mu.Unlock()

	// Step 1: Validate the member exists
	member, exists := lib.members[memberID]
	if !exists {
		return nil, fmt.Errorf("member not found: %s", memberID)
	}

	// Step 2: Validate the book exists
	book, exists := lib.books[isbn]
	if !exists {
		return nil, fmt.Errorf("book not found: %s", isbn)
	}

	// Step 3: Check if the book is already available
	if book.GetAvailableCopy() != nil {
		return nil, fmt.Errorf("book is available, no need to reserve")
	}

	// Step 4: Create the reservation
	reservation := NewReservation(book, member)
	lib.reservations = append(lib.reservations, reservation)

	return reservation, nil
}

// GetMemberLoans returns all active (not returned) loans for a member.
func (lib *Library) GetMemberLoans(memberID string) []*BookLoan {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	activeLoans := make([]*BookLoan, 0)
	for _, loan := range lib.loans {
		if loan.member.id == memberID && !loan.isReturned {
			activeLoans = append(activeLoans, loan)
		}
	}
	return activeLoans
}

// GetOverdueLoans returns all loans that are currently overdue.
func (lib *Library) GetOverdueLoans() []*BookLoan {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	overdueLoans := make([]*BookLoan, 0)
	for _, loan := range lib.loans {
		if loan.IsOverdue() {
			overdueLoans = append(overdueLoans, loan)
		}
	}
	return overdueLoans
}

// PrintCatalog prints all books in the library catalog.
func (lib *Library) PrintCatalog() {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	fmt.Printf("\n📚 %s - Book Catalog\n", lib.name)
	fmt.Println("─────────────────────────────────────────")

	if len(lib.books) == 0 {
		fmt.Println("  No books in catalog")
		return
	}

	for _, book := range lib.books {
		// Show green dot for available, red for all copies loaned
		availabilityIndicator := "🟢" // Available
		if book.GetAvailableCount() == 0 {
			availabilityIndicator = "🔴" // All loaned out
		}
		fmt.Printf("  %s %s\n", availabilityIndicator, book)
	}
}

// ========== MAIN ==========

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("       📚 LIBRARY MANAGEMENT SYSTEM")
	fmt.Println("═══════════════════════════════════════════")

	// Create library
	library := NewLibrary("City Central Library")

	// Add books
	library.AddBook(NewBook("978-0134190440", "Clean Code", "Robert C. Martin", "Pearson", 3))
	library.AddBook(NewBook("978-0201633610", "Design Patterns", "Gang of Four", "Addison-Wesley", 2))
	library.AddBook(NewBook("978-0596007126", "Head First Design Patterns", "Eric Freeman", "O'Reilly", 4))
	library.AddBook(NewBook("978-0132350884", "Clean Architecture", "Robert C. Martin", "Pearson", 2))
	library.AddBook(NewBook("978-1617294549", "Go in Action", "William Kennedy", "Manning", 3))

	// Register members
	member1 := NewMember("M001", "John Doe", "john@email.com", "555-0101")
	member2 := NewMember("M002", "Jane Smith", "jane@email.com", "555-0102")
	library.RegisterMember(member1)
	library.RegisterMember(member2)

	library.PrintCatalog()

	// Search books
	fmt.Println("\n🔍 Searching for 'Clean'...")
	results := library.SearchByTitle("Clean")
	for _, book := range results {
		fmt.Printf("  Found: %s\n", book.title)
	}

	// Issue books
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("📖 Issuing books...")

	loan1, err := library.IssueBook("M001", "978-0134190440")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("✅ Issued '%s' to %s\n", loan1.bookCopy.book.title, member1.name)
		fmt.Printf("   Due date: %s\n", loan1.dueDate.Format("Jan 02, 2006"))
	}

	loan2, err := library.IssueBook("M001", "978-0201633610")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("✅ Issued '%s' to %s\n", loan2.bookCopy.book.title, member1.name)
	}

	loan3, err := library.IssueBook("M002", "978-1617294549")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("✅ Issued '%s' to %s\n", loan3.bookCopy.book.title, member2.name)
	}

	library.PrintCatalog()

	// Show member's books
	fmt.Println("\n📋 John's borrowed books:")
	for _, loan := range library.GetMemberLoans("M001") {
		fmt.Printf("  • %s (Due: %s)\n",
			loan.bookCopy.book.title,
			loan.dueDate.Format("Jan 02"))
	}

	// Return a book
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("📥 Returning book...")

	fine, err := library.ReturnBook(loan1.id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("✅ Returned '%s'\n", loan1.bookCopy.book.title)
		if fine > 0 {
			fmt.Printf("   Fine: $%.2f\n", fine)
		}
	}

	library.PrintCatalog()

	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  KEY DESIGN DECISIONS:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. Book vs BookCopy separation")
	fmt.Println("  2. Member borrowing limits")
	fmt.Println("  3. Fine calculation for overdue")
	fmt.Println("  4. Search by title/author/ISBN")
	fmt.Println("═══════════════════════════════════════════")
}
