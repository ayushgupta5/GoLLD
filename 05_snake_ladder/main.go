package main

import (
	"fmt"
	"math/rand"
)

// ============================================================
// SNAKE AND LADDER GAME - Complete LLD Implementation
// ============================================================
//
// This game demonstrates several important concepts:
// 1. OOP Modeling - Using structs and interfaces to model real-world entities
// 2. Strategy Pattern - Different dice implementations can be swapped easily
// 3. Game State Management - Tracking game progress through states
//
// How the game works:
// - Players take turns rolling a dice and move forward by that many positions
// - If a player lands on a snake's head, they slide down to its tail
// - If a player lands on a ladder's bottom, they climb up to its top
// - First player to reach exactly position 100 wins!
// ============================================================

// ========== DICE (Strategy Pattern) ==========
//
// The Strategy Pattern allows us to define a family of algorithms (dice types),
// encapsulate each one, and make them interchangeable.
// This lets us easily switch between different dice types without changing the game logic.

// Dice interface defines what any dice must be able to do
// Any struct that implements these methods can be used as a Dice
type Dice interface {
	Roll() int        // Roll the dice and return the value
	GetMaxValue() int // Get the maximum possible value from this dice
}

// StandardDice represents a regular dice with configurable number of sides
// By default, it's a 6-sided dice (values 1-6)
type StandardDice struct {
	sides int // Number of sides on the dice
}

// NewStandardDice creates a new standard 6-sided dice
func NewStandardDice() *StandardDice {
	return &StandardDice{sides: 6}
}

// Roll generates a random number between 1 and the number of sides
// rand.Intn(n) returns a value from 0 to n-1, so we add 1 to get 1 to n
func (d *StandardDice) Roll() int {
	return rand.Intn(d.sides) + 1
}

// GetMaxValue returns the maximum possible roll value
func (d *StandardDice) GetMaxValue() int {
	return d.sides
}

// BiasedDice always returns the same value - useful for testing
// Example: Use this when you want to test specific game scenarios
type BiasedDice struct {
	fixedValue int // The value this dice will always return
}

// NewBiasedDice creates a dice that always returns the specified value
func NewBiasedDice(value int) *BiasedDice {
	return &BiasedDice{fixedValue: value}
}

// Roll always returns the fixed value (useful for predictable testing)
func (d *BiasedDice) Roll() int {
	return d.fixedValue
}

// GetMaxValue returns the fixed value since that's the only possible outcome
func (d *BiasedDice) GetMaxValue() int {
	return d.fixedValue
}

// DoubleDice represents two dice rolled together
// This gives a range of 2-12 (both dice combined)
type DoubleDice struct {
	firstDice  *StandardDice // First 6-sided dice
	secondDice *StandardDice // Second 6-sided dice
}

// NewDoubleDice creates a pair of standard dice
func NewDoubleDice() *DoubleDice {
	return &DoubleDice{
		firstDice:  NewStandardDice(),
		secondDice: NewStandardDice(),
	}
}

// Roll returns the sum of rolling both dice
func (d *DoubleDice) Roll() int {
	return d.firstDice.Roll() + d.secondDice.Roll()
}

// GetMaxValue returns 12 (maximum: 6 from each dice)
func (d *DoubleDice) GetMaxValue() int {
	return d.firstDice.GetMaxValue() + d.secondDice.GetMaxValue()
}

// ========== SNAKE ==========

// Snake represents a snake on the board
// Head (start) is always > Tail (end) - moves player DOWN
// Example: Snake with head at 99 and tail at 54 means landing on 99 drops you to 54
type Snake struct {
	head int // Where snake's head is (higher position) - player lands here
	tail int // Where snake's tail is (lower position) - player slides to here
}

// NewSnake creates a new snake with validation
// Returns error if head position is not greater than tail position
func NewSnake(head, tail int) (*Snake, error) {
	if head <= tail {
		return nil, fmt.Errorf("snake head (%d) must be greater than tail (%d)", head, tail)
	}
	return &Snake{head: head, tail: tail}, nil
}

// GetHead returns the snake's head position (where player lands)
func (s *Snake) GetHead() int {
	return s.head
}

// GetTail returns the snake's tail position (where player slides to)
func (s *Snake) GetTail() int {
	return s.tail
}

// String returns a formatted description of the snake
func (s *Snake) String() string {
	slidesDown := s.head - s.tail
	return fmt.Sprintf("🐍 Snake: %d → %d (slides down %d positions)", s.head, s.tail, slidesDown)
}

// ========== LADDER ==========

// Ladder represents a ladder on the board
// Start (bottom) is always < End (top) - moves player UP
// Example: Ladder from 6 to 25 means landing on 6 climbs you to 25
type Ladder struct {
	start int // Bottom of ladder (lower position) - player lands here
	end   int // Top of ladder (higher position) - player climbs to here
}

// NewLadder creates a new ladder with validation
// Returns error if start position is not less than end position
func NewLadder(start, end int) (*Ladder, error) {
	if start >= end {
		return nil, fmt.Errorf("ladder start (%d) must be less than end (%d)", start, end)
	}
	return &Ladder{start: start, end: end}, nil
}

// GetStart returns the ladder's starting position (where player lands)
func (l *Ladder) GetStart() int {
	return l.start
}

// GetEnd returns the ladder's ending position (where player climbs to)
func (l *Ladder) GetEnd() int {
	return l.end
}

// String returns a formatted description of the ladder
func (l *Ladder) String() string {
	climbsUp := l.end - l.start
	return fmt.Sprintf("🪜 Ladder: %d → %d (climbs up %d positions)", l.start, l.end, climbsUp)
}

// ========== PLAYER ==========

// Player represents a game player with their current position on the board
type Player struct {
	id       int    // Unique identifier for the player
	name     string // Display name of the player
	position int    // Current position on the board (0 means not started yet)
}

// NewPlayer creates a new player starting at position 0 (before the board)
func NewPlayer(id int, name string) *Player {
	return &Player{
		id:       id,
		name:     name,
		position: 0, // Start before board (position 0)
	}
}

// GetID returns the player's unique identifier
func (p *Player) GetID() int {
	return p.id
}

// GetName returns the player's display name
func (p *Player) GetName() string {
	return p.name
}

// GetPosition returns the player's current position on the board
func (p *Player) GetPosition() int {
	return p.position
}

// SetPosition updates the player's position on the board
func (p *Player) SetPosition(pos int) {
	p.position = pos
}

// String returns a formatted description of the player
func (p *Player) String() string {
	return fmt.Sprintf("%s (Position: %d)", p.name, p.position)
}

// ========== BOARD ==========

// Board represents the game board containing snakes and ladders
type Board struct {
	size    int             // Total number of squares on the board (typically 100)
	snakes  map[int]*Snake  // Map of position -> snake (key is snake's head position)
	ladders map[int]*Ladder // Map of position -> ladder (key is ladder's start position)
}

// NewBoard creates a new board with the specified size
func NewBoard(size int) *Board {
	return &Board{
		size:    size,
		snakes:  make(map[int]*Snake),
		ladders: make(map[int]*Ladder),
	}
}

// GetSize returns the total number of squares on the board
func (b *Board) GetSize() int {
	return b.size
}

// AddSnake adds a snake to the board with validation
// head: position where snake's head is (player lands here)
// tail: position where snake's tail is (player slides to here)
func (b *Board) AddSnake(head, tail int) error {
	// Validate positions are within board boundaries
	if head > b.size || tail < 1 {
		return fmt.Errorf("snake positions must be within board (1-%d)", b.size)
	}

	// Check if a snake already exists at this position
	if _, exists := b.snakes[head]; exists {
		return fmt.Errorf("snake already exists at position %d", head)
	}

	// Check if a ladder already exists at this position (can't have both)
	if _, exists := b.ladders[head]; exists {
		return fmt.Errorf("ladder already exists at position %d", head)
	}

	// Create and add the snake
	snake, err := NewSnake(head, tail)
	if err != nil {
		return err
	}
	b.snakes[head] = snake
	return nil
}

// AddLadder adds a ladder to the board with validation
// start: position where ladder starts (player lands here)
// end: position where ladder ends (player climbs to here)
func (b *Board) AddLadder(start, end int) error {
	// Validate positions are within board boundaries
	if start < 1 || end > b.size {
		return fmt.Errorf("ladder positions must be within board (1-%d)", b.size)
	}

	// Check if a ladder already exists at this position
	if _, exists := b.ladders[start]; exists {
		return fmt.Errorf("ladder already exists at position %d", start)
	}

	// Check if a snake already exists at this position (can't have both)
	if _, exists := b.snakes[start]; exists {
		return fmt.Errorf("snake already exists at position %d", start)
	}

	// Create and add the ladder
	ladder, err := NewLadder(start, end)
	if err != nil {
		return err
	}
	b.ladders[start] = ladder
	return nil
}

// GetNewPosition checks if the given position has a snake or ladder
// Returns the final position and a message describing what happened
func (b *Board) GetNewPosition(position int) (int, string) {
	// Check for snake at this position
	if snake, exists := b.snakes[position]; exists {
		return snake.GetTail(), fmt.Sprintf("🐍 Oops! Snake bite! Sliding down from %d to %d", position, snake.GetTail())
	}

	// Check for ladder at this position
	if ladder, exists := b.ladders[position]; exists {
		return ladder.GetEnd(), fmt.Sprintf("🪜 Yay! Climbed ladder from %d to %d", position, ladder.GetEnd())
	}

	// No snake or ladder - position stays the same
	return position, ""
}

// IsWinningPosition checks if the given position is the winning position
func (b *Board) IsWinningPosition(position int) bool {
	return position == b.size
}

// PrintBoard displays the board configuration (snakes and ladders)
func (b *Board) PrintBoard() {
	fmt.Printf("\n📋 Board Size: %d\n", b.size)
	fmt.Println("\nSnakes:")
	for _, snake := range b.snakes {
		fmt.Printf("  %s\n", snake)
	}
	fmt.Println("\nLadders:")
	for _, ladder := range b.ladders {
		fmt.Printf("  %s\n", ladder)
	}
}

// ========== GAME ==========

// GameState represents the current state of the game
// Using iota for auto-incrementing constants is a Go idiom
type GameState int

const (
	GameStateNotStarted GameState = iota // Game hasn't started yet (value: 0)
	GameStateInProgress                  // Game is currently being played (value: 1)
	GameStateFinished                    // Game has ended with a winner (value: 2)
)

// Game orchestrates the snake and ladder game
// It manages the board, players, dice, and game flow
type Game struct {
	board       *Board    // The game board with snakes and ladders
	players     []*Player // List of players in the game
	dice        Dice      // The dice used for rolling (can be any Dice implementation)
	currentTurn int       // Index of the player whose turn it is
	state       GameState // Current state of the game
	winner      *Player   // The winning player (nil until game ends)
}

// GameConfig holds all the configuration options for creating a new game
// This pattern makes it easy to customize game setup
type GameConfig struct {
	BoardSize   int      // Size of the board (typically 100)
	Snakes      [][2]int // Array of [head, tail] pairs for snakes
	Ladders     [][2]int // Array of [start, end] pairs for ladders
	PlayerNames []string // Names of all players
	Dice        Dice     // Optional: Custom dice (defaults to StandardDice)
}

// NewGame creates a new game with the given configuration
// Returns an error if the configuration is invalid (e.g., invalid snake/ladder positions)
func NewGame(config GameConfig) (*Game, error) {
	// Validate that we have at least one player
	if len(config.PlayerNames) == 0 {
		return nil, fmt.Errorf("at least one player is required")
	}

	// Validate board size
	if config.BoardSize < 10 {
		return nil, fmt.Errorf("board size must be at least 10")
	}

	// Create the game board with the specified size
	board := NewBoard(config.BoardSize)

	// Add all snakes to the board
	// Each snake is defined as [head, tail] where head > tail
	for _, snakeConfig := range config.Snakes {
		if err := board.AddSnake(snakeConfig[0], snakeConfig[1]); err != nil {
			return nil, fmt.Errorf("failed to add snake: %w", err)
		}
	}

	// Add all ladders to the board
	// Each ladder is defined as [start, end] where start < end
	for _, ladderConfig := range config.Ladders {
		if err := board.AddLadder(ladderConfig[0], ladderConfig[1]); err != nil {
			return nil, fmt.Errorf("failed to add ladder: %w", err)
		}
	}

	// Create player objects with unique IDs starting from 1
	players := make([]*Player, len(config.PlayerNames))
	for index, playerName := range config.PlayerNames {
		players[index] = NewPlayer(index+1, playerName)
	}

	// Use provided dice or default to standard 6-sided dice
	gameDice := config.Dice
	if gameDice == nil {
		gameDice = NewStandardDice()
	}

	return &Game{
		board:       board,
		players:     players,
		dice:        gameDice,
		currentTurn: 0, // First player (index 0) starts
		state:       GameStateNotStarted,
		winner:      nil,
	}, nil
}

// Start begins the game
func (g *Game) Start() {
	g.state = GameStateInProgress
	fmt.Println("\n🎮 Game Started!")
	g.board.PrintBoard()
	fmt.Println("\n══════════════════════════════════════════════════")
}

// GetCurrentPlayer returns the player whose turn it is
func (g *Game) GetCurrentPlayer() *Player {
	return g.players[g.currentTurn]
}

// PlayTurn executes one turn for the current player
// Returns true if the game has ended (current player won), false otherwise
func (g *Game) PlayTurn() bool {
	// Safety check: only play if game is in progress
	if g.state != GameStateInProgress {
		fmt.Println("Game is not in progress!")
		return false
	}

	// Get the player whose turn it is
	currentPlayer := g.GetCurrentPlayer()

	// Step 1: Roll the dice
	diceValue := g.dice.Roll()
	fmt.Printf("\n🎲 %s rolled: %d\n", currentPlayer.GetName(), diceValue)

	// Step 2: Calculate the new position
	currentPosition := currentPlayer.GetPosition()
	newPosition := currentPosition + diceValue

	// Step 3: Check if the roll would exceed the board size
	// In Snake and Ladder, you need EXACT roll to reach the winning position
	if newPosition > g.board.GetSize() {
		fmt.Printf("   %s stays at %d (rolled too high, need exact roll to win)\n",
			currentPlayer.GetName(), currentPosition)
	} else {
		// Step 4: Move the player to the new position
		currentPlayer.SetPosition(newPosition)
		fmt.Printf("   %s moved to %d\n", currentPlayer.GetName(), newPosition)

		// Step 5: Check if landed on a snake or ladder
		finalPosition, eventMessage := g.board.GetNewPosition(newPosition)
		if eventMessage != "" {
			fmt.Printf("   %s\n", eventMessage)
			currentPlayer.SetPosition(finalPosition)
		}

		// Step 6: Check if player has won (reached exactly position 100)
		if g.board.IsWinningPosition(currentPlayer.GetPosition()) {
			g.state = GameStateFinished
			g.winner = currentPlayer
			fmt.Printf("\n🏆 %s WINS! 🎉\n", currentPlayer.GetName())
			return true
		}
	}

	// Step 7: Move to next player's turn
	// Using modulo to cycle through players: 0 -> 1 -> 2 -> 0 -> 1 -> ...
	g.currentTurn = (g.currentTurn + 1) % len(g.players)
	return false
}

// PlayGame plays the entire game until someone wins
// Returns the winning player, or nil if game reaches maximum turns (safety limit)
func (g *Game) PlayGame() *Player {
	g.Start()

	// Safety limit to prevent infinite loops in edge cases
	const maxTurns = 1000
	turnCount := 0

	// Keep playing turns until game ends or we hit the safety limit
	for g.state == GameStateInProgress && turnCount < maxTurns {
		gameEnded := g.PlayTurn()
		if gameEnded {
			break
		}
		turnCount++
	}

	// Warn if game didn't finish naturally
	if turnCount >= maxTurns {
		fmt.Println("⚠️ Game ended due to turn limit!")
	}

	return g.winner
}

// GetStatus returns current game status
func (g *Game) GetStatus() string {
	status := "\n╔══════════════════════════════════════╗\n"
	status += "║         GAME STATUS                  ║\n"
	status += "╠══════════════════════════════════════╣\n"

	for _, p := range g.players {
		marker := "  "
		if g.state == GameStateInProgress && p == g.GetCurrentPlayer() {
			marker = "→ "
		}
		status += fmt.Sprintf("║ %s%s\n", marker, p)
	}

	status += "╚══════════════════════════════════════╝\n"
	return status
}

// ========== MAIN ==========

func main() {
	// Note: In Go 1.20+, random number generation is automatically seeded
	// No need to call rand.Seed() anymore - it's deprecated

	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("       SNAKE AND LADDER - LLD DEMO")
	fmt.Println("═══════════════════════════════════════════")

	// Define game configuration
	// - BoardSize: Standard 100-square board
	// - Snakes: Each pair is [head, tail] - landing on head drops you to tail
	// - Ladders: Each pair is [start, end] - landing on start lifts you to end
	config := GameConfig{
		BoardSize: 100,
		Snakes: [][2]int{
			{99, 54}, // Near win - cruel snake!
			{70, 55},
			{52, 42},
			{25, 2},
			{95, 72},
		},
		Ladders: [][2]int{
			{6, 25},
			{11, 40},
			{60, 85},
			{46, 90},
			{17, 69},
		},
		PlayerNames: []string{"Alice", "Bob", "Charlie"},
		Dice:        NewStandardDice(), // Can swap with NewDoubleDice() or NewBiasedDice()
	}

	// Create and play game
	game, err := NewGame(config)
	if err != nil {
		fmt.Printf("Failed to create game: %v\n", err)
		return
	}

	// Play complete game
	winner := game.PlayGame()

	if winner != nil {
		fmt.Println("\n═══════════════════════════════════════════")
		fmt.Printf("  🎊 Congratulations %s! 🎊\n", winner.GetName())
		fmt.Println("═══════════════════════════════════════════")
	}

	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  KEY DESIGN DECISIONS:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. Dice interface - Strategy pattern")
	fmt.Println("  2. Board encapsulates snake/ladder logic")
	fmt.Println("  3. Game orchestrates the flow")
	fmt.Println("  4. Easy to extend (power-ups, etc.)")
	fmt.Println("═══════════════════════════════════════════")
}
