package main

import "fmt"

// ============================================================================
// INTERFACE SEGREGATION PRINCIPLE (ISP)
// ============================================================================
//
// DEFINITION:
// "Clients should not be forced to depend on interfaces they don't use."
//
// WHAT THIS MEANS (Simple Explanation):
// - Don't create large "fat" interfaces with too many methods
// - Instead, create small, focused interfaces with just a few methods
// - Each struct should only implement the methods it actually needs
//
// WHY IT MATTERS:
// 1. Prevents empty or panic-throwing method implementations
// 2. Makes code easier to understand and maintain
// 3. Makes testing simpler (fewer methods to mock)
// 4. Promotes loose coupling between components
//
// GO'S NATURAL ALIGNMENT WITH ISP:
// Go encourages small interfaces! The Go proverb says:
// "The bigger the interface, the weaker the abstraction"
//
// EXAMPLES FROM GO STANDARD LIBRARY:
// - io.Reader:  Read(p []byte) (n int, err error)
// - io.Writer:  Write(p []byte) (n int, err error)
// - io.Closer:  Close() error
// - error:      Error() string
//
// These are all single-method interfaces - the ultimate form of ISP!
// ============================================================================

// ============================================================================
// PART 1: BAD EXAMPLE - The Problem with Fat Interfaces
// ============================================================================

// MultiFunctionMachineBad is a "fat" interface that forces ALL machines
// to implement every method, even if they don't support that feature.
type MultiFunctionMachineBad interface {
	Print()
	Scan()
	Fax()
	Staple()
}

// BasicPrinterBad represents a simple printer that can ONLY print.
// Problem: It's forced to implement Scan, Fax, and Staple even though
// a basic printer doesn't have these capabilities!
type BasicPrinterBad struct{}

func (printer BasicPrinterBad) Print() {
	fmt.Println("BasicPrinterBad: Printing document...")
}

func (printer BasicPrinterBad) Scan() {
	// BAD: This printer can't scan, but we're forced to implement this method!
	// We have to either panic or return an error, which is poor design.
	panic("BasicPrinterBad: Cannot scan - this printer doesn't have a scanner!")
}

func (printer BasicPrinterBad) Fax() {
	// BAD: This printer can't fax, but we're forced to implement this method!
	panic("BasicPrinterBad: Cannot fax - this printer doesn't have fax capability!")
}

func (printer BasicPrinterBad) Staple() {
	// BAD: This printer can't staple, but we're forced to implement this method!
	panic("BasicPrinterBad: Cannot staple - this printer doesn't have a stapler!")
}

// ============================================================================
// PART 2: GOOD EXAMPLE - Segregated Interfaces (The ISP Way)
// ============================================================================

// Step 1: Create small, focused interfaces (one capability each)

// Printer represents any device that can print documents.
type Printer interface {
	Print()
}

// Scanner represents any device that can scan documents.
type Scanner interface {
	Scan()
}

// Faxer represents any device that can send faxes.
type Faxer interface {
	Fax()
}

// Stapler represents any device that can staple documents.
type Stapler interface {
	Staple()
}

// Step 2: Combine interfaces when needed (Interface Composition)
// Go allows us to embed interfaces inside other interfaces.

// PrinterScanner combines printing and scanning capabilities.
type PrinterScanner interface {
	Printer
	Scanner
}

// AllInOneMachine has all capabilities - printing, scanning, faxing, and stapling.
type AllInOneMachine interface {
	Printer
	Scanner
	Faxer
	Stapler
}

// Step 3: Implement only what each device actually supports

// BasicPrinter is a simple printer that only prints.
// It only implements the Printer interface - nothing more!
type BasicPrinter struct {
	ModelName string
}

func (printer BasicPrinter) Print() {
	fmt.Printf("%s: Printing document...\n", printer.ModelName)
}

// Photocopier can print and scan documents.
// It implements both Printer and Scanner interfaces.
type Photocopier struct {
	ModelName string
}

func (copier Photocopier) Print() {
	fmt.Printf("%s: Printing document...\n", copier.ModelName)
}

func (copier Photocopier) Scan() {
	fmt.Printf("%s: Scanning document...\n", copier.ModelName)
}

// AdvancedOfficePrinter has all capabilities.
// It implements Printer, Scanner, Faxer, and Stapler interfaces.
type AdvancedOfficePrinter struct {
	ModelName string
}

func (officePrinter AdvancedOfficePrinter) Print() {
	fmt.Printf("%s: Printing document...\n", officePrinter.ModelName)
}

func (officePrinter AdvancedOfficePrinter) Scan() {
	fmt.Printf("%s: Scanning document...\n", officePrinter.ModelName)
}

func (officePrinter AdvancedOfficePrinter) Fax() {
	fmt.Printf("%s: Faxing document...\n", officePrinter.ModelName)
}

func (officePrinter AdvancedOfficePrinter) Staple() {
	fmt.Printf("%s: Stapling document...\n", officePrinter.ModelName)
}

// ============================================================================
// PART 3: REAL-WORLD EXAMPLE - Worker Interfaces
// ============================================================================

// BAD APPROACH: One fat interface for all workers
// Problem: A robot can work but can't eat or sleep!
type WorkerBad interface {
	Work()
	Eat()
	Sleep()
	AttendMeeting()
	SubmitTimesheet()
}

// GOOD APPROACH: Separate interfaces for each capability

// Workable represents any entity that can do work.
type Workable interface {
	Work()
}

// Eatable represents any entity that needs to eat.
type Eatable interface {
	Eat()
}

// Sleepable represents any entity that needs to sleep.
type Sleepable interface {
	Sleep()
}

// MeetingAttendable represents any entity that can attend meetings.
type MeetingAttendable interface {
	AttendMeeting()
}

// TimesheetSubmittable represents any entity that submits timesheets.
type TimesheetSubmittable interface {
	SubmitTimesheet()
}

// HumanEmployee implements all worker-related interfaces
// because humans can do all these activities.
type HumanEmployee struct {
	EmployeeName string
}

func (employee HumanEmployee) Work() {
	fmt.Printf("%s is working on their tasks...\n", employee.EmployeeName)
}

func (employee HumanEmployee) Eat() {
	fmt.Printf("%s is taking a lunch break...\n", employee.EmployeeName)
}

func (employee HumanEmployee) Sleep() {
	fmt.Printf("%s is resting at home...\n", employee.EmployeeName)
}

func (employee HumanEmployee) AttendMeeting() {
	fmt.Printf("%s is attending the team meeting...\n", employee.EmployeeName)
}

func (employee HumanEmployee) SubmitTimesheet() {
	fmt.Printf("%s has submitted their timesheet.\n", employee.EmployeeName)
}

// RobotEmployee only implements Workable.
// Robots don't eat, sleep, attend meetings, or submit timesheets!
type RobotEmployee struct {
	RobotID string
}

func (robot RobotEmployee) Work() {
	fmt.Printf("Robot %s is performing automated tasks...\n", robot.RobotID)
}

// Notice: RobotEmployee doesn't implement Eatable, Sleepable, etc.
// This is ISP in action - each type only implements what it needs!

// ============================================================================
// PART 4: PRACTICAL EXAMPLE - Repository Pattern
// ============================================================================

// BAD APPROACH: One giant repository interface
// Problem: Some repositories are read-only, some don't support search!
type RepositoryBad interface {
	Create(entity interface{}) error
	Read(id string) (interface{}, error)
	Update(entity interface{}) error
	Delete(id string) error
	List() ([]interface{}, error)
	Search(query string) ([]interface{}, error)
	Export(format string) ([]byte, error)
	Import(data []byte) error
}

// GOOD APPROACH: Small, focused repository interfaces

// Readable provides read access to data.
type Readable interface {
	Read(id string) (interface{}, error)
}

// Writable provides write access (create and update) to data.
type Writable interface {
	Create(entity interface{}) error
	Update(entity interface{}) error
}

// Deletable provides delete access to data.
type Deletable interface {
	Delete(id string) error
}

// Listable provides ability to list all data.
type Listable interface {
	List() ([]interface{}, error)
}

// Searchable provides search functionality.
type Searchable interface {
	Search(query string) ([]interface{}, error)
}

// Composed interfaces for common use cases

// ReadWriteRepository combines read and write capabilities.
type ReadWriteRepository interface {
	Readable
	Writable
}

// CRUDRepository provides full Create, Read, Update, Delete, and List capabilities.
type CRUDRepository interface {
	Readable
	Writable
	Deletable
	Listable
}

// ReadOnlyUserRepository only implements Readable.
// Perfect for scenarios where data should not be modified.
type ReadOnlyUserRepository struct {
	userData map[string]interface{}
}

func NewReadOnlyUserRepository() *ReadOnlyUserRepository {
	return &ReadOnlyUserRepository{
		userData: make(map[string]interface{}),
	}
}

func (repo *ReadOnlyUserRepository) Read(id string) (interface{}, error) {
	if value, exists := repo.userData[id]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("user not found with id: %s", id)
}

// FullUserRepository implements all CRUD operations.
type FullUserRepository struct {
	userData map[string]interface{}
}

func NewFullUserRepository() *FullUserRepository {
	return &FullUserRepository{
		userData: make(map[string]interface{}),
	}
}

func (repo *FullUserRepository) Read(id string) (interface{}, error) {
	if value, exists := repo.userData[id]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("user not found with id: %s", id)
}

func (repo *FullUserRepository) Create(entity interface{}) error {
	// In a real implementation, you would generate or extract an ID
	fmt.Println("Creating new entity in repository...")
	return nil
}

func (repo *FullUserRepository) Update(entity interface{}) error {
	// In a real implementation, you would update the entity by ID
	fmt.Println("Updating entity in repository...")
	return nil
}

func (repo *FullUserRepository) Delete(id string) error {
	if _, exists := repo.userData[id]; !exists {
		return fmt.Errorf("cannot delete: user not found with id: %s", id)
	}
	delete(repo.userData, id)
	fmt.Printf("Deleted user with id: %s\n", id)
	return nil
}

func (repo *FullUserRepository) List() ([]interface{}, error) {
	allUsers := make([]interface{}, 0, len(repo.userData))
	for _, user := range repo.userData {
		allUsers = append(allUsers, user)
	}
	return allUsers, nil
}

// ============================================================================
// PART 5: HELPER FUNCTIONS (Demonstrating ISP Benefits)
// ============================================================================

// These functions accept small, focused interfaces.
// This means they can work with ANY type that implements the required interface.

// PrintDocument accepts any type that can print.
// Works with: BasicPrinter, Photocopier, AdvancedOfficePrinter
func PrintDocument(device Printer) {
	fmt.Print("  -> ")
	device.Print()
}

// ScanDocument accepts any type that can scan.
// Works with: Photocopier, AdvancedOfficePrinter
// Does NOT work with: BasicPrinter (it doesn't implement Scanner)
func ScanDocument(device Scanner) {
	fmt.Print("  -> ")
	device.Scan()
}

// PrintAndScanDocument accepts any type that can both print AND scan.
// Works with: Photocopier, AdvancedOfficePrinter
func PrintAndScanDocument(device PrinterScanner) {
	fmt.Println("  -> Performing print and scan operation:")
	fmt.Print("     ")
	device.Print()
	fmt.Print("     ")
	device.Scan()
}

// AssignWork accepts any type that can work.
// Works with: HumanEmployee, RobotEmployee
func AssignWork(worker Workable) {
	fmt.Print("  -> ")
	worker.Work()
}

// ============================================================================
// PART 6: KEY INTERVIEW POINTS
// ============================================================================
//
// Q1: What's wrong with large/fat interfaces?
// A1: 1. Implementations are forced to have empty or panic-throwing methods
//     2. Harder to test (need to mock many methods you don't use)
//     3. Harder to understand (too many responsibilities mixed together)
//     4. Tight coupling - changes affect many unrelated implementations
//
// Q2: How small should interfaces be?
// A2: As small as possible while still being useful.
//     Ideally 1-3 methods. Go's standard library often uses 1-method interfaces.
//
// Q3: How is ISP different from SRP (Single Responsibility Principle)?
// A3: - SRP: A struct should have only one reason to change (one responsibility)
//     - ISP: An interface should be focused on one capability
//     Both lead to smaller, more focused components, but at different levels.
//
// Q4: What is Interface Composition?
// A4: Combining small interfaces into larger ones when needed.
//     Example: ReadWriter = Reader + Writer
//
// COMMON MISTAKES TO AVOID:
// 1. Creating interfaces upfront before you need them
// 2. Creating interfaces with 10+ methods
// 3. Forcing every struct to implement a mega-interface
// 4. Not using interface composition to combine small interfaces
// ============================================================================

func main() {
	fmt.Println("============================================")
	fmt.Println("  Interface Segregation Principle (ISP) Demo")
	fmt.Println("============================================")

	// -------------------------
	// Demo 1: Printer Example
	// -------------------------
	fmt.Println("\n--- DEMO 1: Printer Devices ---")

	// Create different printer devices
	basicPrinter := BasicPrinter{ModelName: "BasicPrint 100"}
	photocopier := Photocopier{ModelName: "CopyStar 200"}
	officePrinter := AdvancedOfficePrinter{ModelName: "OfficePro 3000"}

	// BasicPrinter can only print
	fmt.Println("\nBasicPrinter capabilities:")
	basicPrinter.Print()

	// Photocopier can print and scan
	fmt.Println("\nPhotocopier capabilities:")
	photocopier.Print()
	photocopier.Scan()

	// AdvancedOfficePrinter can do everything
	fmt.Println("\nAdvancedOfficePrinter capabilities:")
	officePrinter.Print()
	officePrinter.Scan()
	officePrinter.Fax()
	officePrinter.Staple()

	// -------------------------
	// Demo 2: Worker Example
	// -------------------------
	fmt.Println("\n--- DEMO 2: Worker Types ---")

	humanWorker := HumanEmployee{EmployeeName: "Alice"}
	robotWorker := RobotEmployee{RobotID: "R2D2"}

	fmt.Println("\nHuman employee activities:")
	humanWorker.Work()
	humanWorker.Eat()
	humanWorker.AttendMeeting()

	fmt.Println("\nRobot employee activities:")
	robotWorker.Work()
	// robotWorker.Eat() // This won't compile! Robots don't implement Eatable.
	// This is the benefit of ISP - compile-time safety!

	// -------------------------
	// Demo 3: Functions with Small Interfaces
	// -------------------------
	fmt.Println("\n--- DEMO 3: Functions Using Small Interfaces ---")

	fmt.Println("\nPrintDocument() works with any Printer:")
	PrintDocument(basicPrinter)
	PrintDocument(photocopier)
	PrintDocument(officePrinter)

	fmt.Println("\nScanDocument() works only with Scanner types:")
	ScanDocument(photocopier)
	ScanDocument(officePrinter)
	// ScanDocument(basicPrinter) // Won't compile - BasicPrinter doesn't have Scan()

	fmt.Println("\nPrintAndScanDocument() works only with PrinterScanner types:")
	PrintAndScanDocument(photocopier)
	PrintAndScanDocument(officePrinter)
	// PrintAndScanDocument(basicPrinter) // Won't compile - missing Scan()

	fmt.Println("\nAssignWork() works with any Workable type:")
	AssignWork(humanWorker)
	AssignWork(robotWorker)

	// -------------------------
	// Summary
	// -------------------------
	fmt.Println("\n============================================")
	fmt.Println("  KEY TAKEAWAY")
	fmt.Println("============================================")
	fmt.Println("ISP ensures that types only implement what they")
	fmt.Println("actually need. This leads to cleaner, more")
	fmt.Println("maintainable, and type-safe code!")
	fmt.Println("============================================")
}
