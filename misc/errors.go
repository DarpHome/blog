package misc

import (
	"fmt"
)

type RESTError struct {
	Code    int                  `json:"code"`
	Extra   interface{}          `json:"extra,omitempty"`
	Message string               `json:"message"`
	Fields  map[string]RESTError `json:"fields,omitempty"`
}

func MessageFor(code int, data ...interface{}) string {
	switch code {
	case 40009:
		return "Some fields are incorrect"
	case 40010:
		return "Some fields are required"
	case 40011:
		return "This field is required."
	case 40012:
		return fmt.Sprintf("Expected %s", data[0].(string))
	case 40013:
		return "User with such username already exists."
	case 40014:
		return "User not found"
	case 40015:
		return fmt.Sprintf("Field may not be shorter than %d", data[0].(int))
	case 40016:
		return fmt.Sprintf("Field may not be longer than %d", data[0].(int))
	case 40017:
		return "Password may not contain only digits"
	case 40018:
		return "Password should contain at least one digit"
	case 40019:
		return "Incorrect password"
	case 50010:
		return "row is nil"
	case 50011:
		return "unable scan record: " + data[0].(string)
	case 50012:
		return "unable insert record to database: " + data[0].(string)
	case 50013:
		return "unable download record from database: " + data[0].(string)
	default:
		return ""
	}
}
