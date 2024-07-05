package group

import (
	"context"
	"net/http"
	"sort"

	"github.com/josuebrunel/sportdropin/pkg/errorsmap"
	"github.com/josuebrunel/sportdropin/pkg/service"
	"github.com/josuebrunel/sportdropin/pkg/view"
	"github.com/josuebrunel/sportdropin/pkg/view/component"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/daos"
)

const (
	SeasonStatusInProgress = "inprogress"
	SeasonStatusClosed     = "closed"
	SeasonStatusScheduled  = "scheduled"
)

var (
	seasonSVC service.Service
	memberSVC service.Service
	statSVC   service.Service
)

type (
	SportStatSchema = []map[string]string
	SportMetaData   struct {
		Icon  string          `json:"icon"`
		Stats SportStatSchema `json:"stats"`
	}
)

type GroupHandler struct {
	svc service.Service
}

func NewGroupHandler(db *daos.Dao) *GroupHandler {
	seasonSVC = service.NewService("seasons", "seasonid", db)
	memberSVC = service.NewService("members", "memberid", db)
	statSVC = service.NewService("memberstats", "statid", db)
	return &GroupHandler{svc: service.NewService("groups", "groupid", db)}
}

func (h GroupHandler) GetGroupSportStatSchema(ctx context.Context, groupID string) SportStatSchema {
	group, err := h.svc.GetByID(ctx, groupID, "sport")
	if err != nil {
		return SportStatSchema{}
	}
	var md SportMetaData
	sport := group.Data.ExpandedOne("sport")
	sport.UnmarshalJSONField("data", &md)
	xlog.Debug("schema", "schema", md)
	return md.Stats
}

func (h GroupHandler) GetGroupCurrentSeason(ctx context.Context, groupID string) (service.Record, error) {
	seasons, err := seasonSVC.List(ctx, map[string]any{"group": groupID})
	if err != nil {
		return seasonSVC.GetNewRecord(), err
	}
	if len(seasons.Data) == 0 {
		return nil, nil
	}
	sort.Slice(seasons.Data, func(i, j int) bool {
		return seasons.Data[i].GetString("end_date") < seasons.Data[j].GetString("end_date")
	})
	var currentSeason service.Record
	for _, season := range seasons.Data {
		if season.GetString("status") == SeasonStatusInProgress {
			currentSeason = season
			break
		}
	}
	if currentSeason == nil {
		currentSeason = seasons.Data[0]
	}
	return currentSeason, nil
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
		return view.Render(ctx, http.StatusOK, GroupListView(resp), nil)

	}

}

func (h GroupHandler) Get(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.PathParam(h.svc.GetID())
		xlog.Debug("get", "group-id", id)
		resp, err := h.svc.GetByIDWithBackRel(context, id, service.BackRel{"seasons": map[string]any{"group": id}})
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
		resp, err := h.svc.List(context, filters, "sport")
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

var reverse = func(c echo.Context, name string, values ...any) string {
	path, err := c.Echo().Router().Routes().Reverse(name, values...)
	if err != nil {
		xlog.Error("error while reversing route", "name", name, "values", values, "error", err)
	}
	return path
}
