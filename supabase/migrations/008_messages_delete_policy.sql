-- Ajouter la politique DELETE pour les messages
DROP POLICY IF EXISTS "messages_delete_own" ON messages;
CREATE POLICY "messages_delete_own" ON messages FOR DELETE USING (
    auth.uid() = sender_id AND EXISTS (
        SELECT 1 FROM conversations 
        WHERE conversations.id = messages.conversation_id 
        AND (conversations.user1_id = auth.uid() OR conversations.user2_id = auth.uid())
    )
);