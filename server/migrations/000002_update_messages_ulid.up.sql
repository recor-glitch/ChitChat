-- Update messages table to use ULID and optimize for timestamp-based sorting
-- First, drop the existing messages table
DROP TABLE IF EXISTS messages CASCADE;

-- Recreate messages table with ULID and optimized structure
CREATE TABLE messages (
    id TEXT PRIMARY KEY, -- ULID is 26 characters
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient message retrieval
-- Index for room-based queries (most common)
CREATE INDEX idx_messages_room_id_sent_at ON messages (room_id, sent_at DESC);

-- Index for cursor-based pagination (ULID is lexicographically sortable)
CREATE INDEX idx_messages_room_id_id ON messages (room_id, id DESC);

-- Index for user-based queries
CREATE INDEX idx_messages_user_id ON messages (user_id);

-- Index for tenant-based queries
CREATE INDEX idx_messages_tenant_id ON messages (tenant_id);

-- Index for time-based queries
CREATE INDEX idx_messages_sent_at ON messages (sent_at DESC);

-- Composite index for room + user queries
CREATE INDEX idx_messages_room_user ON messages (room_id, user_id);

-- Add constraint to ensure ULID format (26 characters, alphanumeric)
ALTER TABLE messages ADD CONSTRAINT check_ulid_format 
    CHECK (id ~ '^[0-9A-Z]{26}$');

-- Add constraint to ensure sent_at is not null
ALTER TABLE messages ALTER COLUMN sent_at SET NOT NULL; 