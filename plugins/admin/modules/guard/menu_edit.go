package guard

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"html/template"
	"strconv"
)

type MenuEditParam struct {
	Id       string
	Title    string
	Header   string
	ParentId int64
	Icon     string
	Uri      string
	Roles    []string
	Alert    template.HTML
}

// 檢查MenuEditParam.Alert是否出現警告(不是空值)
func (e MenuEditParam) HasAlert() bool {
	return e.Alert != template.HTML("")
}

// 藉由參數取得multipart/form-data中設置的值，接著驗證token並將multipart/form-data的key、value值設置至Context.UserValue[edit_menu_param]，最後執行迴圈Context.handlers[ctx.index](ctx)
func (g *Guard) MenuEdit(ctx *context.Context) {

	// 藉由參數取得multipart/form-data中的parent_id值
	parentId := ctx.FormValue("parent_id")
	if parentId == "" {
		parentId = "0"
	}

	var (
		parentIdInt, _ = strconv.Atoi(parentId)
		// TokenKey = __go_admin_t_
		// 藉由參數取得multipart/form-data中的__go_admin_t_值
		token          = ctx.FormValue(form.TokenKey)
		alert          template.HTML
	)

	// TokenServiceKey = token_csrf_helper
	// Get透過參數(token_csrf_helper)取得匹配的Service(interface)
	// GetTokenService將參數s轉換成TokenService(struct)類別後回傳
	// CheckToken檢查TokenService.tokens([]string)裡是否有符合參數token的值
    // 如果符合，將在TokenService.tokens([]string)裡將符合的token從[]string拿出
	if !auth.GetTokenService(g.services.Get(auth.TokenServiceKey)).CheckToken(token) {
		alert = getAlert(errors.EditFailWrongToken)
	}

	// title與icon值一定要設置(multipart/form-data)
	// checkEmpty檢查參數(多個key)有在multipart/form-data裡設置(如果值為空則出現錯誤)
	if alert == "" {
		alert = checkEmpty(ctx, "id", "title", "icon")
	}

	// TODO: check the user permission
	// editMenuParamKey = edit_menu_param
	// SetUserValue藉由參數edit_menu_param、&MenuEditParam{...}(struct)設定Context.UserValue
	// 將multipart/form-data的key、value值設置至Context.UserValue[edit_menu_param]
	ctx.SetUserValue(editMenuParamKey, &MenuEditParam{
		Id:       ctx.FormValue("id"),
		Title:    ctx.FormValue("title"),
		Header:   ctx.FormValue("header"),
		ParentId: int64(parentIdInt),
		Icon:     ctx.FormValue("icon"),
		Uri:      ctx.FormValue("uri"),
		Roles:    ctx.Request.Form["roles[]"],
		Alert:    alert,
	})

	// 執行迴圈Context.handlers[ctx.index](ctx)
	ctx.Next()
}

// 將Context.UserValue(map[string]interface{})[edit_menu_param]的值轉換成MenuEditParam(struct)類別
func GetMenuEditParam(ctx *context.Context) *MenuEditParam {
	// editMenuParamKey = edit_menu_param
	// 將Context.UserValue(map[string]interface{})[edit_menu_param]的值轉換成MenuEditParam(struct)類別
	return ctx.UserValue[editMenuParamKey].(*MenuEditParam)
}

// 檢查參數(多個key)有在multipart/form-data裡設置(如果值為空則出現錯誤)
func checkEmpty(ctx *context.Context, key ...string) template.HTML {
	for _, k := range key {
		if ctx.FormValue(k) == "" {
			return getAlert("wrong " + k)
		}
	}
	return template.HTML("")
}
