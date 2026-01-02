package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================
// OBSERVER PATTERN - A Beginner-Friendly Guide
// ============================================================
//
// WHAT IS IT?
// The Observer Pattern creates a subscription mechanism where multiple
// objects (observers) can listen to events from another object (subject).
// When the subject's state changes, all observers are notified automatically.
//
// SIMPLE ANALOGY:
// Think of a YouTube channel. When you subscribe to a channel, you get
// notified whenever a new video is uploaded. You don't need to keep
// checking - the notification comes to you automatically!
//
// KEY COMPONENTS:
// 1. Subject (Publisher)  - The object being watched (e.g., YouTube channel)
// 2. Observer (Subscriber) - Objects that want to be notified (e.g., viewers)
// 3. Register/Subscribe   - Add an observer to the notification list
// 4. Unregister/Unsubscribe - Remove an observer from the list
// 5. Notify               - Tell all observers about a change
//
// WHEN TO USE:
// - When changes in one object should trigger updates in others
// - When you don't know beforehand how many objects need updates
// - For event-driven systems (button clicks, price changes, etc.)
// - For loose coupling between components
//
// REAL WORLD EXAMPLES:
// - YouTube/Newsletter subscriptions
// - Stock price alerts
// - Social media notifications
// - GUI event listeners (button clicks)
// - Message queues and event buses

// ============================================================
// EXAMPLE 1: Stock Price Alert System
// A simple example where investors get notified of stock price changes
// ============================================================

// ----- STEP 1: Define the Observer Interface -----
// This is what any "watcher" must implement to receive updates

// StockObserver defines what methods an observer must have
// Any type that wants to receive stock updates must implement this
type StockObserver interface {
	// OnPriceUpdate is called when the stock price changes
	OnPriceUpdate(stockSymbol string, newPrice float64)
	// GetObserverID returns a unique identifier for this observer
	GetObserverID() string
}

// ----- STEP 2: Define the Subject Interface -----
// This is what any "watchable" object must implement

// StockSubject defines what methods a subject must have
// Any type that can be observed must implement this
type StockSubject interface {
	// AddObserver registers a new observer to receive updates
	AddObserver(observer StockObserver)
	// RemoveObserver unregisters an observer from updates
	RemoveObserver(observer StockObserver)
	// NotifyObservers sends updates to all registered observers
	NotifyObservers()
}

// ----- STEP 3: Implement the Concrete Subject (Stock) -----

// Stock represents a tradeable stock that notifies observers of price changes
type Stock struct {
	symbol    string                   // Stock ticker symbol (e.g., "AAPL")
	price     float64                  // Current stock price
	observers map[string]StockObserver // Map of observer ID -> observer
	mutex     sync.RWMutex             // Protects concurrent access to observers
}

// NewStock creates a new Stock with the given symbol and starting price
func NewStock(symbol string, startingPrice float64) *Stock {
	return &Stock{
		symbol:    symbol,
		price:     startingPrice,
		observers: make(map[string]StockObserver),
	}
}

// AddObserver registers an observer to receive price updates
func (stock *Stock) AddObserver(observer StockObserver) {
	stock.mutex.Lock()
	defer stock.mutex.Unlock()

	observerID := observer.GetObserverID()
	stock.observers[observerID] = observer
	fmt.Printf("📋 [%s] Observer '%s' registered for updates\n", stock.symbol, observerID)
}

// RemoveObserver unregisters an observer from receiving updates
func (stock *Stock) RemoveObserver(observer StockObserver) {
	stock.mutex.Lock()
	defer stock.mutex.Unlock()

	observerID := observer.GetObserverID()
	delete(stock.observers, observerID)
	fmt.Printf("❌ [%s] Observer '%s' unregistered from updates\n", stock.symbol, observerID)
}

// NotifyObservers sends the current price to all registered observers
func (stock *Stock) NotifyObservers() {
	stock.mutex.RLock()
	defer stock.mutex.RUnlock()

	// Iterate through all observers and notify each one
	for _, observer := range stock.observers {
		observer.OnPriceUpdate(stock.symbol, stock.price)
	}
}

// UpdatePrice changes the stock price and notifies all observers
func (stock *Stock) UpdatePrice(newPrice float64) {
	// Store old price for logging
	stock.mutex.Lock()
	oldPrice := stock.price
	stock.price = newPrice
	stock.mutex.Unlock()

	fmt.Printf("\n📈 [%s] Price changed: $%.2f → $%.2f\n", stock.symbol, oldPrice, newPrice)

	// Notify all observers about the price change
	stock.NotifyObservers()
}

// GetCurrentPrice returns the current stock price (thread-safe)
func (stock *Stock) GetCurrentPrice() float64 {
	stock.mutex.RLock()
	defer stock.mutex.RUnlock()
	return stock.price
}

// GetObserverCount returns how many observers are watching this stock
func (stock *Stock) GetObserverCount() int {
	stock.mutex.RLock()
	defer stock.mutex.RUnlock()
	return len(stock.observers)
}

// ----- STEP 4: Implement Concrete Observers -----

// Investor represents a human investor who watches stock prices
type Investor struct {
	investorID   string // Unique ID for this investor
	investorName string // Display name
}

// NewInvestor creates a new Investor with the given ID and name
func NewInvestor(id string, name string) *Investor {
	return &Investor{
		investorID:   id,
		investorName: name,
	}
}

// OnPriceUpdate is called when a watched stock's price changes
func (investor *Investor) OnPriceUpdate(stockSymbol string, newPrice float64) {
	fmt.Printf("  💰 [%s] %s received alert: %s is now $%.2f\n",
		investor.investorID, investor.investorName, stockSymbol, newPrice)
}

// GetObserverID returns this investor's unique identifier
func (investor *Investor) GetObserverID() string {
	return investor.investorID
}

// AutomatedTradingBot represents an automated system that trades based on price
type AutomatedTradingBot struct {
	botID         string  // Unique ID for this bot
	buyThreshold  float64 // Buy when price falls below this
	sellThreshold float64 // Sell when price rises above this
}

// NewAutomatedTradingBot creates a new trading bot with buy/sell thresholds
func NewAutomatedTradingBot(id string, buyBelow float64, sellAbove float64) *AutomatedTradingBot {
	return &AutomatedTradingBot{
		botID:         id,
		buyThreshold:  buyBelow,
		sellThreshold: sellAbove,
	}
}

// OnPriceUpdate analyzes price and generates trading signals
func (bot *AutomatedTradingBot) OnPriceUpdate(stockSymbol string, newPrice float64) {
	// Determine trading action based on price thresholds
	switch {
	case newPrice < bot.buyThreshold:
		fmt.Printf("  🤖 [%s] BUY SIGNAL: %s at $%.2f (below $%.2f threshold)\n",
			bot.botID, stockSymbol, newPrice, bot.buyThreshold)
	case newPrice > bot.sellThreshold:
		fmt.Printf("  🤖 [%s] SELL SIGNAL: %s at $%.2f (above $%.2f threshold)\n",
			bot.botID, stockSymbol, newPrice, bot.sellThreshold)
	default:
		fmt.Printf("  🤖 [%s] HOLD: %s at $%.2f (within range $%.2f-$%.2f)\n",
			bot.botID, stockSymbol, newPrice, bot.buyThreshold, bot.sellThreshold)
	}
}

// GetObserverID returns this bot's unique identifier
func (bot *AutomatedTradingBot) GetObserverID() string {
	return bot.botID
}

// ============================================================
// EXAMPLE 2: Generic Event Bus System
// A more flexible observer pattern using function handlers
// ============================================================

// EventType represents the type/category of an event
type EventType string

// Define common event types as constants for type safety
const (
	EventUserCreated  EventType = "user.created"
	EventUserDeleted  EventType = "user.deleted"
	EventOrderPlaced  EventType = "order.placed"
	EventOrderShipped EventType = "order.shipped"
)

// Event represents something that happened in the system
type Event struct {
	EventType EventType              // What type of event this is
	Data      map[string]interface{} // Flexible data payload
}

// EventHandlerFunc is a function that handles events
// Using a function type makes it easy to subscribe inline functions
type EventHandlerFunc func(event Event)

// EventBus is a central hub for publishing and subscribing to events
type EventBus struct {
	// Map of event type -> list of handler functions
	handlersByType map[EventType][]EventHandlerFunc
	mutex          sync.RWMutex
}

// NewEventBus creates a new EventBus
func NewEventBus() *EventBus {
	return &EventBus{
		handlersByType: make(map[EventType][]EventHandlerFunc),
	}
}

// Subscribe registers a handler function for a specific event type
func (bus *EventBus) Subscribe(eventType EventType, handler EventHandlerFunc) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	// Append the handler to the list for this event type
	bus.handlersByType[eventType] = append(bus.handlersByType[eventType], handler)
	fmt.Printf("📥 Subscribed new handler for '%s' events\n", eventType)
}

// Publish sends an event to all handlers subscribed to that event type
func (bus *EventBus) Publish(event Event) {
	bus.mutex.RLock()
	handlers := bus.handlersByType[event.EventType]
	bus.mutex.RUnlock()

	fmt.Printf("📤 Publishing '%s' event to %d handler(s)\n", event.EventType, len(handlers))

	// Create a WaitGroup to wait for all handlers to complete
	var waitGroup sync.WaitGroup

	// Notify all handlers concurrently (non-blocking)
	for _, handler := range handlers {
		waitGroup.Add(1)
		// Capture handler in closure to avoid race condition
		go func(h EventHandlerFunc) {
			defer waitGroup.Done()
			h(event)
		}(handler)
	}

	// Wait for all handlers to complete
	waitGroup.Wait()
}

// PublishAsync sends an event without waiting for handlers to complete
func (bus *EventBus) PublishAsync(event Event) {
	bus.mutex.RLock()
	handlers := bus.handlersByType[event.EventType]
	bus.mutex.RUnlock()

	// Fire and forget - don't wait for handlers
	for _, handler := range handlers {
		go func(h EventHandlerFunc) {
			h(event)
		}(handler)
	}
}

// ============================================================
// EXAMPLE 3: YouTube-like Subscription System
// A practical example showing channel subscriptions
// ============================================================

// ChannelSubscriber is anyone who can receive notifications from a channel
type ChannelSubscriber interface {
	ReceiveNotification(channelName string, videoTitle string)
	GetSubscriberName() string
}

// VideoChannel represents a content channel that uploads videos
type VideoChannel struct {
	channelName string                       // Name of the channel
	subscribers map[string]ChannelSubscriber // Subscriber name -> subscriber
	mutex       sync.RWMutex
}

// NewVideoChannel creates a new video channel
func NewVideoChannel(name string) *VideoChannel {
	return &VideoChannel{
		channelName: name,
		subscribers: make(map[string]ChannelSubscriber),
	}
}

// Subscribe adds a subscriber to this channel
func (channel *VideoChannel) Subscribe(subscriber ChannelSubscriber) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	subscriberName := subscriber.GetSubscriberName()
	channel.subscribers[subscriberName] = subscriber
	fmt.Printf("🔔 %s subscribed to '%s'\n", subscriberName, channel.channelName)
}

// Unsubscribe removes a subscriber from this channel
func (channel *VideoChannel) Unsubscribe(subscriber ChannelSubscriber) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	subscriberName := subscriber.GetSubscriberName()
	delete(channel.subscribers, subscriberName)
	fmt.Printf("🔕 %s unsubscribed from '%s'\n", subscriberName, channel.channelName)
}

// UploadVideo simulates uploading a video and notifying all subscribers
func (channel *VideoChannel) UploadVideo(videoTitle string) {
	fmt.Printf("\n📹 '%s' uploaded: \"%s\"\n", channel.channelName, videoTitle)
	channel.notifyAllSubscribers(videoTitle)
}

// notifyAllSubscribers sends notifications to all subscribers (private method)
func (channel *VideoChannel) notifyAllSubscribers(videoTitle string) {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()

	for _, subscriber := range channel.subscribers {
		subscriber.ReceiveNotification(channel.channelName, videoTitle)
	}
}

// GetSubscriberCount returns the number of subscribers
func (channel *VideoChannel) GetSubscriberCount() int {
	channel.mutex.RLock()
	defer channel.mutex.RUnlock()
	return len(channel.subscribers)
}

// User represents a viewer who can subscribe to channels
type User struct {
	username string
}

// NewUser creates a new user with the given username
func NewUser(username string) *User {
	return &User{username: username}
}

// ReceiveNotification is called when a subscribed channel uploads a video
func (user *User) ReceiveNotification(channelName string, videoTitle string) {
	fmt.Printf("  📬 %s: New video from '%s' - \"%s\"\n", user.username, channelName, videoTitle)
}

// GetSubscriberName returns the username
func (user *User) GetSubscriberName() string {
	return user.username
}

// ============================================================
// EXAMPLE 4: Newsletter System with Multiple Notification Types
// Shows how different observers can handle the same event differently
// ============================================================

// NewsletterSubscriber defines how to receive newsletter updates
type NewsletterSubscriber interface {
	OnNewsletterPublished(topic string, content string)
	GetSubscriberEmail() string
}

// NewsletterPublisher manages newsletter subscriptions and publishing
type NewsletterPublisher struct {
	newsletterTopic string                          // Topic of this newsletter
	subscribers     map[string]NewsletterSubscriber // Email -> subscriber
	mutex           sync.RWMutex
}

// NewNewsletterPublisher creates a new newsletter publisher
func NewNewsletterPublisher(topic string) *NewsletterPublisher {
	return &NewsletterPublisher{
		newsletterTopic: topic,
		subscribers:     make(map[string]NewsletterSubscriber),
	}
}

// AddSubscriber adds a subscriber to the newsletter
func (publisher *NewsletterPublisher) AddSubscriber(subscriber NewsletterSubscriber) {
	publisher.mutex.Lock()
	defer publisher.mutex.Unlock()

	email := subscriber.GetSubscriberEmail()
	publisher.subscribers[email] = subscriber
	fmt.Printf("✅ %s subscribed to '%s' newsletter\n", email, publisher.newsletterTopic)
}

// RemoveSubscriber removes a subscriber by email
func (publisher *NewsletterPublisher) RemoveSubscriber(email string) {
	publisher.mutex.Lock()
	defer publisher.mutex.Unlock()

	delete(publisher.subscribers, email)
	fmt.Printf("🚫 %s unsubscribed from '%s' newsletter\n", email, publisher.newsletterTopic)
}

// PublishNewsletter sends the newsletter to all subscribers
func (publisher *NewsletterPublisher) PublishNewsletter(content string) {
	publisher.mutex.RLock()
	defer publisher.mutex.RUnlock()

	fmt.Printf("\n📰 Publishing '%s' newsletter to %d subscriber(s)\n",
		publisher.newsletterTopic, len(publisher.subscribers))

	for _, subscriber := range publisher.subscribers {
		subscriber.OnNewsletterPublished(publisher.newsletterTopic, content)
	}
}

// EmailNewsletterSubscriber receives newsletters via email
type EmailNewsletterSubscriber struct {
	emailAddress string
}

// NewEmailNewsletterSubscriber creates an email subscriber
func NewEmailNewsletterSubscriber(email string) *EmailNewsletterSubscriber {
	return &EmailNewsletterSubscriber{emailAddress: email}
}

// OnNewsletterPublished handles receiving a newsletter via email
func (subscriber *EmailNewsletterSubscriber) OnNewsletterPublished(topic string, content string) {
	fmt.Printf("  📧 Email to %s: [%s] %s\n", subscriber.emailAddress, topic, content)
}

// GetSubscriberEmail returns the email address
func (subscriber *EmailNewsletterSubscriber) GetSubscriberEmail() string {
	return subscriber.emailAddress
}

// SlackNewsletterSubscriber receives newsletters via Slack
type SlackNewsletterSubscriber struct {
	emailAddress string // Used as unique identifier
	slackChannel string // Slack channel to post to
}

// NewSlackNewsletterSubscriber creates a Slack subscriber
func NewSlackNewsletterSubscriber(email string, channel string) *SlackNewsletterSubscriber {
	return &SlackNewsletterSubscriber{
		emailAddress: email,
		slackChannel: channel,
	}
}

// OnNewsletterPublished handles receiving a newsletter via Slack
func (subscriber *SlackNewsletterSubscriber) OnNewsletterPublished(topic string, content string) {
	fmt.Printf("  💬 Slack #%s: [%s] %s\n", subscriber.slackChannel, topic, content)
}

// GetSubscriberEmail returns the email address (used as ID)
func (subscriber *SlackNewsletterSubscriber) GetSubscriberEmail() string {
	return subscriber.emailAddress
}

// ============================================================
// KEY INTERVIEW POINTS
// ============================================================
//
// Q: What's the difference between Observer and Pub-Sub patterns?
// A: Observer: Subject directly knows and notifies its observers
//    Pub-Sub: Publishers and subscribers are decoupled via a message broker
//            (they don't know about each other)
//
// Q: How do you handle slow observers?
// A: 1. Run notifications in goroutines (async notification)
//    2. Use buffered channels with timeouts
//    3. Implement a message queue for reliability
//    4. Set notification timeouts
//
// Q: How do you prevent memory leaks?
// A: 1. Always call Unsubscribe/RemoveObserver when done
//    2. Use context.Context for cancellation
//    3. Implement weak references if available
//    4. Add cleanup methods that remove stale observers
//
// Q: What's Push vs Pull notification?
// A: Push: Subject sends data in the notification (our examples)
//    Pull: Subject notifies observers, they then fetch the data
//    Push is simpler; Pull gives observers more control over what they get
//
// Q: How can you implement this with Go channels?
// A: Create a channel for each observer, subject sends to all channels
//    This is more "Go-idiomatic" and handles concurrency naturally
//
// COMMON MISTAKES TO AVOID:
// 1. ❌ Not protecting observers map with mutex (causes race conditions)
// 2. ❌ Forgetting to unsubscribe (causes memory leaks)
// 3. ❌ Synchronous notifications blocking the subject
// 4. ❌ Observers modifying subject state during notification
// 5. ❌ Creating circular dependencies between observers

// ============================================================
// MAIN - Demonstrates all examples
// ============================================================

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║          OBSERVER PATTERN - DEMONSTRATION                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

	// ----- Example 1: Stock Price Alert System -----
	fmt.Println("\n┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│ Example 1: Stock Price Alert System                     │")
	fmt.Println("└─────────────────────────────────────────────────────────┘")

	// Create a stock to observe
	appleStock := NewStock("AAPL", 150.00)

	// Create different types of observers
	investor1 := NewInvestor("INV-001", "John Smith")
	investor2 := NewInvestor("INV-002", "Jane Doe")
	tradingBot := NewAutomatedTradingBot("BOT-001", 145.00, 160.00)

	// Register observers
	appleStock.AddObserver(investor1)
	appleStock.AddObserver(investor2)
	appleStock.AddObserver(tradingBot)

	fmt.Printf("\n👥 Total observers watching AAPL: %d\n", appleStock.GetObserverCount())

	// Simulate price changes - all observers get notified
	appleStock.UpdatePrice(148.00) // Bot: HOLD
	appleStock.UpdatePrice(142.00) // Bot: BUY (below 145)
	appleStock.UpdatePrice(165.00) // Bot: SELL (above 160)

	// Unregister one investor
	appleStock.RemoveObserver(investor1)
	fmt.Printf("\n👥 Observers after John unsubscribed: %d\n", appleStock.GetObserverCount())

	// John won't receive this update
	appleStock.UpdatePrice(155.00)

	// ----- Example 2: Event Bus System -----
	fmt.Println("\n┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│ Example 2: Event Bus System                             │")
	fmt.Println("└─────────────────────────────────────────────────────────┘")

	eventBus := NewEventBus()

	// Subscribe multiple handlers to the same event type
	eventBus.Subscribe(EventUserCreated, func(event Event) {
		email := event.Data["email"]
		fmt.Printf("  📧 Email Service: Sending welcome email to %v\n", email)
	})

	eventBus.Subscribe(EventUserCreated, func(event Event) {
		email := event.Data["email"]
		fmt.Printf("  📊 Analytics Service: Recording new user signup: %v\n", email)
	})

	eventBus.Subscribe(EventOrderPlaced, func(event Event) {
		orderID := event.Data["orderID"]
		fmt.Printf("  📦 Fulfillment Service: Processing order #%v\n", orderID)
	})

	// Publish events
	fmt.Println("\n🎯 Triggering user signup event:")
	eventBus.Publish(Event{
		EventType: EventUserCreated,
		Data:      map[string]interface{}{"email": "newuser@example.com", "name": "Bob"},
	})

	fmt.Println("\n🛒 Triggering order placed event:")
	eventBus.Publish(Event{
		EventType: EventOrderPlaced,
		Data:      map[string]interface{}{"orderID": "ORD-12345", "total": 99.99},
	})

	// ----- Example 3: YouTube Subscription System -----
	fmt.Println("\n┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│ Example 3: YouTube-like Subscription System             │")
	fmt.Println("└─────────────────────────────────────────────────────────┘")

	// Create channels
	techChannel := NewVideoChannel("TechReviews")
	gamingChannel := NewVideoChannel("GamersHub")

	// Create users
	alice := NewUser("Alice")
	bob := NewUser("Bob")
	charlie := NewUser("Charlie")

	// Users subscribe to channels
	techChannel.Subscribe(alice)
	techChannel.Subscribe(bob)
	gamingChannel.Subscribe(bob) // Bob subscribes to both
	gamingChannel.Subscribe(charlie)

	fmt.Printf("\n📊 Tech Channel subscribers: %d\n", techChannel.GetSubscriberCount())
	fmt.Printf("📊 Gaming Channel subscribers: %d\n", gamingChannel.GetSubscriberCount())

	// Upload videos - subscribers get notified
	techChannel.UploadVideo("iPhone 16 Pro Review")
	gamingChannel.UploadVideo("GTA 6 First Look")

	// Alice unsubscribes from tech
	techChannel.Unsubscribe(alice)

	// Alice won't receive this notification
	techChannel.UploadVideo("Best Laptops of 2026")

	// ----- Example 4: Newsletter System -----
	fmt.Println("\n┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│ Example 4: Newsletter System                            │")
	fmt.Println("└─────────────────────────────────────────────────────────┘")

	newsletter := NewNewsletterPublisher("Go Programming Weekly")

	// Add different types of subscribers
	emailSubscriber := NewEmailNewsletterSubscriber("developer@example.com")
	slackSubscriber := NewSlackNewsletterSubscriber("team@example.com", "dev-news")

	newsletter.AddSubscriber(emailSubscriber)
	newsletter.AddSubscriber(slackSubscriber)

	// Publish newsletter - each subscriber receives it their preferred way
	newsletter.PublishNewsletter("Go 1.23 Released with exciting new features!")

	// Small delay to ensure async operations complete
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              DEMONSTRATION COMPLETE                      ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
}
