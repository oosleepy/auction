# Auction System

A real-time auction backend built in Go. 
Primary goal was learning of concurrency, race conditions, and mutex-based safe operations under load.

## Stack
Go, Redis, PostgreSQL

## Key Features
- Multiple concurrent auctions with per-auction mutex locking
- Real-time bid updates via WebSocket
- JWT authentication with protected routes
- Auction history persisted to PostgreSQL
- Race condition test + load tested with `hey`: 14.5K read / 9.2K write req/sec, sub-10ms p95 latency

## Endpoints

| Route | Description |
|---|---|
| `/register` |Register user |
| `/login` |Login, returns JWT |
| `/setbid` |Create/reset an auction |
| `/bid` |Place a bid |
| `/getbid` |Get current auction state |
| `/ws` |WebSocket for live bid updates |
| `/listactive` |List active auctions |
| `/listhistory` |Auction end history |
| `/mybid` |View your bid info |
