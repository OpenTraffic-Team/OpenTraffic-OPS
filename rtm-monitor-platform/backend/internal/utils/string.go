package utils

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"regexp"
	"strings"
	"unicode"
)

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			result[i] = charset[i%len(charset)]
		} else {
			result[i] = charset[n.Int64()]
		}
	}
	return string(result)
}

// GenerateRandomBytes 生成指定长度的随机字节
func GenerateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomBase64 生成指定长度的Base64编码随机字符串
func GenerateRandomBase64(length int) string {
	b, _ := GenerateRandomBytes(length)
	return base64.URLEncoding.EncodeToString(b)
}

// IsEmpty 判断字符串是否为空
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotEmpty 判断字符串是否不为空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// TrimSpace 去除字符串前后空格
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// Contains 判断字符串是否包含子串
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// HasPrefix 判断字符串是否以指定前缀开头
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix 判断字符串是否以指定后缀结尾
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// ToLower 转小写
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToUpper 转大写
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// SubString 截取字符串（支持中文）
func SubString(s string, start, length int) string {
	runes := []rune(s)
	if start >= len(runes) {
		return ""
	}
	end := start + length
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

// RemoveHtmlTags 移除HTML标签
func RemoveHtmlTags(s string) string {
	re := regexp.MustCompile(`<[^?>]*>`)
	return re.ReplaceAllString(s, "")
}

// EscapeHtml 转义HTML特殊字符
func EscapeHtml(s string) string {
	replacer := strings.NewReplacer(
		"\u0026", "\u0026amp;",
		"\u003c", "\u0026lt;",
		"\u003e", "\u0026gt;",
		"\"", "\u0026quot;",
		"'", "\u0026#39;",
	)
	return replacer.Replace(s)
}

// StripXSS 简单的XSS过滤
func StripXSS(s string) string {
	if IsEmpty(s) {
		return s
	}
	s = EscapeHtml(s)
	// 过滤危险事件
	dangerous := []string{"javascript:", "onerror", "onload", "onclick", "onmouseover"}
	lower := strings.ToLower(s)
	for _, d := range dangerous {
		if strings.Contains(lower, d) {
			return ""
		}
	}
	return s
}

// CamelToSnake 驼峰转下划线
func CamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteByte('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// SnakeToCamel 下划线转驼峰
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
