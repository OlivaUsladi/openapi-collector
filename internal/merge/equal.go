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

// сравнивает два значения на полное равенство
func deepEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	mapA, aIsMap := a.(map[string]any)
	mapB, bIsMap := b.(map[string]any)

	if aIsMap && bIsMap {
		if len(mapA) != len(mapB) {
			return false
		}

		for key, aVal := range mapA {
			bVal, exists := mapB[key]
			if !exists {
				return false
			}

			if !deepEqual(aVal, bVal) {
				return false
			}
		}

		return true
	}

	sliceA, aIsSlice := a.([]any)
	sliceB, bIsSlice := b.([]any)

	if aIsSlice && bIsSlice {
		if len(sliceA) != len(sliceB) {
			return false
		}

		for idx := 0; idx < len(sliceA); idx++ {
			if !deepEqual(sliceA[idx], sliceB[idx]) {
				return false
			}
		}

		return true
	}

	return a == b
}
