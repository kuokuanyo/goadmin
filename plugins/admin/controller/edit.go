package controller

import (
	"fmt"
	template2 "html/template"
	"net/http"
	"net/url"

	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/file"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	form2 "github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/guard"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

// ShowForm show form page.
func (h *Handler) ShowForm(ctx *context.Context) {
	// 取得Context.UserValue[show_form_param]的值並轉換成DeleteParam(struct)
	param := guard.GetShowFormParam(ctx)

	// param.Prefix = manager、roles、permission
	// param.Param的資訊有最大顯示資料數、排列順序、依照什麼欄位排序、選擇顯示欄位....等
	h.showForm(ctx, "", param.Prefix, param.Param, false)
}

// 首先透過id取得資料表中的資料，接著對有帶值的欄位更新並加入FormFields，設置需要用到的url(返回、保存...等按鍵)
// 最後匯出HTML語法
func (h *Handler) showForm(ctx *context.Context, alert template2.HTML, prefix string, param parameter.Parameters, isEdit bool, animation ...bool) {
	// BaseTable也屬於Table(interface)
	// 先透過參數prefix取得Table(interface)，接著判斷條件後將[]context.Node加入至Handler.operations後回傳
	panel := h.table(prefix, ctx)
	// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
	user := auth.Auth(ctx)

	// 取得url.Values(map[string][]string)後加入__page(鍵)與值s，最後編碼並回傳
	// ex: ?__goadmin_edit_pk=26&__page=1&__pageSize=10&__pk=26&__sort=id&__sort_type=desc
	paramStr := param.GetRouteParamStr()

	// AorEmpty判斷第一個(condition)參數，如果true則回傳第二個參數，否則回傳""
	// routePathWithPrefix透過參數name取得該路徑名稱的URL，將url中的:__prefix改成第二個參數(prefix)
	// ex: /admin/info/manager/new?__goadmin_edit_pk=26&__page=1&__pageSize=10&__pk=26&__sort=id&__sort_type=desc
	// 新增資料頁面
	newUrl := modules.AorEmpty(panel.GetCanAdd(), h.routePathWithPrefix("show_new", prefix)+paramStr)

	footerKind := "edit"
	// CheckPermissionByUrlMethod檢查權限(藉由url、method)
	// 如果沒有新增資料權限，則為"edit_only"
	if newUrl == "" || !user.CheckPermissionByUrlMethod(newUrl, h.route("show_new").Method(), url.Values{}) {
		footerKind = "edit_only"
	}

	// GetDataWithId(透過id取得資料)透過id取得資料表中的資料，接著對有帶值的欄位更新並加入FormFields後回傳，最後設置值至FormInfo(struct)中
	// formInfo欄位資訊包含資料數值
	formInfo, err := panel.GetDataWithId(param)

	// DeletePK刪除Parameters.Fields[__pk]後回傳
	// GetRouteParamStr取得url.Values(map[string][]string)後加入__page(鍵)與值s，最後編碼並回傳
	// ex:/admin/info/manager/edit?__goadmin_edit_pk=26&__page=1&__pageSize=10&__sort=id&__sort_type=desc(編輯頁面)
	showEditUrl := h.routePathWithPrefix("show_edit", prefix) + param.DeletePK().GetRouteParamStr()

	if err != nil {
		h.HTML(ctx, user, types.Panel{
			Content:     aAlert().Warning(err.Error()),
			Description: template2.HTML(panel.GetForm().Description),
			Title:       template2.HTML(panel.GetForm().Title),
		}, alert == "" || ((len(animation) > 0) && animation[0]))

		if isEdit {
			ctx.AddHeader(constant.PjaxUrlHeader, showEditUrl)
		}
		return
	}

	// EditPKKey = __goadmin_edit_pk
	// 刪除Parameters.Fields[參數(__goadmin_edit_pk)]後回傳
	// GetRouteParamStr取得url.Values(map[string][]string)後加入__page(鍵)與值s，最後編碼並回傳
	// ex: /admin/info/manager?__page=1&__pageSize=10&__sort=id&__sort_type=desc(顯示所有資料頁面)
	infoUrl := h.routePathWithPrefix("info", prefix) + param.DeleteField(constant.EditPKKey).GetRouteParamStr()
	// ex: /admin/edit/manager(編輯功能(post))
	editUrl := h.routePathWithPrefix("edit", prefix)

	// ex: http://localhost:9033/admin/info/manager(標頭中)
	referer := ctx.Headers("Referer")
	if referer != "" && !isInfoUrl(referer) && !isEditUrl(referer, ctx.Query(constant.PrefixKey)) {
		infoUrl = referer
	}

	// 將參數值(BaseTable.PrimaryKey)的值設置至BaseTable.Form(FormPanel(struct)).primaryKey中後回傳FormPanel(struct)
	// f為所有欄位資訊...等資訊
	f := panel.GetForm()

	// 如果url沒設置__goadmin_iframe則為空值
	isNotIframe := ctx.Query(constant.IframeKey) != "true" // ex: true

	// 隱藏資訊(__go_admin_t_、__go_admin_previous_)
	hiddenFields := map[string]string{
		form2.TokenKey:    h.authSrv().AddToken(),
		form2.PreviousKey: infoUrl,
	}

	// 如果url沒設置則都為空
	// IframeKey = __goadmin_iframe
	if ctx.Query(constant.IframeKey) != "" {
		hiddenFields[constant.IframeKey] = ctx.Query(constant.IframeKey)
	}
	// IframeIDKey = __goadmin_iframe_id
	if ctx.Query(constant.IframeIDKey) != "" {
		hiddenFields[constant.IframeIDKey] = ctx.Query(constant.IframeIDKey)
	}

	// formContent尋找{{define "box"}}，將form.GetContent()設置至body以及設置header...等資訊
	// 先將表單資訊設置後，尋找{{define "box"}}將表單包起來
	content := formContent(aForm().
		SetContent(formInfo.FieldList).            // 將欄位資訊及數值設置至表單的content
		SetFieldsHTML(f.HTMLContent).              // ex:""
		SetTabContents(formInfo.GroupFieldList).   // ex:[]
		SetTabHeaders(formInfo.GroupFieldHeaders). // ex:[]
		SetPrefix(h.config.PrefixFixSlash()).      // ex:/admin
		SetInputWidth(f.InputWidth).               // ex:0
		SetHeadWidth(f.HeadWidth).                 // ex:0
		SetPrimaryKey(panel.GetPrimaryKey().Name). // ex:id
		SetUrl(editUrl).
		SetAjax(f.AjaxSuccessJS, f.AjaxErrorJS).
		SetLayout(f.Layout).           // ex:LayoutDefault
		SetHiddenFields(hiddenFields). // 隱藏資訊(__go_admin_t_、__go_admin_previous_)
		SetOperationFooter(formFooter(footerKind,
						f.IsHideContinueEditCheckBox,
						f.IsHideContinueNewCheckBox,
						f.IsHideResetButton)).
		SetHeader(f.HeaderHtml). // ex:HeaderHtml、FooterHtml為[]
		SetFooter(f.FooterHtml), len(formInfo.GroupFieldHeaders) > 0, !isNotIframe, f.IsHideBackButton, f.Header)

	// 一般不會執行
	if f.Wrapper != nil {
		content = f.Wrapper(content)
	}

	h.HTML(ctx, user, types.Panel{
		Content:     alert + content,
		Description: template2.HTML(formInfo.Description),
		Title:       modules.AorBHTML(isNotIframe, template2.HTML(formInfo.Title), ""),
	}, alert == "" || ((len(animation) > 0) && animation[0]))

	// 一般不會執行
	if isEdit {
		ctx.AddHeader(constant.PjaxUrlHeader, showEditUrl)
	}
}

// 首先取得multipart/form-data所設定的參數，接著更新資料數值
// 接著透過處理sql語法後接著取得資料表所有資料後，判斷條件後處理並將值設置至PanelInfo(struct)，panelInfo為頁面上所有的資料
// 接著將所有頁面的HTML處理並回傳(包括標頭、過濾條件、所有顯示的資料)
func (h *Handler) EditForm(ctx *context.Context) {

	// GetEditFormParam回傳Context.UserValue[edit_form_param]的值(struct)
	param := guard.GetEditFormParam(ctx)

	// 如果有上傳頭像檔案才會執行，否則為空map[]
	if len(param.MultiForm.File) > 0 {
		err := file.GetFileEngine(h.config.FileUploadEngine.Name).Upload(param.MultiForm)
		if err != nil {
			if ctx.WantJSON() {
				response.Error(ctx, err.Error())
			} else {
				h.showForm(ctx, aAlert().Warning(err.Error()), param.Prefix, param.Param, true)
			}
			return
		}
	}

	// GetForm在\plugins\admin\modules\table\table.go(BaseTable、table的方法)
	// 將參數值(BaseTable.PrimaryKey)的值設置至BaseTable.Form(FormPanel(struct)).primaryKey中後回傳FormPanel(struct)
	// field為編輯頁面每一欄位的資訊FormField(struct)
	for _, field := range param.Panel.GetForm().FieldList {

		// FormField.FormType為表單型態，ex: select、text、file
		if field.FormType == form.File &&
			len(param.MultiForm.File[field.Field]) == 0 &&
			param.MultiForm.Value[field.Field+"__delete_flag"][0] != "1" {
			//頭像會執行此動作
			// EditFormParam.MultiForm.Value取得在multipart/form-data所設定的參數(map[string][]string)
			// 刪除param.MultiForm.Value[field.Field]值
			delete(param.MultiForm.Value, field.Field)
		}
	}

	// Value()取得EditFormParam.MultiForm.Value(map[string][]string)，multipart/form-data所設定的參數
	// UpdateData先將參數(map[string][]string)資料整理，接著判斷條件後執行資料更新的動作
	err := param.Panel.UpdateData(param.Value())
	if err != nil {
		// 判斷header裡包含accept:json
		if ctx.WantJSON() {
			response.Error(ctx, err.Error())
		} else {
			h.showForm(ctx, aAlert().Warning(err.Error()), param.Prefix, param.Param, true)
		}
		return
	}

	// -------編輯介面不會執行---------
	if param.Panel.GetForm().Responder != nil {
		param.Panel.GetForm().Responder(ctx)
		return
	}

	// -------編輯介面不會執行---------
	if ctx.WantJSON() && !param.IsIframe {
		response.OkWithData(ctx, map[string]interface{}{
			"url": param.PreviousPath,
		})
		return
	}

	// -------編輯介面不會執行---------
	if !param.FromList {
		if isNewUrl(param.PreviousPath, param.Prefix) {
			h.showNewForm(ctx, param.Alert, param.Prefix, param.Param.DeleteEditPk().GetRouteParamStr(), true)
			return
		}

		if isEditUrl(param.PreviousPath, param.Prefix) {
			h.showForm(ctx, param.Alert, param.Prefix, param.Param, true, false)
			return
		}

		ctx.HTML(http.StatusOK, fmt.Sprintf(`<script>location.href="%s"</script>`, param.PreviousPath))
		ctx.AddHeader(constant.PjaxUrlHeader, param.PreviousPath)
		return
	}

	// -------編輯介面不會執行---------
	if param.IsIframe {
		ctx.HTML(http.StatusOK, fmt.Sprintf(`<script>
		swal('%s', '', 'success');
		setTimeout(function(){
			$("#%s", window.parent.document).hide();
			$('.modal-backdrop.fade.in', window.parent.document).hide();
		}, 1000)
</script>`, language.Get("success"), param.IframeID))
		return
	}

	// DeletePK刪除Parameters.Fields[__pk]後回傳
	// DeleteEditPk刪除Parameters.Fields[__goadmin_edit_pk]後回傳
	// param.Prefix = manager、roles or permission
	// showTable首先透過處理sql語法後接著取得資料表所有資料後，判斷條件後處理並將值設置至PanelInfo(struct)，panelInfo為頁面上所有的資料
	// 接著將所有頁面的HTML處理並回傳(包括標頭、過濾條件、所有顯示的資料)
	buf := h.showTable(ctx, param.Prefix, param.Param.DeletePK().DeleteEditPk(), nil)

	ctx.HTML(http.StatusOK, buf.String())

	// add header
	ctx.AddHeader(constant.PjaxUrlHeader, param.PreviousPath)
}
