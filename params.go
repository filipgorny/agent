package agent

// paramString returns a string parameter and whether it was present as a string.
func paramString(params map[string]any, key string) (string, bool) {
	v, ok := params[key].(string)

	return v, ok
}

// paramInt returns an int parameter, tolerating JSON numbers (float64).
func paramInt(params map[string]any, key string) (int, bool) {
	switch v := params[key].(type) {

	case float64:
		return int(v), true

	case int:
		return v, true

	default:
		return 0, false
	}
}

// paramBool returns a bool parameter (false when absent).
func paramBool(params map[string]any, key string) bool {
	v, _ := params[key].(bool)

	return v
}
