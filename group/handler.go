package group

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/josuebrunel/sportdropin/pkg/errorsmap"
	"github.com/josuebrunel/sportdropin/pkg/service"
	"github.com/josuebrunel/sportdropin/pkg/view"
	"github.com/josuebrunel/sportdropin/pkg/view/component"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/daos"
)

const hx_trigger_group = "groupChange"

var (
	seasonSVC service.Service
)

type ErrorResponse struct {
	Error string
}

func NewErrorResponse(err error) ErrorResponse {
	return ErrorResponse{Error: err.Error()}
}

type GroupHandler struct {
	svc service.Service
}

func NewGroupHandler(db *daos.Dao) *GroupHandler {
	seasonSVC = service.NewService("seasons", "seasonid", db)
	return &GroupHandler{svc: service.NewService("groups", "groupid", db)}
}

func (h GroupHandler) Create(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var (
			err error
			req = service.Request{}
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
		id := ctx.PathParam(h.svc.GetID())
		xlog.Debug("get", "group-id", id)
		resp, err := h.svc.GetByIDAndExpand(context, id, map[string]map[string]any{"seasons": map[string]any{"group": id}})
		if err != nil {
			xlog.Error("service", "error", err, "group", id)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		xlog.Debug("get", "group", resp)
		return view.Render(ctx, http.StatusOK, GroupDetailView(resp), nil)
	}
}

func (h GroupHandler) Update(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.PathParam(h.svc.GetID())
		if ctx.Request().Method == http.MethodGet {
			return view.Render(ctx, http.StatusOK, GroupFormView(
				view.NewViewData(h.svc.GetNewRecord(), errorsmap.New()),
				map[string]any{"hx-patch": reverse(ctx, "group.update", id)}), nil)
		}
		req := service.Request{"id": id}
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
		id := ctx.PathParam(h.svc.GetID())
		if err := h.svc.Delete(context, id); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		resp, err := h.svc.List(context, map[string]any{})
		if err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		return view.Render(ctx, http.StatusOK, GroupListView(resp), nil)

	}
}

func (h GroupHandler) SeasonCreate(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		groupID := ctx.PathParam(h.svc.GetID())
		if ctx.Request().Method == http.MethodGet {
			return view.Render(ctx, http.StatusOK,
				GroupSeasonForm(
					view.NewViewData(seasonSVC.GetNewRecord(), errorsmap.New()),
					templ.Attributes{"hx-post": ctx.RouteInfo().Reverse(groupID)}),
				nil)
		}
		req := service.Request{}
		if err := ctx.Bind(&req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		req["group"] = groupID
		_, err := seasonSVC.Create(context, req)
		if err != nil {
			xlog.Error("error while creating season", "req", req, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		seasons, err := seasonSVC.List(context, map[string]any{"group": groupID})
		if err != nil {
			xlog.Error("error while getting seasons", "group", groupID, "error", err)
		}
		xlog.Debug("seasons", "seasons", seasons)
		return view.Render(ctx, http.StatusOK, GroupSeasonList(groupID, seasons), nil)
	}
}

func (h GroupHandler) SeasonList(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		groupID := ctx.PathParam(h.svc.GetID())
		seasons, err := seasonSVC.List(context, map[string]any{"group": groupID})
		if err != nil {
			xlog.Error("error while getting seasons", "group", groupID, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		return view.Render(ctx, http.StatusOK, GroupSeasonList(groupID, seasons), nil)

	}
}

func (h GroupHandler) SeasonEdit(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		groupID := ctx.PathParam(h.svc.GetID())
		seasonID := ctx.PathParam(seasonSVC.GetID())
		vd, _ := seasonSVC.GetByID(context, seasonID)
		if ctx.Request().Method == http.MethodGet {
			return view.Render(ctx, http.StatusOK,
				GroupSeasonForm(vd, templ.Attributes{"hx-patch": ctx.RouteInfo().Reverse(groupID, seasonID)}),
				nil)
		}
		req := service.Request{}
		if err := ctx.Bind(&req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		req["id"] = seasonID
		req["group"] = groupID
		_, err := seasonSVC.Update(context, req)
		if err != nil {
			xlog.Error("error while creating season", "req", req, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		seasons, err := seasonSVC.List(context, map[string]any{"group": groupID})
		if err != nil {
			xlog.Error("error while getting seasons", "group", groupID, "error", err)
		}
		xlog.Debug("seasons", "seasons", seasons)
		return view.Render(ctx, http.StatusOK, GroupSeasonList(groupID, seasons), nil)
	}
}

func (h GroupHandler) SeasonDelete(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		groupID := ctx.PathParam(h.svc.GetID())
		seasonID := ctx.PathParam(seasonSVC.GetID())
		if err := seasonSVC.Delete(context, seasonID); err != nil {
			xlog.Error("error while deleting season", "season", seasonID, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		seasons, err := seasonSVC.List(context, map[string]any{"group": groupID})
		if err != nil {
			xlog.Error("error while getting seasons", "group", groupID, "error", err)
		}
		xlog.Debug("seasons", "seasons", seasons)
		return view.Render(ctx, http.StatusOK, GroupSeasonList(groupID, seasons), nil)
	}
}

var reverse = func(c echo.Context, name string, values ...any) string {
	path, err := c.Echo().Router().Routes().Reverse(name, values...)
	if err != nil {
		xlog.Error("error while reversing route", "name", name, "values", values, "error", err)
	}
	return path
}
