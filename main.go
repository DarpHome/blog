package main

//"github.com/gofiber/fiber/v2"

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/DarpHome/blog/api"
	"github.com/DarpHome/blog/app"
	"github.com/DarpHome/blog/globals"
	"github.com/DarpHome/blog/misc"
	"github.com/DarpHome/blog/utils"
	goflaker "github.com/MCausc78/goflaker"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type AppConfig struct {
	SecretKey  string `json:"secret_key"`
	PgHost     string `json:"pghost"`
	PgPort     int    `json:"pgport"`
	PgUser     string `json:"pguser"`
	PgPassword string `json:"pgpassword"`
	PgDbname   string `json:"pgdbname"`
}

func setupApi(router fiber.Router) error {
	router.Get("/time", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString(strconv.FormatUint(uint64(time.Now().UnixMilli()), 10))
	})
	if err := api.SetupUsers(router); err != nil {
		return err
	}
	if err := api.SetupPosts(router); err != nil {
		return err
	}
	admin := router.Group("/admin")
	adminMiddleware := func(c *fiber.Ctx) error {
		u := utils.FetchUser(c)
		if u == nil || (u.Flags&0x00000002) == 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(misc.RESTError{Code: 0, Message: "Unauthorized"})
		}
		return c.Next()
	}
	admin.Use(adminMiddleware)
	return api.SetupAdmin(admin)
}

func setupAssets(assets fiber.Router) error {
	assets.Static("/assets", "assets/", fiber.Static{
		MaxAge: 600,
	})
	assets.Static("/public", "public/", fiber.Static{
		MaxAge: 60,
	})
	return nil
}

func setupApp(router fiber.Router) error {
	return app.Setup(router)
}

func main() {
	logger := &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
	globals.Logger = logger
	file, err := os.Open("config.json")
	if err != nil {
		globals.Logger.Fatal(err)
	}
	defer file.Close()
	config, err := io.ReadAll(file)
	if err != nil {
		globals.Logger.Fatal(err)
	}
	if err := json.Unmarshal(config, &globals.Config); err != nil {
		globals.Logger.Fatal(err)
	}
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		globals.Config.PgHost,
		globals.Config.PgPort,
		globals.Config.PgUser,
		globals.Config.PgPassword,
		globals.Config.PgDbname,
	))
	if err != nil {
		globals.Logger.Fatal(err)
	}
	defer db.Close()
	globals.Db = db
	if err := globals.SetupDb(db); err != nil {
		globals.Logger.Fatal(err)
	}
	snowflakesBuilder := goflaker.NewBuilder(globals.SnowflakesEpoch)
	globals.Snowflakes = snowflakesBuilder.DefaultGenerator(0)
	app := fiber.New(fiber.Config{
		AppName:           "blog",
		EnablePrintRoutes: true,
		ServerHeader: fmt.Sprintf(
			"Go %s running on %s %s, fiber %s",
			runtime.Version(),
			runtime.GOARCH,
			runtime.GOOS,
			fiber.Version,
		),
		Views: html.New("./views/", ".html"),
	})
	api := app.Group("/api/v1")
	if err := setupApi(api); err != nil {
		globals.Logger.Fatal(err)
	}
	assets := app.Group("/assets")
	if err := setupAssets(assets); err != nil {
		globals.Logger.Fatal(err)
	}
	if err := setupApp(app); err != nil {
		globals.Logger.Fatal(err)
	}
	if err := app.Listen(":8080"); err != nil {
		globals.Logger.Fatal(err)
	}
}
