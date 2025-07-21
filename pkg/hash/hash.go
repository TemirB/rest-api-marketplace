package hash

import (
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func EncryptPassword(password string) (string, error) {
	encryptedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(encryptedPwd), err
}

func ComparePasswords(encryptedPwd, plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encryptedPwd), []byte(plainPwd))
	if err != nil {
		zap.L().Error(
			"failed to compare passwords",
			zap.Error(err),
		)
		return false
	}
	return true
}
