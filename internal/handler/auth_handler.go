package handler

type AuthHandler interface {
	Login(login, password string) (string, error)
	Register(login, password string) (string, error)
	Validate(token string) (string, error)
}
