package dto

// LoginRequest 登录请求 - 与Java LoginBody等效
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Code     string `json:"code"`
	UUID     string `json:"uuid"`
}

// CaptchaImageResponse 验证码响应
type CaptchaImageResponse struct {
	CaptchaEnabled bool   `json:"captchaEnabled"`
	UUID           string `json:"uuid,omitempty"`
	Img            string `json:"img,omitempty"`
	MathStr        string `json:"mathStr,omitempty"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	User        *SysUserVO `json:"user"`
	Roles       []string   `json:"roles"`
	Permissions []string   `json:"permissions"`
}

// SysUserVO 用户视图对象
type SysUserVO struct {
	UserID      int64  `json:"userId"`
	UserName    string `json:"userName"`
	NickName    string `json:"nickName"`
	UserType    string `json:"userType"`
	Email       string `json:"email"`
	Phonenumber string `json:"phonenumber"`
	Sex         string `json:"sex"`
	Avatar      string `json:"avatar"`
	Status      string `json:"status"`
	LoginIP     string `json:"loginIp"`
	LoginDate   string `json:"loginDate"`
}

// RouterMenu Vue路由菜单
type RouterMenu struct {
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Hidden    bool         `json:"hidden"`
	Component string       `json:"component,omitempty"`
	Children  []RouterMenu `json:"children,omitempty"`
	Meta      RouterMeta   `json:"meta"`
}

// RouterMeta 路由元数据
type RouterMeta struct {
	Title   string `json:"title"`
	Icon    string `json:"icon"`
	NoCache bool   `json:"noCache"`
	Link    string `json:"link,omitempty"`
}

// RSAKeyPairResponse RSA密钥对响应
type RSAKeyPairResponse struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}
