package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================
// PUB-SUB / MESSAGE QUEUE SYSTEM - Low Level Design
// ============================================================
//
// What is Pub-Sub?
// - Publishers send messages to topics
// - Subscribers listen to topics and receive all messages
// - One message can be delivered to MULTIPLE subscribers (fan-out)
//
// What is a Message Queue?
// - Producers add messages to a queue
// - Consumers take messages from the queue
// - Each message is delivered to only ONE consumer (point-to-point)
//
// Design Patterns Used:
// - Observer Pattern: Subscribers observe topics for new messages
// - Strategy Pattern: Different subscriber types handle messages differently
// - Producer-Consumer: Queue-based message processing
//
// ============================================================

// ========== MESSAGE ==========
// Message represents a single unit of data in the pub-sub system.
// It contains the content (payload) and metadata (id, topic, timestamp, headers).

type Message struct {
	ID        string            // Unique identifier for the message
	Topic     string            // The topic this message belongs to
	Payload   interface{}       // The actual content (can be any type)
	Timestamp time.Time         // When the message was created
	Headers   map[string]string // Optional key-value metadata
}

// messageCounter is used to generate unique message IDs.
// We use atomic.Int64 for thread-safety when multiple goroutines create messages.
var messageCounter atomic.Int64

// NewMessage creates a new message with a unique ID and the current timestamp.
// Parameters:
//   - topic: The topic name this message is for
//   - payload: The actual message content
func NewMessage(topic string, payload interface{}) *Message {
	// Atomically increment counter to ensure unique IDs even with concurrent access
	newID := messageCounter.Add(1)

	return &Message{
		ID:        fmt.Sprintf("MSG-%d", newID),
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now(),
		Headers:   make(map[string]string),
	}
}

// SetHeader adds a custom header to the message.
// Headers are useful for passing metadata like priority, source, etc.
func (m *Message) SetHeader(key, value string) {
	m.Headers[key] = value
}

// GetHeader retrieves a header value by key.
// Returns empty string if the header doesn't exist.
func (m *Message) GetHeader(key string) string {
	return m.Headers[key]
}

// String returns a human-readable representation of the message.
func (m *Message) String() string {
	return fmt.Sprintf("[%s] %s: %v", m.ID, m.Topic, m.Payload)
}

// ========== SUBSCRIBER INTERFACE ==========
// Subscriber defines what any message receiver must implement.
// This allows us to create different types of subscribers (logger, email, etc.)

type Subscriber interface {
	// GetID returns the unique identifier of this subscriber
	GetID() string

	// OnMessage is called when a new message arrives
	// Each subscriber decides what to do with the message
	OnMessage(msg *Message)
}

// ========== BASE SUBSCRIBER ==========
// BaseSubscriber is a simple implementation of the Subscriber interface.
// It uses a handler function to process messages, making it flexible.

type BaseSubscriber struct {
	id      string         // Unique identifier for this subscriber
	handler func(*Message) // Custom function to handle incoming messages
}

// NewSubscriber creates a new subscriber with a custom message handler.
// Parameters:
//   - id: A unique identifier for this subscriber
//   - handler: A function that will be called for each received message
func NewSubscriber(id string, handler func(*Message)) *BaseSubscriber {
	return &BaseSubscriber{
		id:      id,
		handler: handler,
	}
}

// GetID returns the subscriber's unique identifier.
func (s *BaseSubscriber) GetID() string {
	return s.id
}

// OnMessage processes an incoming message using the handler function.
func (s *BaseSubscriber) OnMessage(msg *Message) {
	if s.handler != nil {
		s.handler(msg)
	}
}

// ========== TOPIC ==========
// Topic is a channel that publishers write to and subscribers read from.
// All subscribers to a topic receive ALL messages published to it.

type Topic struct {
	name        string                // Name of the topic (e.g., "orders", "payments")
	subscribers map[string]Subscriber // Map of subscriber ID to subscriber
	messages    []*Message            // History of all messages (for persistence)
	mutex       sync.RWMutex          // Protects concurrent access to subscribers and messages
}

// NewTopic creates a new topic with the given name.
func NewTopic(name string) *Topic {
	return &Topic{
		name:        name,
		subscribers: make(map[string]Subscriber),
		messages:    make([]*Message, 0),
	}
}

// GetName returns the topic's name.
func (t *Topic) GetName() string {
	return t.name
}

// Subscribe adds a subscriber to this topic.
// The subscriber will receive all future messages published to this topic.
func (t *Topic) Subscribe(subscriber Subscriber) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	subscriberID := subscriber.GetID()
	t.subscribers[subscriberID] = subscriber
}

// Unsubscribe removes a subscriber from this topic.
// The subscriber will no longer receive messages from this topic.
func (t *Topic) Unsubscribe(subscriberID string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	delete(t.subscribers, subscriberID)
}

// Publish sends a message to all subscribers of this topic.
// Messages are delivered asynchronously using goroutines.
func (t *Topic) Publish(msg *Message) {
	// Lock to safely read subscribers and store message
	t.mutex.Lock()
	t.messages = append(t.messages, msg)

	// Copy subscribers to a slice to avoid holding the lock during delivery
	// This prevents deadlocks if a subscriber tries to unsubscribe during delivery
	subscriberList := make([]Subscriber, 0, len(t.subscribers))
	for _, subscriber := range t.subscribers {
		subscriberList = append(subscriberList, subscriber)
	}
	t.mutex.Unlock()

	// Deliver message to each subscriber asynchronously
	// Using goroutines ensures fast publishers aren't blocked by slow subscribers
	for _, subscriber := range subscriberList {
		go subscriber.OnMessage(msg)
	}
}

// GetSubscriberCount returns the number of active subscribers.
func (t *Topic) GetSubscriberCount() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return len(t.subscribers)
}

// GetMessageCount returns the total number of messages published to this topic.
func (t *Topic) GetMessageCount() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return len(t.messages)
}

// ========== MESSAGE BROKER ==========
// MessageBroker is the central hub that manages topics and routes messages.
// Publishers and subscribers interact with the broker instead of topics directly.

type MessageBroker struct {
	topics map[string]*Topic // Map of topic name to topic
	mutex  sync.RWMutex      // Protects concurrent access to topics map
}

// NewMessageBroker creates a new message broker.
func NewMessageBroker() *MessageBroker {
	return &MessageBroker{
		topics: make(map[string]*Topic),
	}
}

// CreateTopic creates a new topic with the given name.
// If the topic already exists, it returns the existing topic.
func (b *MessageBroker) CreateTopic(name string) *Topic {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Return existing topic if it already exists
	if existingTopic, exists := b.topics[name]; exists {
		return existingTopic
	}

	// Create and store new topic
	newTopic := NewTopic(name)
	b.topics[name] = newTopic

	return newTopic
}

// GetTopic returns a topic by name, or nil if it doesn't exist.
func (b *MessageBroker) GetTopic(name string) *Topic {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return b.topics[name]
}

// DeleteTopic removes a topic from the broker.
// Warning: This will disconnect all subscribers from the topic.
func (b *MessageBroker) DeleteTopic(name string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	delete(b.topics, name)
}

// Publish sends a message to all subscribers of the specified topic.
// Returns the created message and an error if the topic doesn't exist.
func (b *MessageBroker) Publish(topicName string, payload interface{}) (*Message, error) {
	topic := b.GetTopic(topicName)
	if topic == nil {
		return nil, fmt.Errorf("topic not found: %s", topicName)
	}

	// Create message and publish to topic
	message := NewMessage(topicName, payload)
	topic.Publish(message)

	return message, nil
}

// Subscribe adds a subscriber to the specified topic.
// Returns an error if the topic doesn't exist.
func (b *MessageBroker) Subscribe(topicName string, subscriber Subscriber) error {
	topic := b.GetTopic(topicName)
	if topic == nil {
		return fmt.Errorf("topic not found: %s", topicName)
	}

	topic.Subscribe(subscriber)
	return nil
}

// Unsubscribe removes a subscriber from the specified topic.
// Returns an error if the topic doesn't exist.
func (b *MessageBroker) Unsubscribe(topicName string, subscriberID string) error {
	topic := b.GetTopic(topicName)
	if topic == nil {
		return fmt.Errorf("topic not found: %s", topicName)
	}

	topic.Unsubscribe(subscriberID)
	return nil
}

// ListTopics returns the names of all topics in the broker.
func (b *MessageBroker) ListTopics() []string {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	topicNames := make([]string, 0, len(b.topics))
	for name := range b.topics {
		topicNames = append(topicNames, name)
	}

	return topicNames
}

// ========== MESSAGE QUEUE (Point-to-Point) ==========
// Unlike Pub-Sub where messages go to ALL subscribers,
// a Queue delivers each message to only ONE consumer.
// This is useful for distributing work among multiple workers.

type MessageQueue struct {
	name     string        // Name of the queue
	messages chan *Message // Buffered channel for storing messages
	capacity int           // Maximum number of messages the queue can hold
}

// NewMessageQueue creates a new message queue with the specified capacity.
// Parameters:
//   - name: A name to identify this queue
//   - capacity: Maximum number of messages to buffer
func NewMessageQueue(name string, capacity int) *MessageQueue {
	return &MessageQueue{
		name:     name,
		messages: make(chan *Message, capacity),
		capacity: capacity,
	}
}

// Enqueue adds a message to the queue.
// Returns the created message.
// Note: This will block if the queue is full!
func (q *MessageQueue) Enqueue(payload interface{}) *Message {
	message := NewMessage(q.name, payload)
	q.messages <- message
	return message
}

// Dequeue removes and returns a message from the queue.
// Returns nil immediately if the queue is empty (non-blocking).
func (q *MessageQueue) Dequeue() *Message {
	select {
	case message := <-q.messages:
		return message
	default:
		// Queue is empty, return nil without blocking
		return nil
	}
}

// DequeueBlocking removes and returns a message from the queue.
// This will BLOCK until a message is available.
// Use this when you want consumers to wait for work.
func (q *MessageQueue) DequeueBlocking() *Message {
	return <-q.messages
}

// Size returns the current number of messages in the queue.
func (q *MessageQueue) Size() int {
	return len(q.messages)
}

// GetCapacity returns the maximum capacity of the queue.
func (q *MessageQueue) GetCapacity() int {
	return q.capacity
}

// ========== CONSUMER GROUP ==========
// ConsumerGroup allows multiple consumers to share work from a topic.
// Messages are distributed among consumers (like a queue) instead of
// being broadcast to all (like pub-sub).

type ConsumerGroup struct {
	id        string      // Unique identifier for this consumer group
	consumers []*Consumer // List of consumers in this group
	topic     *Topic      // The topic this group consumes from
	mutex     sync.Mutex  // Protects concurrent access to consumers list
}

// Consumer represents a single consumer within a consumer group.
type Consumer struct {
	ID      string         // Unique identifier for this consumer
	Handler func(*Message) // Function to process messages
}

// NewConsumerGroup creates a new consumer group for the specified topic.
func NewConsumerGroup(id string, topic *Topic) *ConsumerGroup {
	return &ConsumerGroup{
		id:        id,
		consumers: make([]*Consumer, 0),
		topic:     topic,
	}
}

// AddConsumer adds a new consumer to the group.
// Parameters:
//   - id: Unique identifier for this consumer
//   - handler: Function to process messages assigned to this consumer
func (cg *ConsumerGroup) AddConsumer(id string, handler func(*Message)) {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	newConsumer := &Consumer{
		ID:      id,
		Handler: handler,
	}
	cg.consumers = append(cg.consumers, newConsumer)
}

// GetConsumerCount returns the number of consumers in the group.
func (cg *ConsumerGroup) GetConsumerCount() int {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	return len(cg.consumers)
}

// ========== SPECIALIZED SUBSCRIBERS ==========
// These are concrete implementations of the Subscriber interface
// for common use cases.

// LoggingSubscriber logs all received messages to the console.
// Useful for debugging and monitoring.
type LoggingSubscriber struct {
	id string
}

// NewLoggingSubscriber creates a new logging subscriber.
func NewLoggingSubscriber(id string) *LoggingSubscriber {
	return &LoggingSubscriber{id: id}
}

// GetID returns the subscriber's unique identifier.
func (s *LoggingSubscriber) GetID() string {
	return s.id
}

// OnMessage logs the received message to the console.
func (s *LoggingSubscriber) OnMessage(msg *Message) {
	fmt.Printf("  📥 [%s] Received: %s\n", s.id, msg)
}

// EmailSubscriber simulates sending emails when messages are received.
// In a real system, this would integrate with an email service.
type EmailSubscriber struct {
	id    string
	email string
}

// NewEmailSubscriber creates a new email subscriber.
// Parameters:
//   - id: Unique identifier for this subscriber
//   - email: Email address to send notifications to
func NewEmailSubscriber(id string, email string) *EmailSubscriber {
	return &EmailSubscriber{
		id:    id,
		email: email,
	}
}

// GetID returns the subscriber's unique identifier.
func (s *EmailSubscriber) GetID() string {
	return s.id
}

// OnMessage simulates sending an email notification.
func (s *EmailSubscriber) OnMessage(msg *Message) {
	fmt.Printf("  📧 [%s] Sending email to %s about: %v\n", s.id, s.email, msg.Payload)
}

// ========== MAIN FUNCTION ==========
// Demonstrates the pub-sub and message queue system.

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("       📬 PUB-SUB MESSAGE SYSTEM DEMO")
	fmt.Println("═══════════════════════════════════════════")

	// Step 1: Create the message broker
	// The broker is the central hub for all pub-sub operations
	broker := NewMessageBroker()

	// Step 2: Create topics
	// Topics are channels that group related messages
	broker.CreateTopic("orders")
	broker.CreateTopic("payments")
	broker.CreateTopic("notifications")
	broker.CreateTopic("analytics")

	fmt.Println("\n📋 Created Topics:", broker.ListTopics())

	// Step 3: Set up subscribers
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("👥 Setting up subscribers...")

	// Create subscribers for the "orders" topic
	orderLogger := NewLoggingSubscriber("order-logger")
	inventoryService := NewSubscriber("inventory-service", func(msg *Message) {
		fmt.Printf("  📦 [inventory-service] Processing order: %v\n", msg.Payload)
	})

	broker.Subscribe("orders", orderLogger)
	broker.Subscribe("orders", inventoryService)

	// Create subscribers for the "payments" topic
	paymentLogger := NewLoggingSubscriber("payment-logger")
	accountingService := NewSubscriber("accounting-service", func(msg *Message) {
		fmt.Printf("  💰 [accounting-service] Recording payment: %v\n", msg.Payload)
	})

	broker.Subscribe("payments", paymentLogger)
	broker.Subscribe("payments", accountingService)

	// Create subscriber for the "notifications" topic
	emailNotifier := NewEmailSubscriber("email-notifier", "admin@example.com")
	broker.Subscribe("notifications", emailNotifier)

	// Create analytics subscriber that listens to multiple topics
	analyticsService := NewSubscriber("analytics-service", func(msg *Message) {
		fmt.Printf("  📊 [analytics-service] Tracking: %v\n", msg.Payload)
	})
	broker.Subscribe("analytics", analyticsService)
	broker.Subscribe("orders", analyticsService)   // Also track orders
	broker.Subscribe("payments", analyticsService) // Also track payments

	// Step 4: Publish messages
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("📤 Publishing messages...")

	// Publish an order event
	fmt.Println("\n1️⃣ Order Event:")
	broker.Publish("orders", map[string]interface{}{
		"order_id": "ORD-001",
		"customer": "John Doe",
		"total":    99.99,
	})
	time.Sleep(100 * time.Millisecond) // Wait for async delivery to complete

	// Publish a payment event
	fmt.Println("\n2️⃣ Payment Event:")
	broker.Publish("payments", map[string]interface{}{
		"payment_id": "PAY-001",
		"order_id":   "ORD-001",
		"amount":     99.99,
		"status":     "completed",
	})
	time.Sleep(100 * time.Millisecond)

	// Publish a notification event
	fmt.Println("\n3️⃣ Notification Event:")
	broker.Publish("notifications", "New user signup: jane@email.com")
	time.Sleep(100 * time.Millisecond)

	// Publish multiple orders
	fmt.Println("\n4️⃣ Multiple Orders:")
	for i := 2; i <= 4; i++ {
		broker.Publish("orders", map[string]interface{}{
			"order_id": fmt.Sprintf("ORD-00%d", i),
			"total":    float64(i) * 50.0,
		})
	}
	time.Sleep(200 * time.Millisecond)

	// Step 5: Demonstrate Message Queue (Point-to-Point)
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("📬 Point-to-Point Queue Demo...")

	// Create a task queue
	taskQueue := NewMessageQueue("tasks", 10)

	// Producer adds tasks to the queue
	taskQueue.Enqueue("Task 1: Send email")
	taskQueue.Enqueue("Task 2: Generate report")
	taskQueue.Enqueue("Task 3: Cleanup logs")

	fmt.Printf("Queue size: %d\n", taskQueue.Size())

	// Consumer processes tasks one by one
	// Each task is delivered to exactly ONE consumer
	for taskQueue.Size() > 0 {
		task := taskQueue.Dequeue()
		if task != nil {
			fmt.Printf("  ⚡ Processing: %v\n", task.Payload)
		}
	}

	// Step 6: Demonstrate unsubscribing
	fmt.Println("\n─────────────────────────────────────────")
	fmt.Println("🔕 Unsubscribing order-logger...")
	broker.Unsubscribe("orders", "order-logger")

	fmt.Println("\n5️⃣ Order after unsubscribe:")
	broker.Publish("orders", map[string]interface{}{
		"order_id": "ORD-005",
		"note":     "Logger won't receive this",
	})
	time.Sleep(100 * time.Millisecond)

	// Summary of design decisions
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  KEY DESIGN DECISIONS:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. Topic-based pub-sub (fan-out to all subscribers)")
	fmt.Println("  2. Message Queue for point-to-point (one consumer)")
	fmt.Println("  3. Async delivery via goroutines (non-blocking)")
	fmt.Println("  4. Subscriber interface for flexibility")
	fmt.Println("  5. Thread-safe operations using mutex/atomic")
	fmt.Println("═══════════════════════════════════════════")
}
