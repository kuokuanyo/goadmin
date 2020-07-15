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
	param := guard.GetShowFormParam(ctx)
	h.showForm(ctx, "", param.Prefix, param.Param, false)
}

func (h *Handler) showForm(ctx *context.Context, alert template2.HTML, prefix string, param parameter.Parameters, isEdit bool, animation ...bool) {

	panel := h.table(prefix, ctx)

	user := auth.Auth(ctx)

	paramStr := param.GetRouteParamStr()

	newUrl := modules.AorEmpty(panel.GetCanAdd(), h.routePathWithPrefix("show_new", prefix)+paramStr)
	footerKind := "edit"
	if newUrl == "" || !user.CheckPermissionByUrlMethod(newUrl, h.route("show_new").Method(), url.Values{}) {
		footerKind = "edit_only"
	}

	formInfo, err := panel.GetDataWithId(param)

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

	infoUrl := h.routePathWithPrefix("info", prefix) + param.DeleteField(constant.EditPKKey).GetRouteParamStr()
	editUrl := h.routePathWithPrefix("edit", prefix)

	referer := ctx.Headers("Referer")

	if referer != "" && !isInfoUrl(referer) && !isEditUrl(referer, ctx.Query(constant.PrefixKey)) {
		infoUrl = referer
	}

	f := panel.GetForm()

	isNotIframe := ctx.Query(constant.IframeKey) != "true"

	hiddenFields := map[string]string{
		form2.TokenKey:    h.authSrv().AddToken(),
		form2.PreviousKey: infoUrl,
	}

	if ctx.Query(constant.IframeKey) != "" {
		hiddenFields[constant.IframeKey] = ctx.Query(constant.IframeKey)
	}

	if ctx.Query(constant.IframeIDKey) != "" {
		hiddenFields[constant.IframeIDKey] = ctx.Query(constant.IframeIDKey)
	}

	content := formContent(aForm().
		SetContent(formInfo.FieldList).
		SetFieldsHTML(f.HTMLContent).
		SetTabContents(formInfo.GroupFieldList).
		SetTabHeaders(formInfo.GroupFieldHeaders).
		SetPrefix(h.config.PrefixFixSlash()).
		SetInputWidth(f.InputWidth).
		SetHeadWidth(f.HeadWidth).
		SetPrimaryKey(panel.GetPrimaryKey().Name).
		SetUrl(editUrl).
		SetAjax(f.AjaxSuccessJS, f.AjaxErrorJS).
		SetLayout(f.Layout).
		SetHiddenFields(hiddenFields).
		SetOperationFooter(formFooter(footerKind,
			f.IsHideContinueEditCheckBox,
			f.IsHideContinueNewCheckBox,
			f.IsHideResetButton)).
		SetHeader(f.HeaderHtml).
		SetFooter(f.FooterHtml), len(formInfo.GroupFieldHeaders) > 0, !isNotIframe, f.IsHideBackButton, f.Header)

	if f.Wrapper != nil {
		content = f.Wrapper(content)
	}

	h.HTML(ctx, user, types.Panel{
		Content:     alert + content,
		Description: template2.HTML(formInfo.Description),
		Title:       modules.AorBHTML(isNotIframe, template2.HTML(formInfo.Title), ""),
	}, alert == "" || ((len(animation) > 0) && animation[0]))

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
