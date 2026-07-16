# Auction System

Real-time concurrent auction backend built in Go.

Primary goal: hands-on learning of concurrency, race conditions, and mutex-based safe operations under load.

## Stack
Go · Redis · PostgreSQL · Docker

## Key Features
- Multiple concurrent auctions with per-auction mutex locking
- Real-time bid updates via WebSocket
- JWT authentication with protected routes
- Auction history persisted to PostgreSQL
- Load tested with `hey`: 14.5K read / 9.2K write req/sec, sub-10ms p95 latency
- Race condition verified with Go's `-race` flag

## Running Locally

**Without Docker**
```bash
git clone <repo>
cp .env.example .env   # fill in your values
go run main.go
```

**With Docker**
```bash
git clone <repo>
cp .env.example .env.docker   # fill in your values
make up        # start server
make migrate   # run once to set up DB
make down      # stop server
make reset     # reset all containers
```

## Endpoints

| Route | Auth | Description |
|---|---|---|
| `POST /register` | No | Register user |
| `POST /login` | No | Login, returns JWT |
| `POST /setbid` | Yes | Create/reset an auction |
| `POST /bid` | Yes | Place a bid |
| `GET /getbid` | No | Get current auction state |
| `GET /ws` | No | WebSocket for live bid updates |
| `GET /listactive` | No | List active auctions |
| `GET /listhistory` | No | Auction end history |
| `GET /mybid` | Yes | View your bid info |
