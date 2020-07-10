package guard

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/errors"
)

type MenuDeleteParam struct {
	Id string
}

// 查詢url中參數id的值後將id設置至MenuDeleteParam(struct)，接著將值設置至Context.UserValue[delete_menu_param]中，最後執行迴圈Context.handlers[ctx.index](ctx)
func (g *Guard) MenuDelete(ctx *context.Context) {

	// 查詢url中參數id的值
	id := ctx.Query("id")

	if id == "" {
		// alertWithTitleAndDesc在plugins\admin\modules\guard\edit.go
		// WrongID = wrong id
		// Alert透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel，接著將給定的數據(types.Page(struct))寫入buf(struct)並回傳，最後輸出HTML
		// 將參數Menu、menu、errors.WrongID寫入Panel
		alertWithTitleAndDesc(ctx, "Menu", "menu", errors.WrongID, g.conn, g.navBtns)
		ctx.Abort()
		return
	}

	// TODO: check the user permission
	// deleteMenuParamKey = delete_menu_param
	// 將id設置至MenuDeleteParam(struct)
	// SetUserValue藉由參數delete_menu_param、&MenuDeleteParam{...}(struct)設定Context.UserValue
	// 將參數設置至Context.UserValue[delete_menu_param]
	ctx.SetUserValue(deleteMenuParamKey, &MenuDeleteParam{
		Id: id,
	})

	// 執行迴圈Context.handlers[ctx.index](ctx)
	ctx.Next()
}

// 將Context.UserValue(map[string]interface{})[delete_menu_param]的值轉換成MenuDeleteParam(struct)類別
func GetMenuDeleteParam(ctx *context.Context) *MenuDeleteParam {
	// deleteMenuParamKey = delete_menu_param
	// 將Context.UserValue(map[string]interface{})[delete_menu_param]的值轉換成MenuDeleteParam(struct)類別
	return ctx.UserValue[deleteMenuParamKey].(*MenuDeleteParam)
}
