package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"sync"
)

var (
	rsaKeyPair     *RSAKeyPair
	rsaKeyPairOnce sync.Once
)

// RSAKeyPair RSA密钥对
type RSAKeyPair struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

// GetRSAKeyPair 获取RSA密钥对（单例，首次调用时生成）
func GetRSAKeyPair() *RSAKeyPair {
	rsaKeyPairOnce.Do(func() {
		pair, err := GenerateRSAKeyPair()
		if err != nil {
			panic(fmt.Sprintf("failed to generate RSA key pair: %v", err))
		}
		rsaKeyPair = pair
	})
	return rsaKeyPair
}

// GenerateRSAKeyPair 生成新的RSA密钥对（1024位，与Java一致）
func GenerateRSAKeyPair() (*RSAKeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// PKCS#8私钥
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	privateKeyStr := base64.StdEncoding.EncodeToString(privateKeyDER)

	// X509公钥
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyStr := base64.StdEncoding.EncodeToString(publicKeyDER)

	return &RSAKeyPair{
		PublicKey:  publicKeyStr,
		PrivateKey: privateKeyStr,
	}, nil
}

// RSADecryptByPrivateKey 使用私钥解密（与Java decryptByPrivateKey一致）
func RSADecryptByPrivateKey(privateKeyStr, encryptedText string) (string, error) {
	privateKeyDER, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyDER)
	if err != nil {
		// 尝试PKCS#1格式
		privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyDER)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %w", err)
		}
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("not an RSA private key")
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, encryptedBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(decrypted), nil
}
