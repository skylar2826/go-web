package template

import (
	"bytes"
	"context"
	"html/template"
)

type TemplateEngine interface {
	Render(ctx context.Context, tplName string, data interface{}) ([]byte, error)
}

type GoTemplateEngine struct {
	Template *template.Template
}

func (g *GoTemplateEngine) Render(ctx context.Context, tplName string, data interface{}) ([]byte, error) {
	val := &bytes.Buffer{}
	err := g.Template.ExecuteTemplate(val, tplName, data)
	if err != nil {
		return nil, err
	}
	return val.Bytes(), nil
}
