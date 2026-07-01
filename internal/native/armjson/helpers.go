package armjson

// AsMap returns v as a JSON-object-like map. Non-map and nil values become an
// empty map so callers can safely inspect nested ARM response objects.
func AsMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok && m != nil {
		return m
	}
	return map[string]interface{}{}
}

// Int64 returns a best-effort integer value for numeric ARM response values.
// JSON-decoded numbers commonly arrive as float64; tests and SDK-shaped maps may
// use Go integer types.
func Int64(v interface{}) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case uint:
		return int64(n)
	case uint8:
		return int64(n)
	case uint16:
		return int64(n)
	case uint32:
		return int64(n)
	case uint64:
		return int64(n)
	case float32:
		return int64(n)
	case float64:
		return int64(n)
	default:
		return 0
	}
}
