package captcha

import (
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"

	"rtm-server/pkg/cache"
)

const (
	captchaKeyPrefix = "captcha_codes:"
	captchaExpire    = 2 * time.Minute
)

// redisStore 实现 base64Captcha.Store 接口
type redisStore struct {
	cache  *cache.Cache
	prefix string
	expire time.Duration
}

func (s *redisStore) Set(id string, value string) error {
	key := s.prefix + id
	return s.cache.Set(key, value, s.expire)
}

func (s *redisStore) Get(id string, clear bool) string {
	key := s.prefix + id
	val, err := s.cache.Get(key)
	if err != nil {
		return ""
	}
	if clear {
		_ = s.cache.Delete(key)
	}
	return val
}

func (s *redisStore) Verify(id, answer string, clear bool) bool {
	val := s.Get(id, clear)
	return strings.EqualFold(val, answer)
}

// CaptchaService 验证码服务
type CaptchaService struct {
	captcha *base64Captcha.Captcha
	store   *redisStore
}

// NewCaptchaService 创建验证码服务
func NewCaptchaService(client *redis.Client) *CaptchaService {
	store := &redisStore{
		cache:  cache.NewCache(client),
		prefix: captchaKeyPrefix,
		expire: captchaExpire,
	}

	// 使用数字验证码驱动: 高80, 宽240, 长度6, 扭曲度0.7, 噪点80
	driver := base64Captcha.NewDriverDigit(80, 240, 6, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, store)

	return &CaptchaService{
		captcha: captcha,
		store:   store,
	}
}

// Generate 生成验证码，返回 uuid 和 base64 图片
func (s *CaptchaService) Generate() (uuid, imgBase64 string, err error) {
	id, b64s, _, err := s.captcha.Generate()
	return id, b64s, err
}

// Verify 验证验证码
func (s *CaptchaService) Verify(uuid, code string) bool {
	if uuid == "" || code == "" {
		return false
	}
	return s.store.Verify(uuid, code, true)
}
