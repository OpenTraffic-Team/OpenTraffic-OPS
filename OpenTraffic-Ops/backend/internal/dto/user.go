package dto

// UserQuery 用户查询条件
type UserQuery struct {
	PageQuery
	UserName    string `form:"userName"`
	NickName    string `form:"nickName"`
	Phonenumber string `form:"phonenumber"`
	Status      string `form:"status"`
	BeginTime   string `form:"params[beginTime]"`
	EndTime     string `form:"params[endTime]"`
}

// UserCreateRequest 用户创建请求
type UserCreateRequest struct {
	UserName    string `json:"userName" binding:"required"`
	NickName    string `json:"nickName" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Email       string `json:"email"`
	Phonenumber string `json:"phonenumber"`
	Sex         string `json:"sex"`
	Status      string `json:"status"`
	Remark      string `json:"remark"`
}

// UserUpdateRequest 用户更新请求
type UserUpdateRequest struct {
	UserID      int64  `json:"userId" binding:"required"`
	UserName    string `json:"userName" binding:"required"`
	NickName    string `json:"nickName" binding:"required"`
	Email       string `json:"email"`
	Phonenumber string `json:"phonenumber"`
	Sex         string `json:"sex"`
	Status      string `json:"status"`
	Remark      string `json:"remark"`
}

// ResetPwdRequest 重置密码请求
type ResetPwdRequest struct {
	UserID   int64  `json:"userId" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangeStatusRequest 修改状态请求
type ChangeStatusRequest struct {
	UserID int64  `json:"userId" binding:"required"`
	Status string `json:"status" binding:"required"`
}
