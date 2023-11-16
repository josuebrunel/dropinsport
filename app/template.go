package app

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/josuebrunel/templatesmap"
	"github.com/labstack/echo/v4"
)

type TemplateRenderer struct {
	templateMap *templatesmap.TemplatesMap
}

func (t TemplateRenderer) Render(wr io.Writer, name string, data any, ctx echo.Context) error {
	var context = NewTemplateContext(ctx, data)
	return t.templateMap.Render(wr, name, context)
}

func NewTemplateRenderer(layouts string, pages ...string) (*TemplateRenderer, error) {
	tpl, err := templatesmap.NewTemplatesMap(layouts, pages...)
	if err != nil {
		slog.Error("templatemap", "error", err)
		return nil, err
	}
	return &TemplateRenderer{templateMap: tpl}, nil
}

type TemplateContext map[string]any

func (tc TemplateContext) IsHXRequest() bool {
	req := tc["request"].(http.Request)
	val := req.Header.Get("Hx-Request")
	return val != ""
}

func NewTemplateContext(ctx echo.Context, data any) TemplateContext {
	tc := make(TemplateContext)
	tc["request"] = ctx.Request()
	tc["url"] = ctx.Request().URL.String()
	tc["reverse"] = ctx.Echo().Reverse
	tc["data"] = data
	return tc
}
