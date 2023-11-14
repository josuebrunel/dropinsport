package group

import "github.com/josuebrunel/sportdropin/storage"

type Group struct {
	storage.BaseModel
	Name        *string `json:"name" form:"name" gorm:"primaryKey;not null"`
	Sport       *string `json:"sport" form:"sport" gorm:"primaryKey;not null"`
	Description *string `json:"description" form:"description"`
	City        *string `json:"city" form:"city" gorm:"primaryKey;not null"`
	Country     *string `json:"country" form:"country" gorm:"primaryKey;not null"`
	Street      *string `json:"street" form:"street" gorm:"not null"`
}
