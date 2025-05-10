package authservice

import (
	"context"
	"fmt"

	"github.com/ServiceWeaver/weaver"
)

type AuthService interface {
	Login(ctx context.Context, email string, password string) error
	CheckIsLoggedIn(ctx context.Context) error
}

type impl struct {
	weaver.Implements[AuthService]
	isLoggedIn bool
}

func (s *impl) Init(ctx context.Context) error {
	s.Logger(ctx).Info("Auth Service started")
	s.isLoggedIn = false
	return nil
}

func (s *impl) Login(ctx context.Context, email string, password string) error {
	if email == "" || password == "" {
		return fmt.Errorf("missing email or password")
	}

	s.isLoggedIn = true
	return nil
}

func (s *impl) CheckIsLoggedIn(ctx context.Context) error {
	if !s.isLoggedIn {
		return fmt.Errorf("not logged in")
	}
	return nil
}
