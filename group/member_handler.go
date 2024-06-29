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

func (h GroupHandler) MemberCreate(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		groupID := ctx.PathParam(h.svc.GetID())
		if ctx.Request().Method == http.MethodGet {
			return view.Render(ctx, http.StatusOK,
				GroupMemberForm(
					view.NewViewData(memberSVC.GetNewRecord(), errorsmap.New()),
					templ.Attributes{"hx-post": ctx.RouteInfo().Reverse(groupID)}),
				nil)
		}
		req := service.Request{}
		if err := ctx.Bind(&req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		req["group"] = groupID
		_, err := memberSVC.Create(context, req)
		if err != nil {
			xlog.Error("error while creating member", "req", req, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		members, err := memberSVC.List(context, map[string]any{"group": groupID})
		if err != nil {
			xlog.Error("error while getting members", "group", groupID, "error", err)
		}
		xlog.Debug("members", "members", members)
		return view.Render(ctx, http.StatusOK, GroupMemberList(groupID, members), nil)
	}
}

func (h GroupHandler) MemberList(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		groupID := ctx.PathParam(h.svc.GetID())
		members, err := memberSVC.List(context, map[string]any{"group": groupID})
		if err != nil {
			xlog.Error("error while getting members", "group", groupID, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		return view.Render(ctx, http.StatusOK, GroupMemberList(groupID, members), nil)

	}
}

func (h GroupHandler) MemberEdit(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		groupID := ctx.PathParam(h.svc.GetID())
		memberID := ctx.PathParam(memberSVC.GetID())
		vd, _ := memberSVC.GetByID(context, memberID)
		if ctx.Request().Method == http.MethodGet {
			xlog.Debug("member", "member", vd)
			return view.Render(ctx, http.StatusOK,
				GroupMemberForm(vd, templ.Attributes{"hx-patch": ctx.RouteInfo().Reverse(groupID, memberID)}),
				nil)
		}
		req := service.Request{}
		if err := ctx.Bind(&req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		req["id"] = memberID
		req["group"] = groupID
		_, err := memberSVC.Update(context, req)
		if err != nil {
			xlog.Error("error while creating member", "req", req, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		members, err := memberSVC.List(context, map[string]any{"group": groupID})
		if err != nil {
			xlog.Error("error while getting members", "group", groupID, "error", err)
		}
		xlog.Debug("members", "members", members)
		return view.Render(ctx, http.StatusOK, GroupMemberList(groupID, members), nil)
	}
}

func (h GroupHandler) MemberDelete(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		groupID := ctx.PathParam(h.svc.GetID())
		memberID := ctx.PathParam(memberSVC.GetID())
		if err := memberSVC.Delete(context, memberID); err != nil {
			xlog.Error("error while deleting member", "member", memberID, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		members, err := memberSVC.List(context, map[string]any{"group": groupID})
		if err != nil {
			xlog.Error("error while getting members", "group", groupID, "error", err)
		}
		xlog.Debug("members", "members", members)
		return view.Render(ctx, http.StatusOK, GroupMemberList(groupID, members), nil)
	}
}
