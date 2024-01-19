package group

import (
	"context"
	"net/http"

	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/josuebrunel/sportdropin/storage"
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
		if err = ctx.Bind(req); err != nil {
			return ctx.Render(http.StatusOK, "group-form.html", NewErrorResponse(err))
		}

		resp, err := h.svc.Create(context, req)
		if err != nil {
			xlog.Info("group-handler-create", "errors", err)
			ctx.Response().Header().Set("HX-Retarget", "#group-modal")
			ctx.Response().Header().Set("HX-Reswap", "outerHTML")
			return ctx.Render(http.StatusOK, "group-form.html", resp)
		}
		ctx.Response().Header().Set("HX-Trigger", hx_trigger_group)
		return ctx.Render(resp.GetStatusCode(), "group-detail.html", resp)
	}

}

func (h GroupHandler) Get(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		uuid := ctx.Param(h.svc.GetID())
		xlog.Info("get", "group-uuid", uuid)
		if uuid == "" {
			return ctx.Render(http.StatusOK, "group-form.html", Response{StatusCode: 200, Data: Group{}})
		}
		req := h.svc.GetRequest()
		if err := req.SetID(uuid); err != nil {
			return ctx.Render(http.StatusInternalServerError, "error.html", NewErrorResponse(err))
		}
		resp, err := h.svc.Get(context, req)
		if err != nil {
			xlog.Error("service", "error", err, "resp", resp)
			return ctx.Render(resp.GetStatusCode(), "error.html", NewErrorResponse(err))
		}
		xlog.Info("get", "group", resp)
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
		ctx.Response().Header().Set("HX-Trigger", hx_trigger_group)
		return ctx.Render(resp.GetStatusCode(), "group-form.html", resp)
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
			return ctx.Render(resp.GetStatusCode(), "error.html", NewErrorResponse(err))
		}
		return ctx.Render(resp.GetStatusCode(), "group-list.html", resp)
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
		ctx.Response().Header().Set("HX-Trigger", hx_trigger_group)
		return ctx.Render(resp.GetStatusCode(), "group-list.html", nil)
	}
}
