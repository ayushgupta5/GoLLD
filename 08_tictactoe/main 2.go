package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

/*
============================================================
TIC TAC TOE - Low-Level Design Implementation
============================================================

This implementation demonstrates:
1. Clean separation of concerns (Board, Player, Game, Controller)
2. O(1) win detection using mathematical sum technique
3. Extensible design for different board sizes (NxN)
4. Clear game state management

How O(1) Win Detection Works:
- We assign +1 to X and -1 to O
- We track the sum of each row, column, and both diagonals
- If any sum reaches +N (board size), X wins
- If any sum reaches -N, O wins
- This avoids checking all cells after each move!
============================================================
*/

// ============================================================
// SECTION 1: SYMBOL - Represents what's on each cell
// ============================================================

// Symbol represents the mark on a cell (Empty, X, or O)
type Symbol int

const (
	Empty Symbol = iota // 0 - No mark on the cell
	X                   // 1 - Player X's mark
	O                   // 2 - Player O's mark
)

// String converts Symbol to a displayable character
func (symbol Symbol) String() string {
	switch symbol {
	case X:
		return "X"
	case O:
		return "O"
	default:
		return " " // Empty cell shows as space
	}
}

// ============================================================
// SECTION 2: GAME STATUS - Tracks the current state of game
// ============================================================

// GameStatus represents the current state of the game
type GameStatus int

const (
	InProgress GameStatus = iota // Game is still ongoing
	XWins                        // Player X has won
	OWins                        // Player O has won
	Draw                         // Game ended in a draw
)

// String returns a human-readable game status message
func (status GameStatus) String() string {
	switch status {
	case XWins:
		return "X Wins!"
	case OWins:
		return "O Wins!"
	case Draw:
		return "It's a Draw!"
	default:
		return "In Progress"
	}
}

// ============================================================
// SECTION 3: PLAYER - Represents a game participant
// ============================================================

// Player holds information about a game participant
type Player struct {
	name   string // Player's display name
	symbol Symbol // The symbol this player uses (X or O)
}

// NewPlayer creates a new player with the given name and symbol
func NewPlayer(name string, symbol Symbol) *Player {
	return &Player{
		name:   name,
		symbol: symbol,
	}
}

// GetName returns the player's name
func (player *Player) GetName() string {
	return player.name
}

// GetSymbol returns the player's symbol (X or O)
func (player *Player) GetSymbol() Symbol {
	return player.symbol
}

// ============================================================
// SECTION 4: BOARD - The game grid with win detection
// ============================================================

// Board represents the NxN game grid
// It uses a clever mathematical trick for O(1) win detection:
// - Each row, column, and diagonal has a running sum
// - X adds +1, O adds -1 to the sum
// - If sum equals +N or -N, we have a winner!
type Board struct {
	size            int        // Board dimension (3 for standard 3x3)
	grid            [][]Symbol // 2D array holding all cell values
	rowSums         []int      // Running sum for each row
	columnSums      []int      // Running sum for each column
	mainDiagonalSum int        // Sum of main diagonal (top-left to bottom-right)
	antiDiagonalSum int        // Sum of anti-diagonal (top-right to bottom-left)
	totalMoves      int        // Count of moves made (for draw detection)
}

// NewBoard creates an empty board of the specified size
func NewBoard(size int) *Board {
	// Initialize the 2D grid with empty cells
	grid := make([][]Symbol, size)
	for row := 0; row < size; row++ {
		grid[row] = make([]Symbol, size)
		// Note: Go initializes int slices to 0, which equals Empty
	}

	return &Board{
		size:       size,
		grid:       grid,
		rowSums:    make([]int, size), // All zeros initially
		columnSums: make([]int, size), // All zeros initially
		// mainDiagonalSum and antiDiagonalSum default to 0
	}
}

// GetSize returns the board dimension
func (board *Board) GetSize() int {
	return board.size
}

// IsValidMove checks if a move can be made at the given position
// A move is valid if:
// 1. Row and column are within bounds (0 to size-1)
// 2. The cell is empty (not already occupied)
func (board *Board) IsValidMove(row, col int) bool {
	// Check if position is within the board boundaries
	isWithinBounds := row >= 0 && row < board.size && col >= 0 && col < board.size

	if !isWithinBounds {
		return false
	}

	// Check if the cell is empty
	isCellEmpty := board.grid[row][col] == Empty

	return isCellEmpty
}

// PlaceSymbol puts a symbol on the board and returns true if this move wins
// This is where the O(1) win detection magic happens!
//
// Time Complexity: O(1) - we only update a few sums and check them
// (Compare to O(N) if we checked entire rows/columns/diagonals)
func (board *Board) PlaceSymbol(row, col int, symbol Symbol) bool {
	// Place the symbol on the grid
	board.grid[row][col] = symbol
	board.totalMoves++

	// Determine the value to add to our sums
	// X contributes +1, O contributes -1
	valueToAdd := 1
	if symbol == O {
		valueToAdd = -1
	}

	// Update the running sums
	board.rowSums[row] += valueToAdd
	board.columnSums[col] += valueToAdd

	// Update main diagonal sum (cells where row == col)
	// Main diagonal: (0,0), (1,1), (2,2) in a 3x3 board
	isOnMainDiagonal := row == col
	if isOnMainDiagonal {
		board.mainDiagonalSum += valueToAdd
	}

	// Update anti-diagonal sum (cells where row + col == size - 1)
	// Anti-diagonal: (0,2), (1,1), (2,0) in a 3x3 board
	isOnAntiDiagonal := row+col == board.size-1
	if isOnAntiDiagonal {
		board.antiDiagonalSum += valueToAdd
	}

	// Check if this move resulted in a win
	// For X to win, a sum must equal +size (e.g., +3 for 3x3 board)
	// For O to win, a sum must equal -size (e.g., -3 for 3x3 board)
	winningSumValue := board.size
	if symbol == O {
		winningSumValue = -board.size
	}

	// Check all possible winning conditions
	hasWonByRow := board.rowSums[row] == winningSumValue
	hasWonByColumn := board.columnSums[col] == winningSumValue
	hasWonByMainDiagonal := board.mainDiagonalSum == winningSumValue
	hasWonByAntiDiagonal := board.antiDiagonalSum == winningSumValue

	return hasWonByRow || hasWonByColumn || hasWonByMainDiagonal || hasWonByAntiDiagonal
}

// IsFull checks if all cells are occupied (draw condition)
func (board *Board) IsFull() bool {
	totalCells := board.size * board.size
	return board.totalMoves == totalCells
}

// Display prints the board without coordinates (simple view)
func (board *Board) Display() {
	fmt.Println()

	for row := 0; row < board.size; row++ {
		fmt.Print("  ") // Left padding

		// Print each cell in the row
		for col := 0; col < board.size; col++ {
			fmt.Printf(" %s ", board.grid[row][col])

			// Print vertical separator between cells (not after last cell)
			if col < board.size-1 {
				fmt.Print("|")
			}
		}
		fmt.Println()

		// Print horizontal separator between rows (not after last row)
		if row < board.size-1 {
			fmt.Print("  ") // Left padding
			for col := 0; col < board.size; col++ {
				fmt.Print("---")
				if col < board.size-1 {
					fmt.Print("+")
				}
			}
			fmt.Println()
		}
	}
	fmt.Println()
}

// DisplayWithCoordinates prints the board with row/column numbers
// This helps players know which coordinates to enter
func (board *Board) DisplayWithCoordinates() {
	// Print column header numbers
	fmt.Print("\n    ") // Padding for row numbers
	for col := 0; col < board.size; col++ {
		fmt.Printf(" %d  ", col)
	}
	fmt.Println()

	for row := 0; row < board.size; row++ {
		// Print row number
		fmt.Printf(" %d ", row)

		// Print each cell in the row
		for col := 0; col < board.size; col++ {
			fmt.Printf(" %s ", board.grid[row][col])

			// Print vertical separator between cells
			if col < board.size-1 {
				fmt.Print("|")
			}
		}
		fmt.Println()

		// Print horizontal separator between rows
		if row < board.size-1 {
			fmt.Print("   ") // Padding for row numbers
			for col := 0; col < board.size; col++ {
				fmt.Print("---")
				if col < board.size-1 {
					fmt.Print("+")
				}
			}
			fmt.Println()
		}
	}
	fmt.Println()
}

// ============================================================
// SECTION 5: GAME - Orchestrates the gameplay
// ============================================================

// Game manages the overall game state and player turns
type Game struct {
	board              *Board     // The game board
	players            []*Player  // Array of two players
	currentPlayerIndex int        // Index of current player (0 or 1)
	status             GameStatus // Current game status
}

// NewGame creates a new game with the specified board size and player names
func NewGame(boardSize int, player1Name string, player2Name string) *Game {
	return &Game{
		board: NewBoard(boardSize),
		players: []*Player{
			NewPlayer(player1Name, X), // First player always gets X
			NewPlayer(player2Name, O), // Second player always gets O
		},
		currentPlayerIndex: 0,          // Player 1 (X) goes first
		status:             InProgress, // Game starts in progress
	}
}

// GetCurrentPlayer returns the player whose turn it is
func (game *Game) GetCurrentPlayer() *Player {
	return game.players[game.currentPlayerIndex]
}

// GetStatus returns the current game status
func (game *Game) GetStatus() GameStatus {
	return game.status
}

// IsOver checks if the game has ended (win or draw)
func (game *Game) IsOver() bool {
	return game.status != InProgress
}

// MakeMove attempts to place the current player's symbol at the given position
// Returns an error if the move is invalid
func (game *Game) MakeMove(row, col int) error {
	// Check if game is already over
	if game.IsOver() {
		return fmt.Errorf("game is already over - cannot make more moves")
	}

	// Validate the move
	if !game.board.IsValidMove(row, col) {
		return fmt.Errorf("invalid move: position (%d,%d) is either out of bounds or already occupied", row, col)
	}

	// Get current player and place their symbol
	currentPlayer := game.GetCurrentPlayer()
	isWinningMove := game.board.PlaceSymbol(row, col, currentPlayer.GetSymbol())

	// Check if this move won the game
	if isWinningMove {
		if currentPlayer.GetSymbol() == X {
			game.status = XWins
		} else {
			game.status = OWins
		}
		return nil // Game over - no need to switch turns
	}

	// Check if board is full (draw)
	if game.board.IsFull() {
		game.status = Draw
		return nil
	}

	// Switch to the other player's turn
	// Uses modulo to toggle between 0 and 1
	game.currentPlayerIndex = (game.currentPlayerIndex + 1) % 2

	return nil
}

// DisplayBoard shows the current board state
func (game *Game) DisplayBoard() {
	game.board.DisplayWithCoordinates()
}

// ============================================================
// SECTION 6: GAME CONTROLLER - Handles user interaction
// ============================================================

// GameController manages the interactive gameplay loop
type GameController struct {
	game        *Game         // The game being controlled
	inputReader *bufio.Reader // Reader for user input
}

// NewGameController creates a controller for the given game
func NewGameController(game *Game) *GameController {
	return &GameController{
		game:        game,
		inputReader: bufio.NewReader(os.Stdin),
	}
}

// promptForInput displays a prompt and reads user input
func (controller *GameController) promptForInput(prompt string) string {
	fmt.Print(prompt)
	input, _ := controller.inputReader.ReadString('\n')
	return strings.TrimSpace(input)
}

// parseCoordinates converts user input "row,col" into integers
// Returns an error if the format is incorrect
func (controller *GameController) parseCoordinates(input string) (int, int, error) {
	// Split input by comma
	parts := strings.Split(input, ",")

	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("please enter row,col format (e.g., 1,2)")
	}

	// Parse row number
	row, rowErr := strconv.Atoi(strings.TrimSpace(parts[0]))
	if rowErr != nil {
		return 0, 0, fmt.Errorf("invalid row number: %s", parts[0])
	}

	// Parse column number
	col, colErr := strconv.Atoi(strings.TrimSpace(parts[1]))
	if colErr != nil {
		return 0, 0, fmt.Errorf("invalid column number: %s", parts[1])
	}

	return row, col, nil
}

// StartInteractiveGame runs the main game loop for human players
func (controller *GameController) StartInteractiveGame() {
	fmt.Println("\n🎮 Welcome to Tic Tac Toe!")
	fmt.Println("Enter moves as: row,col (e.g., 0,0 for top-left corner)")
	fmt.Println("─────────────────────────────────────────")

	// Main game loop - continues until game is over
	for !controller.game.IsOver() {
		// Show current board state
		controller.game.DisplayBoard()

		// Get input from current player
		currentPlayer := controller.game.GetCurrentPlayer()
		prompt := fmt.Sprintf("%s (%s), enter your move: ",
			currentPlayer.GetName(), currentPlayer.GetSymbol())
		userInput := controller.promptForInput(prompt)

		// Parse the coordinates
		row, col, parseErr := controller.parseCoordinates(userInput)
		if parseErr != nil {
			fmt.Printf("❌ Error: %v\n", parseErr)
			continue // Ask for input again
		}

		// Try to make the move
		moveErr := controller.game.MakeMove(row, col)
		if moveErr != nil {
			fmt.Printf("❌ Error: %v\n", moveErr)
			continue // Ask for input again
		}
	}

	// Game has ended - show final state
	controller.game.DisplayBoard()
	fmt.Printf("🏆 Game Over: %s\n", controller.game.GetStatus())
}

// ============================================================
// SECTION 7: DEMO FUNCTIONS - Shows the game in action
// ============================================================

// runWinDemo demonstrates a game where X wins
func runWinDemo() {
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("       TIC TAC TOE - Win Scenario Demo")
	fmt.Println("═══════════════════════════════════════════")

	game := NewGame(3, "Alice", "Bob")

	// Predefined sequence of moves that leads to X winning
	// X wins by completing the top row: (0,0), (0,1), (0,2)
	plannedMoves := [][2]int{
		{0, 0}, // X plays top-left
		{1, 1}, // O plays center
		{0, 1}, // X plays top-middle
		{2, 2}, // O plays bottom-right
		{0, 2}, // X plays top-right -> X WINS!
	}

	// Execute each move
	for moveNumber, move := range plannedMoves {
		currentPlayer := game.GetCurrentPlayer()
		row, col := move[0], move[1]

		fmt.Printf("\n📍 Move %d: %s (%s) plays at position (%d,%d)\n",
			moveNumber+1, currentPlayer.GetName(), currentPlayer.GetSymbol(), row, col)

		err := game.MakeMove(row, col)
		if err != nil {
			fmt.Printf("Error occurred: %v\n", err)
			break
		}

		game.DisplayBoard()

		if game.IsOver() {
			break
		}
	}

	fmt.Printf("\n🏆 Final Result: %s\n", game.GetStatus())
}

// runDrawDemo demonstrates a game ending in a draw
func runDrawDemo() {
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("       TIC TAC TOE - Draw Scenario Demo")
	fmt.Println("═══════════════════════════════════════════")

	game := NewGame(3, "Alice", "Bob")

	// Predefined sequence of moves that leads to a draw
	// Neither player completes a line
	plannedMoves := [][2]int{
		{0, 0}, // X: top-left
		{0, 1}, // O: top-middle
		{0, 2}, // X: top-right
		{1, 1}, // O: center
		{1, 0}, // X: middle-left
		{1, 2}, // O: middle-right
		{2, 1}, // X: bottom-middle
		{2, 0}, // O: bottom-left
		{2, 2}, // X: bottom-right -> DRAW (all cells filled, no winner)
	}

	fmt.Println("Executing moves that lead to a draw...")
	fmt.Println()

	// Execute each move
	for _, move := range plannedMoves {
		currentPlayer := game.GetCurrentPlayer()
		row, col := move[0], move[1]

		fmt.Printf("%s plays at (%d,%d)\n", currentPlayer.GetSymbol(), row, col)

		err := game.MakeMove(row, col)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}

		if game.IsOver() {
			break
		}
	}

	game.DisplayBoard()
	fmt.Printf("\n🏆 Final Result: %s\n", game.GetStatus())
}

// printDesignSummary displays the key design decisions
func printDesignSummary() {
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("       KEY DESIGN DECISIONS")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println()
	fmt.Println("  1. O(1) Win Detection:")
	fmt.Println("     - Uses row/column/diagonal sums")
	fmt.Println("     - X adds +1, O adds -1")
	fmt.Println("     - Sum equals ±N means winner!")
	fmt.Println()
	fmt.Println("  2. Flexible Board Size:")
	fmt.Println("     - Works with any NxN board")
	fmt.Println("     - Just change the size parameter")
	fmt.Println()
	fmt.Println("  3. Clean Separation:")
	fmt.Println("     - Board: Grid management & win detection")
	fmt.Println("     - Game: Turn management & game rules")
	fmt.Println("     - Controller: User interaction")
	fmt.Println()
	fmt.Println("  4. Easy to Extend:")
	fmt.Println("     - Add AI player by implementing move logic")
	fmt.Println("     - Add undo/redo with move history")
	fmt.Println("     - Add network play with game serialization")
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════")
}

// ============================================================
// SECTION 8: MAIN ENTRY POINT
// ============================================================

func main() {
	// Run the demo scenarios to show the game works
	runWinDemo()
	runDrawDemo()

	// Print design summary for learning
	printDesignSummary()

	// ─────────────────────────────────────────────────────────
	// INTERACTIVE MODE (uncomment the lines below to play!)
	// ─────────────────────────────────────────────────────────
	// fmt.Println("\n🎮 Starting Interactive Game...")
	// game := NewGame(3, "Player 1", "Player 2")
	// controller := NewGameController(game)
	// controller.StartInteractiveGame()
}
