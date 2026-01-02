package main

import (
	"errors"
	"fmt"
)

// ============================================================
// STATE DESIGN PATTERN
// ============================================================
//
// Definition:
// "The State Pattern allows an object to change its behavior when its
// internal state changes. The object will appear to change its class."
//
// ============================================================
// WHEN TO USE THE STATE PATTERN:
// ============================================================
// 1. When an object's behavior depends on its current state
// 2. When you have well-defined state transitions
// 3. When you want to eliminate large if-else or switch statements
//    that check the current state before performing an action
//
// ============================================================
// CLASSIC INTERVIEW PROBLEMS USING STATE PATTERN:
// ============================================================
// - Vending Machine (most common in interviews!)
// - ATM Machine
// - Traffic Light System
// - Document Workflow (Draft → Review → Published)
// - Order Status (Placed → Shipped → Delivered)
//
// ============================================================
// KEY CONCEPT:
// ============================================================
// Each state is represented by a separate struct that implements
// a common State interface. The context (main object) holds a
// reference to the current state and delegates all state-specific
// behavior to that state object.

// ============================================================
// EXAMPLE 1: VENDING MACHINE
// (Classic Interview Question - Most Frequently Asked!)
// ============================================================

// -----------------------------------------------------
// Step 1: Define the State Interface
// -----------------------------------------------------
// All states must implement these methods. Each state will handle
// these operations differently based on the machine's current state.

type VendingMachineState interface {
	// InsertMoney handles when a user inserts money into the machine
	InsertMoney(amount float64) error

	// SelectProduct handles when a user selects a product
	SelectProduct(productID string) error

	// Dispense handles the product dispensing process
	Dispense() error

	// CancelTransaction handles when a user wants to cancel and get refund
	CancelTransaction() error

	// GetStateName returns the name of the current state (for debugging/display)
	GetStateName() string
}

// -----------------------------------------------------
// Step 2: Define the Context (VendingMachine)
// -----------------------------------------------------
// The VendingMachine is the "context" that holds:
// - The current state
// - Data shared across states (balance, selected product, etc.)
// - References to all possible state objects (for state transitions)

type VendingMachine struct {
	// Current active state
	currentState VendingMachineState

	// Machine data shared across all states
	balance         float64            // Current money inserted by user
	selectedProduct string             // Product chosen by user
	productCatalog  map[string]float64 // Available products with prices

	// Pre-created state instances (we reuse these instead of creating new ones)
	// This is an optimization - we don't create new state objects on every transition
	idleState            VendingMachineState
	hasMoneyState        VendingMachineState
	productSelectedState VendingMachineState
	dispensingState      VendingMachineState
}

// NewVendingMachine creates and initializes a new vending machine
func NewVendingMachine() *VendingMachine {
	// Create the machine with some sample products
	machine := &VendingMachine{
		productCatalog: map[string]float64{
			"COLA":  1.50,
			"CHIPS": 1.00,
			"CANDY": 0.75,
			"WATER": 1.25,
		},
	}

	// Create all state objects and give them a reference to this machine
	// Each state needs access to the machine to:
	// 1. Read/modify machine data (balance, selected product)
	// 2. Trigger state transitions
	machine.idleState = &IdleState{machine: machine}
	machine.hasMoneyState = &HasMoneyState{machine: machine}
	machine.productSelectedState = &ProductSelectedState{machine: machine}
	machine.dispensingState = &DispensingState{machine: machine}

	// Start the machine in the idle state (waiting for user)
	machine.currentState = machine.idleState

	return machine
}

// transitionTo changes the machine to a new state and logs the transition
func (machine *VendingMachine) transitionTo(newState VendingMachineState) {
	oldStateName := machine.currentState.GetStateName()
	newStateName := newState.GetStateName()
	fmt.Printf("  [State Transition: %s → %s]\n", oldStateName, newStateName)
	machine.currentState = newState
}

// The following methods delegate the action to the current state
// This is the key to the State Pattern - the behavior changes based on currentState

func (machine *VendingMachine) InsertMoney(amount float64) error {
	return machine.currentState.InsertMoney(amount)
}

func (machine *VendingMachine) SelectProduct(productID string) error {
	return machine.currentState.SelectProduct(productID)
}

func (machine *VendingMachine) Dispense() error {
	return machine.currentState.Dispense()
}

func (machine *VendingMachine) CancelTransaction() error {
	return machine.currentState.CancelTransaction()
}

// GetStatus returns a human-readable status of the machine
func (machine *VendingMachine) GetStatus() string {
	return fmt.Sprintf("State: %s, Balance: $%.2f",
		machine.currentState.GetStateName(), machine.balance)
}

// -----------------------------------------------------
// Step 3: Implement Each State
// -----------------------------------------------------

// ----- IDLE STATE -----
// The machine is waiting for a user to insert money.
// Valid operations: InsertMoney
// Invalid operations: SelectProduct, Dispense, CancelTransaction

type IdleState struct {
	machine *VendingMachine // Reference to the context
}

func (state *IdleState) InsertMoney(amount float64) error {
	// Accept the money and add to balance
	state.machine.balance += amount
	fmt.Printf("💵 Inserted $%.2f. Current balance: $%.2f\n",
		amount, state.machine.balance)

	// Transition to HasMoney state since we now have money
	state.machine.transitionTo(state.machine.hasMoneyState)
	return nil
}

func (state *IdleState) SelectProduct(productID string) error {
	// Cannot select product without inserting money first
	return errors.New("please insert money first before selecting a product")
}

func (state *IdleState) Dispense() error {
	// Cannot dispense without money and product selection
	return errors.New("please insert money and select a product first")
}

func (state *IdleState) CancelTransaction() error {
	// No transaction in progress to cancel
	return errors.New("no active transaction to cancel")
}

func (state *IdleState) GetStateName() string {
	return "Idle"
}

// ----- HAS MONEY STATE -----
// User has inserted money, waiting for product selection.
// Valid operations: InsertMoney (add more), SelectProduct, CancelTransaction
// Invalid operations: Dispense (need to select product first)

type HasMoneyState struct {
	machine *VendingMachine
}

func (state *HasMoneyState) InsertMoney(amount float64) error {
	// User can add more money
	state.machine.balance += amount
	fmt.Printf("💵 Added $%.2f. Current balance: $%.2f\n",
		amount, state.machine.balance)
	// Stay in HasMoney state (no transition needed)
	return nil
}

func (state *HasMoneyState) SelectProduct(productID string) error {
	// Check if product exists in catalog
	productPrice, productExists := state.machine.productCatalog[productID]
	if !productExists {
		return fmt.Errorf("product '%s' not found in catalog", productID)
	}

	// Check if user has enough balance
	if state.machine.balance < productPrice {
		return fmt.Errorf("insufficient balance: need $%.2f, have $%.2f",
			productPrice, state.machine.balance)
	}

	// Product is available and user has enough money
	state.machine.selectedProduct = productID
	fmt.Printf("✅ Selected %s (Price: $%.2f)\n", productID, productPrice)

	// Transition to ProductSelected state
	state.machine.transitionTo(state.machine.productSelectedState)
	return nil
}

func (state *HasMoneyState) Dispense() error {
	// Cannot dispense without selecting a product first
	return errors.New("please select a product first")
}

func (state *HasMoneyState) CancelTransaction() error {
	// Refund the money and reset
	refundAmount := state.machine.balance
	state.machine.balance = 0
	fmt.Printf("💰 Transaction cancelled. Refunding $%.2f\n", refundAmount)

	// Go back to idle state
	state.machine.transitionTo(state.machine.idleState)
	return nil
}

func (state *HasMoneyState) GetStateName() string {
	return "HasMoney"
}

// ----- PRODUCT SELECTED STATE -----
// User has selected a product, ready to dispense.
// Valid operations: Dispense, CancelTransaction
// Invalid operations: InsertMoney, SelectProduct (product already selected)

type ProductSelectedState struct {
	machine *VendingMachine
}

func (state *ProductSelectedState) InsertMoney(amount float64) error {
	// Product is already selected, cannot accept more money
	return errors.New("product already selected - please collect your item or cancel")
}

func (state *ProductSelectedState) SelectProduct(productID string) error {
	// Product is already selected
	return errors.New("product already selected - please collect your item or cancel")
}

func (state *ProductSelectedState) Dispense() error {
	// Move to dispensing state and dispense the product
	state.machine.transitionTo(state.machine.dispensingState)
	return state.machine.Dispense() // Delegate to dispensing state
}

func (state *ProductSelectedState) CancelTransaction() error {
	// Refund money and reset selection
	refundAmount := state.machine.balance
	state.machine.balance = 0
	state.machine.selectedProduct = ""
	fmt.Printf("💰 Transaction cancelled. Refunding $%.2f\n", refundAmount)

	// Go back to idle state
	state.machine.transitionTo(state.machine.idleState)
	return nil
}

func (state *ProductSelectedState) GetStateName() string {
	return "ProductSelected"
}

// ----- DISPENSING STATE -----
// Machine is actively dispensing the product.
// Valid operations: Dispense (completes the process)
// Invalid operations: Everything else (machine is busy)

type DispensingState struct {
	machine *VendingMachine
}

func (state *DispensingState) InsertMoney(amount float64) error {
	return errors.New("please wait - dispensing in progress")
}

func (state *DispensingState) SelectProduct(productID string) error {
	return errors.New("please wait - dispensing in progress")
}

func (state *DispensingState) Dispense() error {
	// Get the selected product and its price
	productName := state.machine.selectedProduct
	productPrice := state.machine.productCatalog[productName]

	// Dispense the product
	fmt.Printf("🎁 Dispensing %s...\n", productName)

	// Deduct price from balance
	state.machine.balance -= productPrice

	// Return change if any
	if state.machine.balance > 0 {
		fmt.Printf("💰 Returning change: $%.2f\n", state.machine.balance)
	}

	// Reset the machine for next customer
	state.machine.balance = 0
	state.machine.selectedProduct = ""

	// Go back to idle state
	state.machine.transitionTo(state.machine.idleState)
	fmt.Println("✨ Thank you for your purchase!")

	return nil
}

func (state *DispensingState) CancelTransaction() error {
	// Cannot cancel during dispensing - product is already being dispensed
	return errors.New("cannot cancel - dispensing already in progress")
}

func (state *DispensingState) GetStateName() string {
	return "Dispensing"
}

// ============================================================
// EXAMPLE 2: ORDER STATUS (E-Commerce Application)
// ============================================================
// This example shows a simpler state machine with linear progression:
// Pending → Paid → Shipped → Delivered
// (or any state can transition to Cancelled)

// -----------------------------------------------------
// Step 1: Define the Order State Interface
// -----------------------------------------------------

type OrderState interface {
	// Process moves the order to the next logical state
	Process() error

	// Cancel attempts to cancel the order
	Cancel() error

	// GetStatus returns the current status name
	GetStatus() string
}

// -----------------------------------------------------
// Step 2: Define the Order Context
// -----------------------------------------------------

type Order struct {
	orderID      string     // Unique identifier for this order
	currentState OrderState // Current state of the order

	// All possible states
	pendingState   OrderState
	paidState      OrderState
	shippedState   OrderState
	deliveredState OrderState
	cancelledState OrderState
}

// NewOrder creates a new order with the given ID
func NewOrder(orderID string) *Order {
	order := &Order{orderID: orderID}

	// Initialize all states
	order.pendingState = &OrderPendingState{order: order}
	order.paidState = &OrderPaidState{order: order}
	order.shippedState = &OrderShippedState{order: order}
	order.deliveredState = &OrderDeliveredState{order: order}
	order.cancelledState = &OrderCancelledState{order: order}

	// New orders start in pending state
	order.currentState = order.pendingState

	return order
}

func (order *Order) transitionTo(newState OrderState) {
	order.currentState = newState
}

func (order *Order) Process() error {
	return order.currentState.Process()
}

func (order *Order) Cancel() error {
	return order.currentState.Cancel()
}

func (order *Order) GetStatus() string {
	return order.currentState.GetStatus()
}

// -----------------------------------------------------
// Step 3: Implement Each Order State
// -----------------------------------------------------

// ----- PENDING STATE -----
// Order is placed but payment not yet received

type OrderPendingState struct {
	order *Order
}

func (state *OrderPendingState) Process() error {
	fmt.Printf("📝 Order %s: Processing payment...\n", state.order.orderID)
	state.order.transitionTo(state.order.paidState)
	return nil
}

func (state *OrderPendingState) Cancel() error {
	fmt.Printf("❌ Order %s: Cancelled (no payment was made)\n", state.order.orderID)
	state.order.transitionTo(state.order.cancelledState)
	return nil
}

func (state *OrderPendingState) GetStatus() string {
	return "Pending"
}

// ----- PAID STATE -----
// Payment received, ready for shipping

type OrderPaidState struct {
	order *Order
}

func (state *OrderPaidState) Process() error {
	fmt.Printf("📦 Order %s: Shipping order...\n", state.order.orderID)
	state.order.transitionTo(state.order.shippedState)
	return nil
}

func (state *OrderPaidState) Cancel() error {
	fmt.Printf("❌ Order %s: Cancelled (refund initiated)\n", state.order.orderID)
	state.order.transitionTo(state.order.cancelledState)
	return nil
}

func (state *OrderPaidState) GetStatus() string {
	return "Paid"
}

// ----- SHIPPED STATE -----
// Order is in transit to customer

type OrderShippedState struct {
	order *Order
}

func (state *OrderShippedState) Process() error {
	fmt.Printf("✅ Order %s: Delivered successfully!\n", state.order.orderID)
	state.order.transitionTo(state.order.deliveredState)
	return nil
}

func (state *OrderShippedState) Cancel() error {
	// Cannot cancel once shipped
	return errors.New("cannot cancel order that is already shipped")
}

func (state *OrderShippedState) GetStatus() string {
	return "Shipped"
}

// ----- DELIVERED STATE -----
// Order has been delivered - final state

type OrderDeliveredState struct {
	order *Order
}

func (state *OrderDeliveredState) Process() error {
	return errors.New("order is already delivered - no further processing needed")
}

func (state *OrderDeliveredState) Cancel() error {
	return errors.New("cannot cancel a delivered order")
}

func (state *OrderDeliveredState) GetStatus() string {
	return "Delivered"
}

// ----- CANCELLED STATE -----
// Order has been cancelled - final state

type OrderCancelledState struct {
	order *Order
}

func (state *OrderCancelledState) Process() error {
	return errors.New("cannot process a cancelled order")
}

func (state *OrderCancelledState) Cancel() error {
	return errors.New("order is already cancelled")
}

func (state *OrderCancelledState) GetStatus() string {
	return "Cancelled"
}

// ============================================================
// EXAMPLE 3: TRAFFIC LIGHT (Simplest State Pattern Example)
// ============================================================
// This is the most basic example with circular state transitions:
// Red → Green → Yellow → Red → ...

// -----------------------------------------------------
// Step 1: Define the Traffic Light State Interface
// -----------------------------------------------------

type TrafficLightState interface {
	// Change transitions to the next light color
	Change()

	// GetColor returns the current color
	GetColor() string

	// GetDuration returns how long this light stays on (in seconds)
	GetDuration() int
}

// -----------------------------------------------------
// Step 2: Define the Traffic Light Context
// -----------------------------------------------------

type TrafficLight struct {
	currentState TrafficLightState

	// All possible light states
	redState    TrafficLightState
	yellowState TrafficLightState
	greenState  TrafficLightState
}

// NewTrafficLight creates a new traffic light starting with red
func NewTrafficLight() *TrafficLight {
	light := &TrafficLight{}

	// Initialize states
	light.redState = &RedLightState{trafficLight: light}
	light.yellowState = &YellowLightState{trafficLight: light}
	light.greenState = &GreenLightState{trafficLight: light}

	// Traffic lights typically start with red for safety
	light.currentState = light.redState

	return light
}

func (light *TrafficLight) transitionTo(newState TrafficLightState) {
	light.currentState = newState
}

func (light *TrafficLight) Change() {
	light.currentState.Change()
}

func (light *TrafficLight) GetStatus() string {
	return fmt.Sprintf("%s (duration: %d seconds)",
		light.currentState.GetColor(), light.currentState.GetDuration())
}

// -----------------------------------------------------
// Step 3: Implement Each Traffic Light State
// -----------------------------------------------------

// ----- RED LIGHT STATE -----

type RedLightState struct {
	trafficLight *TrafficLight
}

func (state *RedLightState) Change() {
	fmt.Println("🔴 → 🟢 Changing to Green")
	state.trafficLight.transitionTo(state.trafficLight.greenState)
}

func (state *RedLightState) GetColor() string {
	return "🔴 RED"
}

func (state *RedLightState) GetDuration() int {
	return 30 // Red light stays for 30 seconds
}

// ----- YELLOW LIGHT STATE -----

type YellowLightState struct {
	trafficLight *TrafficLight
}

func (state *YellowLightState) Change() {
	fmt.Println("🟡 → 🔴 Changing to Red")
	state.trafficLight.transitionTo(state.trafficLight.redState)
}

func (state *YellowLightState) GetColor() string {
	return "🟡 YELLOW"
}

func (state *YellowLightState) GetDuration() int {
	return 5 // Yellow light stays for 5 seconds
}

// ----- GREEN LIGHT STATE -----

type GreenLightState struct {
	trafficLight *TrafficLight
}

func (state *GreenLightState) Change() {
	fmt.Println("🟢 → 🟡 Changing to Yellow")
	state.trafficLight.transitionTo(state.trafficLight.yellowState)
}

func (state *GreenLightState) GetColor() string {
	return "🟢 GREEN"
}

func (state *GreenLightState) GetDuration() int {
	return 25 // Green light stays for 25 seconds
}

// ============================================================
// KEY INTERVIEW POINTS FOR STATE PATTERN
// ============================================================
//
// Q: What is the difference between State and Strategy patterns?
// A: - Strategy Pattern: The CLIENT chooses which algorithm/strategy to use.
//      The context doesn't change strategy on its own.
//    - State Pattern: State transitions are INTERNAL to the context.
//      The object changes its own state based on its logic.
//      State objects typically hold a reference to the context.
//
// Q: Where should state transition logic be placed?
// A: Two common approaches:
//    1. IN THE STATE ITSELF (used in our examples)
//       - Each state knows what the next state should be
//       - More decentralized, each state is self-contained
//       - Good when states have complex transition logic
//
//    2. IN THE CONTEXT (centralized approach)
//       - Context has a method like changeState() that handles all transitions
//       - States just request transitions, context decides
//       - Good when you need to enforce strict transition rules
//
// Q: How to handle invalid operations in a state?
// A: Always return an error! Never silently ignore invalid operations.
//    This makes debugging easier and prevents unexpected behavior.
//
// Q: Is the State Pattern thread-safe?
// A: Not by default! If multiple goroutines can call methods on the same
//    context, you need to add synchronization (mutex) to protect:
//    - State transitions
//    - Shared data (like balance in vending machine)
//
// ============================================================
// COMMON MISTAKES TO AVOID:
// ============================================================
// ❌ 1. Using if-else chains instead of separate state classes
//       Bad: if state == "idle" { ... } else if state == "hasMoneny" { ... }
//       Good: state.handleAction()
//
// ❌ 2. States knowing too much about other states
//       Each state should only know about itself and the context
//
// ❌ 3. Not implementing all interface methods in each state
//       Even if an action is invalid, implement it and return an error
//
// ❌ 4. Forgetting to handle error cases
//       Always consider: "What happens if this action is invalid?"

// ============================================================
// MAIN FUNCTION - Demonstration
// ============================================================

func main() {
	fmt.Println("=== STATE PATTERN DEMONSTRATION ===")
	fmt.Println()

	// -------------------------------------------------
	// Demo 1: Vending Machine - Normal Purchase Flow
	// -------------------------------------------------
	fmt.Println("--- DEMO 1: Vending Machine (Normal Purchase) ---")
	vendingMachine := NewVendingMachine()
	fmt.Println(vendingMachine.GetStatus())
	fmt.Println()

	// Try to select product without inserting money (should fail)
	fmt.Println("Attempting to select COLA without money...")
	if err := vendingMachine.SelectProduct("COLA"); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	}
	fmt.Println()

	// Insert money
	fmt.Println("Inserting $1.00...")
	vendingMachine.InsertMoney(1.00)
	fmt.Println(vendingMachine.GetStatus())
	fmt.Println()

	// Try to buy cola (costs $1.50, we only have $1.00 - should fail)
	fmt.Println("Attempting to buy COLA ($1.50) with $1.00...")
	if err := vendingMachine.SelectProduct("COLA"); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	}
	fmt.Println()

	// Add more money
	fmt.Println("Adding $1.00 more...")
	vendingMachine.InsertMoney(1.00)
	fmt.Println(vendingMachine.GetStatus())
	fmt.Println()

	// Now buy cola (have $2.00, cola costs $1.50)
	fmt.Println("Selecting COLA...")
	vendingMachine.SelectProduct("COLA")
	fmt.Println()

	// Dispense the product
	fmt.Println("Dispensing product...")
	vendingMachine.Dispense()
	fmt.Println(vendingMachine.GetStatus())
	fmt.Println()

	// -------------------------------------------------
	// Demo 2: Vending Machine - Cancel Flow
	// -------------------------------------------------
	fmt.Println("--- DEMO 2: Vending Machine (Cancel Transaction) ---")
	vendingMachine2 := NewVendingMachine()

	vendingMachine2.InsertMoney(2.00)
	vendingMachine2.SelectProduct("CHIPS")

	fmt.Println("Customer changed their mind, cancelling...")
	vendingMachine2.CancelTransaction()
	fmt.Println(vendingMachine2.GetStatus())
	fmt.Println()

	// -------------------------------------------------
	// Demo 3: Order Status Flow
	// -------------------------------------------------
	fmt.Println("--- DEMO 3: Order Status Flow ---")
	order := NewOrder("ORD-001")
	fmt.Printf("Initial Status: %s\n", order.GetStatus())
	fmt.Println()

	order.Process() // Pending → Paid
	fmt.Printf("After payment - Status: %s\n", order.GetStatus())

	order.Process() // Paid → Shipped
	fmt.Printf("After shipping - Status: %s\n", order.GetStatus())

	order.Process() // Shipped → Delivered
	fmt.Printf("After delivery - Status: %s\n", order.GetStatus())
	fmt.Println()

	// Try to cancel delivered order (should fail)
	fmt.Println("Attempting to cancel delivered order...")
	if err := order.Cancel(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	}
	fmt.Println()

	// -------------------------------------------------
	// Demo 4: Traffic Light
	// -------------------------------------------------
	fmt.Println("--- DEMO 4: Traffic Light System ---")
	trafficLight := NewTrafficLight()

	// Cycle through the lights
	for i := 0; i < 6; i++ {
		fmt.Printf("Current Light: %s\n", trafficLight.GetStatus())
		trafficLight.Change()
	}
}
