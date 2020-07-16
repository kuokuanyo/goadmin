package controller

import (
	"fmt"
	template2 "html/template"
	"net/http"

	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/file"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	form2 "github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/guard"
	"github.com/GoAdminGroup/go-admin/template/types"
)

// ShowNewForm show a new form page.
func (h *Handler) ShowNewForm(ctx *context.Context) {
	// 取得Context.UserValue[show_new_form_param]的值並轉換成ShowNewFormParam(struct)
	param := guard.GetShowNewFormParam(ctx)

	// 取得url.Values(map[string][]string)後加入__page(鍵)與值s，最後編碼並回傳
	// ex:?__page=1&__pageSize=10&__sort=id&__sort_type=desc
	h.showNewForm(ctx, "", param.Prefix, param.Param.GetRouteParamStr(), false)
}

// 首先取得用戶資料、權限等資訊，接著取得新建資料頁面所有的欄位資訊
// 設置需要用到的url(返回、保存、繼續新增...等按鍵)，最後匯出HTML語法
func (h *Handler) showNewForm(ctx *context.Context, alert template2.HTML, prefix, paramStr string, isNew bool) {
	// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
	user := auth.Auth(ctx)

	// BaseTable也屬於Table(interface)
	// 先透過參數prefix取得Table(interface)，接著判斷條件後將[]context.Node加入至Handler.operations後回傳
	panel := h.table(prefix, ctx)

	// GetNewForm(取得新表單)判斷條件(TabGroups)後，設置FormInfo(struct)後並回傳
	// formInfo為新建資料頁面所有的欄位資訊
	formInfo := panel.GetNewForm()
	// routePathWithPrefix透過參數name取得該路徑名稱的URL，將url中的:__prefix改成第二個參數(prefix)
	// ex: /admin/info/manager?__page=1&__pageSize=10&__sort=id&__sort_type=desc(顯示所有資料頁面)，用在返回按鍵
	infoUrl := h.routePathWithPrefix("info", prefix) + paramStr
	// ex: /admin/info/manager/new?__goadmin_edit_pk=26&__page=1&__pageSize=10&__pk=26&__sort=id&__sort_type=desc
	// 新建資料功能(post)
	newUrl := h.routePathWithPrefix("new", prefix)
	// ex:/admin/info/manager/new?__page=1&__pageSize=10&__sort=id&__sort_type=desc
	// 用於如果勾選繼續新增，導向新增資料頁面
	showNewUrl := h.routePathWithPrefix("show_new", prefix) + paramStr

	// ex: http://localhost:9033/admin/info/manager(標頭中)
	referer := ctx.Headers("Referer")

	if referer != "" && !isInfoUrl(referer) && !isNewUrl(referer, ctx.Query(constant.PrefixKey)) {
		infoUrl = referer
	}

	// 將參數值(BaseTable.PrimaryKey)的值設置至BaseTable.Form(FormPanel(struct)).primaryKey中後回傳FormPanel(struct)
	// f為所有欄位資訊...等資訊
	f := panel.GetForm()

	// 如果url沒設置__goadmin_iframe則為空值
	isNotIframe := ctx.Query(constant.IframeKey) != "true"

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
		SetPrefix(h.config.PrefixFixSlash()).      // ex:/admin
		SetFieldsHTML(f.HTMLContent).              // ex:""
		SetContent(formInfo.FieldList).            // 將欄位資訊設置至表單的content
		SetTabContents(formInfo.GroupFieldList).   // ex:[]
		SetTabHeaders(formInfo.GroupFieldHeaders). // ex:[]
		SetUrl(newUrl).
		SetAjax(f.AjaxSuccessJS, f.AjaxErrorJS).
		SetInputWidth(f.InputWidth).               // ex:0
		SetHeadWidth(f.HeadWidth).                 // ex:0
		SetLayout(f.Layout).                       // ex:LayoutDefault
		SetPrimaryKey(panel.GetPrimaryKey().Name). // ex:id
		SetHiddenFields(hiddenFields).             // 隱藏資訊(__go_admin_t_、__go_admin_previous_)
		SetTitle("New").
		SetOperationFooter(formFooter("new", f.IsHideContinueEditCheckBox, f.IsHideContinueNewCheckBox,
						f.IsHideResetButton)).
		SetHeader(f.HeaderHtml). // ex:HeaderHtml、FooterHtml為[]
		SetFooter(f.FooterHtml), len(formInfo.GroupFieldHeaders) > 0, !isNotIframe, f.IsHideBackButton, f.Header)

	// 一般不會執行
	if f.Wrapper != nil {
		content = f.Wrapper(content)
	}

	h.HTML(ctx, user, types.Panel{
		Content:     alert + content,
		Description: template2.HTML(f.Description),
		Title:       modules.AorBHTML(isNotIframe, template2.HTML(f.Title), ""),
	}, alert == "")

	// 一般不會執行
	if isNew {
		ctx.AddHeader(constant.PjaxUrlHeader, showNewUrl)
	}
}

// NewForm insert a table row into database.
// 首先處理multipart/form-data設定數值後將資料加入資料表中，接著處理sql語法後接著取得資料表資料後，判斷條件後處理並將值設置至PanelInfo(struct)，panelInfo為頁面上所有的資料
// 最後將所有頁面的HTML處理並回傳(包括標頭、過濾條件、所有顯示的資料)
func (h *Handler) NewForm(ctx *context.Context) {

	// GetNewFormParam回傳Context.UserValue[new_form_param]的值(struct)
	param := guard.GetNewFormParam(ctx)

	// process uploading files, only support local storage
	// 如果有上傳頭像檔案才會執行，否則為空map[]
	if len(param.MultiForm.File) > 0 {
		err := file.GetFileEngine(h.config.FileUploadEngine.Name).Upload(param.MultiForm)
		if err != nil {
			if ctx.WantJSON() {
				response.Error(ctx, err.Error())
			} else {
				h.showNewForm(ctx, aAlert().Warning(err.Error()), param.Prefix, param.Param.GetRouteParamStr(), true)
			}
			return
		}
	}

	// Value回傳NewFormParam.MultiForm.Value
	// param.Value()為multipart/form-data設定的值
	// InsertData在plugins\admin\modules\table\default.go
	// 處理param.Value()(為multipart/form-data設定數值)後將資料加入資料表中
	err := param.Panel.InsertData(param.Value())
	if err != nil {
		if ctx.WantJSON() {
			response.Error(ctx, err.Error())
		} else {
			h.showNewForm(ctx, aAlert().Warning(err.Error()), param.Prefix, param.Param.GetRouteParamStr(), true)
		}
		return
	}

	// 新增頁面都回傳nil
	if param.Panel.GetForm().Responder != nil {
		param.Panel.GetForm().Responder(ctx)
		return
	}

	// 新增頁面都都不會執行
	if ctx.WantJSON() && !param.IsIframe {
		response.OkWithData(ctx, map[string]interface{}{
			"url": param.PreviousPath,
		})
		return
	}

	// 新增頁面都都不會執行
	if !param.FromList {
		if isNewUrl(param.PreviousPath, param.Prefix) {
			h.showNewForm(ctx, param.Alert, param.Prefix, param.Param.GetRouteParamStr(), true)
			return
		}

		ctx.HTML(http.StatusOK, fmt.Sprintf(`<script>location.href="%s"</script>`, param.PreviousPath))
		ctx.AddHeader(constant.PjaxUrlHeader, param.PreviousPath)
		return
	}

	// 新增頁面都都不會執行
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

	// showTable在plugins\admin\controller\show.go
	// 首先透過處理sql語法後接著取得資料表資料後，判斷條件後處理並將值設置至PanelInfo(struct)，PanelInfo裡的資訊有主題、描述名稱、可以篩選條件的欄位、選擇顯示的欄位....等資訊
	// 接著將所有頁面的HTML處理並回傳(包括標頭、過濾條件、所有顯示的資料)
	buf := h.showTable(ctx, param.Prefix, param.Param, nil)

	ctx.HTML(http.StatusOK, buf.String())
	ctx.AddHeader(constant.PjaxUrlHeader, h.routePathWithPrefix("info", param.Prefix)+param.Param.GetRouteParamStr())
}
