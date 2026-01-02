package main

import (
	"fmt"
	"sync"
)

// ============================================================
// SINGLETON PATTERN
// ============================================================
// Definition: Ensures a class has only ONE instance and provides
// a global point of access to that instance.
//
// Real-World Analogy:
// Think of it like a government - there's only ONE president at a time.
// Everyone in the country accesses the same president, not their own copy.
//
// WHEN TO USE:
// - Database connection pools (expensive to create, share one pool)
// - Configuration managers (one config for entire app)
// - Logger instances (centralized logging)
// - Cache managers (shared cache across app)
//
// WHY USE IT:
// - Controlled access to single instance
// - Reduced memory footprint (only one object in memory)
// - Global state management
// - Lazy initialization (created only when first needed)
//
// INTERVIEW TIP:
// Interviewers often ask: "How do you make it thread-safe in Go?"
// Answer: Use sync.Once - it's the cleanest and most idiomatic approach!

// ============================================================
// EXAMPLE 1: BAD Implementation (NOT Thread-Safe)
// ============================================================
// This shows what NOT to do - has race conditions in concurrent environment

// unsafeConfigInstance stores the singleton (but it's not thread-safe!)
var unsafeConfigInstance *UnsafeConfig

// UnsafeConfig demonstrates a singleton that is NOT thread-safe
type UnsafeConfig struct {
	DatabaseURL string
	APIKey      string
}

// GetUnsafeConfig returns the singleton instance (BAD - has race condition!)
// Problem: If two goroutines call this at the same time:
//   - Goroutine 1: checks if unsafeConfigInstance == nil (yes, it's nil)
//   - Goroutine 2: checks if unsafeConfigInstance == nil (yes, still nil!)
//   - Both goroutines create new instances - BROKEN!
func GetUnsafeConfig() *UnsafeConfig {
	if unsafeConfigInstance == nil {
		// Race condition here! Multiple goroutines could pass the nil check
		unsafeConfigInstance = &UnsafeConfig{}
	}
	return unsafeConfigInstance
}

// ============================================================
// EXAMPLE 2: GOOD Implementation (Thread-Safe using sync.Once)
// ============================================================
// sync.Once ensures the initialization code runs exactly ONCE,
// no matter how many goroutines call it simultaneously.

// AppConfig holds application configuration settings
type AppConfig struct {
	DatabaseURL string
	APIKey      string
	DebugMode   bool
}

// Package-level variables for the singleton
var (
	appConfigInstance *AppConfig // The single instance
	appConfigOnce     sync.Once  // Ensures one-time initialization
)

// GetAppConfig returns the singleton AppConfig instance
// This is thread-safe: sync.Once guarantees single initialization
func GetAppConfig() *AppConfig {
	// Do() runs the function EXACTLY ONCE, even if called from multiple goroutines
	appConfigOnce.Do(func() {
		fmt.Println("  [AppConfig] Initializing... (this message appears only once)")
		appConfigInstance = &AppConfig{
			DatabaseURL: "postgres://localhost:5432/mydb",
			APIKey:      "secret-api-key-123",
			DebugMode:   true,
		}
	})
	return appConfigInstance
}

// ============================================================
// EXAMPLE 3: Database Connection Pool (Practical Use Case)
// ============================================================
// Connection pools are expensive to create, so we use singleton pattern
// to share a single pool across the entire application.

// DatabaseConnectionPool manages a pool of database connections
type DatabaseConnectionPool struct {
	availableConnections []string   // List of available connections
	maxPoolSize          int        // Maximum connections allowed
	mutex                sync.Mutex // Protects concurrent access to the pool
}

// Package-level variables for the connection pool singleton
var (
	connectionPoolInstance *DatabaseConnectionPool
	connectionPoolOnce     sync.Once
)

// GetDatabaseConnectionPool returns the singleton connection pool
func GetDatabaseConnectionPool() *DatabaseConnectionPool {
	connectionPoolOnce.Do(func() {
		fmt.Println("  [ConnectionPool] Creating pool... (happens only once)")
		connectionPoolInstance = &DatabaseConnectionPool{
			availableConnections: make([]string, 0),
			maxPoolSize:          10,
		}
		// Pre-create some connections
		for i := 1; i <= 5; i++ {
			connectionName := fmt.Sprintf("db-connection-%d", i)
			connectionPoolInstance.availableConnections = append(
				connectionPoolInstance.availableConnections,
				connectionName,
			)
		}
	})
	return connectionPoolInstance
}

// BorrowConnection gets a connection from the pool
// Returns the connection name and a boolean indicating success
func (pool *DatabaseConnectionPool) BorrowConnection() (string, bool) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if len(pool.availableConnections) == 0 {
		return "", false // No connections available
	}

	// Get the last connection (LIFO - Last In, First Out)
	lastIndex := len(pool.availableConnections) - 1
	connection := pool.availableConnections[lastIndex]
	pool.availableConnections = pool.availableConnections[:lastIndex]

	return connection, true
}

// ReturnConnection returns a connection back to the pool
func (pool *DatabaseConnectionPool) ReturnConnection(connectionName string) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Only add back if we haven't exceeded max pool size
	if len(pool.availableConnections) < pool.maxPoolSize {
		pool.availableConnections = append(pool.availableConnections, connectionName)
	}
}

// GetAvailableCount returns the number of available connections
func (pool *DatabaseConnectionPool) GetAvailableCount() int {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	return len(pool.availableConnections)
}

// ============================================================
// EXAMPLE 4: Application Logger (Another Practical Use Case)
// ============================================================
// Loggers are typically singletons so all parts of the app
// write to the same log with consistent formatting.

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogLevelDebug LogLevel = iota // 0 - Most verbose
	LogLevelInfo                  // 1 - General information
	LogLevelWarn                  // 2 - Warnings
	LogLevelError                 // 3 - Errors only
)

// logLevelNames maps log levels to their display names
var logLevelNames = map[LogLevel]string{
	LogLevelDebug: "DEBUG",
	LogLevelInfo:  "INFO",
	LogLevelWarn:  "WARN",
	LogLevelError: "ERROR",
}

// ApplicationLogger handles logging throughout the application
type ApplicationLogger struct {
	minimumLevel LogLevel   // Only log messages at or above this level
	mutex        sync.Mutex // Protects concurrent writes
}

// Package-level variables for the logger singleton
var (
	loggerInstance *ApplicationLogger
	loggerOnce     sync.Once
)

// GetLogger returns the singleton logger instance
func GetLogger() *ApplicationLogger {
	loggerOnce.Do(func() {
		fmt.Println("  [Logger] Creating logger... (happens only once)")
		loggerInstance = &ApplicationLogger{
			minimumLevel: LogLevelInfo, // Default: show INFO and above
		}
	})
	return loggerInstance
}

// SetMinimumLevel changes the minimum log level
func (logger *ApplicationLogger) SetMinimumLevel(level LogLevel) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()
	logger.minimumLevel = level
}

// logMessage is a helper that handles the actual logging
func (logger *ApplicationLogger) logMessage(level LogLevel, message string) {
	logger.mutex.Lock()
	defer logger.mutex.Unlock()

	// Only print if the message level is >= minimum level
	if level >= logger.minimumLevel {
		levelName := logLevelNames[level]
		fmt.Printf("  [%s] %s\n", levelName, message)
	}
}

// Debug logs a debug-level message (most verbose)
func (logger *ApplicationLogger) Debug(message string) {
	logger.logMessage(LogLevelDebug, message)
}

// Info logs an info-level message
func (logger *ApplicationLogger) Info(message string) {
	logger.logMessage(LogLevelInfo, message)
}

// Warn logs a warning-level message
func (logger *ApplicationLogger) Warn(message string) {
	logger.logMessage(LogLevelWarn, message)
}

// Error logs an error-level message (most severe)
func (logger *ApplicationLogger) Error(message string) {
	logger.logMessage(LogLevelError, message)
}

// ============================================================
// EXAMPLE 5: Simple Alternative - Package-Level Initialization
// ============================================================
// For simple cases, you can use Go's init() or package-level vars.
// This is automatically a singleton because packages are loaded once.

// GlobalSettings holds application-wide settings
// Initialized when package loads - automatically singleton!
var GlobalSettings = &AppSettings{
	ApplicationName: "MyApp",
	Version:         "1.0.0",
	MaxRetryCount:   3,
}

// AppSettings holds basic application settings
type AppSettings struct {
	ApplicationName string
	Version         string
	MaxRetryCount   int
}

// ============================================================
// KEY INTERVIEW QUESTIONS & ANSWERS
// ============================================================
//
// Q1: What problems does the Singleton pattern solve?
// A1: 1. Ensures only one instance exists (e.g., config, connection pool)
//     2. Provides a global access point to that instance
//     3. Supports lazy initialization (created only when first needed)
//     4. Reduces memory by avoiding duplicate objects
//
// Q2: What are the drawbacks of Singleton?
// A2: 1. Hard to test (global state makes unit testing difficult)
//     2. Hidden dependencies (not visible in function signatures)
//     3. Can lead to tight coupling between components
//     4. Concurrency bugs if not implemented correctly
//
// Q3: How do you make Singleton testable?
// A3: 1. Inject the singleton via interfaces (dependency injection)
//     2. Provide a Reset() function for tests
//     3. Consider if you really need singleton (maybe DI is better)
//
// Q4: Why use sync.Once instead of sync.Mutex for singleton?
// A4: sync.Once is specifically designed for one-time initialization.
//     It's cleaner, more efficient, and more idiomatic in Go.
//     After the first call, subsequent calls have almost zero overhead.
//
// COMMON MISTAKES TO AVOID:
// 1. Forgetting to make it thread-safe
// 2. Overusing singleton (not everything needs to be a singleton)
// 3. Making mutable singletons without proper synchronization
// 4. Not considering alternatives like dependency injection

// ============================================================
// MAIN FUNCTION - Demonstrates all singleton examples
// ============================================================

func main() {
	fmt.Println("============================================================")
	fmt.Println("           SINGLETON PATTERN DEMONSTRATION")
	fmt.Println("============================================================")

	// ---------------------------------------------------------
	// Demo 1: AppConfig Singleton
	// ---------------------------------------------------------
	fmt.Println("\n--- Demo 1: AppConfig Singleton ---")

	fmt.Println("Calling GetAppConfig() first time:")
	config1 := GetAppConfig()
	fmt.Printf("  Got config: DatabaseURL=%s\n", config1.DatabaseURL)

	fmt.Println("\nCalling GetAppConfig() second time:")
	config2 := GetAppConfig()
	fmt.Printf("  Got config: DatabaseURL=%s\n", config2.DatabaseURL)

	fmt.Println("\nVerifying both are the same instance:")
	fmt.Printf("  config1 == config2? %v\n", config1 == config2)

	fmt.Println("\nProof: Change config1, see it in config2:")
	config1.DebugMode = false
	fmt.Printf("  config1.DebugMode = false")
	fmt.Printf("\n  config2.DebugMode = %v (changed!)\n", config2.DebugMode)

	// ---------------------------------------------------------
	// Demo 2: Database Connection Pool Singleton
	// ---------------------------------------------------------
	fmt.Println("\n--- Demo 2: Connection Pool Singleton ---")

	pool := GetDatabaseConnectionPool()
	fmt.Printf("  Available connections: %d\n", pool.GetAvailableCount())

	// Borrow some connections
	conn1, _ := pool.BorrowConnection()
	conn2, _ := pool.BorrowConnection()
	fmt.Printf("  Borrowed: %s, %s\n", conn1, conn2)
	fmt.Printf("  Available after borrowing 2: %d\n", pool.GetAvailableCount())

	// Return one connection
	pool.ReturnConnection(conn1)
	fmt.Printf("  Available after returning 1: %d\n", pool.GetAvailableCount())

	// Verify same pool from another call
	pool2 := GetDatabaseConnectionPool()
	fmt.Printf("  Same pool instance? %v\n", pool == pool2)

	// ---------------------------------------------------------
	// Demo 3: Logger Singleton
	// ---------------------------------------------------------
	fmt.Println("\n--- Demo 3: Logger Singleton ---")

	logger := GetLogger()

	fmt.Println("\nWith LogLevel = DEBUG (show all messages):")
	logger.SetMinimumLevel(LogLevelDebug)
	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")

	fmt.Println("\nWith LogLevel = WARN (show only WARN and ERROR):")
	logger.SetMinimumLevel(LogLevelWarn)
	logger.Debug("This debug message will NOT appear")
	logger.Info("This info message will NOT appear")
	logger.Warn("This warning WILL appear")
	logger.Error("This error WILL appear")

	// ---------------------------------------------------------
	// Demo 4: Package-Level Singleton
	// ---------------------------------------------------------
	fmt.Println("\n--- Demo 4: Package-Level Singleton (GlobalSettings) ---")
	fmt.Printf("  Application: %s\n", GlobalSettings.ApplicationName)
	fmt.Printf("  Version: %s\n", GlobalSettings.Version)
	fmt.Printf("  Max Retries: %d\n", GlobalSettings.MaxRetryCount)

	fmt.Println("\n============================================================")
	fmt.Println("                    END OF DEMO")
	fmt.Println("============================================================")
}
