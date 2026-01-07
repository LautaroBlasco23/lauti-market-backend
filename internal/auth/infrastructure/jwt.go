package infrastructure

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
)

type JWTGenerator struct {
	secret     []byte
	expiration time.Duration
}

func NewJWTGenerator(secret string, expiration time.Duration) *JWTGenerator {
	return &JWTGenerator{
		secret:     []byte(secret),
		expiration: expiration,
	}
}

type Claims struct {
	AuthID string `json:"auth_id"`
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func (g *JWTGenerator) Generate(authID domain.ID, userID domain.UserID) (string, error) {
	claims := Claims{
		AuthID: string(authID),
		UserID: string(userID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(g.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secret)
}

func (g *JWTGenerator) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return g.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
