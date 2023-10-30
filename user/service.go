package user

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	generic "github.com/josuebrunel/sportdropin/pkg/echogeneric"
)

var store = []*User{}

type User struct {
	UUID      uuid.UUID `json:"uuid"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

type Request struct {
	User
}

func (r Request) GetID() string { return r.UUID.String() }
func (r *Request) SetID(id string) error {
	v, err := uuid.Parse(id)
	if err != nil {
		slog.Error("user", "set-id", err)
		return err
	}
	r.User.UUID = v
	return nil
}

type Response struct {
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
	Data       User   `json:"data,ommitempty"`
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
	r.User.UUID = uuid.New()
	store = append(store, &r.User)
	slog.Info("user", "create", r, "store", store)
	return Response{StatusCode: 200, Data: *store[len(store)-1]}, nil
}

func (s Service) Get(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	r := req.(*Request)
	slog.Info("user", "get-gen-req", r)
	for _, u := range store {
		if u.UUID == r.UUID {
			slog.Info("user", "get", *u)
			r := Response{StatusCode: 200, Data: *u}
			return r, nil
		}
	}
	err := errors.New("not-found")
	return Response{StatusCode: 404, Error: err.Error()}, err
}

func (s Service) Update(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	r := req.(*Request)
	for _, u := range store {
		if u.UUID == r.UUID {
			slog.Info("user", "update", *u)
			u.Username = r.Username
			u.FirstName = r.FirstName
			u.LastName = r.LastName
			u.Email = r.Email
			slog.Info("user", "update", r, "store", store)
			r := Response{StatusCode: 200, Data: *u}
			return r, nil
		}
	}
	err := errors.New("not-found")
	return Response{StatusCode: 404, Error: err.Error()}, err
}

func (s Service) Delete(ctx context.Context, req generic.IRequest) (generic.IResponse, error) {
	r := req.(*Request)
	for i, u := range store {
		if u.UUID == r.UUID {
			slog.Info("user", "delete", *u)
			last := len(store) - 1
			store[i], store[last] = store[last], store[i]
			store = store[:last]
		}
	}
	slog.Info("user", "delete", r.UUID, "store", store)
	return Response{StatusCode: 204}, nil
}

func NewService(name, id string) Service {
	return Service{Name: name, ID: id}
}

var _ generic.Service = Service{}
