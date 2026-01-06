package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
)

type JWTService struct {
	secret               []byte
	issuer               string
	audience             string
	accessTokenLifetime  time.Duration
	refreshTokenLifetime time.Duration
}

func NewJWTService(secret, issuer, audience string, accessLifetime, refreshLifetime time.Duration) *JWTService {
	return &JWTService{
		secret:               []byte(secret),
		issuer:               issuer,
		audience:             audience,
		accessTokenLifetime:  accessLifetime,
		refreshTokenLifetime: refreshLifetime,
	}
}

// GenerateAccessToken creates a new access token with user claims
func (s *JWTService) GenerateAccessToken(userID, email, role, sessionID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessTokenLifetime)

	claims := jwt.MapClaims{
		"user_id":    userID,
		"email":      email,
		"role":       role,
		"session_id": sessionID,
		"iat":        now.Unix(),
		"exp":        expiresAt.Unix(),
		"iss":        s.issuer,
		"aud":        s.audience,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateRefreshToken creates a new refresh token
func (s *JWTService) GenerateRefreshToken(userID, sessionID string, lifetime time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(lifetime)

	claims := jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"iat":        now.Unix(),
		"exp":        expiresAt.Unix(),
		"iss":        s.issuer,
		"aud":        s.audience,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateAccessToken validates and parses an access token
func (s *JWTService) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Verify issuer and audience
	if claims["iss"] != s.issuer {
		return nil, ErrInvalidToken
	}
	if claims["aud"] != s.audience {
		return nil, ErrInvalidToken
	}

	// Extract claims
	accessClaims := &AccessTokenClaims{
		UserID:    getStringClaim(claims, "user_id"),
		Email:     getStringClaim(claims, "email"),
		Role:      getStringClaim(claims, "role"),
		SessionID: getStringClaim(claims, "session_id"),
		IssuedAt:  getInt64Claim(claims, "iat"),
		ExpiresAt: getInt64Claim(claims, "exp"),
		Issuer:    getStringClaim(claims, "iss"),
		Audience:  getStringClaim(claims, "aud"),
	}

	return accessClaims, nil
}

// ValidateRefreshToken validates and parses a refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Verify issuer and audience
	if claims["iss"] != s.issuer {
		return nil, ErrInvalidToken
	}
	if claims["aud"] != s.audience {
		return nil, ErrInvalidToken
	}

	// Extract claims
	refreshClaims := &RefreshTokenClaims{
		UserID:    getStringClaim(claims, "user_id"),
		SessionID: getStringClaim(claims, "session_id"),
		IssuedAt:  getInt64Claim(claims, "iat"),
		ExpiresAt: getInt64Claim(claims, "exp"),
		Issuer:    getStringClaim(claims, "iss"),
		Audience:  getStringClaim(claims, "aud"),
	}

	return refreshClaims, nil
}

// Helper functions to extract claims safely
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

func getInt64Claim(claims jwt.MapClaims, key string) int64 {
	if val, ok := claims[key].(float64); ok {
		return int64(val)
	}
	return 0
}
