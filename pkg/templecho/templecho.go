package templecho

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type xcontextkey string

var xc xcontextkey = "xcontext"

func Render(ctx echo.Context, status int, tpl templ.Component, data any) error {
	ctx.Response().Writer.WriteHeader(status)

	cx := context.Background()
	cx = context.WithValue(cx, xc, map[string]any{
		"request": ctx.Request(),
		"url":     ctx.Request().URL.String(),
		"reverse": ctx.Echo().Reverse,
		"csrf":    ctx.Get(middleware.DefaultCSRFConfig.ContextKey).(string),
		"data":    data,
	})

	err := tpl.Render(cx, ctx.Response().Writer)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, "failed to render response template")
	}

	return nil
}

func Get[T any](ctx context.Context, key string) T {
	cx := ctx.Value(xc).(map[string]any)
	var r T
	if v, ok := cx[key]; ok {
		r = v.(T)
	}
	return r
}
