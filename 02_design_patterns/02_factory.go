package main

import "fmt"

// ============================================================
// FACTORY PATTERN
// ============================================================
// Definition: "Define an interface for creating objects, let
// subclasses decide which class to instantiate"
//
// Real-World Analogy:
// Think of a car dealership - you tell them "I want an SUV",
// and they handle all the complex work of getting you one.
// You don't need to know HOW the car is manufactured.
//
// WHEN TO USE:
// - You don't know ahead of time which type to create
// - Object creation logic is complex
// - You want to centralize object creation
// - You want to hide creation details from client
//
// TYPES OF FACTORY PATTERNS:
// 1. Simple Factory - One function/class that creates objects
// 2. Factory Method - Interface for creating, subclasses decide
// 3. Abstract Factory - Creates families of related objects
//
// INTERVIEW TIP:
// Factory is used in almost EVERY LLD problem!
// Payment processors, notification services, vehicle types, etc.

// ============================================================
// EXAMPLE 1: Simple Factory - Vehicle Creation
// ============================================================
// Simple Factory: A single factory class with one method that
// creates different types of objects based on input parameters.

// ------------------------------------------------------------
// Step 1: Define the interface (what ALL vehicles must do)
// ------------------------------------------------------------

// Vehicle is the interface that all vehicle types must implement
type Vehicle interface {
	Drive()          // How the vehicle moves
	GetType() string // Returns the vehicle type name
	GetWheels() int  // Returns number of wheels
}

// ------------------------------------------------------------
// Step 2: Create concrete implementations (Car, Motorcycle, Truck)
// ------------------------------------------------------------

// Car represents a 4-wheeled passenger vehicle
type Car struct {
	brand string // The manufacturer/brand of the car
}

// Drive makes the car move (implements Vehicle interface)
func (car *Car) Drive() {
	fmt.Printf("🚗 %s car is driving on the road\n", car.brand)
}

// GetType returns "Car" (implements Vehicle interface)
func (car *Car) GetType() string {
	return "Car"
}

// GetWheels returns 4 for cars (implements Vehicle interface)
func (car *Car) GetWheels() int {
	return 4
}

// Motorcycle represents a 2-wheeled motor vehicle
type Motorcycle struct {
	brand string // The manufacturer/brand of the motorcycle
}

// Drive makes the motorcycle move (implements Vehicle interface)
func (motorcycle *Motorcycle) Drive() {
	fmt.Printf("🏍️ %s motorcycle is zooming on the road\n", motorcycle.brand)
}

// GetType returns "Motorcycle" (implements Vehicle interface)
func (motorcycle *Motorcycle) GetType() string {
	return "Motorcycle"
}

// GetWheels returns 2 for motorcycles (implements Vehicle interface)
func (motorcycle *Motorcycle) GetWheels() int {
	return 2
}

// Truck represents a large cargo vehicle
type Truck struct {
	brand          string // The manufacturer/brand of the truck
	capacityInTons int    // How much cargo it can carry
}

// Drive makes the truck move (implements Vehicle interface)
func (truck *Truck) Drive() {
	fmt.Printf("🚛 %s truck (capacity: %d tons) is moving\n", truck.brand, truck.capacityInTons)
}

// GetType returns "Truck" (implements Vehicle interface)
func (truck *Truck) GetType() string {
	return "Truck"
}

// GetWheels returns 6 for trucks (implements Vehicle interface)
func (truck *Truck) GetWheels() int {
	return 6
}

// ------------------------------------------------------------
// Step 3: Define vehicle types as constants (type safety)
// ------------------------------------------------------------

// VehicleType represents the type of vehicle to create
type VehicleType string

const (
	VehicleTypeCar        VehicleType = "car"
	VehicleTypeMotorcycle VehicleType = "motorcycle"
	VehicleTypeTruck      VehicleType = "truck"
)

// ------------------------------------------------------------
// Step 4: Create the Factory
// ------------------------------------------------------------

// VehicleFactory creates different types of vehicles
// Benefits:
// - Client code doesn't need to know HOW vehicles are created
// - Easy to add new vehicle types without changing client code
// - Centralizes vehicle creation logic
type VehicleFactory struct{}

// CreateVehicle creates a vehicle based on the type requested
// Parameters:
//   - vehicleType: what kind of vehicle to create
//   - brand: the manufacturer name
//
// Returns:
//   - Vehicle interface (could be Car, Motorcycle, or Truck)
//   - error if the vehicle type is unknown
func (factory *VehicleFactory) CreateVehicle(vehicleType VehicleType, brand string) (Vehicle, error) {
	switch vehicleType {
	case VehicleTypeCar:
		return &Car{brand: brand}, nil

	case VehicleTypeMotorcycle:
		return &Motorcycle{brand: brand}, nil

	case VehicleTypeTruck:
		return &Truck{brand: brand, capacityInTons: 10}, nil

	default:
		return nil, fmt.Errorf("unknown vehicle type: %s", vehicleType)
	}
}

// ============================================================
// EXAMPLE 2: Payment Processor Factory
// ============================================================
// This is a VERY common interview question!
// Shows how to handle objects that need configuration data.

// ------------------------------------------------------------
// Step 1: Define payment method types
// ------------------------------------------------------------

// PaymentMethod represents different payment options
type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodPayPal     PaymentMethod = "paypal"
	PaymentMethodUPI        PaymentMethod = "upi"
	PaymentMethodCrypto     PaymentMethod = "crypto"
)

// ------------------------------------------------------------
// Step 2: Define the PaymentProcessor interface
// ------------------------------------------------------------

// PaymentProcessor defines what all payment processors must do
type PaymentProcessor interface {
	ProcessPayment(amount float64) error // Charge the customer
	Refund(amount float64) error         // Return money to customer
	GetName() string                     // Get the processor name
}

// ------------------------------------------------------------
// Step 3: Implement each payment processor
// ------------------------------------------------------------

// CreditCardProcessor handles credit card payments
type CreditCardProcessor struct {
	cardNumber string // The customer's card number
}

// NewCreditCardProcessor creates a new credit card processor
func NewCreditCardProcessor(cardNumber string) *CreditCardProcessor {
	return &CreditCardProcessor{cardNumber: cardNumber}
}

// ProcessPayment charges the credit card
func (processor *CreditCardProcessor) ProcessPayment(amount float64) error {
	// Show only last 4 digits for security
	lastFourDigits := processor.cardNumber[len(processor.cardNumber)-4:]
	fmt.Printf("💳 Processing $%.2f via Credit Card ending in %s\n", amount, lastFourDigits)
	return nil
}

// Refund returns money to the credit card
func (processor *CreditCardProcessor) Refund(amount float64) error {
	fmt.Printf("💳 Refunding $%.2f to Credit Card\n", amount)
	return nil
}

// GetName returns the processor name
func (processor *CreditCardProcessor) GetName() string {
	return "Credit Card"
}

// PayPalProcessor handles PayPal payments
type PayPalProcessor struct {
	email string // Customer's PayPal email
}

// NewPayPalProcessor creates a new PayPal processor
func NewPayPalProcessor(email string) *PayPalProcessor {
	return &PayPalProcessor{email: email}
}

// ProcessPayment processes payment via PayPal
func (processor *PayPalProcessor) ProcessPayment(amount float64) error {
	fmt.Printf("🅿️ Processing $%.2f via PayPal (%s)\n", amount, processor.email)
	return nil
}

// Refund returns money via PayPal
func (processor *PayPalProcessor) Refund(amount float64) error {
	fmt.Printf("🅿️ Refunding $%.2f to PayPal\n", amount)
	return nil
}

// GetName returns the processor name
func (processor *PayPalProcessor) GetName() string {
	return "PayPal"
}

// UPIProcessor handles UPI payments (popular in India)
type UPIProcessor struct {
	upiID string // Customer's UPI ID (e.g., user@paytm)
}

// NewUPIProcessor creates a new UPI processor
func NewUPIProcessor(upiID string) *UPIProcessor {
	return &UPIProcessor{upiID: upiID}
}

// ProcessPayment processes payment via UPI
func (processor *UPIProcessor) ProcessPayment(amount float64) error {
	fmt.Printf("📱 Processing $%.2f via UPI (%s)\n", amount, processor.upiID)
	return nil
}

// Refund returns money via UPI
func (processor *UPIProcessor) Refund(amount float64) error {
	fmt.Printf("📱 Refunding $%.2f to UPI\n", amount)
	return nil
}

// GetName returns the processor name
func (processor *UPIProcessor) GetName() string {
	return "UPI"
}

// CryptoProcessor handles cryptocurrency payments
type CryptoProcessor struct {
	walletAddress string // Customer's crypto wallet address
}

// NewCryptoProcessor creates a new crypto processor
func NewCryptoProcessor(walletAddress string) *CryptoProcessor {
	return &CryptoProcessor{walletAddress: walletAddress}
}

// ProcessPayment processes payment via cryptocurrency
func (processor *CryptoProcessor) ProcessPayment(amount float64) error {
	// Show only first 8 characters of wallet for display
	shortAddress := processor.walletAddress[:8]
	fmt.Printf("₿ Processing $%.2f via Crypto (%s...)\n", amount, shortAddress)
	return nil
}

// Refund returns money via cryptocurrency
func (processor *CryptoProcessor) Refund(amount float64) error {
	fmt.Printf("₿ Refunding $%.2f to Crypto wallet\n", amount)
	return nil
}

// GetName returns the processor name
func (processor *CryptoProcessor) GetName() string {
	return "Cryptocurrency"
}

// ------------------------------------------------------------
// Step 4: Configuration struct (holds data needed to create processors)
// ------------------------------------------------------------

// PaymentConfig holds all possible configuration options
// Different payment methods will use different fields
type PaymentConfig struct {
	CardNumber    string // For credit card payments
	Email         string // For PayPal payments
	UPIID         string // For UPI payments
	WalletAddress string // For crypto payments
}

// ------------------------------------------------------------
// Step 5: Create the Payment Factory
// ------------------------------------------------------------

// PaymentProcessorFactory creates payment processors
type PaymentProcessorFactory struct{}

// CreateProcessor creates a payment processor based on the method
// Parameters:
//   - method: which payment method to use
//   - config: configuration data for the processor
//
// Returns:
//   - PaymentProcessor interface
//   - error if validation fails or method is unknown
func (factory *PaymentProcessorFactory) CreateProcessor(
	method PaymentMethod,
	config PaymentConfig,
) (PaymentProcessor, error) {

	switch method {
	case PaymentMethodCreditCard:
		// Validate required field
		if config.CardNumber == "" {
			return nil, fmt.Errorf("card number is required for credit card payment")
		}
		return NewCreditCardProcessor(config.CardNumber), nil

	case PaymentMethodPayPal:
		// Validate required field
		if config.Email == "" {
			return nil, fmt.Errorf("email is required for PayPal payment")
		}
		return NewPayPalProcessor(config.Email), nil

	case PaymentMethodUPI:
		// Validate required field
		if config.UPIID == "" {
			return nil, fmt.Errorf("UPI ID is required for UPI payment")
		}
		return NewUPIProcessor(config.UPIID), nil

	case PaymentMethodCrypto:
		// Validate required field
		if config.WalletAddress == "" {
			return nil, fmt.Errorf("wallet address is required for crypto payment")
		}
		return NewCryptoProcessor(config.WalletAddress), nil

	default:
		return nil, fmt.Errorf("unsupported payment method: %s", method)
	}
}

// ============================================================
// EXAMPLE 3: Factory with Registration (Plugin Pattern)
// ============================================================
// This advanced pattern allows adding new types WITHOUT modifying
// the factory code! Great for extensible systems.
//
// How it works:
// 1. Factory maintains a map of "type name" -> "creator function"
// 2. New types register themselves with the factory
// 3. Factory looks up the creator and calls it

// ------------------------------------------------------------
// Step 1: Define the interface
// ------------------------------------------------------------

// DocumentParser parses different document formats
type DocumentParser interface {
	Parse(content string) (map[string]interface{}, error)
	GetFormat() string
}

// ------------------------------------------------------------
// Step 2: Implement parsers for different formats
// ------------------------------------------------------------

// JSONParser parses JSON documents
type JSONParser struct{}

// Parse parses JSON content (simplified for demo)
func (parser *JSONParser) Parse(content string) (map[string]interface{}, error) {
	fmt.Println("  📄 Parsing JSON document...")
	// In real code, you'd use encoding/json
	return map[string]interface{}{"type": "json", "content": content}, nil
}

// GetFormat returns the format this parser handles
func (parser *JSONParser) GetFormat() string {
	return "json"
}

// XMLParser parses XML documents
type XMLParser struct{}

// Parse parses XML content (simplified for demo)
func (parser *XMLParser) Parse(content string) (map[string]interface{}, error) {
	fmt.Println("  📄 Parsing XML document...")
	// In real code, you'd use encoding/xml
	return map[string]interface{}{"type": "xml", "content": content}, nil
}

// GetFormat returns the format this parser handles
func (parser *XMLParser) GetFormat() string {
	return "xml"
}

// YAMLParser parses YAML documents
type YAMLParser struct{}

// Parse parses YAML content (simplified for demo)
func (parser *YAMLParser) Parse(content string) (map[string]interface{}, error) {
	fmt.Println("  📄 Parsing YAML document...")
	// In real code, you'd use gopkg.in/yaml.v3
	return map[string]interface{}{"type": "yaml", "content": content}, nil
}

// GetFormat returns the format this parser handles
func (parser *YAMLParser) GetFormat() string {
	return "yaml"
}

// ------------------------------------------------------------
// Step 3: Create the Registrable Factory
// ------------------------------------------------------------

// ParserCreatorFunc is a function that creates a DocumentParser
type ParserCreatorFunc func() DocumentParser

// ParserFactory creates document parsers using registered creators
type ParserFactory struct {
	// Map from format name to creator function
	registeredParsers map[string]ParserCreatorFunc
}

// NewParserFactory creates a new parser factory
func NewParserFactory() *ParserFactory {
	return &ParserFactory{
		registeredParsers: make(map[string]ParserCreatorFunc),
	}
}

// Register adds a new parser type to the factory
// This allows adding new formats without modifying factory code!
func (factory *ParserFactory) Register(format string, creator ParserCreatorFunc) {
	factory.registeredParsers[format] = creator
	fmt.Printf("  ✅ Registered parser for format: %s\n", format)
}

// Create creates a parser for the given format
func (factory *ParserFactory) Create(format string) (DocumentParser, error) {
	creator, exists := factory.registeredParsers[format]
	if !exists {
		return nil, fmt.Errorf("no parser registered for format: %s", format)
	}
	return creator(), nil
}

// ListRegisteredFormats returns all formats that can be parsed
func (factory *ParserFactory) ListRegisteredFormats() []string {
	formats := make([]string, 0, len(factory.registeredParsers))
	for format := range factory.registeredParsers {
		formats = append(formats, format)
	}
	return formats
}

// ============================================================
// EXAMPLE 4: Notification Factory (Common Interview Question)
// ============================================================
// Shows a simple factory function (not a struct) - also valid!

// NotificationType represents different notification channels
type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
	NotificationTypePush  NotificationType = "push"
)

// Notifier sends notifications through various channels
type Notifier interface {
	Send(recipient, message string) error
	GetType() string
}

// EmailNotifier sends notifications via email
type EmailNotifier struct {
	smtpHost string // Email server hostname
}

// Send sends an email notification
func (notifier *EmailNotifier) Send(recipient, message string) error {
	fmt.Printf("  📧 Email to %s: %s\n", recipient, message)
	return nil
}

// GetType returns "email"
func (notifier *EmailNotifier) GetType() string {
	return "email"
}

// SMSNotifier sends notifications via SMS
type SMSNotifier struct {
	apiKey string // SMS service API key
}

// Send sends an SMS notification
func (notifier *SMSNotifier) Send(recipient, message string) error {
	fmt.Printf("  📱 SMS to %s: %s\n", recipient, message)
	return nil
}

// GetType returns "sms"
func (notifier *SMSNotifier) GetType() string {
	return "sms"
}

// PushNotifier sends push notifications
type PushNotifier struct{}

// Send sends a push notification
func (notifier *PushNotifier) Send(recipient, message string) error {
	fmt.Printf("  🔔 Push to %s: %s\n", recipient, message)
	return nil
}

// GetType returns "push"
func (notifier *PushNotifier) GetType() string {
	return "push"
}

// CreateNotifier is a simple factory FUNCTION (not a struct)
// This is a valid alternative when you don't need factory state
func CreateNotifier(notificationType NotificationType) (Notifier, error) {
	switch notificationType {
	case NotificationTypeEmail:
		return &EmailNotifier{smtpHost: "smtp.example.com"}, nil

	case NotificationTypeSMS:
		return &SMSNotifier{apiKey: "sms-api-key-12345"}, nil

	case NotificationTypePush:
		return &PushNotifier{}, nil

	default:
		return nil, fmt.Errorf("unknown notification type: %s", notificationType)
	}
}

// ============================================================
// KEY INTERVIEW QUESTIONS & ANSWERS
// ============================================================
//
// Q1: When to use Factory vs direct construction (new/struct literal)?
// A1: Use Factory when:
//     - Creation logic is complex (validation, setup)
//     - Type is determined at runtime (user input, config)
//     - You want to hide creation details from client
//     - You need to return interface instead of concrete type
//
// Q2: How does Factory help with testing?
// A2: - You can create a mock factory that returns mock objects
//     - Easy to inject different implementations for testing
//     - Centralizes object creation for easier mocking
//
// Q3: What's the difference between Factory types?
// A3: - Simple Factory: One method creates different types
//     - Factory Method: Subclasses override creation method
//     - Abstract Factory: Creates families of related objects
//
// Q4: How to add new types without modifying factory?
// A4: Use the Registration Pattern (Example 3):
//     - Factory maintains a registry of creators
//     - New types register themselves at startup
//     - Follows Open/Closed Principle (open for extension)
//
// COMMON MISTAKES TO AVOID:
// ❌ Creating factory for simple objects (over-engineering)
// ❌ Huge switch statements (consider registration pattern)
// ❌ Returning concrete types instead of interfaces
// ❌ Not handling invalid input gracefully
// ❌ Forgetting to validate required configuration

// ============================================================
// MAIN FUNCTION - Demonstrates all Factory examples
// ============================================================

func main() {
	fmt.Println("============================================================")
	fmt.Println("           FACTORY PATTERN DEMONSTRATION")
	fmt.Println("============================================================")

	// ---------------------------------------------------------
	// Demo 1: Vehicle Factory (Simple Factory)
	// ---------------------------------------------------------
	fmt.Println("\n--- Demo 1: Vehicle Factory (Simple Factory) ---")
	fmt.Println("Creating different vehicles using the same factory...\n")

	vehicleFactory := &VehicleFactory{}

	// Create a car
	car, err := vehicleFactory.CreateVehicle(VehicleTypeCar, "Tesla")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		car.Drive()
		fmt.Printf("  Type: %s, Wheels: %d\n", car.GetType(), car.GetWheels())
	}

	// Create a motorcycle
	motorcycle, err := vehicleFactory.CreateVehicle(VehicleTypeMotorcycle, "Harley-Davidson")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		motorcycle.Drive()
		fmt.Printf("  Type: %s, Wheels: %d\n", motorcycle.GetType(), motorcycle.GetWheels())
	}

	// Create a truck
	truck, err := vehicleFactory.CreateVehicle(VehicleTypeTruck, "Volvo")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		truck.Drive()
		fmt.Printf("  Type: %s, Wheels: %d\n", truck.GetType(), truck.GetWheels())
	}

	// Try to create an invalid vehicle type
	fmt.Println("\nTrying to create invalid vehicle type:")
	_, err = vehicleFactory.CreateVehicle("spaceship", "SpaceX")
	if err != nil {
		fmt.Printf("  ⚠️ Error (expected): %v\n", err)
	}

	// ---------------------------------------------------------
	// Demo 2: Payment Processor Factory
	// ---------------------------------------------------------
	fmt.Println("\n--- Demo 2: Payment Processor Factory ---")
	fmt.Println("Processing payments with different methods...\n")

	paymentFactory := &PaymentProcessorFactory{}

	// Credit Card payment
	creditCardProcessor, err := paymentFactory.CreateProcessor(
		PaymentMethodCreditCard,
		PaymentConfig{CardNumber: "4111111111111234"},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		creditCardProcessor.ProcessPayment(99.99)
	}

	// PayPal payment
	paypalProcessor, err := paymentFactory.CreateProcessor(
		PaymentMethodPayPal,
		PaymentConfig{Email: "customer@example.com"},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		paypalProcessor.ProcessPayment(49.99)
	}

	// UPI payment
	upiProcessor, err := paymentFactory.CreateProcessor(
		PaymentMethodUPI,
		PaymentConfig{UPIID: "user@paytm"},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		upiProcessor.ProcessPayment(29.99)
	}

	// Crypto payment
	cryptoProcessor, err := paymentFactory.CreateProcessor(
		PaymentMethodCrypto,
		PaymentConfig{WalletAddress: "0x1234567890abcdef1234"},
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		cryptoProcessor.ProcessPayment(199.99)
	}

	// Try missing configuration
	fmt.Println("\nTrying to create processor with missing config:")
	_, err = paymentFactory.CreateProcessor(
		PaymentMethodCreditCard,
		PaymentConfig{}, // Empty - no card number!
	)
	if err != nil {
		fmt.Printf("  ⚠️ Error (expected): %v\n", err)
	}

	// ---------------------------------------------------------
	// Demo 3: Parser Factory with Registration
	// ---------------------------------------------------------
	fmt.Println("\n--- Demo 3: Parser Factory with Registration ---")
	fmt.Println("Registering parsers dynamically...\n")

	parserFactory := NewParserFactory()

	// Register parsers (could be done at application startup)
	parserFactory.Register("json", func() DocumentParser { return &JSONParser{} })
	parserFactory.Register("xml", func() DocumentParser { return &XMLParser{} })
	parserFactory.Register("yaml", func() DocumentParser { return &YAMLParser{} })

	fmt.Printf("\nRegistered formats: %v\n\n", parserFactory.ListRegisteredFormats())

	// Use the registered parsers
	fmt.Println("Parsing documents:")

	jsonParser, _ := parserFactory.Create("json")
	jsonParser.Parse(`{"name": "John", "age": 30}`)

	xmlParser, _ := parserFactory.Create("xml")
	xmlParser.Parse(`<person><name>John</name></person>`)

	yamlParser, _ := parserFactory.Create("yaml")
	yamlParser.Parse("name: John\nage: 30")

	// Try unregistered format
	fmt.Println("\nTrying unregistered format:")
	_, err = parserFactory.Create("csv")
	if err != nil {
		fmt.Printf("  ⚠️ Error (expected): %v\n", err)
	}

	// ---------------------------------------------------------
	// Demo 4: Notification Factory (Function-based)
	// ---------------------------------------------------------
	fmt.Println("\n--- Demo 4: Notification Factory (Simple Function) ---")
	fmt.Println("Sending notifications through different channels...\n")

	// Email notification
	emailNotifier, _ := CreateNotifier(NotificationTypeEmail)
	emailNotifier.Send("user@example.com", "Welcome to our platform!")

	// SMS notification
	smsNotifier, _ := CreateNotifier(NotificationTypeSMS)
	smsNotifier.Send("+1-555-123-4567", "Your OTP is 123456")

	// Push notification
	pushNotifier, _ := CreateNotifier(NotificationTypePush)
	pushNotifier.Send("device-token-abc", "You have a new message!")

	fmt.Println("\n============================================================")
	fmt.Println("                    END OF DEMO")
	fmt.Println("============================================================")
}
