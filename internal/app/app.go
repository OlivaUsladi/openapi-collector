package app

import (
	"flag"
	"fmt"
	"io"
	"openapi-collector/internal/extract"
	"strings"
)

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func Run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("нужно указать команду: oapic list, generate, validate")
	}
	cmd := args[0]
	rest := args[1:]
	//if !strings.Contains(cmd, "oapic") {
	//	return fmt.Errorf("команда должна начинаться с oapic")
	//}
	switch cmd {
	case "list":
		return runList(rest, stdout, stderr)
	case "validate":
		return runValidate(rest, stdout, stderr)
	case "generate":
		return runGenerate(rest, stdout, stderr)
	default:
		return fmt.Errorf("неизвестная команда: %s", cmd)
	}
}

func runList(args []string, stdout, stderr io.Writer) error {
	lst := flag.NewFlagSet("list", flag.ContinueOnError)
	lst.SetOutput(stderr)
	includeTests := lst.Bool("include-tests", false, "анализировать файлы _test.go")
	var excludes stringList
	lst.Var(&excludes, "exclude", "glob-маска исключаемых путей")
	source := lst.String("source", "", "файл или каталог с Go-кодом")

	err := lst.Parse(args)
	if err != nil {
		return err
	}
	if *source == "" {
		return fmt.Errorf("--source не может быть пустым")
	}

	fr, errs := extract.JoinOpenApi(*source, *includeTests, excludes)

	for _, frag := range fr {
		fmt.Fprintf(stdout, "%s:%d:%d  [%s]\n",
			frag.Origin.File, frag.Origin.Line, frag.Origin.Column,
			strings.Join(frag.Sections, ", "))
	}
	fmt.Fprintf(stdout, "Всего фрагментов: %d\n", len(fr))

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(stderr, e)
		}
		return fmt.Errorf("%d ошибки во время выполнения", len(errs))
	}

	return nil
}
