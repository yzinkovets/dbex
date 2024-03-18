package utils

import (
	"database/sql"
	"fmt"
)

func GetStringFromAny(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int, uint32:
		return fmt.Sprintf("%d", v)
	case bool:
		if v {
			return "true"
		} else {
			return "false"
		}
	case sql.NullString:
		if v.Valid {
			return v.String
		} else {
			return "null"
		}
	default:
		return "<unsupported type>"
	}
}
