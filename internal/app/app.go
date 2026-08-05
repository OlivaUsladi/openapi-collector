package app

import (
	"flag"
	"fmt"
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

func Run(args []string) error {
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
		return runList(rest)
	default:
		return fmt.Errorf("неизвестная команда: %s", cmd)
	}
	return nil
}

func runList(args []string) error {
	lst := flag.NewFlagSet("list", flag.ContinueOnError)
	includeTests := lst.Bool("include-tests", false, "анализировать файлы _test.go")
	var excludes stringList
	lst.Var(&excludes, "exclude", "glob-маска исключаемых путей")
	source := lst.String("source", "", "файл или каталог с Go-кодом")
	format := lst.String("format", "yaml", "yaml или json")

	//info, err := os.Stat(*source)
	//if err != nil {
	//	return err
	//}
	//isDir := info.IsDir()

	//files, err := extract.GoFiles(*source, *includeTests, excludes)
	//if err != nil {
	//	return err
	//}

	err := lst.Parse(args)
	if err != nil {
		return err
	}
	if *source == "" {
		return fmt.Errorf("--base и --source не могут быть пустыми")
	}
	fmt.Println(*source, *includeTests, *format)
	return nil
}
