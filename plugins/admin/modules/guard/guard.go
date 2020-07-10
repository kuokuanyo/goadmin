package guard

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/service"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
)

type Guard struct {
	services  service.List
	conn      db.Connection
	tableList table.GeneratorList
	navBtns   *types.Buttons
}

// 將參數s、c、t設置至Guard(struct)後回傳
func New(s service.List, c db.Connection, t table.GeneratorList, b *types.Buttons) *Guard {
	return &Guard{
		services:  s,
		conn:      c,
		tableList: t,
		navBtns:   b,
	}
}

func (g *Guard) table(ctx *context.Context) (table.Table, string) {
	prefix := ctx.Query(constant.PrefixKey)
	return g.tableList[prefix](ctx), prefix
}


// 查詢url裡的參數(__prefix)，如果Guard.tableList存在該prefix(key)則執行迴圈
func (g *Guard) CheckPrefix(ctx *context.Context) {

	// PrefixKey = __prefix
	// 取得Request url裡的參數(__prefix)
	prefix := ctx.Query(constant.PrefixKey)

	if _, ok := g.tableList[prefix]; !ok {
		if ctx.Headers(constant.PjaxHeader) == "" && ctx.Method() != "GET" {
			response.BadRequest(ctx, errors.Msg)
		} else {
			response.Alert(ctx, errors.Msg, errors.Msg, "table model not found", g.conn, g.navBtns,
				template.Missing404Page)
		}
		ctx.Abort()
		return
	}

	// 執行迴圈Context.handlers[ctx.index](ctx)
	ctx.Next()
}

const (
	editFormParamKey   = "edit_form_param"
	deleteParamKey     = "delete_param"
	exportParamKey     = "export_param"
	deleteMenuParamKey = "delete_menu_param"
	editMenuParamKey   = "edit_menu_param"
	newMenuParamKey    = "new_menu_param"
	newFormParamKey    = "new_form_param"
	updateParamKey     = "update_param"
	showFormParamKey   = "show_form_param"
	showNewFormParam   = "show_new_form_param"
)
