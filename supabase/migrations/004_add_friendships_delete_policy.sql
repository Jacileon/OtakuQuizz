-- Ajouter la politique DELETE manquante pour friendships
CREATE POLICY "friendships_delete_own" ON friendships FOR DELETE USING (
    auth.uid() = requester_id OR auth.uid() = addressee_id
);