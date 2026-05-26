-- ============================================================
-- OpenTraffic Ops 监控运维平台 - 业务模块表结构 (PostgreSQL)
-- 重构变更:
--   1) 精简 bu_host_info 表，只保留核心硬件与在线字段
--   2) 新增 bu_host_health 表，Agent 每 3 秒上报健康数据直接写入 PG
--   3) 删除路口、区域、运维、资产、业务归属等无关字段
--   4) 去掉注册审批流程，Agent 上报自动入库
-- ============================================================

-- ============================================================================
-- 一、主机管理
-- ============================================================================

-- ----------------------------
-- 1、主机信息表（精简版）
-- ----------------------------
DROP TABLE IF EXISTS bu_host_info CASCADE;
CREATE TABLE bu_host_info (
  id                 BIGSERIAL       NOT NULL,
  ip                 VARCHAR(64)     NOT NULL,
  name               VARCHAR(128)    DEFAULT '',
  os_type            VARCHAR(50)     DEFAULT '',
  os_version         VARCHAR(100)    DEFAULT '',
  cpu_arch           VARCHAR(20)     DEFAULT '',
  cpu_cores          INT             DEFAULT 0,
  cpu_model          VARCHAR(200)    DEFAULT '',
  mem_total_mb       BIGINT          DEFAULT 0,
  disk_total_gb      BIGINT          DEFAULT 0,
  gpu_info           TEXT,
  mac_address        VARCHAR(100)    DEFAULT '',
  proxy_version      VARCHAR(50)     DEFAULT '',
  heartbeat_interval INT             DEFAULT 3,
  is_online          BOOLEAN         DEFAULT FALSE,
  register_time      TIMESTAMP,
  offline_time       TIMESTAMP,
  last_heartbeat     TIMESTAMP,
  PRIMARY KEY (id)
);
COMMENT ON TABLE bu_host_info IS '主机信息表（精简版）';
COMMENT ON COLUMN bu_host_info.id IS '主键ID';
COMMENT ON COLUMN bu_host_info.ip IS '主机IP';
COMMENT ON COLUMN bu_host_info.name IS '主机名称';
COMMENT ON COLUMN bu_host_info.os_type IS '操作系统类型(linux/windows/darwin)';
COMMENT ON COLUMN bu_host_info.os_version IS '操作系统版本';
COMMENT ON COLUMN bu_host_info.cpu_arch IS 'CPU架构';
COMMENT ON COLUMN bu_host_info.cpu_cores IS 'CPU逻辑核数';
COMMENT ON COLUMN bu_host_info.cpu_model IS 'CPU型号';
COMMENT ON COLUMN bu_host_info.mem_total_mb IS '内存总量(MB)';
COMMENT ON COLUMN bu_host_info.disk_total_gb IS '磁盘总量(GB)';
COMMENT ON COLUMN bu_host_info.gpu_info IS '显卡信息JSON';
COMMENT ON COLUMN bu_host_info.mac_address IS '主网卡MAC地址';
COMMENT ON COLUMN bu_host_info.proxy_version IS 'Proxy版本号';
COMMENT ON COLUMN bu_host_info.heartbeat_interval IS '心跳上报间隔（秒）';
COMMENT ON COLUMN bu_host_info.is_online IS '是否在线';
COMMENT ON COLUMN bu_host_info.register_time IS 'Agent首次上报时间';
COMMENT ON COLUMN bu_host_info.offline_time IS '离线时间';
COMMENT ON COLUMN bu_host_info.last_heartbeat IS '最后心跳时间';

CREATE UNIQUE INDEX idx_bu_host_info_ip ON bu_host_info(ip);
CREATE INDEX idx_bu_host_info_online ON bu_host_info(is_online);
CREATE INDEX idx_bu_host_info_last_heartbeat ON bu_host_info(last_heartbeat);

-- ----------------------------
-- 2、主机健康度表（新增）
-- ----------------------------
DROP TABLE IF EXISTS bu_host_health CASCADE;
CREATE TABLE bu_host_health (
  id            BIGSERIAL         NOT NULL,
  host_id       BIGINT            NOT NULL,
  ip            VARCHAR(64)       NOT NULL,
  cpu_usage     NUMERIC(5,2),
  mem_usage     NUMERIC(5,2),
  mem_used_mb   BIGINT,
  disk_usage    NUMERIC(5,2),
  net_in_kbps   NUMERIC(10,2),
  net_out_kbps  NUMERIC(10,2),
  load_avg      VARCHAR(50),
  is_online     BOOLEAN,
  report_time   TIMESTAMP         NOT NULL,
  create_time   TIMESTAMP         DEFAULT NOW(),
  PRIMARY KEY (id)
);
COMMENT ON TABLE bu_host_health IS '主机健康度表';
COMMENT ON COLUMN bu_host_health.id IS '主键ID';
COMMENT ON COLUMN bu_host_health.host_id IS '关联主机ID';
COMMENT ON COLUMN bu_host_health.ip IS '主机IP（冗余）';
COMMENT ON COLUMN bu_host_health.cpu_usage IS 'CPU使用率%';
COMMENT ON COLUMN bu_host_health.mem_usage IS '内存使用率%';
COMMENT ON COLUMN bu_host_health.mem_used_mb IS '内存使用MB';
COMMENT ON COLUMN bu_host_health.disk_usage IS '磁盘使用率%';
COMMENT ON COLUMN bu_host_health.net_in_kbps IS '网络入流量KB/s';
COMMENT ON COLUMN bu_host_health.net_out_kbps IS '网络出流量KB/s';
COMMENT ON COLUMN bu_host_health.load_avg IS '系统负载';
COMMENT ON COLUMN bu_host_health.is_online IS '上报时在线状态';
COMMENT ON COLUMN bu_host_health.report_time IS 'Agent上报时间';
COMMENT ON COLUMN bu_host_health.create_time IS '记录入库时间';

CREATE INDEX idx_health_ip_time ON bu_host_health(ip, report_time);
CREATE INDEX idx_health_time ON bu_host_health(report_time);
