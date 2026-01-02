package main

import (
	"errors"
	"fmt"
	"strings"
)

// ============================================================
// STRATEGY PATTERN
// "Define a family of algorithms, encapsulate each one, and make them interchangeable"
// ============================================================
//
// WHAT IS STRATEGY PATTERN?
// The Strategy pattern lets you define a family of algorithms (strategies),
// put each of them in a separate class, and make their objects interchangeable.
//
// ANALOGY:
// Think of it like choosing how to travel to work:
// - You can take a car (fast but expensive)
// - You can take a bus (cheap but slow)
// - You can ride a bike (healthy but weather-dependent)
// The destination (goal) is the same, but the strategy (how you get there) differs.
//
// WHY USE STRATEGY PATTERN?
// 1. Multiple algorithms exist for the same task
// 2. You need to switch algorithms at runtime
// 3. You want to avoid long if-else or switch statements
// 4. You want to easily add new algorithms without modifying existing code (Open/Closed Principle)
//
// REAL-WORLD EXAMPLES:
// - Payment methods (Credit Card, PayPal, UPI, Crypto)
// - Sorting algorithms (Bubble Sort, Quick Sort, Merge Sort)
// - File compression (ZIP, GZIP, RAR)
// - Navigation routes (Fastest, Shortest, Scenic)
//
// GO ADVANTAGE:
// Go's interfaces make the Strategy pattern very natural and clean!

// ============================================================
// EXAMPLE 1: Sorting Strategies
// ============================================================
// This example shows how different sorting algorithms can be
// swapped at runtime using the Strategy pattern.
// ============================================================

// SortStrategy defines the interface (contract) that all sorting algorithms must follow.
// Any struct that implements Sort() and GetName() can be used as a sorting strategy.
type SortStrategy interface {
	Sort(numbers []int) []int // Sorts the given numbers and returns sorted result
	GetName() string          // Returns the name of the sorting algorithm
}

// ----- Strategy 1: Bubble Sort -----
// BubbleSort is a simple sorting algorithm that repeatedly steps through the list,
// compares adjacent elements, and swaps them if they are in the wrong order.
// Time Complexity: O(n²) - Not efficient for large datasets
type BubbleSort struct{}

func (b *BubbleSort) Sort(numbers []int) []int {
	// Create a copy to avoid modifying the original slice
	sortedNumbers := make([]int, len(numbers))
	copy(sortedNumbers, numbers)

	totalNumbers := len(sortedNumbers)

	// Outer loop: each pass moves the largest unsorted element to its correct position
	for passIndex := 0; passIndex < totalNumbers-1; passIndex++ {
		// Inner loop: compare adjacent elements
		for currentIndex := 0; currentIndex < totalNumbers-passIndex-1; currentIndex++ {
			// If current element is greater than next element, swap them
			if sortedNumbers[currentIndex] > sortedNumbers[currentIndex+1] {
				sortedNumbers[currentIndex], sortedNumbers[currentIndex+1] =
					sortedNumbers[currentIndex+1], sortedNumbers[currentIndex]
			}
		}
	}
	return sortedNumbers
}

func (b *BubbleSort) GetName() string {
	return "Bubble Sort"
}

// ----- Strategy 2: Quick Sort -----
// QuickSort is an efficient divide-and-conquer sorting algorithm.
// It picks a pivot element and partitions the array around the pivot.
// Time Complexity: O(n log n) on average - Much faster for large datasets
type QuickSort struct{}

func (q *QuickSort) Sort(numbers []int) []int {
	// Create a copy to avoid modifying the original slice
	sortedNumbers := make([]int, len(numbers))
	copy(sortedNumbers, numbers)

	q.quickSortRecursive(sortedNumbers, 0, len(sortedNumbers)-1)
	return sortedNumbers
}

// quickSortRecursive is the recursive helper function for QuickSort
func (q *QuickSort) quickSortRecursive(arr []int, lowIndex, highIndex int) {
	if lowIndex < highIndex {
		// Find the partition index (pivot is now at its correct position)
		pivotIndex := q.partitionArray(arr, lowIndex, highIndex)

		// Recursively sort elements before and after the pivot
		q.quickSortRecursive(arr, lowIndex, pivotIndex-1)
		q.quickSortRecursive(arr, pivotIndex+1, highIndex)
	}
}

// partitionArray rearranges elements around a pivot
func (q *QuickSort) partitionArray(arr []int, lowIndex, highIndex int) int {
	// Choose the last element as the pivot
	pivot := arr[highIndex]

	// Index of the smaller element (elements smaller than pivot go to the left)
	smallerElementIndex := lowIndex - 1

	// Compare each element with the pivot
	for currentIndex := lowIndex; currentIndex < highIndex; currentIndex++ {
		if arr[currentIndex] < pivot {
			smallerElementIndex++
			// Swap: move smaller element to the left side
			arr[smallerElementIndex], arr[currentIndex] = arr[currentIndex], arr[smallerElementIndex]
		}
	}

	// Place the pivot in its correct position
	arr[smallerElementIndex+1], arr[highIndex] = arr[highIndex], arr[smallerElementIndex+1]
	return smallerElementIndex + 1
}

func (q *QuickSort) GetName() string {
	return "Quick Sort"
}

// ----- Context: Sorter -----
// The Sorter is the "Context" in the Strategy pattern.
// It holds a reference to a sorting strategy and delegates the sorting work to it.
// The client code can change the sorting strategy at runtime.
type Sorter struct {
	currentStrategy SortStrategy // The sorting algorithm currently being used
}

// NewSorter creates a new Sorter with the specified sorting strategy
func NewSorter(strategy SortStrategy) *Sorter {
	return &Sorter{currentStrategy: strategy}
}

// SetStrategy allows changing the sorting algorithm at runtime
// This is the key feature of the Strategy pattern!
func (s *Sorter) SetStrategy(strategy SortStrategy) {
	s.currentStrategy = strategy
}

// Sort uses the current strategy to sort the numbers
func (s *Sorter) Sort(numbers []int) []int {
	fmt.Printf("Using %s algorithm\n", s.currentStrategy.GetName())
	return s.currentStrategy.Sort(numbers)
}

// ============================================================
// EXAMPLE 2: Payment Strategies (Interview Favorite!)
// ============================================================
// This is one of the most common examples asked in LLD interviews.
// It demonstrates how different payment methods can be implemented
// using the Strategy pattern.
// ============================================================

// PaymentStrategy defines the interface for all payment methods.
// Each payment method must implement these three methods.
type PaymentStrategy interface {
	Pay(amount float64) error // Process the payment
	Validate() error          // Validate payment details before processing
	GetName() string          // Return the name of the payment method
}

// ----- Strategy 1: Credit Card Payment -----
// CreditCardPayment handles payments made via credit card.
type CreditCardPayment struct {
	cardNumber     string // The 16-digit credit card number
	cvv            string // The 3-digit security code on the back of the card
	expiryDate     string // Card expiry date in MM/YY format
	cardHolderName string // Name of the card holder
}

// NewCreditCardPayment creates a new credit card payment strategy
func NewCreditCardPayment(cardNumber, cvv, expiryDate, cardHolderName string) *CreditCardPayment {
	return &CreditCardPayment{
		cardNumber:     cardNumber,
		cvv:            cvv,
		expiryDate:     expiryDate,
		cardHolderName: cardHolderName,
	}
}

// Validate checks if the credit card details are valid
func (c *CreditCardPayment) Validate() error {
	// Check if card number has minimum required length (simplified validation)
	if len(c.cardNumber) < 12 {
		return errors.New("invalid card number: must have at least 12 digits")
	}
	// CVV must be exactly 3 digits
	if len(c.cvv) != 3 {
		return errors.New("invalid CVV: must be exactly 3 digits")
	}
	// Expiry date is required
	if c.expiryDate == "" {
		return errors.New("expiry date is required")
	}
	return nil
}

// Pay processes the credit card payment
func (c *CreditCardPayment) Pay(amount float64) error {
	// Always validate before processing payment
	if err := c.Validate(); err != nil {
		return fmt.Errorf("credit card validation failed: %w", err)
	}

	// Mask the card number for security (show only last 4 digits)
	maskedCardNumber := strings.Repeat("*", len(c.cardNumber)-4) + c.cardNumber[len(c.cardNumber)-4:]

	fmt.Printf("💳 Processing Credit Card payment of $%.2f\n", amount)
	fmt.Printf("   Card: %s | Holder: %s\n", maskedCardNumber, c.cardHolderName)
	fmt.Println("   ✅ Payment successful!")
	return nil
}

func (c *CreditCardPayment) GetName() string {
	return "Credit Card"
}

// ----- Strategy 2: PayPal Payment -----
// PayPalPayment handles payments made via PayPal account.
type PayPalPayment struct {
	emailAddress string // PayPal account email
	password     string // PayPal account password (would be handled securely in real apps)
}

// NewPayPalPayment creates a new PayPal payment strategy
func NewPayPalPayment(email, password string) *PayPalPayment {
	return &PayPalPayment{
		emailAddress: email,
		password:     password,
	}
}

// Validate checks if the PayPal credentials are valid
func (p *PayPalPayment) Validate() error {
	if p.emailAddress == "" {
		return errors.New("PayPal email address is required")
	}
	// Basic email validation (simplified)
	if !strings.Contains(p.emailAddress, "@") {
		return errors.New("invalid email address format")
	}
	if p.password == "" {
		return errors.New("PayPal password is required")
	}
	return nil
}

// Pay processes the PayPal payment
func (p *PayPalPayment) Pay(amount float64) error {
	if err := p.Validate(); err != nil {
		return fmt.Errorf("PayPal validation failed: %w", err)
	}

	fmt.Printf("🅿️  Processing PayPal payment of $%.2f\n", amount)
	fmt.Printf("   Account: %s\n", p.emailAddress)
	fmt.Println("   ✅ Payment successful!")
	return nil
}

func (p *PayPalPayment) GetName() string {
	return "PayPal"
}

// ----- Strategy 3: UPI Payment -----
// UPIPayment handles payments made via UPI (Unified Payments Interface).
// UPI is a popular payment method in India.
type UPIPayment struct {
	upiID string // UPI ID format: username@bankname (e.g., user@okbank)
}

// NewUPIPayment creates a new UPI payment strategy
func NewUPIPayment(upiID string) *UPIPayment {
	return &UPIPayment{upiID: upiID}
}

// Validate checks if the UPI ID is valid
func (u *UPIPayment) Validate() error {
	if u.upiID == "" {
		return errors.New("UPI ID is required")
	}
	// UPI ID must contain '@' symbol
	if !strings.Contains(u.upiID, "@") {
		return errors.New("invalid UPI ID format: must contain '@' symbol")
	}
	return nil
}

// Pay processes the UPI payment
func (u *UPIPayment) Pay(amount float64) error {
	if err := u.Validate(); err != nil {
		return fmt.Errorf("UPI validation failed: %w", err)
	}

	fmt.Printf("📱 Processing UPI payment of $%.2f\n", amount)
	fmt.Printf("   UPI ID: %s\n", u.upiID)
	fmt.Println("   ✅ Payment successful!")
	return nil
}

func (u *UPIPayment) GetName() string {
	return "UPI"
}

// ----- Context: ShoppingCart -----
// ShoppingCart is the "Context" that uses a payment strategy.
// It contains items to purchase and delegates the payment to the selected strategy.
type ShoppingCart struct {
	items           []CartItem      // List of items in the cart
	paymentStrategy PaymentStrategy // The selected payment method
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	Name  string
	Price float64
}

// NewShoppingCart creates a new empty shopping cart
func NewShoppingCart() *ShoppingCart {
	return &ShoppingCart{
		items: make([]CartItem, 0),
	}
}

// AddItem adds a new item to the shopping cart
func (cart *ShoppingCart) AddItem(itemName string, price float64) {
	item := CartItem{Name: itemName, Price: price}
	cart.items = append(cart.items, item)
	fmt.Printf("   Added: %s ($%.2f)\n", itemName, price)
}

// GetTotal calculates the total price of all items in the cart
func (cart *ShoppingCart) GetTotal() float64 {
	var total float64
	for _, item := range cart.items {
		total += item.Price
	}
	return total
}

// GetItemNames returns a list of all item names in the cart
func (cart *ShoppingCart) GetItemNames() []string {
	names := make([]string, len(cart.items))
	for i, item := range cart.items {
		names[i] = item.Name
	}
	return names
}

// SetPaymentMethod allows the user to select a payment strategy
func (cart *ShoppingCart) SetPaymentMethod(strategy PaymentStrategy) {
	cart.paymentStrategy = strategy
	fmt.Printf("   Payment method set to: %s\n", strategy.GetName())
}

// Checkout processes the payment using the selected payment strategy
func (cart *ShoppingCart) Checkout() error {
	// Check if a payment method has been selected
	if cart.paymentStrategy == nil {
		return errors.New("checkout failed: no payment method selected")
	}

	// Check if cart has items
	if len(cart.items) == 0 {
		return errors.New("checkout failed: cart is empty")
	}

	// Display checkout summary
	fmt.Println("\n   📦 Checkout Summary:")
	fmt.Printf("   Items: %v\n", cart.GetItemNames())
	fmt.Printf("   Total Amount: $%.2f\n", cart.GetTotal())
	fmt.Println()

	// Process payment using the selected strategy
	return cart.paymentStrategy.Pay(cart.GetTotal())
}

// ============================================================
// EXAMPLE 3: Compression Strategies
// ============================================================
// This example demonstrates how file compression can use different
// algorithms (ZIP, GZIP) that can be swapped at runtime.
// ============================================================

// CompressionStrategy defines the interface for all compression algorithms.
type CompressionStrategy interface {
	Compress(data []byte) []byte   // Compresses the given data
	Decompress(data []byte) []byte // Decompresses the given data
	GetExtension() string          // Returns the file extension for this compression type
}

// ----- Strategy 1: ZIP Compression -----
// ZIPCompression simulates compression using the ZIP algorithm.
type ZIPCompression struct{}

func (z *ZIPCompression) Compress(data []byte) []byte {
	fmt.Println("   🗜️  Compressing using ZIP algorithm...")
	// Simulate compression by adding a header (in real code, use compress/zip package)
	compressedData := append([]byte("ZIP:"), data...)
	return compressedData
}

func (z *ZIPCompression) Decompress(data []byte) []byte {
	fmt.Println("   📂 Decompressing ZIP file...")
	// Remove the "ZIP:" header to get original data
	if len(data) > 4 {
		return data[4:]
	}
	return data
}

func (z *ZIPCompression) GetExtension() string {
	return ".zip"
}

// ----- Strategy 2: GZIP Compression -----
// GZIPCompression simulates compression using the GZIP algorithm.
type GZIPCompression struct{}

func (g *GZIPCompression) Compress(data []byte) []byte {
	fmt.Println("   🗜️  Compressing using GZIP algorithm...")
	// Simulate compression by adding a header (in real code, use compress/gzip package)
	compressedData := append([]byte("GZIP:"), data...)
	return compressedData
}

func (g *GZIPCompression) Decompress(data []byte) []byte {
	fmt.Println("   📂 Decompressing GZIP file...")
	// Remove the "GZIP:" header to get original data
	if len(data) > 5 {
		return data[5:]
	}
	return data
}

func (g *GZIPCompression) GetExtension() string {
	return ".gz"
}

// ----- Context: FileCompressor -----
// FileCompressor uses a compression strategy to compress/decompress files.
type FileCompressor struct {
	currentStrategy CompressionStrategy
}

// NewFileCompressor creates a new FileCompressor with the specified strategy
func NewFileCompressor(strategy CompressionStrategy) *FileCompressor {
	return &FileCompressor{currentStrategy: strategy}
}

// SetStrategy allows changing the compression algorithm at runtime
func (fc *FileCompressor) SetStrategy(strategy CompressionStrategy) {
	fc.currentStrategy = strategy
}

// CompressFile compresses the given data and returns the new filename and compressed data
func (fc *FileCompressor) CompressFile(filename string, data []byte) (newFilename string, compressedData []byte) {
	compressedData = fc.currentStrategy.Compress(data)
	newFilename = filename + fc.currentStrategy.GetExtension()
	fmt.Printf("   ✅ File compressed: %s (%d bytes -> %d bytes)\n",
		newFilename, len(data), len(compressedData))
	return newFilename, compressedData
}

// ============================================================
// EXAMPLE 4: Navigation/Route Calculation Strategies
// ============================================================
// This example shows how a GPS navigation system can offer different
// route options (fastest, shortest, scenic) using the Strategy pattern.
// ============================================================

// RouteStrategy defines the interface for different route calculation methods.
type RouteStrategy interface {
	CalculateRoute(fromLocation, toLocation string) string // Calculates and returns the route description
	GetEstimatedTime() int                                 // Returns estimated travel time in minutes
	GetName() string                                       // Returns the name of this route type
}

// ----- Strategy 1: Fastest Route -----
// FastestRoute prioritizes speed over distance (uses highways).
type FastestRoute struct{}

func (fr *FastestRoute) CalculateRoute(fromLocation, toLocation string) string {
	return fmt.Sprintf("🚀 Highway route from %s to %s (fastest, may have tolls)", fromLocation, toLocation)
}

func (fr *FastestRoute) GetEstimatedTime() int {
	return 30 // 30 minutes
}

func (fr *FastestRoute) GetName() string {
	return "Fastest Route"
}

// ----- Strategy 2: Shortest Route -----
// ShortestRoute prioritizes distance over time (takes the most direct path).
type ShortestRoute struct{}

func (sr *ShortestRoute) CalculateRoute(fromLocation, toLocation string) string {
	return fmt.Sprintf("📏 Direct route from %s to %s (shortest distance, may have traffic)", fromLocation, toLocation)
}

func (sr *ShortestRoute) GetEstimatedTime() int {
	return 45 // 45 minutes due to potential traffic
}

func (sr *ShortestRoute) GetName() string {
	return "Shortest Route"
}

// ----- Strategy 3: Scenic Route -----
// ScenicRoute prioritizes beautiful views over speed or distance.
type ScenicRoute struct{}

func (scr *ScenicRoute) CalculateRoute(fromLocation, toLocation string) string {
	return fmt.Sprintf("🏔️  Scenic route from %s to %s (enjoy beautiful views!)", fromLocation, toLocation)
}

func (scr *ScenicRoute) GetEstimatedTime() int {
	return 90 // 90 minutes but worth the views!
}

func (scr *ScenicRoute) GetName() string {
	return "Scenic Route"
}

// ----- Context: Navigator -----
// Navigator is a GPS system that uses different route strategies.
type Navigator struct {
	currentRouteStrategy RouteStrategy
}

// NewNavigator creates a new Navigator with the specified route strategy
func NewNavigator(strategy RouteStrategy) *Navigator {
	return &Navigator{currentRouteStrategy: strategy}
}

// SetStrategy allows changing the route strategy at runtime
func (nav *Navigator) SetStrategy(strategy RouteStrategy) {
	nav.currentRouteStrategy = strategy
	fmt.Printf("   Route preference changed to: %s\n", strategy.GetName())
}

// Navigate calculates and displays the route between two locations
func (nav *Navigator) Navigate(fromLocation, toLocation string) {
	fmt.Printf("\n   🗺️  Navigation using %s:\n", nav.currentRouteStrategy.GetName())
	routeDescription := nav.currentRouteStrategy.CalculateRoute(fromLocation, toLocation)
	fmt.Printf("   %s\n", routeDescription)
	fmt.Printf("   ⏱️  Estimated time: %d minutes\n", nav.currentRouteStrategy.GetEstimatedTime())
}

// ============================================================
// KEY INTERVIEW POINTS
// ============================================================
//
// Q: How is Strategy different from Factory?
// A: Factory is about CREATING objects
//    Strategy is about BEHAVIOR/ALGORITHMS
//    Often used together: Factory creates the right Strategy
//
// Q: When to use Strategy vs if-else?
// A: Strategy when:
//    - Algorithms are complex
//    - Need to switch at runtime
//    - Want to add new algorithms easily (OCP)
//    Use if-else for simple, few cases
//
// Q: Can you have multiple strategies in one context?
// A: Yes! E.g., ShoppingCart could have:
//    - PaymentStrategy
//    - DiscountStrategy
//    - ShippingStrategy
//
// Q: What's the relationship to DIP?
// A: Strategy IS Dependency Inversion!
//    Context depends on Strategy interface (abstraction)
//    not concrete implementations
//
// ❌ Common Mistakes:
// 1. Using Strategy for trivial algorithms (over-engineering)
// 2. Creating strategies with too many methods (should be focused)
// 3. Not using constructor/setter injection
// 4. Strategy that maintains too much state

func main() {
	fmt.Println("=== Strategy Pattern Demo ===\n")

	// ============================================================
	// DEMO 1: Sorting Strategies
	// ============================================================
	fmt.Println("--- Sorting Strategies ---")
	unsortedNumbers := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Printf("Original: %v\n\n", unsortedNumbers)

	// Create a sorter with Bubble Sort strategy
	sorter := NewSorter(&BubbleSort{})
	bubbleSortResult := sorter.Sort(unsortedNumbers)
	fmt.Printf("Sorted: %v\n\n", bubbleSortResult)

	// Change strategy at runtime to Quick Sort
	sorter.SetStrategy(&QuickSort{})
	quickSortResult := sorter.Sort(unsortedNumbers)
	fmt.Printf("Sorted: %v\n\n", quickSortResult)

	// ============================================================
	// DEMO 2: Payment Strategies
	// ============================================================
	fmt.Println("--- Payment Strategies ---")
	cart := NewShoppingCart()
	cart.AddItem("Laptop", 999.99)
	cart.AddItem("Mouse", 29.99)

	// Pay with Credit Card
	creditCardPayment := NewCreditCardPayment("4111111111111234", "123", "12/25", "John Doe")
	cart.SetPaymentMethod(creditCardPayment)
	if err := cart.Checkout(); err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	}

	// Same cart, different payment method - PayPal
	paypalPayment := NewPayPalPayment("user@example.com", "password")
	cart.SetPaymentMethod(paypalPayment)
	if err := cart.Checkout(); err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	}

	// Pay with UPI
	upiPayment := NewUPIPayment("user@okbank")
	cart.SetPaymentMethod(upiPayment)
	if err := cart.Checkout(); err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	}

	// ============================================================
	// DEMO 3: Compression Strategies
	// ============================================================
	fmt.Println("\n--- Compression Strategies ---")
	fileData := []byte("Hello, World!")

	// Compress using ZIP
	compressor := NewFileCompressor(&ZIPCompression{})
	zipFilename, zipCompressed := compressor.CompressFile("data.txt", fileData)
	fmt.Printf("   Compressed to: %s, Content: %s\n\n", zipFilename, zipCompressed)

	// Change strategy to GZIP
	compressor.SetStrategy(&GZIPCompression{})
	gzipFilename, gzipCompressed := compressor.CompressFile("data.txt", fileData)
	fmt.Printf("   Compressed to: %s, Content: %s\n", gzipFilename, gzipCompressed)

	// ============================================================
	// DEMO 4: Navigation Strategies
	// ============================================================
	fmt.Println("\n--- Navigation Strategies ---")

	// Start with fastest route
	navigator := NewNavigator(&FastestRoute{})
	navigator.Navigate("Home", "Office")

	// Change to scenic route
	fmt.Println()
	navigator.SetStrategy(&ScenicRoute{})
	navigator.Navigate("Home", "Office")

	// Change to shortest route
	fmt.Println()
	navigator.SetStrategy(&ShortestRoute{})
	navigator.Navigate("Home", "Office")
}
