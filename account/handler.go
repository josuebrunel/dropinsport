package account

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/josuebrunel/sportdropin/pkg/errorsmap"
	"github.com/josuebrunel/sportdropin/pkg/models"
	"github.com/josuebrunel/sportdropin/pkg/pbclient"
	pb "github.com/josuebrunel/sportdropin/pkg/pbclient"
	"github.com/josuebrunel/sportdropin/pkg/view"
	"github.com/josuebrunel/sportdropin/pkg/view/component"
	"github.com/josuebrunel/sportdropin/pkg/xsession"
	"github.com/labstack/echo/v5"
)

type (
	Request          = map[string]any
	UserModel        = pb.UserRecord
	RequestLoginForm struct {
		Username string `json:"username" form:"username"`
		Password string `json:"password" form:"password"`
	}
	RequestRegisterForm struct {
		Email string `json:"email" form:"email"`
		RequestLoginForm
		PasswordConfirm string `json:"passwordConfirm" form:"passwordConfirm"`
	}
)

type AccountHandler struct {
	Collection string
	pathParam  string
	api        *pb.Client
}

func NewAccountHandler(baseURL string) AccountHandler {
	api := pb.New(baseURL)
	return AccountHandler{Collection: "users", pathParam: "accountid", api: &api}
}

func (a AccountHandler) GetToken(cx context.Context) string {
	return xsession.Get[string](cx, xsession.SessionName)
}

func (a AccountHandler) Login(cx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Method == http.MethodGet {
			return view.Render(c, http.StatusOK, LoginFormView(
				errorsmap.New(),
				templ.Attributes{"method": http.MethodPost, "action": view.ReverseX(c, "account.login")}),
				nil,
			)
		}
		var (
			req RequestLoginForm
			err error
		)
		if err = c.Bind(&req); err != nil {
			return view.Render(c, http.StatusOK, component.Error(err.Error()), nil)
		}
		resp, err := a.api.UserAuth(req.Username, req.Password)
		if err != nil {
			em := errorsmap.New()
			data := pb.ResponseTo[pb.ResponseError](resp)
			if strings.EqualFold(data.Message, "") {
				em["error"] = err
			} else {
				em["error"] = errors.New(data.Message)
			}
			return view.Render(c, http.StatusOK, LoginFormView(em,
				templ.Attributes{"method": http.MethodPost, "action": view.ReverseX(c, "account.login")}),
				nil,
			)
		}
		user := pb.ResponseTo[pb.ResponseAuth](resp)
		xsession.Set(c.Request().Context(), xsession.SessionName, user.Token)
		return c.Redirect(http.StatusFound, view.ReverseX(c, "account.get", user.Record.ID))
	}
}

func (a AccountHandler) Get(cx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := a.GetToken(c.Request().Context())
		id := c.PathParam(a.pathParam)
		resp, err := a.api.RecordGet("users", id, pb.QHeaders{"Authorization": token})
		if err != nil {
			return view.Render(c, http.StatusOK, component.Error(err.Error()), nil)
		}
		user := pb.ResponseTo[UserModel](resp)
		return view.Render(c, http.StatusOK, ProfileView(user), nil)
	}
}

func (a AccountHandler) Create(cx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		em := errorsmap.New()
		if c.Request().Method == http.MethodGet {
			return view.Render(c, http.StatusOK, RegisterFormView(
				em,
				templ.Attributes{"method": http.MethodPost, "action": view.ReverseX(c, "account.register")}),
				nil,
			)
		}
		var (
			req RequestRegisterForm
			err error
		)
		if err = c.Bind(&req); err != nil {
			em["error"] = err
			return view.Render(c, http.StatusOK, RegisterFormView(
				em,
				templ.Attributes{"method": http.MethodPost, "action": view.ReverseX(c, "account.register")}),
				nil,
			)
		}
		resp, err := a.api.RecordCreate("users", pb.NewQData(req))
		if err != nil {
			em["error"] = err
			return view.Render(c, http.StatusOK, RegisterFormView(
				em,
				templ.Attributes{"method": http.MethodPost, "action": view.ReverseX(c, "account.register")}),
				nil,
			)

		}
		_ = pbclient.ResponseTo[pbclient.UserRecord](resp)
		return c.Redirect(http.StatusFound, view.ReverseX(c, "account.login"))
	}
}

func (a AccountHandler) Update(cx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			id    = c.PathParam(a.pathParam)
			em    = errorsmap.New()
			token = a.GetToken(c.Request().Context())
			user  = UserModel{}
		)
		resp, err := a.api.RecordGet("users", id, pb.QHeaders{"Authorization": token})
		if err != nil {
			return view.Render(c, http.StatusOK, component.Error(err.Error()), nil)
		}
		user = pb.ResponseTo[UserModel](resp)
		if c.Request().Method == http.MethodGet {
			return view.Render(c, http.StatusOK, EditFormView(user, em,
				templ.Attributes{"hx-patch": view.ReverseX(c, "account.update", id), "hx-target": "#content"}),
				nil,
			)
		}
		var req RequestRegisterForm
		if err = c.Bind(&req); err != nil {
			em["error"] = err
			return view.Render(c, http.StatusOK, EditFormView(user, em,
				templ.Attributes{"hx-patch": view.ReverseX(c, "account.update", id), "hx-target": "#content"}),
				nil,
			)
		}
		_, err = a.api.RecordUpdate("users", id, pb.NewQData(req), pb.QHeaders{"Authorization": token})
		if err != nil {
			em["error"] = err
			return view.Render(c, http.StatusOK, EditFormView(user, em,
				templ.Attributes{"hx-patch": view.ReverseX(c, "account.update", id), "hx-target": "#content"}),
				nil,
			)
		}
		resp, err = a.api.RecordGet("users", id, pb.QHeaders{"Authorization": token})
		if err != nil {
			return view.Render(c, http.StatusOK, component.Error(err.Error()), nil)
		}
		user = pb.ResponseTo[UserModel](resp)
		return view.Render(c, http.StatusOK, ProfileView(user), nil)
	}
}

func (a AccountHandler) Groups(cx context.Context) echo.HandlerFunc {
	return func(c echo.Context) error {
		var (
			id    = c.PathParam(a.pathParam)
			token = a.GetToken(c.Request().Context())
		)
		resp, err := a.api.RecordGet("users", id, pb.QHeaders{"Authorization": token}, pb.QExpand{"groups_via_user.sport"})
		if err != nil {
			return view.Render(c, http.StatusOK, component.Error(err.Error()), nil)
		}
		user := pb.ResponseTo[models.UserExpandGroup](resp)
		return view.Render(c, http.StatusOK, GroupListView(user.Expand.Groups, templ.Attributes{}), nil)
	}
}
