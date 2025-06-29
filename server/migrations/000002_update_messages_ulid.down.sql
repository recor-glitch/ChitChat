-- Revert messages table back to UUID structure
-- First, drop the existing messages table
DROP TABLE IF EXISTS messages CASCADE;

-- Recreate messages table with UUID (original structure)
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create basic indexes for UUID structure
CREATE INDEX idx_messages_room_id ON messages (room_id);
CREATE INDEX idx_messages_user_id ON messages (user_id);
CREATE INDEX idx_messages_tenant_id ON messages (tenant_id);
CREATE INDEX idx_messages_sent_at ON messages (sent_at DESC); 