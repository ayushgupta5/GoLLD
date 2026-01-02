package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// SHOPPING CART SYSTEM - Low Level DESIGN
// ============================================================================
//
// This system demonstrates:
// - Strategy Pattern: Different discount strategies (percentage, flat, buy-x-get-y)
// - Factory Pattern: Creating orders from carts
// - Entity Modeling: Products, Cart Items, Shopping Cart, Orders
// - Thread-safe operations using mutex locks
// - Category-based tax calculation
//
// ============================================================================

// ============================================================================
// SECTION 1: PRODUCT CATEGORY ENUM
// ============================================================================

// ProductCategory represents different types of products with associated tax rates.
// Using iota for auto-incrementing values (Electronics=0, Clothing=1, etc.)
type ProductCategory int

const (
	CategoryElectronics ProductCategory = iota // 0 - Electronics (18% tax)
	CategoryClothing                           // 1 - Clothing (12% tax)
	CategoryBooks                              // 2 - Books (0% tax - exempt)
	CategoryGrocery                            // 3 - Grocery (5% tax)
)

// String returns a human-readable name for the product category.
func (category ProductCategory) String() string {
	names := [...]string{"Electronics", "Clothing", "Books", "Grocery"}
	if int(category) < len(names) {
		return names[category]
	}
	return "Unknown"
}

// TaxRate returns the tax rate for each product category.
// Different categories have different tax rates based on regulations.
func (category ProductCategory) TaxRate() float64 {
	taxRates := map[ProductCategory]float64{
		CategoryElectronics: 0.18, // 18% tax on electronics
		CategoryClothing:    0.12, // 12% tax on clothing
		CategoryBooks:       0.0,  // Books are tax-exempt
		CategoryGrocery:     0.05, // 5% tax on groceries
	}

	if rate, exists := taxRates[category]; exists {
		return rate
	}
	return 0.0
}

// ============================================================================
// SECTION 2: PRODUCT ENTITY
// ============================================================================

// Product represents an item available for purchase in the store.
type Product struct {
	id          string          // Unique identifier (e.g., "P001")
	name        string          // Display name (e.g., "iPhone 15 Pro")
	description string          // Detailed description of the product
	price       float64         // Price per unit in dollars
	category    ProductCategory // Category for tax calculation
	stockCount  int             // Number of units available
	mutex       sync.Mutex      // Protects concurrent access to stock
}

// NewProduct creates and initializes a new Product instance.
func NewProduct(id, name string, price float64, category ProductCategory, initialStock int) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: "",
		price:       price,
		category:    category,
		stockCount:  initialStock,
	}
}

// Getter methods for Product fields
func (product *Product) GetID() string                { return product.id }
func (product *Product) GetName() string              { return product.name }
func (product *Product) GetPrice() float64            { return product.price }
func (product *Product) GetCategory() ProductCategory { return product.category }

// GetStock returns the current stock count (thread-safe).
func (product *Product) GetStock() int {
	product.mutex.Lock()
	defer product.mutex.Unlock()
	return product.stockCount
}

// ReduceStock decreases the stock count by the specified quantity.
// Returns an error if there isn't enough stock available.
func (product *Product) ReduceStock(quantity int) error {
	product.mutex.Lock()
	defer product.mutex.Unlock()

	if quantity > product.stockCount {
		return fmt.Errorf("insufficient stock: requested %d, available %d", quantity, product.stockCount)
	}

	product.stockCount -= quantity
	return nil
}

// AddStock increases the stock count by the specified quantity.
// Used when restocking or when orders are cancelled.
func (product *Product) AddStock(quantity int) {
	product.mutex.Lock()
	defer product.mutex.Unlock()
	product.stockCount += quantity
}

// ============================================================================
// SECTION 3: CART ITEM ENTITY
// ============================================================================

// CartItem represents a product with a specific quantity in a shopping cart.
type CartItem struct {
	product  *Product // Reference to the product
	quantity int      // Number of units in the cart
}

// NewCartItem creates a new CartItem instance.
func NewCartItem(product *Product, quantity int) *CartItem {
	return &CartItem{
		product:  product,
		quantity: quantity,
	}
}

// GetProduct returns the product reference.
func (item *CartItem) GetProduct() *Product {
	return item.product
}

// GetQuantity returns the quantity of this item.
func (item *CartItem) GetQuantity() int {
	return item.quantity
}

// GetSubtotal calculates the price for this item (price × quantity).
// Tax is NOT included in the subtotal.
func (item *CartItem) GetSubtotal() float64 {
	return item.product.GetPrice() * float64(item.quantity)
}

// GetTax calculates the tax amount for this item.
// Tax = Subtotal × Category Tax Rate
func (item *CartItem) GetTax() float64 {
	return item.GetSubtotal() * item.product.GetCategory().TaxRate()
}

// ============================================================================
// SECTION 4: DISCOUNT STRATEGY PATTERN
// ============================================================================
//
// The Strategy Pattern allows us to define a family of algorithms (discount types),
// encapsulate each one, and make them interchangeable.
//
// Benefits:
// - Easy to add new discount types without modifying existing code
// - Each discount type has its own calculation logic
// - Discounts can be swapped at runtime
//

// DiscountStrategy is the interface that all discount types must implement.
type DiscountStrategy interface {
	// CalculateDiscount computes the discount amount based on the subtotal
	CalculateDiscount(subtotal float64) float64

	// GetDescription returns a human-readable description of the discount
	GetDescription() string
}

// ----------------------------------------------------------------------------
// Strategy 1: Percentage Discount (e.g., "10% OFF")
// ----------------------------------------------------------------------------

// PercentageDiscount applies a percentage-based discount.
// Example: 10% off on a $100 order = $10 discount
type PercentageDiscount struct {
	percentage float64 // Discount percentage (e.g., 10 for 10%)
	couponCode string  // Coupon code that activates this discount
}

// NewPercentageDiscount creates a new percentage-based discount.
func NewPercentageDiscount(couponCode string, percentage float64) *PercentageDiscount {
	return &PercentageDiscount{
		couponCode: couponCode,
		percentage: percentage,
	}
}

// CalculateDiscount returns the discount amount (subtotal × percentage / 100).
func (discount *PercentageDiscount) CalculateDiscount(subtotal float64) float64 {
	return subtotal * discount.percentage / 100
}

// GetDescription returns a readable description of this discount.
func (discount *PercentageDiscount) GetDescription() string {
	return fmt.Sprintf("%.0f%% OFF (Code: %s)", discount.percentage, discount.couponCode)
}

// ----------------------------------------------------------------------------
// Strategy 2: Flat Discount (e.g., "$20 OFF")
// ----------------------------------------------------------------------------

// FlatDiscount applies a fixed amount discount.
// Example: $20 off on any order (discount can't exceed subtotal)
type FlatDiscount struct {
	amount     float64 // Fixed discount amount in dollars
	couponCode string  // Coupon code that activates this discount
}

// NewFlatDiscount creates a new flat-amount discount.
func NewFlatDiscount(couponCode string, amount float64) *FlatDiscount {
	return &FlatDiscount{
		couponCode: couponCode,
		amount:     amount,
	}
}

// CalculateDiscount returns the discount amount.
// The discount cannot exceed the subtotal.
func (discount *FlatDiscount) CalculateDiscount(subtotal float64) float64 {
	if discount.amount > subtotal {
		return subtotal // Don't discount more than the subtotal
	}
	return discount.amount
}

// GetDescription returns a readable description of this discount.
func (discount *FlatDiscount) GetDescription() string {
	return fmt.Sprintf("$%.2f OFF (Code: %s)", discount.amount, discount.couponCode)
}

// ----------------------------------------------------------------------------
// Strategy 3: Buy X Get Y Free Discount
// ----------------------------------------------------------------------------

// BuyXGetYDiscount applies a "Buy X Get Y Free" discount for a specific product.
// Example: Buy 2 get 1 free - customer pays for 2, gets 3
type BuyXGetYDiscount struct {
	productID string // Product this discount applies to
	buyCount  int    // Number of items to buy
	freeCount int    // Number of free items
}

// NewBuyXGetYDiscount creates a new buy-x-get-y discount.
func NewBuyXGetYDiscount(productID string, buyCount, freeCount int) *BuyXGetYDiscount {
	return &BuyXGetYDiscount{
		productID: productID,
		buyCount:  buyCount,
		freeCount: freeCount,
	}
}

// CalculateDiscount returns the discount amount.
// Note: This is a simplified implementation. A full implementation would
// need access to the cart items to calculate the actual discount.
func (discount *BuyXGetYDiscount) CalculateDiscount(subtotal float64) float64 {
	// In a full implementation, this would:
	// 1. Find the product in the cart
	// 2. Calculate how many free items the customer gets
	// 3. Return the value of those free items
	return 0
}

// GetDescription returns a readable description of this discount.
func (discount *BuyXGetYDiscount) GetDescription() string {
	return fmt.Sprintf("Buy %d Get %d Free", discount.buyCount, discount.freeCount)
}

// ============================================================================
// SECTION 5: SHOPPING CART
// ============================================================================

// cartIDGenerator generates unique IDs for shopping carts (thread-safe).
type cartIDGenerator struct {
	counter int
	mutex   sync.Mutex
}

var cartIDGen = &cartIDGenerator{counter: 0}

// NextID generates the next unique cart ID.
func (gen *cartIDGenerator) NextID() string {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()
	gen.counter++
	return fmt.Sprintf("CART-%d", gen.counter)
}

// Cart represents a shopping cart that holds items before checkout.
type Cart struct {
	id              string               // Unique cart identifier
	userID          string               // ID of the user who owns this cart
	items           map[string]*CartItem // Map of productID -> CartItem
	appliedDiscount DiscountStrategy     // Currently applied discount (can be nil)
	mutex           sync.Mutex           // Protects concurrent access to cart
}

// NewCart creates a new empty shopping cart for a user.
func NewCart(userID string) *Cart {
	return &Cart{
		id:              cartIDGen.NextID(),
		userID:          userID,
		items:           make(map[string]*CartItem),
		appliedDiscount: nil,
	}
}

// GetID returns the cart's unique identifier.
func (cart *Cart) GetID() string {
	return cart.id
}

// AddItem adds a product to the cart with the specified quantity.
// If the product already exists in the cart, the quantity is increased.
func (cart *Cart) AddItem(product *Product, quantity int) error {
	cart.mutex.Lock()
	defer cart.mutex.Unlock()

	// Check if enough stock is available
	if product.GetStock() < quantity {
		return fmt.Errorf("insufficient stock for '%s': requested %d, available %d",
			product.GetName(), quantity, product.GetStock())
	}

	// If product already in cart, increase quantity; otherwise, add new item
	if existingItem, exists := cart.items[product.GetID()]; exists {
		existingItem.quantity += quantity
	} else {
		cart.items[product.GetID()] = NewCartItem(product, quantity)
	}

	fmt.Printf("  ✅ Added %d x %s to cart\n", quantity, product.GetName())
	return nil
}

// RemoveItem removes a product from the cart completely.
func (cart *Cart) RemoveItem(productID string) {
	cart.mutex.Lock()
	defer cart.mutex.Unlock()
	delete(cart.items, productID)
}

// UpdateQuantity changes the quantity of a product in the cart.
// If quantity is 0 or negative, the item is removed from the cart.
func (cart *Cart) UpdateQuantity(productID string, newQuantity int) error {
	cart.mutex.Lock()
	defer cart.mutex.Unlock()

	item, exists := cart.items[productID]
	if !exists {
		return fmt.Errorf("product '%s' is not in the cart", productID)
	}

	// Remove item if quantity is zero or negative
	if newQuantity <= 0 {
		delete(cart.items, productID)
		return nil
	}

	// Check stock availability
	if item.product.GetStock() < newQuantity {
		return fmt.Errorf("insufficient stock: requested %d, available %d",
			newQuantity, item.product.GetStock())
	}

	item.quantity = newQuantity
	return nil
}

// ApplyDiscount sets a discount strategy on the cart.
func (cart *Cart) ApplyDiscount(discount DiscountStrategy) {
	cart.mutex.Lock()
	defer cart.mutex.Unlock()
	cart.appliedDiscount = discount
	fmt.Printf("  🏷️  Discount applied: %s\n", discount.GetDescription())
}

// calculateSubtotalInternal computes subtotal without locking (used internally).
func (cart *Cart) calculateSubtotalInternal() float64 {
	var subtotal float64
	for _, item := range cart.items {
		subtotal += item.GetSubtotal()
	}
	return subtotal
}

// calculateTaxInternal computes tax without locking (used internally).
func (cart *Cart) calculateTaxInternal() float64 {
	var totalTax float64
	for _, item := range cart.items {
		totalTax += item.GetTax()
	}
	return totalTax
}

// GetSubtotal returns the total price of all items before tax and discount.
func (cart *Cart) GetSubtotal() float64 {
	cart.mutex.Lock()
	defer cart.mutex.Unlock()
	return cart.calculateSubtotalInternal()
}

// GetTax returns the total tax amount for all items.
func (cart *Cart) GetTax() float64 {
	cart.mutex.Lock()
	defer cart.mutex.Unlock()
	return cart.calculateTaxInternal()
}

// GetDiscount returns the discount amount based on the applied discount strategy.
func (cart *Cart) GetDiscount() float64 {
	cart.mutex.Lock()
	defer cart.mutex.Unlock()

	if cart.appliedDiscount == nil {
		return 0
	}
	subtotal := cart.calculateSubtotalInternal()
	return cart.appliedDiscount.CalculateDiscount(subtotal)
}

// GetTotal returns the final amount: Subtotal + Tax - Discount.
func (cart *Cart) GetTotal() float64 {
	cart.mutex.Lock()
	defer cart.mutex.Unlock()

	subtotal := cart.calculateSubtotalInternal()
	tax := cart.calculateTaxInternal()

	var discountAmount float64
	if cart.appliedDiscount != nil {
		discountAmount = cart.appliedDiscount.CalculateDiscount(subtotal)
	}

	return subtotal + tax - discountAmount
}

// GetItemCount returns the total number of items in the cart.
func (cart *Cart) GetItemCount() int {
	cart.mutex.Lock()
	defer cart.mutex.Unlock()

	totalCount := 0
	for _, item := range cart.items {
		totalCount += item.quantity
	}
	return totalCount
}

// IsEmpty checks if the cart has no items.
func (cart *Cart) IsEmpty() bool {
	cart.mutex.Lock()
	defer cart.mutex.Unlock()
	return len(cart.items) == 0
}

// PrintCart displays the cart contents in a formatted receipt-style layout.
func (cart *Cart) PrintCart() {
	cart.mutex.Lock()
	defer cart.mutex.Unlock()

	fmt.Println("\n╔════════════════════════════════════════════════╗")
	fmt.Println("║              🛒 SHOPPING CART                  ║")
	fmt.Println("╠════════════════════════════════════════════════╣")

	if len(cart.items) == 0 {
		fmt.Println("║  Your cart is empty                            ║")
	} else {
		for _, item := range cart.items {
			fmt.Printf("  %s x%d\n", item.product.GetName(), item.quantity)
			fmt.Printf("    $%.2f each = $%.2f (Tax: $%.2f)\n",
				item.product.GetPrice(), item.GetSubtotal(), item.GetTax())
		}
	}

	// Calculate totals (already holding lock, so use internal methods)
	subtotal := cart.calculateSubtotalInternal()
	tax := cart.calculateTaxInternal()

	var discountAmount float64
	if cart.appliedDiscount != nil {
		discountAmount = cart.appliedDiscount.CalculateDiscount(subtotal)
	}

	total := subtotal + tax - discountAmount

	fmt.Println("╠════════════════════════════════════════════════╣")
	fmt.Printf("  Subtotal: $%.2f\n", subtotal)
	fmt.Printf("  Tax:      $%.2f\n", tax)

	if cart.appliedDiscount != nil {
		fmt.Printf("  Discount: -$%.2f (%s)\n", discountAmount, cart.appliedDiscount.GetDescription())
	}

	fmt.Println("  ────────────────────────────────")
	fmt.Printf("  TOTAL:    $%.2f\n", total)
	fmt.Println("╚════════════════════════════════════════════════╝")
}

// ============================================================================
// SECTION 6: ORDER STATUS ENUM
// ============================================================================

// OrderStatus represents the lifecycle state of an order.
type OrderStatus int

const (
	OrderStatusPending   OrderStatus = iota // 0 - Order created, awaiting confirmation
	OrderStatusConfirmed                    // 1 - Order confirmed
	OrderStatusShipped                      // 2 - Order shipped to customer
	OrderStatusDelivered                    // 3 - Order delivered to customer
	OrderStatusCancelled                    // 4 - Order was cancelled
)

// String returns a human-readable name for the order status.
func (status OrderStatus) String() string {
	names := [...]string{"Pending", "Confirmed", "Shipped", "Delivered", "Cancelled"}
	if int(status) < len(names) {
		return names[status]
	}
	return "Unknown"
}

// ============================================================================
// SECTION 7: ORDER ENTITY
// ============================================================================

// orderIDGenerator generates unique IDs for orders (thread-safe).
type orderIDGenerator struct {
	counter int
	mutex   sync.Mutex
}

var orderIDGen = &orderIDGenerator{counter: 0}

// NextID generates the next unique order ID.
func (gen *orderIDGenerator) NextID() string {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()
	gen.counter++
	return fmt.Sprintf("ORD-%d", gen.counter)
}

// Order represents a confirmed purchase made from a shopping cart.
type Order struct {
	id              string      // Unique order identifier
	userID          string      // ID of the user who placed the order
	items           []*CartItem // List of items in the order
	subtotal        float64     // Total before tax and discount
	taxAmount       float64     // Total tax amount
	discountAmount  float64     // Discount applied
	totalAmount     float64     // Final amount charged
	status          OrderStatus // Current status of the order
	createdAt       time.Time   // When the order was placed
	shippingAddress string      // Delivery address
}

// NewOrderFromCart creates a new Order from a shopping cart.
// This is an example of the Factory Pattern - creating complex objects.
// The function also reduces inventory for each purchased item.
func NewOrderFromCart(cart *Cart, shippingAddress string) (*Order, error) {
	// Validate cart is not empty
	if cart.IsEmpty() {
		return nil, fmt.Errorf("cannot create order: cart is empty")
	}

	// Calculate totals before creating order
	subtotal := cart.GetSubtotal()
	taxAmount := cart.GetTax()
	discountAmount := cart.GetDiscount()
	totalAmount := cart.GetTotal()

	// Create the order
	order := &Order{
		id:              orderIDGen.NextID(),
		userID:          cart.userID,
		items:           make([]*CartItem, 0),
		subtotal:        subtotal,
		taxAmount:       taxAmount,
		discountAmount:  discountAmount,
		totalAmount:     totalAmount,
		status:          OrderStatusPending,
		createdAt:       time.Now(),
		shippingAddress: shippingAddress,
	}

	// Copy items from cart and reduce inventory
	// This is done in a single transaction to ensure consistency
	for _, item := range cart.items {
		// Attempt to reduce stock
		err := item.product.ReduceStock(item.quantity)
		if err != nil {
			// If stock reduction fails, we should ideally rollback
			// For simplicity, we just return an error here
			return nil, fmt.Errorf("failed to reserve '%s': %v", item.product.GetName(), err)
		}

		// Add item to order
		order.items = append(order.items, item)
	}

	return order, nil
}

// Getter methods for Order
func (order *Order) GetID() string          { return order.id }
func (order *Order) GetStatus() OrderStatus { return order.status }
func (order *Order) GetTotal() float64      { return order.totalAmount }

// Confirm changes the order status to Confirmed.
func (order *Order) Confirm() {
	order.status = OrderStatusConfirmed
}

// Ship changes the order status to Shipped.
func (order *Order) Ship() {
	order.status = OrderStatusShipped
}

// Deliver changes the order status to Delivered.
func (order *Order) Deliver() {
	order.status = OrderStatusDelivered
}

// Cancel changes the order status to Cancelled.
// In a full implementation, this would also restore the inventory.
func (order *Order) Cancel() {
	order.status = OrderStatusCancelled
	// TODO: Restore inventory for cancelled items
}

// PrintOrder displays the order details in a formatted confirmation layout.
func (order *Order) PrintOrder() {
	fmt.Printf(`
╔════════════════════════════════════════════════╗
║              📦 ORDER CONFIRMATION             ║
╠════════════════════════════════════════════════╣
  Order ID: %s
  Status: %s
  Date: %s
  
  Items:
`,
		order.id,
		order.status,
		order.createdAt.Format("Jan 02, 2006"))

	for _, item := range order.items {
		fmt.Printf("    • %s x%d = $%.2f\n",
			item.product.GetName(), item.quantity, item.GetSubtotal())
	}

	fmt.Printf(`
  ────────────────────────────────
  Subtotal: $%.2f
  Tax:      $%.2f
  Discount: -$%.2f
  TOTAL:    $%.2f
  
  Shipping to: %s
╚════════════════════════════════════════════════╝
`,
		order.subtotal,
		order.taxAmount,
		order.discountAmount,
		order.totalAmount,
		order.shippingAddress)
}

// ============================================================================
// SECTION 8: MAIN - DEMONSTRATION
// ============================================================================

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("        🛒 SHOPPING CART SYSTEM")
	fmt.Println("═══════════════════════════════════════════")

	// =========================================
	// STEP 1: Create product catalog
	// =========================================
	products := []*Product{
		NewProduct("P001", "iPhone 15 Pro", 999.00, CategoryElectronics, 10),
		NewProduct("P002", "MacBook Air M3", 1299.00, CategoryElectronics, 5),
		NewProduct("P003", "Cotton T-Shirt", 29.99, CategoryClothing, 100),
		NewProduct("P004", "Go Programming Book", 49.99, CategoryBooks, 50),
		NewProduct("P005", "Organic Coffee", 15.99, CategoryGrocery, 200),
	}

	// Display available products
	fmt.Println("\n📦 Available Products:")
	fmt.Println("─────────────────────────────────────────")
	for _, product := range products {
		fmt.Printf("  %s: %s - $%.2f (Stock: %d) [%s]\n",
			product.GetID(),
			product.GetName(),
			product.GetPrice(),
			product.GetStock(),
			product.GetCategory())
	}

	// =========================================
	// STEP 2: Create a shopping cart
	// =========================================
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("🛒 Adding items to cart...")

	shoppingCart := NewCart("USER001")

	// Add items to cart
	shoppingCart.AddItem(products[0], 1) // 1 iPhone
	shoppingCart.AddItem(products[2], 2) // 2 T-Shirts
	shoppingCart.AddItem(products[3], 1) // 1 Book
	shoppingCart.AddItem(products[4], 3) // 3 Coffee

	// Display cart contents
	shoppingCart.PrintCart()

	// =========================================
	// STEP 3: Apply a discount code
	// =========================================
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("🏷️  Applying coupon code...")

	// Create a 10% discount using the Strategy Pattern
	percentageDiscount := NewPercentageDiscount("SAVE10", 10)
	shoppingCart.ApplyDiscount(percentageDiscount)

	// Display cart with discount applied
	shoppingCart.PrintCart()

	// =========================================
	// STEP 4: Create order from cart
	// =========================================
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("📦 Creating order...")

	order, err := NewOrderFromCart(shoppingCart, "123 Main St, New York, NY 10001")
	if err != nil {
		fmt.Printf("❌ Error creating order: %v\n", err)
		return
	}

	// Confirm the order
	order.Confirm()

	// Display order confirmation
	order.PrintOrder()

	// =========================================
	// STEP 5: Show updated inventory
	// =========================================
	fmt.Println("\n📦 Updated Inventory:")
	fmt.Println("─────────────────────────────────────────")
	for _, product := range products {
		fmt.Printf("  %s: %d in stock\n", product.GetName(), product.GetStock())
	}

	// =========================================
	// SUMMARY: Key Design Decisions
	// =========================================
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  KEY DESIGN DECISIONS:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. Strategy Pattern for flexible discount types")
	fmt.Println("  2. Category-based tax rates (18%, 12%, 5%, 0%)")
	fmt.Println("  3. Inventory reduced on order creation")
	fmt.Println("  4. Factory Pattern: Cart → Order conversion")
	fmt.Println("  5. Thread-safe operations using mutex locks")
	fmt.Println("  6. Clear separation of entities and logic")
	fmt.Println("═══════════════════════════════════════════")
}
