package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"octoops/internal/config"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

func getAesKey() ([]byte, error) {
	key := config.GetAesKey()
	if len(key) != 32 {
		return nil, errors.New("octoops.security.aes_key must be 32 bytes (AES-256)")
	}
	return []byte(key), nil
}

// aesPrefix 密文前缀，用于区分格式
const aesPrefix = "gcm:"

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
	return aesPrefix + encoded, nil
}

// DecryptAES decrypts AES-GCM values produced by EncryptAES.
func DecryptAES(cipherBase64 string) (string, error) {
	key, err := getAesKey()
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(cipherBase64, aesPrefix) {
		return "", errors.New("密文格式不正确，应为 " + aesPrefix + " 开头的 AES-GCM 密文")
	}
	raw := strings.TrimPrefix(cipherBase64, aesPrefix)
	cipherText, err := base64.StdEncoding.DecodeString(raw)
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

// ValidatePasswordComplexity validates password strength:
// at least 8 characters and at least 3 of these 4 categories:
// lowercase letters, uppercase letters, digits, special characters.
func ValidatePasswordComplexity(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return errors.New("密码长度至少为8位")
	}

	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	categories := 0
	if hasLower {
		categories++
	}
	if hasUpper {
		categories++
	}
	if hasDigit {
		categories++
	}
	if hasSpecial {
		categories++
	}

	if categories < 3 {
		return errors.New("密码需包含数字、大写字母、小写字母、特殊字符中的至少三种")
	}

	return nil
}
