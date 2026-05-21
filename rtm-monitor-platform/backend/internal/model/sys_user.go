package model

// SysUser 用户模型
type SysUser struct {
	UserID      int64   `json:"userId" gorm:"primaryKey;column:user_id;comment:用户ID"`
	UserName    string  `json:"userName" gorm:"column:user_name;size:30;not null;uniqueIndex;comment:用户账号"`
	NickName    string  `json:"nickName" gorm:"column:nick_name;size:30;not null;comment:用户昵称"`
	UserType    string  `json:"userType" gorm:"column:user_type;size:2;default:00;comment:用户类型（00系统用户）"`
	Email       string  `json:"email" gorm:"column:email;size:50;comment:用户邮箱"`
	Phonenumber string  `json:"phonenumber" gorm:"column:phonenumber;size:11;comment:手机号码"`
	Sex         string  `json:"sex" gorm:"column:sex;size:1;default:0;comment:用户性别（0男 1女 2未知）"`
	Avatar      string  `json:"avatar" gorm:"column:avatar;size:100;comment:头像地址"`
	Password    string  `json:"-" gorm:"column:password;size:100;comment:密码"`
	Status      string  `json:"status" gorm:"column:status;size:1;default:0;comment:帐号状态（0正常 1停用）"`
	LoginIP     string  `json:"loginIp" gorm:"column:login_ip;size:128;comment:最后登录IP"`
	LoginDate   *string `json:"loginDate" gorm:"column:login_date;comment:最后登录时间"`
	BaseEntity
}

func (SysUser) TableName() string {
	return "sys_user"
}
