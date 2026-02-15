-- 模型配置表
CREATE TABLE llm_models (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(64) NOT NULL,
  provider VARCHAR(32) NOT NULL,
  base_url VARCHAR(256) NOT NULL,
  api_key VARCHAR(256) NOT NULL,
  model VARCHAR(64) NOT NULL,
  max_tokens INT NOT NULL DEFAULT 4096,
  temperature DECIMAL(3,2) NOT NULL DEFAULT 0.70,
  timeout INT NOT NULL DEFAULT 60,
  purpose VARCHAR(32) NOT NULL DEFAULT 'default',
  is_active BOOLEAN NOT NULL DEFAULT true,
  extra_config JSONB NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);

COMMENT ON TABLE llm_models IS '大语言模型配置表';
COMMENT ON COLUMN llm_models.name IS '模型显示名称';
COMMENT ON COLUMN llm_models.provider IS '服务商标识: deepseek/qwen/zhipu/moonshot/openai';
COMMENT ON COLUMN llm_models.base_url IS 'API端点URL';
COMMENT ON COLUMN llm_models.api_key IS 'API密钥(加密存储)';
COMMENT ON COLUMN llm_models.model IS '模型标识';
COMMENT ON COLUMN llm_models.max_tokens IS '最大输出Token数';
COMMENT ON COLUMN llm_models.temperature IS '温度参数';
COMMENT ON COLUMN llm_models.timeout IS '超时时间(秒)';
COMMENT ON COLUMN llm_models.purpose IS '用途: default/chapter_parse/subject_extract';
COMMENT ON COLUMN llm_models.is_active IS '是否启用';
COMMENT ON COLUMN llm_models.extra_config IS '扩展配置(JSONB)';

CREATE INDEX idx_llm_models_purpose ON llm_models(purpose) WHERE deleted_at IS NULL;
CREATE INDEX idx_llm_models_provider ON llm_models(provider) WHERE deleted_at IS NULL;

-- 调用日志表
CREATE TABLE llm_call_logs (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  model_id BIGINT NULL,
  provider VARCHAR(32) NOT NULL,
  model VARCHAR(64) NOT NULL,
  purpose VARCHAR(32) NOT NULL,
  prompt_tokens INT NOT NULL DEFAULT 0,
  completion_tokens INT NOT NULL DEFAULT 0,
  total_tokens INT NOT NULL DEFAULT 0,
  duration_ms INT NOT NULL DEFAULT 0,
  status SMALLINT NOT NULL DEFAULT 1,
  error_message TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE llm_call_logs IS 'LLM调用日志表';
COMMENT ON COLUMN llm_call_logs.user_id IS '调用用户ID';
COMMENT ON COLUMN llm_call_logs.model_id IS '关联模型配置ID';
COMMENT ON COLUMN llm_call_logs.provider IS '服务商标识';
COMMENT ON COLUMN llm_call_logs.model IS '实际使用的模型标识';
COMMENT ON COLUMN llm_call_logs.purpose IS '调用用途';
COMMENT ON COLUMN llm_call_logs.prompt_tokens IS '输入Token数';
COMMENT ON COLUMN llm_call_logs.completion_tokens IS '输出Token数';
COMMENT ON COLUMN llm_call_logs.total_tokens IS '总Token数';
COMMENT ON COLUMN llm_call_logs.duration_ms IS '调用耗时(毫秒)';
COMMENT ON COLUMN llm_call_logs.status IS '状态: 1成功/2失败/3超时';
COMMENT ON COLUMN llm_call_logs.error_message IS '错误信息';

CREATE INDEX idx_llm_call_logs_user_id ON llm_call_logs(user_id);
CREATE INDEX idx_llm_call_logs_purpose ON llm_call_logs(purpose);
CREATE INDEX idx_llm_call_logs_created_at ON llm_call_logs(created_at);
CREATE INDEX idx_llm_call_logs_status ON llm_call_logs(status);
