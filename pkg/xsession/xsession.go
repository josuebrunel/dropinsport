package xsession

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

const SessionName = "xtoken"

type XUser struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	Email string `json:"email"`
}

type SessionConfig struct {
	Skipper        middleware.Skipper
	SessionManager *scs.SessionManager
}

var (
	DefaultSessionConfig = SessionConfig{
		Skipper: middleware.DefaultSkipper,
	}
	SessionManager *scs.SessionManager
)

func init() {
	SessionManager = scs.New()
	SessionManager.Lifetime = 23 * time.Hour
}

func LoadAndSave(sessionManager *scs.SessionManager) echo.MiddlewareFunc {
	c := DefaultSessionConfig
	c.SessionManager = sessionManager

	return LoadAndSaveWithConfig(c)
}

func LoadAndSaveWithConfig(config SessionConfig) echo.MiddlewareFunc {

	if config.Skipper == nil {
		config.Skipper = DefaultSessionConfig.Skipper
	}

	if config.SessionManager == nil {
		panic("Session middleware requires a session manager")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			ctx := c.Request().Context()

			var token string
			cookie, err := c.Cookie(config.SessionManager.Cookie.Name)
			if err == nil {
				token = cookie.Value
			}

			ctx, err = config.SessionManager.Load(ctx, token)
			if err != nil {
				return err
			}

			c.SetRequest(c.Request().WithContext(ctx))

			c.Response().Before(func() {
				if config.SessionManager.Status(ctx) != scs.Unmodified {
					responseCookie := &http.Cookie{
						Name:     config.SessionManager.Cookie.Name,
						Path:     config.SessionManager.Cookie.Path,
						Domain:   config.SessionManager.Cookie.Domain,
						Secure:   config.SessionManager.Cookie.Secure,
						HttpOnly: config.SessionManager.Cookie.HttpOnly,
						SameSite: config.SessionManager.Cookie.SameSite,
					}

					switch config.SessionManager.Status(ctx) {
					case scs.Modified:
						token, _, err := config.SessionManager.Commit(ctx)
						if err != nil {
							panic(err)
						}

						responseCookie.Value = token

					case scs.Destroyed:
						responseCookie.Expires = time.Unix(1, 0)
						responseCookie.MaxAge = -1
					}

					c.SetCookie(responseCookie)
					addHeaderIfMissing(c.Response(), "Cache-Control", `no-cache="Set-Cookie"`)
					addHeaderIfMissing(c.Response(), "Vary", "Cookie")
				}
			})

			return next(c)
		}
	}
}

func addHeaderIfMissing(w http.ResponseWriter, key, value string) {
	for _, h := range w.Header()[key] {
		if h == value {
			return
		}
	}
	w.Header().Add(key, value)
}

func LoginRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess := GetUser(c.Request().Context())
		if strings.EqualFold(sess.Token, "") {
			return c.Redirect(http.StatusFound, "/account/login")
		}
		return next(c)
	}
}

func Get[T any](c context.Context, name string) T {
	var r T
	if v, ok := SessionManager.Get(c, name).(T); ok {
		r = v
	}
	return r
}

func Set(c context.Context, name string, value any) {
	SessionManager.Put(c, name, value)
}

func SetUser(c context.Context, u XUser) {
	b, _ := json.Marshal(u)
	Set(c, SessionName, b)
}

func GetUser(c context.Context) XUser {
	b := Get[[]byte](c, SessionName)
	var u XUser
	json.Unmarshal(b, &u)
	return u
}

func IsAuthenticated(c context.Context) bool {
	return !(GetUser(c) == (XUser{}))
}
