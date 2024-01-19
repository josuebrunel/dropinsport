package storage

import (
	"time"

	"github.com/google/uuid"
	"github.com/josuebrunel/sportdropin/pkg/xlog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ErrNotFound = gorm.ErrRecordNotFound

type BaseModel struct {
	UUID      *uuid.UUID     `json:"uuid,omitempty"`
	CreatedAt *time.Time     `json:"created_at,omitempty"`
	UpdatedAt *time.Time     `json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	u := uuid.New()
	b.UUID = &u
	return nil
}

type Storer interface {
	Create(any) (int64, error)
	Get(any, map[string]any) (int64, error)
	List(any, map[string]any) (int64, error)
	Update(any) (int64, error)
	Delete(any, map[string]any) (int64, error)
}

type Store struct {
	DSN string
	db  *gorm.DB
}

func NewStore(dsn string) (*Store, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	s := Store{DSN: dsn, db: db}
	return &s, err
}

func (s Store) RunMigrations(models ...any) error {
	xlog.Info("running db migrations", "models", models)
	return s.db.AutoMigrate(models...)
}

func (s Store) Create(m any) (int64, error) {
	result := s.db.Create(m)
	if result.Error != nil {
		xlog.Error("storage", "error", result.Error)
	}
	return result.RowsAffected, result.Error
}

func (s Store) Get(m any, filter map[string]any) (int64, error) {
	result := s.db.Where(filter).First(m)
	if result.Error != nil {
		xlog.Error("storage", "error", result.Error)
	}
	return result.RowsAffected, result.Error
}

func (s Store) List(m any, filter map[string]any) (int64, error) {
	result := s.db.Where(filter).Find(m)
	if result.Error != nil {
		xlog.Error("storage", "error", result.Error)
	}
	return result.RowsAffected, result.Error
}

func (s Store) Update(m any) (int64, error) {
	result := s.db.Model(m).Updates(m)
	if result.Error != nil {
		xlog.Error("storage", "error", result.Error)
	}
	return result.RowsAffected, result.Error
}

func (s Store) Delete(m any, filter map[string]any) (int64, error) {
	result := s.db.Where(filter).Delete(m)
	if result.Error != nil {
		xlog.Error("storage", "error", result.Error)
	}
	return result.RowsAffected, result.Error
}
