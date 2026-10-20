-- Travel Guangdong - MySQL schema (tourist flow)
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
  KEY idx_routes_user (user_id)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS blog_likes (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL,
  blog_id    BIGINT UNSIGNED NOT NULL,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_blog_like (user_id, blog_id)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS guide_sessions (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL,
  title      VARCHAR(200) NOT NULL DEFAULT 'New Conversation',
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  KEY idx_guide_sessions_user (user_id)
) ENGINE = InnoDB;

CREATE TABLE IF NOT EXISTS guide_messages (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  session_id BIGINT UNSIGNED NOT NULL,
  role       VARCHAR(20)  NOT NULL,
  content    TEXT         NOT NULL,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_guide_messages_session (session_id),
  CONSTRAINT fk_guide_msg_session FOREIGN KEY (session_id) REFERENCES guide_sessions (id)
) ENGINE = InnoDB;
