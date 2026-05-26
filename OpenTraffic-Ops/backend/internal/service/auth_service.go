package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"opentraffic-ops-backend/internal/config"
	"opentraffic-ops-backend/internal/constant"
	"opentraffic-ops-backend/internal/dto"
	"opentraffic-ops-backend/internal/model"
	"opentraffic-ops-backend/internal/repository"
	"opentraffic-ops-backend/internal/utils"
	"opentraffic-ops-backend/pkg/cache"
	"opentraffic-ops-backend/pkg/captcha"
	"opentraffic-ops-backend/pkg/crypto"
	"opentraffic-ops-backend/pkg/jwt"
)

const loginTokenPrefix = "login_tokens:"

// AuthService 认证服务
type AuthService struct {
	userRepo       *repository.UserRepository
	cache          *cache.Cache
	loginLogService *LoginLogService
	captchaService *captcha.CaptchaService
	jwtSecret      string
	jwtExpire      time.Duration
	captchaType    string
}

// NewAuthService 创建认证服务
func NewAuthService(db *gorm.DB, redisClient *redis.Client, loginLogService ...*LoginLogService) *AuthService {
	cfg := config.GlobalConfig
	expireTime := time.Duration(cfg.JWT.ExpireTime) * time.Minute

	s := &AuthService{
		userRepo:       repository.NewUserRepository(db),
		cache:          cache.NewCache(redisClient),
		jwtSecret:      cfg.JWT.Secret,
		jwtExpire:      expireTime,
		captchaType:    cfg.App.CaptchaType,
		captchaService: captcha.NewCaptchaService(redisClient),
	}
	if len(loginLogService) > 0 {
		s.loginLogService = loginLogService[0]
	}
	return s
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest, clientIP string, userAgent string) (string, error) {
	browser, os := utils.ParseUserAgent(userAgent)

	// 1. 验证码校验
	if !s.captchaService.Verify(req.UUID, req.Code) {
		zap.L().Warn("登录失败：验证码错误", zap.String("username", req.Username), zap.String("ip", clientIP))
		s.recordLoginLog(ctx, req.Username, "1", clientIP, browser, os, "验证码错误")
		return "", fmt.Errorf("验证码错误或已过期")
	}

	// 2. 查找用户
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		zap.L().Warn("登录失败：用户不存在", zap.String("username", req.Username), zap.String("ip", clientIP))
		s.recordLoginLog(ctx, req.Username, "1", clientIP, browser, os, "用户不存在")
		return "", fmt.Errorf("用户不存在/密码错误")
	}

	// 3. 检查用户状态
	if user.Status == "1" {
		s.recordLoginLog(ctx, req.Username, "1", clientIP, browser, os, "账号已停用")
		return "", fmt.Errorf("对不起，您的账号已停用")
	}

	// 4. 解密密码（RSA）并验证（BCrypt）
	rsaPair := crypto.GetRSAKeyPair()
	decryptedPassword, err := crypto.RSADecryptByPrivateKey(rsaPair.PrivateKey, req.Password)
	if err != nil {
		zap.L().Warn("登录失败：密码解密失败", zap.String("username", req.Username), zap.Error(err))
		s.recordLoginLog(ctx, req.Username, "1", clientIP, browser, os, "密码解密失败")
		return "", fmt.Errorf("用户不存在/密码错误")
	}

	if !crypto.VerifyPassword(user.Password, decryptedPassword) {
		zap.L().Warn("登录失败：密码错误", zap.String("username", req.Username), zap.String("ip", clientIP))
		s.recordLoginLog(ctx, req.Username, "1", clientIP, browser, os, "密码错误")
		return "", fmt.Errorf("用户不存在/密码错误")
	}

	// 5. 生成UUID
	uuid := utils.GenerateRandomBase64(16)

	// 6. 构建LoginUser并存储到Redis
	nowMs := time.Now().UnixMilli()
	expireMs := nowMs + int64(s.jwtExpire/time.Millisecond)

	loginUser := &model.LoginUser{
		UserID:     user.UserID,
		Token:      uuid,
		LoginTime:  nowMs,
		ExpireTime: expireMs,
		IPAddr:     clientIP,
		User:       user,
	}

	loginUserJSON, _ := json.Marshal(loginUser)
	loginKey := loginTokenPrefix + uuid
	s.cache.Set(loginKey, string(loginUserJSON), s.jwtExpire)

	// 7. 更新登录信息
	nowStr := utils.NowStr()
	s.userRepo.UpdateLoginInfo(ctx, user.UserID, clientIP, nowStr)

	// 8. 生成JWT Token
	token, err := jwt.GenerateToken(user.UserID, user.UserName, uuid, s.jwtSecret, s.jwtExpire)
	if err != nil {
		s.recordLoginLog(ctx, req.Username, "1", clientIP, browser, os, "生成令牌失败")
		return "", fmt.Errorf("生成令牌失败: %w", err)
	}

	// 记录登录成功日志
	s.recordLoginLog(ctx, req.Username, "0", clientIP, browser, os, "登录成功")

	zap.L().Info("用户登录成功",
		zap.String("username", user.UserName),
		zap.Int64("userId", user.UserID),
		zap.String("ip", clientIP),
	)

	return token, nil
}

// GenerateCaptcha 生成验证码
func (s *AuthService) GenerateCaptcha() (string, string, error) {
	return s.captchaService.Generate()
}

// Logout 用户登出
func (s *AuthService) Logout(uuid string) error {
	loginKey := loginTokenPrefix + uuid
	return s.cache.Delete(loginKey)
}

// GetLoginUser 从Redis获取登录用户
func (s *AuthService) GetLoginUser(uuid string) (*model.LoginUser, error) {
	loginKey := loginTokenPrefix + uuid
	jsonStr, err := s.cache.Get(loginKey)
	if err != nil {
		return nil, err
	}

	var loginUser model.LoginUser
	if err := json.Unmarshal([]byte(jsonStr), &loginUser); err != nil {
		return nil, err
	}

	return &loginUser, nil
}

// RefreshToken 刷新Token有效期
func (s *AuthService) RefreshToken(loginUser *model.LoginUser) error {
	nowMs := time.Now().UnixMilli()
	expireMs := nowMs + int64(s.jwtExpire/time.Millisecond)
	loginUser.LoginTime = nowMs
	loginUser.ExpireTime = expireMs

	loginUserJSON, _ := json.Marshal(loginUser)
	loginKey := loginTokenPrefix + loginUser.Token
	return s.cache.Set(loginKey, string(loginUserJSON), s.jwtExpire)
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(ctx context.Context, loginUser *model.LoginUser) (*dto.UserInfoResponse, error) {
	user := loginUser.User
	if user == nil {
		return nil, fmt.Errorf("用户信息不存在")
	}

	// 转换用户VO
	userVO := &dto.SysUserVO{
		UserID:      user.UserID,
		UserName:    user.UserName,
		NickName:    user.NickName,
		UserType:    user.UserType,
		Email:       user.Email,
		Phonenumber: user.Phonenumber,
		Sex:         user.Sex,
		Avatar:      user.Avatar,
		Status:      user.Status,
		LoginIP:     user.LoginIP,
		LoginDate:   "",
	}
	if user.LoginDate != nil {
		userVO.LoginDate = *user.LoginDate
	}

	// 简化权限：管理员返回全部权限，普通用户返回基本权限
	var permissions []string
	var roles []string
	if user.UserID == constant.SuperAdminUserID {
		roles = []string{"admin"}
		permissions = []string{"*:*:*"}
	} else {
		roles = []string{"common"}
		permissions = []string{}
	}

	return &dto.UserInfoResponse{
		User:        userVO,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}

// GetRouters 获取路由菜单（简化版，返回固定路由）
func (s *AuthService) GetRouters(ctx context.Context, userID int64) ([]dto.RouterMenu, error) {
	// 简化路由：返回监控平台核心路由
	return []dto.RouterMenu{
		{
			Name:      "Monitor",
			Path:      "/monitor",
			Hidden:    false,
			Component: "Layout",
			Meta: dto.RouterMeta{
				Title: "监控平台",
				Icon:  "monitor",
			},
			Children: []dto.RouterMenu{
				{
					Name:      "HostInfo",
					Path:      "host-info",
					Component: "business/host-info/index",
					Meta: dto.RouterMeta{
						Title: "主机信息",
						Icon:  "server",
					},
				},
				{
					Name:      "HostHealth",
					Path:      "host-health",
					Component: "business/host-health/index",
					Meta: dto.RouterMeta{
						Title: "主机健康度",
						Icon:  "chart",
					},
				},
			},
		},
		{
			Name:      "System",
			Path:      "/system",
			Hidden:    false,
			Component: "Layout",
			Meta: dto.RouterMeta{
				Title: "系统管理",
				Icon:  "system",
			},
			Children: []dto.RouterMenu{
				{
					Name:      "User",
					Path:      "user",
					Component: "system/user/index",
					Meta: dto.RouterMeta{
						Title: "用户管理",
						Icon:  "user",
					},
				},
				{
					Name:      "OperLog",
					Path:      "oper-log",
					Component: "monitor/operlog/index",
					Meta: dto.RouterMeta{
						Title: "操作日志",
						Icon:  "log",
					},
				},
				{
					Name:      "LoginLog",
					Path:      "login-log",
					Component: "monitor/logininfor/index",
					Meta: dto.RouterMeta{
						Title: "登录日志",
						Icon:  "logininfor",
					},
				},
			},
		},
	}, nil
}

// recordLoginLog 记录登录日志
func (s *AuthService) recordLoginLog(ctx context.Context, userName, status, ip, browser, os, msg string) {
	if s.loginLogService == nil {
		return
	}
	go func() {
		_ = s.loginLogService.Create(ctx, &dto.LoginLogCreateRequest{
			UserName: userName,
			IPAddr:   ip,
			Browser:  browser,
			OS:       os,
			Status:   status,
			Msg:      msg,
		})
	}()
}
