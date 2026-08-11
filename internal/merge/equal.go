package merge

// копия разобранного YAML фрагмента
func copyValue(v any) any {
	m, ok := v.(map[string]any)
	if ok {
		out := make(map[string]any)
		for key, value := range m {
			out[key] = copyValue(value)
		}
		return out
	}

	s, ok := v.([]any)
	if ok {
		out := make([]any, len(s))
		for i, value := range s {
			out[i] = copyValue(value)
		}
		return out
	}

	return v
}
