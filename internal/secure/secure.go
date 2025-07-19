package secure

import (
	"golang.org/x/crypto/bcrypt"
)

func EncryptPassword(pwd string) (string, error) {
	encryptedPwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(encryptedPwd), err
}

func ComparePasswords(encryptedPwd, plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encryptedPwd), []byte(plainPwd))
	return err == nil
}
