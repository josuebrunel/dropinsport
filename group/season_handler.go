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
)

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
