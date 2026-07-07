-- Keyset pagination on ListUsers orders by (created_at, id).
CREATE INDEX users_created_at_id_idx ON users (created_at, id);
