# Chess Game - Low Level Design

## 🎯 Problem Statement

Design a Chess game that:
1. Supports all standard chess pieces
2. Validates legal moves for each piece
3. Detects check and checkmate
4. Handles turns between players

## 🧠 Interviewer's Mindset

This is a COMPLEX problem testing:
1. **OOP Modeling** - Piece hierarchy
2. **Move Validation** - Each piece has different rules
3. **Game State** - Check, Checkmate detection
4. **Polymorphism** - All pieces share interface

## 📋 Key Entities

- **Board**: 8x8 grid of cells
- **Piece**: King, Queen, Rook, Bishop, Knight, Pawn
- **Player**: White/Black
- **Move**: From position to position
- **Game**: Orchestrates gameplay

