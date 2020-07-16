package guard

import (
	"html/template"
	"mime/multipart"
	"strings"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
)

type ShowNewFormParam struct {
	Panel  table.Table
	Prefix string
	Param  parameter.Parameters
}

// 將值設置至Context.UserValue[show_new_form_param]
func (g *Guard) ShowNewForm(ctx *context.Context) {
	// 取得url中__prefix的值
	// prefix = manager、roles、permission
	panel, prefix := g.table(ctx)

	if !panel.GetCanAdd() {
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

	if panel.GetOnlyUpdateForm() {
		ctx.Redirect(config.Url("/info/" + prefix + "/edit"))
		ctx.Abort()
		return
	}

	// showNewFormParam = show_new_form_param
	ctx.SetUserValue(showNewFormParam, &ShowNewFormParam{
		Panel:  panel,
		Prefix: prefix,
		Param: parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField,
			panel.GetInfo().GetSort()),
	})
	ctx.Next()
}

// 取得Context.UserValue[show_new_form_param]的值並轉換成ShowNewFormParam(struct)
func GetShowNewFormParam(ctx *context.Context) *ShowNewFormParam {
	return ctx.UserValue[showNewFormParam].(*ShowNewFormParam)
}

type NewFormParam struct {
	Panel        table.Table
	Id           string
	Prefix       string
	Param        parameter.Parameters
	Path         string
	MultiForm    *multipart.Form
	PreviousPath string
	FromList     bool
	IsIframe     bool
	IframeID     string
	Alert        template.HTML
}

// 回傳NewFormParam.MultiForm.Value
func (e NewFormParam) Value() form.Values {
	return e.MultiForm.Value
}

// NewForm(新增表單)新增用戶、角色、權限等表單資訊，首先取得multipart/form-data設定的參數值並驗證token是否正確
// 接著取得頁面size、資料排列方式、選擇欄位...等資訊後設置至Parameters(struct)，最後設定Context.UserValue並執行新增表單的動作
func (g *Guard) NewForm(ctx *context.Context) {
	// form.PreviousKey  = __go_admin_previous_
	// 藉由參數取得multipart/form-data中的__go_admin_previous_值
	// ex:/admin/info/manager?__page=1&__pageSize=10&__sort=id&__sort_type=desc
	previous := ctx.FormValue(form.PreviousKey)

	// 取得url中__prefix的值
	// prefix = manager、roles、permission
	panel, prefix := g.table(ctx)

	// 取得匹配的service.Service然後轉換成Connection(interface)類別
	conn := db.GetConnection(g.services)

	// form.TokenKey  = __go_admin_t_
	// 藉由參數取得multipart/form-data中的__go_admin_t_值
	token := ctx.FormValue(form.TokenKey)

	// GetTokenService將參數g.services.Get(auth.TokenServiceKey)轉換成TokenService(struct)類別後回傳
	// GetTokenService透過參數(token_csrf_helper)取得匹配的Service(interface)
	// CheckToken檢查TokenService.tokens([]string)裡是否有符合參數toCheckToken的值
	// 如果符合，將在TokenService.tokens([]string)裡將符合的toCheckToken從[]string拿出
	// 檢查token是否正確
	if !auth.GetTokenService(g.services.Get(auth.TokenServiceKey)).CheckToken(token) {
		alert(ctx, panel, errors.CreateFailWrongToken, conn, g.navBtns)
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

	// 取得在multipart/form-data所設定的參數(map[string][]string)
	values := ctx.Request.MultipartForm.Value

	// newFormParamKey = new_form_param
	// SetUserValue藉由參數key、value設定Context.UserValue
	ctx.SetUserValue(newFormParamKey, &NewFormParam{
		Panel:        panel,
		Id:           "",
		Prefix:       prefix,                                                // manage or roles or permissions
		Param:        param,                                                 // 頁面size、資料排列方式、選擇欄位...等資訊
		IsIframe:     form.Values(values).Get(constant.IframeKey) == "true", // ex:false
		IframeID:     form.Values(values).Get(constant.IframeIDKey),         // ex:空
		Path:         strings.Split(previous, "?")[0],                       // ex:/admin/info/manager(roles or permissions)
		MultiForm:    ctx.Request.MultipartForm,                             // 在multipart/form-data所設定的參數
		PreviousPath: previous,                                              // ex: /admin/info/manager?__page=1&__pageSize=10&__sort=id&__sort_type=desc
		FromList:     fromList,
	})
	ctx.Next()
}

func GetNewFormParam(ctx *context.Context) *NewFormParam {
	return ctx.UserValue[newFormParamKey].(*NewFormParam)
}
