package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	mrand "math/rand"
	"time"
)

func GenerateSalt(length int) (string, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

func HashPassword(password, salt string) string {
	hasher := sha256.New()
	hasher.Write([]byte(salt + password))
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func VerifyPassword(hash, password, salt string) bool {
	return hash == HashPassword(password, salt)
}

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := mrand.New(mrand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GenerateMd5(data string) string {
	hash := md5.New()
	hash.Write([]byte(data))
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}
