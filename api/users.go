package api

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/DarpHome/blog/globals"
	"github.com/DarpHome/blog/misc"
	"github.com/DarpHome/blog/utils"
	"github.com/gofiber/fiber/v2"
)

func SetupUsers(router fiber.Router) error {
	router.Get("/users/@me", func(c *fiber.Ctx) error {
		u := utils.FetchUser(c)
		if u == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(misc.RESTError{Code: 0, Message: "Unauthorized"})
		}
		return c.Status(fiber.StatusOK).JSON(u.Build())
	})
	router.Get("/users/:id", func(c *fiber.Ctx) error {
		u := utils.FetchUser(c)
		if u == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(misc.RESTError{Code: 0, Message: "Unauthorized"})
		}
		stringId := c.Params("id")
		id, err := strconv.ParseUint(stringId, 10, 64)
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		var t *misc.User
		if id == u.Id {
			t = u
		} else {
			row := globals.Db.QueryRow("SELECT id, avatar, username, flags, bio FROM users WHERE id = $1", t)
			if row == nil {
				return c.Status(fiber.StatusNotFound).JSON(misc.RESTError{Code: 40014, Message: misc.MessageFor(40014)})
			}
			var avatar sql.NullString
			var username string
			var flags int64
			var bio sql.NullString
			if err := row.Scan(&id, &avatar, &username, &flags, &bio); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return c.Status(fiber.StatusNotFound).JSON(misc.RESTError{Code: 40014, Message: misc.MessageFor(40014)})
				} else {
					return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
						Code:    50013,
						Message: misc.MessageFor(50013, err.Error()),
					})
				}
			}
			if !avatar.Valid {
				avatar.String = ""
			}
			if !bio.Valid {
				bio.String = ""
			}
			t = &misc.User{
				Id:       id,
				Avatar:   avatar.String,
				Username: username,
				Flags:    flags,
				Bio:      bio.String,
			}
		}
		return c.Status(fiber.StatusOK).JSON(t.Build())
	})
	router.Post("/auth/register", func(c *fiber.Ctx) error {
		var data map[string]interface{}
		if err := c.App().Config().JSONDecoder(c.Body(), &data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{Code: 0, Message: "invalid json"})
		}
		fields := map[string]misc.RESTError{}
		dusername, ok := data["username"]
		var username string
		if !ok {
			fields["username"] = misc.RESTError{
				Code:    40011,
				Message: misc.MessageFor(40011),
			}
		} else if s, ok := dusername.(string); ok {
			l := len(s)
			if l < 2 {
				fields["username"] = misc.RESTError{
					Code:    40015,
					Extra:   2,
					Message: misc.MessageFor(40015, 2),
				}
			} else if l > 32 {
				fields["username"] = misc.RESTError{
					Code:    40016,
					Extra:   32,
					Message: misc.MessageFor(40016, 32),
				}
			}
			username = s
		} else {
			fields["username"] = misc.RESTError{
				Code:    40012,
				Extra:   "string",
				Message: misc.MessageFor(40012, "string"),
			}
		}
		dpassword, ok := data["password"]
		var password string
		if !ok {
			fields["password"] = misc.RESTError{
				Code:    40011,
				Message: misc.MessageFor(40011),
			}
		} else if s, ok := dpassword.(string); ok {
			if err := utils.CheckPassword(s); err != nil {
				fields["password"] = *err
			} else {
				password = s
			}
		} else {
			fields["password"] = misc.RESTError{
				Code:    40012,
				Extra:   "string",
				Message: misc.MessageFor(40012, "string"),
			}
		}
		if len(fields) != 0 {
			return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{
				Code:    40010,
				Message: misc.MessageFor(40010),
				Fields:  fields,
			})
		}
		var claimed bool
		claimedRow := globals.Db.QueryRow("SELECT COUNT(*) <> 0 FROM users WHERE username = $1", username)
		if claimedRow == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
				Code:    50010,
				Message: misc.MessageFor(50010),
			})
		}
		if err := claimedRow.Err(); err != nil {
			message := err.Error()
			return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
				Code:    50013,
				Extra:   message,
				Message: misc.MessageFor(50013, message),
			})
		}
		if err := claimedRow.Scan(&claimed); err != nil {
			message := err.Error()
			return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
				Code:    50011,
				Extra:   message,
				Message: misc.MessageFor(50011, message),
			})
		}
		if claimed {
			return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{
				Code:    40013,
				Message: misc.MessageFor(40013),
			})
		}
		id := globals.Snowflakes.Make(0).Value()
		hash := utils.Hash(id, password)
		row := globals.Db.QueryRow(
			"INSERT INTO users(id, password, token, username) VALUES ($1, $2, $3, $4) "+
				"RETURNING id, avatar, password, token, username, flags",
			id,
			hash,
			time.Now().Unix(),
			username,
		)
		if row == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
				Code:    50012,
				Message: misc.MessageFor(50012),
			})
		}
		var resultId, resultTokenTimestamp, resultFlags int64
		var avatar sql.NullString
		if err := row.Scan(&resultId, &avatar, &password, &resultTokenTimestamp, &username, &resultFlags); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
				Code:    50011,
				Message: misc.MessageFor(50011, err.Error()),
			})
		}
		return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
			"token": utils.GenerateToken(uint64(resultId), resultTokenTimestamp, password),
			"user": map[string]interface{}{
				"id": strconv.FormatUint(uint64(resultId), 10),
				"avatar": func(s sql.NullString) interface{} {
					if s.Valid {
						return s.String
					}
					return nil
				}(avatar),
				"bio":      nil,
				"flags":    resultFlags,
				"username": username,
			},
		})
	})
	router.Post("/auth/login", func(c *fiber.Ctx) error {
		var data map[string]interface{}
		if err := c.App().Config().JSONDecoder(c.Body(), &data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{Code: 0, Message: "invalid json"})
		}
		fields := map[string]misc.RESTError{}
		dusername, ok := data["username"]
		var username string
		if !ok {
			fields["username"] = misc.RESTError{
				Code:    40011,
				Message: misc.MessageFor(40011),
			}
		} else if s, ok := dusername.(string); ok {
			/* l := len(s)
			if l < 2 {
				fields["username"] = misc.RESTError{
					Code:    40015,
					Message: misc.MessageFor(40015, 2),
				}
			} else if l > 32 {
				fields["username"] = misc.RESTError{
					Code:    40016,
					Message: misc.MessageFor(40016, 32),
				}
			} */
			username = s
		} else {
			fields["username"] = misc.RESTError{
				Code:    40012,
				Message: misc.MessageFor(40012, "string"),
			}
		}
		dpassword, ok := data["password"]
		var password string
		if !ok {
			fields["password"] = misc.RESTError{
				Code:    40011,
				Message: misc.MessageFor(40011),
			}
		} else if s, ok := dpassword.(string); ok {
			/* if err := utils.CheckPassword(s); err != nil {
				fields["password"] = *err
			} else {
				password = s
			} */
			password = s
		} else {
			fields["password"] = misc.RESTError{
				Code:    40012,
				Message: misc.MessageFor(40012, "string"),
			}
		}
		if len(fields) != 0 {
			return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{
				Code:    40010,
				Message: misc.MessageFor(40010),
				Fields:  fields,
			})
		}
		r := globals.Db.QueryRow(
			"SELECT id, password, token FROM users WHERE username = $1",
			username,
		)
		if r == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
				Code:    50010,
				Message: misc.MessageFor(50010),
			})
		}
		if err := r.Err(); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
					Code:    50013,
					Message: misc.MessageFor(50013, err.Error()),
				})
			}
			return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{
				Code:    40014,
				Message: misc.MessageFor(40014),
			})
		}
		var u misc.User
		if err := r.Scan(&u.Id, &u.Password, &u.TokenTimestamp); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
					Code:    50011,
					Message: misc.MessageFor(50011, err.Error()),
				})
			}
			return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{
				Code:    40014,
				Message: misc.MessageFor(40014, err.Error()),
			})
		}
		hash := utils.Hash(u.Id, password)
		if hash != u.Password {
			return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{
				Code:    40019,
				Message: misc.MessageFor(40019),
			})
		}
		return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
			"token": utils.GenerateToken(u.Id, u.TokenTimestamp, u.Password),
		})
	})
	return nil
}
