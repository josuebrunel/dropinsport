package group

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/josuebrunel/sportdropin/pkg/collection"
	"github.com/josuebrunel/sportdropin/pkg/service"
	"github.com/josuebrunel/sportdropin/pkg/util"
	"github.com/josuebrunel/sportdropin/pkg/view"
	"github.com/josuebrunel/sportdropin/pkg/view/component"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/labstack/echo/v5"
)

func memberStatsToData(schema SportStatSchema, rr service.RecordSlice) []map[string]string {
	data := collection.Transform(rr, func(r service.Record) map[string]string {
		d := map[string]string{}
		d["id"] = r.GetId()
		d["username"] = r.GetString("username")
		stat := r.ExpandedOne("memberstats")
		xlog.Debug("member stats", "stats", stat)
		if stat != nil {
			var stats = make(map[string]string)
			d["stats_id"] = stat.GetId()
			stat.UnmarshalJSONField("stats", &stats)
			for _, k := range schema {
				d[k["abbr"]] = stats[k["abbr"]]
			}
		}
		return d
	})
	return data
}

func formDataToRequests(groupID, seasonID string, formData map[string]any, schema SportStatSchema) service.Requests {
	requests := map[string]service.Request{}
	stats := map[string]map[string]string{}
	for field, value := range formData {
		sf := strings.Split(field, ":")
		if len(sf) > 1 {
			id, fname := sf[0], sf[1]
			// process stat fields
			if collection.Exists(schema, func(s map[string]string) bool { return strings.EqualFold(s["abbr"], fname) }) {
				if v, ok := stats[id]; ok {
					v[fname] = util.F64Fmt(util.AssertType[float64](value), "%.f")
				} else {
					stats[id] = map[string]string{fname: util.F64Fmt(util.AssertType[float64](value), "%.f")}
				}
				// process request fields
			} else {
				if m, ok := requests[id]; ok {
					m[fname] = value
					requests[id] = m
				} else {
					requests[id] = service.Request{fname: value}
				}
			}
		}
	}
	r := service.Requests{}
	for i, v := range requests {
		v["member"] = i
		v["group"] = groupID
		v["season"] = seasonID
		if s, ok := stats[i]; ok {
			js, err := json.Marshal(&s)
			if err != nil {
				xlog.Error("error while marshalling stats", "member", i, "stats", stats[i])
			}
			v["stats"] = js
		}
		r = append(r, v)
	}
	xlog.Debug("requests-data", "requests", r)
	return r
}

func (h GroupHandler) StatCreate(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		groupID := ctx.PathParam(h.svc.GetID())
		schema := h.GetGroupSportStatSchema(context, groupID)
		curSeason, err := h.GetGroupCurrentSeason(context, groupID)
		if err != nil {
			xlog.Error("error while getting current season", "group", groupID, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		if ctx.Request().Method == http.MethodGet {
			members, err := memberSVC.ListWithBackRel(
				context, service.Filters{"group": groupID},
				service.BackRel{
					"memberstats": map[string]any{"member": ":id", "group": groupID, "season": curSeason.GetId()},
				},
			)
			if err != nil {
				xlog.Error("error while getting members and stats", "group", groupID, "error", err)
				return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
			}
			data := memberStatsToData(schema, members.V())
			xlog.Debug("members stats", "stats", data)
			return view.Render(ctx, http.StatusOK,
				GroupStatForm(
					schema, data,
					templ.Attributes{"hx-post": ctx.RouteInfo().Reverse(groupID), "hx-target": "#content"}),
				nil)
		}
		req := service.Request{}
		if err := ctx.Bind(&req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}

		group, _ := h.GetGroup(groupID)
		reqs := formDataToRequests(groupID, curSeason.GetId(), req, schema)
		xlog.Debug("request data", "requests", reqs)
		_, err = statSVC.BulkUpsert(context, reqs)
		if err != nil {
			xlog.Error("error while creating stat", "reqs", reqs, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		members, err := memberSVC.ListWithBackRel(
			context, service.Filters{"group": groupID},
			service.BackRel{
				"memberstats": map[string]any{"member": ":id", "group": groupID, "season": curSeason.GetId()},
			},
		)
		if err != nil {
			xlog.Error("error while getting members and stats", "group", groupID, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		data := memberStatsToData(schema, members.V())
		xlog.Debug("members stats", "stats", data)
		return view.Render(ctx, http.StatusOK, GroupStatList(group, data, schema), nil)
	}
}

func (h GroupHandler) StatList(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		groupID := ctx.PathParam(h.svc.GetID())
		group, _ := h.GetGroup(groupID)
		curSeason, err := h.GetGroupCurrentSeason(context, groupID)
		if err != nil {
			xlog.Error("error while getting current season", "group", groupID, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		members, err := memberSVC.ListWithBackRel(
			context, service.Filters{"group": groupID},
			service.BackRel{
				"memberstats": map[string]any{"member": ":id", "group": groupID, "season": curSeason.GetId()},
			},
		)
		if err != nil {
			xlog.Error("error while getting members and stats", "group", groupID, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		schema := h.GetGroupSportStatSchema(context, groupID)
		xlog.Debug("members and stats", "members", members, "schema", schema, "group")
		data := memberStatsToData(schema, members.V())
		xlog.Debug("stats", "stats", data)
		return view.Render(ctx, http.StatusOK, GroupStatList(group, data, schema), nil)
	}
}
