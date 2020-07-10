package response

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/menu"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"net/http"
)

// 成功，回傳code:200 and msg:ok
func Ok(ctx *context.Context) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"msg":  "ok",
	})
}

// 成功，回傳code:200 and msg
func OkWithMsg(ctx *context.Context, msg string) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"msg":  msg,
	})
}

// 成功，回傳code:200 and msg:ok and data
func OkWithData(ctx *context.Context, data map[string]interface{}) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code": http.StatusOK,
		"msg":  "ok",
		"data": data,
	})
}

// 錯誤請求，回傳code:400 and msg
func BadRequest(ctx *context.Context, msg string) {
	ctx.JSON(http.StatusBadRequest, map[string]interface{}{
		"code": http.StatusBadRequest,
		// Get依照設定的語言給予訊息
		"msg":  language.Get(msg),
	})
}

// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel，接著將給定的數據(types.Page(struct))寫入buf(struct)並回傳，最後輸出HTML
// 將參數desc、title、msg寫入Panel
func Alert(ctx *context.Context, desc, title, msg string, conn db.Connection, btns *types.Buttons,
	pageType ...template.PageType) {

	// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
	user := auth.Auth(ctx)

	pt := template.Error500Page
	if len(pageType) > 0 {
		pt = pageType[0]
	}

	pageTitle, description, content := template.GetPageContentFromPageType(title, desc, msg, pt)

	// Get判斷templateMap(map[string]Template)的key鍵是否參數config.GetTheme()，有則回傳Template(interface)
	// GetTemplate為Template(interface)的方法
	tmpl, tmplName := template.Default().GetTemplate(ctx.IsPjax())
	// 將給定的數據(types.Page(struct))寫入buf(struct)並回傳
	buf := template.Execute(template.ExecuteParam{
		User:     user,
		TmplName: tmplName,
		Tmpl:     tmpl,
		Panel: types.Panel{
			Content:     content,
			Description: description,
			Title:       pageTitle,
		},
		Config:    *config.Get(),
		// GetGlobalMenu回傳參數user(struct)的Menu(設置menuList、menuOption、MaxOrder)
		// 設定menu的active
		// URLRemovePrefix globalCfg(Config struct).prefix將URL的前綴去除
		Menu:      menu.GetGlobalMenu(user, conn).SetActiveClass(config.URLRemovePrefix(ctx.Path())),
		Animation: true,
		Buttons:   *btns,
	})
	// 將buf輸出成HTML
	ctx.HTML(http.StatusOK, buf.String())
}

// 錯誤，回傳code:500 and msg
func Error(ctx *context.Context, msg string) {
	ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
		"code": http.StatusInternalServerError,
		"msg":  language.Get(msg),
	})
}

// 錯誤，回傳code:403 and msg
func Denied(ctx *context.Context, msg string) {
	ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
		"code": http.StatusForbidden,
		"msg":  language.Get(msg),
	})
}

// OffLineHandler(function)
// 判斷站點是否要關閉，如要關閉，判斷method是否為get以及header裡包含accept:html後輸出HTML
var OffLineHandler = func(ctx *context.Context) {
	// GetSiteOff回傳globalCfg.SiteOff(站點關閉)
	if config.GetSiteOff() {
		// 判斷method是否為get以及header裡包含accept:html
		if ctx.WantHTML() {
			//輸出HTML
			ctx.HTML(http.StatusOK, `<html><body><h1>The website is offline</h1></body></html>`)
		} else {
			ctx.JSON(http.StatusForbidden, map[string]interface{}{
				"code": http.StatusForbidden,
				"msg":  language.Get(errors.SiteOff),
			})
		}
		// Context.index = 63
		ctx.Abort()
	}
}
