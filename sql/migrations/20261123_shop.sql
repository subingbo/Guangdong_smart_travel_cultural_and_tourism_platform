-- Shop MVP: products, cart, orders. Idempotent for existing databases.
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
