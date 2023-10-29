package generic

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type IRequest interface {
	GetID() string
	SetID(id string) error
}
type IResponse interface {
	GetStatusCode() int
}

type Service interface {
	GetName() string
	GetID() string
	GetRequest() IRequest
	GetResponse() IResponse
	Create(context.Context, IRequest) (IResponse, error)
	Get(context.Context, IRequest) (IResponse, error)
	Update(context.Context, IRequest) (IResponse, error)
	Delete(context.Context, IRequest) (IResponse, error)
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
			return ctx.JSON(resp.GetStatusCode(), resp)
		}
		return ctx.JSON(status, resp)
	}
}

func (s GenericServiceHandler) Get(ctx context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.Param(s.svc.GetID())
		s.e.Logger.Info("generic-get", "param", id)
		req := s.svc.GetRequest()
		if err := req.SetID(id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		resp, err := s.svc.Get(ctx.Request().Context(), req)
		if err != nil {
			return ctx.JSON(resp.GetStatusCode(), resp)
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
		s.e.Logger.Infof("generic-update", "param", id)
		if err = req.SetID(id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		resp, err := s.svc.Update(ctx.Request().Context(), req)
		if err != nil {
			return ctx.JSON(resp.GetStatusCode(), resp)
		}
		return ctx.JSON(http.StatusOK, resp)
	}
}

func (s GenericServiceHandler) Delete(ctx context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.Param(s.svc.GetID())
		s.e.Logger.Infof("generic-delete", "param", id)
		req := s.svc.GetRequest()
		if err := req.SetID(id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		resp, err := s.svc.Delete(ctx.Request().Context(), req)
		if err != nil {
			return ctx.JSON(resp.GetStatusCode(), resp)
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

func MountService(e *echo.Echo, svc Service) GenericServiceHandler {
	ctx := context.Background()
	h := GenericServiceHandler{svc: svc, e: e}
	g := h.e.Group(svc.GetName())
	paramPath := h.GetPathParamName()
	g.POST("/", h.Create(ctx))
	g.GET(paramPath, h.Get(ctx))
	g.PATCH(paramPath, h.Update(ctx))
	g.DELETE(paramPath, h.Delete(ctx))
	return h
}
