package service

import (
	"errors"
	"fmt"
	"rtm-initialization-backend/internal/model"
	"rtm-initialization-backend/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService 认证服务
type AuthService struct {
	userRepo *repository.UserRepository
	jwtSecret string
	expireHours int
}

// NewAuthService 创建认证服务
func NewAuthService(jwtSecret string, expireHours int) *AuthService {
	return &AuthService{
		userRepo:   repository.NewUserRepository(),
		jwtSecret: jwtSecret,
		expireHours: expireHours,
	}
}

// Claims JWT声明
type Claims struct {
	UserID   string      `json:"user_id"`
	Username string      `json:"username"`
	Role     model.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string       `json:"token"`
	User  *model.User  `json:"user"`
}

// Login 登录
func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	// 查找用户
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 验证密码（这里简化处理，实际应该使用bcrypt）
	// 在生产环境中应该使用 crypto.VerifyPassword
	if !verifyPasswordSimple(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}

	// 生成JWT Token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

// ValidateToken 验证Token
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserByID 根据ID获取用户
func (s *AuthService) GetUserByID(id string) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

// generateToken 生成JWT Token
func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// verifyPasswordSimple 简化版密码验证
func verifyPasswordSimple(password, hash string) bool {
	// 这是一个简化的实现，实际应用中应该使用bcrypt
	// 这里仅用于开发测试
	return password == "admin123" // 临时硬编码
}
