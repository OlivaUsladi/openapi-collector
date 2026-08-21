package app

import (
	"errors"
	"fmt"
	"io"
	"openapi-collector/internal/extract"
	"openapi-collector/internal/merge"
	"openapi-collector/internal/model"
	"openapi-collector/internal/refs"
	"openapi-collector/internal/spec"
	"openapi-collector/internal/validate"
)

var ErrIssues = errors.New("issues found")

type buildResult struct {
	doc      map[string]any
	owners   map[string]model.Origin
	errs     []error
	warnings []string
}

func buildSpec(basePath, source string, includeTests bool, excludes []string, checkRefs bool) (*buildResult, error) {
	baseDoc, err := spec.LoadBase(basePath)
	if err != nil {
		return nil, err
	}

	fragments, extractErrs := extract.JoinOpenApi(source, includeTests, excludes)
	result := merge.Merge(baseDoc, fragments)

	build := &buildResult{doc: result.Doc, owners: result.Owners}
	build.errs = append(build.errs, extractErrs...)
	build.errs = append(build.errs, result.Errors...)
	for _, c := range result.Conflicts {
		build.errs = append(build.errs, errors.New(c.String()))
	}

	if checkRefs {
		refErrs, warnings := refs.Check(result.Doc, result.Owners)
		build.errs = append(build.errs, refErrs...)
		build.warnings = append(build.warnings, warnings...)
	}
	build.errs = append(build.errs, validate.Check(result.Doc, result.Owners)...)

	return build, nil
}

func (b *buildResult) report(stderr io.Writer, verbose bool) error {
	fmt.Fprintf(stderr, "Errors: %d, warnings: %d\n", len(b.errs), len(b.warnings))
	for _, e := range b.errs {
		fmt.Fprintln(stderr, e)
	}
	if verbose {
		for _, w := range b.warnings {
			fmt.Fprintln(stderr, "warning:", w)
		}
	}
	if len(b.errs) > 0 {
		return fmt.Errorf("сборка: %d проблемы: %w", len(b.errs), ErrIssues)
	}
	return nil
}
