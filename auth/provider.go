package auth

import (
	"fmt"
)

type User struct {
	ID    int64
	Email string
}

type ProviderKind int

const (
	KindPasswordProvider ProviderKind = iota
)

var (
	ErrInvalidUserCredentials error = fmt.Errorf("invalid user credentials")
)

type ProviderPayload struct {
	Kind ProviderKind
	data providerPayloadData
}

type providerPayloadData interface {
	AsPayload() *ProviderPayload
}

func newPayload(kind ProviderKind, data providerPayloadData) *ProviderPayload {
	return &ProviderPayload{
		data: data,
		Kind: kind,
	}
}

type ProviderPayloadBase struct {
	ProviderPayload
}

func (payload *ProviderPayloadBase) AsPayload() *ProviderPayload {
	return &payload.ProviderPayload
}

type UserPasswordPayload struct {
	ProviderPayloadBase
	Email    string
	Password string
}

func (payload *ProviderPayload) AsPasswordUserPayload() *UserPasswordPayload {
	return payload.data.(*UserPasswordPayload)
}

func NewPasswordUserPayload(email string, password string) *ProviderPayload {
	data := UserPasswordPayload{
		Email:    email,
		Password: password,
	}
	return newPayload(KindPasswordProvider, &data)
}

type Provider interface {
	signIn(payload *ProviderPayload) (User, error)
	getKind() ProviderKind
}

type AuthProvider struct {
	providers map[ProviderKind]Provider
}

func NewAuthProvider() AuthProvider {
	return AuthProvider{
		providers: make(map[ProviderKind]Provider),
	}
}

func (auth AuthProvider) AddProvider(provider Provider) {
	auth.providers[provider.getKind()] = provider
}

func (auth AuthProvider) SignIn(payload *ProviderPayload) (User, error) {
	provider := auth.providers[payload.Kind]
	if provider == nil {
		return User{}, fmt.Errorf("missing provider for kind")
	}

	return provider.signIn(payload)
}
