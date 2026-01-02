package main

import (
	"errors"
	"fmt"
)

// ============================================================
// OPEN/CLOSED PRINCIPLE (OCP)
// "Open for extension, closed for modification"
// ============================================================
//
// WHAT THIS MEANS:
// - You should be able to ADD new functionality
// - WITHOUT CHANGING existing code
//
// HOW TO ACHIEVE IN GO:
// - Use interfaces!
// - New features = new implementations of interface
// - Existing code stays untouched

// ============================================================
// SCENARIO: Payment Processing System
// ============================================================
// You start with Credit Card payments.
// Later, you need to add: PayPal, UPI, Crypto, etc.
// How do you add new payment methods without breaking existing code?

// ============================================================
// ❌ BAD EXAMPLE - Violates OCP
// ============================================================
// Every new payment method requires modifying this function.
// This is problematic because:
// 1. Existing tested code gets modified
// 2. Risk of introducing bugs in working code
// 3. Function grows larger and harder to maintain

func ProcessPaymentBad(amount float64, paymentType string) error {
	// Validate the amount first
	if amount <= 0 {
		return errors.New("payment amount must be positive")
	}

	switch paymentType {
	case "credit_card":
		fmt.Printf("Processing $%.2f via Credit Card\n", amount)
	case "paypal":
		fmt.Printf("Processing $%.2f via PayPal\n", amount)
	case "upi":
		fmt.Printf("Processing $%.2f via UPI\n", amount)
	default:
		// ❌ PROBLEM: Adding new payment = modifying this function
		// What if this function is tested and deployed?
		// What if someone adds a bug while adding new payment?
		return fmt.Errorf("unsupported payment type: %s", paymentType)
	}
	return nil
}

// ============================================================
// ✅ GOOD EXAMPLE - Follows OCP
// ============================================================
// Use interface to define a contract.
// Any payment method must implement this interface.

// PaymentProcessor defines what every payment method must do.
// This is our "contract" - any new payment method just needs to
// implement these two methods.
type PaymentProcessor interface {
	// ProcessPayment handles the actual payment logic
	ProcessPayment(amount float64) error

	// GetPaymentMethod returns the name of the payment method
	GetPaymentMethod() string
}

// ============================================================
// HELPER FUNCTIONS
// ============================================================

// maskString hides sensitive data, showing only the last few characters.
// Example: "4111111111111234" becomes "************1234"
func maskString(input string, visibleChars int) string {
	if len(input) <= visibleChars {
		return input // Nothing to mask if string is too short
	}

	// Create masked portion with asterisks
	maskedLength := len(input) - visibleChars
	masked := ""
	for i := 0; i < maskedLength; i++ {
		masked += "*"
	}

	// Append the visible portion
	return masked + input[len(input)-visibleChars:]
}

// validateAmount checks if the payment amount is valid
func validateAmount(amount float64) error {
	if amount <= 0 {
		return errors.New("payment amount must be greater than zero")
	}
	return nil
}

// ============================================================
// PAYMENT PROCESSOR IMPLEMENTATIONS
// ============================================================

// --- Credit Card Processor ---

// CreditCardProcessor handles credit card payments.
type CreditCardProcessor struct {
	CardNumber string // Full card number (will be masked when displayed)
	CVV        string // Card verification value
	ExpiryDate string // Format: MM/YY
}

// ProcessPayment processes a credit card payment.
func (processor *CreditCardProcessor) ProcessPayment(amount float64) error {
	// Validate the amount
	if err := validateAmount(amount); err != nil {
		return err
	}

	// Validate card number exists
	if len(processor.CardNumber) < 4 {
		return errors.New("invalid card number: too short")
	}

	// In real code: call payment gateway API here
	// For demo, we just print the transaction
	maskedCard := maskString(processor.CardNumber, 4)
	fmt.Printf("💳 Processing $%.2f via Credit Card ending in %s\n",
		amount, maskedCard)

	return nil
}

// GetPaymentMethod returns the name of this payment method.
func (processor *CreditCardProcessor) GetPaymentMethod() string {
	return "Credit Card"
}

// --- PayPal Processor ---

// PayPalProcessor handles PayPal payments.
type PayPalProcessor struct {
	Email string // PayPal account email
}

// ProcessPayment processes a PayPal payment.
func (processor *PayPalProcessor) ProcessPayment(amount float64) error {
	// Validate the amount
	if err := validateAmount(amount); err != nil {
		return err
	}

	// Validate email exists
	if processor.Email == "" {
		return errors.New("PayPal email is required")
	}

	fmt.Printf("🅿️  Processing $%.2f via PayPal account: %s\n", amount, processor.Email)
	return nil
}

// GetPaymentMethod returns the name of this payment method.
func (processor *PayPalProcessor) GetPaymentMethod() string {
	return "PayPal"
}

// --- UPI Processor ---

// UPIProcessor handles UPI payments (popular in India).
type UPIProcessor struct {
	UPIID string // UPI ID like "user@bank"
}

// ProcessPayment processes a UPI payment.
func (processor *UPIProcessor) ProcessPayment(amount float64) error {
	// Validate the amount
	if err := validateAmount(amount); err != nil {
		return err
	}

	// Validate UPI ID exists
	if processor.UPIID == "" {
		return errors.New("UPI ID is required")
	}

	fmt.Printf("📱 Processing $%.2f via UPI ID: %s\n", amount, processor.UPIID)
	return nil
}

// GetPaymentMethod returns the name of this payment method.
func (processor *UPIProcessor) GetPaymentMethod() string {
	return "UPI"
}

// ============================================================
// THE MAGIC OF OCP: PaymentService
// ============================================================
// This service works with the INTERFACE, not concrete types.
// It doesn't need to change when we add new payment methods!

// PaymentService processes any payment that implements PaymentProcessor.
type PaymentService struct{}

// Process handles a payment using any PaymentProcessor implementation.
// Adding a new payment method? This code stays UNCHANGED!
func (service *PaymentService) Process(processor PaymentProcessor, amount float64) error {
	fmt.Printf("Starting payment via %s...\n", processor.GetPaymentMethod())

	// Process the payment using the provided processor
	if err := processor.ProcessPayment(amount); err != nil {
		return fmt.Errorf("payment failed: %w", err)
	}

	fmt.Println("Payment successful! ✅")
	return nil
}

// ============================================================
// EXTENDING THE SYSTEM: Adding Cryptocurrency Payment
// ============================================================
// Notice: We DON'T modify PaymentService or any existing processor.
// We just ADD a new struct that implements PaymentProcessor.
// This is the power of OCP!

// CryptoProcessor handles cryptocurrency payments.
type CryptoProcessor struct {
	WalletAddress string // Crypto wallet address
	CoinType      string // Cryptocurrency type: BTC, ETH, etc.
}

// ProcessPayment processes a cryptocurrency payment.
func (processor *CryptoProcessor) ProcessPayment(amount float64) error {
	// Validate the amount
	if err := validateAmount(amount); err != nil {
		return err
	}

	// Validate wallet address
	if len(processor.WalletAddress) < 8 {
		return errors.New("invalid wallet address: too short")
	}

	// Validate coin type
	if processor.CoinType == "" {
		return errors.New("coin type is required")
	}

	// Mask the wallet address for security
	maskedWallet := processor.WalletAddress[:8] + "..."
	fmt.Printf("₿  Processing $%.2f via %s to wallet: %s\n",
		amount, processor.CoinType, maskedWallet)

	return nil
}

// GetPaymentMethod returns the name of this payment method.
func (processor *CryptoProcessor) GetPaymentMethod() string {
	return "Cryptocurrency (" + processor.CoinType + ")"
}

// ============================================================
// REAL-WORLD EXAMPLE #2: Notification System
// ============================================================
// Another example of OCP - adding new notification channels
// without modifying the NotificationService.

// Notifier defines the contract for sending notifications.
type Notifier interface {
	// Send delivers the message through this notification channel
	Send(message string) error

	// GetType returns the type of notification channel
	GetType() string
}

// --- Email Notifier ---

// EmailNotifier sends notifications via email.
type EmailNotifier struct {
	RecipientEmail string // Email address to send to
}

// Send delivers a message via email.
func (notifier *EmailNotifier) Send(message string) error {
	if notifier.RecipientEmail == "" {
		return errors.New("recipient email is required")
	}
	if message == "" {
		return errors.New("message cannot be empty")
	}

	fmt.Printf("📧 Email to %s: %s\n", notifier.RecipientEmail, message)
	return nil
}

// GetType returns the notification channel type.
func (notifier *EmailNotifier) GetType() string {
	return "Email"
}

// --- SMS Notifier ---

// SMSNotifier sends notifications via SMS.
type SMSNotifier struct {
	PhoneNumber string // Phone number to send SMS to
}

// Send delivers a message via SMS.
func (notifier *SMSNotifier) Send(message string) error {
	if notifier.PhoneNumber == "" {
		return errors.New("phone number is required")
	}
	if message == "" {
		return errors.New("message cannot be empty")
	}

	fmt.Printf("📱 SMS to %s: %s\n", notifier.PhoneNumber, message)
	return nil
}

// GetType returns the notification channel type.
func (notifier *SMSNotifier) GetType() string {
	return "SMS"
}

// --- Slack Notifier ---

// SlackNotifier sends notifications to a Slack channel.
type SlackNotifier struct {
	ChannelName string // Slack channel name (without #)
}

// Send delivers a message to a Slack channel.
func (notifier *SlackNotifier) Send(message string) error {
	if notifier.ChannelName == "" {
		return errors.New("channel name is required")
	}
	if message == "" {
		return errors.New("message cannot be empty")
	}

	fmt.Printf("💬 Slack to #%s: %s\n", notifier.ChannelName, message)
	return nil
}

// GetType returns the notification channel type.
func (notifier *SlackNotifier) GetType() string {
	return "Slack"
}

// ============================================================
// NOTIFICATION SERVICE
// ============================================================

// NotificationService manages multiple notification channels.
// It works with ANY notifier that implements the Notifier interface.
type NotificationService struct {
	notifiers []Notifier // List of registered notification channels
}

// NewNotificationService creates a new service with the provided notifiers.
func NewNotificationService(notifiers ...Notifier) *NotificationService {
	return &NotificationService{
		notifiers: notifiers,
	}
}

// AddNotifier registers a new notification channel.
// This method allows adding notifiers after service creation.
func (service *NotificationService) AddNotifier(notifier Notifier) {
	service.notifiers = append(service.notifiers, notifier)
}

// NotifyAll sends a message through ALL registered notifiers.
func (service *NotificationService) NotifyAll(message string) {
	if len(service.notifiers) == 0 {
		fmt.Println("Warning: No notifiers registered")
		return
	}

	for _, notifier := range service.notifiers {
		if err := notifier.Send(message); err != nil {
			fmt.Printf("❌ Failed to send via %s: %v\n", notifier.GetType(), err)
		}
	}
}

// ============================================================
// KEY INTERVIEW POINTS
// ============================================================
//
// Q: How does OCP help in real projects?
// A: 1. Reduces risk - existing tested code stays unchanged
//    2. Easier code reviews - only review new code
//    3. Parallel development - teams can add features independently
//    4. Better testability - each implementation can be tested in isolation
//
// Q: What enables OCP in Go?
// A: Interfaces! They define contracts that allow new implementations
//    without changing existing code.
//
// Q: Is it always possible to follow OCP?
// A: No! Sometimes you NEED to modify existing code (bug fixes, etc.)
//    OCP is about DESIGNING for extension from the start.
//
// Q: What's the relationship between OCP and interfaces?
// A: Interfaces are the mechanism that enables OCP in Go.
//    By depending on interfaces instead of concrete types,
//    code becomes open for extension (new implementations)
//    but closed for modification (existing code unchanged).

// ============================================================
// MAIN FUNCTION - DEMONSTRATION
// ============================================================

func main() {
	fmt.Println("=== Open/Closed Principle Demo ===")
	fmt.Println()

	// ---- PAYMENT PROCESSING DEMO ----
	fmt.Println("--- Payment Processing Example ---")
	fmt.Println()

	// Create the payment service (works with ANY payment processor)
	paymentService := &PaymentService{}

	// Example 1: Credit Card Payment
	creditCardProcessor := &CreditCardProcessor{
		CardNumber: "4111111111111234",
		CVV:        "123",
		ExpiryDate: "12/25",
	}
	err := paymentService.Process(creditCardProcessor, 99.99)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println()

	// Example 2: PayPal Payment
	paypalProcessor := &PayPalProcessor{
		Email: "user@example.com",
	}
	err = paymentService.Process(paypalProcessor, 49.99)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println()

	// Example 3: UPI Payment
	upiProcessor := &UPIProcessor{
		UPIID: "user@okbank",
	}
	err = paymentService.Process(upiProcessor, 29.99)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Println()

	// Example 4: Crypto Payment
	// NOTE: This is a NEW payment method!
	// But PaymentService code remains UNCHANGED - that's OCP!
	cryptoProcessor := &CryptoProcessor{
		WalletAddress: "0x1234567890abcdef1234567890abcdef",
		CoinType:      "ETH",
	}
	err = paymentService.Process(cryptoProcessor, 199.99)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println()
	fmt.Println("--- Notification System Example ---")
	fmt.Println()

	// ---- NOTIFICATION SYSTEM DEMO ----

	// Create notification service with multiple channels
	notificationService := NewNotificationService(
		&EmailNotifier{RecipientEmail: "user@example.com"},
		&SMSNotifier{PhoneNumber: "+1-234-567-8900"},
		&SlackNotifier{ChannelName: "order-alerts"},
	)

	// Send notification through all channels
	notificationService.NotifyAll("Your order has been shipped!")

	fmt.Println()
	fmt.Println("=== Demo Complete ===")
}
