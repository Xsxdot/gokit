package security

import (
	"context"
	"strings"
	"time"

	errorc "github.com/xsxdot/gokit/err"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JwtAuth 泛型 JWT 认证器
// T 必须实现 jwt.Claims 接口（通常是 *Claims 结构体）
type JwtAuth[T jwt.Claims] struct {
	secret       []byte
	expireTime   time.Duration
	newClaims    func() T                  // Claims 构造函数，用于 ParseWithClaims
	contextKey   any                       // Context 存储键
	saveToLocals func(c *fiber.Ctx, claims T) // Locals 存储回调
	setExpiry    func(claims T, duration time.Duration) int64 // 设置过期时间回调
}

// NewJwtAuth 创建泛型 JWT 认证器
func NewJwtAuth[T jwt.Claims](
	secret []byte,
	expireTime time.Duration,
	newClaims func() T,
	contextKey any,
	saveToLocals func(c *fiber.Ctx, claims T),
	setExpiry func(claims T, duration time.Duration) int64,
) *JwtAuth[T] {
	if expireTime <= 0 {
		expireTime = 24 * time.Hour
	}
	return &JwtAuth[T]{
		secret:       secret,
		expireTime:   expireTime,
		newClaims:    newClaims,
		contextKey:   contextKey,
		saveToLocals: saveToLocals,
		setExpiry:    setExpiry,
	}
}

// CreateToken 创建 JWT Token
// 返回签名的 token 字符串、过期时间戳（Unix）、错误
func (a *JwtAuth[T]) CreateToken(claims T) (string, int64, error) {
	expiryTs := a.setExpiry(claims, a.expireTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString(a.secret)
	return signedString, expiryTs, err
}

// ParseToken 解析 JWT Token
func (a *JwtAuth[T]) ParseToken(tokenString string) (T, error) {
	token, err := jwt.ParseWithClaims(tokenString, a.newClaims(), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errorc.New("无效的签名方法", nil).NoAuth()
		}
		return a.secret, nil
	})
	if err != nil {
		return a.newClaims(), errorc.New("invalid token", err).NoAuth()
	}

	claims, ok := token.Claims.(T)
	if !ok || !token.Valid {
		return a.newClaims(), errorc.New("invalid token", nil).NoAuth()
	}
	return claims, nil
}

// ParseAuthHeader 从 Fiber Context 解析 Authorization Header 并返回 Claims
func (a *JwtAuth[T]) ParseAuthHeader(c *fiber.Ctx) (T, error) {
	auth := c.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return a.newClaims(), errorc.New("authorization header is required", nil).NoAuth()
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	return a.ParseToken(token)
}

// SaveToContext 将 Claims 保存到 Fiber Context
// 同时调用 saveToLocals 回调保存特定字段到 Locals
func (a *JwtAuth[T]) SaveToContext(c *fiber.Ctx, claims T) {
	if a.saveToLocals != nil {
		a.saveToLocals(c, claims)
	}
	userCtx := c.UserContext()
	userCtx = context.WithValue(userCtx, a.contextKey, claims)
	c.SetUserContext(userCtx)
}

// GetClaimsFromCtx 从 context.Context 获取 Claims
func (a *JwtAuth[T]) GetClaimsFromCtx(ctx context.Context) (T, error) {
	claims, ok := ctx.Value(a.contextKey).(T)
	if !ok {
		return a.newClaims(), errorc.New("claims not found or invalid", nil).NoAuth()
	}
	return claims, nil
}

