package bsonutil

import "strings"

// CoerceToBool maps legacy BSON values (string, number, etc.) to bool.
func CoerceToBool(v interface{}) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		s := strings.TrimSpace(strings.ToLower(x))
		switch s {
		case "true", "1", "yes", "y", "si", "sí":
			return true
		default:
			return false
		}
	case float64:
		return x != 0
	case int32:
		return x != 0
	case int64:
		return x != 0
	case int:
		return x != 0
	default:
		return false
	}
}
