# Snake and Ladder Game - Low Level Design

## 🎯 Problem Statement

Design a Snake and Ladder game that can:
1. Support multiple players
2. Handle snakes (move down) and ladders (move up)
3. Roll dice and move players
4. Determine winner when someone reaches 100

## 🧠 Interviewer's Mindset

This is a great OOP modeling problem. Interviewers evaluate:
1. **Entity Identification**: Board, Player, Snake, Ladder, Dice
2. **Game State Management**: Whose turn, current positions
3. **Rules Implementation**: Movement logic, win condition
4. **Extensibility**: Can you add power-ups, multiple dice?

## ❓ Questions to Ask

1. Board size? (Standard 100 squares)
2. Number of players?
3. Multiple dice? Special dice?
4. Can snake/ladder chains happen?
5. What if you land beyond 100? (Bounce back or stay?)

## 📋 Key Entities

- **Board**: Contains snakes, ladders, size
- **Player**: Name, current position
- **Snake**: Start (head), end (tail) - moves DOWN
- **Ladder**: Start (bottom), end (top) - moves UP
- **Dice**: Roll random number
- **Game**: Orchestrates gameplay

## 🎨 Design Patterns Used

1. **Strategy Pattern**: Different dice types
2. **Observer Pattern**: Notify on player move (optional)
3. **Factory Pattern**: Create game with config

