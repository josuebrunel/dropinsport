package group

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	generic "github.com/josuebrunel/sportdropin/pkg/echogeneric"
	"github.com/josuebrunel/sportdropin/storage"
)

type Request struct {
	Group
}

func (r Request) GetID() string { return r.UUID.String() }
func (r *Request) SetID(id string) error {
	v, err := uuid.Parse(id)
	if err != nil {
		slog.Error("group", "set-id", err)
		return err
	}
	r.Group.UUID = &v
	return err
}

type Response struct {
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
	Data       any    `json:"data,omitempty"`
}

func (r Response) GetStatusCode() int {
	return r.StatusCode
}

type Service struct {
	Name  string
	ID    string
	store storage.Storer
}

func (s Service) GetName() string {
	return s.Name
}

func (s Service) GetID() string {
	return s.ID
}

func (s Service) GetRequest() generic.IRequest {
	return &Request{}
}

func (s Service) GetResponse() generic.IResponse {
	return Response{}
}

func (s Service) GetModel() any {
	return Group{}
}

func (s Service) Create(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	r := req.(*Request)
	var (
		err  error
		resp = Response{
			StatusCode: 201,
			Error:      "",
			Data:       Group{},
		}
	)
	if _, err = s.store.Create(&r.Group); err != nil {
		slog.Error("error while creating", "group", r.Group, "error", err)
		resp.Error = err.Error()
		resp.StatusCode = 500
	} else {
		slog.Info("group", "created", r)
		resp.Data = r.Group
	}
	return resp, err
}

func (s Service) Get(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	r := req.(*Request)
	var (
		err  error
		g    = Group{}
		resp = Response{
			StatusCode: 200,
			Error:      "",
			Data:       &g,
		}
		filter = map[string]any{"uuid": r.GetID()}
	)

	if _, err = s.store.Get(&g, filter); err != nil {
		slog.Error("error while getting", "group", r.Group.UUID, "error", err)
		if errors.Is(err, storage.ErrNotFound) {
			resp.StatusCode = 404
			resp.Error = err.Error()
		} else {
			resp.StatusCode = 500
			resp.Error = err.Error()
		}
	}
	slog.Info("storage", "get-group", resp.Data)
	return resp, err
}

func (s Service) List(ctx context.Context, filters map[string]any) (generic.IResponse, error) {
	var (
		err    error
		groups []Group
		resp   = Response{
			StatusCode: 200,
			Error:      "",
			Data:       groups,
		}
	)

	if _, err = s.store.List(&resp.Data, filters); err != nil {
		resp.StatusCode = 500
		resp.Error = err.Error()
	}
	return resp, err

}

func (s Service) Update(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	r := req.(*Request)
	var (
		err  error
		resp = Response{
			StatusCode: 202,
			Error:      "",
			Data:       r.Group,
		}
	)

	if _, err := s.store.Update(&r.Group); err != nil {
		slog.Error("error while updating", "group", r.Group.UUID, "error", err)
		if errors.Is(err, storage.ErrNotFound) {
			resp.StatusCode = 404
			resp.Error = err.Error()
		} else {
			resp.StatusCode = 500
			resp.Error = err.Error()
		}
	}
	return resp, err
}

func (s Service) Delete(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	var (
		r      = req.(*Request)
		filter = map[string]any{"uuid": r.UUID.String()}
		resp   = Response{StatusCode: 204}
		err    error
	)
	if _, err := s.store.Delete(&r.Group, filter); err != nil {
		slog.Error("error while deleting", "group", r.Group.UUID, "error", err)
		if errors.Is(err, storage.ErrNotFound) {
			resp.StatusCode = 404
			resp.Error = err.Error()
		} else {
			resp.StatusCode = 500
			resp.Error = err.Error()
		}
	}
	return resp, err
}

func NewService(name, id string, store storage.Storer) Service {
	return Service{Name: name, ID: id, store: store}
}

var _ generic.Service = Service{}
