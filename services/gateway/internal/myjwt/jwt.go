package myjwt

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour

	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
}

type Claims struct {
	Type string `json:"type"`
	jwt.RegisteredClaims
}

func NewManager(accessSecret, refreshSecret string) (*TokenManager, error) {
	accessSecret = strings.TrimSpace(accessSecret)
	refreshSecret = strings.TrimSpace(refreshSecret)

	if len(accessSecret) < 32 {
		return nil, errors.New("access secret must be at least 32 bytes")
	}

	if len(refreshSecret) < 32 {
		return nil, errors.New("refresh secret must be at least 32 bytes")
	}

	if accessSecret == refreshSecret {
		return nil, errors.New("access and refresh secrets must be different")
	}

	return &TokenManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
	}, nil
}

func (m *TokenManager) createToken(
	userID string,
	tokenType string,
	ttl time.Duration,
	secret []byte,
) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", errors.New("user id is required")
	}

	now := time.Now()

	claims := Claims{
		Type: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        rand.Text(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign %s token: %w", tokenType, err)
	}

	return signedToken, nil
}

func (m *TokenManager) CreateAccessToken(userID string) (string, error) {
	return m.createToken(
		userID,
		accessTokenType,
		accessTokenTTL,
		m.accessSecret,
	)
}

func (m *TokenManager) CreateTokens(userID string) (string, string, error) {
	accessToken, err := m.CreateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := m.createToken(
		userID,
		refreshTokenType,
		refreshTokenTTL,
		m.refreshSecret,
	)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (m *TokenManager) VerifyAccessToken(tokenString string) (*Claims, error) {
	return m.verifyToken(
		tokenString,
		m.accessSecret,
		accessTokenType,
	)
}

func (m *TokenManager) VerifyRefreshToken(tokenString string) (*Claims, error) {
	return m.verifyToken(
		tokenString,
		m.refreshSecret,
		refreshTokenType,
	)
}

func (m *TokenManager) verifyToken(
	tokenString string,
	secret []byte,
	expectedType string,
) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return secret, nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, fmt.Errorf("verify %s token: %w", expectedType, err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Type != expectedType {
		return nil, errors.New("unexpected token type")
	}

	if claims.Subject == "" {
		return nil, errors.New("missing subject claim")
	}

	return claims, nil
}
