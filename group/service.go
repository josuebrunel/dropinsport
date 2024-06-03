package group

import (
	"context"
	"errors"

	"github.com/google/uuid"
	generic "github.com/josuebrunel/sportdropin/pkg/echogeneric"
	"github.com/josuebrunel/sportdropin/pkg/errorsmap"
	"github.com/josuebrunel/sportdropin/pkg/storage"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
)

func isStrNil(s *string) bool {
	if s == nil || *s == "" {
		return true
	}
	return false
}

type Request struct {
	Group
}

func (r Request) Valid() errorsmap.EMap {
	var em = errorsmap.New()
	if isStrNil(r.Group.Name) {
		em["name"] = errors.New("<name> is required")
	}
	if isStrNil(r.Group.Sport) {
		em["sport"] = errors.New("<sport> is required")
	}
	if isStrNil(r.Group.Street) {
		em["street"] = errors.New("<street> is required")
	}
	if isStrNil(r.Group.City) {
		em["city"] = errors.New("<city> is required")
	}
	if isStrNil(r.Group.Country) {
		em["country"] = errors.New("<country> is required")
	}
	return em
}

func (r Request) GetID() string { return r.UUID.String() }
func (r *Request) SetID(id string) error {
	v, err := uuid.Parse(id)
	if err != nil {
		xlog.Error("group", "set-id", err)
		return err
	}
	r.Group.UUID = v
	return err
}

type Response struct {
	StatusCode int            `json:"status_code"`
	Errors     errorsmap.EMap `json:"errors"`
	Data       any            `json:"data,omitempty"`
}

func (r Response) GetStatusCode() int {
	return r.StatusCode
}

func (r Response) One() Group {
	if r.Data == nil {
		return Group{}
	}
	return r.Data.(Group)
}

func (r Response) All() []Group {
	if r.Data == nil || r.Data.([]Group) == nil {
		return []Group{}
	}
	return r.Data.([]Group)
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
			Errors:     errorsmap.New(),
			Data:       Group{},
		}
	)
	if em := r.Valid(); !em.Nil() {
		xlog.Error("error while validating request", "group", r.Group, "errors", em)
		resp.Errors = em
		resp.StatusCode = 400
		return resp, em
	}
	if _, err = s.store.Create(&r.Group); err != nil {
		xlog.Error("error while creating", "group", r.Group, "error", err)
		resp.Errors["error"] = err
		resp.StatusCode = 500
	} else {
		xlog.Info("group", "created", r)
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
			Errors:     errorsmap.New(),
			Data:       g,
		}
		filter = map[string]any{"uuid": r.GetID()}
	)

	if _, err = s.store.Get(&g, filter); err != nil {
		xlog.Error("error while getting", "group", r.Group.UUID, "error", err)
		if errors.Is(err, storage.ErrNotFound) {
			resp.StatusCode = 404
			resp.Errors["error"] = err
		} else {
			resp.StatusCode = 500
			resp.Errors["error"] = err
		}
	}
	resp.Data = g
	xlog.Info("storage", "get-group", resp.Data)
	return resp, err
}

func (s Service) List(ctx context.Context, filters map[string]any) (generic.IResponse, error) {
	var (
		err    error
		groups []Group
		resp   = Response{
			StatusCode: 200,
			Errors:     errorsmap.New(),
			Data:       groups,
		}
	)

	if _, err = s.store.List(&resp.Data, filters); err != nil {
		resp.StatusCode = 500
		resp.Errors["error"] = err
	}
	return resp, err

}

func (s Service) Update(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	r := req.(*Request)
	var (
		err  error
		resp = Response{
			StatusCode: 202,
			Errors:     errorsmap.New(),
			Data:       r.Group,
		}
	)

	if _, err := s.store.Update(&r.Group); err != nil {
		xlog.Error("error while updating", "group", r.Group.UUID, "error", err)
		if errors.Is(err, storage.ErrNotFound) {
			resp.StatusCode = 404
			resp.Errors["error"] = err
		} else {
			resp.StatusCode = 500
			resp.Errors["error"] = err
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
		xlog.Error("error while deleting", "group", r.Group.UUID, "error", err)
		if errors.Is(err, storage.ErrNotFound) {
			resp.StatusCode = 404
			resp.Errors["error"] = err
		} else {
			resp.StatusCode = 500
			resp.Errors["error"] = err
		}
	}
	return resp, err
}

func NewService(name, id string, store storage.Storer) Service {
	return Service{Name: name, ID: id, store: store}
}

var _ generic.Service = Service{}
