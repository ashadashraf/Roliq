package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type Claims struct {
	Subject       string `json:"sub"`
	Issuer        string `json:"iss"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

type Verifier interface {
	Verify(ctx context.Context, rawToken string) (Claims, error)
}

type OIDCVerifier struct {
	verifier *oidc.IDTokenVerifier
}

func NewVerifier(ctx context.Context, issuer, jwksURL, audience string) (*OIDCVerifier, error) {
	config := &oidc.Config{ClientID: audience}
	if jwksURL != "" {
		return &OIDCVerifier{verifier: oidc.NewVerifier(issuer, oidc.NewRemoteKeySet(ctx, jwksURL), config)}, nil
	}
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}
	return &OIDCVerifier{verifier: provider.Verifier(config)}, nil
}

func (v *OIDCVerifier) Verify(ctx context.Context, rawToken string) (Claims, error) {
	token, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return Claims{}, fmt.Errorf("verify token: %w", err)
	}
	var claims Claims
	if err := token.Claims(&claims); err != nil {
		return Claims{}, fmt.Errorf("decode claims: %w", err)
	}
	claims.Email = strings.TrimSpace(strings.ToLower(claims.Email))
	if claims.Subject == "" || claims.Issuer == "" || claims.Email == "" || !claims.EmailVerified {
		return Claims{}, fmt.Errorf("token must contain subject, issuer, and a verified email")
	}
	return claims, nil
}
