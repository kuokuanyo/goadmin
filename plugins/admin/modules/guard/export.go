package guard

import (
	"strings"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
)

type ExportParam struct {
	Panel  table.Table
	Id     []string
	Prefix string
	IsAll  bool
}

// 取得參數取得multipart/form-data的值後將值設置至Context.UserValue[export_param]
func (g *Guard) Export(ctx *context.Context) {
	// 取得url中__prefix的值
	// prefix = manager、roles、permission
	panel, prefix := g.table(ctx)
	if !panel.GetExportable() {
		alert(ctx, panel, errors.OperationNotAllow, g.conn, g.navBtns)
		ctx.Abort()
		return
	}

	idStr := make([]string, 0)
	// 取得參數取得multipart/form-data的id值
	// 如果取得當前頁或全部資料id會回傳空值，如果有選擇導出特定資料則會有id值
	ids := ctx.FormValue("id")
	if ids != "" {
		idStr = strings.Split(ctx.FormValue("id"), ",")
	}

	// exportParamKey = export_param
	// 將值設置至Context.UserValue[export_param]
	ctx.SetUserValue(exportParamKey, &ExportParam{
		Panel:  panel,
		Id:     idStr,
		Prefix: prefix,
		// 透過參數取得multipart/form-data的is_all值(判斷是否取得全部資料)
		IsAll: ctx.FormValue("is_all") == "true",
	})
	ctx.Next()
}

// 取得Context.UserValue[export_param]的值並轉換成DeleteParam(struct)
func GetExportParam(ctx *context.Context) *ExportParam {
	return ctx.UserValue[exportParamKey].(*ExportParam)
}
