package model

import "time"

// LoginUser 登录用户信息 - 与Java版本LoginUser等效
type LoginUser struct {
	UserID        int64     `json:"userId"`
	Token         string    `json:"token"`          // UUID
	LoginTime     int64     `json:"loginTime"`      // 毫秒时间戳
	ExpireTime    int64     `json:"expireTime"`     // 毫秒时间戳
	IPAddr        string    `json:"ipaddr"`
	LoginLocation string    `json:"loginLocation"`
	Browser       string    `json:"browser"`
	OS            string    `json:"os"`
	User          *SysUser  `json:"user"`
}

// IsExpired 检查是否过期
func (l *LoginUser) IsExpired() bool {
	return time.Now().UnixMilli() >= l.ExpireTime
}

// NeedRefresh 是否需要刷新（不足20分钟过期）
func (l *LoginUser) NeedRefresh() bool {
	return l.ExpireTime-time.Now().UnixMilli() <= 20*60*1000
}
