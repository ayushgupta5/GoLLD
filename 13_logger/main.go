package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// ============================================================
// LOGGER SYSTEM - Low Level Design
// ============================================================
//
// This logger system demonstrates several important design patterns:
//
// 1. SINGLETON PATTERN: Ensures only one logger instance exists globally
// 2. STRATEGY PATTERN: Different handlers (console, file) can be swapped
// 3. CHAIN OF RESPONSIBILITY: Filters process messages in sequence
// 4. THREAD SAFETY: Uses mutexes to prevent race conditions
//
// ============================================================

// ==================== LOG LEVEL ====================
// LogLevel represents the severity of a log message.
// Lower values = less severe, Higher values = more severe.

type LogLevel int

const (
	DEBUG LogLevel = iota // 0 - Detailed debugging information
	INFO                  // 1 - General informational messages
	WARN                  // 2 - Warning messages for potential issues
	ERROR                 // 3 - Error messages for failures
	FATAL                 // 4 - Critical errors that may crash the app
)

// logLevelNames maps LogLevel to human-readable strings
var logLevelNames = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

// String returns the string representation of a LogLevel
func (level LogLevel) String() string {
	if level < DEBUG || level > FATAL {
		return "UNKNOWN"
	}
	return logLevelNames[level]
}

// logLevelColors maps LogLevel to terminal color codes
var logLevelColors = map[LogLevel]string{
	DEBUG: "\033[36m", // Cyan - for debug messages
	INFO:  "\033[32m", // Green - for info messages
	WARN:  "\033[33m", // Yellow - for warnings
	ERROR: "\033[31m", // Red - for errors
	FATAL: "\033[35m", // Magenta - for fatal errors
}

// Color returns the ANSI color code for the log level
func (level LogLevel) Color() string {
	if color, exists := logLevelColors[level]; exists {
		return color
	}
	return "" // No color for unknown levels
}

// ==================== LOG MESSAGE ====================
// LogMessage holds all information about a single log entry.

type LogMessage struct {
	Level     LogLevel  // Severity level of the message
	Message   string    // The actual log content
	Timestamp time.Time // When the message was created
	Source    string    // Which component generated this log
}

// NewLogMessage creates a new log message with the current timestamp
func NewLogMessage(level LogLevel, message string, source string) *LogMessage {
	return &LogMessage{
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Source:    source,
	}
}

// ==================== LOG HANDLER INTERFACE ====================
// LogHandler defines how log messages are output (console, file, etc.)
// This is the STRATEGY PATTERN - different strategies for handling logs.

type LogHandler interface {
	// Handle processes and outputs a log message
	Handle(message *LogMessage)

	// SetLevel sets the minimum level this handler will process
	SetLevel(level LogLevel)

	// GetLevel returns the current minimum level
	GetLevel() LogLevel
}

// ==================== CONSOLE HANDLER ====================
// ConsoleHandler outputs log messages to the terminal (stdout).

type ConsoleHandler struct {
	minimumLevel LogLevel   // Only log messages at or above this level
	useColors    bool       // Whether to use colored output
	mutex        sync.Mutex // Prevents concurrent writes from mixing up
}

// NewConsoleHandler creates a handler that writes to the console
func NewConsoleHandler(minimumLevel LogLevel) *ConsoleHandler {
	return &ConsoleHandler{
		minimumLevel: minimumLevel,
		useColors:    true, // Colors enabled by default
	}
}

// SetLevel changes the minimum log level
func (handler *ConsoleHandler) SetLevel(level LogLevel) {
	handler.minimumLevel = level
}

// GetLevel returns the current minimum log level
func (handler *ConsoleHandler) GetLevel() LogLevel {
	return handler.minimumLevel
}

// Handle writes the log message to console if it meets the level threshold
func (handler *ConsoleHandler) Handle(message *LogMessage) {
	// Skip messages below our minimum level
	if message.Level < handler.minimumLevel {
		return
	}

	// Lock to prevent garbled output from concurrent goroutines
	handler.mutex.Lock()
	defer handler.mutex.Unlock()

	// Format the timestamp in a readable way
	formattedTime := message.Timestamp.Format("2006-01-02 15:04:05")

	// ANSI reset code to clear color after the message
	const colorReset = "\033[0m"

	if handler.useColors {
		// Colored output: [timestamp] LEVEL [source] message
		fmt.Printf("%s[%s] %s [%s] %s%s\n",
			message.Level.Color(),
			formattedTime,
			message.Level,
			message.Source,
			message.Message,
			colorReset,
		)
	} else {
		// Plain output without colors
		fmt.Printf("[%s] %s [%s] %s\n",
			formattedTime,
			message.Level,
			message.Source,
			message.Message,
		)
	}
}

// ==================== FILE HANDLER ====================
// FileHandler writes log messages to a file for persistent storage.

type FileHandler struct {
	minimumLevel LogLevel   // Only log messages at or above this level
	filePath     string     // Path to the log file
	file         *os.File   // The open file handle
	mutex        sync.Mutex // Prevents concurrent writes
}

// NewFileHandler creates a handler that writes to a file
// Returns an error if the file cannot be opened/created
func NewFileHandler(minimumLevel LogLevel, filePath string) (*FileHandler, error) {
	// Open file in append mode, create if doesn't exist
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &FileHandler{
		minimumLevel: minimumLevel,
		filePath:     filePath,
		file:         file,
	}, nil
}

// SetLevel changes the minimum log level
func (handler *FileHandler) SetLevel(level LogLevel) {
	handler.minimumLevel = level
}

// GetLevel returns the current minimum log level
func (handler *FileHandler) GetLevel() LogLevel {
	return handler.minimumLevel
}

// Handle writes the log message to file if it meets the level threshold
func (handler *FileHandler) Handle(message *LogMessage) {
	// Skip messages below our minimum level
	if message.Level < handler.minimumLevel {
		return
	}

	// Lock to prevent file corruption from concurrent writes
	handler.mutex.Lock()
	defer handler.mutex.Unlock()

	// Format the log line (no colors in files)
	formattedTime := message.Timestamp.Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] %s [%s] %s\n",
		formattedTime,
		message.Level,
		message.Source,
		message.Message,
	)

	// Write to file (ignoring errors for simplicity)
	_, _ = handler.file.WriteString(logLine)
}

// Close closes the log file - always call this when done!
func (handler *FileHandler) Close() error {
	if handler.file != nil {
		return handler.file.Close()
	}
	return nil
}

// ==================== LOG FILTER INTERFACE ====================
// LogFilter decides whether a message should be logged.
// This is the CHAIN OF RESPONSIBILITY PATTERN - filters can be linked.

type LogFilter interface {
	// SetNext links this filter to the next one in the chain
	SetNext(nextFilter LogFilter)

	// ShouldLog returns true if the message passes this filter
	ShouldLog(message *LogMessage) bool
}

// ==================== LEVEL FILTER ====================
// LevelFilter only allows messages at or above a minimum level.

type LevelFilter struct {
	minimumLevel LogLevel  // Minimum level to allow
	nextFilter   LogFilter // Next filter in chain (can be nil)
}

// NewLevelFilter creates a filter that blocks messages below the given level
func NewLevelFilter(minimumLevel LogLevel) *LevelFilter {
	return &LevelFilter{minimumLevel: minimumLevel}
}

// SetNext sets the next filter in the chain
func (filter *LevelFilter) SetNext(nextFilter LogFilter) {
	filter.nextFilter = nextFilter
}

// ShouldLog returns true if message level meets minimum requirement
func (filter *LevelFilter) ShouldLog(message *LogMessage) bool {
	// Block messages below minimum level
	if message.Level < filter.minimumLevel {
		return false
	}

	// If there's a next filter, check it too
	if filter.nextFilter != nil {
		return filter.nextFilter.ShouldLog(message)
	}

	// Passed all filters!
	return true
}

// ==================== SOURCE FILTER ====================
// SourceFilter only allows messages from specific sources (components).

type SourceFilter struct {
	allowedSources map[string]bool // Map of allowed source names
	nextFilter     LogFilter       // Next filter in chain (can be nil)
}

// NewSourceFilter creates a filter that only allows specified sources
// If sources is empty, all sources are allowed
func NewSourceFilter(sources []string) *SourceFilter {
	allowedMap := make(map[string]bool)
	for _, source := range sources {
		allowedMap[source] = true
	}
	return &SourceFilter{allowedSources: allowedMap}
}

// SetNext sets the next filter in the chain
func (filter *SourceFilter) SetNext(nextFilter LogFilter) {
	filter.nextFilter = nextFilter
}

// ShouldLog returns true if message source is in the allowed list
func (filter *SourceFilter) ShouldLog(message *LogMessage) bool {
	// If we have an allow list and source is not in it, block
	if len(filter.allowedSources) > 0 && !filter.allowedSources[message.Source] {
		return false
	}

	// If there's a next filter, check it too
	if filter.nextFilter != nil {
		return filter.nextFilter.ShouldLog(message)
	}

	// Passed all filters!
	return true
}

// ==================== LOGGER (SINGLETON) ====================
// Logger is the main logging system. Only one instance exists (Singleton).
// It manages handlers (where to log) and filters (what to log).

type Logger struct {
	handlers []LogHandler // List of output destinations
	filters  []LogFilter  // List of message filters
	mutex    sync.RWMutex // Read-write lock for thread safety
}

// Global singleton variables
var (
	loggerInstance *Logger   // The single logger instance
	loggerOnce     sync.Once // Ensures instance is created only once
)

// GetLogger returns the singleton logger instance.
// This is thread-safe and always returns the same instance.
func GetLogger() *Logger {
	// sync.Once ensures this block runs exactly once, even with concurrent calls
	loggerOnce.Do(func() {
		loggerInstance = &Logger{
			handlers: make([]LogHandler, 0),
			filters:  make([]LogFilter, 0),
		}
	})
	return loggerInstance
}

// AddHandler registers a new output handler (console, file, etc.)
func (logger *Logger) AddHandler(handler LogHandler) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	logger.handlers = append(logger.handlers, handler)
}

// AddFilter registers a new filter to control which messages are logged
func (logger *Logger) AddFilter(filter LogFilter) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	logger.filters = append(logger.filters, filter)
}

// log is the internal method that processes all log messages
func (logger *Logger) log(level LogLevel, source string, message string) {
	// Create the log message with current timestamp
	logMessage := NewLogMessage(level, message, source)

	// Use read lock since we're only reading handlers/filters
	logger.mutex.RLock()
	defer logger.mutex.RUnlock()

	// Check all filters - if any filter blocks, don't log
	for _, filter := range logger.filters {
		if !filter.ShouldLog(logMessage) {
			return // Message was filtered out
		}
	}

	// Send message to all registered handlers
	for _, handler := range logger.handlers {
		handler.Handle(logMessage)
	}
}

// ==================== PUBLIC LOGGING METHODS ====================
// These are the main methods users call to log messages.

// Debug logs a debug-level message
func (logger *Logger) Debug(source string, message string) {
	logger.log(DEBUG, source, message)
}

// Info logs an info-level message
func (logger *Logger) Info(source string, message string) {
	logger.log(INFO, source, message)
}

// Warn logs a warning-level message
func (logger *Logger) Warn(source string, message string) {
	logger.log(WARN, source, message)
}

// Error logs an error-level message
func (logger *Logger) Error(source string, message string) {
	logger.log(ERROR, source, message)
}

// Fatal logs a fatal-level message
func (logger *Logger) Fatal(source string, message string) {
	logger.log(FATAL, source, message)
}

// ==================== FORMATTED LOGGING METHODS ====================
// These methods support printf-style formatting.

// Debugf logs a formatted debug message
func (logger *Logger) Debugf(source string, format string, args ...interface{}) {
	logger.Debug(source, fmt.Sprintf(format, args...))
}

// Infof logs a formatted info message
func (logger *Logger) Infof(source string, format string, args ...interface{}) {
	logger.Info(source, fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message
func (logger *Logger) Warnf(source string, format string, args ...interface{}) {
	logger.Warn(source, fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message
func (logger *Logger) Errorf(source string, format string, args ...interface{}) {
	logger.Error(source, fmt.Sprintf(format, args...))
}

// Fatalf logs a formatted fatal message
func (logger *Logger) Fatalf(source string, format string, args ...interface{}) {
	logger.Fatal(source, fmt.Sprintf(format, args...))
}

// ==================== NAMED LOGGER ====================
// NamedLogger wraps the main Logger with a fixed source name.
// This is convenient when a component needs to log many messages.

type NamedLogger struct {
	componentName string  // The name that appears in all log messages
	logger        *Logger // Reference to the main logger
}

// NewNamedLogger creates a logger pre-configured with a component name
func NewNamedLogger(componentName string) *NamedLogger {
	return &NamedLogger{
		componentName: componentName,
		logger:        GetLogger(),
	}
}

// Debug logs a debug message with the component name
func (named *NamedLogger) Debug(message string) {
	named.logger.Debug(named.componentName, message)
}

// Info logs an info message with the component name
func (named *NamedLogger) Info(message string) {
	named.logger.Info(named.componentName, message)
}

// Warn logs a warning message with the component name
func (named *NamedLogger) Warn(message string) {
	named.logger.Warn(named.componentName, message)
}

// Error logs an error message with the component name
func (named *NamedLogger) Error(message string) {
	named.logger.Error(named.componentName, message)
}

// Fatal logs a fatal message with the component name
func (named *NamedLogger) Fatal(message string) {
	named.logger.Fatal(named.componentName, message)
}

// Debugf logs a formatted debug message
func (named *NamedLogger) Debugf(format string, args ...interface{}) {
	named.logger.Debugf(named.componentName, format, args...)
}

// Infof logs a formatted info message
func (named *NamedLogger) Infof(format string, args ...interface{}) {
	named.logger.Infof(named.componentName, format, args...)
}

// Warnf logs a formatted warning message
func (named *NamedLogger) Warnf(format string, args ...interface{}) {
	named.logger.Warnf(named.componentName, format, args...)
}

// Errorf logs a formatted error message
func (named *NamedLogger) Errorf(format string, args ...interface{}) {
	named.logger.Errorf(named.componentName, format, args...)
}

// Fatalf logs a formatted fatal message
func (named *NamedLogger) Fatalf(format string, args ...interface{}) {
	named.logger.Fatalf(named.componentName, format, args...)
}

// ==================== MAIN - DEMONSTRATION ====================

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("          📝 LOGGER SYSTEM DEMO")
	fmt.Println("═══════════════════════════════════════════")

	// Step 1: Get the singleton logger instance
	logger := GetLogger()

	// Step 2: Add a console handler that shows INFO and above
	// (DEBUG messages won't appear on console)
	consoleHandler := NewConsoleHandler(INFO)
	logger.AddHandler(consoleHandler)

	// Step 3: Add a file handler that logs everything (including DEBUG)
	fileHandler, err := NewFileHandler(DEBUG, "/tmp/app.log")
	if err != nil {
		fmt.Printf("Warning: Could not create file handler: %v\n", err)
	} else {
		logger.AddHandler(fileHandler)
		defer fileHandler.Close() // Important: Close file when done!
	}

	// Step 4: Add a level filter (allows DEBUG and above)
	levelFilter := NewLevelFilter(DEBUG)
	logger.AddFilter(levelFilter)

	// ========== Demo 1: Basic Logging ==========
	fmt.Println("\n📋 Demo 1: Basic Logging")
	fmt.Println("─────────────────────────────────────────")

	logger.Debug("Main", "Debug message (only in file, not console)")
	logger.Info("Main", "Application started successfully")
	logger.Warn("Main", "Low memory warning - consider cleanup")
	logger.Error("Main", "Failed to connect to database")

	// ========== Demo 2: Named Loggers ==========
	fmt.Println("\n📋 Demo 2: Named Loggers (for different components)")
	fmt.Println("─────────────────────────────────────────")

	// Create named loggers for different services
	userLogger := NewNamedLogger("UserService")
	orderLogger := NewNamedLogger("OrderService")
	paymentLogger := NewNamedLogger("PaymentService")

	// Log from each service
	userLogger.Info("User login successful")
	userLogger.Infof("User %s logged in from IP %s", "john@email.com", "192.168.1.1")

	orderLogger.Info("New order created")
	orderLogger.Infof("Order #%d created for user ID %d", 12345, 100)

	paymentLogger.Warn("Payment gateway responding slowly")
	paymentLogger.Error("Payment failed: connection timeout")

	// ========== Demo 3: Multi-Component System ==========
	fmt.Println("\n📋 Demo 3: Simulating a Multi-Component System")
	fmt.Println("─────────────────────────────────────────")

	// Create loggers for infrastructure components
	databaseLogger := NewNamedLogger("Database")
	cacheLogger := NewNamedLogger("Cache")
	apiLogger := NewNamedLogger("API")

	// Simulate system startup
	databaseLogger.Info("Connected to PostgreSQL successfully")
	cacheLogger.Info("Redis connection established")
	apiLogger.Infof("HTTP server listening on port %d", 8080)

	// Simulate normal operation
	databaseLogger.Warnf("Slow query detected: took %dms", 500)
	cacheLogger.Info("Cache hit for key: user:123")
	apiLogger.Errorf("Request failed with status: %s", "404 Not Found")

	// ========== Demo 4: Source Filtering ==========
	fmt.Println("\n📋 Demo 4: Source Filtering (showing only specific components)")
	fmt.Println("─────────────────────────────────────────")

	// Create a new logger instance for source filter demo
	// Note: In real apps, you'd configure filters at startup
	sourceFilter := NewSourceFilter([]string{"Database", "API"})
	logger.AddFilter(sourceFilter)

	// Now only Database and API logs will appear
	databaseLogger.Info("This message WILL appear (Database is allowed)")
	cacheLogger.Info("This message will NOT appear (Cache is filtered out)")
	apiLogger.Info("This message WILL appear (API is allowed)")

	// ========== Summary ==========
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  📚 KEY DESIGN PATTERNS USED:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. SINGLETON: One global logger instance")
	fmt.Println("  2. STRATEGY: Pluggable handlers (console/file)")
	fmt.Println("  3. CHAIN OF RESPONSIBILITY: Filter chain")
	fmt.Println("  4. THREAD SAFETY: Mutex locks prevent races")
	fmt.Println("  5. NAMED LOGGER: Convenient component logging")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("\n✅ Check /tmp/app.log for file output!")
}
