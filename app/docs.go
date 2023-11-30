package app

import (
	"github.com/DarpHome/blog/globals"
	"github.com/gofiber/fiber/v2"
)

func SetupDocs(router fiber.Router) error {
	router.Get("/", func(c *fiber.Ctx) error {
		return c.Render("docs/index", fiber.Map{"Title": globals.Config.AppName, "SnowflakesEpoch": globals.SnowflakesEpoch})
	})
	router.Get("/auth", func(c *fiber.Ctx) error {
		return c.Render("docs/auth", fiber.Map{"Title": globals.Config.AppName})
	})
	router.Get("/users", func(c *fiber.Ctx) error {
		return c.Render("docs/users", fiber.Map{"Title": globals.Config.AppName})
	})
	return nil
}
