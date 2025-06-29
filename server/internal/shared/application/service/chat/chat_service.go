package chat

import (
	"ChitChat/internal/shared/application/service/db"
	"ChitChat/internal/shared/application/service/websocket"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"
)

type ChatService struct {
	db        *pgxpool.Pool
	wsService *websocket.WebSocketService
}

func NewChatService(database *pgxpool.Pool, wsService *websocket.WebSocketService) *ChatService {
	return &ChatService{
		db:        database,
		wsService: wsService,
	}
}

// CreateRoom creates a new chat room (group or direct) and adds members
func (s *ChatService) CreateRoom(room *db.Room, userIDs []string) error {
	ctx := context.Background()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get tenant ID from the first user
	var tenantID string
	err = tx.QueryRow(ctx, "SELECT tenant_id FROM users WHERE id = $1", userIDs[0]).Scan(&tenantID)
	if err != nil {
		return err
	}

	// Create the room
	err = tx.QueryRow(ctx, `
		INSERT INTO rooms (id, tenant_id, name, type) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, tenant_id, name, type, created_at
	`, room.ID, tenantID, room.Name, room.Type).Scan(&room.ID, &room.TenantID, &room.Name, &room.Type, &room.CreatedAt)
	if err != nil {
		return err
	}

	// Add all users as room members
	for _, userID := range userIDs {
		_, err = tx.Exec(ctx, "INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)", room.ID, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// FindOrCreateDirectRoom finds an existing direct message room between two users or creates a new one
func (s *ChatService) FindOrCreateDirectRoom(userID1, userID2 string) (*db.Room, error) {
	ctx := context.Background()

	// First, try to find an existing direct room between these users
	var room db.Room
	err := s.db.QueryRow(ctx, `
		SELECT r.id, r.tenant_id, r.name, r.type, r.created_at
		FROM rooms r
		JOIN room_members rm1 ON r.id = rm1.room_id
		JOIN room_members rm2 ON r.id = rm2.room_id
		WHERE r.type = 'direct' 
		AND rm1.user_id = $1 
		AND rm2.user_id = $2
		LIMIT 1
	`, userID1, userID2).Scan(&room.ID, &room.TenantID, &room.Name, &room.Type, &room.CreatedAt)

	if err == nil {
		// Room found, return it
		return &room, nil
	}

	// Room not found, create a new one
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get tenant ID from one of the users
	var tenantID string
	err = tx.QueryRow(ctx, "SELECT tenant_id FROM users WHERE id = $1", userID1).Scan(&tenantID)
	if err != nil {
		return nil, err
	}

	// Create new room
	roomID := uuid.New().String()
	err = tx.QueryRow(ctx, `
		INSERT INTO rooms (id, tenant_id, name, type) 
		VALUES ($1, $2, 'Direct Message', 'direct') 
		RETURNING id, tenant_id, name, type, created_at
	`, roomID, tenantID).Scan(&room.ID, &room.TenantID, &room.Name, &room.Type, &room.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Add both users as room members
	_, err = tx.Exec(ctx, "INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)", roomID, userID1)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, "INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)", roomID, userID2)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &room, nil
}

// SaveMessage saves a message to the database
func (s *ChatService) SaveMessage(message *db.Message) error {
	ctx := context.Background()

	// Generate ULID for message ID if not provided
	if message.ID == "" {
		message.ID = s.generateULID()
	}

	// Get tenant ID from the user
	var tenantID string
	err := s.db.QueryRow(ctx, "SELECT tenant_id FROM users WHERE id = $1", message.UserID).Scan(&tenantID)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO messages (id, tenant_id, room_id, user_id, content) 
		VALUES ($1, $2, $3, $4, $5)
	`, message.ID, tenantID, message.RoomID, message.UserID, message.Content)

	if err != nil {
		return err
	}

	// Broadcast message via WebSocket if service is available
	if s.wsService != nil {
		fmt.Printf("Broadcasting message to room %s: %s\n", message.RoomID, message.Content)
		s.wsService.BroadcastMessage(message.RoomID, message)
	} else {
		fmt.Printf("WebSocket service not available for broadcasting\n")
	}

	return nil
}

// generateULID creates a new ULID for message IDs
func (s *ChatService) generateULID() string {
	return ulid.Make().String()
}

// VerifyUserIsRoomMember checks if a user is a member of a specific room
func (s *ChatService) VerifyUserIsRoomMember(userID, roomID string) (bool, error) {
	ctx := context.Background()

	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM room_members 
			WHERE room_id = $1 AND user_id = $2
		)
	`, roomID, userID).Scan(&exists)

	return exists, err
}

// GetMessagesByRoom retrieves messages for a specific room with pagination
func (s *ChatService) GetMessagesByRoom(roomID string, page, limit int) ([]db.Message, error) {
	ctx := context.Background()

	offset := (page - 1) * limit

	rows, err := s.db.Query(ctx, `
		SELECT m.id, m.tenant_id, m.room_id, m.user_id, m.content, m.sent_at
		FROM messages m
		WHERE m.room_id = $1
		ORDER BY m.id DESC
		LIMIT $2 OFFSET $3
	`, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []db.Message
	for rows.Next() {
		var msg db.Message
		err := rows.Scan(&msg.ID, &msg.TenantID, &msg.RoomID, &msg.UserID, &msg.Content, &msg.SentAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// GetMessagesByRoomCursor retrieves messages for a specific room using cursor-based pagination
func (s *ChatService) GetMessagesByRoomCursor(roomID string, cursor string, limit int) ([]db.Message, string, error) {
	ctx := context.Background()

	var rows pgx.Rows
	var err error

	if cursor == "" {
		// First page - get the most recent messages using ULID sorting
		rows, err = s.db.Query(ctx, `
			SELECT m.id, m.tenant_id, m.room_id, m.user_id, m.content, m.sent_at
			FROM messages m
			WHERE m.room_id = $1
			ORDER BY m.id DESC
			LIMIT $2
		`, roomID, limit+1) // Get one extra to determine if there are more pages
	} else {
		// Subsequent pages - use ULID cursor for efficient pagination
		rows, err = s.db.Query(ctx, `
			SELECT m.id, m.tenant_id, m.room_id, m.user_id, m.content, m.sent_at
			FROM messages m
			WHERE m.room_id = $1
			AND m.id < $2
			ORDER BY m.id DESC
			LIMIT $3
		`, roomID, cursor, limit+1) // Get one extra to determine if there are more pages
	}

	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var messages []db.Message
	for rows.Next() {
		var msg db.Message
		err := rows.Scan(&msg.ID, &msg.TenantID, &msg.RoomID, &msg.UserID, &msg.Content, &msg.SentAt)
		if err != nil {
			return nil, "", err
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	// Determine next cursor
	var nextCursor string
	if len(messages) > limit {
		// Remove the extra message and set next cursor
		nextCursor = messages[limit-1].ID
		messages = messages[:limit]
	}

	return messages, nextCursor, nil
}

// GetMessagesByRoomCursorForward retrieves newer messages (for real-time updates)
func (s *ChatService) GetMessagesByRoomCursorForward(roomID string, cursor string, limit int) ([]db.Message, string, error) {
	ctx := context.Background()

	var rows pgx.Rows
	var err error

	if cursor == "" {
		// If no cursor, return empty (no newer messages)
		return []db.Message{}, "", nil
	}

	// Get messages newer than the cursor using ULID sorting
	rows, err = s.db.Query(ctx, `
		SELECT m.id, m.tenant_id, m.room_id, m.user_id, m.content, m.sent_at
		FROM messages m
		WHERE m.room_id = $1
		AND m.id > $2
		ORDER BY m.id ASC
		LIMIT $3
	`, roomID, cursor, limit+1) // Get one extra to determine if there are more pages

	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var messages []db.Message
	for rows.Next() {
		var msg db.Message
		err := rows.Scan(&msg.ID, &msg.TenantID, &msg.RoomID, &msg.UserID, &msg.Content, &msg.SentAt)
		if err != nil {
			return nil, "", err
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	// Determine next cursor
	var nextCursor string
	if len(messages) > limit {
		// Remove the extra message and set next cursor
		nextCursor = messages[limit-1].ID
		messages = messages[:limit]
	}

	return messages, nextCursor, nil
}

// GetUserChatRooms retrieves all chat rooms that a user is a member of
func (s *ChatService) GetUserChatRooms(userID string) ([]db.Room, error) {
	ctx := context.Background()

	rows, err := s.db.Query(ctx, `
		SELECT r.id, r.tenant_id, r.name, r.type, r.created_at
		FROM rooms r
		JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = $1
		ORDER BY r.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []db.Room
	for rows.Next() {
		var room db.Room
		err := rows.Scan(&room.ID, &room.TenantID, &room.Name, &room.Type, &room.CreatedAt)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

// CheckIfDirectRoomExists checks if a direct message room already exists between two users
func (s *ChatService) CheckIfDirectRoomExists(userID1, userID2 string) (*db.Room, error) {
	ctx := context.Background()

	var room db.Room
	err := s.db.QueryRow(ctx, `
		SELECT r.id, r.tenant_id, r.name, r.type, r.created_at
		FROM rooms r
		JOIN room_members rm1 ON r.id = rm1.room_id
		JOIN room_members rm2 ON r.id = rm2.room_id
		WHERE r.type = 'direct' 
		AND rm1.user_id = $1 
		AND rm2.user_id = $2
		LIMIT 1
	`, userID1, userID2).Scan(&room.ID, &room.TenantID, &room.Name, &room.Type, &room.CreatedAt)

	if err != nil {
		return nil, err // Room doesn't exist
	}

	return &room, nil
}

// ValidateUsersExist checks if all provided user IDs exist in the database
func (s *ChatService) ValidateUsersExist(userIDs []string) error {
	ctx := context.Background()

	for _, userID := range userIDs {
		var exists bool
		err := s.db.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
		`, userID).Scan(&exists)

		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("user with ID %s does not exist", userID)
		}
	}

	return nil
}

// GetRoomMembers retrieves all members of a specific room
func (s *ChatService) GetRoomMembers(roomID string) ([]string, error) {
	ctx := context.Background()

	rows, err := s.db.Query(ctx, `
		SELECT user_id FROM room_members WHERE room_id = $1
	`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}

// AddMemberToRoom adds a user to a group chat room
func (s *ChatService) AddMemberToRoom(roomID, userID string) error {
	ctx := context.Background()

	// Check if room exists and is a group room
	var roomType string
	err := s.db.QueryRow(ctx, "SELECT type FROM rooms WHERE id = $1", roomID).Scan(&roomType)
	if err != nil {
		return fmt.Errorf("room not found")
	}

	if roomType != "group" {
		return fmt.Errorf("can only add members to group rooms")
	}

	// Check if user is already a member
	isMember, err := s.VerifyUserIsRoomMember(userID, roomID)
	if err != nil {
		return err
	}
	if isMember {
		return fmt.Errorf("user is already a member of this room")
	}

	// Add user to room
	_, err = s.db.Exec(ctx, "INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)", roomID, userID)
	return err
}

// RemoveMemberFromRoom removes a user from a group chat room
func (s *ChatService) RemoveMemberFromRoom(roomID, userID string) error {
	ctx := context.Background()

	// Check if room exists and is a group room
	var roomType string
	err := s.db.QueryRow(ctx, "SELECT type FROM rooms WHERE id = $1", roomID).Scan(&roomType)
	if err != nil {
		return fmt.Errorf("room not found")
	}

	if roomType != "group" {
		return fmt.Errorf("can only remove members from group rooms")
	}

	// Remove user from room
	_, err = s.db.Exec(ctx, "DELETE FROM room_members WHERE room_id = $1 AND user_id = $2", roomID, userID)
	return err
}

// GetRoomInfo retrieves detailed information about a room including member count
func (s *ChatService) GetRoomInfo(roomID string) (*db.Room, int, error) {
	ctx := context.Background()

	// Get room details
	var room db.Room
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, name, type, created_at
		FROM rooms WHERE id = $1
	`, roomID).Scan(&room.ID, &room.TenantID, &room.Name, &room.Type, &room.CreatedAt)
	if err != nil {
		return nil, 0, err
	}

	// Get member count
	var memberCount int
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM room_members WHERE room_id = $1
	`, roomID).Scan(&memberCount)
	if err != nil {
		return nil, 0, err
	}

	return &room, memberCount, nil
}

// GetRecentMessages gets the most recent messages from a room (for preview)
func (s *ChatService) GetRecentMessages(roomID string, limit int) ([]db.Message, error) {
	ctx := context.Background()

	rows, err := s.db.Query(ctx, `
		SELECT m.id, m.tenant_id, m.room_id, m.user_id, m.content, m.sent_at
		FROM messages m
		WHERE m.room_id = $1
		ORDER BY m.id DESC
		LIMIT $2
	`, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []db.Message
	for rows.Next() {
		var msg db.Message
		err := rows.Scan(&msg.ID, &msg.TenantID, &msg.RoomID, &msg.UserID, &msg.Content, &msg.SentAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
