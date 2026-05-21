package utils

import "strings"

// ParseUserAgent 解析User-Agent，返回(浏览器, 操作系统)
func ParseUserAgent(ua string) (browser, os string) {
	if ua == "" {
		return "未知", "未知"
	}

	ua = strings.ToLower(ua)

	// 解析操作系统
	if strings.Contains(ua, "windows") {
		os = "Windows"
	} else if strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os") {
		os = "Mac OS"
	} else if strings.Contains(ua, "linux") {
		os = "Linux"
	} else if strings.Contains(ua, "android") {
		os = "Android"
	} else if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		os = "iOS"
	} else {
		os = "未知"
	}

	// 解析浏览器
	if strings.Contains(ua, "edg") {
		browser = "Edge"
	} else if strings.Contains(ua, "chrome") && !strings.Contains(ua, "chromium") {
		browser = "Chrome"
	} else if strings.Contains(ua, "firefox") {
		browser = "Firefox"
	} else if strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome") {
		browser = "Safari"
	} else if strings.Contains(ua, "opera") || strings.Contains(ua, "opr") {
		browser = "Opera"
	} else if strings.Contains(ua, "trident") || strings.Contains(ua, "msie") {
		browser = "IE"
	} else {
		browser = "未知"
	}

	return browser, os
}
