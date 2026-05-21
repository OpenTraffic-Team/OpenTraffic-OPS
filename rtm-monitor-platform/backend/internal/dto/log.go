package dto

// OperLogQuery 操作日志查询条件
type OperLogQuery struct {
	PageQuery
	Title        string `form:"title"`
	OperName     string `form:"operName"`
	BusinessType *int   `form:"businessType"`
	Status       *int   `form:"status"`
	BeginTime    string `form:"params[beginTime]"`
	EndTime      string `form:"params[endTime]"`
}

// LoginLogQuery 登录日志查询条件
type LoginLogQuery struct {
	PageQuery
	UserName  string `form:"userName"`
	IPAddr    string `form:"ipaddr"`
	Status    string `form:"status"`
	BeginTime string `form:"params[beginTime]"`
	EndTime   string `form:"params[endTime]"`
}

// OperLogCreateRequest 操作日志创建请求
type OperLogCreateRequest struct {
	Title         string `json:"title"`
	BusinessType  int    `json:"businessType"`
	Method        string `json:"method"`
	RequestMethod string `json:"requestMethod"`
	OperatorType  int    `json:"operatorType"`
	OperName      string `json:"operName"`
	DeptName      string `json:"deptName"`
	OperURL       string `json:"operUrl"`
	OperIP        string `json:"operIp"`
	OperLocation  string `json:"operLocation"`
	OperParam     string `json:"operParam"`
	JsonResult    string `json:"jsonResult"`
	Status        int    `json:"status"`
	ErrorMsg      string `json:"errorMsg"`
	CostTime      int64  `json:"costTime"`
}

// LoginLogCreateRequest 登录日志创建请求
type LoginLogCreateRequest struct {
	UserName      string `json:"userName"`
	IPAddr        string `json:"ipaddr"`
	LoginLocation string `json:"loginLocation"`
	Browser       string `json:"browser"`
	OS            string `json:"os"`
	Status        string `json:"status"`
	Msg           string `json:"msg"`
}
