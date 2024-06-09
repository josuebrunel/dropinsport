package group

import (
	"context"
	"net/http"

	"github.com/josuebrunel/sportdropin/pkg/errorsmap"
	"github.com/josuebrunel/sportdropin/pkg/view"
	"github.com/josuebrunel/sportdropin/pkg/view/component"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/daos"
)

const hx_trigger_group = "groupChange"

type ErrorResponse struct {
	Error string
}

func NewErrorResponse(err error) ErrorResponse {
	return ErrorResponse{Error: err.Error()}
}

type GroupHandler struct {
	svc Service
}

func NewGroupHandler(db *daos.Dao) *GroupHandler {
	return &GroupHandler{svc: NewService("groups", "uuid", db)}
}

func (h GroupHandler) Create(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var (
			err error
			req = Request{}
		)
		if ctx.Request().Method == http.MethodGet {
			return view.Render(ctx, http.StatusOK, GroupFormView(
				view.NewViewData(h.svc.GetNewRecord(), errorsmap.New()),
				map[string]any{"hx-post": reverse(ctx, "group.create")}), nil)
		}
		if err = ctx.Bind(&req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}

		_, err = h.svc.Create(context, req)
		if err != nil {
			xlog.Error("group-handler-create", "errors", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		resp, err := h.svc.List(context, map[string]any{})
		if err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		ctx.Response().Header().Set("HX-Trigger", hx_trigger_group)
		return view.Render(ctx, http.StatusOK, GroupListView(resp), nil)

	}

}

func (h GroupHandler) Get(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uuid := ctx.PathParam(h.svc.GetID())
		xlog.Info("get", "group-uuid", uuid)
		req := Request{UUID: uuid}
		resp, err := h.svc.Get(context, req)
		if err != nil {
			xlog.Error("service", "error", err, "group", uuid)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		xlog.Info("get", "group", uuid)
		return view.Render(ctx, http.StatusOK, GroupFormView(resp, map[string]any{
			"hx-patch": reverse(ctx, "group.update", uuid)}), nil)
	}
}

func (h GroupHandler) Update(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uuid := ctx.PathParam(h.svc.GetID())
		if ctx.Request().Method == http.MethodGet {
			return view.Render(ctx, http.StatusOK, GroupFormView(
				view.NewViewData(h.svc.GetNewRecord(), errorsmap.New()),
				map[string]any{"hx-patch": reverse(ctx, "group.update", uuid)}), nil)
		}
		req := Request{UUID: uuid}
		if err := ctx.Bind(&req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		_, err := h.svc.Update(context, req)
		if err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		resp, err := h.svc.List(context, map[string]any{})
		if err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		ctx.Response().Header().Set("HX-Trigger", hx_trigger_group)
		return view.Render(ctx, http.StatusOK, GroupListView(resp), nil)
	}
}

func (h GroupHandler) List(context context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var filters = make(map[string]any)
		if city := c.QueryParam("search"); city != "" {
			filters["city"] = city
		}
		resp, err := h.svc.List(context, filters)
		if err != nil {
			return view.Render(c, http.StatusOK, component.Error(err.Error()), nil)
		}
		return view.Render(c, http.StatusOK, GroupListView(resp), nil)
	}
}

func (h GroupHandler) Delete(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uuid := ctx.PathParam(h.svc.GetID())
		req := Request{UUID: uuid}
		if err := h.svc.Delete(context, req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		resp, err := h.svc.List(context, map[string]any{})
		if err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		ctx.Response().Header().Set("HX-Trigger", hx_trigger_group)
		return view.Render(ctx, http.StatusOK, GroupListView(resp), nil)

	}
}

var reverse = func(c echo.Context, name string, values ...any) string {
	path, err := c.Echo().Router().Routes().Reverse(name, values...)
	if err != nil {
		xlog.Error("error while reversing route", "name", name, "values", values, "error", err)
	}
	return path
}
