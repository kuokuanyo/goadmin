package models

import (
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
)

// OperationLogModel is operation log model structure.
// OperationLogModel紀錄使用者操作行為
type OperationLogModel struct {
	//Base(struct)在plugins\admin\models\base.go
	Base

	Id        int64
	UserId    int64
	Path      string
	Method    string
	Ip        string
	Input     string
	CreatedAt string
	UpdatedAt string
}

// OperationLog return a default operation log model.
// 回傳預設的OperationLogModel(struct)，資料表名為goadmin_operation_log
// goadmin_operation_log資料表為紀錄使用者操作行為
func OperationLog() OperationLogModel {
	return OperationLogModel{Base: Base{TableName: "goadmin_operation_log"}}
}

// Find return a default operation log model of given id.
// 透過參數(id)尋找符合資料，將資訊設置至OperationLogModel(struct)
func (t OperationLogModel) Find(id interface{}) OperationLogModel {
	// 在plugins\admin\models\base.go
	// Table藉由給定的參數(t.TableName)回傳sql(struct)
	// Find在modules\db\statement.go中
	// Find藉由參數id取得符合資料
	item, _ := t.Table(t.TableName).Find(id)
	// 透過參數(m map[string]interface{})將資訊設置至OperationLogModel(struct)
	return t.MapToModel(item)
}

// 將參數conn(Connection(interface))設置至OperationLogModel.Base.Conn(struct)
func (t OperationLogModel) SetConn(con db.Connection) OperationLogModel {
	t.Conn = con
	return t
}

// New create a new operation log model.
// 新增一筆使用者行為資料至資料表，回傳OperationLogModel(struct)
func (t OperationLogModel) New(userId int64, path, method, ip, input string) OperationLogModel {

	// 插入使用者行為資料
	// OperationLogModel.Base.TableName
	// Table藉由給定的參數(t.TableName)回傳sql(struct)
	id, _ := t.Table(t.TableName).Insert(dialect.H{
		"user_id": userId,
		"path":    path,
		"method":  method,
		"ip":      ip,
		"input":   input,
	})

	t.Id = id
	t.UserId = userId
	t.Path = path
	t.Method = method
	t.Ip = ip
	t.Input = input

	return t
}

// MapToModel get the operation log model from given map.
// 透過參數(m map[string]interface{})將資訊設置至OperationLogModel(struct)
func (t OperationLogModel) MapToModel(m map[string]interface{}) OperationLogModel {
	t.Id = m["id"].(int64)
	t.UserId = m["user_id"].(int64)
	t.Path, _ = m["path"].(string)
	t.Method, _ = m["method"].(string)
	t.Ip, _ = m["ip"].(string)
	t.Input, _ = m["input"].(string)
	t.CreatedAt, _ = m["created_at"].(string)
	t.UpdatedAt, _ = m["updated_at"].(string)
	return t
}
