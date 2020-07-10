// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package dialect

import (
	"fmt"
	"strings"

	"github.com/GoAdminGroup/go-admin/modules/config"
)

// Dialect is methods set of different driver.
type Dialect interface {
	// GetName get dialect's name
	GetName() string

	// ShowColumns show columns of specified table
	ShowColumns(table string) string

	// ShowTables show tables of database
	ShowTables() string

	// Insert
	Insert(comp *SQLComponent) string

	// Delete
	Delete(comp *SQLComponent) string

	// Update
	Update(comp *SQLComponent) string

	// Select
	Select(comp *SQLComponent) string

	// GetDelimiter return the delimiter of Dialect.
	GetDelimiter() string
}

// GetDialect return the default Dialect.
func GetDialect() Dialect {
	return GetDialectByDriver(config.GetDatabases().GetDefault().Driver)
}

// GetDialectByDriver return the Dialect of given driver.
// 不同資料庫引擎有不同使用符號
func GetDialectByDriver(driver string) Dialect {
	switch driver {
	case "mysql":
		return mysql{
			commonDialect: commonDialect{delimiter: "`"},
		}
	case "mssql":
		return mssql{
			commonDialect: commonDialect{delimiter: "["},
		}
	case "postgresql":
		return postgresql{
			commonDialect: commonDialect{delimiter: `"`},
		}
	case "sqlite":
		return sqlite{
			commonDialect: commonDialect{delimiter: "`"},
		}
	default:
		return commonDialect{delimiter: "`"}
	}
}

// H is a shorthand of map.
// 要新增的資料放置(map[欄位]數值)
type H map[string]interface{}

// SQLComponent is a sql components set.
// sql 指令元件
type SQLComponent struct {
	Fields     []string
	Functions  []string
	TableName  string
	Wheres     []Where
	Leftjoins  []Join
	Args       []interface{}
	Order      string
	Offset     string
	Limit      string
	WhereRaws  string
	UpdateRaws []RawUpdate
	Group      string
	Statement  string
	Values     H
}

// Where contains the operation and field.
type Where struct {
	Operation string
	Field     string
	Qmark     string
}

// Join contains the table and field and operation.
type Join struct {
	Table     string
	FieldA    string
	Operation string
	FieldB    string
}

// RawUpdate contains the expression and arguments.
// 表達式及參數
type RawUpdate struct {
	Expression string
	Args       []interface{}
}

// *******************************
// 輔助函式
// sql 相關語法
// *******************************
// 顯示資料數
func (sql *SQLComponent) getLimit() string {
	if sql.Limit == "" {
		return ""
	}
	return " limit " + sql.Limit + " "
}

// 跳過前面幾筆資料
func (sql *SQLComponent) getOffset() string {
	if sql.Offset == "" {
		return ""
	}
	return " offset " + sql.Offset + " "
}

// 排列順序
func (sql *SQLComponent) getOrderBy() string {
	if sql.Order == "" {
		return ""
	}
	return " order by " + sql.Order + " "
}

// 分組
func (sql *SQLComponent) getGroupBy() string {
	if sql.Group == "" {
		return ""
	}
	return " group by " + sql.Group + " "
}

// join其他表
func (sql *SQLComponent) getJoins(delimiter string) string {
	if len(sql.Leftjoins) == 0 {
		return ""
	}
	joins := ""
	for _, join := range sql.Leftjoins {
		joins += " left join " + wrap(delimiter, join.Table) + " on " + join.FieldA + " " + join.Operation + " " + join.FieldB + " "
	}
	return joins
}

// 取得特定欄位
func (sql *SQLComponent) getFields(delimiter string) string {
	if len(sql.Fields) == 0 {
		return "*"
	}
	fields := ""
	if len(sql.Leftjoins) == 0 {
		for k, field := range sql.Fields {
			if sql.Functions[k] != "" {
				fields += sql.Functions[k] + "(" + wrap(delimiter, field) + "),"
			} else {
				fields += wrap(delimiter, field) + ","
			}
		}
	} else {
		for _, field := range sql.Fields {
			arr := strings.Split(field, ".")
			if len(arr) > 1 {
				fields += arr[0] + "." + wrap(delimiter, arr[1]) + ","
			} else {
				fields += wrap(delimiter, field) + ","
			}
		}
	}
	return fields[:len(fields)-1]
}

func wrap(delimiter, field string) string {
	if field == "*" {
		return "*"
	}
	if delimiter == "[" {
		return fmt.Sprintf("[%s]", field)
	}
	return delimiter + field + delimiter
}

// where = ...
func (sql *SQLComponent) getWheres(delimiter string) string {
	if len(sql.Wheres) == 0 {
		if sql.WhereRaws != "" {
			return " where " + sql.WhereRaws
		}
		return ""
	}
	wheres := " where "
	var arr []string
	for _, where := range sql.Wheres {
		//arr[0]為資料表名稱
		arr = strings.Split(where.Field, ".")
		if len(arr) > 1 {
			wheres += arr[0] + "." + wrap(delimiter, arr[1]) + " " + where.Operation + " " + where.Qmark + " and "
		} else {
			wheres += wrap(delimiter, where.Field) + " " + where.Operation + " " + where.Qmark + " and "
		}
	}

	if sql.WhereRaws != "" {
		return wheres + sql.WhereRaws
	}
	return wheres[:len(wheres)-5]
}

// update語法
func (sql *SQLComponent) prepareUpdate(delimiter string) {
	fields := ""
	args := make([]interface{}, 0)

	if len(sql.Values) != 0 {

		for key, value := range sql.Values {
			fields += wrap(delimiter, key) + " = ?, "
			args = append(args, value)
		}

		if len(sql.UpdateRaws) == 0 {
			fields = fields[:len(fields)-2]
		} else {
			for i := 0; i < len(sql.UpdateRaws); i++ {
				if i == len(sql.UpdateRaws)-1 {
					fields += sql.UpdateRaws[i].Expression + " "
				} else {
					fields += sql.UpdateRaws[i].Expression + ","
				}
				args = append(args, sql.UpdateRaws[i].Args...)
			}
		}

		sql.Args = append(args, sql.Args...)
	} else {
		if len(sql.UpdateRaws) == 0 {
			panic("prepareUpdate: wrong parameter")
		} else {
			for i := 0; i < len(sql.UpdateRaws); i++ {
				if i == len(sql.UpdateRaws)-1 {
					fields += sql.UpdateRaws[i].Expression + " "
				} else {
					fields += sql.UpdateRaws[i].Expression + ","
				}
				args = append(args, sql.UpdateRaws[i].Args...)
			}
		}
		sql.Args = append(args, sql.Args...)
	}

	sql.Statement = "update " + sql.TableName + " set " + fields + sql.getWheres(delimiter)
}

// 插入數值語法
func (sql *SQLComponent) prepareInsert(delimiter string) {
	fields := " ("
	quesMark := "("

	for key, value := range sql.Values {
		fields += wrap(delimiter, key) + ","
		quesMark += "?,"
		sql.Args = append(sql.Args, value)
	}
	fields = fields[:len(fields)-1] + ")"
	quesMark = quesMark[:len(quesMark)-1] + ")"

	sql.Statement = "insert into " + sql.TableName + fields + " values " + quesMark
}
