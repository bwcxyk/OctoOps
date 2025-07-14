package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"octoops/internal/config"
	"golang.org/x/crypto/bcrypt"
)

// 获取加密密钥（32 字节，AES-256）
func getAesKey() ([]byte, error) {
	key := config.GetAliyunAesKey()
	if len(key) != 32 {
		return nil, errors.New("octoops.aliyun.aes_key must be 32 bytes (AES-256)")
	}
	return []byte(key), nil
}

// AES加密，返回base64字符串
func EncryptAES(plainText string) (string, error) {
	key, err := getAesKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	plain := []byte(plainText)
	if len(plain)%aes.BlockSize != 0 {
		pad := aes.BlockSize - len(plain)%aes.BlockSize
		for i := 0; i < pad; i++ {
			plain = append(plain, byte(pad))
		}
	}
	cipherText := make([]byte, len(plain))
	iv := key[:aes.BlockSize]
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, plain)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// AES解密，输入base64字符串
func DecryptAES(cipherBase64 string) (string, error) {
	key, err := getAesKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
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
	// 去除填充
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
