package service

import (
	"database/sql"
	"testing"

	"navigation/internal/domain"
)

type authTestStore struct {
	user     domain.User
	settings domain.AppSettings
}

func (s *authTestStore) GetUser() (domain.User, error) {
	if s.user.Username == "" {
		return domain.User{}, sql.ErrNoRows
	}
	return s.user, nil
}

func (s *authTestStore) SaveUser(user domain.User) error {
	s.user = user
	return nil
}

func (s *authTestStore) GetSettings() (domain.AppSettings, error) {
	return s.settings, nil
}

func (s *authTestStore) SaveSettings(settings domain.AppSettings) error {
	s.settings = settings
	return nil
}

func TestNewAuthServiceGeneratesRandomInitialCredential(t *testing.T) {
	store := &authTestStore{}
	auth, err := NewAuthService(store)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}

	credential, ok := auth.InitialCredential()
	if !ok {
		t.Fatal("InitialCredential() missing")
	}
	if credential.Username != DefaultUsername {
		t.Fatalf("username = %q, want %q", credential.Username, DefaultUsername)
	}
	if credential.Password == "" || credential.Password == "admin" {
		t.Fatalf("generated password = %q, want non-empty non-default password", credential.Password)
	}
	if _, _, err := auth.Login(credential.Username, credential.Password); err != nil {
		t.Fatalf("Login() with generated password error = %v", err)
	}
	if _, _, err := auth.Login(DefaultUsername, "admin"); err == nil {
		t.Fatal("Login() with old default password succeeded")
	}

	nextAuth, err := NewAuthService(store)
	if err != nil {
		t.Fatalf("second NewAuthService() error = %v", err)
	}
	if _, ok := nextAuth.InitialCredential(); ok {
		t.Fatal("InitialCredential() should only be present when a user was created")
	}
}

func TestUpdateAccountPasswordClearsExistingSessions(t *testing.T) {
	store := &authTestStore{}
	auth, err := NewAuthService(store)
	if err != nil {
		t.Fatalf("NewAuthService() error = %v", err)
	}
	credential, ok := auth.InitialCredential()
	if !ok {
		t.Fatal("InitialCredential() missing")
	}
	firstToken, _, err := auth.Login(credential.Username, credential.Password)
	if err != nil {
		t.Fatalf("first Login() error = %v", err)
	}
	secondToken, _, err := auth.Login(credential.Username, credential.Password)
	if err != nil {
		t.Fatalf("second Login() error = %v", err)
	}

	user, passwordChanged, err := auth.UpdateAccount("owner", credential.Password, "strong-new-password")
	if err != nil {
		t.Fatalf("UpdateAccount() error = %v", err)
	}
	if !passwordChanged || user.Username != "owner" {
		t.Fatalf("user = %#v passwordChanged = %v, want renamed account with password change", user, passwordChanged)
	}
	if _, ok := auth.UserBySession(firstToken); ok {
		t.Fatal("first old session is still valid")
	}
	if _, ok := auth.UserBySession(secondToken); ok {
		t.Fatal("second old session is still valid")
	}
	if _, _, err := auth.Login("owner", "strong-new-password"); err != nil {
		t.Fatalf("Login() with new password error = %v", err)
	}
}
