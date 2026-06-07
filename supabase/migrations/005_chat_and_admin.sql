-- ============================================================
-- TABLE CONVERSATIONS (regroupe les messages entre 2 users)
-- ============================================================
CREATE TABLE conversations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user1_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    user2_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    last_message_at timestamptz,
    created_at timestamptz DEFAULT now(),
    UNIQUE(user1_id, user2_id),
    CHECK (user1_id != user2_id),
    CHECK (user1_id < user2_id) -- Toujours dans l'ordre pour éviter les doublons
);

CREATE INDEX idx_conversations_user1 ON conversations(user1_id);
CREATE INDEX idx_conversations_user2 ON conversations(user2_id);
CREATE INDEX idx_conversations_last_message ON conversations(last_message_at DESC);

-- ============================================================
-- TABLE MESSAGES
-- ============================================================
CREATE TABLE messages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id uuid NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    content text NOT NULL CHECK (length(content) > 0 AND length(content) <= 5000),
    is_read boolean DEFAULT false,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_messages_unread ON messages(conversation_id, is_read) WHERE is_read = false;

-- ============================================================
-- TABLE ADMIN_CONVERSATIONS (chat user ↔ admin)
-- ============================================================
CREATE TABLE admin_conversations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    admin_id uuid REFERENCES user_profiles(id) ON DELETE SET NULL,
    subject text DEFAULT 'Support',
    status text DEFAULT 'open' CHECK (status IN ('open', 'assigned', 'closed')),
    last_message_at timestamptz,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_admin_conversations_user ON admin_conversations(user_id);
CREATE INDEX idx_admin_conversations_admin ON admin_conversations(admin_id);
CREATE INDEX idx_admin_conversations_status ON admin_conversations(status);

-- ============================================================
-- TABLE ADMIN_MESSAGES
-- ============================================================
CREATE TABLE admin_messages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id uuid NOT NULL REFERENCES admin_conversations(id) ON DELETE CASCADE,
    sender_id uuid NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    content text NOT NULL CHECK (length(content) > 0 AND length(content) <= 5000),
    is_read boolean DEFAULT false,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_admin_messages_conversation ON admin_messages(conversation_id, created_at DESC);
CREATE INDEX idx_admin_messages_sender ON admin_messages(sender_id);

-- ============================================================
-- AJOUTER STATUT ADMIN
-- ============================================================
ALTER TABLE user_profiles ADD COLUMN is_admin boolean DEFAULT false;

-- ============================================================
-- TRIGGER: Mettre à jour last_message_at sur conversations
-- ============================================================
CREATE OR REPLACE FUNCTION update_conversation_last_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE conversations SET last_message_at = NEW.created_at WHERE id = NEW.conversation_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_conversation_last_message
    AFTER INSERT ON messages
    FOR EACH ROW EXECUTE FUNCTION update_conversation_last_message();

-- ============================================================
-- TRIGGER: Mettre à jour last_message_at sur admin_conversations
-- ============================================================
CREATE OR REPLACE FUNCTION update_admin_conversation_last_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE admin_conversations SET last_message_at = NEW.created_at WHERE id = NEW.conversation_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_admin_conversation_last_message
    AFTER INSERT ON admin_messages
    FOR EACH ROW EXECUTE FUNCTION update_admin_conversation_last_message();

-- ============================================================
-- ROW LEVEL SECURITY
-- ============================================================
ALTER TABLE conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE admin_conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE admin_messages ENABLE ROW LEVEL SECURITY;

-- conversations: les users peuvent voir leurs propres conversations
CREATE POLICY "conversations_select_own" ON conversations FOR SELECT USING (
    auth.uid() = user1_id OR auth.uid() = user2_id
);
CREATE POLICY "conversations_insert_own" ON conversations FOR INSERT WITH CHECK (
    auth.uid() = user1_id OR auth.uid() = user2_id
);
CREATE POLICY "conversations_update_own" ON conversations FOR UPDATE USING (
    auth.uid() = user1_id OR auth.uid() = user2_id
);

-- messages: les users peuvent voir les messages de leurs conversations
CREATE POLICY "messages_select_own" ON messages FOR SELECT USING (
    EXISTS (
        SELECT 1 FROM conversations 
        WHERE conversations.id = messages.conversation_id 
        AND (conversations.user1_id = auth.uid() OR conversations.user2_id = auth.uid())
    )
);
CREATE POLICY "messages_insert_own" ON messages FOR INSERT WITH CHECK (
    auth.uid() = sender_id AND EXISTS (
        SELECT 1 FROM conversations 
        WHERE conversations.id = messages.conversation_id 
        AND (conversations.user1_id = auth.uid() OR conversations.user2_id = auth.uid())
    )
);
CREATE POLICY "messages_update_own" ON messages FOR UPDATE USING (
    EXISTS (
        SELECT 1 FROM conversations 
        WHERE conversations.id = messages.conversation_id 
        AND (conversations.user1_id = auth.uid() OR conversations.user2_id = auth.uid())
    )
);

-- admin_conversations: users voient les leurs, admins voient tout
CREATE POLICY "admin_conversations_select_own" ON admin_conversations FOR SELECT USING (
    auth.uid() = user_id OR EXISTS (
        SELECT 1 FROM user_profiles WHERE id = auth.uid() AND is_admin = true
    )
);
CREATE POLICY "admin_conversations_insert_own" ON admin_conversations FOR INSERT WITH CHECK (
    auth.uid() = user_id
);
CREATE POLICY "admin_conversations_update_admin" ON admin_conversations FOR UPDATE USING (
    EXISTS (SELECT 1 FROM user_profiles WHERE id = auth.uid() AND is_admin = true)
);

-- admin_messages: users voient les leurs, admins voient tout
CREATE POLICY "admin_messages_select_own" ON admin_messages FOR SELECT USING (
    EXISTS (
        SELECT 1 FROM admin_conversations 
        WHERE admin_conversations.id = admin_messages.conversation_id 
        AND (admin_conversations.user_id = auth.uid() OR EXISTS (
            SELECT 1 FROM user_profiles WHERE id = auth.uid() AND is_admin = true
        ))
    )
);
CREATE POLICY "admin_messages_insert_own" ON admin_messages FOR INSERT WITH CHECK (
    auth.uid() = sender_id
);