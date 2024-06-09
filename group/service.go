package group

import (
	"context"

	"github.com/google/uuid"
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
)

type Request struct {
	UUID    string `json:"uuid" form:"uuid"`
	Name    string `json:"name" form:"name"`
	Sport   string `json:"sport" form:"sport"`
	Street  string `json:"street" form:"street"`
	City    string `json:"city" form:"city"`
	Country string `json:"country" form:"country"`
}

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

func (s Service) Create(ctx context.Context, req Request) (view.ViewData[Record], error) {
	em := errorsmap.New()

	collection, err := s.db.FindCollectionByNameOrId(s.Name)
	if err != nil {
		xlog.Error("error while finding collection", "collection", s.Name, "error", err)
		em["error"] = err
		return view.NewViewData(&models.Record{}, em), err
	}

	record := models.NewRecord(collection)
	record.Set("uuid", uuid.New().String())
	record.Set("name", req.Name)
	record.Set("sport", req.Sport)
	record.Set("street", req.Street)
	record.Set("city", req.City)
	record.Set("country", req.Country)

	if err = s.db.SaveRecord(record); err != nil {
		xlog.Error("error while inserting", "record", record, "error", err)
		em["error"] = err
		return view.NewViewData(&models.Record{}, em), err
	}

	return view.NewViewData(record, em), nil
}

func (s Service) Get(ctx context.Context, req Request) (view.ViewData[Record], error) {
	em := errorsmap.New()
	record, err := s.db.FindFirstRecordByData(s.Name, "uuid", req.UUID)
	if err != nil {
		xlog.Error("error while getting", "record", req.UUID, "error", err)
		em["error"] = err
		return view.NewViewData(&models.Record{}, em), err
	}
	return view.NewViewData(record, em), nil
}

func (s Service) List(ctx context.Context, filters map[string]any) (view.ViewData[RecordSlice], error) {
	em := errorsmap.New()

	hashExp := dbx.HashExp{}
	for k, v := range filters {
		hashExp[k] = v
	}

	records, err := s.db.FindRecordsByExpr(s.Name, dbx.NewExp("1={:one}", dbx.Params{"one": 1}), hashExp)
	if err != nil {
		xlog.Error("error while listing records", "error", err)
		em["error"] = err
		return view.NewViewData[RecordSlice](nil, em), err
	}

	return view.NewViewData(records, em), nil
}

func (s Service) Update(ctx context.Context, req Request) (view.ViewData[Record], error) {
	em := errorsmap.New()

	record, err := s.db.FindFirstRecordByData(s.Name, "uuid", req.UUID)
	if err != nil {
		xlog.Error("error while getting", "record", req.UUID, "error", err)
		em["error"] = err
		return view.NewViewData(&models.Record{}, em), err
	}

	record.Set("name", req.Name)
	record.Set("sport", req.Sport)
	record.Set("street", req.Street)
	record.Set("city", req.City)
	record.Set("country", req.Country)

	if err = s.db.SaveRecord(record); err != nil {
		xlog.Error("error while updating", "record", record, "error", err)
		em["error"] = err
		return view.NewViewData(&models.Record{}, em), err
	}

	return view.NewViewData(record, em), err
}

func (s Service) Delete(ctx context.Context, req Request) error {
	em := errorsmap.New()

	record, err := s.db.FindFirstRecordByData(s.Name, "uuid", req.UUID)
	if err != nil {
		xlog.Error("error while getting", "record", req.UUID, "error", err)
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
