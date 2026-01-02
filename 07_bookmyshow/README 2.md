# BookMyShow - Movie Ticket Booking System

## 🎯 Problem Statement

Design a movie ticket booking system (like BookMyShow) that:
1. List movies in a city
2. Show available shows and seats
3. Book seats with payment
4. Handle concurrent bookings

## 🧠 Interviewer's Mindset

This tests:
1. **Entity Modeling** - Movie, Theatre, Show, Seat, Booking
2. **Concurrency** - Multiple users booking same seat
3. **Real-world Complexity** - Pricing, seat types, cancellation

## 📋 Key Entities

- **Movie**: Title, duration, genre
- **Theatre**: Name, city, screens
- **Screen**: Seats arrangement
- **Show**: Movie + Screen + Time
- **Seat**: Row, number, type, price
- **Booking**: User + Show + Seats + Payment

