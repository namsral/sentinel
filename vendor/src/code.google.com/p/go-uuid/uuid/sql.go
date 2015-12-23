package uuid

import (
	"database/sql/driver"
	"errors"
)

var ErrIncompatibleType = errors.New("incompatible type for UUID")

// Value implements the driver.Valuer interface.
func (u UUID) Value() (driver.Value, error) {
	return u.String(), nil
}

// Scan implements the driver.Scanner interface.
func (u *UUID) Scan(src interface{}) error {
	var source string
	switch src.(type) {
	case string:
		source = src.(string)
	case []byte:
		source = string(src.([]byte))
	default:
		return ErrIncompatibleType
	}
	*u = Parse(source)
	return nil
}
