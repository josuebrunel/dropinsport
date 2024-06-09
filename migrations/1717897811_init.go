package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		group := &models.Collection{
			Name: "groups",
			Schema: schema.NewSchema(
				&schema.SchemaField{
					Name:     "uuid",
					Type:     schema.FieldTypeText,
					Required: true,
					Unique:   true,
				},
				&schema.SchemaField{
					Name:     "name",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name: "description",
					Type: schema.FieldTypeText,
				},
				&schema.SchemaField{
					Name:     "sport",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "street",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "city",
					Type:     schema.FieldTypeText,
					Required: true,
				},
				&schema.SchemaField{
					Name:     "country",
					Type:     schema.FieldTypeText,
					Required: true,
				},
			),
		}
		dao := daos.New(db)
		err := dao.SaveCollection(group)
		return err
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("groups")
		if err != nil {
			return err
		}
		if err := dao.DeleteCollection(collection); err != nil {
			return err
		}
		return nil
	})
}
