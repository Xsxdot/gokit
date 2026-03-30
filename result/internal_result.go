package result

import (
	"github.com/xsxdot/gokit/utils"

	errorc "github.com/xsxdot/gokit/err"

	"github.com/gofiber/fiber/v2"
)

func InternalOK(c *fiber.Ctx, v interface{}) error {
	return c.Status(200).JSON(v)
}

func InternalBadRequestNormal(c *fiber.Ctx, message string, err error) error {
	return errorc.New(message, err).WithTraceID(utils.Context(c))
}

func InternalBadRequest(c *fiber.Ctx, err error) error {
	return err
}

func InternalOnce(c *fiber.Ctx, v interface{}, err error) error {
	if err == nil {
		return InternalOK(c, v)
	} else {
		return InternalBadRequest(c, err)
	}
}
