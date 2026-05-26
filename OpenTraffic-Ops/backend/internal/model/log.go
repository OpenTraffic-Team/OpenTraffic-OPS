package model

// SysOperLog 操作日志模型
type SysOperLog struct {
	OperID        int64  `json:"operId" gorm:"primaryKey;column:oper_id;comment:日志主键"`
	Title         string `json:"title" gorm:"column:title;size:50;comment:模块标题"`
	BusinessType  int    `json:"businessType" gorm:"column:business_type;comment:业务类型（0其它 1新增 2修改 3删除）"`
	Method        string `json:"method" gorm:"column:method;size:100;comment:方法名称"`
	RequestMethod string `json:"requestMethod" gorm:"column:request_method;size:10;comment:请求方式"`
	OperatorType  int    `json:"operatorType" gorm:"column:operator_type;comment:操作类别（0其它 1后台用户 2手机端用户）"`
	OperName      string `json:"operName" gorm:"column:oper_name;size:50;comment:操作人员"`
	DeptName      string `json:"deptName" gorm:"column:dept_name;size:50;comment:部门名称"`
	OperURL       string `json:"operUrl" gorm:"column:oper_url;size:255;comment:请求URL"`
	OperIP        string `json:"operIp" gorm:"column:oper_ip;size:128;comment:主机地址"`
	OperLocation  string `json:"operLocation" gorm:"column:oper_location;size:255;comment:操作地点"`
	OperParam     string `json:"operParam" gorm:"column:oper_param;type:text;comment:请求参数"`
	JsonResult    string `json:"jsonResult" gorm:"column:json_result;type:text;comment:返回参数"`
	Status        int    `json:"status" gorm:"column:status;comment:操作状态（0正常 1异常）"`
	ErrorMsg      string `json:"errorMsg" gorm:"column:error_msg;size:2000;comment:错误消息"`
	OperTime      string `json:"operTime" gorm:"column:oper_time;comment:操作时间"`
	CostTime      int64  `json:"costTime" gorm:"column:cost_time;comment:消耗时间"`
}

func (SysOperLog) TableName() string {
	return "sys_oper_log"
}

// SysLoginLog 登录日志模型
type SysLoginLog struct {
	InfoID        int64  `json:"infoId" gorm:"primaryKey;column:info_id;comment:访问ID"`
	UserName      string `json:"userName" gorm:"column:user_name;size:50;comment:用户账号"`
	IPAddr        string `json:"ipaddr" gorm:"column:ipaddr;size:128;comment:登录IP地址"`
	LoginLocation string `json:"loginLocation" gorm:"column:login_location;size:255;comment:登录地点"`
	Browser       string `json:"browser" gorm:"column:browser;size:50;comment:浏览器类型"`
	OS            string `json:"os" gorm:"column:os;size:50;comment:操作系统"`
	Status        string `json:"status" gorm:"column:status;size:1;default:0;comment:登录状态（0成功 1失败）"`
	Msg           string `json:"msg" gorm:"column:msg;size:255;comment:提示消息"`
	LoginTime     string `json:"loginTime" gorm:"column:login_time;comment:访问时间"`
}

func (SysLoginLog) TableName() string {
	return "sys_login_log"
}
