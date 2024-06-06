package group

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/josuebrunel/sportdropin/pkg/errorsmap"
	"github.com/josuebrunel/sportdropin/pkg/models"
	"github.com/josuebrunel/sportdropin/pkg/util"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Request struct {
	UUID    string `json:"uuid" form:"uuid"`
	Name    string `json:"name" form:"name"`
	Sport   string `json:"sport" form:"sport"`
	Street  string `json:"street" form:"street"`
	City    string `json:"city" form:"city"`
	Country string `json:"country" form:"country"`
}

func (r Request) Validate() error {
	// em := errorsmap.New()
	// if err := validate.Var(r.Name, "required"); err != nil {
	// 	em["name"] = err
	// }
	// if err := validate.Var(r.Sport, "required"); err != nil {
	// 	em["sport"] = err
	// }
	// if err := validate.Var(r.Street, "required"); err != nil {
	// 	em["street"] = err
	// }
	// if err := validate.Var(r.City, "required"); err != nil {
	// 	em["city"] = err
	// }
	// if err := validate.Var(r.Country, "required"); err != nil {
	// 	em["country"] = err
	// }
	return nil
}

type Response struct {
	Errors errorsmap.EMap
	Status int
	Data   any
}

func (r Response) One() models.Group {
	var g models.Group
	if d, ok := r.Data.(*models.Group); ok {
		g = util.Deref(d)
	}
	return g
}

func (r Response) All() models.GroupSlice {
	var gg models.GroupSlice
	if d, ok := r.Data.(models.GroupSlice); ok {
		gg = d
	}
	return gg
}

type Service struct {
	Name string
	ID   string
	db   *sql.DB
}

func (s Service) GetID() string {
	return s.ID
}

func NewService(name, id string, db *sql.DB) Service {
	return Service{
		Name: name,
		ID:   id,
		db:   db,
	}
}

func (s Service) Create(ctx context.Context, req Request) (Response, error) {
	var (
		err  error
		errM = errorsmap.New()
		g    = GroupMFromRequest(req)
		r    = Response{Status: 201, Data: g, Errors: errM}
	)
	if err = req.Validate(); err != nil {
		xlog.Error("error while validating", "group", g, "error", err)
		r.Errors = err.(errorsmap.EMap)
		return r, errM
	}
	g.UUID = uuid.New().String()
	if err = g.Insert(ctx, s.db, boil.Infer()); err != nil {
		xlog.Error("error while creating", "group", g, "error", err)
		r.Errors = err.(errorsmap.EMap)
		r.Status = 400
		return r, errM
	}
	xlog.Info("group", "created", g)
	return r, err
}

func (s Service) Get(ctx context.Context, req Request) (Response, error) {
	var (
		err error
		g   *models.Group
		r   = Response{Status: 200, Data: g, Errors: errorsmap.New()}
	)

	if g, err = models.FindGroup(ctx, s.db, req.UUID); err != nil {
		xlog.Error("error while getting", "group", req.UUID, "error", err)
		r.Errors["error"] = err
		r.Status = 500
		return r, err
	}
	r.Data = g
	return r, err
}

func (s Service) List(ctx context.Context, filters map[string]any) (Response, error) {
	var (
		err      error
		groups   models.GroupSlice
		qFilters []qm.QueryMod
		r        = Response{Status: 200, Errors: errorsmap.New()}
	)

	for k, v := range filters {
		qFilters = append(qFilters, qm.Where(k+"=?", v))
	}

	if groups, err = models.Groups(qFilters...).All(ctx, s.db); err != nil {
		xlog.Error("error while listing", "error", err)
		r.Errors["error"] = err
		return r, err
	}
	r.Data = groups
	return r, err

}

func (s Service) Update(ctx context.Context, req Request) (Response, error) {
	var (
		err error
		g   *models.Group
		r   = Response{Status: 200, Data: g, Errors: errorsmap.New()}
	)
	if err = req.Validate(); err != nil {
		xlog.Error("error while validating", "group", g, "error", err)
		r.Errors = err.(errorsmap.EMap)
		return r, err
	}
	if g, err = models.FindGroup(ctx, s.db, req.UUID); err != nil {
		xlog.Error("error while getting", "group", req.UUID, "error", err)
		return r, err
	}

	g = GroupMFromRequest(req)
	g.UUID = req.UUID
	if _, err = g.Update(ctx, s.db, boil.Infer()); err != nil {
		xlog.Error("error while updating", "group", g, "error", err)
		r.Status = 500
		r.Errors["error"] = err
		return r, err
	}
	r.Data = g
	return r, err
}

func (s Service) Delete(ctx context.Context, req Request) error {
	var (
		err error
		g   *models.Group
	)
	if g, err = models.FindGroup(ctx, s.db, req.UUID); err != nil {
		xlog.Error("error while getting", "group", req.UUID, "error", err)
		return err
	}
	if _, err = g.Delete(ctx, s.db); err != nil {
		xlog.Error("error while deleting", "group", g, "error", err)
		return err
	}
	return err
}

func GroupMFromRequest(req Request) *models.Group {
	return &models.Group{
		Name:    req.Name,
		Sport:   req.Sport,
		Street:  null.StringFrom(req.Street),
		City:    null.StringFrom(req.City),
		Country: null.StringFrom(req.Country),
	}
}
