# CaveJoz API

A social media backend written in Go, with real-time notifications and direct messaging over WebSocket.

Built as a learning project to explore WebSocket concurrency patterns in Go — connection lifecycle management, safe concurrent writes, keepalive, and bidirectional protocols.

## Features

- **Auth** — JWT access tokens with refresh token rotation
- **Profiles** — bio, avatar, and cover photo uploads
- **Friendships** — request / accept / reject / unfriend / block, with a self-referencing many-to-many relationship
- **Posts** — CRUD with `public` / `friends` visibility enforced at read time
- **Feed** — chronological, paginated, own posts plus accepted friends'
- **Likes & comments**
- **Notifications** — a single table covering four event types, delivered live over WebSocket with a REST fallback
- **Direct messaging** — real-time chat with typing indicators, read receipts, and cursor-paginated history

## Stack

| | |
|---|---|
| Language | Go |
| Framework | [Gin](https://github.com/gin-gonic/gin) |
| Database | MySQL |
| WebSocket | [gorilla/websocket](https://github.com/gorilla/websocket) |
| Auth | JWT |

## Project structure

```
cavejoz-go-api/
├── databases/     # DB connection and initialization
├── hub/           # WebSocket hub, client, read/write pumps
├── middlewares/   # Authenticate middleware
├── models/        # Data models and queries
├── routes/        # Route registration and handlers
├── utils/         # Token generation/verification, helpers
└── main.go
```

## Getting started

### Requirements

- Go 1.21+
- MySQL 8.0+

### Setup

```bash
git clone https://github.com/BatJoz21/cavejoz-go-api.git
cd cavejoz-go-api
go mod download
```

Create a `.env` file in the project root:

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=your_user
DB_PASSWORD=your_password
DB_NAME=cavejoz

JWT_SECRET=your_secret_key
```

Import the schema, then run:

```bash
go run main.go
```

The API listens on `:8080`.

## API

Public routes are at the root; authenticated routes require an `Authorization: Bearer <token>` header.

## WebSocket protocol

### Connecting

Browsers can't set custom headers on a WebSocket handshake, so the access token can't travel in an `Authorization` header. Instead:

1. `POST /ws-ticket` with the access token — returns an opaque, single-use ticket valid for 30 seconds
2. `GET /api/v1/ws?ticket=<ticket>` — the ticket is consumed on redemption

Tickets carry no claims and are useless once spent, so exposure in a URL or a server log is low-risk.

### Server → client

```json
{ "type": "notification", "notification": { ... } }
{ "type": "message", "message": { ... } }
{ "type": "typing", "conversation_id": 5, "user_id": 3 }
```

### Client → server

```json
{ "type": "send_message", "conversation_id": 5, "content": "hello" }
{ "type": "typing", "conversation_id": 5 }
```

The sender's identity always comes from the authenticated connection, never from the payload. Every inbound frame is authorized independently — an open socket proves identity, not permission.

### Design notes

**Connection registry.** The hub holds `map[int64]map[*Client]bool` — a set of connections per user, so a user signed in across several tabs receives every push on all of them. Guarded by an `RWMutex`, since lookups vastly outnumber connects and disconnects.

**One writer per connection.** `*websocket.Conn` permits a single concurrent writer. Rather than locking around each write, every connection gets a buffered channel and a dedicated `WritePump` goroutine — the only thing that ever writes to that socket. Pushes queue and return immediately, so a slow client can't block anything else. A full buffer drops the message rather than blocking.

**Keepalive.** The server pings every 54 seconds against a 60-second read deadline. Browsers reply with a pong automatically, which resets the deadline. A client whose network vanished without a clean close is reaped when the deadline expires, rather than lingering in the hub forever.

**Persistence before delivery.** Notifications and messages are written to the database *before* the push, and the push is best-effort — if the recipient is offline, the row is still there for their next fetch. The socket is a fast path, not the source of truth.

**Read tracking.** Conversations store a per-participant watermark (`last_read_id`) rather than a flag on every message. Marking a thread read is one row written regardless of how many messages were unread, and "unread" is a range comparison against the existing index.

## License

MIT
