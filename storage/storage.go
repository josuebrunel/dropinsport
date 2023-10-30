package storage

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ErrNotFound = gorm.ErrRecordNotFound

type BaseModel struct {
	UUID      uuid.UUID      `json:"uuid,omitempty" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	b.UUID = uuid.New()
	return nil
}

type Storer interface {
	Create(any) (int64, error)
	Get(any, map[string]any) (int64, error)
	Update(any) (int64, error)
	Delete(any, map[string]any) (int64, error)
}

type Store struct {
	DSN string
	db  *gorm.DB
}

func NewStore(dsn string) (Store, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	s := Store{DSN: dsn, db: db}
	return s, err
}

func (s Store) RunMigrations(models ...any) error {
	slog.Info("running db migrations", "models", models)
	return s.db.AutoMigrate(models...)
}

func (s Store) Create(m any) (int64, error) {
	result := s.db.Create(m)
	return result.RowsAffected, result.Error
}

func (s Store) Get(m any, filter map[string]any) (int64, error) {
	result := s.db.Where(filter).First(m)
	return result.RowsAffected, result.Error
}

func (s Store) Update(m any) (int64, error) {
	result := s.db.Model(m).Updates(m)
	return result.RowsAffected, result.Error
}

func (s Store) Delete(m any, filter map[string]any) (int64, error) {
	result := s.db.Where(filter).Delete(m)
	return result.RowsAffected, result.Error
}
