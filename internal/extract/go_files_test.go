package extract

import (
	"testing"
)

func TestGoFiles(t *testing.T) {
	str, err1 := GoFiles("C:\\Users\\Alexandra\\GolandProjects\\openapi-collector\\testdata", false)
	if err1 != nil {
		t.Fatal(err1)
	}
	if len(str) == 0 {
		t.Errorf("ожидался хотя бы один файл")
	}
	//Вот это временный тест
	if str[0] != "C:\\Users\\Alexandra\\GolandProjects\\openapi-collector\\testdata\\tests\\test1.go" {
		t.Errorf("ожидался файл test1.go, получился %s", str[0])
	}
}
