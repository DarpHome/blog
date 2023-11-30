package misc

import "strconv"

type User struct {
	Id             uint64
	Avatar         string
	Password       string
	TokenTimestamp int64
	Username       string
	Flags          int64
	Bio            string
}

func (u *User) Build() interface{} {
	r := map[string]interface{}{
		"id":       strconv.FormatUint(u.Id, 10),
		"username": u.Username,
		"flags":    u.Flags,
		"bio":      u.Bio,
	}
	if u.Avatar == "" {
		r["avatar"] = nil
	} else {
		r["avatar"] = u.Avatar
	}
	return r
}
