package extract

import "testing"

func TestFindMarkedComments(t *testing.T) {
	comments, err := FindAPIComments("../../testdata/tests/test1.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) == 0 {
		t.Fatal("ожидался хотя бы один комментарий с @openapi")
	}
}
