package output

import (
	"strings"
	"testing"
)

func TestMarshalJSON(t *testing.T) {
	first, err := MarshalJSON(sampleDoc())
	if err != nil {
		t.Fatal(err)
	}
	second, err := MarshalJSON(sampleDoc())
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("повторный вызов дал другой результат")
	}
	text := string(first)
	if !strings.Contains(text, "Задачи <и> дела") {
		t.Errorf("юникод и <> не должны экранироваться:\n%s", text)
	}
	if strings.Contains(text, `\u0417`) || strings.Contains(text, `\u003c`) {
		t.Errorf("найдено лишнее экранирование:\n%s", text)
	}
}
