package database

import "database/sql"

type Entry struct {
	Uid           int64
	Project       string
	Note          sql.NullString
	EntryDatetime string
	Name          sql.NullString // Property Name.
	Value         sql.NullString // Property Value.
}
