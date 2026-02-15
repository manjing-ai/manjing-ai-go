CREATE TABLE voices (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  age_group SMALLINT NOT NULL,
  gender SMALLINT NOT NULL,
  dialect SMALLINT NOT NULL DEFAULT 1,
  tone SMALLINT NOT NULL DEFAULT 1,
  sample_url VARCHAR(512) NULL,
  type SMALLINT NOT NULL DEFAULT 1,
  owner_user_id BIGINT NULL,
  extra_data JSONB NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);

COMMENT ON TABLE voices IS '声音表';
COMMENT ON COLUMN voices.name IS '声音名称';
COMMENT ON COLUMN voices.age_group IS '年龄段: 1儿童/2少年/3青年/4中年/5老年';
COMMENT ON COLUMN voices.gender IS '性别: 1男/2女';
COMMENT ON COLUMN voices.dialect IS '方言口音: 1标准普通话/2东北话/3四川话/4粤语/5台湾腔/6港式普通话/7外国口音';
COMMENT ON COLUMN voices.tone IS '音色: 1标准/2清亮/3浑厚/4沙哑/5柔和/6尖细/7气声/8鼻音/9金属';
COMMENT ON COLUMN voices.sample_url IS '试听音频URL';
COMMENT ON COLUMN voices.type IS '类型: 1官方/2用户';
COMMENT ON COLUMN voices.owner_user_id IS '所属用户ID，type=2时必填';
COMMENT ON COLUMN voices.extra_data IS '扩展字段(JSONB)';

CREATE INDEX idx_voices_type_age_gender ON voices(type, age_group, gender);
CREATE INDEX idx_voices_owner ON voices(owner_user_id);
CREATE INDEX idx_voices_dialect_tone ON voices(dialect, tone);
CREATE INDEX idx_voices_deleted_at ON voices(deleted_at);
