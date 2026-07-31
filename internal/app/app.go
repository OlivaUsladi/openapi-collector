package app

import (
	"flag"
	"fmt"
)

func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("нужно указать команду: oapic list, generate, validate")
	}
	cmd := args[0]
	//if !strings.Contains(cmd, "oapic") {
	//	return fmt.Errorf("команда должна начинаться с oapic")
	//}
	lst := flag.NewFlagSet("list", flag.ContinueOnError)
	rest := args[1:]
	switch cmd {
	case "list":
		base := lst.String("base", "", "базовая спецификация")
		source := lst.String("source", "", "файл или каталог с Go-кодом")
		out := lst.String("out", "openapi.yaml", "выходной файл команды generate")
		format := lst.String("format", "yaml", "yaml или json")
		err := lst.Parse(rest)
		if err != nil {
			return err
		}
		if *base == "" || *source == "" {
			return fmt.Errorf("--base и --source не могут быть пустыми")
		}
		fmt.Println(*base, *source, *out, *format)
	default:
		return fmt.Errorf("неизвестная команда: %s", cmd)
	}
	return nil
}
