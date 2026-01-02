package main

import "fmt"

// ============================================================
// DEPENDENCY INVERSION PRINCIPLE (DIP)
// "Depend on abstractions, not concretions"
// ============================================================
//
// WHAT IS DIP? (Explained Simply)
// --------------------------------
// Imagine you have a TV remote. The remote doesn't care which TV brand
// you have - it just needs the TV to understand "power", "volume", etc.
// The remote depends on the CONCEPT of a TV (abstraction), not a specific brand.
//
// IN PROGRAMMING TERMS:
// - High-level modules (like a UserService) should NOT directly depend
//   on low-level modules (like MySQLDatabase)
// - Both should depend on INTERFACES (abstractions)
//
// WHY IS THIS IMPORTANT?
// - Easy to swap implementations (change MySQL to PostgreSQL easily)
// - Easy to test (use fake/mock implementations for testing)
// - Loose coupling = changes in one part don't break others
//
// REAL-WORLD ANALOGY:
// Your laptop charger plugs into a wall socket (interface).
// You don't care if electricity comes from solar, nuclear, or coal.
// The socket is the abstraction that decouples your laptop from the power source.

// ============================================================
// ❌ BAD EXAMPLE: Direct Dependency on Concrete Types
// ============================================================
// This shows what NOT to do - UserServiceBad is tightly coupled
// to MySQLDatabase. If we want to change the database, we must
// modify the UserServiceBad code!

// MySQLDatabaseBad is a concrete (specific) implementation
// "Concrete" means it's a real, specific thing - not an abstraction
type MySQLDatabaseBad struct {
	connectionString string // The address to connect to the database
}

// Save stores data in the MySQL database
func (db *MySQLDatabaseBad) Save(data string) {
	fmt.Printf("MySQL: Saving '%s' to database\n", data)
}

// Get retrieves data from the MySQL database
func (db *MySQLDatabaseBad) Get(id string) string {
	return fmt.Sprintf("MySQL: Data for %s", id)
}

// UserServiceBad directly depends on MySQLDatabaseBad - THIS IS THE PROBLEM!
// Problems with this approach:
// 1. Cannot easily switch to PostgreSQL or MongoDB
// 2. Cannot easily test (need real MySQL running)
// 3. Any change to MySQLDatabaseBad might break UserServiceBad
type UserServiceBad struct {
	db *MySQLDatabaseBad // ❌ Concrete type = tight coupling!
}

// NewUserServiceBad creates the service with a hardcoded MySQL dependency
// Notice how the database is created INSIDE this function - no flexibility!
func NewUserServiceBad() *UserServiceBad {
	return &UserServiceBad{
		db: &MySQLDatabaseBad{connectionString: "mysql://localhost:3306"},
	}
}

// SaveUser saves a user - directly calls MySQL methods
func (service *UserServiceBad) SaveUser(name string) {
	service.db.Save(name) // ❌ Tightly coupled to MySQL
}

// GetUser retrieves a user by ID
func (service *UserServiceBad) GetUser(id string) string {
	return service.db.Get(id) // ❌ Tightly coupled to MySQL
}

// ============================================================
// ✅ GOOD EXAMPLE: Depend on Abstractions (Interfaces)
// ============================================================
// This is the RIGHT way! UserService depends on a Database INTERFACE.
// We can easily swap MySQL for PostgreSQL, MongoDB, or even a fake
// database for testing - all without changing UserService!

// Database is an INTERFACE - it defines WHAT a database should do,
// not HOW it does it. Any struct that has these methods is a "Database".
type Database interface {
	Save(data string) error        // Store data, return error if failed
	Get(id string) (string, error) // Retrieve data by ID
	Delete(id string) error        // Remove data by ID
}

// -------------------- MySQL Implementation --------------------

// MySQLDB implements the Database interface using MySQL
type MySQLDB struct {
	connectionString string // Example: "mysql://localhost:3306"
}

// NewMySQLDB creates a new MySQL database connection
func NewMySQLDB(connectionString string) *MySQLDB {
	return &MySQLDB{connectionString: connectionString}
}

// Save stores data in MySQL (implements Database interface)
func (db *MySQLDB) Save(data string) error {
	fmt.Printf("MySQL: Saving '%s'\n", data)
	return nil // nil means no error - success!
}

// Get retrieves data from MySQL (implements Database interface)
func (db *MySQLDB) Get(id string) (string, error) {
	return fmt.Sprintf("MySQL: Data for %s", id), nil
}

// Delete removes data from MySQL (implements Database interface)
func (db *MySQLDB) Delete(id string) error {
	fmt.Printf("MySQL: Deleting %s\n", id)
	return nil
}

// -------------------- PostgreSQL Implementation --------------------

// PostgreSQLDB implements the Database interface using PostgreSQL
type PostgreSQLDB struct {
	connectionString string // Example: "postgres://localhost:5432"
}

// NewPostgreSQLDB creates a new PostgreSQL database connection
func NewPostgreSQLDB(connectionString string) *PostgreSQLDB {
	return &PostgreSQLDB{connectionString: connectionString}
}

// Save stores data in PostgreSQL (implements Database interface)
func (db *PostgreSQLDB) Save(data string) error {
	fmt.Printf("PostgreSQL: Saving '%s'\n", data)
	return nil
}

// Get retrieves data from PostgreSQL (implements Database interface)
func (db *PostgreSQLDB) Get(id string) (string, error) {
	return fmt.Sprintf("PostgreSQL: Data for %s", id), nil
}

// Delete removes data from PostgreSQL (implements Database interface)
func (db *PostgreSQLDB) Delete(id string) error {
	fmt.Printf("PostgreSQL: Deleting %s\n", id)
	return nil
}

// -------------------- MongoDB Implementation --------------------

// MongoDatabase implements the Database interface using MongoDB
type MongoDatabase struct {
	connectionString string // Example: "mongodb://localhost:27017"
}

// NewMongoDatabase creates a new MongoDB database connection
func NewMongoDatabase(connectionString string) *MongoDatabase {
	return &MongoDatabase{connectionString: connectionString}
}

// Save stores data in MongoDB (implements Database interface)
func (db *MongoDatabase) Save(data string) error {
	fmt.Printf("MongoDB: Saving '%s'\n", data)
	return nil
}

// Get retrieves data from MongoDB (implements Database interface)
func (db *MongoDatabase) Get(id string) (string, error) {
	return fmt.Sprintf("MongoDB: Data for %s", id), nil
}

// Delete removes data from MongoDB (implements Database interface)
func (db *MongoDatabase) Delete(id string) error {
	fmt.Printf("MongoDB: Deleting %s\n", id)
	return nil
}

// -------------------- InMemory Implementation (For Testing!) --------------------

// InMemoryDatabase stores data in memory - perfect for unit tests!
// No actual database needed - fast and isolated tests.
type InMemoryDatabase struct {
	storedData map[string]string // Stores key-value pairs in memory
}

// NewInMemoryDatabase creates a new in-memory database
func NewInMemoryDatabase() *InMemoryDatabase {
	return &InMemoryDatabase{storedData: make(map[string]string)}
}

// Save stores data in memory (implements Database interface)
func (db *InMemoryDatabase) Save(data string) error {
	db.storedData["latest"] = data
	fmt.Printf("InMemory: Saving '%s'\n", data)
	return nil
}

// Get retrieves data from memory (implements Database interface)
func (db *InMemoryDatabase) Get(id string) (string, error) {
	if value, exists := db.storedData[id]; exists {
		return value, nil
	}
	return "", fmt.Errorf("not found: %s", id)
}

// Delete removes data from memory (implements Database interface)
func (db *InMemoryDatabase) Delete(id string) error {
	delete(db.storedData, id)
	fmt.Printf("InMemory: Deleting %s\n", id)
	return nil
}

// -------------------- DIPUserService (Uses the Database Interface) --------------------

// DIPUserService depends on Database INTERFACE, not a concrete type
// This is the KEY to DIP - we can swap ANY database implementation!
// (Named DIPUserService to avoid conflict with UserService in 01_srp.go)
type DIPUserService struct {
	db Database // ✅ Interface type = loose coupling!
}

// NewDIPUserService creates a DIPUserService with dependency injection
// "Dependency Injection" means we PASS the dependency from outside
// instead of creating it inside this function
func NewDIPUserService(db Database) *DIPUserService {
	return &DIPUserService{db: db}
}

// SaveUser saves a user to whatever database was injected
func (service *DIPUserService) SaveUser(name string) error {
	return service.db.Save(name)
}

// GetUser retrieves a user from whatever database was injected
func (service *DIPUserService) GetUser(id string) (string, error) {
	return service.db.Get(id)
}

// ============================================================
// COMPLETE EXAMPLE: Notification Service
// ============================================================
// This example shows how DIP makes it easy to send notifications
// via different channels (Email, SMS, Push) using ONE NotificationService!

// MessageSender is an interface for sending messages
// Any struct that has a Send method can be a MessageSender
type MessageSender interface {
	Send(recipient string, message string) error
}

// -------------------- Email Sender --------------------

// EmailSender sends messages via email
type EmailSender struct {
	smtpHost string // Email server address (e.g., "smtp.gmail.com")
	smtpPort int    // Email server port (e.g., 587)
}

// NewEmailSender creates a new email sender
func NewEmailSender(host string, port int) *EmailSender {
	return &EmailSender{smtpHost: host, smtpPort: port}
}

// Send sends an email message (implements MessageSender interface)
func (sender *EmailSender) Send(recipient string, message string) error {
	fmt.Printf("📧 Email to %s: %s\n", recipient, message)
	return nil
}

// -------------------- SMS Sender --------------------

// SMSSender sends messages via SMS text
type SMSSender struct {
	apiKey string // API key for SMS service (e.g., Twilio)
}

// NewSMSSender creates a new SMS sender
func NewSMSSender(apiKey string) *SMSSender {
	return &SMSSender{apiKey: apiKey}
}

// Send sends an SMS message (implements MessageSender interface)
func (sender *SMSSender) Send(recipient string, message string) error {
	fmt.Printf("📱 SMS to %s: %s\n", recipient, message)
	return nil
}

// -------------------- Push Notification Sender --------------------

// PushNotificationSender sends push notifications to mobile devices
type PushNotificationSender struct {
	fcmKey string // Firebase Cloud Messaging key
}

// NewPushNotificationSender creates a new push notification sender
func NewPushNotificationSender(fcmKey string) *PushNotificationSender {
	return &PushNotificationSender{fcmKey: fcmKey}
}

// Send sends a push notification (implements MessageSender interface)
func (sender *PushNotificationSender) Send(recipient string, message string) error {
	fmt.Printf("🔔 Push to %s: %s\n", recipient, message)
	return nil
}

// -------------------- Mock Sender (For Testing!) --------------------

// MockSender is a fake sender for testing - records messages instead of sending
type MockSender struct {
	SentMessages []string // Stores all "sent" messages for verification
}

// Send records the message instead of actually sending it
func (sender *MockSender) Send(recipient string, message string) error {
	sender.SentMessages = append(sender.SentMessages, fmt.Sprintf("%s: %s", recipient, message))
	return nil
}

// -------------------- DIPNotificationService --------------------

// DIPNotificationService can send notifications using ANY MessageSender
// Thanks to DIP, we can easily switch between Email, SMS, or Push!
// (Named DIPNotificationService to avoid conflicts with other files)
type DIPNotificationService struct {
	sender MessageSender // The abstraction - could be any sender type!
}

// NewDIPNotificationService creates a notification service with the given sender
func NewDIPNotificationService(sender MessageSender) *DIPNotificationService {
	return &DIPNotificationService{sender: sender}
}

// NotifyUser sends a notification to a user
func (service *DIPNotificationService) NotifyUser(userID string, message string) error {
	return service.sender.Send(userID, message)
}

// ============================================================
// ADVANCED: Multiple Dependencies with Constructor Injection
// ============================================================
// Real-world services often need MULTIPLE dependencies.
// DIP makes it easy to inject all of them!

// Logger interface - defines how to log messages
type Logger interface {
	Log(message string)
}

// -------------------- Console Logger --------------------

// ConsoleLogger prints log messages to the terminal
type ConsoleLogger struct{}

// Log prints a message to the console
func (logger *ConsoleLogger) Log(message string) {
	fmt.Printf("[LOG] %s\n", message)
}

// -------------------- File Logger --------------------

// FileLogger writes log messages to a file
type FileLogger struct {
	filePath string // Path to the log file
}

// NewFileLogger creates a new file logger
func NewFileLogger(filePath string) *FileLogger {
	return &FileLogger{filePath: filePath}
}

// Log writes a message to the log file (simulated here with print)
func (logger *FileLogger) Log(message string) {
	fmt.Printf("[FILE: %s] %s\n", logger.filePath, message)
}

// -------------------- Order Service (Uses Multiple Dependencies) --------------------

// OrderService demonstrates a service with multiple injected dependencies
// It needs: a database, a message sender, and a logger
type OrderService struct {
	db       Database      // For storing orders
	notifier MessageSender // For notifying users
	logger   Logger        // For logging events
}

// NewOrderService creates an order service with all its dependencies
// This is "Constructor Injection" - all dependencies passed at creation time
func NewOrderService(db Database, notifier MessageSender, logger Logger) *OrderService {
	return &OrderService{
		db:       db,
		notifier: notifier,
		logger:   logger,
	}
}

// PlaceOrder creates a new order, saves it, and notifies the user
func (service *OrderService) PlaceOrder(userID string, itemName string) error {
	// Step 1: Log the start of the operation
	service.logger.Log(fmt.Sprintf("Placing order for %s: %s", userID, itemName))

	// Step 2: Save the order to the database
	orderData := fmt.Sprintf("Order: %s - %s", userID, itemName)
	if err := service.db.Save(orderData); err != nil {
		return err
	}

	// Step 3: Notify the user about their order
	notificationMessage := fmt.Sprintf("Your order for %s is placed!", itemName)
	if err := service.notifier.Send(userID, notificationMessage); err != nil {
		service.logger.Log(fmt.Sprintf("Failed to notify user: %v", err))
		// We don't return error here - notification failure shouldn't fail the order
	}

	// Step 4: Log success
	service.logger.Log("Order placed successfully")
	return nil
}

// ============================================================
// KEY INTERVIEW POINTS
// ============================================================
//
// Q: What is Dependency Injection?
// A: Passing dependencies (usually as interfaces) to a component
//    instead of the component creating them itself.
//    Three ways: Constructor injection, Setter injection, Method injection.
//
// Q: How does DIP help testing?
// A: You can inject mock implementations during tests!
//    No need for actual database, email server, etc.
//
// Q: What's the difference between DIP and Dependency Injection?
// A: DIP is the PRINCIPLE (depend on abstractions)
//    DI is a TECHNIQUE to achieve DIP
//
// Q: When should you create an interface?
// A: When you need to:
//    - Swap implementations
//    - Test with mocks
//    - Decouple modules
//
// ❌ Common Mistakes:
// 1. Creating interfaces for everything (premature abstraction)
// 2. Huge interfaces (violates ISP)
// 3. Passing concrete types when you need flexibility
// 4. Not using constructor injection (hard to test)

// handleError is a helper function to check and print errors
// In real applications, you'd handle errors more gracefully
func handleError(err error, context string) {
	if err != nil {
		fmt.Printf("Error %s: %v\n", context, err)
	}
}

func main() {
	fmt.Println("=== Dependency Inversion Principle Demo ===\n")

	// ============================================================
	// PART 1: Demonstrating the BAD approach (tightly coupled)
	// ============================================================
	fmt.Println("--- ❌ BAD Example: Tightly Coupled Service ---")
	badService := NewUserServiceBad()
	badService.SaveUser("Alice")
	userData := badService.GetUser("user1")
	fmt.Printf("Retrieved: %s\n", userData)
	fmt.Println("Problem: Cannot switch database without modifying UserServiceBad!\n")

	// ============================================================
	// PART 2: Demonstrating the GOOD approach (loosely coupled)
	// ============================================================
	fmt.Println("--- ✅ GOOD Example: Loosely Coupled with Interfaces ---")
	fmt.Println("Same UserService works with ANY database implementation!\n")

	// Example 1: Using MySQL
	fmt.Println("1. Using MySQL Database:")
	mysqlDB := NewMySQLDB("mysql://localhost:3306")
	mysqlUserService := NewDIPUserService(mysqlDB)
	err := mysqlUserService.SaveUser("John")
	handleError(err, "saving user to MySQL")

	// Example 2: Using PostgreSQL - no code change in DIPUserService!
	fmt.Println("\n2. Using PostgreSQL Database:")
	postgresDB := NewPostgreSQLDB("postgres://localhost:5432")
	postgresUserService := NewDIPUserService(postgresDB)
	err = postgresUserService.SaveUser("Jane")
	handleError(err, "saving user to PostgreSQL")

	// Example 3: Using MongoDB
	fmt.Println("\n3. Using MongoDB Database:")
	mongoDB := NewMongoDatabase("mongodb://localhost:27017")
	mongoUserService := NewDIPUserService(mongoDB)
	err = mongoUserService.SaveUser("Bob")
	handleError(err, "saving user to MongoDB")

	// Example 4: Using InMemory (for testing)
	fmt.Println("\n4. Using InMemory Database (for testing):")
	inMemoryDB := NewInMemoryDatabase()
	inMemoryUserService := NewDIPUserService(inMemoryDB)
	err = inMemoryUserService.SaveUser("TestUser")
	handleError(err, "saving user to InMemory")

	// ============================================================
	// PART 3: Notification Service Example
	// ============================================================
	fmt.Println("\n--- Notification Service Example ---")
	fmt.Println("Same NotificationService works with ANY message sender!\n")

	// Email notifications
	fmt.Println("1. Sending via Email:")
	emailSender := NewEmailSender("smtp.example.com", 587)
	emailNotificationService := NewDIPNotificationService(emailSender)
	err = emailNotificationService.NotifyUser("user@example.com", "Welcome!")
	handleError(err, "sending email")

	// SMS notifications
	fmt.Println("\n2. Sending via SMS:")
	smsSender := NewSMSSender("api-key-123")
	smsNotificationService := NewDIPNotificationService(smsSender)
	err = smsNotificationService.NotifyUser("+1234567890", "Your OTP is 123456")
	handleError(err, "sending SMS")

	// Push notifications
	fmt.Println("\n3. Sending via Push Notification:")
	pushSender := NewPushNotificationSender("fcm-key")
	pushNotificationService := NewDIPNotificationService(pushSender)
	err = pushNotificationService.NotifyUser("device-token-abc", "New message!")
	handleError(err, "sending push notification")

	// ============================================================
	// PART 4: Order Service with Multiple Dependencies
	// ============================================================
	fmt.Println("\n--- Order Service (Multiple Dependencies) ---")
	fmt.Println("OrderService needs: Database + MessageSender + Logger\n")

	// Create order service with console logger
	fmt.Println("1. Using Console Logger:")
	consoleLogger := &ConsoleLogger{}
	orderServiceWithConsole := NewOrderService(
		NewMySQLDB("mysql://localhost:3306"),
		NewEmailSender("smtp.example.com", 587),
		consoleLogger,
	)
	err = orderServiceWithConsole.PlaceOrder("user123", "iPhone 15")
	handleError(err, "placing order with console logger")

	// Create order service with file logger
	fmt.Println("\n2. Using File Logger:")
	fileLogger := NewFileLogger("/var/log/orders.log")
	orderServiceWithFile := NewOrderService(
		NewPostgreSQLDB("postgres://localhost:5432"),
		NewSMSSender("sms-api-key"),
		fileLogger,
	)
	err = orderServiceWithFile.PlaceOrder("user456", "MacBook Pro")
	handleError(err, "placing order with file logger")

	// ============================================================
	// PART 5: Testing with Mock Dependencies
	// ============================================================
	fmt.Println("\n--- Testing with Mock Dependencies ---")
	fmt.Println("Using fake implementations for unit testing!\n")

	// Create mock dependencies
	mockDB := NewInMemoryDatabase()
	mockSender := &MockSender{}
	mockLogger := &ConsoleLogger{}

	// Create order service with mock dependencies
	testOrderService := NewOrderService(mockDB, mockSender, mockLogger)
	err = testOrderService.PlaceOrder("testuser", "Test Item")
	handleError(err, "placing test order")

	// Verify mock sender captured the message
	fmt.Println("\nMessages captured by MockSender during test:")
	for i, msg := range mockSender.SentMessages {
		fmt.Printf("  %d. %s\n", i+1, msg)
	}

	// ============================================================
	// SUMMARY
	// ============================================================
	fmt.Println("\n=== Summary ===")
	fmt.Println("✅ DIP allows us to:")
	fmt.Println("   1. Swap implementations easily (MySQL → PostgreSQL)")
	fmt.Println("   2. Test with mock/fake dependencies")
	fmt.Println("   3. Build loosely coupled, maintainable systems")
	fmt.Println("   4. Follow 'Program to an interface, not an implementation'")
}
