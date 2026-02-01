CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR(64),
  email VARCHAR(128),
  phone VARCHAR(32),
  password_hash VARCHAR(256) NOT NULL,
  avatar_url VARCHAR(512),
  status SMALLINT DEFAULT 1,
  role VARCHAR(32) DEFAULT 'user',
  last_login_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE email IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone ON users (phone) WHERE phone IS NOT NULL;
