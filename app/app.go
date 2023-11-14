package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/josuebrunel/sportdropin/app/config"
	"github.com/josuebrunel/sportdropin/group"
	"github.com/josuebrunel/sportdropin/pkg/templatemap"
	"github.com/josuebrunel/sportdropin/storage"
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

type TemplateMapWrapper struct {
	templateMap *templatemap.TemplateMap
}

func (t TemplateMapWrapper) Render(wr io.Writer, name string, data any, ctx echo.Context) error {
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = ctx.Echo().Reverse
	}
	return t.templateMap.Render(wr, name, data)
}

func (a App) Run() {
	// Mount storage
	store, err := storage.NewStore(a.Opts.GetDBDSN())
	if err != nil {
		slog.Error("error while initializing storage", "err", err)
		return
	}
	// Setup
	e := echo.New()
	e.Pre(middleware.AddTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	tpl, err := templatemap.NewTemplateMap("templates/layouts/*.html", "templates/pages/*.html")
	if err != nil {
		slog.Error("templatemap", "error", err)
		return
	}
	renderer := TemplateMapWrapper{templateMap: tpl}
	e.Renderer = renderer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// Mount handlers
	groupSVC := group.NewService("group", "uuid", store)

	groupHandler := group.NewGroupHandler(store)

	e.Static("/static/", "static")
	e.GET("/", groupHandler.List(ctx))
	e.POST("/group/", groupHandler.Create(ctx))
	e.GET("/group/", groupHandler.Get(ctx))
	e.GET("/group/:uuid/", groupHandler.Get(ctx))
	e.GET("/group/:uuid/delete/", groupHandler.Delete(ctx))

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
