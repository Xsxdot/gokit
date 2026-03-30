package security

import (
	"context"
	"strings"
	"time"

	errorc "gokit/err"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	UserKey        = "user"
	UserSuperAdmin = "ROLE_USER_SUPER_ADMIN"
)

type UserClaims struct {
	jwt.RegisteredClaims
	ID          int64    `json:"id"`
	Username    string   `json:"username,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// UserAuth 用户认证器（使用泛型实现）
type UserAuth struct {
	jwtAuth *JwtAuth[*UserClaims]
}

// NewUserAuth 创建用户认证器
func NewUserAuth(secret []byte, expireTime time.Duration) *UserAuth {
	jwtAuth := NewJwtAuth(
		secret,
		expireTime,
		func() *UserClaims { return &UserClaims{} },
		UserKey,
		func(c *fiber.Ctx, claims *UserClaims) {
			c.Locals("user_id", claims.ID)
		},
		func(claims *UserClaims, duration time.Duration) int64 {
			now := time.Now()
			claims.ExpiresAt = jwt.NewNumericDate(now.Add(duration))
			claims.IssuedAt = jwt.NewNumericDate(now)
			return claims.ExpiresAt.Unix()
		},
	)
	return &UserAuth{jwtAuth: jwtAuth}
}

// CreateSimpleToken 创建用户 token（简化版）
func (a *UserAuth) CreateSimpleToken(userID int64, userName string) (string, error) {
	claims := &UserClaims{
		ID:       userID,
		Username: userName,
	}
	token, _, err := a.jwtAuth.CreateToken(claims)
	return token, err
}

// CreateToken 创建用户 token
func (a *UserAuth) CreateToken(claims *UserClaims) (string, int64, error) {
	return a.jwtAuth.CreateToken(claims)
}

// NoAuthRequired 无需校验权限，直接放行
func (a *UserAuth) NoAuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

// OptionalAuth 可选校验，有 token 则验证并保存 ID
func (a *UserAuth) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth != "" && strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")
			claims, err := a.jwtAuth.ParseToken(token)
			if err == nil {
				a.jwtAuth.SaveToContext(c, claims)
			}
		}
		return c.Next()
	}
}

// RequireAuth 必须通过校验，并保存 ID
func (a *UserAuth) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := a.jwtAuth.ParseAuthHeader(c)
		if err != nil {
			return err
		}
		a.jwtAuth.SaveToContext(c, claims)
		return c.Next()
	}
}

// RequirePermission 要求特定权限
func (a *UserAuth) RequirePermission(permissionCode string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := a.jwtAuth.ParseAuthHeader(c)
		if err != nil {
			return err
		}

		// 保存用户信息到上下文
		a.jwtAuth.SaveToContext(c, claims)

		hasPermission := false
		for _, perm := range claims.Permissions {
			if perm == permissionCode {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return errorc.New("permission denied", nil).Forbidden()
		}

		return c.Next()
	}
}

// 便捷函数（供外部直接调用）

// GetUserID 从上下文中获取用户 ID
func GetUserID(c *fiber.Ctx) (int64, error) {
	if c == nil {
		return 0, errorc.New("fiber context is nil", nil).WithCode(errorc.ErrorCodeInternal)
	}
	id, ok := c.Locals("user_id").(int64)
	if !ok || id == 0 {
		return 0, errorc.New("user id not found or invalid", nil).NoAuth()
	}
	return id, nil
}

// GetUserRoles 从上下文中获取用户角色
func GetUserRoles(c *fiber.Ctx) ([]string, error) {
	if c == nil {
		return nil, errorc.New("fiber context is nil", nil).WithCode(errorc.ErrorCodeInternal)
	}
	ctx := c.UserContext()
	claims, ok := ctx.Value(UserKey).(*UserClaims)
	if !ok {
		return nil, errorc.New("user claims not found or invalid", nil).NoAuth()
	}
	return claims.Permissions, nil
}

// IsUserSuper 检查用户是否是超级管理员
func IsUserSuper(c *fiber.Ctx) bool {
	if c == nil {
		return false
	}
	isSuper, ok := c.Locals("is_super").(bool)
	return ok && isSuper
}

func GetUserClaimsByCtx(ctx context.Context) (*UserClaims, error) {
	claims, ok := ctx.Value(UserKey).(*UserClaims)
	if !ok {
		return nil, errorc.New("user claims not found or invalid", nil).NoAuth()
	}
	return claims, nil
}

// ParseToken 解析用户令牌（供外部使用）
func (a *UserAuth) ParseToken(token string) (*UserClaims, error) {
	return a.jwtAuth.ParseToken(token)
}