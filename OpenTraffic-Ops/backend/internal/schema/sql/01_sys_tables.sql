-- ============================================================
-- OpenTraffic Ops 监控运维平台 - 系统管理模块表结构 (PostgreSQL)
-- 来源: Java若依(RuoYi)框架 sys_* 表
-- ============================================================

-- ----------------------------
-- 2、用户表
-- ----------------------------
DROP TABLE IF EXISTS sys_user CASCADE;
CREATE TABLE sys_user (
  user_id           BIGSERIAL       NOT NULL,
  dept_id           BIGINT          DEFAULT NULL,
  user_name         VARCHAR(30)     NOT NULL,
  nick_name         VARCHAR(30)     NOT NULL,
  user_type         VARCHAR(2)      DEFAULT '00',
  email             VARCHAR(50)     DEFAULT '',
  phonenumber       VARCHAR(11)     DEFAULT '',
  sex               CHAR(1)         DEFAULT '0',
  avatar            VARCHAR(100)    DEFAULT '',
  password          VARCHAR(100)    DEFAULT '',
  status            CHAR(1)         DEFAULT '0',
  del_flag          CHAR(1)         DEFAULT '0',
  login_ip          VARCHAR(128)    DEFAULT '',
  login_date        TIMESTAMP,
  create_by         VARCHAR(64)     DEFAULT '',
  create_time       TIMESTAMP,
  update_by         VARCHAR(64)     DEFAULT '',
  update_time       TIMESTAMP,
  remark            VARCHAR(500)    DEFAULT NULL,
  PRIMARY KEY (user_id)
);
COMMENT ON TABLE sys_user IS '用户信息表';
COMMENT ON COLUMN sys_user.user_id IS '用户ID';
COMMENT ON COLUMN sys_user.dept_id IS '部门ID';
COMMENT ON COLUMN sys_user.user_name IS '用户账号';
COMMENT ON COLUMN sys_user.nick_name IS '用户昵称';
COMMENT ON COLUMN sys_user.user_type IS '用户类型（00系统用户）';
COMMENT ON COLUMN sys_user.email IS '用户邮箱';
COMMENT ON COLUMN sys_user.phonenumber IS '手机号码';
COMMENT ON COLUMN sys_user.sex IS '用户性别（0男 1女 2未知）';
COMMENT ON COLUMN sys_user.avatar IS '头像地址';
COMMENT ON COLUMN sys_user.password IS '密码';
COMMENT ON COLUMN sys_user.status IS '帐号状态（0正常 1停用）';
COMMENT ON COLUMN sys_user.del_flag IS '删除标志（0代表存在 2代表删除）';
COMMENT ON COLUMN sys_user.login_ip IS '最后登录IP';
COMMENT ON COLUMN sys_user.login_date IS '最后登录时间';
COMMENT ON COLUMN sys_user.create_by IS '创建者';
COMMENT ON COLUMN sys_user.create_time IS '创建时间';
COMMENT ON COLUMN sys_user.update_by IS '更新者';
COMMENT ON COLUMN sys_user.update_time IS '更新时间';
COMMENT ON COLUMN sys_user.remark IS '备注';

-- ----------------------------
-- 10、操作日志表
-- ----------------------------
DROP TABLE IF EXISTS sys_oper_log CASCADE;
CREATE TABLE sys_oper_log (
  oper_id           BIGSERIAL       NOT NULL,
  title             VARCHAR(50)     DEFAULT '',
  business_type     INT             DEFAULT 0,
  method            VARCHAR(100)    DEFAULT '',
  request_method    VARCHAR(10)     DEFAULT '',
  operator_type     INT             DEFAULT 0,
  oper_name         VARCHAR(50)     DEFAULT '',
  dept_name         VARCHAR(50)     DEFAULT '',
  oper_url          VARCHAR(255)    DEFAULT '',
  oper_ip           VARCHAR(128)    DEFAULT '',
  oper_location     VARCHAR(255)    DEFAULT '',
  oper_param        VARCHAR(2000)   DEFAULT '',
  json_result       VARCHAR(2000)   DEFAULT '',
  status            INT             DEFAULT 0,
  error_msg         VARCHAR(2000)   DEFAULT '',
  oper_time         TIMESTAMP,
  cost_time         BIGINT          DEFAULT 0,
  PRIMARY KEY (oper_id)
);
COMMENT ON TABLE sys_oper_log IS '操作日志记录';
COMMENT ON COLUMN sys_oper_log.oper_id IS '日志主键';
COMMENT ON COLUMN sys_oper_log.title IS '模块标题';
COMMENT ON COLUMN sys_oper_log.business_type IS '业务类型（0其它 1新增 2修改 3删除 4授权 5导出 6导入 7强退 8生成代码 9清空数据）';
COMMENT ON COLUMN sys_oper_log.method IS '方法名称';
COMMENT ON COLUMN sys_oper_log.request_method IS '请求方式';
COMMENT ON COLUMN sys_oper_log.operator_type IS '操作类别（0其它 1后台用户 2手机端用户）';
COMMENT ON COLUMN sys_oper_log.oper_name IS '操作人员';
COMMENT ON COLUMN sys_oper_log.dept_name IS '部门名称';
COMMENT ON COLUMN sys_oper_log.oper_url IS '请求URL';
COMMENT ON COLUMN sys_oper_log.oper_ip IS '主机地址';
COMMENT ON COLUMN sys_oper_log.oper_location IS '操作地点';
COMMENT ON COLUMN sys_oper_log.oper_param IS '请求参数';
COMMENT ON COLUMN sys_oper_log.json_result IS '返回参数';
COMMENT ON COLUMN sys_oper_log.status IS '操作状态（0正常 1异常）';
COMMENT ON COLUMN sys_oper_log.error_msg IS '错误消息';
COMMENT ON COLUMN sys_oper_log.oper_time IS '操作时间';
COMMENT ON COLUMN sys_oper_log.cost_time IS '消耗时间';

-- ----------------------------
-- 14、登录日志表
-- ----------------------------
DROP TABLE IF EXISTS sys_login_log CASCADE;
CREATE TABLE sys_login_log (
  info_id           BIGSERIAL       NOT NULL,
  user_name         VARCHAR(50)     DEFAULT '',
  ipaddr            VARCHAR(128)    DEFAULT '',
  login_location    VARCHAR(255)    DEFAULT '',
  browser           VARCHAR(50)     DEFAULT '',
  os                VARCHAR(50)     DEFAULT '',
  status            CHAR(1)         DEFAULT '0',
  msg               VARCHAR(255)    DEFAULT '',
  login_time        TIMESTAMP,
  PRIMARY KEY (info_id)
);
COMMENT ON TABLE sys_login_log IS '系统访问记录';
COMMENT ON COLUMN sys_login_log.info_id IS '访问ID';
COMMENT ON COLUMN sys_login_log.user_name IS '用户账号';
COMMENT ON COLUMN sys_login_log.ipaddr IS '登录IP地址';
COMMENT ON COLUMN sys_login_log.login_location IS '登录地点';
COMMENT ON COLUMN sys_login_log.browser IS '浏览器类型';
COMMENT ON COLUMN sys_login_log.os IS '操作系统';
COMMENT ON COLUMN sys_login_log.status IS '登录状态（0成功 1失败）';
COMMENT ON COLUMN sys_login_log.msg IS '提示消息';
COMMENT ON COLUMN sys_login_log.login_time IS '访问时间';