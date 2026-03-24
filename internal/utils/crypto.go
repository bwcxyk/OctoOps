package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"octoops/internal/config"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func getAesKey() ([]byte, error) {
	key := config.GetAliyunAesKey()
	if len(key) != 32 {
		return nil, errors.New("octoops.aliyun.aes_key must be 32 bytes (AES-256)")
	}
	return []byte(key), nil
}

// EncryptAES encrypts with AES-GCM and returns a versioned base64 string.
func EncryptAES(plainText string) (string, error) {
	key, err := getAesKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	cipherText := gcm.Seal(nil, nonce, []byte(plainText), nil)
	encoded := base64.StdEncoding.EncodeToString(append(nonce, cipherText...))
	return "gcm:" + encoded, nil
}

// DecryptAES decrypts AES-GCM values and falls back to legacy AES-CBC.
func DecryptAES(cipherBase64 string) (string, error) {
	key, err := getAesKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(cipherBase64, "gcm:") {
		raw := strings.TrimPrefix(cipherBase64, "gcm:")
		cipherText, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return "", err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return "", err
		}
		if len(cipherText) < gcm.NonceSize() {
			return "", errors.New("cipherText is too short")
		}
		nonce := cipherText[:gcm.NonceSize()]
		enc := cipherText[gcm.NonceSize():]
		plain, err := gcm.Open(nil, nonce, enc, nil)
		if err != nil {
			return "", err
		}
		return string(plain), nil
	}

	// Legacy AES-CBC (fixed IV) for backward compatibility
	cipherText, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err != nil {
		return "", err
	}
	if len(cipherText)%aes.BlockSize != 0 {
		return "", errors.New("cipherText is not a multiple of the block size")
	}
	iv := key[:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)
	plain := make([]byte, len(cipherText))
	mode.CryptBlocks(plain, cipherText)
	// Remove PKCS7 padding
	pad := int(plain[len(plain)-1])
	if pad > aes.BlockSize || pad == 0 {
		return "", errors.New("invalid padding")
	}
	return string(plain[:len(plain)-pad]), nil
}

// HashPassword 使用bcrypt加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword 验证密码
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
