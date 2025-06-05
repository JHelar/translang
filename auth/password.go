package auth

import (
	"fmt"
	"translang/db"
)

type PasswordProvider struct {
	db *db.DBClient
}

func NewPasswordProvider(db *db.DBClient) PasswordProvider {
	return PasswordProvider{
		db: db,
	}
}

func (provider PasswordProvider) signIn(payload ProviderPayload) (User, error) {
	passwordUserPayload := payload.AsPasswordUserPayload()

	fmt.Printf("Email: %s\nPassword: %s\n", passwordUserPayload.Email, passwordUserPayload.Password)

	return User{}, nil
}

func (provider PasswordProvider) getKind() ProviderKind {
	return KindPasswordProvider
}
