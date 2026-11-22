-- AI agent migration: run lifecycle, steps, plan commits.
-- The existing guide_sessions/guide_messages tables are reused as the
-- conversation store so the legacy /api/guide/chat contract keeps working.
-- Apply once on an existing database.

ALTER TABLE guide_sessions
  ADD COLUMN plan_json    JSON         NULL,
  ADD COLUMN plan_version INT          NOT NULL DEFAULT 0,
  ADD COLUMN last_run_id  BIGINT UNSIGNED NULL;

ALTER TABLE guide_messages
  MODIFY COLUMN content MEDIUMTEXT NOT NULL,
  ADD COLUMN run_id           BIGINT UNSIGNED NULL,
  ADD COLUMN tool_calls       JSON         NULL,
  ADD COLUMN status           VARCHAR(20)  NOT NULL DEFAULT 'done',
  ADD COLUMN model            VARCHAR(50)  NULL,
  ADD COLUMN prompt_tokens    INT          NOT NULL DEFAULT 0,
  ADD COLUMN completion_tokens INT         NOT NULL DEFAULT 0,
  ADD KEY idx_guide_messages_run (run_id);

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
