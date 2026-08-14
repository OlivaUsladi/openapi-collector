package app

import (
	"flag"
	"fmt"
	"io"
	"openapi-collector/internal/extract"
	"openapi-collector/internal/merge"
	"openapi-collector/internal/spec"
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
	var excludes stringList
	fs.Var(&excludes, "exclude", "glob-маска исключаемых путей")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *base == "" || *source == "" {
		return fmt.Errorf("--base и --source обязательны")
	}

	baseDoc, err := spec.LoadBase(*base)
	if err != nil {
		return err
	}

	fragments, extractErrs := extract.JoinOpenApi(*source, *includeTests, excludes)
	result := merge.Merge(baseDoc, fragments)

	totalErrors := len(extractErrs) + len(result.Errors) + len(result.Conflicts)
	fmt.Fprintf(stderr, "Errors: %d, warnings: 0\n", totalErrors)

	for _, e := range extractErrs {
		fmt.Fprintln(stderr, e)
	}
	for _, e := range result.Errors {
		fmt.Fprintln(stderr, e)
	}
	for _, c := range result.Conflicts {
		fmt.Fprintln(stderr, c)
	}

	if totalErrors > 0 {
		return fmt.Errorf("validate: %d", totalErrors)
	}
	fmt.Fprintln(stdout, "OK")
	return nil
}
