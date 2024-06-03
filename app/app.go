package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/josuebrunel/sportdropin/app/config"
	"github.com/josuebrunel/sportdropin/group"
	"github.com/josuebrunel/sportdropin/pkg/storage"
	"github.com/josuebrunel/sportdropin/pkg/view"
	"github.com/josuebrunel/sportdropin/pkg/view/base"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type App struct {
	Opts config.Config
}

func NewApp() App {
	opts := config.NewConfig()
	return App{Opts: opts}
}

func (a App) Run() {
	// Mount storage
	store, err := storage.NewStore(a.Opts.GetDBDSN())
	if err != nil {
		xlog.Error("error while initializing storage", "err", err)
		return
	}
	// Setup
	e := echo.New()
	e.Pre(middleware.AddTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup: "form:csrf,header:csrf",
	}))
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogError:     true,
		LogRemoteIP:  true,
		LogMethod:    true,
		LogURIPath:   true,
		LogRoutePath: true,
		LogHost:      true,
		LogProtocol:  true,
		HandleError:  true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			xlog.Info("request", "values", values)
			return nil
		},
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// Mount handlers
	groupSVC := group.NewService("group", "uuid", store)

	groupHandler := group.NewGroupHandler(store)

	e.Static("/static", "static")
	e.GET("/", func(c echo.Context) error { return view.Render(c, http.StatusOK, base.Index(), nil) })
	g := e.Group("/group")
	g.GET("/", groupHandler.List(ctx)).Name = "group.list"
	g.POST("/create/", groupHandler.Create(ctx)).Name = "group.create"
	g.GET("/create/", groupHandler.Create(ctx)).Name = "group.create"
	g.GET("/:uuid/", groupHandler.Get(ctx)).Name = "group.get"
	g.PATCH("/:uuid/edit/", groupHandler.Update(ctx)).Name = "group.update"
	g.GET("/:uuid/edit/", groupHandler.Update(ctx)).Name = "group.update"
	g.DELETE("/:uuid/", groupHandler.Delete(ctx)).Name = "group.delete"

	// Migrate models
	models := []any{groupSVC.GetModel()}
	store.RunMigrations(models...)
	// Start server
	go func() {
		if err := e.Start(a.Opts.HTTPAddr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
