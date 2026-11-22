-- Travel Guangdong - MySQL schema (tourist flow)
-- The docker entrypoint client defaults to latin1; force UTF-8 for seed text.
SET NAMES utf8mb4;
-- MySQL 8.0+, utf8mb4

CREATE DATABASE IF NOT EXISTS travel_gd
  DEFAULT CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE travel_gd;

CREATE TABLE IF NOT EXISTS users (
  id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  username     VARCHAR(50)  NOT NULL,
  password_hash VARCHAR(100) NOT NULL,
  nickname     VARCHAR(50)  NOT NULL DEFAULT 'Guest User',
  avatar_url   VARCHAR(500) NOT NULL DEFAULT '',
  language     VARCHAR(5)   NOT NULL DEFAULT 'en',
  role         VARCHAR(16)  NOT NULL DEFAULT 'user',
  created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_users_username (username)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS scenic_images (
  id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  image_url      VARCHAR(500) NOT NULL,
  title_cn       VARCHAR(100) NOT NULL,
  title_en       VARCHAR(100) NOT NULL,
  location       VARCHAR(100) NOT NULL DEFAULT '',
  tag            VARCHAR(50)  NOT NULL DEFAULT '',
  description_cn TEXT         NOT NULL,
  description_en TEXT         NOT NULL,
  sort_order     INT          NOT NULL DEFAULT 0,
  created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS places (
  id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name_cn        VARCHAR(100) NOT NULL,
  name_en        VARCHAR(100) NOT NULL,
  lat            DECIMAL(10, 7) NOT NULL,
  lng            DECIMAL(10, 7) NOT NULL,
  tags           JSON         NULL,
  indoor         TINYINT(1)   NOT NULL DEFAULT 0,
  duration       INT          NOT NULL DEFAULT 60,
  description_cn TEXT         NOT NULL,
  description_en TEXT         NOT NULL,
  created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS knowledge (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  title_cn   VARCHAR(100) NOT NULL,
  title_en   VARCHAR(100) NOT NULL,
  content_cn TEXT         NOT NULL,
  content_en TEXT         NOT NULL,
  tags       JSON         NULL,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS favorites (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL,
  place_id   BIGINT UNSIGNED NOT NULL,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_fav_user_place (user_id, place_id),
  CONSTRAINT fk_fav_place FOREIGN KEY (place_id) REFERENCES places (id)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS travel_blogs (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL,
  title      VARCHAR(200) NOT NULL,
  content    TEXT         NOT NULL,
  images     JSON         NULL,
  likes      INT          NOT NULL DEFAULT 0,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_blogs_user (user_id)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS routes (
  id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id        BIGINT UNSIGNED NOT NULL,
  route_name     VARCHAR(200) NOT NULL,
  duration_days  INT          NOT NULL DEFAULT 1,
  total_distance VARCHAR(50)  NOT NULL DEFAULT '',
  mode           VARCHAR(20)  NOT NULL DEFAULT 'driving',
  pace           VARCHAR(20)  NOT NULL DEFAULT 'balanced',
  interests      JSON         NULL,
  waypoints      JSON         NULL,
  polyline       JSON         NULL,
  created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_blogs_user (user_id),
  KEY idx_blogs_created (created_at)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS blog_likes (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL,
  blog_id    BIGINT UNSIGNED NOT NULL,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_blog_like (user_id, blog_id),
  KEY idx_likes_blog (blog_id)
) ENGINE = InnoDB;

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

CREATE TABLE IF NOT EXISTS guide_sessions (
  id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id      BIGINT UNSIGNED NOT NULL,
  title        VARCHAR(200) NOT NULL DEFAULT 'New Conversation',
  plan_json    JSON         NULL,
  plan_version INT          NOT NULL DEFAULT 0,
  last_run_id  BIGINT UNSIGNED NULL,
  created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_guide_sessions_user (user_id)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS guide_messages (
  id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  session_id        BIGINT UNSIGNED NOT NULL,
  run_id            BIGINT UNSIGNED NULL,
  role              VARCHAR(20)  NOT NULL,
  content           MEDIUMTEXT   NOT NULL,
  tool_calls        JSON         NULL,
  status            VARCHAR(20)  NOT NULL DEFAULT 'done',
  model             VARCHAR(50)  NULL,
  prompt_tokens     INT          NOT NULL DEFAULT 0,
  completion_tokens INT          NOT NULL DEFAULT 0,
  created_at        DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_guide_messages_session (session_id),
  KEY idx_guide_messages_run (run_id),
  CONSTRAINT fk_guide_msg_session FOREIGN KEY (session_id) REFERENCES guide_sessions (id)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS ai_runs (
  id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  conversation_id   BIGINT UNSIGNED NOT NULL,
  user_id           BIGINT UNSIGNED NOT NULL,
  client_request_id CHAR(36)        NULL,
  status            VARCHAR(20)     NOT NULL DEFAULT 'queued',
  model             VARCHAR(50)     NOT NULL DEFAULT '',
  input             TEXT            NOT NULL,
  output            MEDIUMTEXT      NOT NULL,
  error_code        VARCHAR(50)     NULL,
  error_message     VARCHAR(500)    NULL,
  step_count        INT             NOT NULL DEFAULT 0,
  prompt_tokens     INT             NOT NULL DEFAULT 0,
  completion_tokens INT             NOT NULL DEFAULT 0,
  plan_json         JSON            NULL,
  plan_version      INT             NOT NULL DEFAULT 0,
  route_id          BIGINT UNSIGNED NULL,
  created_at        DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  started_at        DATETIME        NULL,
  finished_at       DATETIME        NULL,
  updated_at        DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_ai_runs_client (user_id, client_request_id),
  KEY idx_ai_runs_conversation (conversation_id, id),
  KEY idx_ai_runs_status (status),
  CONSTRAINT fk_ai_runs_conversation FOREIGN KEY (conversation_id) REFERENCES guide_sessions (id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS ai_steps (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  run_id     BIGINT UNSIGNED NOT NULL,
  step_no    INT             NOT NULL,
  type       VARCHAR(20)     NOT NULL,
  name       VARCHAR(100)    NOT NULL DEFAULT '',
  arguments  MEDIUMTEXT      NULL,
  result     MEDIUMTEXT      NULL,
  latency_ms INT             NOT NULL DEFAULT 0,
  error      VARCHAR(500)    NULL,
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_ai_steps_run (run_id, step_no),
  CONSTRAINT fk_ai_steps_run FOREIGN KEY (run_id) REFERENCES ai_runs (id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS ai_plan_commits (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id         BIGINT UNSIGNED NOT NULL,
  run_id          BIGINT UNSIGNED NOT NULL,
  idempotency_key VARCHAR(64)     NOT NULL,
  route_id        BIGINT UNSIGNED NOT NULL,
  created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_ai_plan_commit (user_id, idempotency_key),
  KEY idx_ai_plan_commit_run (run_id),
  CONSTRAINT fk_ai_plan_commit_run FOREIGN KEY (run_id) REFERENCES ai_runs (id)
) ENGINE = InnoDB;

-- --- shop -----------------------------------------------------------------
CREATE TABLE IF NOT EXISTS products (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name_cn         VARCHAR(120)    NOT NULL,
  name_en         VARCHAR(160)    NOT NULL,
  description_cn  TEXT            NOT NULL,
  description_en  TEXT            NOT NULL,
  price_cents     INT UNSIGNED    NOT NULL,
  category        VARCHAR(32)     NOT NULL DEFAULT 'souvenir',
  image_url       VARCHAR(500)    NOT NULL DEFAULT '',
  stock           INT             NOT NULL DEFAULT 0,
  sold_count      INT             NOT NULL DEFAULT 0,
  rating          DECIMAL(2,1)    NOT NULL DEFAULT 5.0,
  review_count    INT             NOT NULL DEFAULT 0,
  status          TINYINT         NOT NULL DEFAULT 1,
  created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_products_category (category, status),
  KEY idx_products_created (created_at)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS cart_items (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL,
  product_id BIGINT UNSIGNED NOT NULL,
  quantity   INT             NOT NULL DEFAULT 1,
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_cart_user_product (user_id, product_id),
  KEY idx_cart_user (user_id),
  CONSTRAINT fk_cart_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT fk_cart_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS orders (
  id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_no        VARCHAR(32)     NOT NULL,
  user_id         BIGINT UNSIGNED NOT NULL,
  total_cents     INT UNSIGNED    NOT NULL,
  status          VARCHAR(20)     NOT NULL DEFAULT 'pending',
  idempotency_key VARCHAR(64)     NULL,
  created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_orders_no (order_no),
  UNIQUE KEY uk_orders_idempotency (user_id, idempotency_key),
  KEY idx_orders_user (user_id, created_at),
  CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users (id)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS order_items (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  order_id    BIGINT UNSIGNED NOT NULL,
  product_id  BIGINT UNSIGNED NOT NULL,
  name_cn     VARCHAR(120)    NOT NULL,
  name_en     VARCHAR(160)    NOT NULL,
  price_cents INT UNSIGNED    NOT NULL,
  quantity    INT             NOT NULL,
  image_url   VARCHAR(500)    NOT NULL DEFAULT '',
  KEY idx_order_items_order (order_id),
  CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS product_reviews (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  product_id BIGINT UNSIGNED NOT NULL,
  user_id    BIGINT UNSIGNED NOT NULL,
  rating     TINYINT         NOT NULL DEFAULT 5,
  content    VARCHAR(1000)   NOT NULL,
  created_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_product_reviews_product (product_id, id),
  CONSTRAINT fk_product_reviews_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
  CONSTRAINT fk_product_reviews_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS activities (
  id             BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  title_cn       VARCHAR(160)    NOT NULL,
  title_en       VARCHAR(200)    NOT NULL,
  description_cn TEXT            NOT NULL,
  description_en TEXT            NOT NULL,
  location       VARCHAR(160)    NOT NULL,
  start_at       DATETIME        NOT NULL,
  end_at         DATETIME        NOT NULL,
  image_url      VARCHAR(500)    NOT NULL DEFAULT '',
  organizer      VARCHAR(160)    NOT NULL DEFAULT '',
  tag            VARCHAR(50)     NOT NULL DEFAULT '',
  status         VARCHAR(20)     NOT NULL DEFAULT 'upcoming',
  created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_activities_start (start_at),
  KEY idx_activities_status (status, start_at)
) ENGINE = InnoDB;
