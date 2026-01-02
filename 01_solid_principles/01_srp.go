package main

import (
	"errors"
	"fmt"
	"strings"
)

// ============================================================
// SINGLE RESPONSIBILITY PRINCIPLE (SRP)
// ============================================================
//
// WHAT IS SRP?
// ------------
// "A struct (or class) should have only ONE reason to change"
//
// Think of it like a job:
// - A chef cooks food (one job)
// - A waiter serves food (one job)
// - A cashier handles payments (one job)
//
// If one person did ALL these jobs, any change to cooking, serving,
// or payment rules would affect that one person. That's messy!
//
// Similarly, each struct in our code should have ONE specific job.
//
// ============================================================

// ============================================================
// ❌ BAD EXAMPLE - Violates SRP (DON'T DO THIS)
// ============================================================
//
// This BadUser struct is trying to do TOO MANY things:
//   Job 1: Store user information (ID, Name, Email)
//   Job 2: Validate the user data
//   Job 3: Save user to database
//   Job 4: Send welcome emails
//
// WHY IS THIS BAD?
// ----------------
// 1. If we change how emails are sent → we modify BadUser
// 2. If we change the database → we modify BadUser
// 3. If we add new validation rules → we modify BadUser
// 4. Testing is hard → need to set up database AND email just to test validation!

// BadUser - Does too many things (violates SRP)
type BadUser struct {
	ID    int    // Unique identifier for the user
	Name  string // User's full name
	Email string // User's email address
}

// Validate checks if the user data is valid
// PROBLEM: Validation logic is inside the User struct
func (bu *BadUser) Validate() bool {
	hasName := bu.Name != ""
	hasEmail := bu.Email != ""
	return hasName && hasEmail
}

// SaveToDatabase saves the user to a database
// PROBLEM: Database logic is inside the User struct
func (bu *BadUser) SaveToDatabase() {
	fmt.Println("💾 [BAD] Saving to database:", bu.Name)
}

// SendWelcomeEmail sends a welcome email to the user
// PROBLEM: Email logic is inside the User struct
func (bu *BadUser) SendWelcomeEmail() {
	fmt.Println("📧 [BAD] Sending email to:", bu.Email)
}

// ============================================================
// ✅ GOOD EXAMPLE - Follows SRP (DO THIS!)
// ============================================================
//
// Now we split responsibilities into separate structs:
//   - User           → Only stores user data
//   - UserValidator  → Only validates user data
//   - UserRepository → Only handles database operations
//   - EmailService   → Only sends emails
//   - UserService    → Coordinates all the above
//
// Each struct has ONE job = ONE reason to change!
//
// We also use INTERFACES to make components replaceable and testable.
//
// ============================================================

// ------------------------------------------------------------
// CUSTOM ERRORS - Clear error messages for validation
// ------------------------------------------------------------
// Using custom errors makes it easy to identify what went wrong

var (
	// ErrEmptyName is returned when user name is empty
	ErrEmptyName = errors.New("name cannot be empty")

	// ErrEmptyEmail is returned when user email is empty
	ErrEmptyEmail = errors.New("email cannot be empty")

	// ErrInvalidEmail is returned when email format is invalid
	ErrInvalidEmail = errors.New("email must contain '@' symbol")
)

// ------------------------------------------------------------
// INTERFACES - Define contracts for each responsibility
// ------------------------------------------------------------
// Interfaces allow us to swap implementations easily (great for testing!)

// Validator defines the contract for user validation
type Validator interface {
	Validate(user *User) error
}

// Repository defines the contract for data storage operations
type Repository interface {
	Save(user *User) error
	FindByID(id int) (*User, error)
}

// UserEmailNotifier defines the contract for sending email notifications to users
type UserEmailNotifier interface {
	SendWelcomeEmail(user *User) error
}

// ------------------------------------------------------------
// STEP 1: User struct - Stores user data ONLY
// ------------------------------------------------------------
// Job: Hold user information (nothing else!)
type User struct {
	ID    int    // Unique identifier for the user
	Name  string // User's full name
	Email string // User's email address
}

// String returns a readable representation of the User
// This is Go's way of implementing toString()
func (u *User) String() string {
	return fmt.Sprintf("User{ID: %d, Name: %q, Email: %q}", u.ID, u.Name, u.Email)
}

// ------------------------------------------------------------
// STEP 2: UserValidator - Validates user data ONLY
// ------------------------------------------------------------
// Job: Check if user data meets our rules
type UserValidator struct {
	// This struct doesn't need any fields
	// It just contains validation methods
}

// NewUserValidator creates a new UserValidator instance
func NewUserValidator() *UserValidator {
	return &UserValidator{}
}

// Validate checks if a user's data is valid
// Returns an error if something is wrong, nil if everything is OK
func (v *UserValidator) Validate(user *User) error {
	// Check 1: Name must not be empty
	if strings.TrimSpace(user.Name) == "" {
		return ErrEmptyName
	}

	// Check 2: Email must not be empty
	if strings.TrimSpace(user.Email) == "" {
		return ErrEmptyEmail
	}

	// Check 3: Email must contain '@' symbol (basic format check)
	if !strings.Contains(user.Email, "@") {
		return ErrInvalidEmail
	}

	// All checks passed!
	return nil
}

// ------------------------------------------------------------
// STEP 3: UserRepository - Handles database operations ONLY
// ------------------------------------------------------------
// Job: Save and retrieve users from the database
type UserRepository struct {
	// In a real application, this would have:
	// db *sql.DB  // database connection

	// For demo purposes, we use an in-memory map to store users
	users map[int]*User
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository() *UserRepository {
	return &UserRepository{
		users: make(map[int]*User),
	}
}

// Save stores a user in the database
func (r *UserRepository) Save(user *User) error {
	fmt.Printf("💾 [GOOD] Saving user '%s' to database\n", user.Name)
	// In real code: INSERT INTO users (name, email) VALUES (?, ?)

	// Store in our in-memory map (simulating database)
	r.users[user.ID] = user
	return nil
}

// FindByID retrieves a user from the database by their ID
func (r *UserRepository) FindByID(id int) (*User, error) {
	fmt.Printf("🔍 [GOOD] Looking for user with ID: %d\n", id)
	// In real code: SELECT * FROM users WHERE id = ?

	// Look up in our in-memory map
	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}
	return user, nil
}

// ------------------------------------------------------------
// STEP 4: EmailService - Sends emails ONLY
// ------------------------------------------------------------
// Job: Send emails to users
type EmailService struct {
	// In a real application, this would have:
	// smtpServer string  // email server address
	// apiKey     string  // email service API key
}

// NewEmailService creates a new EmailService instance
func NewEmailService() *EmailService {
	return &EmailService{}
}

// SendWelcomeEmail sends a welcome email to a new user
func (e *EmailService) SendWelcomeEmail(user *User) error {
	fmt.Printf("📧 [GOOD] Sending welcome email to '%s' at '%s'\n", user.Name, user.Email)
	// In real code: Use email library to send actual email
	return nil
}

// ------------------------------------------------------------
// STEP 5: UserService - Coordinates all the pieces
// ------------------------------------------------------------
// Job: Orchestrate the user creation process
// This is like a manager that tells other specialists what to do
type UserService struct {
	validator  Validator         // Handles validation (interface type)
	repository Repository        // Handles database (interface type)
	notifier   UserEmailNotifier // Handles notifications (interface type)
	nextUserID int               // Tracks the next available user ID
}

// NewUserService creates a new UserService with all its dependencies
// This is called a "constructor function" in Go
//
// By accepting interfaces, we can inject different implementations:
// - Real implementations for production
// - Mock implementations for testing
func NewUserService(validator Validator, repository Repository, notifier UserEmailNotifier) *UserService {
	return &UserService{
		validator:  validator,
		repository: repository,
		notifier:   notifier,
		nextUserID: 1, // Start user IDs at 1
	}
}

// CreateUser creates a new user by coordinating all the steps
func (s *UserService) CreateUser(name, email string) (*User, error) {
	// Create a temporary user object for validation
	// We don't assign the ID yet because validation might fail
	tempUser := &User{
		Name:  name,
		Email: email,
	}

	// STEP A: Validate the user data FIRST
	if err := s.validator.Validate(tempUser); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// STEP B: Assign ID only AFTER validation passes
	// This prevents wasting IDs on invalid users
	tempUser.ID = s.nextUserID
	s.nextUserID++

	// STEP C: Save to database
	if err := s.repository.Save(tempUser); err != nil {
		return nil, fmt.Errorf("could not save user: %w", err)
	}

	// STEP D: Send welcome email (non-critical operation)
	if err := s.notifier.SendWelcomeEmail(tempUser); err != nil {
		// Email failure is not critical, so we just log a warning
		// We don't return an error because the user was created successfully
		fmt.Printf("⚠️  Warning: Could not send welcome email: %v\n", err)
	}

	return tempUser, nil
}

// GetUser retrieves a user by their ID
func (s *UserService) GetUser(id int) (*User, error) {
	return s.repository.FindByID(id)
}

// ============================================================
// WHY IS SRP BETTER? (The Benefits)
// ============================================================
//
// 1. 🧪 EASY TO TEST:
//    - Test UserValidator without setting up a database
//    - Test UserRepository without an email server
//    - Each piece can be tested separately (isolation)
//    - Use mock implementations via interfaces
//
// 2. 🔧 EASY TO CHANGE:
//    - Want to change email provider? → Only modify EmailService
//    - Switching databases? → Only modify UserRepository
//    - New validation rule? → Only modify UserValidator
//
// 3. 📖 EASY TO READ:
//    - Clear what each struct does (one job each)
//    - Easy to find where specific logic lives
//    - New team members understand code faster
//
// 4. ♻️  REUSABLE:
//    - EmailService can send other types of emails
//    - UserRepository can be used by other services
//    - Components work like LEGO blocks!
//
// 5. 🔌 PLUGGABLE (New benefit from interfaces!):
//    - Swap EmailService with SMSService
//    - Change from MySQL to PostgreSQL
//    - Use mock services for testing
//
// ============================================================

// ============================================================
// MAIN FUNCTION - See SRP in Action!
// ============================================================

func main() {
	printHeader()

	demonstrateBadExample()

	demonstrateGoodExample()

	printSummary()
}

// printHeader displays the program title
func printHeader() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║     SINGLE RESPONSIBILITY PRINCIPLE (SRP) - DEMO          ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// demonstrateBadExample shows code that violates SRP
func demonstrateBadExample() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("❌ BAD EXAMPLE: One struct doing EVERYTHING")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	badUser := &BadUser{
		ID:    1,
		Name:  "Alice Smith",
		Email: "alice@example.com",
	}

	fmt.Printf("Created BadUser: %+v\n", badUser)
	fmt.Println()

	// All these methods are on the same struct - BAD!
	if badUser.Validate() {
		fmt.Println("✓ Validation passed")
		badUser.SaveToDatabase()   // Database logic in User - messy!
		badUser.SendWelcomeEmail() // Email logic in User - messy!
	}

	fmt.Println()
	fmt.Println("Problem: BadUser struct has 4 reasons to change!")
	fmt.Println("  1. Data structure changes")
	fmt.Println("  2. Validation rules change")
	fmt.Println("  3. Database changes")
	fmt.Println("  4. Email system changes")
	fmt.Println()
}

// demonstrateGoodExample shows code that follows SRP
func demonstrateGoodExample() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ GOOD EXAMPLE: Each struct has ONE job")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Create individual components (each with ONE responsibility)
	validator := NewUserValidator()
	repository := NewUserRepository()
	emailService := NewEmailService()

	// Create the service that coordinates all components
	userService := NewUserService(validator, repository, emailService)

	// Test Case 1: Create a valid user
	fmt.Println("📝 Test 1: Creating a valid user...")
	createdUser, err := userService.CreateUser("Bob Johnson", "bob@example.com")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Success! Created: %s\n", createdUser)
	}

	fmt.Println()

	// Test Case 2: Try to create a user with empty name
	fmt.Println("📝 Test 2: Trying to create a user with empty name...")
	_, err = userService.CreateUser("", "invalid@example.com")
	if err != nil {
		fmt.Printf("❌ Error (expected!): %v\n", err)
	}

	fmt.Println()

	// Test Case 3: Try to create a user with invalid email
	fmt.Println("📝 Test 3: Trying to create a user with invalid email...")
	_, err = userService.CreateUser("Charlie", "invalid-email")
	if err != nil {
		fmt.Printf("❌ Error (expected!): %v\n", err)
	}

	fmt.Println()

	// Test Case 4: Retrieve a user from the repository
	fmt.Println("📝 Test 4: Retrieving the user we created...")
	foundUser, err := userService.GetUser(1)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Found: %s\n", foundUser)
	}

	fmt.Println()
}

// printSummary displays the key benefits and takeaways
func printSummary() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📚 SUMMARY: Benefits of SRP")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Each struct has only ONE reason to change:")
	fmt.Println("  • User           → Only if data structure changes")
	fmt.Println("  • UserValidator  → Only if validation rules change")
	fmt.Println("  • UserRepository → Only if database changes")
	fmt.Println("  • EmailService   → Only if email system changes")
	fmt.Println("  • UserService    → Only if business flow changes")

	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║  🎯 KEY TAKEAWAY: One Struct = One Job = One Reason to    ║")
	fmt.Println("║     Change. This makes code easier to test, change, and  ║")
	fmt.Println("║     understand!                                          ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
}
