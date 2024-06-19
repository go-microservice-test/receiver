package migrations

import (
	"go-test/db-utils/models"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"slices"
	"sync"
)

func MigrateAllTables(db *gorm.DB) error {
	return MigrateAnimals(db)
}

func MigrateAnimals(db *gorm.DB) error {
	// define as a transaction block
	return db.Transaction(func(tx *gorm.DB) error {
		// apply manual changes for every condition
		// no table
		if !tx.Migrator().HasTable(&models.Animal{}) {
			if err := tx.Migrator().CreateTable(&models.Animal{}); err != nil {
				return err
			}
			// no additional checks required
			return nil
		}
		// get column names in existing table
		columns, err := db.Migrator().ColumnTypes(&models.Animal{})
		if err != nil {
			return err
		}
		// get columns names in the gorm schema
		s, err := schema.Parse(&models.Animal{}, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			panic("failed to parse schema")
		}
		var schemaColumns []string
		for _, field := range s.Fields {
			schemaColumns = append(schemaColumns, field.DBName)
		}
		// add missing columns
		for _, column := range schemaColumns {
			if !tx.Migrator().HasColumn(&models.Animal{}, column) {
				tx.Migrator().AddColumn(&models.Animal{}, column)
			}
		}
		// remove redundant columns
		for _, column := range columns {
			if !slices.Contains(schemaColumns, column.Name()) {
				tx.Migrator().DropColumn(&models.Animal{}, column.Name())
			}
		}

		// table migrated
		return nil
	})
}
