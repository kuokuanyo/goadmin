package controller

import (
	"encoding/json"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"

	"github.com/GoAdminGroup/go-admin/context"
)

// RecordOperationLog record all operation logs, store into database.
// 記錄所有操作行為至資料表(goadmin_operation_log)中
func (h *Handler) RecordOperationLog(ctx *context.Context) {
	// 查詢Context.UserValue["user值"]之後轉換成UserModel類別
	if user, ok := ctx.UserValue["user"].(models.UserModel); ok {
		var input []byte
		// 解析的表單(form)參數
		form := ctx.Request.MultipartForm
		if form != nil {
			// 編碼放置input
			input, _ = json.Marshal((*form).Value)
		}

		// OperationLog在plugins\admin\models\operation_log.go
		// OperationLog回傳預設的OperationLogModel(struct)，資料表名為goadmin_operation_log
		// goadmin_operation_log資料表為紀錄使用者操作行為
		// SetConn將參數h.conn(Connection(interface))設置至OperationLogModel.Base.Conn(struct)
		// New新增一筆使用者操作紀錄至資料表，回傳OperationLogModel(struct)
		// 資料表input欄位為儲存使用的參數(例如新建使用者(form-data參數)，{"__go_admin_previous_":["/admin/info/manager?__page=1\u0026__pageSize=10\u0026__sort=id\u0026__sort_type=desc"],"__go_admin_t_":["972c6941-35fc-4401-9e95-e07a53c5370e"],"avatar":[""],"avatar__delete_flag":["0"],"name":["iiiii"],"password":["admin"],"password_again":["admin"],"username":["iiiii"]})
		models.OperationLog().SetConn(h.conn).New(user.Id, ctx.Path(), ctx.Method(), ctx.LocalIP(), string(input))
	}
}
