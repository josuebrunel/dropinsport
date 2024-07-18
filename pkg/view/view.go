package view

import (
	"context"
	"net/http"
	"net/url"

	"github.com/a-h/templ"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/labstack/echo/v5"
)

type xcontextkey string

var xc xcontextkey = "xcontext"

type ViewData[T any] struct {
	Data   T
	Errors map[string]error
}

func (v ViewData[T]) ErrNil(key string) bool   { return v.Errors[key] == nil }
func (v ViewData[T]) ErrGet(key string) string { return v.Errors[key].Error() }
func (v ViewData[T]) V() T                     { return v.Data }

func NewViewData[T any](data T, errs map[string]error) ViewData[T] {
	return ViewData[T]{Data: data, Errors: errs}
}

func Render(ctx echo.Context, status int, tpl templ.Component, data any) error {
	ctx.Response().Writer.WriteHeader(status)

	var csrf string
	if v := ctx.Get("csrf"); v != nil {
		csrf = v.(string)
	}
	cx := ctx.Request().Context()
	cx = context.WithValue(cx, xc, map[string]any{
		"request": ctx.Request(),
		"url":     ctx.Request().URL.String(),
		"reverse": ctx.Echo().Router().Routes().Reverse,
		"csrf":    csrf,
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

func Reverse(cx context.Context, name string, values ...any) string {
	reverse := Get[func(string, ...any) (string, error)](cx, "reverse")
	path, err := reverse(name, values...)
	if err != nil {
		xlog.Error("failed to reverse route", "name", name, "values", values, "error", err)
	}
	return path
}

func ReverseX(c echo.Context, name string, values ...any) string {
	path, err := c.Echo().Router().Routes().Reverse(name, values...)
	if err != nil {
		xlog.Error("error while reversing route", "name", name, "values", values, "error", err)
	}
	return path
}

func WithQS(url_ string, qs map[string]string) string {
	u, err := url.Parse(url_)
	if err != nil {
		xlog.Error("failed to parse url", "url", url_, "error", err)
		return url_
	}
	q := u.Query()
	for k, v := range qs {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
