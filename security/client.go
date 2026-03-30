package security

import (
	"context"
	"time"

	errorc "github.com/xsxdot/gokit/err"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const ClientKey = "client"

type ClientClaims struct {
	jwt.RegisteredClaims
	ClientID  int64  `json:"clientId"`
	ClientKey string `json:"clientKey"`
	Status    int8   `json:"status"`
}

// ClientAuth Client 认证器（使用泛型实现）
type ClientAuth struct {
	jwtAuth *JwtAuth[*ClientClaims]
}

// NewClientAuth 创建 Client 认证器
// Client 认证器只用于解析 token，不需要 expireTime
func NewClientAuth(secret []byte) *ClientAuth {
	jwtAuth := NewJwtAuth(
		secret,
		0, // Client 不创建 token，expireTime 无意义
		func() *ClientClaims { return &ClientClaims{} },
		ClientKey,
		func(c *fiber.Ctx, claims *ClientClaims) {
			c.Locals("client_id", claims.ClientID)
			c.Locals("client_key", claims.ClientKey)
			c.Locals("client_status", claims.Status)
		},
		func(claims *ClientClaims, duration time.Duration) int64 {
			// Client 不创建 token，此回调不会被调用
			return 0
		},
	)
	return &ClientAuth{jwtAuth: jwtAuth}
}

// ParseToken 解析 Client Token
func (a *ClientAuth) ParseToken(tokenString string) (*ClientClaims, error) {
	return a.jwtAuth.ParseToken(tokenString)
}

// SaveClientToContext 将 Client Claims 保存到 Context
func (a *ClientAuth) SaveClientToContext(c *fiber.Ctx, claims *ClientClaims) {
	a.jwtAuth.SaveToContext(c, claims)
}

// RequireClientAuth 必须通过校验，并保存 client 信息
func (a *ClientAuth) RequireClientAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := a.jwtAuth.ParseAuthHeader(c)
		if err != nil {
			return err
		}
		a.jwtAuth.SaveToContext(c, claims)
		return c.Next()
	}
}

// 便捷函数（供外部直接调用）

func GetClientID(c *fiber.Ctx) (int64, error) {
	if c == nil {
		return 0, errorc.New("fiber context is nil", nil).WithCode(errorc.ErrorCodeInternal)
	}
	id, ok := c.Locals("client_id").(int64)
	if !ok || id == 0 {
		return 0, errorc.New("client id not found or invalid", nil).NoAuth()
	}
	return id, nil
}

func GetClientKey(c *fiber.Ctx) (string, error) {
	if c == nil {
		return "", errorc.New("fiber context is nil", nil).WithCode(errorc.ErrorCodeInternal)
	}
	key, ok := c.Locals("client_key").(string)
	if !ok || key == "" {
		return "", errorc.New("client key not found or invalid", nil).NoAuth()
	}
	return key, nil
}

func GetClientClaimsByCtx(ctx context.Context) (*ClientClaims, error) {
	claims, ok := ctx.Value(ClientKey).(*ClientClaims)
	if !ok {
		return nil, errorc.New("client claims not found or invalid", nil).NoAuth()
	}
	return claims, nil
}