package main

import (
	"fmt"
)

// ============================================================
// CHESS GAME - Low Level Design
// ============================================================
//
// This implementation demonstrates:
// - Polymorphism: Each piece type implements the Piece interface
// - Encapsulation: Board manages piece placement, Game manages rules
// - Single Responsibility: Each struct has a clear, focused purpose
//
// Board Layout (0-indexed):
// Row 0: Black's back rank (Rook, Knight, Bishop, Queen, King, ...)
// Row 1: Black's pawns
// Row 6: White's pawns
// Row 7: White's back rank
//
// Columns: 0-7 correspond to files a-h
// ============================================================

// ========== COLOR ENUM ==========
// Represents the two players in chess: White and Black

type Color int

const (
	White Color = iota // White = 0, moves first
	Black              // Black = 1
)

// String returns a human-readable representation of the color
func (c Color) String() string {
	if c == White {
		return "White"
	}
	return "Black"
}

// Opponent returns the opposing color
func (c Color) Opponent() Color {
	if c == White {
		return Black
	}
	return White
}

// ========== PIECE TYPE ENUM ==========
// Represents the six different types of chess pieces

type PieceType int

const (
	TypeKing   PieceType = iota // The most important piece - game ends if checkmated
	TypeQueen                   // Most powerful piece - moves like Rook + Bishop
	TypeRook                    // Moves horizontally or vertically
	TypeBishop                  // Moves diagonally
	TypeKnight                  // Moves in L-shape, can jump over pieces
	TypePawn                    // Moves forward, captures diagonally
)

// ========== POSITION ==========
// Represents a square on the chess board using row and column indices

type Position struct {
	Row int // 0-7, where 0 is Black's back rank and 7 is White's back rank
	Col int // 0-7, where 0 is column 'a' and 7 is column 'h'
}

// NewPosition creates a new Position with the given row and column
func NewPosition(row, col int) Position {
	return Position{Row: row, Col: col}
}

// IsValid checks if the position is within the 8x8 board boundaries
func (p Position) IsValid() bool {
	return p.Row >= 0 && p.Row < 8 && p.Col >= 0 && p.Col < 8
}

// String converts position to chess notation (e.g., "e4", "a1")
func (p Position) String() string {
	return fmt.Sprintf("%c%d", 'a'+p.Col, 8-p.Row)
}

// ========== PIECE INTERFACE ==========
// Piece defines the contract that all chess pieces must implement
// This enables polymorphism - we can treat all pieces uniformly

type Piece interface {
	GetColor() Color                              // Returns which player owns this piece
	GetType() PieceType                           // Returns the type of piece (King, Queen, etc.)
	GetSymbol() string                            // Returns Unicode symbol for display
	CanMove(from, to Position, board *Board) bool // Validates if this piece can make the move
	Copy() Piece                                  // Creates a deep copy of the piece
}

// ========== BASE PIECE ==========
// BasePiece contains common fields shared by all piece types
// This is embedded in each specific piece type (King, Queen, etc.)

type BasePiece struct {
	color     Color     // Which player owns this piece
	pieceType PieceType // The type of piece
	hasMoved  bool      // Tracks if piece has moved (important for castling and pawn double-move)
}

// Common getter methods for all pieces
func (bp *BasePiece) GetColor() Color    { return bp.color }
func (bp *BasePiece) GetType() PieceType { return bp.pieceType }
func (bp *BasePiece) SetMoved()          { bp.hasMoved = true }
func (bp *BasePiece) HasMoved() bool     { return bp.hasMoved }

// ========== KING ==========
// The King is the most important piece - if checkmated, the game is over
// Movement: One square in any direction (horizontal, vertical, or diagonal)

type King struct {
	BasePiece
}

// NewKing creates a new King piece of the specified color
func NewKing(color Color) *King {
	return &King{BasePiece{color: color, pieceType: TypeKing}}
}

// GetSymbol returns the Unicode chess symbol for display
func (k *King) GetSymbol() string {
	if k.color == White {
		return "♔"
	}
	return "♚"
}

// CanMove validates King movement: one square in any direction
func (k *King) CanMove(from, to Position, _ *Board) bool {
	rowDiff := abs(to.Row - from.Row)
	colDiff := abs(to.Col - from.Col)

	// King can move exactly one square in any direction
	// rowDiff+colDiff > 0 ensures the king actually moves somewhere
	return rowDiff <= 1 && colDiff <= 1 && (rowDiff+colDiff > 0)
}

// Copy creates a deep copy of this King piece
func (k *King) Copy() Piece {
	return &King{BasePiece{color: k.color, pieceType: k.pieceType, hasMoved: k.hasMoved}}
}

// ========== QUEEN ==========
// The Queen is the most powerful piece
// Movement: Any number of squares horizontally, vertically, or diagonally (combines Rook + Bishop)

type Queen struct {
	BasePiece
}

// NewQueen creates a new Queen piece of the specified color
func NewQueen(color Color) *Queen {
	return &Queen{BasePiece{color: color, pieceType: TypeQueen}}
}

// GetSymbol returns the Unicode chess symbol for display
func (q *Queen) GetSymbol() string {
	if q.color == White {
		return "♕"
	}
	return "♛"
}

// CanMove validates Queen movement: combines Rook and Bishop movement patterns
func (q *Queen) CanMove(from, to Position, _ *Board) bool {
	// Queen can move like a Rook OR like a Bishop
	return canMoveAsRook(from, to) || canMoveAsBishop(from, to)
}

// Copy creates a deep copy of this Queen piece
func (q *Queen) Copy() Piece {
	return &Queen{BasePiece{color: q.color, pieceType: q.pieceType, hasMoved: q.hasMoved}}
}

// ========== ROOK ==========
// The Rook moves horizontally or vertically any number of squares
// Also participates in castling with the King

type Rook struct {
	BasePiece
}

// NewRook creates a new Rook piece of the specified color
func NewRook(color Color) *Rook {
	return &Rook{BasePiece{color: color, pieceType: TypeRook}}
}

// GetSymbol returns the Unicode chess symbol for display
func (r *Rook) GetSymbol() string {
	if r.color == White {
		return "♖"
	}
	return "♜"
}

// CanMove validates Rook movement: horizontal or vertical lines
func (r *Rook) CanMove(from, to Position, _ *Board) bool {
	return canMoveAsRook(from, to)
}

// Copy creates a deep copy of this Rook piece
func (r *Rook) Copy() Piece {
	return &Rook{BasePiece{color: r.color, pieceType: r.pieceType, hasMoved: r.hasMoved}}
}

// ========== BISHOP ==========
// The Bishop moves diagonally any number of squares
// Each Bishop is restricted to squares of one color throughout the game

type Bishop struct {
	BasePiece
}

// NewBishop creates a new Bishop piece of the specified color
func NewBishop(color Color) *Bishop {
	return &Bishop{BasePiece{color: color, pieceType: TypeBishop}}
}

// GetSymbol returns the Unicode chess symbol for display
func (b *Bishop) GetSymbol() string {
	if b.color == White {
		return "♗"
	}
	return "♝"
}

// CanMove validates Bishop movement: diagonal lines
func (b *Bishop) CanMove(from, to Position, _ *Board) bool {
	return canMoveAsBishop(from, to)
}

// Copy creates a deep copy of this Bishop piece
func (b *Bishop) Copy() Piece {
	return &Bishop{BasePiece{color: b.color, pieceType: b.pieceType, hasMoved: b.hasMoved}}
}

// ========== KNIGHT ==========
// The Knight moves in an "L" shape: 2 squares in one direction, then 1 square perpendicular
// Special ability: Can jump over other pieces

type Knight struct {
	BasePiece
}

// NewKnight creates a new Knight piece of the specified color
func NewKnight(color Color) *Knight {
	return &Knight{BasePiece{color: color, pieceType: TypeKnight}}
}

// GetSymbol returns the Unicode chess symbol for display
func (n *Knight) GetSymbol() string {
	if n.color == White {
		return "♘"
	}
	return "♞"
}

// CanMove validates Knight movement: L-shape (2+1 or 1+2 squares)
func (n *Knight) CanMove(from, to Position, _ *Board) bool {
	rowDiff := abs(to.Row - from.Row)
	colDiff := abs(to.Col - from.Col)

	// Knight moves in L-shape: 2 squares one way, 1 square perpendicular
	// This creates 8 possible target squares from any position
	return (rowDiff == 2 && colDiff == 1) || (rowDiff == 1 && colDiff == 2)
}

// Copy creates a deep copy of this Knight piece
func (n *Knight) Copy() Piece {
	return &Knight{BasePiece{color: n.color, pieceType: n.pieceType, hasMoved: n.hasMoved}}
}

// ========== PAWN ==========
// The Pawn is the most numerous piece (8 per player)
// Movement: Forward one square (or two from starting position)
// Capture: Diagonally one square
// Special: Can be promoted when reaching the opposite end of the board

type Pawn struct {
	BasePiece
}

// NewPawn creates a new Pawn piece of the specified color
func NewPawn(color Color) *Pawn {
	return &Pawn{BasePiece{color: color, pieceType: TypePawn}}
}

// GetSymbol returns the Unicode chess symbol for display
func (p *Pawn) GetSymbol() string {
	if p.color == White {
		return "♙"
	}
	return "♟"
}

// CanMove validates Pawn movement:
// - One square forward (two from starting position)
// - Capture diagonally one square
func (p *Pawn) CanMove(from, to Position, board *Board) bool {
	// Direction: White moves UP (row decreases: -1), Black moves DOWN (row increases: +1)
	direction := 1 // Black moves down (increasing row numbers)
	if p.color == White {
		direction = -1 // White moves up (decreasing row numbers)
	}

	rowDiff := to.Row - from.Row
	colDiff := abs(to.Col - from.Col)

	// Case 1: Forward move (no capture allowed)
	if colDiff == 0 {
		// One square forward
		if rowDiff == direction && board.GetPiece(to) == nil {
			return true
		}

		// Two squares forward from starting position (first move only)
		// White pawns start at row 6, Black pawns start at row 1
		startRow := 1 // Black's starting row
		if p.color == White {
			startRow = 6 // White's starting row
		}

		if from.Row == startRow && rowDiff == 2*direction {
			// Check both the middle square and destination are empty
			middlePos := NewPosition(from.Row+direction, from.Col)
			return board.GetPiece(middlePos) == nil && board.GetPiece(to) == nil
		}
	}

	// Case 2: Diagonal capture (must capture an opponent's piece)
	if colDiff == 1 && rowDiff == direction {
		targetPiece := board.GetPiece(to)
		return targetPiece != nil && targetPiece.GetColor() != p.color
	}

	return false
}

// Copy creates a deep copy of this Pawn piece
func (p *Pawn) Copy() Piece {
	return &Pawn{BasePiece{color: p.color, pieceType: p.pieceType, hasMoved: p.hasMoved}}
}

// ========== HELPER FUNCTIONS ==========
// Movement pattern helpers used by multiple pieces

// canMoveAsRook checks if the move is along a row or column (horizontal/vertical)
func canMoveAsRook(from, to Position) bool {
	return from.Row == to.Row || from.Col == to.Col
}

// canMoveAsBishop checks if the move is along a diagonal
func canMoveAsBishop(from, to Position) bool {
	return abs(to.Row-from.Row) == abs(to.Col-from.Col)
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// sign returns -1, 0, or 1 based on the sign of x
func sign(x int) int {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
}

// ========== BOARD ==========
// Board represents the 8x8 chess board and manages piece placement
// It provides methods for piece manipulation and position checking

type Board struct {
	cells [8][8]Piece // 2D array storing pieces at each position
}

// NewBoard creates a new board with pieces in starting positions
func NewBoard() *Board {
	board := &Board{}
	board.setupPieces()
	return board
}

// setupPieces places all pieces in their standard starting positions
func (b *Board) setupPieces() {
	// Black pieces (row 0) - opponent's back rank
	b.cells[0][0] = NewRook(Black)
	b.cells[0][1] = NewKnight(Black)
	b.cells[0][2] = NewBishop(Black)
	b.cells[0][3] = NewQueen(Black)
	b.cells[0][4] = NewKing(Black)
	b.cells[0][5] = NewBishop(Black)
	b.cells[0][6] = NewKnight(Black)
	b.cells[0][7] = NewRook(Black)

	// Black pawns (row 1)
	for col := 0; col < 8; col++ {
		b.cells[1][col] = NewPawn(Black)
	}

	// White pawns (row 6)
	for col := 0; col < 8; col++ {
		b.cells[6][col] = NewPawn(White)
	}

	// White pieces (row 7) - player's back rank
	b.cells[7][0] = NewRook(White)
	b.cells[7][1] = NewKnight(White)
	b.cells[7][2] = NewBishop(White)
	b.cells[7][3] = NewQueen(White)
	b.cells[7][4] = NewKing(White)
	b.cells[7][5] = NewBishop(White)
	b.cells[7][6] = NewKnight(White)
	b.cells[7][7] = NewRook(White)
}

// GetPiece returns the piece at the given position, or nil if empty/invalid
func (b *Board) GetPiece(pos Position) Piece {
	if !pos.IsValid() {
		return nil
	}
	return b.cells[pos.Row][pos.Col]
}

// SetPiece places a piece at the given position
func (b *Board) SetPiece(pos Position, piece Piece) {
	if pos.IsValid() {
		b.cells[pos.Row][pos.Col] = piece
	}
}

// MovePiece moves a piece from one position to another
// Returns the captured piece (if any), or nil
func (b *Board) MovePiece(from, to Position) Piece {
	piece := b.GetPiece(from)
	capturedPiece := b.GetPiece(to)

	// Move the piece to the new position
	b.SetPiece(to, piece)
	b.SetPiece(from, nil)

	// Mark the piece as having moved (important for castling and pawn double-move)
	if king, ok := piece.(*King); ok {
		king.SetMoved()
	} else if rook, ok := piece.(*Rook); ok {
		rook.SetMoved()
	} else if pawn, ok := piece.(*Pawn); ok {
		pawn.SetMoved()
	}

	return capturedPiece
}

// IsPathClear checks if the path between two positions is clear (for Rook, Bishop, Queen)
// This ensures pieces cannot jump over other pieces (except Knights)
func (b *Board) IsPathClear(from, to Position) bool {
	rowDirection := sign(to.Row - from.Row)
	colDirection := sign(to.Col - from.Col)

	// Start from the square next to 'from' and check each square until 'to'
	current := NewPosition(from.Row+rowDirection, from.Col+colDirection)
	for current != to {
		if b.GetPiece(current) != nil {
			return false // Path is blocked
		}
		current.Row += rowDirection
		current.Col += colDirection
	}
	return true
}

// FindKing locates the King of the specified color on the board
func (b *Board) FindKing(color Color) Position {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := b.cells[row][col]
			if piece != nil && piece.GetType() == TypeKing && piece.GetColor() == color {
				return NewPosition(row, col)
			}
		}
	}
	// Return invalid position if king not found (should never happen in a valid game)
	return Position{Row: -1, Col: -1}
}

// IsSquareUnderAttack checks if a square is under attack by any piece of the given color
func (b *Board) IsSquareUnderAttack(pos Position, byColor Color) bool {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			piece := b.cells[row][col]
			if piece != nil && piece.GetColor() == byColor {
				attackerPos := NewPosition(row, col)
				if piece.CanMove(attackerPos, pos, b) {
					// For sliding pieces (Rook, Bishop, Queen), also verify path is clear
					if piece.GetType() != TypeKnight && piece.GetType() != TypePawn && piece.GetType() != TypeKing {
						if !b.IsPathClear(attackerPos, pos) {
							continue // Path blocked, this piece can't attack the square
						}
					}
					return true
				}
			}
		}
	}
	return false
}

// Copy creates a deep copy of the board for move simulation
func (b *Board) Copy() *Board {
	newBoard := &Board{}
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			if b.cells[row][col] != nil {
				newBoard.cells[row][col] = b.cells[row][col].Copy()
			}
		}
	}
	return newBoard
}

// Print displays the board with pieces and coordinates
func (b *Board) Print() {
	fmt.Println("\n    a   b   c   d   e   f   g   h")
	fmt.Println("  ┌───┬───┬───┬───┬───┬───┬───┬───┐")
	for row := 0; row < 8; row++ {
		fmt.Printf("%d │", 8-row)
		for col := 0; col < 8; col++ {
			piece := b.cells[row][col]
			if piece != nil {
				fmt.Printf(" %s │", piece.GetSymbol())
			} else {
				fmt.Print("   │")
			}
		}
		fmt.Printf(" %d\n", 8-row)
		if row < 7 {
			fmt.Println("  ├───┼───┼───┼───┼───┼───┼───┼───┤")
		}
	}
	fmt.Println("  └───┴───┴───┴───┴───┴───┴───┴───┘")
	fmt.Println("    a   b   c   d   e   f   g   h")
}

// ========== PLAYER ==========
// Player represents a chess player with a name and assigned color

type Player struct {
	name  string // Player's display name
	color Color  // The color pieces this player controls
}

// NewPlayer creates a new player with the given name and color
func NewPlayer(name string, color Color) *Player {
	return &Player{name: name, color: color}
}

// GetName returns the player's name
func (p *Player) GetName() string {
	return p.name
}

// GetColor returns the player's piece color
func (p *Player) GetColor() Color {
	return p.color
}

// ========== GAME STATUS ==========
// GameStatus represents the current state of the chess game

type GameStatus int

const (
	StatusOngoing   GameStatus = iota // Game is in progress
	StatusCheck                       // Current player's king is in check
	StatusCheckmate                   // Current player is checkmated (game over)
	StatusStalemate                   // Current player has no legal moves but is not in check (draw)
)

// String returns a human-readable description of the game status
func (gs GameStatus) String() string {
	switch gs {
	case StatusOngoing:
		return "Ongoing"
	case StatusCheck:
		return "Check"
	case StatusCheckmate:
		return "Checkmate"
	case StatusStalemate:
		return "Stalemate"
	default:
		return "Unknown"
	}
}

// ========== GAME ==========
// Game manages the chess game state, rules, and turn-based play
// It orchestrates interactions between the board and players

type Game struct {
	board       *Board     // The chess board with all pieces
	players     [2]*Player // Array of two players [White, Black]
	currentTurn Color      // Which player's turn it is
	status      GameStatus // Current game status (ongoing, check, checkmate, stalemate)
	moveHistory []string   // Record of all moves made in the game
}

// NewGame creates a new chess game with two players
// White player always moves first
func NewGame(whitePlayerName, blackPlayerName string) *Game {
	return &Game{
		board: NewBoard(),
		players: [2]*Player{
			NewPlayer(whitePlayerName, White),
			NewPlayer(blackPlayerName, Black),
		},
		currentTurn: White, // White moves first
		status:      StatusOngoing,
		moveHistory: make([]string, 0),
	}
}

// GetCurrentPlayer returns the player whose turn it is
func (g *Game) GetCurrentPlayer() *Player {
	if g.currentTurn == White {
		return g.players[0]
	}
	return g.players[1]
}

// GetStatus returns the current game status
func (g *Game) GetStatus() GameStatus {
	return g.status
}

// IsValidMove checks if a move is valid according to chess rules
// Returns (true, "") if valid, or (false, reason) if invalid
func (g *Game) IsValidMove(from, to Position) (bool, string) {
	// Check 1: There must be a piece at the source position
	piece := g.board.GetPiece(from)
	if piece == nil {
		return false, "No piece at source position"
	}

	// Check 2: The piece must belong to the current player
	if piece.GetColor() != g.currentTurn {
		return false, "Not your turn"
	}

	// Check 3: Cannot capture your own piece
	targetPiece := g.board.GetPiece(to)
	if targetPiece != nil && targetPiece.GetColor() == g.currentTurn {
		return false, "Cannot capture your own piece"
	}

	// Check 4: The piece must be able to make this move
	if !piece.CanMove(from, to, g.board) {
		return false, "Invalid move for this piece"
	}

	// Check 5: Path must be clear (except for knights which can jump)
	if piece.GetType() != TypeKnight && piece.GetType() != TypePawn && piece.GetType() != TypeKing {
		if !g.board.IsPathClear(from, to) {
			return false, "Path is blocked"
		}
	}

	// Check 6: Move must not leave own king in check
	if g.wouldLeaveKingInCheck(from, to) {
		return false, "Move would leave your king in check"
	}

	return true, ""
}

// wouldLeaveKingInCheck simulates a move and checks if it would leave the king in check
func (g *Game) wouldLeaveKingInCheck(from, to Position) bool {
	// Create a copy of the board to simulate the move
	simulatedBoard := g.board.Copy()

	// Simulate the move
	simulatedBoard.MovePiece(from, to)

	// Find where our king is after the move
	kingPos := simulatedBoard.FindKing(g.currentTurn)

	// Check if the king would be under attack
	return simulatedBoard.IsSquareUnderAttack(kingPos, g.currentTurn.Opponent())
}

// hasAnyLegalMove checks if the player with the given color has any legal moves
func (g *Game) hasAnyLegalMove(color Color) bool {
	for fromRow := 0; fromRow < 8; fromRow++ {
		for fromCol := 0; fromCol < 8; fromCol++ {
			fromPos := NewPosition(fromRow, fromCol)
			piece := g.board.GetPiece(fromPos)

			// Skip empty squares and opponent's pieces
			if piece == nil || piece.GetColor() != color {
				continue
			}

			// Try all possible destination squares
			for toRow := 0; toRow < 8; toRow++ {
				for toCol := 0; toCol < 8; toCol++ {
					toPos := NewPosition(toRow, toCol)

					// Skip same square
					if fromPos == toPos {
						continue
					}

					// Temporarily save current turn to check moves
					originalTurn := g.currentTurn
					g.currentTurn = color

					// Check if this move is valid
					valid, _ := g.IsValidMove(fromPos, toPos)

					// Restore original turn
					g.currentTurn = originalTurn

					if valid {
						return true
					}
				}
			}
		}
	}
	return false
}

// Move executes a move if it's valid
// Returns an error if the move is invalid
func (g *Game) Move(from, to Position) error {
	// Validate the move
	valid, reason := g.IsValidMove(from, to)
	if !valid {
		return fmt.Errorf(reason)
	}

	// Get piece info before moving (for recording the move)
	piece := g.board.GetPiece(from)

	// Execute the move
	captured := g.board.MovePiece(from, to)

	// Record the move in history
	moveStr := fmt.Sprintf("%s: %s %s→%s", g.currentTurn, piece.GetSymbol(), from, to)
	if captured != nil {
		moveStr += fmt.Sprintf(" (captured %s)", captured.GetSymbol())
	}
	g.moveHistory = append(g.moveHistory, moveStr)
	fmt.Printf("✅ %s\n", moveStr)

	// Switch to the other player's turn
	g.currentTurn = g.currentTurn.Opponent()

	// Update game status (check for check, checkmate, stalemate)
	g.updateGameStatus()

	return nil
}

// updateGameStatus checks and updates the game status after each move
func (g *Game) updateGameStatus() {
	// Find the current player's king
	kingPos := g.board.FindKing(g.currentTurn)
	opponentColor := g.currentTurn.Opponent()

	// Check if the current player's king is under attack
	isInCheck := g.board.IsSquareUnderAttack(kingPos, opponentColor)

	// Check if the current player has any legal moves
	hasLegalMoves := g.hasAnyLegalMove(g.currentTurn)

	if isInCheck {
		if hasLegalMoves {
			g.status = StatusCheck
			fmt.Printf("⚠️  %s King is in CHECK!\n", g.currentTurn)
		} else {
			g.status = StatusCheckmate
			fmt.Printf("🏆 CHECKMATE! %s wins!\n", opponentColor)
		}
	} else {
		if hasLegalMoves {
			g.status = StatusOngoing
		} else {
			g.status = StatusStalemate
			fmt.Printf("🤝 STALEMATE! The game is a draw.\n")
		}
	}
}

// PrintBoard displays the current board state
func (g *Game) PrintBoard() {
	g.board.Print()
}

// GetMoveHistory returns the list of all moves made in the game
func (g *Game) GetMoveHistory() []string {
	return g.moveHistory
}

// ========== MAIN ==========
// Entry point demonstrating the chess game functionality

func main() {
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("           ♔ CHESS GAME ♚")
	fmt.Println("═══════════════════════════════════════════")

	// Create a new game with two players
	game := NewGame("Alice", "Bob")

	// Display the initial board
	fmt.Println("\n📋 Initial Board Setup:")
	game.PrintBoard()

	// Demo: Play the Italian Game opening (a popular chess opening)
	fmt.Println("\n📍 Playing the Italian Game opening...")
	fmt.Println("─────────────────────────────────────────")

	// Define a series of moves demonstrating the opening
	// Each move is defined as [from_position, to_position]
	moves := [][2]Position{
		{NewPosition(6, 4), NewPosition(4, 4)}, // Move 1: White pawn e2→e4
		{NewPosition(1, 4), NewPosition(3, 4)}, // Move 1: Black pawn e7→e5
		{NewPosition(7, 6), NewPosition(5, 5)}, // Move 2: White knight g1→f3
		{NewPosition(0, 1), NewPosition(2, 2)}, // Move 2: Black knight b8→c6
		{NewPosition(7, 5), NewPosition(4, 2)}, // Move 3: White bishop f1→c4
		{NewPosition(0, 5), NewPosition(3, 2)}, // Move 3: Black bishop f8→c5
	}

	// Execute each move
	for _, move := range moves {
		fromPos := move[0]
		toPos := move[1]

		err := game.Move(fromPos, toPos)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		}

		// Check if game is over
		if game.GetStatus() == StatusCheckmate || game.GetStatus() == StatusStalemate {
			break
		}
	}

	// Display the final board position
	fmt.Println("\n📋 Current Board Position:")
	game.PrintBoard()

	// Print design summary
	fmt.Println("\n═══════════════════════════════════════════")
	fmt.Println("  KEY DESIGN PATTERNS & PRINCIPLES:")
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println("  1. Piece Interface    - Polymorphism")
	fmt.Println("  2. BasePiece Embed    - Code Reuse")
	fmt.Println("  3. Board Encapsulation - Single Responsibility")
	fmt.Println("  4. Game Orchestration  - Separation of Concerns")
	fmt.Println("  5. Move Validation     - Defensive Programming")
	fmt.Println("═══════════════════════════════════════════")
}
