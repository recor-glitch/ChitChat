# ChitChat API Documentation

## Authentication

All endpoints except `/auth/*` require a valid JWT token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

## WebSocket Endpoints

### WebSocket Connection

**Endpoint:** `GET /chat/ws`  
**Description:** Establishes a WebSocket connection for real-time messaging  
**Authentication:** Required (JWT token)  
**Query Parameters:**

- `user_id`: Automatically set from JWT token

**WebSocket Message Types:**

#### Subscribe to Room

```json
{
  "type": "subscribe",
  "content": "room_id"
}
```

#### Unsubscribe from Room

```json
{
  "type": "unsubscribe",
  "content": "room_id"
}
```

#### Ping (Keep Alive)

```json
{
  "type": "ping"
}
```

**WebSocket Response Types:**

#### New Message

```json
{
  "type": "new_message",
  "room_id": "room_id",
  "content": {
    "id": "message_ulid",
    "room_id": "room_id",
    "user_id": "user_id",
    "content": "message content",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Subscription Confirmation

```json
{
  "type": "subscribed",
  "content": {
    "room_id": "room_id",
    "status": "success"
  }
}
```

#### Error Message

```json
{
  "type": "error",
  "content": {
    "message": "error description"
  }
}
```

### WebSocket Statistics

**Endpoint:** `GET /chat/ws/stats`  
**Description:** Get WebSocket connection statistics  
**Authentication:** Required  
**Query Parameters:**

- `room_id` (optional): Get statistics for specific room

**Response:**

```json
{
  "message": "WebSocket statistics retrieved",
  "stats": {
    "total_connected_clients": 5,
    "room_clients": 3
  }
}
```

## REST API Endpoints

### Authentication

#### Sign Up

**Endpoint:** `POST /auth/signup`  
**Description:** Register a new user  
**Body:**

```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "secure_password"
}
```

#### Sign In

**Endpoint:** `POST /auth/signin`  
**Description:** Authenticate user and get JWT token  
**Body:**

```json
{
  "email": "john@example.com",
  "password": "secure_password"
}
```

### Chat Rooms

#### Get All Chat Rooms

**Endpoint:** `GET /chat/rooms`  
**Description:** Get all chat rooms for the authenticated user  
**Response:**

```json
{
  "message": "Chat rooms retrieved successfully",
  "rooms": [
    {
      "id": "room_id",
      "name": "General Chat",
      "type": "group",
      "created_at": "2024-01-01T00:00:00Z",
      "members": [
        {
          "id": "user_id",
          "username": "john_doe",
          "email": "john@example.com"
        }
      ]
    }
  ]
}
```

#### Create Chat Room

**Endpoint:** `POST /chat/rooms`  
**Description:** Create a new group chat room  
**Body:**

```json
{
  "name": "Project Team",
  "member_ids": ["user_id_1", "user_id_2"]
}
```

#### Get Room Information

**Endpoint:** `GET /chat/rooms/:id`  
**Description:** Get detailed information about a specific room  
**Response:**

```json
{
  "message": "Room information retrieved",
  "room": {
    "id": "room_id",
    "name": "Project Team",
    "type": "group",
    "created_at": "2024-01-01T00:00:00Z",
    "members": [
      {
        "id": "user_id",
        "username": "john_doe",
        "email": "john@example.com"
      }
    ]
  }
}
```

### Messages

#### Get Messages by Room (Cursor Pagination)

**Endpoint:** `GET /chat/rooms/:id/messages`  
**Description:** Get messages from a specific room with cursor-based pagination  
**Query Parameters:**

- `cursor` (optional): ULID cursor for pagination (default: latest messages)
- `limit` (optional): Number of messages to retrieve (default: 50, max: 100)

**Response:**

```json
{
  "message": "Messages retrieved successfully",
  "messages": [
    {
      "id": "01HXYZ1234567890ABCDEFGH",
      "room_id": "room_id",
      "user_id": "user_id",
      "content": "Hello everyone!",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "01HXYZ1234567890ABCDEFGH",
    "has_more": true
  }
}
```

#### Send Message to Room

**Endpoint:** `POST /chat/rooms/:id/messages`  
**Description:** Send a message to a specific room  
**Body:**

```json
{
  "content": "Hello everyone!"
}
```

**Response:**

```json
{
  "message": "Message sent successfully",
  "message_data": {
    "id": "01HXYZ1234567890ABCDEFGH",
    "room_id": "room_id",
    "user_id": "user_id",
    "content": "Hello everyone!",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### Room Members

#### Add Member to Room

**Endpoint:** `POST /chat/rooms/:id/members`  
**Description:** Add a user to a group chat room  
**Body:**

```json
{
  "user_id": "user_id_to_add"
}
```

#### Remove Member from Room

**Endpoint:** `DELETE /chat/rooms/:id/members`  
**Description:** Remove a user from a group chat room  
**Body:**

```json
{
  "user_id": "user_id_to_remove"
}
```

### Direct Messages

#### Send Direct Message

**Endpoint:** `POST /chat/direct/message`  
**Description:** Send a direct message to another user  
**Body:**

```json
{
  "recipient_id": "recipient_user_id",
  "content": "Hello there!"
}
```

#### Get Direct Message Room

**Endpoint:** `GET /chat/direct/room/:recipient_id`  
**Description:** Get or create a direct message room with another user  
**Response:**

```json
{
  "message": "Direct message room retrieved",
  "room": {
    "id": "room_id",
    "type": "direct",
    "created_at": "2024-01-01T00:00:00Z",
    "members": [
      {
        "id": "user_id_1",
        "username": "john_doe",
        "email": "john@example.com"
      },
      {
        "id": "user_id_2",
        "username": "jane_smith",
        "email": "jane@example.com"
      }
    ]
  }
}
```

#### Get Direct Messages (Cursor Pagination)

**Endpoint:** `GET /chat/direct/messages/:recipient_id`  
**Description:** Get direct messages between current user and recipient with cursor pagination  
**Query Parameters:**

- `cursor` (optional): ULID cursor for pagination (default: latest messages)
- `limit` (optional): Number of messages to retrieve (default: 50, max: 100)

**Response:**

```json
{
  "message": "Direct messages retrieved successfully",
  "messages": [
    {
      "id": "01HXYZ1234567890ABCDEFGH",
      "room_id": "room_id",
      "user_id": "user_id",
      "content": "Hello there!",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "next_cursor": "01HXYZ1234567890ABCDEFGH",
    "has_more": true
  }
}
```

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "Error description"
}
```

Common HTTP status codes:

- `200`: Success
- `400`: Bad Request
- `401`: Unauthorized
- `403`: Forbidden
- `404`: Not Found
- `500`: Internal Server Error

## Real-Time Messaging Flow

1. **Connect to WebSocket:** Establish WebSocket connection with JWT token
2. **Subscribe to Rooms:** Send subscribe message for each room you want to receive messages from
3. **Send Messages:** Use REST API to send messages to rooms
4. **Receive Messages:** Get real-time message updates via WebSocket
5. **Handle Disconnection:** WebSocket automatically reconnects and resubscribes

## Example WebSocket Usage

```javascript
// Connect to WebSocket
const ws = new WebSocket("ws://localhost:4000/chat/ws?user_id=your_user_id");

// Subscribe to a room
ws.send(
  JSON.stringify({
    type: "subscribe",
    content: "room_id",
  })
);

// Listen for messages
ws.onmessage = function (event) {
  const message = JSON.parse(event.data);
  if (message.type === "new_message") {
    console.log("New message:", message.content);
  }
};

// Send message via REST API
fetch("/chat/rooms/room_id/messages", {
  method: "POST",
  headers: {
    Authorization: "Bearer your_jwt_token",
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    content: "Hello everyone!",
  }),
});
```
