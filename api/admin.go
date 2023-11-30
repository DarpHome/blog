package api

import (
	"runtime"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SetupAdmin(admin fiber.Router) error {
	admin.Post("/gc", func(c *fiber.Ctx) error {
		now := time.Now().UnixMilli()
		runtime.GC()
		elapsed := time.Now().UnixMilli() - now
		return c.Status(fiber.StatusOK).SendString(strconv.FormatUint(uint64(elapsed), 10))
	})
	return nil
}
