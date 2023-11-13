package templatemap

import (
	"html/template"
	"io"
	"path/filepath"
)

type TemplateMap struct {
	templates map[string]*template.Template
}

func (tm TemplateMap) Render(wr io.Writer, name string, data any) error {
	return tm.templates[name].ExecuteTemplate(wr, name, data)
}

func NewTemplateMap(layoutsPath, pagesPath string) (*TemplateMap, error) {
	tpls := make(map[string]*template.Template)
	layouts, err := filepath.Glob(layoutsPath)
	if err != nil {
		return nil, err
	}
	pages, err := filepath.Glob(pagesPath)
	if err != nil {
		return nil, err
	}
	for _, p := range pages {
		files := append(layouts, p)
		tpls[filepath.Base(p)] = template.Must(template.ParseFiles(files...))
	}
	return &TemplateMap{templates: tpls}, nil
}
