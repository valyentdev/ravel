package corroclient

import "log/slog"

func scanJSONNumber(number float64, d interface{}) error {
	switch dest := d.(type) {
	case *int:
		*dest = int(number)
		return nil
	case *int8:
		*dest = int8(number)
		return nil
	case *int16:
		*dest = int16(number)
		return nil
	case *int32:
		*dest = int32(number)
		return nil
	case *int64:
		*dest = int64(number)
		return nil
	case *uint:
		*dest = uint(number)
		return nil
	case *uint8:
		*dest = uint8(number)
		return nil
	case *uint16:
		*dest = uint16(number)
		return nil
	case *uint32:
		*dest = uint32(number)
		return nil
	case *uint64:
		*dest = uint64(number)
		return nil
	case *float32:
		*dest = float32(number)
		return nil
	case *float64:
		*dest = float64(number)
		return nil
	case *bool:
		*dest = number != 0
		return nil
	}

	return ErrScan
}

func scanJSONString(s string, dest any) error {
	if dest == nil {
		return ErrScan
	}

	switch d := dest.(type) {
	case *string:
		*d = s
		return nil
	case *[]byte:
		*d = []byte(s)
		return nil
	}

	slog.Error("scanJSONString", "dest", s)
	return ErrScan
}

func scanJSONBool(b any, dest interface{}) error {

	switch d := dest.(type) {
	case *bool:
		*d = b.(bool)
		return nil
	}
	return ErrScan
}
