package guard

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
)

type DeleteParam struct {
	Panel  table.Table
	Id     string
	Prefix string
}

// 取得參數取得multipart/form-data的id值後將值設置至Context.UserValue[delete_param]
func (g *Guard) Delete(ctx *context.Context) {
	// 取得url中__prefix的值
	// prefix = manager、roles、permission
	panel, prefix := g.table(ctx)
	if !panel.GetDeletable() {
		alert(ctx, panel, errors.OperationNotAllow, g.conn, g.navBtns)
		ctx.Abort()
		return
	}

	// 取得參數取得multipart/form-data的id值
	id := ctx.FormValue("id")
	if id == "" {
		alert(ctx, panel, errors.WrongID, g.conn, g.navBtns)
		ctx.Abort()
		return
	}
	// deleteParamKey = delete_param
	// 將值設置至Context.UserValue[delete_param]
	ctx.SetUserValue(deleteParamKey, &DeleteParam{
		Panel:  panel,
		Id:     id,
		Prefix: prefix,
	})
	ctx.Next()
}

// 取得Context.UserValue[delete_param]的值並轉換成DeleteParam(struct)
func GetDeleteParam(ctx *context.Context) *DeleteParam {
	return ctx.UserValue[deleteParamKey].(*DeleteParam)
}
