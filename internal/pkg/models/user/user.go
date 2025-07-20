package user

type User struct {
	Login    string
	Password string
}

func NewUser(login, password string) *User {
	return &User{
		Login:    login,
		Password: password,
	}
}
