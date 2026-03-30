package utils

import (
	"context"
	"gokit/consts"

	"github.com/gofiber/fiber/v2"
	uuid "github.com/google/uuid"
)

func Context(c *fiber.Ctx) context.Context {

	ctx := c.UserContext()
	if ctx.Value(consts.TraceKey) == nil {
		return context.WithValue(context.Background(), consts.TraceKey, uuid.New().String())
	}
	return ctx
}
