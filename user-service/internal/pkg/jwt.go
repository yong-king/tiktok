package pkg

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"user-service/internal/pkg/tracing"
)

type CustomClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey []byte
	issuer    string
	expire    time.Duration
}

var (
	ErrAccessTokenExpired = errors.New("access token expired")
	ErrInvalidToken       = errors.New("invalid token")
)

func NewJWTManager(secretKey string, issuer string, expireSeconds int64) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		expire:    time.Duration(expireSeconds) * time.Second,
	}
}

func (j *JWTManager) Issuer() string {
	return j.issuer
}

func (j *JWTManager) SecretKey() []byte {
	return j.secretKey
}

// GenerateToken 生成 token
func (j *JWTManager) CreateToken(ctx context.Context, userID int64) (accessToken, refreshToken string, err error) {
	// 生成 Access Token（15 分钟有效）
	accessClaims := &CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    j.issuer,
		},
	}
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(j.secretKey)
	if err != nil {
		return "", "", err
	}

	// 生成 Refresh Token（7 天有效）
	refreshClaims := &CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			Issuer:    j.issuer,
		},
	}
	refresh, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(j.secretKey)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

// ParseToken 解析 token
func (j *JWTManager) ParseToken(ctx context.Context, tokenString string) (*CustomClaims, error) {

	ctx, span := tracing.StartSpan(ctx, "JWTManager.ParseToken")
	defer span.End()

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return j.secretKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrAccessTokenExpired
		}
		return nil, ErrInvalidToken
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// RefreshAccessToken 刷新token
func (j *JWTManager) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := j.ParseToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	accessClaims := &CustomClaims{
		UserID: claims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			Issuer:    j.issuer,
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(j.secretKey)
}

func (j *JWTManager) VerifyAndRefreshTokens(ctx context.Context, accessToken, refreshToken string) (string, error) {

	ctx, span := tracing.StartSpan(ctx, "JWTManager.VerifyAndRefreshTokens")
	defer span.End()

	// 1. 先校验 Access Token
	_, err := j.ParseToken(ctx, accessToken)
	if err == nil {
		// Access Token 有效，直接返回它
		return accessToken, nil
	}
	if errors.Is(err, ErrAccessTokenExpired) {
		// 2. Access Token 过期，尝试刷新
		newAccessToken, err := j.RefreshAccessToken(ctx, refreshToken)
		if err != nil {
			// Refresh Token 过期或无效，要求重新登录
			return "", errors.New("refresh token expired or invalid, please login again")
		}
		// 刷新成功，返回新 Access Token
		return newAccessToken, nil
	}
	// Access Token 其他错误，直接返回错误
	return "", err
}
