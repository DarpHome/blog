package app

import (
	"github.com/DarpHome/blog/globals"
	"github.com/gofiber/fiber/v2"
)

func Setup(router fiber.Router) error {
	docs := router.Group("/docs")
	if err := SetupDocs(docs); err != nil {
		return err
	}
	router.Get("/auth/register", func(c *fiber.Ctx) error {
		return c.Render("auth/register", fiber.Map{"Title": globals.Config.AppName})
	})
	router.Get("/auth/login", func(c *fiber.Ctx) error {
		return c.Render("auth/login", fiber.Map{"Title": globals.Config.AppName})
	})
	router.Get("/profile", func(c *fiber.Ctx) error {
		return c.Render("profile", fiber.Map{"Title": globals.Config.AppName})
	})
	return nil
}
