CREATE TABLE IF NOT EXISTS projects (
  id BIGSERIAL PRIMARY KEY,
  owner_user_id BIGINT NOT NULL,
  name VARCHAR(128) NOT NULL,
  narrative_mode SMALLINT NOT NULL DEFAULT 1,
  cover_resource_id BIGINT,
  video_aspect_ratio VARCHAR(16) NOT NULL DEFAULT '16:9',
  style_ref VARCHAR(256),
  status SMALLINT NOT NULL DEFAULT 1,
  extra_data JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_projects_owner_created
  ON projects (owner_user_id, created_at DESC)
  WHERE status <> 3;

CREATE INDEX IF NOT EXISTS idx_projects_status_updated
  ON projects (status, updated_at DESC);

ALTER TABLE projects ADD CONSTRAINT chk_projects_status
  CHECK (status IN (1, 2, 3));

ALTER TABLE projects ADD CONSTRAINT chk_projects_narrative_mode
  CHECK (narrative_mode IN (1, 2));
