package guard

import (
	"fmt"
	tmpl "html/template"
	"mime/multipart"
	"regexp"
	"strings"

	"github.com/GoAdminGroup/go-admin/template/types"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template"
)

type ShowFormParam struct {
	Panel  table.Table
	Id     string
	Prefix string
	Param  parameter.Parameters
}

func (g *Guard) ShowForm(ctx *context.Context) {

	panel, prefix := g.table(ctx)

	if !panel.GetEditable() {
		alert(ctx, panel, errors.OperationNotAllow, g.conn, g.navBtns)
		ctx.Abort()
		return
	}

	if panel.GetOnlyInfo() {
		ctx.Redirect(config.Url("/info/" + prefix))
		ctx.Abort()
		return
	}

	if panel.GetOnlyDetail() {
		ctx.Redirect(config.Url("/info/" + prefix + "/detail"))
		ctx.Abort()
		return
	}

	if panel.GetOnlyNewForm() {
		ctx.Redirect(config.Url("/info/" + prefix + "/new"))
		ctx.Abort()
		return
	}

	id := ctx.Query(constant.EditPKKey)
	if id == "" && prefix != "site" {
		alert(ctx, panel, errors.WrongPK(panel.GetPrimaryKey().Name), g.conn, g.navBtns)
		ctx.Abort()
		return
	}
	if prefix == "site" {
		id = "1"
	}

	ctx.SetUserValue(showFormParamKey, &ShowFormParam{
		Panel:  panel,
		Id:     id,
		Prefix: prefix,
		Param: parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField,
			panel.GetInfo().GetSort()).WithPKs(id),
	})
	ctx.Next()
}

func GetShowFormParam(ctx *context.Context) *ShowFormParam {
	return ctx.UserValue[showFormParamKey].(*ShowFormParam)
}

type EditFormParam struct {
	Panel        table.Table
	Id           string
	Prefix       string
	Param        parameter.Parameters
	Path         string
	MultiForm    *multipart.Form
	PreviousPath string
	Alert        tmpl.HTML
	FromList     bool
	IsIframe     bool
	IframeID     string
}

func (e EditFormParam) Value() form.Values {
	return e.MultiForm.Value
}

// EditForm(編輯表單)編輯用戶、角色、權限等表單資訊，首先取得multipart/form-data設定的參數值並驗證token是否正確
// 接著取得頁面size、資料排列方式、選擇欄位...等資訊後設置至Parameters(struct)，最後設定Context.UserValue並執行編輯表單的動作
func (g *Guard) EditForm(ctx *context.Context) {

	// form.PreviousKey  = __go_admin_previous_
	// 藉由參數取得multipart/form-data中的__go_admin_previous_值
	// ex:/admin/info/manager?__page=1&__pageSize=10&__sort=id&__sort_type=desc
	previous := ctx.FormValue(form.PreviousKey)


	// 取得url中__prefix的值
	panel, prefix := g.table(ctx)

	// GetEditable回傳BaseTable.Editable(是否可以編輯)
	if !panel.GetEditable() {
		alert(ctx, panel, errors.OperationNotAllow, g.conn, g.navBtns)
		ctx.Abort()
		return
	}

	// form.TokenKey  = __go_admin_t_
	// 藉由參數取得multipart/form-data中的__go_admin_t_值
	token := ctx.FormValue(form.TokenKey)

	// GetTokenService將參數g.services.Get(auth.TokenServiceKey)轉換成TokenService(struct)類別後回傳
	// GetTokenService透過參數(token_csrf_helper)取得匹配的Service(interface)
	// CheckToken檢查TokenService.tokens([]string)裡是否有符合參數toCheckToken的值
	// 如果符合，將在TokenService.tokens([]string)裡將符合的toCheckToken從[]string拿出
	// 檢查token是否正確
	if !auth.GetTokenService(g.services.Get(auth.TokenServiceKey)).CheckToken(token) {
		alert(ctx, panel, errors.EditFailWrongToken, g.conn, g.navBtns)
		ctx.Abort()
		return
	}

	// 判斷參數是否是info url(true)
	fromList := isInfoUrl(previous)

	// GetInfo將參數值設置至base.Info(InfoPanel(struct)).primaryKey中後回傳
	// GetParamFromURL在plugins\admin\modules\parameter\parameter.go
	// GetParamFromURL(從URL中取得參數)取得頁面size、資料排列方式、選擇欄位...等資訊後設置至Parameters(struct)並回傳
	param := parameter.GetParamFromURL(previous, panel.GetInfo().DefaultPageSize,
		// GetPrimaryKey回傳BaseTable.PrimaryKey
		panel.GetInfo().GetSort(), panel.GetPrimaryKey().Name)

	if fromList {
		// GetRouteParamStr取得url.Values後加入__page(鍵)與值，最後編碼並回傳
		previous = config.Url("/info/" + prefix + param.GetRouteParamStr())
	}

	multiForm := ctx.Request.MultipartForm

	// 取得id
	// GetPrimaryKey在plugins\admin\modules\table\table.go
	// GetPrimaryKey回傳BaseTable.PrimaryKey
	id := multiForm.Value[panel.GetPrimaryKey().Name][0]

	// 取得在multipart/form-data所設定的參數
	values := ctx.Request.MultipartForm.Value

	// SetUserValue藉由參數key、value設定Context.UserValue
	ctx.SetUserValue(editFormParamKey, &EditFormParam{
		Panel:     panel,
		Id:        id,
		Prefix:    prefix,                          // manage or roles or permissions
		Param:     param.WithPKs(id),               // 將參數(多個string)結合並設置至Parameters.Fields["__pk"]後回傳
		Path:      strings.Split(previous, "?")[0], // ex:/admin/info/manager(roles or permissions)
		MultiForm: multiForm,
		// constant.IframeKey = __goadmin_iframe
		IsIframe: form.Values(values).Get(constant.IframeKey) == "true", // ex:false
		// constant.IframeIDKey = __goadmin_iframe_id
		IframeID:     form.Values(values).Get(constant.IframeIDKey),
		PreviousPath: previous, // ex: /admin/info/manager?__page=1&__pageSize=10&__sort=id&__sort_type=desc
		FromList:     fromList, // ex: true
	})
	ctx.Next()
}

// 判斷參數是否是info url
func isInfoUrl(s string) bool {
	reg, _ := regexp.Compile("(.*?)info/(.*?)$")
	sub := reg.FindStringSubmatch(s)
	return len(sub) > 2 && !strings.Contains(sub[2], "/")
}

func GetEditFormParam(ctx *context.Context) *EditFormParam {
	return ctx.UserValue[editFormParamKey].(*EditFormParam)
}

func alert(ctx *context.Context, panel table.Table, msg string, conn db.Connection, btns *types.Buttons) {
	if ctx.WantJSON() {
		response.BadRequest(ctx, msg)
	} else {
		response.Alert(ctx, panel.GetInfo().Description, panel.GetInfo().Title, msg, conn, btns)
	}
}

// Alert透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel，接著將給定的數據(types.Page(struct))寫入buf(struct)並回傳，最後輸出HTML
func alertWithTitleAndDesc(ctx *context.Context, title, desc, msg string, conn db.Connection, btns *types.Buttons) {
	// Alert在plugins\admin\modules\response\response.go
	// Alert透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel，接著將給定的數據(types.Page(struct))寫入buf(struct)並回傳，最後輸出HTML
	// 將參數desc、title、msg寫入Panel
	response.Alert(ctx, desc, title, msg, conn, btns)
}

func getAlert(msg string) tmpl.HTML {
	// GetTheme回傳globalCfg.Theme
	// Get判斷templateMap(map[string]Template)的key鍵是否參數theme，有則回傳Template(interface)
	// Alert為Template(interface)的方法
	return template.Get(config.GetTheme()).Alert().Warning(msg)
}
