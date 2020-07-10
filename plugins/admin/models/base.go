package models

import (
	"database/sql"
	"github.com/GoAdminGroup/go-admin/modules/db"
)

// Base is base model structure.
type Base struct {
	TableName string

	Conn db.Connection
	Tx   *sql.Tx
}

// 將參數con(Connection(interface))設置至Base.Conn
func (b Base) SetConn(con db.Connection) Base {
	b.Conn = con
	return b
}

// 藉由給定的table回傳sql(struct)
func (b Base) Table(table string) *db.SQL {
	// Table在modules/db/statement.go中
	// Table藉由給定的table回傳sql(struct)
	// WithDriver藉由給定的conn回傳sql(struct)
	return db.Table(table).WithDriver(b.Conn)
}
