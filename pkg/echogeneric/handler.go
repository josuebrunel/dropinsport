package generic

import (
	"context"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// IRequest represents the interface for request objects.
type IRequest interface {
	GetID() string         // Get the ID from the request.
	SetID(id string) error // Set the ID for the request.
}

// IResponse represents the interface for response objects.
type IResponse interface {
	GetStatusCode() int // Get the HTTP status code for the response.
}

// Service is an interface representing a generic service with CRUD operations.
type Service interface {
	GetName() string                                     // Get the service name.
	GetID() string                                       // Get the ID field name for the service.
	GetRequest() IRequest                                // Get an instance of the request object.
	GetResponse() IResponse                              // Get an instance of the response object.
	Create(context.Context, IRequest) (IResponse, error) // Create a resource.
	Get(context.Context, IRequest) (IResponse, error)    // Get a resource.
	Update(context.Context, IRequest) (IResponse, error) // Update a resource.
	Delete(context.Context, IRequest) (IResponse, error) // Delete a resource.
}

// GenericServiceHandler is a handler for generic service operations.
type GenericServiceHandler struct {
	svc Service
	e   *echo.Echo
}

// Create is a handler for the create operation.
func (s GenericServiceHandler) Create(context context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var (
			err error
			req = s.svc.GetRequest()
		)
		// Try to bind payload.
		if err = ctx.Bind(req); err != nil {
			s.e.Logger.Error("failed to bind payload for %s", s.svc.GetName())
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
		// Let the target service process the request.
		resp, err := s.svc.Create(context, req)
		if err != nil {
			return ctx.JSON(resp.GetStatusCode(), resp)
		}
		return ctx.JSON(resp.GetStatusCode(), resp)
	}
}

// Get is a handler for the get operation.
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
		return ctx.JSON(resp.GetStatusCode(), resp)
	}
}

// Update is a handler for the update operation.
func (s GenericServiceHandler) Update(ctx context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		var err error
		req := s.svc.GetRequest()
		// Try to bind payload.
		if err = ctx.Bind(req); err != nil {
			s.e.Logger.Error("failed to bind payload for %s", s.svc.GetName())
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
		id := ctx.Param(s.svc.GetID())
		s.e.Logger.Info("generic-update", "param", id)
		if err = req.SetID(id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		resp, err := s.svc.Update(ctx.Request().Context(), req)
		if err != nil {
			return ctx.JSON(resp.GetStatusCode(), resp)
		}
		return ctx.JSON(resp.GetStatusCode(), resp)
	}
}

// Delete is a handler for the delete operation.
func (s GenericServiceHandler) Delete(ctx context.Context) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		id := ctx.Param(s.svc.GetID())
		s.e.Logger.Info("generic-delete", "param", id)
		req := s.svc.GetRequest()
		if err := req.SetID(id); err != nil {
			return ctx.JSON(http.StatusInternalServerError, err.Error())
		}
		resp, err := s.svc.Delete(ctx.Request().Context(), req)
		if err != nil {
			return ctx.JSON(resp.GetStatusCode(), resp)
		}
		return ctx.JSON(resp.GetStatusCode(), resp)
	}
}

// GetPathParamName returns the path parameter name used for routing.
func (s GenericServiceHandler) GetPathParamName() string {
	var sb strings.Builder
	sb.WriteString("/:")
	sb.WriteString(s.svc.GetID())
	sb.WriteString("/")
	return sb.String()
}

// MountService creates and mounts a GenericServiceHandler for the provided service on the given Echo instance.
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
