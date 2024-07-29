package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/josuebrunel/sportdropin/pkg/errorsmap"
	"github.com/josuebrunel/sportdropin/pkg/view"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type (
	Record      = *models.Record
	RecordSlice = []*models.Record
	Request     = map[string]any
	Requests    = []Request
	BackRel     = map[string]map[string]any
	Filters     = map[string]any
)

type Service struct {
	Name string
	ID   string
	db   *daos.Dao
}

func (s Service) GetID() string { return s.ID }

func NewService(name, id string, db *daos.Dao) Service {
	return Service{
		Name: name,
		ID:   id,
		db:   db,
	}
}

func (s Service) GetCollection() *models.Collection {
	c, err := s.db.FindCollectionByNameOrId(s.Name)
	if err != nil {
		xlog.Error("failed to find collection", "collection", s.Name, "error", err)
		return nil
	}
	return c
}

func (s Service) GetNewRecord() *models.Record {
	return models.NewRecord(s.GetCollection())
}

func (s Service) GetNewRecords(size int) RecordSlice {
	return make([]*models.Record, size)
}

func (s Service) create(ctx context.Context, req Request) (Record, error) {
	collection, err := s.db.FindCollectionByNameOrId(s.Name)
	if err != nil {
		xlog.Error("error while finding collection", "collection", s.Name, "error", err)
		return s.GetNewRecord(), err
	}
	record := models.NewRecord(collection)
	for k, v := range req {
		record.Set(k, v)
	}
	if err = s.db.SaveRecord(record); err != nil {
		xlog.Error("error while inserting", "record", record, "error", err)
		return s.GetNewRecord(), err
	}
	return record, nil
}

func (s Service) Create(ctx context.Context, req Request) (view.ViewData[Record], error) {
	em := errorsmap.New()
	record, err := s.create(ctx, req)
	em["error"] = err
	if err != nil {
		return view.NewViewData(s.GetNewRecord(), em), err
	}
	return view.NewViewData(record, em), nil
}

func (s Service) BulkCreate(ctx context.Context, req []Request) (view.ViewData[RecordSlice], error) {
	em := errorsmap.New()
	records := RecordSlice{}
	for i, r := range req {
		record, err := s.create(ctx, r)
		em["error"] = err
		if err != nil {
			em[fmt.Sprintf("%d", i)] = err
			continue
		}
		records = append(records, record)
	}
	return view.NewViewData(records, em), nil
}

func (s Service) getByID(ctx context.Context, id string) (Record, error) {
	record, err := s.db.FindRecordById(s.Name, id)
	if err != nil {
		xlog.Error("error while getting", "record", id, "error", err)
		return s.GetNewRecord(), err
	}
	return record, nil
}

func (s Service) GetByID(ctx context.Context, id string, expand ...string) (view.ViewData[Record], error) {
	em := errorsmap.New()
	record, err := s.getByID(ctx, id)
	em["error"] = err
	if len(expand) > 0 {
		s.db.ExpandRecord(record, expand, nil)
	}
	xlog.Debug("record", "record", record)
	return view.NewViewData(record, em), err
}

func (s Service) GetByIDWithBackRel(ctx context.Context, id string, expand BackRel) (view.ViewData[Record], error) {
	em := errorsmap.New()
	record, err := s.getByID(ctx, id)
	em["error"] = err
	rels := map[string]any{}
	for k, v := range expand {
		rels[k] = s.FindRecordsByExpr(ctx, k, dbx.HashExp(v))
	}
	record.SetExpand(rels)
	xlog.Debug("record", "record", record)
	return view.NewViewData(record, em), err
}

func (s Service) GetByData(ctx context.Context, key string, value any) (view.ViewData[Record], error) {
	em := errorsmap.New()
	record, err := s.db.FindFirstRecordByData(s.Name, key, value)
	if err != nil {
		xlog.Error("error while getting record", "key", key, "value", value, "error", err)
		em["error"] = err
		return view.NewViewData(s.GetNewRecord(), em), err

	}
	return view.NewViewData(record, em), nil
}

func (s Service) list(ctx context.Context, filters map[string]any) (RecordSlice, error) {
	hashExp := dbx.HashExp{}
	for k, v := range filters {
		hashExp[k] = v
	}

	records, err := s.db.FindRecordsByExpr(s.Name, dbx.NewExp("1={:one}", dbx.Params{"one": 1}), hashExp)
	if err != nil {
		xlog.Error("error while listing records", "error", err)
		return records, err
	}
	return records, nil
}

func (s Service) List(ctx context.Context, filters map[string]any, expand ...string) (view.ViewData[RecordSlice], error) {
	em := errorsmap.New()

	records, err := s.list(ctx, filters)
	if err != nil {
		xlog.Error("error while listing records", "error", err)
		em["error"] = err
		return view.NewViewData[RecordSlice](nil, em), err
	}
	if len(expand) > 0 {
		s.db.ExpandRecords(records, expand, nil)
	}
	return view.NewViewData(records, em), nil
}

func (s Service) ListWithBackRel(ctx context.Context, filters map[string]any, expand BackRel) (view.ViewData[RecordSlice], error) {
	em := errorsmap.New()
	records, err := s.list(ctx, filters)
	if err != nil {
		xlog.Error("error while listing records", "error", err)
		em["error"] = err
		return view.NewViewData[RecordSlice](records, em), err
	}
	for _, record := range records {
		rels := map[string]any{}
		for rel, ff := range expand {
			exp := dbx.HashExp{}
			for k, v := range ff {
				if strings.HasPrefix(v.(string), ":") {
					exp[k] = record.GetId()
				} else {
					exp[k] = v
				}
			}
			rels[rel] = s.FindRecordsByExpr(ctx, rel, exp)
		}
		record.SetExpand(rels)
	}
	xlog.Debug("records list", "records", records)
	return view.NewViewData(records, em), nil
}

func (s Service) upsert(ctx context.Context, req Request) (Record, error) {
	record, err := s.db.FindRecordById(s.Name, req["id"].(string))
	if err != nil {
		xlog.Error("error while getting", "record", req["id"], "error", err)
		record = s.GetNewRecord()
	}

	record.Load(req)
	if err = s.db.SaveRecord(record); err != nil {
		xlog.Error("error while updating", "record", record, "error", err)
		return record, err
	}
	return record, nil
}

func (s Service) Upsert(ctx context.Context, req Request) (view.ViewData[Record], error) {
	em := errorsmap.New()
	record, err := s.upsert(ctx, req)
	if err != nil {
		em["error"] = err
	}
	return view.NewViewData(record, em), err
}

func (s Service) BulkUpsert(ctx context.Context, reqs Requests) (view.ViewData[RecordSlice], error) {
	em := errorsmap.New()
	records := RecordSlice{}
	for i, r := range reqs {
		record, err := s.upsert(ctx, r)
		em["error"] = err
		if err != nil {
			em[fmt.Sprintf("%d", i)] = err
			continue
		}
		records = append(records, record)
	}
	return view.NewViewData(records, em), nil
}

func (s Service) Update(ctx context.Context, req Request) (view.ViewData[Record], error) {
	em := errorsmap.New()

	record, err := s.db.FindFirstRecordByData(s.Name, "id", req[s.ID])
	if err != nil {
		xlog.Error("error while getting", "record", req[s.ID], "error", err)
		em["error"] = err
		return view.NewViewData(s.GetNewRecord(), em), err
	}

	for k, v := range req {
		if k == s.ID {
			continue
		}
		record.Set(k, v)
	}

	if err = s.db.SaveRecord(record); err != nil {
		xlog.Error("error while updating", "record", record, "error", err)
		em["error"] = err
		return view.NewViewData(s.GetNewRecord(), em), err
	}

	return view.NewViewData(record, em), err
}

func (s Service) Delete(ctx context.Context, id string) error {
	em := errorsmap.New()

	record, err := s.db.FindFirstRecordByData(s.Name, "id", id)
	if err != nil {
		xlog.Error("error while getting", "record", id, "error", err)
		em["error"] = err
		return err
	}

	if err = s.db.DeleteRecord(record); err != nil {
		xlog.Error("error while deleting", "record", record, "error", err)
		em["error"] = err
		return err
	}

	return nil
}

func (s Service) FindRecordsByExpr(ctx context.Context, name string, filters dbx.HashExp) RecordSlice {
	var records = RecordSlice{}
	records, err := s.db.FindRecordsByExpr(name, dbx.NewExp("1={:one}", dbx.Params{"one": 1}), filters)
	if err != nil {
		xlog.Error("error while finding records", "error", err)
		return records
	}
	return records
}

func UnmarshalTo(r Record, m any) error {
	b, err := r.MarshalJSON()
	if err != nil {
		xlog.Error("failed to marshal json", "collection", r.Collection(), "id", r.GetId())
		return err
	}
	if err = json.Unmarshal(b, m); err != nil {
		xlog.Error("failed to unmarshal record to target", "collection", r.Collection(), "id", r.GetId())
	}
	return nil
}
