// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/service"
)

const (
	// DriverMysql is a const value of mysql driver.
	DriverMysql = "mysql"
	// DriverSqlite is a const value of sqlite driver.
	DriverSqlite = "sqlite"
	// DriverPostgresql is a const value of postgresql driver.
	DriverPostgresql = "postgresql"
	// DriverMssql is a const value of mssql driver.
	DriverMssql = "mssql"
)

// Connection is a connection handler of database.
// 資料庫連接的處理程序
// 方法建立在資料庫引擎名稱.go中(使用mysql(/modules/db/mysql.go中))
type Connection interface {
	// Query is the query method of sql.
	// 查詢
	Query(query string, args ...interface{}) ([]map[string]interface{}, error)

	// Exec is the exec method of sql.
	// 執行
	Exec(query string, args ...interface{}) (sql.Result, error)

	// QueryWithConnection is the query method with given connection of sql.
	// 查詢(有給定sql連接)
	QueryWithConnection(conn, query string, args ...interface{}) ([]map[string]interface{}, error)

	// ExecWithConnection is the exec method with given connection of sql.
	// 執行(有給定sql連接)
	ExecWithConnection(conn, query string, args ...interface{}) (sql.Result, error)

	QueryWithTx(tx *sql.Tx, query string, args ...interface{}) ([]map[string]interface{}, error)

	ExecWithTx(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error)

	BeginTxWithReadUncommitted() *sql.Tx
	BeginTxWithReadCommitted() *sql.Tx
	BeginTxWithRepeatableRead() *sql.Tx
	BeginTx() *sql.Tx
	BeginTxWithLevel(level sql.IsolationLevel) *sql.Tx

	BeginTxWithReadUncommittedAndConnection(conn string) *sql.Tx
	BeginTxWithReadCommittedAndConnection(conn string) *sql.Tx
	BeginTxWithRepeatableReadAndConnection(conn string) *sql.Tx
	BeginTxAndConnection(conn string) *sql.Tx
	BeginTxWithLevelAndConnection(conn string, level sql.IsolationLevel) *sql.Tx

	// InitDB initialize the database connections.
	// 初始化資料庫連接
	InitDB(cfg map[string]config.Database) Connection

	// GetName get the connection name.
	Name() string

	Close() []error

	// GetDelimiter get the default testDelimiter.
	GetDelimiter() string

	GetDB(key string) *sql.DB
}

// GetConnectionByDriver return the Connection by given driver name.
// 藉由參數(driver = mysql、mssql...)取得Connection(interface)
func GetConnectionByDriver(driver string) Connection {
	switch driver {
	case "mysql":
		return GetMysqlDB()
	case "mssql":
		return GetMssqlDB()
	case "sqlite":
		return GetSqliteDB()
	case "postgresql":
		return GetPostgresqlDB()
	default:
		panic("driver not found!")
	}
}

// 將參數srv轉換為Connect(interface)回傳並回傳
func GetConnectionFromService(srv interface{}) Connection {
	if v, ok := srv.(Connection); ok {
		return v
	}
	panic("wrong service")
}

// service.List類別為map[string]Service，Service是interface(Name方法)
// 取得匹配的service.Service然後轉換成Connection(interface)類別
func GetConnection(srvs service.List) Connection {
	// config.GetDatabases()設置DatabaseList，在modules\config\config.go
	// GetDefault取得預設資料庫DatabaseList["default"]的值
	// srvs.Get透過資料庫driver取的Service(interface)
	if v, ok := srvs.Get(config.GetDatabases().GetDefault().Driver).(Connection); ok {
		return v
	}
	panic("wrong service")
}

// 取得資料庫引擎的Aggregation表達式，將參數值加入表達式
func GetAggregationExpression(driver, field, headField, delimiter string) string {
	switch driver {
	case "postgresql":
		return fmt.Sprintf("string_agg(%s::character varying, '%s') as %s", field, delimiter, headField)
	case "mysql":
		return fmt.Sprintf("group_concat(%s separator '%s') as %s", field, delimiter, headField)
	case "sqlite":
		return fmt.Sprintf("group_concat(%s, '%s') as %s", field, delimiter, headField)
	case "mssql":
		return fmt.Sprintf("string_agg(%s, '%s') as [%s]", field, delimiter, headField)
	default:
		panic("wrong driver")
	}
}

const (
	INSERT = 0
	DELETE = 1
	UPDATE = 2
	QUERY  = 3
)

var ignoreErrors = [...][]string{
	// insert
	{
		"LastInsertId is not supported",
		"There is no generated identity value",
	},
	// delete
	{
		"no affect",
	},
	// update
	{
		"LastInsertId is not supported",
		"There is no generated identity value",
		"no affect",
	},
	// query
	{
		"LastInsertId is not supported",
		"There is no generated identity value",
		"no affect",
		"out of index",
	},
}

//檢查是否有錯誤
func CheckError(err error, t int) bool {
	if err == nil {
		return false
	}
	for _, msg := range ignoreErrors[t] {
		if strings.Contains(err.Error(), msg) {
			return false
		}
	}
	return true
}
