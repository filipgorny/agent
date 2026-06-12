package core

// ParamString returns a string parameter and whether it was present as a string.
func ParamString(params map[string]any, key string) (string, bool) {
	v, ok := params[key].(string)

	return v, ok
}

// ParamInt returns an int parameter, tolerating JSON numbers (float64).
func ParamInt(params map[string]any, key string) (int, bool) {
	switch v := params[key].(type) {

	case float64:
		return int(v), true

	case int:
		return v, true

	default:
		return 0, false
	}
}

// ParamBool returns a bool parameter (false when absent).
func ParamBool(params map[string]any, key string) bool {
	v, _ := params[key].(bool)

	return v
}
