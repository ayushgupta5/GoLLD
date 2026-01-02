package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// ATM MACHINE - Low Level Design
// ============================================================================
//
// This implementation demonstrates three important design concepts:
//
// 1. STATE PATTERN: The ATM transitions through different states (Idle,
//    CardInserted, Authenticated, etc.) and validates operations based on
//    the current state.
//
// 2. CHAIN OF RESPONSIBILITY: Cash dispensers are linked in a chain. When
//    withdrawing money, each dispenser handles what it can and passes the
//    remaining amount to the next dispenser in the chain.
//
// 3. THREAD SAFETY: All shared data is protected with mutex locks to ensure
//    safe concurrent access.
//
// ============================================================================

// ============================================================================
// SECTION 1: ENUMS (Constants representing different types/states)
// ============================================================================

// TransactionType represents the type of transaction performed at the ATM.
// Using iota creates auto-incrementing integer constants starting from 0.
type TransactionType int

const (
	TransactionBalanceInquiry TransactionType = iota // Value: 0
	TransactionWithdraw                              // Value: 1
	TransactionDeposit                               // Value: 2
)

// String returns a human-readable name for the transaction type.
// This is useful for logging and displaying transaction details.
func (transactionType TransactionType) String() string {
	names := [...]string{"Balance Inquiry", "Withdrawal", "Deposit"}
	// Ensure we don't go out of bounds
	if transactionType < 0 || int(transactionType) >= len(names) {
		return "Unknown"
	}
	return names[transactionType]
}

// ATMState represents the current operational state of the ATM machine.
// The ATM can only perform certain operations based on its current state.
type ATMState int

const (
	StateIdle                ATMState = iota // ATM is waiting for a card
	StateCardInserted                        // Card has been inserted
	StateAuthenticated                       // PIN has been verified
	StateTransactionSelected                 // User has selected a transaction (reserved for future use)
	StateProcessing                          // Transaction is being processed (reserved for future use)
)

// String returns a human-readable name for the ATM state.
func (state ATMState) String() string {
	names := [...]string{"Idle", "Card Inserted", "Authenticated", "Transaction Selected", "Processing"}
	if state < 0 || int(state) >= len(names) {
		return "Unknown"
	}
	return names[state]
}

// ============================================================================
// SECTION 2: CARD - Represents a bank card
// ============================================================================

// Card represents a bank card that can be used at the ATM.
// Each card is linked to a bank account via the accountID.
type Card struct {
	cardNumber string // The card number (e.g., "4111111111111111")
	pin        string // The PIN code for authentication
	accountID  string // The ID of the linked bank account
}

// NewCard creates a new card with the given details.
// Parameters:
//   - cardNumber: The card number
//   - pin: The PIN code for the card
//   - accountID: The ID of the bank account linked to this card
func NewCard(cardNumber, pin, accountID string) *Card {
	return &Card{
		cardNumber: cardNumber,
		pin:        pin,
		accountID:  accountID,
	}
}

// ============================================================================
// SECTION 3: ACCOUNT - Represents a bank account
// ============================================================================

// Account represents a bank account with balance management.
// All operations on the balance are thread-safe using a mutex.
type Account struct {
	id           string     // Unique account identifier
	holderName   string     // Name of the account holder
	balance      float64    // Current balance in the account
	balanceMutex sync.Mutex // Mutex to protect balance from concurrent access
}

// NewAccount creates a new bank account.
// Parameters:
//   - id: Unique identifier for the account
//   - holderName: Name of the account holder
//   - initialBalance: Starting balance for the account
func NewAccount(id, holderName string, initialBalance float64) *Account {
	return &Account{
		id:         id,
		holderName: holderName,
		balance:    initialBalance,
	}
}

// GetBalance returns the current balance of the account.
// This method is thread-safe.
func (account *Account) GetBalance() float64 {
	account.balanceMutex.Lock()
	defer account.balanceMutex.Unlock()
	return account.balance
}

// Withdraw deducts the specified amount from the account balance.
// Returns an error if there are insufficient funds.
// This method is thread-safe.
func (account *Account) Withdraw(amount float64) error {
	account.balanceMutex.Lock()
	defer account.balanceMutex.Unlock()

	if amount > account.balance {
		return errors.New("insufficient funds in account")
	}
	account.balance -= amount
	return nil
}

// Deposit adds the specified amount to the account balance.
// This method is thread-safe.
func (account *Account) Deposit(amount float64) {
	account.balanceMutex.Lock()
	defer account.balanceMutex.Unlock()
	account.balance += amount
}

// ============================================================================
// SECTION 4: CASH DISPENSER - Chain of Responsibility Pattern
// ============================================================================
//
// The Chain of Responsibility pattern is used here to handle cash dispensing.
// Each dispenser handles a specific denomination (e.g., $100, $50, $20, $10).
// When dispensing cash:
//   1. The first dispenser tries to dispense as many notes as possible
//   2. The remaining amount is passed to the next dispenser in the chain
//   3. This continues until the full amount is dispensed or it fails
//
// ============================================================================

// CashDispenser is an interface that defines the contract for cash dispensers.
// Any cash dispenser must implement these two methods.
type CashDispenser interface {
	SetNext(nextDispenser CashDispenser) // Sets the next dispenser in the chain
	Dispense(amount int) error           // Dispenses the given amount
}

// NoteDispenser handles dispensing notes of a specific denomination.
// It implements the CashDispenser interface.
type NoteDispenser struct {
	denomination   int           // Value of each note (e.g., 100, 50, 20, 10)
	availableNotes int           // Number of notes available
	nextDispenser  CashDispenser // Reference to the next dispenser in the chain
	mutex          sync.Mutex    // Mutex for thread-safe operations
}

// NewNoteDispenser creates a new dispenser for a specific denomination.
// Parameters:
//   - denomination: The value of notes this dispenser handles (e.g., 100 for $100 notes)
//   - initialNoteCount: The initial number of notes in this dispenser
func NewNoteDispenser(denomination, initialNoteCount int) *NoteDispenser {
	return &NoteDispenser{
		denomination:   denomination,
		availableNotes: initialNoteCount,
	}
}

// SetNext sets the next dispenser in the chain.
// This allows creating a linked chain of dispensers.
func (dispenser *NoteDispenser) SetNext(nextDispenser CashDispenser) {
	dispenser.nextDispenser = nextDispenser
}

// Dispense attempts to dispense the requested amount using this denomination.
// Any remaining amount is passed to the next dispenser in the chain.
// Returns an error if the exact amount cannot be dispensed.
func (dispenser *NoteDispenser) Dispense(amount int) error {
	dispenser.mutex.Lock()
	defer dispenser.mutex.Unlock()

	// Base case: nothing left to dispense
	if amount <= 0 {
		return nil
	}

	// Calculate how many notes of this denomination we need
	notesNeeded := amount / dispenser.denomination

	// Limit to available notes
	if notesNeeded > dispenser.availableNotes {
		notesNeeded = dispenser.availableNotes
	}

	// Dispense notes if we can
	if notesNeeded > 0 {
		dispenser.availableNotes -= notesNeeded
		amountDispensed := notesNeeded * dispenser.denomination
		remainingAmount := amount - amountDispensed

		fmt.Printf("   💵 Dispensing %d x $%d notes\n", notesNeeded, dispenser.denomination)

		// Pass remaining amount to next dispenser in chain
		if remainingAmount > 0 {
			if dispenser.nextDispenser != nil {
				return dispenser.nextDispenser.Dispense(remainingAmount)
			}
			// No next dispenser and we still have remaining amount
			return errors.New("cannot dispense exact amount: insufficient denominations")
		}
		return nil
	}

	// We couldn't dispense any notes of this denomination, try next dispenser
	if dispenser.nextDispenser != nil {
		return dispenser.nextDispenser.Dispense(amount)
	}

	// No next dispenser and we couldn't dispense anything
	return errors.New("cannot dispense amount: no suitable denominations available")
}

// GetAvailableNotes returns the number of notes currently available.
// This method is thread-safe.
func (dispenser *NoteDispenser) GetAvailableNotes() int {
	dispenser.mutex.Lock()
	defer dispenser.mutex.Unlock()
	return dispenser.availableNotes
}

// AddNotes adds more notes to this dispenser (used for restocking).
// This method is thread-safe.
func (dispenser *NoteDispenser) AddNotes(count int) {
	dispenser.mutex.Lock()
	defer dispenser.mutex.Unlock()
	dispenser.availableNotes += count
}

// ============================================================================
// SECTION 5: TRANSACTION - Records all ATM transactions
// ============================================================================

// Transaction represents a single ATM transaction.
// All transactions are logged for record-keeping.
type Transaction struct {
	id              string          // Unique transaction identifier
	transactionType TransactionType // Type of transaction
	amount          float64         // Amount involved in the transaction
	accountID       string          // Account involved in the transaction
	timestamp       time.Time       // When the transaction occurred
	status          string          // Status: "Pending", "Completed", or "Failed: <reason>"
}

// transactionCounter is a thread-safe counter for generating unique transaction IDs.
// Using atomic operations ensures thread safety without explicit locking.
var transactionCounter int64

// NewTransaction creates a new transaction record.
// Parameters:
//   - transactionType: The type of transaction (Balance Inquiry, Withdrawal, Deposit)
//   - amount: The amount involved (0 for balance inquiry)
//   - accountID: The ID of the account involved
func NewTransaction(transactionType TransactionType, amount float64, accountID string) *Transaction {
	// Atomically increment the counter to ensure unique IDs across goroutines
	newID := atomic.AddInt64(&transactionCounter, 1)

	return &Transaction{
		id:              fmt.Sprintf("TXN-%d", newID),
		transactionType: transactionType,
		amount:          amount,
		accountID:       accountID,
		timestamp:       time.Now(),
		status:          "Pending",
	}
}

// ============================================================================
// SECTION 6: ATM MACHINE - The main ATM class
// ============================================================================

// ATM represents the ATM machine with all its components.
type ATM struct {
	id       string   // Unique ATM identifier
	location string   // Physical location of the ATM
	state    ATMState // Current operational state

	currentCard    *Card    // Currently inserted card (nil if none)
	currentAccount *Account // Currently active account (nil if none)

	// Cash Dispensers (Chain of Responsibility)
	// The chain is: $100 -> $50 -> $20 -> $10
	cashDispenserChain CashDispenser  // Head of the dispenser chain
	dispenser100       *NoteDispenser // $100 note dispenser
	dispenser50        *NoteDispenser // $50 note dispenser
	dispenser20        *NoteDispenser // $20 note dispenser
	dispenser10        *NoteDispenser // $10 note dispenser

	// Bank data (in a real system, this would be a database)
	registeredAccounts map[string]*Account // Map of account ID -> Account
	registeredCards    map[string]*Card    // Map of card number -> Card

	// Transaction history
	transactionHistory []*Transaction

	// Mutex for thread-safe ATM operations
	atmMutex sync.Mutex
}

// NewATM creates and initializes a new ATM machine.
// Parameters:
//   - id: Unique identifier for this ATM
//   - location: Physical location description
func NewATM(id, location string) *ATM {
	atm := &ATM{
		id:                 id,
		location:           location,
		state:              StateIdle,
		registeredAccounts: make(map[string]*Account),
		registeredCards:    make(map[string]*Card),
		transactionHistory: make([]*Transaction, 0),
	}

	// Initialize cash dispensers with default amounts
	// Each dispenser starts with 100 notes
	atm.dispenser100 = NewNoteDispenser(100, 100) // 100 x $100 = $10,000
	atm.dispenser50 = NewNoteDispenser(50, 100)   // 100 x $50 = $5,000
	atm.dispenser20 = NewNoteDispenser(20, 100)   // 100 x $20 = $2,000
	atm.dispenser10 = NewNoteDispenser(10, 100)   // 100 x $10 = $1,000

	// Build the Chain of Responsibility
	// $100 notes are tried first, then $50, then $20, then $10
	atm.dispenser100.SetNext(atm.dispenser50)
	atm.dispenser50.SetNext(atm.dispenser20)
	atm.dispenser20.SetNext(atm.dispenser10)

	// Set the head of the chain
	atm.cashDispenserChain = atm.dispenser100

	return atm
}

// RegisterAccount adds a bank account to the ATM's known accounts.
// In a real system, accounts would be stored in a central database.
func (atm *ATM) RegisterAccount(account *Account) {
	atm.registeredAccounts[account.id] = account
}

// RegisterCard adds a card to the ATM's known cards.
// In a real system, cards would be validated against a central system.
func (atm *ATM) RegisterCard(card *Card) {
	atm.registeredCards[card.cardNumber] = card
}

// InsertCard simulates inserting a card into the ATM.
// Returns an error if the ATM is busy or the card is not recognized.
func (atm *ATM) InsertCard(cardNumber string) error {
	atm.atmMutex.Lock()
	defer atm.atmMutex.Unlock()

	// Check if ATM is available
	if atm.state != StateIdle {
		return errors.New("ATM is currently busy, please wait")
	}

	// Validate the card
	card, cardExists := atm.registeredCards[cardNumber]
	if !cardExists {
		return errors.New("card not recognized")
	}

	// Card accepted - update state
	atm.currentCard = card
	atm.state = StateCardInserted

	// Show masked card number for security (only last 4 digits visible)
	maskedCardNumber := cardNumber[len(cardNumber)-4:]
	fmt.Printf("💳 Card inserted: ****%s\n", maskedCardNumber)

	return nil
}

// EnterPIN validates the entered PIN against the card's PIN.
// Returns an error if the PIN is incorrect or no card is inserted.
func (atm *ATM) EnterPIN(enteredPIN string) error {
	atm.atmMutex.Lock()
	defer atm.atmMutex.Unlock()

	// Verify state - card must be inserted first
	if atm.state != StateCardInserted {
		return errors.New("please insert your card first")
	}

	// Validate PIN
	if atm.currentCard.pin != enteredPIN {
		// Wrong PIN - eject card for security
		// Note: We call ejectCardInternal to avoid deadlock (already holding mutex)
		atm.ejectCardInternal()
		return errors.New("incorrect PIN - card has been ejected")
	}

	// PIN correct - find the linked account
	account, accountExists := atm.registeredAccounts[atm.currentCard.accountID]
	if !accountExists {
		atm.ejectCardInternal()
		return errors.New("account not found - card has been ejected")
	}

	// Authentication successful
	atm.currentAccount = account
	atm.state = StateAuthenticated
	fmt.Printf("✅ Authentication successful. Welcome, %s!\n", account.holderName)

	return nil
}

// CheckBalance displays and returns the current account balance.
// Also logs a balance inquiry transaction.
func (atm *ATM) CheckBalance() (float64, error) {
	atm.atmMutex.Lock()
	defer atm.atmMutex.Unlock()

	// Verify user is authenticated
	if atm.state != StateAuthenticated {
		return 0, errors.New("please authenticate first")
	}

	// Get current balance
	currentBalance := atm.currentAccount.GetBalance()

	// Log this transaction
	transaction := NewTransaction(TransactionBalanceInquiry, 0, atm.currentAccount.id)
	transaction.status = "Completed"
	atm.transactionHistory = append(atm.transactionHistory, transaction)

	fmt.Printf("💰 Current Balance: $%.2f\n", currentBalance)
	return currentBalance, nil
}

// Withdraw dispenses cash and deducts from the account.
// The amount must be positive and a multiple of 10.
func (atm *ATM) Withdraw(amount int) error {
	atm.atmMutex.Lock()
	defer atm.atmMutex.Unlock()

	// Verify user is authenticated
	if atm.state != StateAuthenticated {
		return errors.New("please authenticate first")
	}

	// Validate withdrawal amount
	if amount <= 0 {
		return errors.New("withdrawal amount must be positive")
	}
	if amount%10 != 0 {
		return errors.New("withdrawal amount must be a multiple of $10")
	}

	// Check sufficient balance
	currentBalance := atm.currentAccount.GetBalance()
	if float64(amount) > currentBalance {
		return errors.New("insufficient funds in your account")
	}

	// Create transaction record
	transaction := NewTransaction(TransactionWithdraw, float64(amount), atm.currentAccount.id)

	// Attempt to dispense cash
	fmt.Printf("\n💵 Dispensing $%d...\n", amount)
	dispensingError := atm.cashDispenserChain.Dispense(amount)
	if dispensingError != nil {
		transaction.status = "Failed: " + dispensingError.Error()
		atm.transactionHistory = append(atm.transactionHistory, transaction)
		return fmt.Errorf("withdrawal failed: %w", dispensingError)
	}

	// Cash dispensed successfully - deduct from account
	withdrawError := atm.currentAccount.Withdraw(float64(amount))
	if withdrawError != nil {
		// This shouldn't happen as we already checked the balance
		transaction.status = "Failed: " + withdrawError.Error()
		atm.transactionHistory = append(atm.transactionHistory, transaction)
		return withdrawError
	}

	// Transaction completed successfully
	transaction.status = "Completed"
	atm.transactionHistory = append(atm.transactionHistory, transaction)

	fmt.Printf("✅ Please take your cash: $%d\n", amount)
	fmt.Printf("   Remaining balance: $%.2f\n", atm.currentAccount.GetBalance())

	return nil
}

// Deposit adds cash to the account.
// The amount must be positive.
func (atm *ATM) Deposit(amount float64) error {
	atm.atmMutex.Lock()
	defer atm.atmMutex.Unlock()

	// Verify user is authenticated
	if atm.state != StateAuthenticated {
		return errors.New("please authenticate first")
	}

	// Validate deposit amount
	if amount <= 0 {
		return errors.New("deposit amount must be positive")
	}

	// Create transaction record
	transaction := NewTransaction(TransactionDeposit, amount, atm.currentAccount.id)

	// Add to account
	atm.currentAccount.Deposit(amount)

	// Transaction completed successfully
	transaction.status = "Completed"
	atm.transactionHistory = append(atm.transactionHistory, transaction)

	fmt.Printf("✅ Deposited: $%.2f\n", amount)
	fmt.Printf("   New balance: $%.2f\n", atm.currentAccount.GetBalance())

	return nil
}

// EjectCard ejects the current card and resets the ATM to idle state.
// This is the public method that acquires the lock.
func (atm *ATM) EjectCard() {
	atm.atmMutex.Lock()
	defer atm.atmMutex.Unlock()
	atm.ejectCardInternal()
}

// ejectCardInternal is the internal method that ejects the card.
// This method assumes the caller already holds the mutex.
// This pattern avoids deadlock when called from other methods that hold the lock.
func (atm *ATM) ejectCardInternal() {
	atm.currentCard = nil
	atm.currentAccount = nil
	atm.state = StateIdle
	fmt.Println("💳 Card ejected. Thank you for using our ATM!")
}

// DisplayCashStatus shows the current cash inventory in the ATM.
func (atm *ATM) DisplayCashStatus() {
	fmt.Println("\n╔════════════════════════════════════════╗")
	fmt.Println("║           ATM CASH STATUS              ║")
	fmt.Println("╠════════════════════════════════════════╣")
	fmt.Printf("║ $100 notes: %-3d (Total: $%d)\n", atm.dispenser100.GetAvailableNotes(), atm.dispenser100.GetAvailableNotes()*100)
	fmt.Printf("║ $50 notes:  %-3d (Total: $%d)\n", atm.dispenser50.GetAvailableNotes(), atm.dispenser50.GetAvailableNotes()*50)
	fmt.Printf("║ $20 notes:  %-3d (Total: $%d)\n", atm.dispenser20.GetAvailableNotes(), atm.dispenser20.GetAvailableNotes()*20)
	fmt.Printf("║ $10 notes:  %-3d (Total: $%d)\n", atm.dispenser10.GetAvailableNotes(), atm.dispenser10.GetAvailableNotes()*10)
	fmt.Println("╚════════════════════════════════════════╝")
}

// ============================================================================
// SECTION 7: MAIN - Demonstration of the ATM system
// ============================================================================

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("           🏧 ATM MACHINE DEMO")
	fmt.Println("═══════════════════════════════════════════")

	// Step 1: Create a new ATM machine
	atm := NewATM("ATM-001", "Main Street Branch")

	// Step 2: Set up bank accounts (simulating bank database)
	johnAccount := NewAccount("ACC001", "John Doe", 5000.00)
	janeAccount := NewAccount("ACC002", "Jane Smith", 10000.00)
	atm.RegisterAccount(johnAccount)
	atm.RegisterAccount(janeAccount)

	// Step 3: Register cards (simulating card issuance)
	johnCard := NewCard("4111111111111111", "1234", "ACC001")
	janeCard := NewCard("4222222222222222", "5678", "ACC002")
	atm.RegisterCard(johnCard)
	atm.RegisterCard(janeCard)

	atm.DisplayCashStatus()

	// ========== DEMO: John Doe's ATM Session ==========
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("📌 USER SESSION: John Doe")
	fmt.Println("─────────────────────────────────────────")

	// Step 4: Insert card
	err := atm.InsertCard("4111111111111111")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Step 5: Demonstrate wrong PIN handling
	fmt.Println("\n⚠️  Attempting authentication with wrong PIN...")
	err = atm.EnterPIN("0000") // Wrong PIN
	if err != nil {
		fmt.Printf("   ❌ %v\n", err)
	}

	// Step 6: Insert card again and use correct PIN
	err = atm.InsertCard("4111111111111111")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	err = atm.EnterPIN("1234") // Correct PIN
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Step 7: Check balance
	fmt.Println("\n📋 Checking account balance...")
	_, err = atm.CheckBalance()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Step 8: Withdraw cash
	fmt.Println("\n💸 Withdrawing $280...")
	err = atm.Withdraw(280)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Step 9: Deposit cash
	fmt.Println("\n💰 Depositing $500...")
	err = atm.Deposit(500)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Step 10: Check final balance
	fmt.Println("\n📋 Checking final balance...")
	_, err = atm.CheckBalance()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Step 11: End session
	fmt.Println()
	atm.EjectCard()

	atm.DisplayCashStatus()

	// ========== Summary of Design Patterns Used ==========
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  KEY DESIGN PATTERNS DEMONSTRATED:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. STATE PATTERN: ATM transitions through")
	fmt.Println("     states (Idle → CardInserted → Authenticated)")
	fmt.Println()
	fmt.Println("  2. CHAIN OF RESPONSIBILITY: Cash dispensers")
	fmt.Println("     are chained ($100 → $50 → $20 → $10)")
	fmt.Println()
	fmt.Println("  3. THREAD SAFETY: All shared data protected")
	fmt.Println("     with mutex locks for concurrent access")
	fmt.Println()
	fmt.Println("  4. TRANSACTION LOGGING: All operations are")
	fmt.Println("     recorded for audit purposes")
	fmt.Println("═══════════════════════════════════════════")
}
