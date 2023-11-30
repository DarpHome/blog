package utils

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/DarpHome/blog/globals"
	"github.com/DarpHome/blog/misc"
	"github.com/MCausc78/cgorithm"
	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

func CheckPassword(password string) *misc.RESTError {
	l := len(password)
	if l < 6 {
		return &misc.RESTError{
			Code:    40015,
			Extra:   6,
			Message: misc.MessageFor(40015, 6),
		}
	} else if l > 72 {
		return &misc.RESTError{
			Code:    40016,
			Extra:   72,
			Message: misc.MessageFor(40016, 72),
		}
	} else if cgorithm.All([]rune(password), func(_ int, c rune) bool {
		return c >= '0' && c <= '9'
	}) {
		return &misc.RESTError{
			Code:    40017,
			Message: misc.MessageFor(40017),
		}
	} else if !cgorithm.Any([]rune(password), func(c rune) bool {
		return c >= '0' && c <= '9'
	}) {
		return &misc.RESTError{
			Code:    40018,
			Message: misc.MessageFor(40018),
		}
	}
	return nil
}

func GenerateToken(id uint64, tokenTimestamp int64, password string) string {
	return fmt.Sprintf(
		"%s.%s.%x",
		base64.StdEncoding.EncodeToString([]byte(strconv.FormatUint(id, 10))),
		base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(tokenTimestamp, 32))),
		sha256.Sum256([]byte(fmt.Sprintf("'%s'&%d|%d<===>%s", globals.Config.SecretKey, id, tokenTimestamp, password))),
	)
}

func FetchUser(c *fiber.Ctx) *misc.User {
	h := c.Get("Authorization")
	if h == "" {
		return nil
	}
	parts := strings.Split(h, ".")
	if len(parts) != 3 {
		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil
	}
	id, err := strconv.ParseUint(string(decoded), 10, 64)
	if err != nil {
		return nil
	}
	r := globals.Db.QueryRow("SELECT id, avatar, password, token, username, flags, bio FROM users WHERE id = $1", id)
	var flags int64
	var avatar, bio sql.NullString
	var password string
	var tokenTimestamp uint64
	var username string
	if err := r.Scan(&id, &avatar, &password, &tokenTimestamp, &username, &flags, &bio); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			globals.Logger.Error(err)
		}
		return nil
	}
	if h != GenerateToken(id, int64(tokenTimestamp), password) {
		return nil
	}
	if !avatar.Valid {
		avatar.String = ""
	}
	if !bio.Valid {
		bio.String = ""
	}
	return &misc.User{
		Id:       id,
		Avatar:   avatar.String,
		Password: password,
		Username: username,
		Flags:    flags,
		Bio:      bio.String,
	}
}

func ResolveUsers(ids []uint64) (map[uint64]misc.User, int, error) {
	state := 0
	rows, err := globals.Db.Query(
		"SELECT id, avatar, password, token, username, flags, bio FROM users WHERE id = ANY($1)",
		pq.Array(ids),
	)
	if err != nil {
		return nil, state, err
	}
	defer rows.Close()
	res := map[uint64]misc.User{}
	for rows.Next() {
		state++
		var avatar, bio sql.NullString
		var flags int64
		var password string
		var id, tokenTimestamp uint64
		var username string
		if err := rows.Scan(&id, &avatar, &password, &tokenTimestamp, &username, &flags, &bio); err != nil {
			return nil, state, err
		}
		if !avatar.Valid {
			avatar.String = ""
		}
		if !bio.Valid {
			bio.String = ""
		}
		res[id] = misc.User{
			Id:       id,
			Avatar:   avatar.String,
			Password: password,
			Username: username,
			Flags:    flags,
			Bio:      bio.String,
		}
	}
	return res, 0, nil
}

func ResolveUser(id uint64) *misc.User {
	r := globals.Db.QueryRow("SELECT id, avatar, password, token, username, flags, bio FROM users WHERE id = $1", id)
	var flags int64
	var avatar, bio sql.NullString
	var password string
	var tokenTimestamp uint64
	var username string
	if err := r.Scan(&id, &avatar, &password, &tokenTimestamp, &username, &flags, &bio); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			globals.Logger.Error(err)
		}
		return nil
	}
	if !avatar.Valid {
		avatar.String = ""
	}
	if !bio.Valid {
		bio.String = ""
	}
	return &misc.User{
		Id:       id,
		Avatar:   avatar.String,
		Password: password,
		Username: username,
		Flags:    flags,
		Bio:      bio.String,
	}
}

func Hash(id uint64, password string) string {
	hash := sha256.Sum256([]byte(strconv.FormatUint(id, 10) + ":" + password))
	return hex.EncodeToString(hash[:])
}
