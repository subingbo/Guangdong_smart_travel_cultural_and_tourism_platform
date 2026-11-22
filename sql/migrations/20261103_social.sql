-- Travel Guangdong - social features upgrade (idempotent).
-- Apply with: sudo mysql < sql/migrations/20261103_social.sql
USE travel_gd;

CREATE TABLE IF NOT EXISTS comments (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  blog_id    BIGINT UNSIGNED NOT NULL,
  user_id    BIGINT UNSIGNED NOT NULL,
  content    VARCHAR(1000) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_comments_blog (blog_id, created_at),
  KEY idx_comments_user (user_id),
  CONSTRAINT fk_comments_blog FOREIGN KEY (blog_id) REFERENCES travel_blogs (id) ON DELETE CASCADE,
  CONSTRAINT fk_comments_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS follows (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  follower_id BIGINT UNSIGNED NOT NULL,
  followee_id BIGINT UNSIGNED NOT NULL,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_follow_pair (follower_id, followee_id),
  KEY idx_follow_followee (followee_id, created_at),
  CONSTRAINT fk_follow_follower FOREIGN KEY (follower_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT fk_follow_followee FOREIGN KEY (followee_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS notifications (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL COMMENT 'recipient',
  actor_id   BIGINT UNSIGNED NOT NULL,
  type       VARCHAR(16) NOT NULL COMMENT 'like | comment | follow',
  blog_id    BIGINT UNSIGNED NULL,
  comment_id BIGINT UNSIGNED NULL,
  is_read    TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_notif_user (user_id, is_read, created_at),
  CONSTRAINT fk_notif_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT fk_notif_actor FOREIGN KEY (actor_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT fk_notif_blog FOREIGN KEY (blog_id) REFERENCES travel_blogs (id) ON DELETE CASCADE
) ENGINE = InnoDB;

-- Add indexes only when missing so this file can be re-applied safely.
DROP PROCEDURE IF EXISTS tg_add_index_if_missing;
DELIMITER //
CREATE PROCEDURE tg_add_index_if_missing(IN tbl VARCHAR(64), IN idx VARCHAR(64), IN ddl TEXT)
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema = DATABASE() AND table_name = tbl AND index_name = idx
  ) THEN
    SET @s = ddl;
    PREPARE stmt FROM @s;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
  END IF;
END //
DELIMITER ;

CALL tg_add_index_if_missing('travel_blogs', 'idx_blogs_created',
  'ALTER TABLE travel_blogs ADD INDEX idx_blogs_created (created_at)');
CALL tg_add_index_if_missing('blog_likes', 'idx_likes_blog',
  'ALTER TABLE blog_likes ADD INDEX idx_likes_blog (blog_id)');

DROP PROCEDURE tg_add_index_if_missing;
