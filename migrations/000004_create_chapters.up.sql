CREATE TABLE IF NOT EXISTS chapters (
  id BIGSERIAL PRIMARY KEY,
  project_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  content TEXT,
  summary VARCHAR(256),
  order_index INT NOT NULL DEFAULT 0,
  status SMALLINT NOT NULL DEFAULT 1,
  extra_data JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_chapters_project_order
  ON chapters (project_id, order_index);

CREATE INDEX IF NOT EXISTS idx_chapters_project_status
  ON chapters (project_id, status);

CREATE INDEX IF NOT EXISTS idx_chapters_deleted_at
  ON chapters (deleted_at);

ALTER TABLE chapters ADD CONSTRAINT chk_chapters_status
  CHECK (status IN (1, 2, 3));
