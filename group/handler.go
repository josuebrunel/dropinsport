package group

import (
	"context"
	"net/http"

	"github.com/josuebrunel/sportdropin/pkg/storage"
	"github.com/josuebrunel/sportdropin/pkg/view"
	"github.com/josuebrunel/sportdropin/pkg/view/component"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/labstack/echo/v4"
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

func NewGroupHandler(store storage.Storer) *GroupHandler {
	return &GroupHandler{svc: NewService("group", "uuid", store)}
}

func (h GroupHandler) Create(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var (
			err error
			req = h.svc.GetRequest()
		)
		if ctx.Request().Method == http.MethodGet {
			return view.Render(ctx, http.StatusOK, GroupFormView(Response{}, map[string]any{"hx-post": reverse(ctx, "group.create")}), nil)
		}
		if err = ctx.Bind(req); err != nil {
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
		return view.Render(ctx, http.StatusOK, GroupListView(resp.(Response)), nil)

	}

}

func (h GroupHandler) Get(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uuid := ctx.Param(h.svc.GetID())
		xlog.Info("get", "group-uuid", uuid)
		req := h.svc.GetRequest()
		if err := req.SetID(uuid); err != nil {
			return view.Render(ctx, http.StatusOK, GroupFormView(Response{}, map[string]any{"hx-get": reverse(ctx, "group.list")}), nil)
		}
		resp, err := h.svc.Get(context, req)
		if err != nil {
			xlog.Error("service", "error", err, "resp", resp)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		xlog.Info("get", "group", resp)
		return view.Render(ctx, http.StatusOK, GroupFormView(resp.(Response), map[string]any{
			"hx-patch": reverse(ctx, "group.update", resp.(Response).One().UUID.String())}), nil)
	}
}

func (h GroupHandler) Update(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uuid := ctx.Param(h.svc.GetID())
		if ctx.Request().Method == http.MethodGet {
			return view.Render(ctx, http.StatusOK, GroupFormView(Response{},
				map[string]any{"hx-patch": reverse(ctx, "group.update", uuid)}), nil)
		}
		req := h.svc.GetRequest()
		if err := ctx.Bind(req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		if err := req.SetID(uuid); err != nil {
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
		return view.Render(ctx, http.StatusOK, GroupListView(resp.(Response)), nil)
	}
}

func (h GroupHandler) List(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var filters = make(map[string]any)
		if city := ctx.QueryParam("search"); city != "" {
			filters["city"] = city
		}
		resp, err := h.svc.List(context, filters)
		if err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		return view.Render(ctx, http.StatusOK, GroupListView(resp.(Response)), nil)
	}
}

func (h GroupHandler) Delete(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uuid := ctx.Param(h.svc.GetID())
		req := h.svc.GetRequest()
		if err := req.SetID(uuid); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		_, err := h.svc.Delete(context, req)
		if err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		resp, err := h.svc.List(context, map[string]any{})
		if err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		ctx.Response().Header().Set("HX-Trigger", hx_trigger_group)
		return view.Render(ctx, http.StatusOK, GroupListView(resp.(Response)), nil)

	}
}

var reverse = func(c echo.Context, name string, values ...any) string {
	return c.Echo().Reverse(name, values...)
}
