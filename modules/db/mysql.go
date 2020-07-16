// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package db

import (
	"database/sql"
	"github.com/GoAdminGroup/go-admin/modules/config"
)

// SQLTx is an in-progress database transaction.
// 正在進行程序
// 程序必須以Commit 或 Rollback方式結束
type SQLTx struct {
	Tx *sql.Tx
}

// Mysql is a Connection of mysql.
// 在base.go中
type Mysql struct {
	Base
}

// GetMysqlDB return the global mysql connection.
// 回傳 global mysql連接
func GetMysqlDB() *Mysql {
	return &Mysql{
		Base: Base{
			DbList: make(map[string]*sql.DB),
		},
	}
}

// Name implements the method Connection.Name.
func (db *Mysql) Name() string {
	return "mysql"
}

// GetDelimiter implements the method Connection.GetDelimiter.
// mysql 使用的分隔符
func (db *Mysql) GetDelimiter() string {
	return "`"
}

// InitDB implements the method Connection.InitDB.
// 初始化資料庫連線並啟動引擎
func (db *Mysql) InitDB(cfgs map[string]config.Database) Connection {
	db.Once.Do(func() {
		for conn, cfg := range cfgs {

			if cfg.Dsn == "" {
				cfg.Dsn = cfg.User + ":" + cfg.Pwd + "@tcp(" + cfg.Host + ":" + cfg.Port + ")/" +
					cfg.Name + cfg.ParamStr()
			}

			sqlDB, err := sql.Open("mysql", cfg.Dsn)

			if err != nil {
				if sqlDB != nil {
					_ = sqlDB.Close()
				}
				panic(err)
			} else {
				// Largest set up the database connection reduce time wait
				sqlDB.SetMaxIdleConns(cfg.MaxIdleCon)
				sqlDB.SetMaxOpenConns(cfg.MaxOpenCon)

				db.DbList[conn] = sqlDB
			}
			//啟動資料庫引擎
			if err := sqlDB.Ping(); err != nil {
				panic(err)
			}
		}
	})
	return db
}

// QueryWithConnection implements the method Connection.QueryWithConnection.
// 有給定參數連接(conn)名稱，透過參數con查詢db.DbList[con]資料並回傳
func (db *Mysql) QueryWithConnection(con string, query string, args ...interface{}) ([]map[string]interface{}, error) {
	// CommonQuery查詢資料並回傳
	return CommonQuery(db.DbList[con], query, args...)
}

// ExecWithConnection implements the method Connection.ExecWithConnection.
// 有給定連接(conn)名稱
func (db *Mysql) ExecWithConnection(con string, query string, args ...interface{}) (sql.Result, error) {
	return CommonExec(db.DbList[con], query, args...)
}

// Query implements the method Connection.Query.
// 沒有給定連接(conn)名稱，透過參數查詢db.DbList["default"]資料並回傳
func (db *Mysql) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	// CommonQuery查詢資料並回傳
	return CommonQuery(db.DbList["default"], query, args...)
}

// Exec implements the method Connection.Exec.
// 沒有給定連接(conn)名稱
func (db *Mysql) Exec(query string, args ...interface{}) (sql.Result, error) {
	return CommonExec(db.DbList["default"], query, args...)
}

// BeginTxWithReadUncommitted starts a transaction with level LevelReadUncommitted.
func (db *Mysql) BeginTxWithReadUncommitted() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelReadUncommitted)
}

// BeginTxWithReadCommitted starts a transaction with level LevelReadCommitted.
func (db *Mysql) BeginTxWithReadCommitted() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelReadCommitted)
}

// BeginTxWithRepeatableRead starts a transaction with level LevelRepeatableRead.
func (db *Mysql) BeginTxWithRepeatableRead() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelRepeatableRead)
}

// BeginTx starts a transaction with level LevelDefault.
func (db *Mysql) BeginTx() *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], sql.LevelDefault)
}

// BeginTxWithLevel starts a transaction with given transaction isolation level.
func (db *Mysql) BeginTxWithLevel(level sql.IsolationLevel) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList["default"], level)
}

// BeginTxWithReadUncommittedAndConnection starts a transaction with level LevelReadUncommitted and connection.
func (db *Mysql) BeginTxWithReadUncommittedAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelReadUncommitted)
}

// BeginTxWithReadCommittedAndConnection starts a transaction with level LevelReadCommitted and connection.
func (db *Mysql) BeginTxWithReadCommittedAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelReadCommitted)
}

// BeginTxWithRepeatableReadAndConnection starts a transaction with level LevelRepeatableRead and connection.
func (db *Mysql) BeginTxWithRepeatableReadAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelRepeatableRead)
}

// BeginTxAndConnection starts a transaction with level LevelDefault and connection.
func (db *Mysql) BeginTxAndConnection(conn string) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], sql.LevelDefault)
}

// BeginTxWithLevelAndConnection starts a transaction with given transaction isolation level and connection.
func (db *Mysql) BeginTxWithLevelAndConnection(conn string, level sql.IsolationLevel) *sql.Tx {
	return CommonBeginTxWithLevel(db.DbList[conn], level)
}

// QueryWithTx is query method within the transaction.
// QueryWithTx是transaction的查詢方法
func (db *Mysql) QueryWithTx(tx *sql.Tx, query string, args ...interface{}) ([]map[string]interface{}, error) {
	// 在performer.go中
	// 與CommonQuery一樣
	return CommonQueryWithTx(tx, query, args...)
}

// ExecWithTx is exec method within the transaction.
// QueryWithTx是transaction的執行方法(與CommonExec一樣)
func (db *Mysql) ExecWithTx(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	// 在performer.go中
	// 與CommonExec一樣(差別在tx執行)
	return CommonExecWithTx(tx, query, args...)
}
