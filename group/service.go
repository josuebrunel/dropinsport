package group

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	generic "github.com/josuebrunel/sportdropin/pkg/echogeneric"
)

var store = []*Group{}

type Group struct {
	UUID        uuid.UUID `json:"uuid"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
}

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
	r.Group.UUID = v
	return err
}

type Response struct {
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
	Data       Group  `json:"data,ommitempty"`
}

func (r Response) GetStatusCode() int {
	return r.StatusCode
}

type Service struct {
	Name string
	ID   string
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

func (s Service) Create(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	r := req.(*Request)
	r.UUID = uuid.New()
	store = append(store, &r.Group)
	slog.Info("group", "create", r, "store", store)
	return Response{StatusCode: 200, Data: *store[len(store)-1]}, nil
}

func (s Service) Get(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	r := req.(*Request)
	slog.Info("group", "get-gen-req", r)
	for _, g := range store {
		if g.UUID.String() == r.UUID.String() {
			slog.Info("group", "get", *g)
			r := Response{StatusCode: 200, Data: *g}
			return r, nil
		}
	}
	err := errors.New("not-found")
	return Response{StatusCode: 404, Error: err.Error()}, err
}

func (s Service) Update(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	r := req.(*Request)
	for _, g := range store {
		if g.UUID.String() == r.UUID.String() {
			slog.Info("group", "update", *g)
			g.Name = r.Name
			g.Description = r.Description
			slog.Info("group", "update", r, "store", store)
			r := Response{StatusCode: 200, Data: *g}
			return r, nil
		}
	}
	err := errors.New("not-found")
	return Response{StatusCode: 404, Error: err.Error()}, err
}

func (s Service) Delete(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	r := req.(*Request)
	for i, g := range store {
		if g.UUID == r.UUID {
			slog.Info("group", "delete", *g)
			last := len(store) - 1
			store[i], store[last] = store[last], store[i]
			store = store[:last]
		}
	}
	slog.Info("group", "delete", r.UUID, "store", store)
	return Response{StatusCode: 204}, nil
}

func NewService(name, id string) Service {
	return Service{Name: name, ID: id}
}

// var _ gh.Service = Service{}
