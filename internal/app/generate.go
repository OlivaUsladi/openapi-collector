package app

import (
	"flag"
	"fmt"
	"io"
	"openapi-collector/internal/atomicfile"
	"openapi-collector/internal/output"
	"path/filepath"
	"strings"
)

func runGenerate(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	base := fs.String("base", "", "базовая спецификация (YAML или JSON)")
	source := fs.String("source", "", "файл или каталог с Go-кодом")
	out := fs.String("out", "", "файл результата (по умолчанию stdout)")
	format := fs.String("format", "yaml", "формат результата: yaml или json")
	includeTests := fs.Bool("include-tests", false, "анализировать файлы _test.go")
	verbose := fs.Bool("verbose", false, "печатать предупреждения")
	var excludes stringList
	fs.Var(&excludes, "exclude", "glob-маска исключаемых путей")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *base == "" || *source == "" {
		return fmt.Errorf("--base и --source обязательны")
	}
	if *format != "yaml" && *format != "json" {
		return fmt.Errorf("--format должен быть yaml или json, получено %q", *format)
	}

	build, err := buildSpec(*base, *source, *includeTests, excludes, true)
	if err != nil {
		return err
	}

	err = build.report(stderr, *verbose)
	if err != nil {
		return err
	}

	var data []byte
	if *format == "json" {
		data, err = output.MarshalJSON(build.doc)
	} else {
		data, err = output.MarshalYAML(build.doc)
	}
	if err != nil {
		return err
	}

	if *out == "" {
		_, err = stdout.Write(data)
		return err
	}

	ext := strings.TrimPrefix(filepath.Ext(*out), ".")
	if ext == "yml" {
		ext = "yaml"
	}
	if ext != "" && ext != *format {
		fmt.Fprintf(stderr, "warning: расширение файла %q не совпадает с форматом %q\n", *out, *format)
	}
	return atomicfile.WriteFile(*out, data)
}
