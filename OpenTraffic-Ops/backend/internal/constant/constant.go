package constant

// HTTP请求方法
const (
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"
)

// 用户状态
const (
	UserStatusNormal  = "0"
	UserStatusDisable = "1"
)

// 删除标志
const (
	DelFlagExist  = "0"
	DelFlagDelete = "2"
)

// 是否
const (
	Yes = "1"
	No  = "0"
)

// 操作状态
const (
	OperStatusSuccess = "0"
	OperStatusFail    = "1"
)

// 登录状态
const (
	LoginStatusSuccess = "0"
	LoginStatusFail    = "1"
)

// 验证码类型
const (
	CaptchaTypeMath = "math"
	CaptchaTypeChar = "char"
)

// JWT相关
const (
	JWTTokenPrefix  = "Bearer "
	JWTLoginUserKey = "login_user_key"
)

// Redis Key前缀
const (
	RedisKeyCaptcha       = "captcha_codes:"
	RedisKeyLoginToken    = "login_tokens:"
	RedisKeyLoginUser     = "login_user_key:"
	RedisKeyOnlineUser    = "online_user:"
	RedisKeyPasswordError = "pwd_error_count:"
	RedisKeyCacheable     = "cacheable:"
)

// 超级管理员
const (
	SuperAdminUserID = 1
)

// 字符编码
const (
	UTF8 = "UTF-8"
	GBK  = "GBK"
)

// 内容类型
const (
	ContentTypeJSON      = "application/json"
	ContentTypeForm      = "application/x-www-form-urlencoded"
	ContentTypeMultipart = "multipart/form-data"
	ContentTypeExcel     = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
)

// Agent聊天消息角色
const (
	ChatRoleUser      = "user"
	ChatRoleAssistant = "assistant"
)

// Agent聊天会话类型
const (
	ChatSessionTypeControl  = "control"
	ChatSessionTypePerceive = "perceive"
)

// 告警通道类型
const (
	ChannelTypePlatform = "platform"
	ChannelTypeEmail    = "email"
	ChannelTypeDingTalk = "dingtalk"
	ChannelTypeWechat   = "wechat"
)

// 告警通知状态
const (
	NotifyStatusPending = "pending"
	NotifyStatusSuccess = "success"
	NotifyStatusFailed  = "failed"
)
