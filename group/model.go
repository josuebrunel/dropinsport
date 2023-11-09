package group

import "github.com/josuebrunel/sportdropin/storage"

type Group struct {
	storage.BaseModel
	Name        *string `json:"name" form:"name" gorm:"not null"`
	Sport       *string `json:"sport" form:"sport" gorm:"not null"`
	Description *string `json:"description" form:"description"`
	City        *string `json:"city" form:"city" gorm:"not null"`
	Country     *string `json:"country" form:"country" gorm:"not null"`
	Street      *string `json:"street" form:"street" gorm:"not null"`
}
