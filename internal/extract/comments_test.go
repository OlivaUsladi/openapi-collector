package extract

import (
	"strings"
	"testing"
)

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

func TestFindMarkedComments2(t *testing.T) {
	src := `package p


/* @openapi
paths:
  /tasks:
    get:
      operationId: listTasks
*/
func F() {}
`
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(frags) != 1 {
		t.Fatalf("ожидался 1 фрагмент, получено %d", len(frags))
	}
	if strings.Join(frags[0].Sections, ",") != "paths" {
		t.Errorf("ожидался раздел paths, получено %v", frags[0].Sections)
	}
	if !strings.Contains(frags[0].Raw, "operationId: listTasks") {
		t.Errorf("в Raw ожидалась подстрока operationId: listTasks, Raw:\n%s", frags[0].Raw)
	}
}

func TestFindMarkedComments3(t *testing.T) {
	src := `package p
/*
@openapi
components:
  schemas:
    Task:
      type: object
*/
func F() {}
`
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(frags) != 1 {
		t.Fatalf("ожидался 1 фрагмент, получено %d", len(frags))
	}
	if strings.Join(frags[0].Sections, ",") != "components" {
		t.Errorf("ожидался раздел components, получено %v", frags[0].Sections)
	}
	if !strings.Contains(frags[0].Raw, "Task:") {
		t.Errorf("в Raw ожидалась подстрока Task:, Raw:\n%s", frags[0].Raw)
	}
}

func TestFindMarkedComments4(t *testing.T) {
	src := `package p


/*
 * @openapi
 * components:
 *   schemas:
 *     Error:
 *       type: object
 */
func F() {}
`
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(frags) != 1 {
		t.Fatalf("ожидался 1 фрагмент, получено %d", len(frags))
	}
	if strings.Join(frags[0].Sections, ",") != "components" {
		t.Errorf("ожидался раздел components, получено %v", frags[0].Sections)
	}
	if !strings.Contains(frags[0].Raw, "Error:") {
		t.Errorf("в Raw ожидалась подстрока Error:, Raw:\n%s", frags[0].Raw)
	}
}

func TestFindMarkedComments5(t *testing.T) {
	src := `package p


// @openapi
// components:
//   schemas:
//     Error:
//       type: object
func F() {}
`
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(frags) != 1 {
		t.Fatalf("ожидался 1 фрагмент, получено %d", len(frags))
	}
	if strings.Join(frags[0].Sections, ",") != "components" {
		t.Errorf("ожидался раздел components, получено %v", frags[0].Sections)
	}
	if !strings.Contains(frags[0].Raw, "Error:") {
		t.Errorf("в Raw ожидалась подстрока Error:, Raw:\n%s", frags[0].Raw)
	}
}

func TestFindMarkedComment5(t *testing.T) {
	src := `package p


// @openapi
// paths:
//   /a:
//     get:
//       operationId: a
func A() {}


// @openapi
// paths:
//   /b:
//     get:
//       operationId: b
func B() {}
`
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(frags) != 2 {
		t.Fatalf("ожидалось 2 фрагмента, получено %d", len(frags))
	}
	if strings.Join(frags[0].Sections, ",") != "paths" {
		t.Errorf("ожидался раздел components, получено %v", frags[0].Sections)
	}
	if !strings.Contains(frags[0].Raw, "operationId: a") {
		t.Errorf("в Raw ожидалась подстрока operationId: a, Raw:\n%s", frags[0].Raw)
	}

	if strings.Join(frags[1].Sections, ",") != "paths" {
		t.Errorf("ожидался раздел components, получено %v", frags[0].Sections)
	}
	if !strings.Contains(frags[1].Raw, "operationId: b") {
		t.Errorf("в Raw ожидалась подстрока operationId: b, Raw:\n%s", frags[0].Raw)
	}
}

func TestFindMarkedComments6(t *testing.T) {
	src := `package p
const example = "@openapi"
`
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(frags) != 0 {
		t.Fatalf("ожидалось 0 фрагментов, получено %d", len(frags))
	}
}

func TestFindMarkedComments7(t *testing.T) {
	src := `package p

// про @openapi будет ниже в коде
func F() {}
`
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(frags) != 0 {
		t.Fatalf("ожидалось 0 фрагментов, получено %d", len(frags))
	}
}

func TestFindMarkedComments8(t *testing.T) {
	src := `package p
// @openapi
func F() {}
`
	_, err := FindApiComment("test.go", "test.go", []byte(src))
	if err == nil {
		t.Fatal("ожидалась ошибка пустой @openapi фрагмент")
	}
	if !strings.Contains(err.Error(), "пустой @openapi, нет YAML после маркера") {
		t.Fatalf("ожидалась ошибка с текстом пустой @openapi, нет YAML после маркера, получено: %v", err)
	}
}

func TestFindMarkedComments9(t *testing.T) {
	src := `package p

// @openapi
// - just
// - a text
func F() {}
`
	_, err := FindApiComment("test.go", "test.go", []byte(src))
	if err == nil {
		t.Fatal("ожидалась ошибка @openapi фрагемент не map")
	}
	if !strings.Contains(err.Error(), "@openapi фрагемент не map") {
		t.Fatalf("ожидалась ошибка с текстом @openapi фрагемент не map, получено: %v", err)
	}
}

func TestFindMarkedComments10(t *testing.T) {
	src := `package p

// обычный комментарий
func F() {}
`
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	if len(frags) != 0 {
		t.Fatalf("ожидалось 0 фрагментов, получено %d", len(frags))
	}
}

func TestFindMarkedComments11(t *testing.T) {
	src := `package p
// @openapi
// paths:
//   /x:
//     get:
//       operationId: x
func F() {}
`
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(frags) != 1 {
		t.Fatalf("ожидался 1 фрагмент, получено %d", len(frags))
	}
	if frags[0].Origin.Line != 2 || frags[0].Origin.Column != 1 {
		t.Errorf("ожидалась позиция 2:1, получено %d:%d", frags[0].Origin.Line, frags[0].Origin.Column)
	}
}

func TestFindMarkedComments12(t *testing.T) {
	src := "package p\n\n// @openapi\n//\tpaths: {}\nfunc F() {}\n"
	_, err := FindApiComment("test.go", "test.go", []byte(src))
	if err == nil {
		t.Fatal("ожидалась ошибка табуляции")
	}
	if !strings.Contains(err.Error(), "ошибка табуляции") {
		t.Fatalf("ожидалась ошибка с текстом 'ошибка табуляции', получено: %v", err)
	}
}

func TestFindMarkedComments13(t *testing.T) {
	src := "package p\r\n\r\n// @openapi\r\n// paths:\r\n//   /x:\r\n//     get:\r\n//       operationId: x\r\nfunc F() {}\r\n"
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(frags) != 1 {
		t.Fatalf("CRLF: ожидался 1 фрагмент, получено %d", len(frags))
	}
}

func TestFindMarkedComments14(t *testing.T) {
	src := "\xEF\xBB\xBFpackage p\n\n// @openapi\n// paths:\n//   /x:\n//     get:\n//       operationId: x\nfunc F() {}\n"
	frags, err := FindApiComment("test.go", "test.go", []byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if len(frags) != 1 {
		t.Fatalf("BOM: ожидался 1 фрагмент, получено %d", len(frags))
	}
}

func TestFindMarkedComments15(t *testing.T) {
	src := `package p
// @openapi
// просто комментарий
func F() {}
`
	_, err := FindApiComment("test.go", "test.go", []byte(src))
	if err == nil {
		t.Fatal("ожидалась ошибка пустого YAML-фрагмента")
	}
}
