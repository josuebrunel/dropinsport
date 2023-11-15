package group

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/josuebrunel/sportdropin/storage"
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Error string
}

func NewErrorResponse(err error) ErrorResponse {
	return ErrorResponse{Error: err.Error()}
}

type GroupHandler struct {
	svc Service
}

func NewGroupHandler(store storage.Storer) *GroupHandler {
	return &GroupHandler{svc: NewService("group", "uuid", store)}
}

func (h GroupHandler) Create(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var (
			err error
			req = h.svc.GetRequest()
		)
		if err = ctx.Bind(req); err != nil {
			return ctx.Render(http.StatusBadRequest, "error.html", NewErrorResponse(err))
		}

		resp, err := h.svc.Create(context, req)
		if err != nil {
			return ctx.Render(resp.GetStatusCode(), "error.html", NewErrorResponse(err))
		}
		r := resp.(Response)
		g := r.Data.(Group)
		return ctx.Redirect(http.StatusFound, fmt.Sprintf("/group/%s/", g.UUID))
	}
}

func (h GroupHandler) Get(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uuid := ctx.Param(h.svc.GetID())
		slog.Info("get", "group-uuid", uuid)
		if uuid == "" {
			return ctx.Render(http.StatusOK, "group-form.html", Response{StatusCode: 200, Data: Group{}})
		}
		req := h.svc.GetRequest()
		if err := req.SetID(uuid); err != nil {
			return ctx.Render(http.StatusInternalServerError, "error.html", NewErrorResponse(err))
		}
		resp, err := h.svc.Get(context, req)
		if err != nil {
			slog.Error("service", "error", err, "resp", resp)
			return ctx.Render(resp.GetStatusCode(), "error.html", NewErrorResponse(err))
		}
		slog.Info("get", "group", resp)
		return ctx.Render(resp.GetStatusCode(), "group-form.html", resp)
	}
}

func (h GroupHandler) Update(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		req := h.svc.GetRequest()
		if err := ctx.Bind(req); err != nil {
			return ctx.Render(http.StatusBadRequest, "error.html", NewErrorResponse(err))
		}
		uuid := ctx.Param(h.svc.GetID())
		if err := req.SetID(uuid); err != nil {
			return ctx.Render(http.StatusInternalServerError, "error.html", NewErrorResponse(err))
		}
		resp, err := h.svc.Update(context, req)
		if err != nil {
			return ctx.Render(resp.GetStatusCode(), "error.html", NewErrorResponse(err))
		}
		return ctx.Render(resp.GetStatusCode(), "group-form.html", resp)
	}
}

func (h GroupHandler) List(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var filters = make(map[string]any)
		if city := ctx.QueryParam("search"); city != "" {
			filters["city"] = city
		}

		var tpl = "group-list.html"
		if ctx.Request().Header.Get("Hx-Request") == "true" {
			tpl = "group-hx-list.html"
		}
		resp, err := h.svc.List(context, filters)
		if err != nil {
			return ctx.Render(resp.GetStatusCode(), "error.html", NewErrorResponse(err))
		}
		return ctx.Render(resp.GetStatusCode(), tpl, resp)
	}
}

func (h GroupHandler) Delete(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uuid := ctx.Param(h.svc.GetID())
		req := h.svc.GetRequest()
		if err := req.SetID(uuid); err != nil {
			return ctx.Render(http.StatusInternalServerError, "error.html", NewErrorResponse(err))
		}
		resp, err := h.svc.Delete(context, req)
		if err != nil {
			return ctx.Render(resp.GetStatusCode(), "error.html", NewErrorResponse(err))
		}
		return ctx.Redirect(http.StatusFound, "/")
	}
}
