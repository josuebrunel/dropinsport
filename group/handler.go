package group

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/josuebrunel/sportdropin/pkg/errorsmap"
	"github.com/josuebrunel/sportdropin/pkg/models"
	pb "github.com/josuebrunel/sportdropin/pkg/pbclient"
	"github.com/josuebrunel/sportdropin/pkg/service"
	"github.com/josuebrunel/sportdropin/pkg/view"
	"github.com/josuebrunel/sportdropin/pkg/view/component"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/josuebrunel/sportdropin/pkg/xsession"
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
	sportSVC  service.Service
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
	api pb.Client
}

func NewGroupHandler(db *daos.Dao, url string) *GroupHandler {
	seasonSVC = service.NewService("seasons", "seasonid", db)
	memberSVC = service.NewService("members", "memberid", db)
	statSVC = service.NewService("memberstats", "statid", db)
	sportSVC = service.NewService("sports", "sportid", db)
	return &GroupHandler{svc: service.NewService("groups", "groupid", db), api: pb.New(url)}
}

func (h GroupHandler) GetGroup(id string) (service.Record, error) {
	v, err := h.svc.GetByID(context.Background(), id, "user", "sport", "seasons_via_group")
	if err != nil {
		xlog.Error("error while getting group", "group", id, "error", err)
	}
	return v.V(), err
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
		xlog.Error("failed to get group seasons", "group", groupID)
		return seasonSVC.GetNewRecord(), err
	}
	sort.Slice(seasons.Data, func(i, j int) bool {
		return seasons.Data[i].GetString("end_date") < seasons.Data[j].GetString("end_date")
	})
	var currentSeason = seasonSVC.GetNewRecord()
	for _, season := range seasons.Data {
		if strings.EqualFold(season.GetString("status"), SeasonStatusInProgress) {
			currentSeason = season
			break
		}
	}
	if currentSeason == nil && len(seasons.Data) > 0 {
		currentSeason = seasons.Data[0]
	}
	return currentSeason, nil
}

func (h GroupHandler) GetSports(ctx context.Context) (view.ViewData[service.RecordSlice], error) {
	sports, err := sportSVC.List(ctx, service.Filters{})
	if err != nil {
		xlog.Error("failed to get sport list", "error", err)
		return view.ViewData[service.RecordSlice]{}, err
	}
	return sports, nil

}

func (h GroupHandler) Create(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var (
			err error
			req = service.Request{}
		)
		if ctx.Request().Method == http.MethodGet {
			sports, err := h.GetSports(context)
			if err != nil {
				xlog.Error("error while getting sports", "error", err)
			}
			return view.Render(ctx, http.StatusOK, GroupFormView(
				view.NewViewData(h.svc.GetNewRecord(), errorsmap.New()), sports,
				templ.Attributes{"target": "#content", "hx-post": reverse(ctx, "group.create")}), nil)
		}
		if err = ctx.Bind(&req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}

		_, err = h.svc.Create(context, req)
		if err != nil {
			xlog.Error("group-handler-create", "errors", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		return ctx.Redirect(http.StatusFound, view.ReverseX(ctx, "account.get", xsession.GetUser(ctx.Request().Context()).ID))
	}

}

func (h GroupHandler) Get(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.PathParam(h.svc.GetID())
		token := xsession.Get[string](ctx.Request().Context(), xsession.SessionName)
		xlog.Debug("get", "group-id", id, "token", token)
		resp, err := h.api.RecordGet("groups", id, pb.QHeaders{"Authorization": token}, pb.QExpand{"sport"})
		if err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		group := pb.ResponseTo[models.Group](resp)
		xlog.Debug("get group", "group", group)
		return view.Render(ctx, http.StatusSeeOther, GroupDetailView(group), nil)
	}
}

func (h GroupHandler) Update(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.PathParam(h.svc.GetID())
		group, _ := h.GetGroup(id)
		if ctx.Request().Method == http.MethodGet {
			sports, err := h.GetSports(context)
			if err != nil {
				xlog.Error("error while getting sports", "error", err)
			}
			return view.Render(ctx, http.StatusOK, GroupFormView(
				view.NewViewData(group, errorsmap.New()), sports,
				templ.Attributes{"target": "#content", "hx-patch": reverse(ctx, "group.update", id)}), nil)
		}
		req := service.Request{"id": id}
		if err := ctx.Bind(&req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		_, err := h.svc.Update(context, req)
		if err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		return ctx.Redirect(http.StatusSeeOther, view.ReverseX(ctx, "account.get", xsession.GetUser(ctx.Request().Context()).ID))
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
		return ctx.Redirect(http.StatusSeeOther, view.ReverseX(ctx, "account.get", xsession.GetUser(ctx.Request().Context()).ID))

	}
}

var reverse = func(c echo.Context, name string, values ...any) string {
	path, err := c.Echo().Router().Routes().Reverse(name, values...)
	if err != nil {
		xlog.Error("error while reversing route", "name", name, "values", values, "error", err)
	}
	return path
}
