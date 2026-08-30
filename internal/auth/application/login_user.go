package application

import (
	"strings"

	"github.com/armandoalvarado/sofia-backend/internal/auth/domain"
	"golang.org/x/crypto/bcrypt"
)

type LoginUserUseCase struct {
	repo domain.UserRepository
}

func NewLoginUserUseCase(repo domain.UserRepository) *LoginUserUseCase {
	return &LoginUserUseCase{repo: repo}
}

func (uc *LoginUserUseCase) Execute(email, password string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, domain.ErrInvalidEmail
	}
	if password == "" {
		return nil, domain.ErrInvalidPassword
	}

	user, err := uc.repo.FindByEmail(email)

	if user == nil || err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	match := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if match != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}
