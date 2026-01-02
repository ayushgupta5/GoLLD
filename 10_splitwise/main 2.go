package main

import (
	"errors"
	"fmt"
	"math"
	"sync"
)

// ============================================================
// SPLITWISE - Expense Sharing System
// ============================================================
//
// This application demonstrates how to build a simplified version of
// Splitwise - an expense sharing app where friends can split bills.
//
// Key Concepts Demonstrated:
// - Strategy Pattern: Different split strategies (Equal, Exact, Percent)
// - Complex Relationships: User-to-User balance tracking
// - Thread Safety: Using mutex for concurrent access
//
// How it works:
// 1. Users are created and added to the system
// 2. When someone pays for an expense, they specify how to split it
// 3. The system tracks who owes whom and by how much
// ============================================================

// ==================== SPLIT TYPE ENUM ====================

// SplitType defines how an expense should be divided among participants
type SplitType int

const (
	// SplitTypeEqual - Divide equally among all participants
	SplitTypeEqual SplitType = iota
	// SplitTypeExact - Each person pays a specific amount
	SplitTypeExact
	// SplitTypePercent - Each person pays a percentage of the total
	SplitTypePercent
)

// String returns a human-readable name for the split type
func (splitType SplitType) String() string {
	names := [...]string{"Equal", "Exact", "Percent"}
	if splitType < SplitTypeEqual || splitType > SplitTypePercent {
		return "Unknown"
	}
	return names[splitType]
}

// ==================== USER ====================

// User represents a person who can participate in expense sharing
type User struct {
	id    string // Unique identifier for the user
	name  string // Display name
	email string // Email address
	phone string // Phone number
}

// NewUser creates a new User with the given details
func NewUser(id, name, email, phone string) *User {
	return &User{
		id:    id,
		name:  name,
		email: email,
		phone: phone,
	}
}

// GetID returns the unique identifier of the user
func (user *User) GetID() string {
	return user.id
}

// GetName returns the display name of the user
func (user *User) GetName() string {
	return user.name
}

// GetEmail returns the email address of the user
func (user *User) GetEmail() string {
	return user.email
}

// GetPhone returns the phone number of the user
func (user *User) GetPhone() string {
	return user.phone
}

// ==================== SPLIT INTERFACE (Strategy Pattern) ====================

// Split is the interface that all split strategies must implement.
// This is an example of the Strategy Pattern - we define a common interface
// and multiple implementations that can be used interchangeably.
type Split interface {
	GetUserID() string  // Returns the ID of the user this split belongs to
	GetAmount() float64 // Returns the amount this user needs to pay
}

// ==================== EQUAL SPLIT ====================

// EqualSplit represents a split where the amount is calculated automatically
// by dividing the total equally among all participants.
type EqualSplit struct {
	userID string  // ID of the user participating in this split
	amount float64 // Amount calculated after dividing equally
}

// NewEqualSplit creates a new equal split for a user.
// The amount will be calculated later when the expense is processed.
func NewEqualSplit(userID string) *EqualSplit {
	return &EqualSplit{userID: userID}
}

// GetUserID returns the user ID for this split
func (split *EqualSplit) GetUserID() string {
	return split.userID
}

// GetAmount returns the calculated amount for this split
func (split *EqualSplit) GetAmount() float64 {
	return split.amount
}

// SetAmount sets the calculated amount for this split
// This is called by the expense manager after calculating equal shares
func (split *EqualSplit) SetAmount(amount float64) {
	split.amount = amount
}

// ==================== EXACT SPLIT ====================

// ExactSplit represents a split where the exact amount is specified upfront.
// Use this when each person pays a different, predetermined amount.
type ExactSplit struct {
	userID string  // ID of the user participating in this split
	amount float64 // Exact amount this user needs to pay
}

// NewExactSplit creates a new exact split with a specified amount
func NewExactSplit(userID string, amount float64) *ExactSplit {
	return &ExactSplit{
		userID: userID,
		amount: amount,
	}
}

// GetUserID returns the user ID for this split
func (split *ExactSplit) GetUserID() string {
	return split.userID
}

// GetAmount returns the exact amount for this split
func (split *ExactSplit) GetAmount() float64 {
	return split.amount
}

// ==================== PERCENT SPLIT ====================

// PercentSplit represents a split where each person pays a percentage of total.
// The percentages of all participants must add up to 100%.
type PercentSplit struct {
	userID     string  // ID of the user participating in this split
	percentage float64 // Percentage of total expense (e.g., 25 for 25%)
	amount     float64 // Calculated amount based on percentage
}

// NewPercentSplit creates a new percentage-based split
func NewPercentSplit(userID string, percentage float64) *PercentSplit {
	return &PercentSplit{
		userID:     userID,
		percentage: percentage,
	}
}

// GetUserID returns the user ID for this split
func (split *PercentSplit) GetUserID() string {
	return split.userID
}

// GetAmount returns the calculated amount based on percentage
func (split *PercentSplit) GetAmount() float64 {
	return split.amount
}

// GetPercentage returns the percentage for this split
func (split *PercentSplit) GetPercentage() float64 {
	return split.percentage
}

// SetAmount sets the calculated amount after applying percentage to total
func (split *PercentSplit) SetAmount(amount float64) {
	split.amount = amount
}

// ==================== EXPENSE ====================

// Expense represents a single expense that needs to be split among users.
// For example: A dinner bill of $100 paid by Alice, split among 4 friends.
type Expense struct {
	id          string    // Unique identifier for the expense
	totalAmount float64   // Total amount of the expense
	description string    // Description (e.g., "Dinner at restaurant")
	paidByUser  *User     // The user who paid for this expense
	splits      []Split   // How the expense should be divided
	splitType   SplitType // Type of split (Equal, Exact, or Percent)
}

// expenseIDGenerator handles thread-safe generation of expense IDs
var (
	expenseIDCounter int
	expenseIDMutex   sync.Mutex
)

// generateExpenseID creates a unique expense ID in a thread-safe manner
func generateExpenseID() string {
	expenseIDMutex.Lock()
	defer expenseIDMutex.Unlock()
	expenseIDCounter++
	return fmt.Sprintf("EXP-%d", expenseIDCounter)
}

// NewExpense creates a new expense with the given details
func NewExpense(totalAmount float64, description string, paidByUser *User, splits []Split, splitType SplitType) *Expense {
	return &Expense{
		id:          generateExpenseID(),
		totalAmount: totalAmount,
		description: description,
		paidByUser:  paidByUser,
		splits:      splits,
		splitType:   splitType,
	}
}

// GetID returns the unique identifier of the expense
func (expense *Expense) GetID() string {
	return expense.id
}

// GetDescription returns the description of the expense
func (expense *Expense) GetDescription() string {
	return expense.description
}

// GetTotalAmount returns the total amount of the expense
func (expense *Expense) GetTotalAmount() float64 {
	return expense.totalAmount
}

// ==================== BALANCE SHEET ====================

// BalanceSheet tracks the debts between users.
// It maintains a two-way relationship: if Alice owes Bob $50,
// the sheet shows both "Alice owes Bob $50" and "Bob is owed $50 by Alice"
type BalanceSheet struct {
	// userBalances is a nested map: userBalances[userA][userB] = amount
	// Positive amount means userA owes userB
	// Negative amount means userB owes userA (i.e., userA is owed)
	userBalances map[string]map[string]float64
	mutex        sync.RWMutex // Protects concurrent access to userBalances
}

// NewBalanceSheet creates a new empty balance sheet
func NewBalanceSheet() *BalanceSheet {
	return &BalanceSheet{
		userBalances: make(map[string]map[string]float64),
	}
}

// UpdateBalance records that borrowerID owes lenderID the specified amount.
// This updates both directions: borrower's debt increases, lender's credit increases.
func (balanceSheet *BalanceSheet) UpdateBalance(borrowerID, lenderID string, amount float64) {
	balanceSheet.mutex.Lock()
	defer balanceSheet.mutex.Unlock()

	// Initialize maps if they don't exist
	if balanceSheet.userBalances[borrowerID] == nil {
		balanceSheet.userBalances[borrowerID] = make(map[string]float64)
	}
	if balanceSheet.userBalances[lenderID] == nil {
		balanceSheet.userBalances[lenderID] = make(map[string]float64)
	}

	// borrower owes lender (positive value for borrower)
	balanceSheet.userBalances[borrowerID][lenderID] += amount

	// lender is owed by borrower (negative value for lender, meaning they should receive)
	balanceSheet.userBalances[lenderID][borrowerID] -= amount
}

// GetBalancesForUser returns all balances for a specific user.
// Positive values mean the user owes money, negative values mean they are owed.
func (balanceSheet *BalanceSheet) GetBalancesForUser(userID string) map[string]float64 {
	balanceSheet.mutex.RLock()
	defer balanceSheet.mutex.RUnlock()

	result := make(map[string]float64)

	if balances, exists := balanceSheet.userBalances[userID]; exists {
		for otherUserID, amount := range balances {
			// Only include significant amounts (ignore tiny rounding differences)
			if math.Abs(amount) > 0.01 {
				result[otherUserID] = amount
			}
		}
	}

	return result
}

// GetAllDebts returns all outstanding debts (only positive balances).
// This shows who owes money to whom without duplicating reverse relationships.
func (balanceSheet *BalanceSheet) GetAllDebts() map[string]map[string]float64 {
	balanceSheet.mutex.RLock()
	defer balanceSheet.mutex.RUnlock()

	result := make(map[string]map[string]float64)

	for borrowerID, balances := range balanceSheet.userBalances {
		for lenderID, amount := range balances {
			// Only include positive amounts (debts owed)
			if amount > 0.01 {
				if result[borrowerID] == nil {
					result[borrowerID] = make(map[string]float64)
				}
				result[borrowerID][lenderID] = amount
			}
		}
	}

	return result
}

// ==================== EXPENSE MANAGER ====================

// ExpenseManager is the main controller that manages users, expenses, and balances.
// It acts as a facade, providing a simplified interface to the underlying complexity.
type ExpenseManager struct {
	users        map[string]*User // All registered users (key: userID)
	expenses     []*Expense       // List of all expenses
	balanceSheet *BalanceSheet    // Tracks who owes whom
	mutex        sync.RWMutex     // Protects concurrent access
}

// NewExpenseManager creates a new expense manager
func NewExpenseManager() *ExpenseManager {
	return &ExpenseManager{
		users:        make(map[string]*User),
		expenses:     make([]*Expense, 0),
		balanceSheet: NewBalanceSheet(),
	}
}

// AddUser registers a new user in the system
func (manager *ExpenseManager) AddUser(user *User) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.users[user.GetID()] = user
}

// GetUser retrieves a user by their ID
func (manager *ExpenseManager) GetUser(userID string) *User {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	return manager.users[userID]
}

// AddExpense adds a new expense and updates all relevant balances.
// Returns an error if the split validation fails.
func (manager *ExpenseManager) AddExpense(expense *Expense) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	// Step 1: Validate and calculate split amounts
	if err := manager.validateAndCalculateSplits(expense); err != nil {
		return err
	}

	// Step 2: Store the expense
	manager.expenses = append(manager.expenses, expense)

	// Step 3: Update balances for each participant
	payerID := expense.paidByUser.GetID()
	for _, split := range expense.splits {
		participantID := split.GetUserID()

		// Skip if the participant is the payer (they don't owe themselves)
		if participantID == payerID {
			continue
		}

		// Record that this participant owes the payer
		manager.balanceSheet.UpdateBalance(participantID, payerID, split.GetAmount())
	}

	return nil
}

// validateAndCalculateSplits ensures the split is valid and calculates amounts where needed.
// For Equal splits: calculates equal share for each participant
// For Exact splits: validates that amounts sum to total
// For Percent splits: validates percentages sum to 100 and calculates amounts
func (manager *ExpenseManager) validateAndCalculateSplits(expense *Expense) error {
	switch expense.splitType {

	case SplitTypeEqual:
		// Calculate equal share: total / number of participants
		numberOfParticipants := len(expense.splits)
		if numberOfParticipants == 0 {
			return errors.New("no participants specified for equal split")
		}

		equalShare := expense.totalAmount / float64(numberOfParticipants)
		// Round to 2 decimal places for currency
		equalShare = roundToTwoDecimals(equalShare)

		// Assign the calculated amount to each equal split
		for _, split := range expense.splits {
			if equalSplit, isEqualSplit := split.(*EqualSplit); isEqualSplit {
				equalSplit.SetAmount(equalShare)
			}
		}

	case SplitTypeExact:
		// Validate that exact amounts sum to the total expense
		sumOfAmounts := 0.0
		for _, split := range expense.splits {
			sumOfAmounts += split.GetAmount()
		}

		if math.Abs(sumOfAmounts-expense.totalAmount) > 0.01 {
			return fmt.Errorf(
				"exact split amounts ($%.2f) don't match total expense ($%.2f)",
				sumOfAmounts,
				expense.totalAmount,
			)
		}

	case SplitTypePercent:
		// Validate that percentages sum to 100
		sumOfPercentages := 0.0
		for _, split := range expense.splits {
			if percentSplit, isPercentSplit := split.(*PercentSplit); isPercentSplit {
				sumOfPercentages += percentSplit.GetPercentage()
			}
		}

		if math.Abs(sumOfPercentages-100.0) > 0.01 {
			return fmt.Errorf(
				"percentages (%.2f%%) don't add up to 100%%",
				sumOfPercentages,
			)
		}

		// Calculate actual amounts based on percentages
		for _, split := range expense.splits {
			if percentSplit, isPercentSplit := split.(*PercentSplit); isPercentSplit {
				calculatedAmount := expense.totalAmount * percentSplit.GetPercentage() / 100.0
				calculatedAmount = roundToTwoDecimals(calculatedAmount)
				percentSplit.SetAmount(calculatedAmount)
			}
		}

	default:
		return fmt.Errorf("unknown split type: %v", expense.splitType)
	}

	return nil
}

// roundToTwoDecimals rounds a float to 2 decimal places (for currency)
func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

// PrintBalancesForUser displays all balances for a specific user
func (manager *ExpenseManager) PrintBalancesForUser(userID string) {
	user := manager.GetUser(userID)
	if user == nil {
		fmt.Printf("Error: User with ID '%s' not found\n", userID)
		return
	}

	balances := manager.balanceSheet.GetBalancesForUser(userID)
	if len(balances) == 0 {
		fmt.Printf("%s has no outstanding balances\n", user.GetName())
		return
	}

	fmt.Printf("\nBalances for %s:\n", user.GetName())
	for otherUserID, amount := range balances {
		otherUser := manager.GetUser(otherUserID)
		if otherUser == nil {
			continue
		}

		if amount > 0 {
			// Positive amount = this user owes the other user
			fmt.Printf("  -> Owes %s: $%.2f\n", otherUser.GetName(), amount)
		} else {
			// Negative amount = this user is owed by the other user
			fmt.Printf("  <- Is owed by %s: $%.2f\n", otherUser.GetName(), -amount)
		}
	}
}

// PrintAllBalances displays all outstanding balances in the system
func (manager *ExpenseManager) PrintAllBalances() {
	fmt.Println("\n+------------------------------------------+")
	fmt.Println("|            ALL BALANCES                  |")
	fmt.Println("+------------------------------------------+")

	allDebts := manager.balanceSheet.GetAllDebts()

	if len(allDebts) == 0 {
		fmt.Println("| No outstanding balances                  |")
		fmt.Println("+------------------------------------------+")
		return
	}

	for borrowerID, debts := range allDebts {
		borrower := manager.GetUser(borrowerID)
		for lenderID, amount := range debts {
			lender := manager.GetUser(lenderID)
			if borrower != nil && lender != nil {
				fmt.Printf("| %s owes %s: $%.2f\n",
					borrower.GetName(),
					lender.GetName(),
					amount,
				)
			}
		}
	}
	fmt.Println("+------------------------------------------+")
}

// ==================== GROUP (Optional Extension) ====================

// Group represents a collection of users who frequently share expenses.
// For example: "Roommates", "Trip to Paris", "Office Lunch Group"
type Group struct {
	id      string           // Unique identifier for the group
	name    string           // Display name of the group
	members map[string]*User // Members of the group (key: userID)
	mutex   sync.RWMutex     // Protects concurrent access
}

// NewGroup creates a new group with the given ID and name
func NewGroup(id, name string) *Group {
	return &Group{
		id:      id,
		name:    name,
		members: make(map[string]*User),
	}
}

// GetID returns the unique identifier of the group
func (group *Group) GetID() string {
	return group.id
}

// GetName returns the name of the group
func (group *Group) GetName() string {
	return group.name
}

// AddMember adds a user to the group
func (group *Group) AddMember(user *User) {
	group.mutex.Lock()
	defer group.mutex.Unlock()
	group.members[user.GetID()] = user
}

// RemoveMember removes a user from the group
func (group *Group) RemoveMember(userID string) {
	group.mutex.Lock()
	defer group.mutex.Unlock()
	delete(group.members, userID)
}

// GetMembers returns a list of all members in the group
func (group *Group) GetMembers() []*User {
	group.mutex.RLock()
	defer group.mutex.RUnlock()

	memberList := make([]*User, 0, len(group.members))
	for _, member := range group.members {
		memberList = append(memberList, member)
	}
	return memberList
}

// GetMemberCount returns the number of members in the group
func (group *Group) GetMemberCount() int {
	group.mutex.RLock()
	defer group.mutex.RUnlock()
	return len(group.members)
}

// ==================== MAIN FUNCTION ====================

func main() {
	fmt.Println("============================================")
	fmt.Println("       SPLITWISE - Expense Sharing")
	fmt.Println("============================================")

	// Step 1: Create the expense manager (our main controller)
	expenseManager := NewExpenseManager()

	// Step 2: Create some users
	alice := NewUser("U1", "Alice", "alice@email.com", "1111111111")
	bob := NewUser("U2", "Bob", "bob@email.com", "2222222222")
	charlie := NewUser("U3", "Charlie", "charlie@email.com", "3333333333")
	diana := NewUser("U4", "Diana", "diana@email.com", "4444444444")

	// Step 3: Register users with the expense manager
	expenseManager.AddUser(alice)
	expenseManager.AddUser(bob)
	expenseManager.AddUser(charlie)
	expenseManager.AddUser(diana)

	fmt.Println("\nUsers registered: Alice, Bob, Charlie, Diana")

	// ============================================
	// EXAMPLE 1: Equal Split
	// ============================================
	// Alice pays $100 for dinner, split equally among 4 friends
	// Each person's share: $100 / 4 = $25
	fmt.Println("\n--------------------------------------------")
	fmt.Println("EXPENSE 1: Dinner - $100 paid by Alice")
	fmt.Println("Split Type: EQUAL (divided equally among all)")

	dinnerExpense := NewExpense(
		100.0,    // Total amount
		"Dinner", // Description
		alice,    // Paid by Alice
		[]Split{ // Participants
			NewEqualSplit("U1"), // Alice
			NewEqualSplit("U2"), // Bob
			NewEqualSplit("U3"), // Charlie
			NewEqualSplit("U4"), // Diana
		},
		SplitTypeEqual, // Split type
	)

	if err := expenseManager.AddExpense(dinnerExpense); err != nil {
		fmt.Printf("Error adding expense: %v\n", err)
	} else {
		fmt.Println("Result: Each person owes $25")
		fmt.Println("  - Bob, Charlie, Diana each owe Alice $25")
	}

	expenseManager.PrintAllBalances()

	// ============================================
	// EXAMPLE 2: Exact Split
	// ============================================
	// Bob pays $50 for a movie, with specific amounts for each person
	fmt.Println("\n--------------------------------------------")
	fmt.Println("EXPENSE 2: Movie - $50 paid by Bob")
	fmt.Println("Split Type: EXACT (specific amounts)")

	movieExpense := NewExpense(
		50.0,    // Total amount
		"Movie", // Description
		bob,     // Paid by Bob
		[]Split{ // Participants with exact amounts
			NewExactSplit("U1", 20.0), // Alice pays $20
			NewExactSplit("U2", 10.0), // Bob pays $10 (himself)
			NewExactSplit("U3", 20.0), // Charlie pays $20
		},
		SplitTypeExact, // Split type
	)

	if err := expenseManager.AddExpense(movieExpense); err != nil {
		fmt.Printf("Error adding expense: %v\n", err)
	} else {
		fmt.Println("Result: Alice: $20, Bob: $10, Charlie: $20")
		fmt.Println("  - Alice owes Bob $20, Charlie owes Bob $20")
	}

	expenseManager.PrintAllBalances()

	// ============================================
	// EXAMPLE 3: Percentage Split
	// ============================================
	// Charlie pays $200 for groceries, split by percentage
	fmt.Println("\n--------------------------------------------")
	fmt.Println("EXPENSE 3: Groceries - $200 paid by Charlie")
	fmt.Println("Split Type: PERCENT (by percentage of total)")

	groceriesExpense := NewExpense(
		200.0,       // Total amount
		"Groceries", // Description
		charlie,     // Paid by Charlie
		[]Split{ // Participants with percentages
			NewPercentSplit("U1", 40.0), // Alice: 40% = $80
			NewPercentSplit("U2", 30.0), // Bob: 30% = $60
			NewPercentSplit("U3", 20.0), // Charlie: 20% = $40 (himself)
			NewPercentSplit("U4", 10.0), // Diana: 10% = $20
		},
		SplitTypePercent, // Split type
	)

	if err := expenseManager.AddExpense(groceriesExpense); err != nil {
		fmt.Printf("Error adding expense: %v\n", err)
	} else {
		fmt.Println("Result:")
		fmt.Println("  - Alice: 40% = $80")
		fmt.Println("  - Bob: 30% = $60")
		fmt.Println("  - Charlie: 20% = $40 (paid himself)")
		fmt.Println("  - Diana: 10% = $20")
	}

	expenseManager.PrintAllBalances()

	// ============================================
	// Display Individual User Balances
	// ============================================
	fmt.Println("\n--------------------------------------------")
	fmt.Println("INDIVIDUAL BALANCES:")
	expenseManager.PrintBalancesForUser("U1") // Alice
	expenseManager.PrintBalancesForUser("U2") // Bob
	expenseManager.PrintBalancesForUser("U3") // Charlie
	expenseManager.PrintBalancesForUser("U4") // Diana

	// ============================================
	// Summary of Key Design Patterns Used
	// ============================================
	fmt.Println("\n============================================")
	fmt.Println("KEY DESIGN PATTERNS & CONCEPTS:")
	fmt.Println("============================================")
	fmt.Println("1. Strategy Pattern: Different split types")
	fmt.Println("   (EqualSplit, ExactSplit, PercentSplit)")
	fmt.Println("   all implement the same Split interface")
	fmt.Println("")
	fmt.Println("2. Facade Pattern: ExpenseManager provides")
	fmt.Println("   a simple interface to complex subsystems")
	fmt.Println("")
	fmt.Println("3. Thread Safety: Mutex locks protect shared")
	fmt.Println("   data for concurrent access")
	fmt.Println("")
	fmt.Println("4. Bidirectional Balance Tracking: Debts are")
	fmt.Println("   recorded both ways for easy lookup")
	fmt.Println("============================================")
}
