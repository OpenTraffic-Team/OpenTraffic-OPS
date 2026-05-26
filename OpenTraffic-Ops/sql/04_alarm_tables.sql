-- ============================================================
-- RTM监控运维平台 - 告警模块表结构 (PostgreSQL)
-- ============================================================

-- ----------------------------
-- 1、告警通道表
-- ----------------------------
DROP TABLE IF EXISTS bu_alarm_channel CASCADE;
CREATE TABLE bu_alarm_channel (
  id           BIGSERIAL         NOT NULL,
  name         VARCHAR(100)      NOT NULL,
  channel_type VARCHAR(32)       NOT NULL,
  config       TEXT,
  status       VARCHAR(1)        DEFAULT '0',
  is_default   VARCHAR(1)        DEFAULT '0',
  remark       VARCHAR(500)      DEFAULT '',
  create_by    VARCHAR(64)       DEFAULT '',
  create_time  TIMESTAMP,
  update_by    VARCHAR(64)       DEFAULT '',
  update_time  TIMESTAMP,
  del_flag     VARCHAR(1)        DEFAULT '0',
  PRIMARY KEY (id)
);
COMMENT ON TABLE bu_alarm_channel IS '告警通道表';
COMMENT ON COLUMN bu_alarm_channel.name IS '通道名称';
COMMENT ON COLUMN bu_alarm_channel.channel_type IS '通道类型：email/sms/dingtalk/wechat/platform';
COMMENT ON COLUMN bu_alarm_channel.config IS '通道配置JSON';
COMMENT ON COLUMN bu_alarm_channel.status IS '状态：0启用 1禁用';
COMMENT ON COLUMN bu_alarm_channel.is_default IS '是否默认：0否 1是';

CREATE INDEX idx_alarm_channel_type ON bu_alarm_channel(channel_type);
CREATE INDEX idx_alarm_channel_status ON bu_alarm_channel(status);

-- ----------------------------
-- 2、告警规则表
-- ----------------------------
DROP TABLE IF EXISTS bu_alarm_rule CASCADE;
CREATE TABLE bu_alarm_rule (
  id           BIGSERIAL         NOT NULL,
  name         VARCHAR(100)      NOT NULL,
  rule_type    VARCHAR(32)       NOT NULL,
  metric_type  VARCHAR(32)       NOT NULL,
  host_id      BIGINT            DEFAULT 0,
  threshold    NUMERIC(10,2),
  compare_op   VARCHAR(16),
  duration     INT               DEFAULT 0,
  severity     VARCHAR(16)       DEFAULT 'warning',
  channel_ids  TEXT,
  status       VARCHAR(1)        DEFAULT '0',
  remark       VARCHAR(500)      DEFAULT '',
  create_by    VARCHAR(64)       DEFAULT '',
  create_time  TIMESTAMP,
  update_by    VARCHAR(64)       DEFAULT '',
  update_time  TIMESTAMP,
  del_flag     VARCHAR(1)        DEFAULT '0',
  PRIMARY KEY (id)
);
COMMENT ON TABLE bu_alarm_rule IS '告警规则表';
COMMENT ON COLUMN bu_alarm_rule.name IS '规则名称';
COMMENT ON COLUMN bu_alarm_rule.rule_type IS '规则类型：metric(指标)/service(服务)';
COMMENT ON COLUMN bu_alarm_rule.metric_type IS '指标/服务类型：cpu/mem/disk/network/load/host_offline/agent_offline';
COMMENT ON COLUMN bu_alarm_rule.host_id IS '关联主机ID，0表示全部主机';
COMMENT ON COLUMN bu_alarm_rule.threshold IS '阈值（百分比或数值）';
COMMENT ON COLUMN bu_alarm_rule.compare_op IS '比较运算符：gt/lt/ge/le/eq';
COMMENT ON COLUMN bu_alarm_rule.duration IS '持续时间（秒），持续超过才告警';
COMMENT ON COLUMN bu_alarm_rule.severity IS '告警级别：warning(警告)/critical(严重)';
COMMENT ON COLUMN bu_alarm_rule.channel_ids IS '告警通道ID数组JSON';
COMMENT ON COLUMN bu_alarm_rule.status IS '状态：0启用 1禁用';

CREATE INDEX idx_alarm_rule_type ON bu_alarm_rule(rule_type);
CREATE INDEX idx_alarm_rule_metric ON bu_alarm_rule(metric_type);
CREATE INDEX idx_alarm_rule_host ON bu_alarm_rule(host_id);
CREATE INDEX idx_alarm_rule_status ON bu_alarm_rule(status);

-- ----------------------------
-- 3、告警记录表
-- ----------------------------
DROP TABLE IF EXISTS bu_alarm_record CASCADE;
CREATE TABLE bu_alarm_record (
  id            BIGSERIAL         NOT NULL,
  rule_id       BIGINT            NOT NULL,
  rule_name     VARCHAR(100)      DEFAULT '',
  host_id       BIGINT,
  host_ip       VARCHAR(64)       DEFAULT '',
  host_name     VARCHAR(100)      DEFAULT '',
  alarm_type    VARCHAR(32)       NOT NULL,
  metric_type   VARCHAR(32)       DEFAULT '',
  current_value NUMERIC(10,2)     DEFAULT 0,
  threshold     NUMERIC(10,2)     DEFAULT 0,
  severity      VARCHAR(16)       DEFAULT 'warning',
  content       TEXT,
  status        VARCHAR(16)       DEFAULT 'triggered',
  trigger_time  TIMESTAMP,
  resolve_time  TIMESTAMP,
  notify_status VARCHAR(16)       DEFAULT 'pending',
  create_time   TIMESTAMP         DEFAULT NOW(),
  PRIMARY KEY (id)
);
COMMENT ON TABLE bu_alarm_record IS '告警记录表';
COMMENT ON COLUMN bu_alarm_record.rule_id IS '关联规则ID';
COMMENT ON COLUMN bu_alarm_record.rule_name IS '规则名称（快照）';
COMMENT ON COLUMN bu_alarm_record.host_id IS '主机ID';
COMMENT ON COLUMN bu_alarm_record.host_ip IS '主机IP';
COMMENT ON COLUMN bu_alarm_record.host_name IS '主机名称';
COMMENT ON COLUMN bu_alarm_record.alarm_type IS '告警类型：metric/service';
COMMENT ON COLUMN bu_alarm_record.metric_type IS '指标/服务类型';
COMMENT ON COLUMN bu_alarm_record.current_value IS '当前值';
COMMENT ON COLUMN bu_alarm_record.threshold IS '阈值';
COMMENT ON COLUMN bu_alarm_record.severity IS '告警级别';
COMMENT ON COLUMN bu_alarm_record.content IS '告警内容';
COMMENT ON COLUMN bu_alarm_record.status IS '状态：triggered(已触发)/resolved(已恢复)/acknowledged(已确认)';
COMMENT ON COLUMN bu_alarm_record.trigger_time IS '触发时间';
COMMENT ON COLUMN bu_alarm_record.resolve_time IS '恢复时间';
COMMENT ON COLUMN bu_alarm_record.notify_status IS '通知状态：pending/success/failed';

CREATE INDEX idx_alarm_record_rule ON bu_alarm_record(rule_id);
CREATE INDEX idx_alarm_record_host ON bu_alarm_record(host_id);
CREATE INDEX idx_alarm_record_status ON bu_alarm_record(status);
CREATE INDEX idx_alarm_record_severity ON bu_alarm_record(severity);
CREATE INDEX idx_alarm_record_trigger_time ON bu_alarm_record(trigger_time);

-- ----------------------------
-- 4、告警通知日志表
-- ----------------------------
DROP TABLE IF EXISTS bu_alarm_notify_log CASCADE;
CREATE TABLE bu_alarm_notify_log (
  id            BIGSERIAL         NOT NULL,
  record_id     BIGINT            NOT NULL,
  channel_id    BIGINT            NOT NULL,
  channel_name  VARCHAR(100)      DEFAULT '',
  channel_type  VARCHAR(32)       DEFAULT '',
  status        VARCHAR(16)       DEFAULT 'failed',
  response      TEXT,
  send_time     TIMESTAMP,
  create_time   TIMESTAMP         DEFAULT NOW(),
  PRIMARY KEY (id)
);
COMMENT ON TABLE bu_alarm_notify_log IS '告警通知日志表';
COMMENT ON COLUMN bu_alarm_notify_log.record_id IS '关联告警记录ID';
COMMENT ON COLUMN bu_alarm_notify_log.channel_id IS '通道ID';
COMMENT ON COLUMN bu_alarm_notify_log.channel_name IS '通道名称';
COMMENT ON COLUMN bu_alarm_notify_log.channel_type IS '通道类型';
COMMENT ON COLUMN bu_alarm_notify_log.status IS '发送状态：success/failed';
COMMENT ON COLUMN bu_alarm_notify_log.response IS '响应内容';
COMMENT ON COLUMN bu_alarm_notify_log.send_time IS '发送时间';

CREATE INDEX idx_notify_log_record ON bu_alarm_notify_log(record_id);
CREATE INDEX idx_notify_log_channel ON bu_alarm_notify_log(channel_id);
CREATE INDEX idx_notify_log_send_time ON bu_alarm_notify_log(send_time);
