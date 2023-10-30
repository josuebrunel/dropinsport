package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/josuebrunel/sportdropin/app/config"
	"github.com/josuebrunel/sportdropin/group"
	generic "github.com/josuebrunel/sportdropin/pkg/echogeneric"
	"github.com/josuebrunel/sportdropin/storage"
	"github.com/josuebrunel/sportdropin/user"
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
		slog.Error("error while initializing storage", "err", err)
		return
	}
	// Setup
	e := echo.New()
	e.Pre(middleware.AddTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	// Mount handlers
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})
	groupSVC := group.NewService("group", "uuid", store)
	generic.MountService(e, groupSVC)
	userSVC := user.NewService("user", "uuid")
	generic.MountService(e, userSVC)

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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
