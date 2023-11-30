package api

import (
	"database/sql"
	"slices"
	"strconv"

	"github.com/DarpHome/blog/globals"
	"github.com/DarpHome/blog/misc"
	"github.com/DarpHome/blog/utils"
	"github.com/MCausc78/cgorithm"
	"github.com/gofiber/fiber/v2"
)

type Post struct {
	Id              uint64
	Author          uint64
	Tags            int
	Title           string
	EditedTimestamp int64
	Body            string
}

func SetupPosts(router fiber.Router) error {
	router.Post("/posts", func(c *fiber.Ctx) error {
		u := utils.FetchUser(c)
		if u == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(misc.RESTError{Code: 0, Message: "Unauthorized"})
		}
		var data map[string]interface{}
		if err := c.App().Config().JSONDecoder(c.Body(), &data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{Code: 0, Message: "invalid json"})
		}
		fields := map[string]misc.RESTError{}
		var tags int
		dtags, ok := data["tags"]
		if ok {
			if i, ok := dtags.(int); ok {
				tags = i & 0b1111111111111111
			} else {
				fields["tags"] = misc.RESTError{Code: 40012, Message: misc.MessageFor(40012, "int")}
			}
		}
		var title string
		dtitle, ok := data["title"]
		if !ok {
			fields["title"] = misc.RESTError{Code: 40011, Message: misc.MessageFor(40011)}
		} else if s, ok := dtitle.(string); !ok {
			fields["title"] = misc.RESTError{Code: 40012, Message: misc.MessageFor(40012, "string")}
		} else {
			l := len(s)
			if l < 20 {
				fields["title"] = misc.RESTError{Code: 40015, Message: misc.MessageFor(40015, 20)}
			} else if l > 60 {
				fields["title"] = misc.RESTError{Code: 40016, Message: misc.MessageFor(40016, 60)}
			}
			title = s
		}
		var body string
		dbody, ok := data["body"]
		if !ok {
			fields["body"] = misc.RESTError{Code: 40011, Message: misc.MessageFor(40011)}
		} else if s, ok := dbody.(string); !ok {
			fields["body"] = misc.RESTError{Code: 40012, Message: misc.MessageFor(40012, "string")}
		} else {
			l := len(s)
			if l < 100 {
				fields["body"] = misc.RESTError{Code: 40015, Message: misc.MessageFor(40015, 100)}
			} else if l > 32000 {
				fields["body"] = misc.RESTError{Code: 40016, Message: misc.MessageFor(40016, 32000)}
			}
			body = s
		}
		if len(fields) != 0 {
			return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{
				Code:    40010,
				Message: misc.MessageFor(40010),
				Fields:  fields,
			})
		}
		row := globals.Db.QueryRow(
			"INSERT INTO posts (id, author, tags, title, body) VALUES ($1, $2, $3, $4, $5) "+
				"RETURNING id, author, tags, title, body",
			globals.Snowflakes.Make(0).Value(),
			u.Id,
			tags,
			title,
			body,
		)
		if row == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
				Code:    50010,
				Message: "row is nil",
			})
		}
		if err := row.Err(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
				Code:    50013,
				Message: misc.MessageFor(50013, err.Error()),
			})
		}
		var id, author uint64
		if err := row.Scan(&id, &author, &tags, &title, &body); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
				Code:    50011,
				Message: misc.MessageFor(50011, err.Error()),
			})
		}
		return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
			"id":               strconv.FormatUint(id, 10),
			"author":           u.Build(),
			"tags":             tags,
			"title":            title,
			"edited_timestamp": nil,
			"body":             body,
		})
	})
	router.Get("/posts", func(c *fiber.Ctx) error {
		/* u := utils.FetchUser(c)
		if u == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(misc.RESTError{Code: 0, Message: "Unauthorized"})
		} */
		before := c.QueryInt("before", 0)
		after := c.QueryInt("after", (1<<63)-1)
		rows, err := globals.Db.Query(
			"SELECT id, author, tags, title, edited_timestamp, body FROM posts WHERE id >= $1 AND id <= $2 ORDER BY id >> 22 DESC LIMIT 20",
			before,
			after,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
				Code:    50013,
				Message: misc.MessageFor(50013, err.Error()),
			})
		}
		defer rows.Close()
		var posts = []Post{}
		for rows.Next() {
			var post Post
			var editedTimestamp sql.NullInt64
			if err := rows.Scan(&post.Id, &post.Author, &post.Tags, &post.Title, &editedTimestamp, &post.Body); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(misc.RESTError{
					Code:    50011,
					Message: misc.MessageFor(50011, err.Error()),
				})
			}
			if editedTimestamp.Valid {
				post.EditedTimestamp = editedTimestamp.Int64
			}
			posts = append(posts, post)
		}
		users := []uint64{}
		for _, post := range posts {
			if slices.Contains(users, post.Author) {
				continue
			}
			users = append(users, post.Author)
		}
		resolvedUsers, state, err := utils.ResolveUsers(users)
		if err != nil {
			if state == 0 {
				return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{
					Code:    50013,
					Message: misc.MessageFor(50013, err.Error()),
				})
			}
			return c.Status(fiber.StatusBadRequest).JSON(misc.RESTError{
				Code:    50011,
				Message: misc.MessageFor(50011, err.Error()),
			})
		}
		buildPost := func(post Post) interface{} {
			var author interface{}
			if ob, ok := resolvedUsers[post.Author]; ok {
				author = ob.Build()
			} else {
				authorId := strconv.FormatUint(post.Author, 10)
				author = map[string]interface{}{
					"id":     authorId,
					"avatar": nil,
					"bio":    nil,
					"flags":  1,
				}
			}
			r := map[string]interface{}{
				"id":     strconv.FormatUint(post.Id, 10),
				"author": author,
				"tags":   post.Tags,
				"title":  post.Title,
			}
			if post.EditedTimestamp == 0 {
				r["edited_timestamp"] = nil
			} else {
				r["edited_timestamp"] = post.EditedTimestamp
			}
			r["body"] = post.Body
			return r
		}
		/*return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
			"posts": cgorithm.Transform(posts, func(_ int, post Post) interface{} {
				return buildPost(post)
			}),
			"resolved": map[string]interface{}{
				"users": cgorithm.MTransform(resolvedUsers, func(k uint64, v misc.User) (string, interface{}) {
					return strconv.FormatUint(k, 10), v.Build()
				}),
			},
		})*/

		return c.Status(fiber.StatusOK).JSON(cgorithm.Transform(posts, func(_ int, post Post) interface{} {
			return buildPost(post)

		}))
	})
	return nil
}
