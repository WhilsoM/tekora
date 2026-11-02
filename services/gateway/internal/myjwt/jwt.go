package myjwt

import (
	"crypto/rand"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("secret-key")
var refreshSecret = []byte("secret-key2")

type Claims struct {
	Type string `json:"type"`
	jwt.RegisteredClaims
}

func createRefreshToken(userID string) (string, error) {
	refreshClaims := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{
			Type: "refresh",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userID,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ID:        rand.Text(),
			},
		},
	)

	refreshToken, err := refreshClaims.SignedString(refreshSecret)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func CreateAccessToken(userID string) (string, error) {
	accessClaims := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{
			Type: "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userID,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				ID:        rand.Text(),
			},
		},
	)

	accessToken, err := accessClaims.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func CreateTokens(userID string) (string, string, error) {
	if userID == "" {
		return "", "", errors.New("user id is required")
	}

	accessToken, err := CreateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := createRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func VerifyAccessToken(tokenString string) (*Claims, error) {
	return verifyToken(tokenString, secretKey, "access")
}

func VerifyRefreshToken(tokenString string) (*Claims, error) {
	return verifyToken(tokenString, refreshSecret, "refresh")
}

func verifyToken(
	tokenString string,
	secret []byte,
	expectedType string,
) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return secret, nil
		},
		jwt.WithValidMethods(
			[]string{jwt.SigningMethodHS256.Alg()},
		),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Type != expectedType {
		return nil, errors.New("unexpected type")
	}

	userID, err := claims.GetSubject()
	if err != nil || userID == "" {
		return nil, errors.New("missing or invalid sub claim")
	}

	return claims, nil
}
