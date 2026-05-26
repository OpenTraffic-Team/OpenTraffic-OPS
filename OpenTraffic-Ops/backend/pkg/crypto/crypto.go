package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// RSAPrivateKeyToBase64 将RSA私钥转为Base64编码的PKCS#1格式
func RSAPrivateKeyToBase64(key *rsa.PrivateKey) string {
	der := x509.MarshalPKCS1PrivateKey(key)
	return base64.StdEncoding.EncodeToString(der)
}

// RSAPublicKeyToBase64 将RSA公钥转为Base64编码的PKCS#1格式
func RSAPublicKeyToBase64(key *rsa.PublicKey) string {
	der := x509.MarshalPKCS1PublicKey(key)
	return base64.StdEncoding.EncodeToString(der)
}

// RSAPublicKeyToPEM 将RSA公钥转为PEM格式
func RSAPublicKeyToPEM(key *rsa.PublicKey) string {
	der := x509.MarshalPKCS1PublicKey(key)
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: der,
	}
	return string(pem.EncodeToMemory(block))
}

// RSADecryptBase64 使用Base64编码的私钥解密数据
func RSADecryptBase64(encryptedData, base64PrivateKey string) (string, error) {
	privateKeyDer, err := base64.StdEncoding.DecodeString(base64PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 private key: %w", err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyDer)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 encrypted data: %w", err)
	}

	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(decrypted), nil
}

// HashPassword 使用BCrypt哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword 验证密码
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateRandomBase64 生成指定长度的Base64随机字符串
func GenerateRandomBase64(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
