package group

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/josuebrunel/sportdropin/pkg/collection"
	"github.com/josuebrunel/sportdropin/pkg/models"
	"github.com/josuebrunel/sportdropin/pkg/service"
	"github.com/josuebrunel/sportdropin/pkg/util"
	"github.com/josuebrunel/sportdropin/pkg/view"
	"github.com/josuebrunel/sportdropin/pkg/view/component"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/labstack/echo/v5"
)

func memberStatsToData(sport models.Sport, rr service.RecordSlice) []map[string]string {
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
			for _, k := range sport.Data.Stats {
				d[k.Abbr] = stats[k.Abbr]
			}
		}
		return d
	})
	return data
}

func formDataToRequests(groupID string, formData map[string]any, sport models.Sport) service.Requests {
	requests := map[string]service.Request{}
	stats := map[string]map[string]string{}
	for field, value := range formData {
		sf := strings.Split(field, ":")
		if len(sf) > 1 {
			id, fname := sf[0], sf[1]
			// process stat fields
			if collection.Exists(sport.Data.Stats, func(s models.SportStat) bool { return strings.EqualFold(s.Abbr, fname) }) {
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
		v["season"] = formData["season"]
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
		group, _ := h.GetGroup(groupID)
		sport := h.GetGroupSport(context, groupID)
		seasonID := ctx.QueryParam("season")
		if strings.EqualFold(seasonID, "") {
			curSeason, err := h.GetGroupCurrentSeason(context, groupID)
			if err != nil {
				xlog.Error("error while getting current season", "group", groupID, "error", err)
				return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
			}
			seasonID = curSeason.GetId()
		}
		var err error
		if ctx.Request().Method == http.MethodGet {
			members, err := memberSVC.ListWithBackRel(
				context, service.Filters{"group": groupID},
				service.BackRel{
					"memberstats": map[string]any{"member": ":id", "group": groupID, "season": seasonID},
				},
			)
			if err != nil {
				xlog.Error("error while getting members and stats", "group", groupID, "error", err)
				return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
			}
			var m models.Group
			_ = service.UnmarshalTo(group, &m)
			m.Extra = models.Extra{"curseason": seasonID}

			data := memberStatsToData(sport, members.V())
			sort.Slice(data, func(i, j int) bool {
				return util.F64(data[i][sport.Data.Top.Abbr]) > util.F64(data[j][sport.Data.Top.Abbr])
			})
			xlog.Debug("members stats", "stats", data)
			return view.Render(ctx, http.StatusOK,
				GroupStatForm(
					m, sport, data,
					templ.Attributes{"hx-post": ctx.RouteInfo().Reverse(groupID), "hx-target": "#content"}),
				nil)
		}
		req := service.Request{}
		if err := ctx.Bind(&req); err != nil {
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}

		seasonID = req["season"].(string)
		reqs := formDataToRequests(groupID, req, sport)
		xlog.Debug("request data", "requests", reqs)
		_, err = statSVC.BulkUpsert(context, reqs)
		if err != nil {
			xlog.Error("error while creating stat", "reqs", reqs, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		members, err := memberSVC.ListWithBackRel(
			context, service.Filters{"group": groupID},
			service.BackRel{
				"memberstats": map[string]any{"member": ":id", "group": groupID, "season": seasonID},
			},
		)
		if err != nil {
			xlog.Error("error while getting members and stats", "group", groupID, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		data := memberStatsToData(sport, members.V())
		sort.Slice(data, func(i, j int) bool {
			return util.F64(data[i][sport.Data.Top.Abbr]) > util.F64(data[j][sport.Data.Top.Abbr])
		})
		xlog.Debug("members stats", "stats", data)
		var m models.Group
		_ = service.UnmarshalTo(group, &m)
		m.Extra = models.Extra{"curseason": seasonID}
		return view.Render(ctx, http.StatusOK, GroupStatList(m, data, sport), nil)
	}
}

func (h GroupHandler) StatList(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		groupID := ctx.PathParam(h.svc.GetID())
		group, _ := h.GetGroup(groupID)
		seasonID := ctx.QueryParam("season")
		if strings.EqualFold(seasonID, "") {
			curSeason, err := h.GetGroupCurrentSeason(context, groupID)
			if err != nil {
				xlog.Error("error while getting current season", "group", groupID, "error", err)
				return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
			}
			seasonID = curSeason.GetId()
		}

		members, err := memberSVC.ListWithBackRel(
			context, service.Filters{"group": groupID},
			service.BackRel{
				"memberstats": map[string]any{"member": ":id", "group": groupID, "season": seasonID},
			},
		)
		if err != nil {
			xlog.Error("error while getting members and stats", "group", groupID, "error", err)
			return view.Render(ctx, http.StatusOK, component.Error(err.Error()), nil)
		}
		sport := h.GetGroupSport(context, groupID)
		xlog.Debug("members and stats", "members", members, "sport", sport, "group", group)
		data := memberStatsToData(sport, members.V())
		sort.Slice(data, func(i, j int) bool {
			return util.F64(data[i][sport.Data.Top.Abbr]) > util.F64(data[j][sport.Data.Top.Abbr])
		})
		xlog.Debug("stats", "stats", data)
		var m models.Group
		_ = service.UnmarshalTo(group, &m)
		m.Extra = models.Extra{"curseason": seasonID}
		return view.Render(ctx, http.StatusOK, GroupStatList(m, data, sport), nil)
	}
}
