package app

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type ServiceRequest interface {
	GetID() string
	SetID(id string)
}
type ServiceResponse interface {
	GetStatusCode() int
}

type Service interface {
	GetName() string
	GetID() string
	GetRequest() ServiceRequest
	GetResponse() ServiceResponse
	Create(context.Context, ServiceRequest) (ServiceResponse, error)
	Get(context.Context, ServiceRequest) (ServiceResponse, error)
	Update(context.Context, ServiceRequest) (ServiceResponse, error)
	Delete(context.Context, ServiceRequest) (ServiceResponse, error)
}

type GenericServiceHandler struct {
	svc Service
	e   *echo.Echo
}

func (s GenericServiceHandler) Create(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var (
			err    error
			status int
			req    = s.svc.GetRequest()
		)
		// try to bing payload
		if err = ctx.Bind(req); err != nil {
			s.e.Logger.Error("failed to bind payalod for %s", s.svc.GetName())
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
		// let the target service process the request
		resp, err := s.svc.Create(context, req)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		return ctx.JSON(status, resp)
	}
}

func (s GenericServiceHandler) Get(ctx context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.Param(s.svc.GetID())
		req := s.svc.GetRequest()
		req.SetID(id)
		resp, err := s.svc.Get(ctx.Request().Context(), req)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		return ctx.JSON(http.StatusOK, resp)
	}
}

func (s GenericServiceHandler) Update(ctx context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var err error
		req := s.svc.GetRequest()
		// try to bing payload
		if err = ctx.Bind(req); err != nil {
			s.e.Logger.Error("failed to bind payalod for %s", s.svc.GetName())
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
		id := ctx.Param(s.svc.GetID())
		req.SetID(id)
		resp, err := s.svc.Update(ctx.Request().Context(), req)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		return ctx.JSON(http.StatusOK, resp)
	}
}

func (s GenericServiceHandler) Delete(ctx context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.Param(s.svc.GetID())
		req := s.svc.GetRequest()
		req.SetID(id)
		resp, err := s.svc.Delete(ctx.Request().Context(), req)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		return ctx.JSON(http.StatusOK, resp)
	}
}

func (s GenericServiceHandler) GetPathParamName() string {
	var sb strings.Builder
	sb.WriteString("/:")
	sb.WriteString(s.svc.GetID())
	sb.WriteString("/")
	return sb.String()
}

func NewGenericServiceHandler(e *echo.Echo, svc Service) GenericServiceHandler {
	ctx := context.Background()
	h := GenericServiceHandler{svc: svc, e: e}
	g := h.e.Group(svc.GetName())
	paramPath := h.GetPathParamName()
	g.POST("/", h.Create(ctx))
	g.GET(paramPath, h.Get(ctx))
	g.PATCH(paramPath, h.Get(ctx))
	g.DELETE(paramPath, h.Get(ctx))
	return h
}
