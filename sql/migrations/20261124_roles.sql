-- Roles for the admin/merchant consoles. Run once on an existing database
-- created before 2026-11-24; fresh installs get role from sql/schema.sql.
ALTER TABLE users ADD COLUMN role VARCHAR(16) NOT NULL DEFAULT 'user' AFTER language;

-- Demo staff accounts (password: admin123 / merchant123).
INSERT INTO users (username, password_hash, nickname, avatar_url, language, role) VALUES
('admin', '$2b$12$FYK5lEuFlMIrzGOVr9FUz./BYJotv7.hQmbW9K663nmGzrtrfa49O', '平台管理员', 'https://i.pravatar.cc/150?img=12', 'zh', 'admin'),
('merchant', '$2b$12$IPf0yKPoPQYCXqg3Q1nCNOM58ajJes.oFAGJC3JWSfyZxp6rRUz02', '商家小助手', 'https://i.pravatar.cc/150?img=32', 'zh', 'merchant')
ON DUPLICATE KEY UPDATE role = VALUES(role);
