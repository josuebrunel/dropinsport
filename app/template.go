package app

import (
	"html/template"
	"io"
	"net/http"

	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/josuebrunel/templatesmap"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type TemplateRenderer struct {
	templateMap *templatesmap.TemplatesMap
}

func (t TemplateRenderer) Render(wr io.Writer, name string, data any, ctx echo.Context) error {
	var context = NewTemplateContext(ctx, data)
	return t.templateMap.Render(wr, name, context)
}

func NewTemplateRenderer(layouts string, funcs template.FuncMap, pages ...string) (*TemplateRenderer, error) {
	tpl, err := templatesmap.NewTemplatesMap(layouts, funcs, pages...)
	if err != nil {
		xlog.Error("templatemap", "error", err)
		return nil, err
	}
	return &TemplateRenderer{templateMap: tpl}, nil
}

type TemplateContext map[string]any

func (tc TemplateContext) IsHXRequest() bool {
	req := tc["request"].(*http.Request)
	val := req.Header.Get("HX-Request")
	return val != ""
}

func NewTemplateContext(ctx echo.Context, data any) TemplateContext {
	tc := make(TemplateContext)
	tc["request"] = ctx.Request()
	tc["url"] = ctx.Request().URL.String()
	tc["reverse"] = ctx.Echo().Reverse
	tc["data"] = data
	tc["csrf"] = ctx.Get(middleware.DefaultCSRFConfig.ContextKey).(string)
	return tc
}
