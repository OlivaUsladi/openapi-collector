package merge

import "testing"

func TestDeepEqual(t *testing.T) {
	a := map[string]any{"x": []any{1, "s", map[string]any{"k": true}}}
	b := map[string]any{"x": []any{1, "s", map[string]any{"k": true}}}
	if !deepEqual(any(a), any(b)) {
		t.Error("одинаковые структуры должны быть равны")
	}
	c := map[string]any{"x": []any{"s", 1, map[string]any{"k": true}}}
	if deepEqual(any(a), any(c)) {
		t.Error("порядок элементов массива значим")
	}
}
