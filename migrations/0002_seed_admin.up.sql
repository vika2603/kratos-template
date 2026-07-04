-- Demo-only seed: username "admin", password "admin123456".
INSERT INTO users (username, email, password_hash)
VALUES ('admin', 'admin@example.com', '$2y$10$tO/LZ9Lyc0JDy5uhoz2ByuoCmlaujJshEzmqQPHFRTRFbGGK0/xMK')
ON CONFLICT (username) DO NOTHING;
