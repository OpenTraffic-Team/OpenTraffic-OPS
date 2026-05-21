-- ============================================================
-- RTM监控运维平台 - Agent聊天会话模块表结构 (PostgreSQL)
-- ============================================================

-- ----------------------------
-- 1、聊天会话表
-- ----------------------------
DROP TABLE IF EXISTS bu_chat_session CASCADE;
CREATE TABLE bu_chat_session (
  id               BIGSERIAL         NOT NULL,
  user_id          BIGINT            NOT NULL,
  session_type     VARCHAR(16)       DEFAULT 'control',
  title            VARCHAR(128)      DEFAULT '',
  agent_session_id VARCHAR(128)      DEFAULT '',
  last_message_at  TIMESTAMP,
  message_count    INT               DEFAULT 0,
  remark           VARCHAR(500)      DEFAULT '',
  create_by        VARCHAR(64)       DEFAULT '',
  create_time      TIMESTAMP,
  update_by        VARCHAR(64)       DEFAULT '',
  update_time      TIMESTAMP,
  del_flag         VARCHAR(1)        DEFAULT '0',
  PRIMARY KEY (id)
);
COMMENT ON TABLE bu_chat_session IS 'Agent聊天会话表';
COMMENT ON COLUMN bu_chat_session.user_id IS '所属用户ID（JWT解析）';
COMMENT ON COLUMN bu_chat_session.session_type IS '会话类型：control(控制Agent) / perceive(感知Agent)';
COMMENT ON COLUMN bu_chat_session.title IS '会话标题（首次截取用户消息生成）';
COMMENT ON COLUMN bu_chat_session.agent_session_id IS '外部Agent服务返回的session_id，用于上下文延续';
COMMENT ON COLUMN bu_chat_session.last_message_at IS '最近一次消息时间（列表排序键）';
COMMENT ON COLUMN bu_chat_session.message_count IS '累计消息条数';
COMMENT ON COLUMN bu_chat_session.del_flag IS '删除标志：0存在 2删除';

-- 列表查询主索引：按用户 + 会话类型筛选 + 未删除 + 按 last_message_at DESC 排序
CREATE INDEX idx_chat_session_user_type_lastmsg ON bu_chat_session(user_id, session_type, last_message_at DESC NULLS LAST, id DESC) WHERE del_flag = '0';

-- ----------------------------
-- 2、聊天消息表（追加写，无软删除字段）
-- ----------------------------
DROP TABLE IF EXISTS bu_chat_message CASCADE;
CREATE TABLE bu_chat_message (
  id          BIGSERIAL         NOT NULL,
  session_id  BIGINT            NOT NULL,
  role        VARCHAR(16)       NOT NULL,
  content     TEXT,
  seq         INT               DEFAULT 0,
  create_time TIMESTAMP         DEFAULT NOW(),
  PRIMARY KEY (id),
  CONSTRAINT fk_chat_message_session FOREIGN KEY (session_id)
    REFERENCES bu_chat_session(id) ON DELETE CASCADE,
  CONSTRAINT uk_chat_message_session_seq UNIQUE (session_id, seq)
);
COMMENT ON TABLE bu_chat_message IS 'Agent聊天消息表';
COMMENT ON COLUMN bu_chat_message.session_id IS '所属会话ID（FK，ON DELETE CASCADE）';
COMMENT ON COLUMN bu_chat_message.role IS '消息角色：user(用户)/assistant(助手)';
COMMENT ON COLUMN bu_chat_message.content IS '消息内容（markdown）';
COMMENT ON COLUMN bu_chat_message.seq IS '会话内顺序号';
COMMENT ON COLUMN bu_chat_message.create_time IS '入库时间';

-- session_id + seq 已由 UNIQUE 约束创建索引，无需重复创建
