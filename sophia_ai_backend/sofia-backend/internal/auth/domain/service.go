package domain

type UserService interface {
	Register(name, email, password string) (*User, error)
	Login(email, password string) (*User, error)
}
