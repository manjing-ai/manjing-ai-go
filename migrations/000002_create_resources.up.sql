CREATE TABLE IF NOT EXISTS resources (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  name VARCHAR(255) NOT NULL,

  type VARCHAR(32) NOT NULL DEFAULT 'image',
  category VARCHAR(32),

  object_key VARCHAR(512) NOT NULL,
  file_name VARCHAR(255),
  file_ext VARCHAR(16),
  mime_type VARCHAR(64),
  width INT,
  height INT,
  aspect VARCHAR(16),
  size_bytes BIGINT,

  status VARCHAR(16) NOT NULL DEFAULT 'active',
  extra_data JSONB,

  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_resources_user_created
  ON resources (user_id, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_resources_user_type
  ON resources (user_id, type)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_resources_status
  ON resources (status, created_at)
  WHERE deleted_at IS NULL;

ALTER TABLE resources ADD CONSTRAINT chk_resources_status
  CHECK (status IN ('pending', 'active', 'deleted'));

ALTER TABLE resources ADD CONSTRAINT chk_resources_type
  CHECK (type IN ('image', 'video', 'audio', 'other'));
