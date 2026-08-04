package extract

import "testing"

func TestFindMarkedComments(t *testing.T) {
	src := `package p
/* @openapi
paths:
  /tasks:
    get:
      operationId: listTasks
*/
func pupupu() {}
`
	expextedCount := 1
	expectedSection := "paths"
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(frags) != expextedCount {
		t.Fatalf("ожидалось %d фрагментов, получено %d", expextedCount, len(frags))
	}
	for _, s := range frags[0].Sections {
		if s != expectedSection {
			t.Errorf("ожидалась секция %q, получены %v", expectedSection, frags[0].Sections)
		}
	}

}
