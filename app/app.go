package app

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/josuebrunel/sportdropin/app/config"
	"github.com/josuebrunel/sportdropin/group"
	_ "github.com/josuebrunel/sportdropin/migrations"
	"github.com/josuebrunel/sportdropin/pkg/view"
	"github.com/josuebrunel/sportdropin/pkg/view/base"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

type App struct {
	Opts config.Config
}

func NewApp() App {
	opts := config.NewConfig()
	return App{Opts: opts}
}

func (a App) Run() {
	// pocket base app
	app := pocketbase.New()
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		ctx := app.RootCmd.Context()
		groupHandler := group.NewGroupHandler(app.Dao())

		e.Router.Use(middleware.Logger())
		e.Router.Use(middleware.CORS())
		e.Router.Use(middleware.Recover())

		e.Router.Static("/static", "static")
		e.Router.GET("/", func(c echo.Context) error { return view.Render(c, http.StatusOK, base.Index(), nil) })
		g := e.Router.Group("/group")
		g.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
			TokenLookup: "form:csrf,header:csrf",
		}))
		g.AddRoute(echo.Route{
			Method:  http.MethodGet,
			Path:    "",
			Handler: groupHandler.List(ctx),
			Name:    "group.list",
		})
		g.AddRoute(echo.Route{
			Method:  http.MethodGet,
			Path:    "/:uuid",
			Handler: groupHandler.Get(ctx),
			Name:    "group.get",
		})
		g.AddRoute(echo.Route{
			Method:  http.MethodGet,
			Path:    "/create",
			Handler: groupHandler.Create(ctx),
			Name:    "group.create",
		})
		g.AddRoute(echo.Route{
			Method:  http.MethodPost,
			Path:    "/create",
			Handler: groupHandler.Create(ctx),
			Name:    "group.created",
		})
		g.AddRoute(echo.Route{
			Method:  http.MethodGet,
			Path:    "/:uuid/edit",
			Handler: groupHandler.Update(ctx),
			Name:    "group.update",
		})
		g.AddRoute(echo.Route{
			Method:  http.MethodPatch,
			Path:    "/:uuid/edit",
			Handler: groupHandler.Update(ctx),
			Name:    "group.update",
		})
		g.AddRoute(echo.Route{
			Method:  http.MethodDelete,
			Path:    "/:uuid",
			Handler: groupHandler.Delete(ctx),
			Name:    "group.delete",
		})
		return nil
	})

	// add migration command
	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Admin UI
		// (the isGoRun check is to enable it only during development)
		Automigrate: isGoRun,
	})

	if err := app.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatal("shutting down the server")
	}
}
