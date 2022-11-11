package servetpl

import (
	"fmt"
	"html/template"
	"os"
	textTemplate "text/template"

	"github.com/eternal-flame-AD/yoake/internal/servetpl/funcmap"
)

func ParseTemplateFileAs[M interface{ ~map[string]any }, T interface {
	*template.Template | *textTemplate.Template
	Parse(string) (T, error)
	New(name string) T
	Funcs(funcs M) T
}](tpl T, name string, path string) (T, error) {
	slurpedFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading template file %s: %w", path, err)
	}

	res, err := tpl.New(name).Funcs(funcmap.GetFuncMap()).Parse(string(slurpedFile))
	return res, err
}
