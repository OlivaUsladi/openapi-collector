package app

import (
	"flag"
	"fmt"
	"io"
)

/*
Пример вывода на контрольном наборе с ошибками:
Errors: 4, warnings: 1
*/

func runValidate(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	base := fs.String("base", "", "базовая спецификация (YAML или JSON)")
	source := fs.String("source", "", "файл или каталог с Go-кодом")
	includeTests := fs.Bool("include-tests", false, "анализировать файлы _test.go")
	verbose := fs.Bool("verbose", false, "печатать предупреждения")
	var excludes stringList
	fs.Var(&excludes, "exclude", "glob-маска исключаемых путей")

	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if *base == "" || *source == "" {
		return fmt.Errorf("--base и --source обязательны")
	}

	build, err := buildSpec(*base, *source, *includeTests, excludes, true)
	if err != nil {
		return err
	}
	err = build.report(stderr, *verbose)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, "OK")
	return nil
}
