package extract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJoinOpenApi1(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	src := `package p

// @openapi
// paths:
//   /a:
//     get:
//       operationId: a
func F() {}
`
	err := os.WriteFile(path, []byte(src), 0644)
	if err != nil {
		t.Fatal(err)
	}

	frags, errs := JoinOpenApi(path, false, nil)
	if len(errs) != 0 {
		t.Fatalf("неожиданные ошибки: %v", errs)
	}
	if len(frags) != 1 {
		t.Fatalf("ожидался 1 фрагмент, получено %d", len(frags))
	}
	if frags[0].Origin.File != "test.go" {
		t.Errorf("ожидался File=test.go, получено %s", frags[0].Origin.File)
	}
}

func TestJoinOpenApi2(t *testing.T) {
	dir := t.TempDir()
	src := `package p

// @openapi
// paths:
//   /a:
//     get:
//       operationId: a
func F() {}
`
	err := os.WriteFile(filepath.Join(dir, "test.go"), []byte(src), 0644)
	if err != nil {
		t.Fatal(err)
	}

	frags, errs := JoinOpenApi(dir, false, nil)
	if len(errs) != 0 {
		t.Fatalf("неожиданные ошибки: %v", errs)
	}
	if len(frags) != 1 {
		t.Fatalf("ожидался 1 фрагмент, получено %d", len(frags))
	}
}

func TestJoinOpenApi3(t *testing.T) {
	_, errs := JoinOpenApi(filepath.Join(t.TempDir(), "no-file"), false, nil)
	if len(errs) == 0 {
		t.Fatal("ожидалась ошибка для несуществующего source")
	}
}

func TestJoinOpenApi4(t *testing.T) {
	dir := t.TempDir()
	src1 := `package p
func F( {
`

	src2 := `package p
// @openapi
// paths: {}
func G() {}
`
	err := os.WriteFile(filepath.Join(dir, "bad.go"), []byte(src1), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(dir, "good.go"), []byte(src2), 0644)
	if err != nil {
		t.Fatal(err)
	}

	frags, errs := JoinOpenApi(dir, false, nil)
	if len(errs) != 1 {
		t.Fatalf("ожидалась 1 ошибка (bad.go), получено %d: %v", len(errs), errs)
	}
	if len(frags) != 1 {
		t.Fatalf("good.go должен был обработаться, получено %d фрагментов", len(frags))
	}
}
