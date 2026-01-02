package main

import (
	"fmt"
	"math"
)

// ============================================================
// LISKOV SUBSTITUTION PRINCIPLE (LSP)
// ============================================================
//
// DEFINITION:
// "Objects of a superclass should be replaceable with objects
// of its subclasses without breaking the application."
//
// IN SIMPLE TERMS:
// If you write code that works with a base type (or interface),
// it should work correctly with ANY type that implements it.
// No surprises, no "special cases", no broken behavior.
//
// IN GO CONTEXT:
// Any struct implementing an interface must honor the interface's
// contract completely. If calling a method through the interface,
// the behavior should be predictable and consistent.
//
// ANALOGY:
// If you order "coffee" at a cafe, you expect caffeine and liquid.
// Whether it's espresso, latte, or americano - they all fulfill
// the "coffee contract". But if someone gives you water and says
// "it's a type of coffee" - that violates the contract!

// ============================================================
// PART 1: THE CLASSIC PROBLEM - Rectangle and Square
// ============================================================
// This is the most famous example of LSP violation.
// Mathematically: "A square IS-A rectangle" (all squares are rectangles)
// In code: This relationship breaks behavior!

// -------------------- BAD EXAMPLE --------------------

// BadRectangle represents a rectangle with width and height
type BadRectangle struct {
	width  float64
	height float64
}

// SetWidth sets the width of the rectangle
func (r *BadRectangle) SetWidth(newWidth float64) {
	r.width = newWidth
}

// SetHeight sets the height of the rectangle
func (r *BadRectangle) SetHeight(newHeight float64) {
	r.height = newHeight
}

// GetArea calculates and returns the area
func (r *BadRectangle) GetArea() float64 {
	return r.width * r.height
}

// BadSquare tries to "extend" Rectangle - THIS CAUSES PROBLEMS!
// A square has equal sides, so it must override the setters
type BadSquare struct {
	BadRectangle // Embedding (Go's form of "inheritance")
}

// SetWidth for square must set BOTH dimensions to keep sides equal
func (s *BadSquare) SetWidth(newWidth float64) {
	s.width = newWidth
	s.height = newWidth // Must keep sides equal - THIS BREAKS EXPECTATIONS!
}

// SetHeight for square must set BOTH dimensions to keep sides equal
func (s *BadSquare) SetHeight(newHeight float64) {
	s.width = newHeight // Must keep sides equal - THIS BREAKS EXPECTATIONS!
	s.height = newHeight
}

// demonstrateLSPViolation shows why Square extending Rectangle is problematic
func demonstrateLSPViolation() {
	fmt.Println("=== BAD: Rectangle-Square Problem (LSP Violation) ===")
	fmt.Println()

	// This function expects standard rectangle behavior
	testRectangle := func(rect *BadRectangle) {
		rect.SetWidth(5)  // Set width to 5
		rect.SetHeight(4) // Set height to 4
		expectedArea := 5.0 * 4.0
		actualArea := rect.GetArea()

		fmt.Printf("  Expected Area: %.0f (5 × 4)\n", expectedArea)
		fmt.Printf("  Actual Area:   %.0f\n", actualArea)

		if actualArea != expectedArea {
			fmt.Println("  ❌ VIOLATION! Area doesn't match expectation!")
		} else {
			fmt.Println("  ✅ Correct behavior")
		}
		fmt.Println()
	}

	// Test with Rectangle - works correctly
	fmt.Println("Testing with Rectangle:")
	rectangle := &BadRectangle{}
	testRectangle(rectangle)

	// Test with Square - BREAKS!
	// When we call SetHeight(4), it also changes width to 4
	// So we get 4×4=16 instead of 5×4=20
	fmt.Println("Testing with Square (passed as Rectangle):")
	square := &BadSquare{}
	testRectangle(&square.BadRectangle) // Pass the embedded Rectangle
	// Note: In Go, the overridden methods on BadSquare won't be called
	// when passing &square.BadRectangle. But conceptually, this demonstrates
	// why the mathematical "is-a" relationship doesn't work in code.

	fmt.Println("WHY THIS IS BAD:")
	fmt.Println("  - Code written for Rectangle doesn't work with Square")
	fmt.Println("  - Square's behavior surprises the caller")
	fmt.Println("  - You need special handling for different 'types'")
	fmt.Println()
}

// -------------------- GOOD EXAMPLE --------------------
// Solution: Don't use inheritance! Use interfaces based on BEHAVIOR.

// Shape defines what any shape can do - calculate area and perimeter
type Shape interface {
	GetArea() float64
	GetPerimeter() float64
	GetName() string
}

// Rectangle is a proper rectangle implementation
type Rectangle struct {
	Width  float64
	Height float64
}

func (r Rectangle) GetArea() float64 {
	return r.Width * r.Height
}

func (r Rectangle) GetPerimeter() float64 {
	return 2 * (r.Width + r.Height)
}

func (r Rectangle) GetName() string {
	return "Rectangle"
}

// Square is a completely separate type - NOT extending Rectangle!
type Square struct {
	Side float64
}

func (s Square) GetArea() float64 {
	return s.Side * s.Side
}

func (s Square) GetPerimeter() float64 {
	return 4 * s.Side
}

func (s Square) GetName() string {
	return "Square"
}

// Circle also implements Shape - shows LSP working correctly
type Circle struct {
	Radius float64
}

func (c Circle) GetArea() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) GetPerimeter() float64 {
	return 2 * math.Pi * c.Radius
}

func (c Circle) GetName() string {
	return "Circle"
}

// printShapeDetails works with ANY shape - LSP is satisfied!
// No matter which Shape you pass, it behaves correctly.
func printShapeDetails(shape Shape) {
	fmt.Printf("  %s: Area = %.2f, Perimeter = %.2f\n",
		shape.GetName(), shape.GetArea(), shape.GetPerimeter())
}

func demonstrateGoodShapes() {
	fmt.Println("=== GOOD: Shape Interface (LSP Compliant) ===")
	fmt.Println()

	// Create different shapes
	shapes := []Shape{
		Rectangle{Width: 5, Height: 4},
		Square{Side: 5},
		Circle{Radius: 3},
	}

	// ALL shapes work correctly through the interface!
	for _, shape := range shapes {
		printShapeDetails(shape)
	}

	fmt.Println()
	fmt.Println("WHY THIS IS GOOD:")
	fmt.Println("  - Each shape is independent - no inheritance issues")
	fmt.Println("  - All shapes honor the Shape interface contract")
	fmt.Println("  - Code using Shape works with ANY shape implementation")
	fmt.Println()
}

// ============================================================
// PART 2: REAL-WORLD EXAMPLE - Payment Processing
// ============================================================

// -------------------- BAD EXAMPLE --------------------
// Problem: Gift cards can't receive refunds, but interface forces them to

// BadPaymentProcessor defines payment operations
// This interface assumes ALL payment methods can do both pay AND refund
// Named "BadPaymentProcessor" to avoid conflict with PaymentProcessor in 02_ocp.go
type BadPaymentProcessor interface {
	ProcessPayment(amount float64) error
	ProcessRefund(amount float64) error
}

// CreditCardPayment handles credit card transactions
type CreditCardPayment struct {
	CardNumber  string
	CreditLimit float64
}

func (c *CreditCardPayment) ProcessPayment(amount float64) error {
	if amount > c.CreditLimit {
		return fmt.Errorf("amount $%.2f exceeds credit limit $%.2f", amount, c.CreditLimit)
	}
	fmt.Printf("  ✅ Paid $%.2f via Credit Card (****%s)\n", amount, c.CardNumber[len(c.CardNumber)-4:])
	return nil
}

func (c *CreditCardPayment) ProcessRefund(amount float64) error {
	fmt.Printf("  ✅ Refunded $%.2f to Credit Card (****%s)\n", amount, c.CardNumber[len(c.CardNumber)-4:])
	return nil
}

// DebitCardPayment handles debit card transactions
type DebitCardPayment struct {
	CardNumber string
	Balance    float64
}

func (d *DebitCardPayment) ProcessPayment(amount float64) error {
	if amount > d.Balance {
		return fmt.Errorf("insufficient balance: have $%.2f, need $%.2f", d.Balance, amount)
	}
	d.Balance -= amount
	fmt.Printf("  ✅ Paid $%.2f via Debit Card. Remaining balance: $%.2f\n", amount, d.Balance)
	return nil
}

func (d *DebitCardPayment) ProcessRefund(amount float64) error {
	d.Balance += amount
	fmt.Printf("  ✅ Refunded $%.2f to Debit Card. New balance: $%.2f\n", amount, d.Balance)
	return nil
}

// BadGiftCard violates LSP - it CAN'T refund but is forced to implement it
type BadGiftCard struct {
	CardCode string
	Balance  float64
}

func (g *BadGiftCard) ProcessPayment(amount float64) error {
	if amount > g.Balance {
		return fmt.Errorf("insufficient gift card balance: have $%.2f, need $%.2f", g.Balance, amount)
	}
	g.Balance -= amount
	fmt.Printf("  ✅ Paid $%.2f via Gift Card. Remaining: $%.2f\n", amount, g.Balance)
	return nil
}

func (g *BadGiftCard) ProcessRefund(_ float64) error {
	// ❌ LSP VIOLATION: This method exists but doesn't work!
	// Callers expect refund to work, but gift cards can't receive refunds
	// Note: We use _ for amount since we can't actually refund to gift cards
	return fmt.Errorf("❌ gift cards cannot receive refunds (LSP violation!)")
}

func demonstrateBadPayment() {
	fmt.Println("=== BAD: Payment Processor (LSP Violation) ===")
	fmt.Println()

	// Process payments and refunds
	payments := []BadPaymentProcessor{
		&CreditCardPayment{CardNumber: "1234567890123456", CreditLimit: 1000},
		&DebitCardPayment{CardNumber: "9876543210987654", Balance: 500},
		&BadGiftCard{CardCode: "GIFT-123", Balance: 100},
	}

	for _, payment := range payments {
		fmt.Printf("Processing with %T:\n", payment)

		// Payment works for all
		if err := payment.ProcessPayment(50); err != nil {
			fmt.Printf("  Payment Error: %v\n", err)
		}

		// Refund BREAKS for gift card - LSP violation!
		if err := payment.ProcessRefund(25); err != nil {
			fmt.Printf("  Refund Error: %v\n", err)
		}
		fmt.Println()
	}

	fmt.Println("PROBLEM: Gift card claims to support refunds but fails at runtime!")
	fmt.Println()
}

// -------------------- GOOD EXAMPLE --------------------
// Solution: Split interface into smaller, focused interfaces

// Payable - for any payment method that can accept payments
type Payable interface {
	Pay(amount float64) error
	GetPaymentMethodName() string
}

// Refundable - only for payment methods that support refunds
type Refundable interface {
	Refund(amount float64) error
}

// FullPaymentMethod - for methods that support both (interface composition)
type FullPaymentMethod interface {
	Payable
	Refundable
}

// GoodCreditCard supports both payment and refund
type GoodCreditCard struct {
	CardNumber  string
	CreditLimit float64
}

func (c *GoodCreditCard) Pay(amount float64) error {
	if amount > c.CreditLimit {
		return fmt.Errorf("exceeds credit limit")
	}
	fmt.Printf("  ✅ Paid $%.2f via Credit Card\n", amount)
	return nil
}

func (c *GoodCreditCard) Refund(amount float64) error {
	fmt.Printf("  ✅ Refunded $%.2f to Credit Card\n", amount)
	return nil
}

func (c *GoodCreditCard) GetPaymentMethodName() string {
	return "Credit Card"
}

// GoodDebitCard supports both payment and refund
type GoodDebitCard struct {
	CardNumber string
	Balance    float64
}

func (d *GoodDebitCard) Pay(amount float64) error {
	if amount > d.Balance {
		return fmt.Errorf("insufficient balance")
	}
	d.Balance -= amount
	fmt.Printf("  ✅ Paid $%.2f via Debit Card. Balance: $%.2f\n", amount, d.Balance)
	return nil
}

func (d *GoodDebitCard) Refund(amount float64) error {
	d.Balance += amount
	fmt.Printf("  ✅ Refunded $%.2f to Debit Card. Balance: $%.2f\n", amount, d.Balance)
	return nil
}

func (d *GoodDebitCard) GetPaymentMethodName() string {
	return "Debit Card"
}

// GoodGiftCard ONLY implements Payable - honest about its capabilities!
type GoodGiftCard struct {
	CardCode string
	Balance  float64
}

func (g *GoodGiftCard) Pay(amount float64) error {
	if amount > g.Balance {
		return fmt.Errorf("insufficient gift card balance")
	}
	g.Balance -= amount
	fmt.Printf("  ✅ Paid $%.2f via Gift Card. Balance: $%.2f\n", amount, g.Balance)
	return nil
}

func (g *GoodGiftCard) GetPaymentMethodName() string {
	return "Gift Card"
}

// Note: GoodGiftCard does NOT implement Refund - and that's correct!
// It honestly declares it can only Pay, not Refund.

func demonstrateGoodPayment() {
	fmt.Println("=== GOOD: Separate Interfaces (LSP Compliant) ===")
	fmt.Println()

	// All payment methods can pay
	fmt.Println("Processing payments (all methods support this):")
	payableMethods := []Payable{
		&GoodCreditCard{CardNumber: "1234", CreditLimit: 1000},
		&GoodDebitCard{CardNumber: "5678", Balance: 500},
		&GoodGiftCard{CardCode: "GIFT-123", Balance: 100},
	}

	for _, method := range payableMethods {
		fmt.Printf("  %s: ", method.GetPaymentMethodName())
		if err := method.Pay(50); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	fmt.Println()
	fmt.Println("Processing refunds (only refundable methods):")
	// Only use methods that implement Refundable
	refundableMethods := []FullPaymentMethod{
		&GoodCreditCard{CardNumber: "1234", CreditLimit: 1000},
		&GoodDebitCard{CardNumber: "5678", Balance: 500},
		// GoodGiftCard is NOT here - it doesn't implement Refundable!
	}

	for _, method := range refundableMethods {
		fmt.Printf("  %s: ", method.GetPaymentMethodName())
		if err := method.Refund(25); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	fmt.Println()
	fmt.Println("WHY THIS IS GOOD:")
	fmt.Println("  - Gift card doesn't pretend to support refunds")
	fmt.Println("  - Compiler prevents calling Refund on gift card")
	fmt.Println("  - No runtime surprises - interfaces are honest!")
	fmt.Println()
}

// ============================================================
// PART 3: BIRD EXAMPLE - Classic Interface Segregation + LSP
// ============================================================

// -------------------- BAD EXAMPLE --------------------
// Problem: Not all birds can fly, but interface forces them to

// BadBird interface assumes all birds can fly - WRONG!
// This interface is kept unused intentionally to show what NOT to do.
// It would force penguins to implement Fly() which they can't do!
type BadBird interface {
	Fly() // Problem: What about penguins? Ostriches?
	Eat()
}

// Ensure BadBird is referenced (prevents unused warning)
var _ BadBird = nil

// A penguin implementing BadBird would have to fake the Fly method:
// func (p Penguin) Fly() { panic("I can't fly!") } // LSP Violation!

// -------------------- GOOD EXAMPLE --------------------
// Solution: Separate behaviors into focused interfaces

// Bird - base behavior all birds share
type Bird interface {
	Eat()
	GetSpecies() string
}

// FlyingBird - only for birds that can actually fly
type FlyingBird interface {
	Bird
	Fly()
}

// SwimmingBird - for birds that can swim
type SwimmingBird interface {
	Bird
	Swim()
}

// Sparrow can fly but not swim
type Sparrow struct{}

func (s Sparrow) Eat()               { fmt.Println("    Sparrow eating seeds...") }
func (s Sparrow) Fly()               { fmt.Println("    Sparrow flying through the sky!") }
func (s Sparrow) GetSpecies() string { return "Sparrow" }

// Penguin can swim but not fly
type Penguin struct{}

func (p Penguin) Eat()               { fmt.Println("    Penguin eating fish...") }
func (p Penguin) Swim()              { fmt.Println("    Penguin swimming gracefully!") }
func (p Penguin) GetSpecies() string { return "Penguin" }

// Duck can both fly AND swim!
type Duck struct{}

func (d Duck) Eat()               { fmt.Println("    Duck eating bread...") }
func (d Duck) Fly()               { fmt.Println("    Duck flying over the pond!") }
func (d Duck) Swim()              { fmt.Println("    Duck swimming in the pond!") }
func (d Duck) GetSpecies() string { return "Duck" }

func demonstrateBirdExample() {
	fmt.Println("=== GOOD: Bird Interfaces (LSP Compliant) ===")
	fmt.Println()

	// All birds can eat
	fmt.Println("All birds eating:")
	allBirds := []Bird{Sparrow{}, Penguin{}, Duck{}}
	for _, bird := range allBirds {
		fmt.Printf("  %s:\n", bird.GetSpecies())
		bird.Eat()
	}

	fmt.Println()
	fmt.Println("Flying birds flying:")
	flyingBirds := []FlyingBird{Sparrow{}, Duck{}}
	for _, bird := range flyingBirds {
		fmt.Printf("  %s:\n", bird.GetSpecies())
		bird.Fly()
	}
	// Penguin is NOT in flyingBirds - because it can't fly!

	fmt.Println()
	fmt.Println("Swimming birds swimming:")
	swimmingBirds := []SwimmingBird{Penguin{}, Duck{}}
	for _, bird := range swimmingBirds {
		fmt.Printf("  %s:\n", bird.GetSpecies())
		bird.Swim()
	}
	// Sparrow is NOT in swimmingBirds - because it can't swim!

	fmt.Println()
	fmt.Println("WHY THIS IS GOOD:")
	fmt.Println("  - Penguin doesn't need to fake a Fly() method")
	fmt.Println("  - Duck can implement multiple interfaces (both flying and swimming)")
	fmt.Println("  - Each bird only claims capabilities it actually has")
	fmt.Println()
}

// ============================================================
// KEY TAKEAWAYS FOR INTERVIEWS
// ============================================================
//
// Q: How do you identify LSP violations?
// A: Look for these red flags:
//    1. Methods that throw "not supported" or "not implemented" errors
//    2. Methods with empty implementations that do nothing
//    3. Type checks or switch statements to handle "special" types
//    4. Subtype requires stricter conditions than parent
//
// Q: How do you fix LSP violations in Go?
// A: Follow these principles:
//    1. Design interfaces based on BEHAVIOR, not hierarchies
//    2. Use small, focused interfaces (Interface Segregation)
//    3. Prefer composition over inheritance
//    4. Don't force types to implement methods they can't support
//
// Q: Why is LSP important?
// A: Benefits:
//    1. Code using interfaces works with ANY implementation
//    2. No "special case" handling scattered throughout codebase
//    3. New implementations can be added without changing existing code
//    4. More testable - can easily use mock implementations

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║    LISKOV SUBSTITUTION PRINCIPLE (LSP) DEMONSTRATION     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Part 1: Rectangle and Square Problem
	demonstrateLSPViolation()

	fmt.Println("────────────────────────────────────────────────────────────")
	demonstrateGoodShapes()

	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	// Part 2: Payment Processing
	demonstrateBadPayment()

	fmt.Println("────────────────────────────────────────────────────────────")
	demonstrateGoodPayment()

	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()

	// Part 3: Bird Example
	demonstrateBirdExample()

	fmt.Println("════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("SUMMARY:")
	fmt.Println("  ❌ BAD:  Force types to implement unsupported behaviors")
	fmt.Println("  ✅ GOOD: Design interfaces around actual capabilities")
	fmt.Println()
	fmt.Println("Remember: If it walks like a duck and quacks like a duck,")
	fmt.Println("but needs batteries - you have the wrong abstraction!")
	fmt.Println()
}
