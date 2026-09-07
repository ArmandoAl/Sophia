package application

import (
	"strings"

	"github.com/armandoalvarado/sofia-backend/internal/auth/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserUseCase struct {
	repo domain.UserRepository
}

func NewRegisterUserUseCase(repo domain.UserRepository) *RegisterUserUseCase {
	return &RegisterUserUseCase{repo: repo}
}

func (uc *RegisterUserUseCase) Execute(name, email, password string) (*domain.User, error) {
	name = strings.TrimSpace(name)
	email = normalizeEmail(email)

	if name == "" {
		return nil, domain.ErrInvalidName
	}
	if email == "" {
		return nil, domain.ErrInvalidEmail
	}
	if password == "" {
		return nil, domain.ErrInvalidPassword
	}
	if len(password) < 8 {
		return nil, domain.ErrWeakPassword
	}

	existingUser, errFind := uc.repo.FindByEmail(email)

	if errFind != nil {
		return nil, domain.ErrInternalServerError
	}

	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hashedPassword, errHash := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if errHash != nil {
		return nil, domain.ErrInternalServerError
	}

	user := domain.NewUser(
		generateID(),
		name,
		email,
		string(hashedPassword),
	)

	var errSave error
	if saver, ok := uc.repo.(domain.TransactionalDefaultSaver); ok {
		errSave = saver.SaveWithDefaults(user)
	} else {
		errSave = uc.repo.Save(user)
	}
	if errSave != nil {
		return nil, domain.ErrInternalServerError
	}

	return user, nil
}

func generateID() string {
	return uuid.New().String()
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
