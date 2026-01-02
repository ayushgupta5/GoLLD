package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ============================================================
// NOTIFICATION SYSTEM - Low Level Design
// ============================================================
//
// This system demonstrates how to build a scalable notification
// service that can send messages through multiple channels
// (Email, SMS, Push, Slack).
//
// Design Patterns Used:
// 1. Strategy Pattern - Different notification channels
// 2. Decorator Pattern - Add retry/logging capabilities
// 3. Template Pattern - Reusable notification templates
//
// ============================================================

// ==================== ENUMS (Type Definitions) ====================
//
// In Go, we simulate enums using custom types with constants.
// This provides type safety and better code readability.

// NotificationType represents the channel through which
// a notification will be sent (Email, SMS, Push, or Slack)
type NotificationType int

const (
	NotificationTypeEmail NotificationType = iota // 0 - Email notifications
	NotificationTypeSMS                           // 1 - SMS text messages
	NotificationTypePush                          // 2 - Mobile push notifications
	NotificationTypeSlack                         // 3 - Slack messages
)

// String converts NotificationType to a readable string
func (notificationType NotificationType) String() string {
	typeNames := []string{"Email", "SMS", "Push", "Slack"}
	if int(notificationType) < len(typeNames) {
		return typeNames[notificationType]
	}
	return "Unknown"
}

// NotificationPriority determines how urgent a notification is
type NotificationPriority int

const (
	PriorityLow      NotificationPriority = iota // 0 - Can be delayed
	PriorityMedium                               // 1 - Normal priority
	PriorityHigh                                 // 2 - Should be sent soon
	PriorityCritical                             // 3 - Must be sent immediately (ignores quiet hours)
)

// String converts NotificationPriority to a readable string
func (priority NotificationPriority) String() string {
	priorityNames := []string{"Low", "Medium", "High", "Critical"}
	if int(priority) < len(priorityNames) {
		return priorityNames[priority]
	}
	return "Unknown"
}

// NotificationStatus tracks the current state of a notification
type NotificationStatus int

const (
	StatusPending  NotificationStatus = iota // 0 - Waiting to be sent
	StatusSent                               // 1 - Successfully delivered
	StatusFailed                             // 2 - Failed to send
	StatusRetrying                           // 3 - Retrying after failure
)

// String converts NotificationStatus to a readable string
func (status NotificationStatus) String() string {
	statusNames := []string{"Pending", "Sent", "Failed", "Retrying"}
	if int(status) < len(statusNames) {
		return statusNames[status]
	}
	return "Unknown"
}

// ==================== NOTIFICATION MODEL ====================
//
// Notification holds all the information needed to send a message
// to a user through a specific channel.

type Notification struct {
	ID         string               // Unique identifier for this notification
	UserID     string               // Target user who will receive this notification
	Title      string               // Subject/Title of the notification
	Message    string               // Body content of the notification
	Channel    NotificationType     // Which channel to use (Email, SMS, etc.)
	Priority   NotificationPriority // How urgent is this notification
	Status     NotificationStatus   // Current delivery status
	CreatedAt  time.Time            // When was this notification created
	SentAt     time.Time            // When was this notification actually sent
	RetryCount int                  // How many times we've tried to send this
	Metadata   map[string]string    // Additional data (e.g., tracking info)
}

// notificationIDCounter generates unique IDs for notifications
// Note: In production, use UUID or database-generated IDs
var notificationIDCounter int
var notificationIDMutex sync.Mutex // Protects counter in concurrent access

// NewNotification creates a new notification with sensible defaults
func NewNotification(
	userID string,
	title string,
	message string,
	channel NotificationType,
	priority NotificationPriority,
) *Notification {
	// Thread-safe ID generation
	notificationIDMutex.Lock()
	notificationIDCounter++
	id := fmt.Sprintf("NOTIF-%d", notificationIDCounter)
	notificationIDMutex.Unlock()

	return &Notification{
		ID:        id,
		UserID:    userID,
		Title:     title,
		Message:   message,
		Channel:   channel,
		Priority:  priority,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]string),
	}
}

// ==================== NOTIFICATION CHANNEL INTERFACE ====================
//
// Strategy Pattern: Define a common interface for all notification channels.
// This allows the service to work with any channel without knowing
// the implementation details.

type NotificationChannel interface {
	// Send delivers the notification and returns any error
	Send(notification *Notification) error
	// GetType returns the type of this channel
	GetType() NotificationType
}

// ==================== EMAIL CHANNEL ====================

// EmailChannel handles sending email notifications
type EmailChannel struct {
	SMTPHost string // Email server hostname
	SMTPPort int    // Email server port
	FromAddr string // Sender email address
}

// NewEmailChannel creates a new email channel with SMTP configuration
func NewEmailChannel(host string, port int, fromAddress string) *EmailChannel {
	return &EmailChannel{
		SMTPHost: host,
		SMTPPort: port,
		FromAddr: fromAddress,
	}
}

// Send delivers an email notification
func (emailChannel *EmailChannel) Send(notification *Notification) error {
	// In a real implementation, this would connect to SMTP server
	// and send the actual email. Here we simulate the send.
	fmt.Printf("  📧 EMAIL to %s\n", notification.UserID)
	fmt.Printf("     Subject: %s\n", notification.Title)
	fmt.Printf("     Body: %s\n", notification.Message)
	return nil
}

// GetType returns the channel type (Email)
func (emailChannel *EmailChannel) GetType() NotificationType {
	return NotificationTypeEmail
}

// ==================== SMS CHANNEL ====================

// SMSChannel handles sending SMS text messages
type SMSChannel struct {
	Provider string // SMS provider name (e.g., "twilio")
	APIKey   string // API key for authentication
}

// NewSMSChannel creates a new SMS channel
func NewSMSChannel(provider string, apiKey string) *SMSChannel {
	return &SMSChannel{
		Provider: provider,
		APIKey:   apiKey,
	}
}

// Send delivers an SMS notification
func (smsChannel *SMSChannel) Send(notification *Notification) error {
	// In a real implementation, this would call the SMS provider's API
	fmt.Printf("  📱 SMS to %s: %s\n", notification.UserID, notification.Message)
	return nil
}

// GetType returns the channel type (SMS)
func (smsChannel *SMSChannel) GetType() NotificationType {
	return NotificationTypeSMS
}

// ==================== PUSH NOTIFICATION CHANNEL ====================

// PushChannel handles sending mobile push notifications
type PushChannel struct {
	FCMKey string // Firebase Cloud Messaging API key
}

// NewPushChannel creates a new push notification channel
func NewPushChannel(fcmKey string) *PushChannel {
	return &PushChannel{FCMKey: fcmKey}
}

// Send delivers a push notification
func (pushChannel *PushChannel) Send(notification *Notification) error {
	// In a real implementation, this would call FCM or APNS
	fmt.Printf("  🔔 PUSH to %s: %s - %s\n",
		notification.UserID,
		notification.Title,
		notification.Message,
	)
	return nil
}

// GetType returns the channel type (Push)
func (pushChannel *PushChannel) GetType() NotificationType {
	return NotificationTypePush
}

// ==================== SLACK CHANNEL ====================

// SlackChannel handles sending Slack messages
type SlackChannel struct {
	WebhookURL string // Slack incoming webhook URL
}

// NewSlackChannel creates a new Slack channel
func NewSlackChannel(webhookURL string) *SlackChannel {
	return &SlackChannel{WebhookURL: webhookURL}
}

// Send delivers a Slack notification
func (slackChannel *SlackChannel) Send(notification *Notification) error {
	// In a real implementation, this would POST to the webhook URL
	fmt.Printf("  💬 SLACK: [%s] %s\n", notification.Title, notification.Message)
	return nil
}

// GetType returns the channel type (Slack)
func (slackChannel *SlackChannel) GetType() NotificationType {
	return NotificationTypeSlack
}

// ==================== CHANNEL DECORATORS ====================
//
// Decorator Pattern: Wrap channels to add extra functionality
// like retry logic or logging, without modifying the original channel.

// RetryDecorator wraps a channel to add automatic retry on failure
type RetryDecorator struct {
	wrappedChannel NotificationChannel // The channel being decorated
	maxRetries     int                 // Maximum number of retry attempts
	retryDelay     time.Duration       // Time to wait between retries
}

// NewRetryDecorator creates a decorator that adds retry capability
func NewRetryDecorator(
	channel NotificationChannel,
	maxRetries int,
	retryDelay time.Duration,
) *RetryDecorator {
	return &RetryDecorator{
		wrappedChannel: channel,
		maxRetries:     maxRetries,
		retryDelay:     retryDelay,
	}
}

// Send attempts to deliver the notification with retries on failure
func (decorator *RetryDecorator) Send(notification *Notification) error {
	var lastError error

	// Try sending up to (maxRetries + 1) times
	for attempt := 0; attempt <= decorator.maxRetries; attempt++ {
		// If this is a retry, wait before trying again
		if attempt > 0 {
			notification.Status = StatusRetrying
			fmt.Printf("     ⟳ Retry attempt %d/%d...\n", attempt, decorator.maxRetries)
			time.Sleep(decorator.retryDelay)
		}

		// Attempt to send
		lastError = decorator.wrappedChannel.Send(notification)
		if lastError == nil {
			// Success! No need to retry
			return nil
		}

		// Track the retry count
		notification.RetryCount++
	}

	// All retries exhausted, return the last error
	return lastError
}

// GetType returns the wrapped channel's type
func (decorator *RetryDecorator) GetType() NotificationType {
	return decorator.wrappedChannel.GetType()
}

// LoggingDecorator wraps a channel to add logging
type LoggingDecorator struct {
	wrappedChannel NotificationChannel
}

// NewLoggingDecorator creates a decorator that adds logging
func NewLoggingDecorator(channel NotificationChannel) *LoggingDecorator {
	return &LoggingDecorator{wrappedChannel: channel}
}

// Send logs the attempt and result of sending a notification
func (decorator *LoggingDecorator) Send(notification *Notification) error {
	// Log before sending
	fmt.Printf("  [LOG] Sending %s notification %s\n",
		notification.Channel,
		notification.ID,
	)

	// Send the notification
	err := decorator.wrappedChannel.Send(notification)

	// Log the result
	if err != nil {
		fmt.Printf("  [LOG] ✗ Failed: %v\n", err)
	} else {
		fmt.Printf("  [LOG] ✓ Success: %s sent\n", notification.ID)
	}

	return err
}

// GetType returns the wrapped channel's type
func (decorator *LoggingDecorator) GetType() NotificationType {
	return decorator.wrappedChannel.GetType()
}

// ==================== USER PREFERENCES ====================
//
// UserPreferences stores user-specific notification settings
// like which channels they want to receive, quiet hours, etc.

type UserPreferences struct {
	UserID          string                    // User identifier
	EnabledChannels map[NotificationType]bool // Which channels user has enabled
	Email           string                    // User's email address
	Phone           string                    // User's phone number
	PushToken       string                    // User's device push token
	QuietHoursStart int                       // Start of quiet hours (0-23)
	QuietHoursEnd   int                       // End of quiet hours (0-23)
}

// NewUserPreferences creates preferences with default settings
func NewUserPreferences(userID string) *UserPreferences {
	return &UserPreferences{
		UserID: userID,
		// Default: Email and Push enabled, SMS disabled
		EnabledChannels: map[NotificationType]bool{
			NotificationTypeEmail: true,
			NotificationTypeSMS:   false,
			NotificationTypePush:  true,
			NotificationTypeSlack: true,
		},
		QuietHoursStart: 0, // No quiet hours by default
		QuietHoursEnd:   0,
	}
}

// IsChannelEnabled checks if user has enabled a specific channel
func (prefs *UserPreferences) IsChannelEnabled(channel NotificationType) bool {
	enabled, exists := prefs.EnabledChannels[channel]
	return exists && enabled
}

// IsQuietHours checks if current time is within user's quiet hours
// During quiet hours, only Critical notifications are sent
func (prefs *UserPreferences) IsQuietHours() bool {
	// If start and end are same, quiet hours are disabled
	if prefs.QuietHoursStart == prefs.QuietHoursEnd {
		return false
	}

	currentHour := time.Now().Hour()

	// Normal case: quiet hours don't span midnight (e.g., 9-17)
	if prefs.QuietHoursStart < prefs.QuietHoursEnd {
		return currentHour >= prefs.QuietHoursStart && currentHour < prefs.QuietHoursEnd
	}

	// Overnight case: quiet hours span midnight (e.g., 22-7)
	return currentHour >= prefs.QuietHoursStart || currentHour < prefs.QuietHoursEnd
}

// ==================== NOTIFICATION TEMPLATE ====================
//
// Templates allow reusing notification content with placeholders
// that get filled in at send time.

type NotificationTemplate struct {
	ID          string           // Unique template identifier
	Name        string           // Human-readable template name
	TitleFormat string           // Title with {placeholders}
	BodyFormat  string           // Body with {placeholders}
	Channel     NotificationType // Default channel for this template
}

// NewTemplate creates a new notification template
func NewTemplate(
	id string,
	name string,
	titleFormat string,
	bodyFormat string,
	channel NotificationType,
) *NotificationTemplate {
	return &NotificationTemplate{
		ID:          id,
		Name:        name,
		TitleFormat: titleFormat,
		BodyFormat:  bodyFormat,
		Channel:     channel,
	}
}

// Render fills in the template placeholders with actual values
// Parameters should be a map like {"name": "John", "order_id": "12345"}
func (template *NotificationTemplate) Render(parameters map[string]string) (title string, body string) {
	title = template.TitleFormat
	body = template.BodyFormat

	// Replace each placeholder with its value
	for key, value := range parameters {
		placeholder := "{" + key + "}"
		title = strings.ReplaceAll(title, placeholder, value)
		body = strings.ReplaceAll(body, placeholder, value)
	}

	return title, body
}

// ==================== NOTIFICATION SERVICE ====================
//
// The main service that coordinates all notification operations.
// It manages channels, user preferences, templates, and queuing.

type NotificationService struct {
	channels          map[NotificationType]NotificationChannel // Registered channels
	userPreferences   map[string]*UserPreferences              // User settings by userID
	templates         map[string]*NotificationTemplate         // Templates by ID
	notificationQueue chan *Notification                       // Async processing queue
	history           []*Notification                          // Sent notification history
	mutex             sync.RWMutex                             // Thread-safety lock
}

// NewNotificationService creates and initializes a new service
func NewNotificationService() *NotificationService {
	service := &NotificationService{
		channels:          make(map[NotificationType]NotificationChannel),
		userPreferences:   make(map[string]*UserPreferences),
		templates:         make(map[string]*NotificationTemplate),
		notificationQueue: make(chan *Notification, 100), // Buffer for 100 notifications
		history:           make([]*Notification, 0),
	}

	// Start background worker to process queued notifications
	go service.processNotificationQueue()

	return service
}

// RegisterChannel adds a notification channel to the service
func (service *NotificationService) RegisterChannel(channel NotificationChannel) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.channels[channel.GetType()] = channel
}

// SetUserPreferences saves notification preferences for a user
func (service *NotificationService) SetUserPreferences(preferences *UserPreferences) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.userPreferences[preferences.UserID] = preferences
}

// AddTemplate registers a new notification template
func (service *NotificationService) AddTemplate(template *NotificationTemplate) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.templates[template.ID] = template
}

// SendNotification immediately sends a notification
// Returns an error if sending fails or is blocked by preferences
func (service *NotificationService) SendNotification(notification *Notification) error {
	// Get the channel and user preferences (read lock)
	service.mutex.RLock()
	channel, channelExists := service.channels[notification.Channel]
	userPrefs := service.userPreferences[notification.UserID]
	service.mutex.RUnlock()

	// Check if the channel is configured
	if !channelExists {
		return fmt.Errorf("channel %s is not configured", notification.Channel)
	}

	// Check user preferences if they exist
	if userPrefs != nil {
		// Check if user has disabled this channel
		if !userPrefs.IsChannelEnabled(notification.Channel) {
			return fmt.Errorf("user has disabled %s notifications", notification.Channel)
		}

		// Check quiet hours (Critical notifications bypass quiet hours)
		if userPrefs.IsQuietHours() && notification.Priority != PriorityCritical {
			return fmt.Errorf("quiet hours active - notification queued for later")
		}
	}

	// Send the notification
	err := channel.Send(notification)
	if err != nil {
		notification.Status = StatusFailed
		return err
	}

	// Mark as sent and record the time
	notification.Status = StatusSent
	notification.SentAt = time.Now()

	// Add to history (write lock)
	service.mutex.Lock()
	service.history = append(service.history, notification)
	service.mutex.Unlock()

	return nil
}

// QueueNotification adds a notification to the async processing queue
// Use this for non-urgent notifications to avoid blocking
func (service *NotificationService) QueueNotification(notification *Notification) {
	service.notificationQueue <- notification
}

// processNotificationQueue is a background worker that processes
// queued notifications one by one
func (service *NotificationService) processNotificationQueue() {
	for notification := range service.notificationQueue {
		err := service.SendNotification(notification)
		if err != nil {
			fmt.Printf("  [QUEUE] Failed to send %s: %v\n", notification.ID, err)
		}
	}
}

// SendFromTemplate creates and sends a notification using a template
func (service *NotificationService) SendFromTemplate(
	userID string,
	templateID string,
	parameters map[string]string,
) error {
	// Get the template
	service.mutex.RLock()
	template, exists := service.templates[templateID]
	service.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("template not found: %s", templateID)
	}

	// Render the template with parameters
	title, body := template.Render(parameters)

	// Create and send the notification
	notification := NewNotification(userID, title, body, template.Channel, PriorityMedium)
	return service.SendNotification(notification)
}

// SendToMultipleChannels sends the same message through multiple channels
// Useful for critical alerts that need maximum visibility
func (service *NotificationService) SendToMultipleChannels(
	userID string,
	title string,
	message string,
	channels []NotificationType,
	priority NotificationPriority,
) {
	for _, channelType := range channels {
		notification := NewNotification(userID, title, message, channelType, priority)
		err := service.SendNotification(notification)
		if err != nil {
			fmt.Printf("  [MULTI] Failed on %s: %v\n", channelType, err)
		}
	}
}

// GetNotificationHistory returns a copy of the notification history
func (service *NotificationService) GetNotificationHistory() []*Notification {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	// Return a copy to prevent external modification
	historyCopy := make([]*Notification, len(service.history))
	copy(historyCopy, service.history)
	return historyCopy
}

// ==================== MAIN - DEMO ====================

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("       🔔 NOTIFICATION SYSTEM DEMO")
	fmt.Println("═══════════════════════════════════════════")

	// ========== STEP 1: Create the notification service ==========
	service := NewNotificationService()

	// ========== STEP 2: Configure notification channels ==========
	// We use decorators to add logging and retry capabilities

	// Email channel with retry and logging
	emailChannel := NewLoggingDecorator(
		NewRetryDecorator(
			NewEmailChannel("smtp.example.com", 587, "noreply@example.com"),
			3,           // Max 3 retries
			time.Second, // 1 second between retries
		),
	)

	// Other channels with logging only
	smsChannel := NewLoggingDecorator(
		NewSMSChannel("twilio", "api-key-here"),
	)
	pushChannel := NewLoggingDecorator(
		NewPushChannel("fcm-key-here"),
	)
	slackChannel := NewLoggingDecorator(
		NewSlackChannel("https://hooks.slack.com/services/..."),
	)

	// Register all channels with the service
	service.RegisterChannel(emailChannel)
	service.RegisterChannel(smsChannel)
	service.RegisterChannel(pushChannel)
	service.RegisterChannel(slackChannel)

	// ========== STEP 3: Configure user preferences ==========
	userPrefs := NewUserPreferences("user123")
	userPrefs.Email = "user@example.com"
	userPrefs.Phone = "+1234567890"
	userPrefs.EnabledChannels[NotificationTypeSMS] = true // Enable SMS
	// userPrefs.QuietHoursStart = 22 // Uncomment to test quiet hours
	// userPrefs.QuietHoursEnd = 7
	service.SetUserPreferences(userPrefs)

	// ========== STEP 4: Add notification templates ==========
	welcomeTemplate := NewTemplate(
		"welcome",
		"Welcome Email",
		"Welcome to {app_name}!",
		"Hi {name}, thanks for joining {app_name}. Get started now!",
		NotificationTypeEmail,
	)

	orderShippedTemplate := NewTemplate(
		"order_shipped",
		"Order Shipped",
		"Your order #{order_id} has shipped!",
		"Track your package: {tracking_url}",
		NotificationTypePush,
	)

	service.AddTemplate(welcomeTemplate)
	service.AddTemplate(orderShippedTemplate)

	// ========== STEP 5: Send various notifications ==========
	fmt.Println("\n📤 Sending Notifications...")
	fmt.Println("─────────────────────────────────────────")

	// Example 1: Direct email notification
	fmt.Println("\n1️⃣  Direct Email Notification:")
	passwordResetNotif := NewNotification(
		"user123",
		"Password Reset Request",
		"Click the link below to reset your password.",
		NotificationTypeEmail,
		PriorityHigh,
	)
	service.SendNotification(passwordResetNotif)

	// Example 2: SMS notification (for OTP)
	fmt.Println("\n2️⃣  SMS Notification (OTP):")
	otpNotif := NewNotification(
		"user123",
		"", // SMS typically doesn't have a title
		"Your verification code is 123456. Valid for 5 minutes.",
		NotificationTypeSMS,
		PriorityCritical,
	)
	service.SendNotification(otpNotif)

	// Example 3: Push notification
	fmt.Println("\n3️⃣  Push Notification:")
	saleNotif := NewNotification(
		"user123",
		"Flash Sale! 🎉",
		"50% off on all items for the next 2 hours!",
		NotificationTypePush,
		PriorityMedium,
	)
	service.SendNotification(saleNotif)

	// Example 4: Using a template
	fmt.Println("\n4️⃣  Notification from Template:")
	service.SendFromTemplate("user123", "welcome", map[string]string{
		"name":     "John",
		"app_name": "MyApp",
	})

	// Example 5: Multi-channel security alert
	fmt.Println("\n5️⃣  Multi-Channel Security Alert:")
	service.SendToMultipleChannels(
		"user123",
		"⚠️ Security Alert",
		"New login detected from an unknown device in New York, USA.",
		[]NotificationType{
			NotificationTypeEmail,
			NotificationTypePush,
			NotificationTypeSMS,
		},
		PriorityCritical,
	)

	// Example 6: Slack notification for team
	fmt.Println("\n6️⃣  Slack Team Notification:")
	deployNotif := NewNotification(
		"team-devops",
		"🚀 Deployment Complete",
		"Version 2.0.0 has been deployed to production successfully.",
		NotificationTypeSlack,
		PriorityMedium,
	)
	service.SendNotification(deployNotif)

	// ========== SUMMARY ==========
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  📚 KEY DESIGN PATTERNS USED:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. Strategy Pattern")
	fmt.Println("     → Different channels (Email, SMS, Push, Slack)")
	fmt.Println("     → All implement NotificationChannel interface")
	fmt.Println()
	fmt.Println("  2. Decorator Pattern")
	fmt.Println("     → RetryDecorator adds automatic retry logic")
	fmt.Println("     → LoggingDecorator adds send logging")
	fmt.Println("     → Can be stacked for combined functionality")
	fmt.Println()
	fmt.Println("  3. Template Pattern")
	fmt.Println("     → Reusable notification templates")
	fmt.Println("     → Dynamic content with placeholders")
	fmt.Println()
	fmt.Println("  4. Additional Features:")
	fmt.Println("     → User preferences (channel opt-in/out)")
	fmt.Println("     → Quiet hours support")
	fmt.Println("     → Async queue processing")
	fmt.Println("     → Thread-safe operations")
	fmt.Println("═══════════════════════════════════════════")
}
