package auth

import (
	"translang/db"
	"translang/dto"

	"golang.org/x/crypto/bcrypt"
)

type PasswordProvider struct {
	db *db.DBClient
}

func NewPasswordProvider(db *db.DBClient) PasswordProvider {
	return PasswordProvider{
		db: db,
	}
}

func (provider PasswordProvider) signIn(payload *ProviderPayload) (User, error) {
	passwordUserPayload := payload.AsPasswordUserPayload()

	user, err := dto.GetPasswordUserByEmail(passwordUserPayload.Email, provider.db)
	if err != nil {
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(passwordUserPayload.Password))
	if err != nil {
		return User{}, ErrInvalidUserCredentials
	}

	return User{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

func (provider PasswordProvider) getKind() ProviderKind {
	return KindPasswordProvider
}
