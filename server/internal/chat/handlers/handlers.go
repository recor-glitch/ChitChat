package handlers

import (
	"ChitChat/internal/shared/application/service/chat"
	"ChitChat/internal/shared/application/service/db"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

type ChatHandlers struct {
	chatService *chat.ChatService
}

func NewChatHandlers(chatService *chat.ChatService) *ChatHandlers {
	return &ChatHandlers{
		chatService: chatService,
	}
}

// GetChatRooms retrieves all chat rooms for the authenticated user
func (h *ChatHandlers) GetChatRooms(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Query database to get user's chat rooms
	rooms, err := h.chatService.GetUserChatRooms(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve chat rooms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Chat rooms retrieved",
		"rooms":   rooms,
	})
}

// CreateChatRoom creates a new chat room (group or direct)
func (h *ChatHandlers) CreateChatRoom(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req db.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that the current user is included in user_ids
	userIncluded := false
	for _, id := range req.UserIDs {
		if id == userID {
			userIncluded = true
			break
		}
	}
	if !userIncluded {
		req.UserIDs = append(req.UserIDs, userID)
	}

	// Validate that all users exist
	if err := h.chatService.ValidateUsersExist(req.UserIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For direct messages, ensure only 2 users
	if req.Type == "direct" && len(req.UserIDs) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Direct messages must have exactly 2 users"})
		return
	}

	// For direct messages, check if room already exists
	if req.Type == "direct" {
		existingRoom, err := h.chatService.CheckIfDirectRoomExists(req.UserIDs[0], req.UserIDs[1])
		if err == nil && existingRoom != nil {
			// Room already exists, return it
			c.JSON(http.StatusOK, gin.H{
				"message": "Direct message room already exists",
				"room":    existingRoom,
			})
			return
		}
	}

	// Create new room
	room := db.Room{
		ID:   uuid.New().String(),
		Name: req.Name,
		Type: req.Type,
	}

	// Save room and room members to database
	if err := h.chatService.CreateRoom(&room, req.UserIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Chat room created",
		"room":    room,
	})
}

// GetMessagesByRoom retrieves messages for a specific chat room
func (h *ChatHandlers) GetMessagesByRoom(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room ID is required"})
		return
	}

	// Verify user is member of this room
	isMember, err := h.chatService.VerifyUserIsRoomMember(userID, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room membership"})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not a member of this room"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Ensure reasonable limits
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Query messages from database with pagination
	messages, err := h.chatService.GetMessagesByRoom(roomID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Messages retrieved",
		"room_id":  roomID,
		"messages": messages,
		"page":     page,
		"limit":    limit,
	})
}

// SendMessage sends a message to a specific chat room
func (h *ChatHandlers) SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room ID is required"})
		return
	}

	var req db.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify user is member of this room
	isMember, err := h.chatService.VerifyUserIsRoomMember(userID, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room membership"})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not a member of this room"})
		return
	}

	// Create message with ULID
	message := db.Message{
		ID:      ulid.Make().String(),
		RoomID:  roomID,
		UserID:  userID,
		Content: req.Content,
	}

	// Save message to database
	if err := h.chatService.SaveMessage(&message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message sent",
		"room_id": roomID,
		"msg":     message,
	})
}

// SendDirectMessage sends a direct message to a specific user
func (h *ChatHandlers) SendDirectMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req db.DirectMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find or create direct message room between users
	room, err := h.chatService.FindOrCreateDirectRoom(userID, req.RecipientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find or create direct message room"})
		return
	}

	// Create message with ULID
	message := db.Message{
		ID:      ulid.Make().String(),
		RoomID:  room.ID,
		UserID:  userID,
		Content: req.Content,
	}

	// Save message to database
	if err := h.chatService.SaveMessage(&message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Direct message sent",
		"msg":     message,
		"room":    room,
	})
}

// GetDirectMessageRoom gets or creates a direct message room between two users
func (h *ChatHandlers) GetDirectMessageRoom(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	recipientID := c.Param("recipient_id")
	if recipientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipient ID is required"})
		return
	}

	// Find or create direct message room between these users
	room, err := h.chatService.FindOrCreateDirectRoom(userID, recipientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find or create direct message room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Direct message room retrieved",
		"room":    room,
	})
}

// AddMemberToRoom adds a user to a group chat room
func (h *ChatHandlers) AddMemberToRoom(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room ID is required"})
		return
	}

	var req struct {
		MemberID string `json:"member_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify current user is member of this room
	isMember, err := h.chatService.VerifyUserIsRoomMember(userID, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room membership"})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not a member of this room"})
		return
	}

	// Add member to room
	if err := h.chatService.AddMemberToRoom(roomID, req.MemberID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Member added to room",
		"room_id":   roomID,
		"member_id": req.MemberID,
	})
}

// RemoveMemberFromRoom removes a user from a group chat room
func (h *ChatHandlers) RemoveMemberFromRoom(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room ID is required"})
		return
	}

	var req struct {
		MemberID string `json:"member_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify current user is member of this room
	isMember, err := h.chatService.VerifyUserIsRoomMember(userID, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room membership"})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not a member of this room"})
		return
	}

	// Remove member from room
	if err := h.chatService.RemoveMemberFromRoom(roomID, req.MemberID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Member removed from room",
		"room_id":   roomID,
		"member_id": req.MemberID,
	})
}

// GetRoomInfo gets detailed information about a room
func (h *ChatHandlers) GetRoomInfo(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room ID is required"})
		return
	}

	// Verify user is member of this room
	isMember, err := h.chatService.VerifyUserIsRoomMember(userID, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room membership"})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not a member of this room"})
		return
	}

	// Get room info
	room, memberCount, err := h.chatService.GetRoomInfo(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get room information"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Room information retrieved",
		"room":         room,
		"member_count": memberCount,
	})
}

// GetDirectMessages gets direct messages between the authenticated user and another user
func (h *ChatHandlers) GetDirectMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	recipientID := c.Param("recipient_id")
	if recipientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipient ID is required"})
		return
	}

	// Find or create direct message room between these users
	room, err := h.chatService.FindOrCreateDirectRoom(userID, recipientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find or create direct message room"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Ensure reasonable limits
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Get messages from the direct message room
	messages, err := h.chatService.GetMessagesByRoom(room.ID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve direct messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Direct messages retrieved",
		"recipient_id": recipientID,
		"room_id":      room.ID,
		"messages":     messages,
		"page":         page,
		"limit":        limit,
	})
}

// GetMessagesByRoomCursor gets messages from a room using cursor-based pagination
func (h *ChatHandlers) GetMessagesByRoomCursor(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room ID is required"})
		return
	}

	// Verify user is member of this room
	isMember, err := h.chatService.VerifyUserIsRoomMember(userID, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room membership"})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not a member of this room"})
		return
	}

	// Parse cursor and limit parameters (both optional)
	cursor := c.DefaultQuery("cursor", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Ensure reasonable limits
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Get messages using cursor pagination
	messages, nextCursor, err := h.chatService.GetMessagesByRoomCursor(roomID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}

	response := gin.H{
		"message":  "Messages retrieved",
		"room_id":  roomID,
		"messages": messages,
		"limit":    limit,
	}

	if nextCursor != "" {
		response["next_cursor"] = nextCursor
		response["has_more"] = true
	} else {
		response["has_more"] = false
	}

	c.JSON(http.StatusOK, response)
}

// GetDirectMessagesCursor gets direct messages using cursor-based pagination
func (h *ChatHandlers) GetDirectMessagesCursor(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	recipientID := c.Param("recipient_id")
	if recipientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Recipient ID is required"})
		return
	}

	// Find or create direct message room between these users
	room, err := h.chatService.FindOrCreateDirectRoom(userID, recipientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find or create direct message room"})
		return
	}

	// Parse cursor and limit parameters (both optional)
	cursor := c.DefaultQuery("cursor", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Ensure reasonable limits
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Get messages using cursor pagination
	messages, nextCursor, err := h.chatService.GetMessagesByRoomCursor(room.ID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve direct messages"})
		return
	}

	response := gin.H{
		"message":      "Direct messages retrieved",
		"recipient_id": recipientID,
		"room_id":      room.ID,
		"messages":     messages,
		"limit":        limit,
	}

	if nextCursor != "" {
		response["next_cursor"] = nextCursor
		response["has_more"] = true
	} else {
		response["has_more"] = false
	}

	c.JSON(http.StatusOK, response)
}

// GetNewMessages gets newer messages since a specific cursor (for real-time updates)
func (h *ChatHandlers) GetNewMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	roomID := c.Param("id")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room ID is required"})
		return
	}

	// Verify user is member of this room
	isMember, err := h.chatService.VerifyUserIsRoomMember(userID, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room membership"})
		return
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not a member of this room"})
		return
	}

	// Parse cursor and limit parameters (cursor optional, limit optional)
	cursor := c.DefaultQuery("cursor", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Ensure reasonable limits
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// If no cursor provided, return empty result (no new messages)
	if cursor == "" {
		c.JSON(http.StatusOK, gin.H{
			"message":  "New messages retrieved",
			"room_id":  roomID,
			"messages": []db.Message{},
			"limit":    limit,
			"has_more": false,
		})
		return
	}

	// Get newer messages using cursor pagination
	messages, nextCursor, err := h.chatService.GetMessagesByRoomCursorForward(roomID, cursor, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve new messages"})
		return
	}

	response := gin.H{
		"message":  "New messages retrieved",
		"room_id":  roomID,
		"messages": messages,
		"limit":    limit,
	}

	if nextCursor != "" {
		response["next_cursor"] = nextCursor
		response["has_more"] = true
	} else {
		response["has_more"] = false
	}

	c.JSON(http.StatusOK, response)
}
