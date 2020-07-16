package controller

import (
	"bytes"
	"crypto/md5"
	"fmt"
	template2 "html/template"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/guard"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/icon"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/action"
	"github.com/GoAdminGroup/html"
)

// ShowInfo show info page.
func (h *Handler) ShowInfo(ctx *context.Context) {

	prefix := ctx.Query(constant.PrefixKey)

	panel := h.table(prefix, ctx)

	if panel.GetOnlyUpdateForm() {
		ctx.Redirect(h.routePathWithPrefix("show_edit", prefix))
		return
	}

	if panel.GetOnlyNewForm() {
		ctx.Redirect(h.routePathWithPrefix("show_new", prefix))
		return
	}

	if panel.GetOnlyDetail() {
		ctx.Redirect(h.routePathWithPrefix("detail", prefix))
		return
	}

	params := parameter.GetParam(ctx.Request.URL, panel.GetInfo().DefaultPageSize, panel.GetInfo().SortField,
		panel.GetInfo().GetSort())

	buf := h.showTable(ctx, prefix, params, panel)
	ctx.HTML(http.StatusOK, buf.String())
}

// 首先透過處理sql語法後接著取得資料表資料後，判斷條件後處理並將值設置至PanelInfo(struct)，PanelInfo裡的資訊有主題、描述名稱、可以篩選條件的欄位、選擇顯示的欄位....等資訊
// 最後判斷用戶權限並取得編輯、新增、刪除、輸出...等url資訊儲存至[]string
func (h *Handler) showTableData(ctx *context.Context, prefix string, params parameter.Parameters,
	panel table.Table, urlNamePrefix string) (table.Table, table.PanelInfo, []string, error) {

	// -------用戶編輯介面會執行---------
	if panel == nil {
		// 先透過參數prefix取得Table(interface)，接著判斷條件後將[]context.Node加入至Handler.operations後回傳
		panel = h.table(prefix, ctx)
	}

	// WithIsAll在plugins\admin\modules\parameter\parameter.go
	// WithIsAll判斷條件後，在param.Fields[false](map[string][]string)加入false
	// GetData在\plugins\admin\modules\table\default.go(table、DefaultTable的方法)
	// 透過參數並且將欄位、join語法...等資訊處理後，回傳[]TheadItem、欄位名稱、joinFields(ex:group_concat(goadmin_roles.`name`...)、join語法(left join....)、合併的資料表、可篩選過濾的欄位資訊後
	// 接著取得資料表資料後，判斷條件處理field(struct)數值(DefaultTable.Info.FieldList的迴圈)後將值加入map[string]types.InfoItem後回傳(如果沒有選擇顯示的欄位則直接跳過該欄位)
	// 最後將值設置至PanelInfo(struct)並回傳，PanelInfo裡的資訊有主題、描述名稱、可以篩選條件的欄位、選擇顯示的欄位....等資訊
	panelInfo, err := panel.GetData(params.WithIsAll(false))

	if err != nil {
		return panel, panelInfo, nil, err
	}

	// DeleteIsAll刪除Parameters.Fields(map[string][]string)[__is_all]
	// GetRouteParamStr取得url.Values後加入__page(鍵)與值，最後編碼並回傳
	// 將Parameters.Fields裡的值編碼(ex: ?__page=1&__pageSize=10&__sort=id&__sort_type=desc...)
	paramStr := params.DeleteIsAll().GetRouteParamStr()

	// AorEmpty判斷第一個(condition)參數，如果true則回傳第二個參數，否則回傳""
	// 透過參數name取得該路徑名稱的URL，將url中的:__prefix改成第二個參數(prefix)
	// 第一個參數都在判斷按鈕是否隱藏
	editUrl := modules.AorEmpty(!panel.GetInfo().IsHideEditButton, h.routePathWithPrefix(urlNamePrefix+"show_edit", prefix)+paramStr)  // ex: /admin/info/manager/edit(用戶)
	newUrl := modules.AorEmpty(!panel.GetInfo().IsHideNewButton, h.routePathWithPrefix(urlNamePrefix+"show_new", prefix)+paramStr)     // ex: /admin/info/manager/new(用戶)
	deleteUrl := modules.AorEmpty(!panel.GetInfo().IsHideDeleteButton, h.routePathWithPrefix(urlNamePrefix+"delete", prefix))          // ex: /admin/delete/manager(用戶)
	exportUrl := modules.AorEmpty(!panel.GetInfo().IsHideExportButton, h.routePathWithPrefix(urlNamePrefix+"export", prefix)+paramStr) // ex: /admin/export/manager(用戶)
	detailUrl := modules.AorEmpty(!panel.GetInfo().IsHideDetailButton, h.routePathWithPrefix(urlNamePrefix+"detail", prefix)+paramStr) // ex: /admin/export/manager(用戶)

	// ex: /admin/info/manager(用戶)
	infoUrl := h.routePathWithPrefix(urlNamePrefix+"info", prefix)
	// ex: /admin/update/manager(用戶)
	updateUrl := h.routePathWithPrefix(urlNamePrefix+"update", prefix)

	// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
	// 取得權限、角色、可用menu
	user := auth.Auth(ctx)

	// 藉由參數檢查權限，如果有權限回傳第一個參數(path)，反之回傳""
	editUrl = user.GetCheckPermissionByUrlMethod(editUrl, h.route(urlNamePrefix+"show_edit").Method())  //get
	newUrl = user.GetCheckPermissionByUrlMethod(newUrl, h.route(urlNamePrefix+"show_new").Method())     //get
	deleteUrl = user.GetCheckPermissionByUrlMethod(deleteUrl, h.route(urlNamePrefix+"delete").Method()) //post
	exportUrl = user.GetCheckPermissionByUrlMethod(exportUrl, h.route(urlNamePrefix+"export").Method()) //post
	detailUrl = user.GetCheckPermissionByUrlMethod(detailUrl, h.route(urlNamePrefix+"detail").Method()) //get

	return panel, panelInfo, []string{editUrl, newUrl, deleteUrl, exportUrl, detailUrl, infoUrl, updateUrl}, nil
}

// 首先透過處理sql語法後接著取得資料表資料後，判斷條件後處理並將值設置至PanelInfo(struct)，panelInfo裡的資訊有主題、描述名稱、可以篩選條件的欄位、選擇顯示的欄位....等資訊
// 接著將所有頁面的HTML處理並回傳(包括標頭、過濾條件、所有顯示的資料)
func (h *Handler) showTable(ctx *context.Context, prefix string, params parameter.Parameters, panel table.Table) *bytes.Buffer {

	// 首先透過處理sql語法後接著取得資料表資料後，判斷條件後處理並將值設置至PanelInfo(struct)，PanelInfo裡的資訊有主題、描述名稱、可以篩選條件的欄位、選擇顯示的欄位....等資訊
	// 最後判斷用戶權限並取得編輯、新增、刪除、輸出...等url資訊儲存至[]string
	// panelInfo為頁面上所有的資料...等資料
	panel, panelInfo, urls, err := h.showTableData(ctx, prefix, params, panel, "")

	if err != nil {
		return h.Execute(ctx, auth.Auth(ctx), types.Panel{
			Content: aAlert().SetTitle(errors.MsgWithIcon).
				SetTheme("warning").
				SetContent(template2.HTML(err.Error())).
				GetContent(),
			Description: template2.HTML(errors.Msg),
			Title:       template2.HTML(errors.Msg),
		}, params.Animation)
	}

	editUrl, newUrl, deleteUrl, exportUrl, detailUrl, infoUrl, updateUrl := urls[0], urls[1], urls[2], urls[3], urls[4], urls[5], urls[6]

	// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
	// 取得權限、角色、可用menu
	user := auth.Auth(ctx)

	var (
		body       template2.HTML
		dataTable  types.DataTableAttribute
		info       = panel.GetInfo()
		actionBtns = info.Action
		actionJs   template2.JS
		allBtns    = make(types.Buttons, 0)
	)

	// 新增、編輯頁面的InfoPanel.Buttons為空
	for _, b := range info.Buttons {
		if b.URL() == "" || b.METHOD() == "" || user.CheckPermissionByUrlMethod(b.URL(), b.METHOD(), url.Values{}) {
			allBtns = append(allBtns, b)
		}
	}

	// 取得HTML及JSON
	// 上面為空因此這裡也是空值
	btns, btnsJs := allBtns.Content()

	allActionBtns := make(types.Buttons, 0)

	// 用戶頁面的info.ActionButtons為空
	for _, b := range info.ActionButtons {
		if b.URL() == "" || b.METHOD() == "" || user.CheckPermissionByUrlMethod(b.URL(), b.METHOD(), url.Values{}) {
			allActionBtns = append(allActionBtns, b)
		}
	}

	// 上面為空，因此這裡不執行
	if actionBtns == template.HTML("") && len(allActionBtns) > 0 {
		ext := template.HTML("")
		if deleteUrl != "" {
			ext = html.LiEl().SetClass("divider").Get()
			allActionBtns = append([]types.Button{types.GetActionButton(language.GetFromHtml("delete"),
				types.NewDefaultAction(`data-id='{{.Id}}' style="cursor: pointer;"`,
					ext, "", ""), "grid-row-delete")}, allActionBtns...)
		}
		ext = template.HTML("")
		if detailUrl != "" {
			if editUrl == "" && deleteUrl == "" {
				ext = html.LiEl().SetClass("divider").Get()
			}
			allActionBtns = append([]types.Button{types.GetActionButton(language.GetFromHtml("detail"),
				action.Jump(detailUrl+"&"+constant.DetailPKKey+"={{.Id}}", ext))}, allActionBtns...)
		}
		if editUrl != "" {
			if detailUrl == "" && deleteUrl == "" {
				ext = html.LiEl().SetClass("divider").Get()
			}
			allActionBtns = append([]types.Button{types.GetActionButton(language.GetFromHtml("edit"),
				action.Jump(editUrl+"&"+constant.EditPKKey+"={{.Id}}", ext))}, allActionBtns...)
		}

		var content template2.HTML
		content, actionJs = allActionBtns.Content()

		actionBtns = html.Div(
			html.A(icon.Icon(icon.EllipsisV),
				html.M{"color": "#676565"},
				html.M{"class": "dropdown-toggle", "href": "#", "data-toggle": "dropdown"},
			)+html.Ul(content,
				html.M{"min-width": "20px !important", "left": "-32px", "overflow": "hidden"},
				html.M{"class": "dropdown-menu", "role": "menu", "aria-labelledby": "dLabel"}),
			html.M{"text-align": "center"}, html.M{"class": "dropdown"})
	}

	// ------用戶頁面為false-------------
	// Valid判斷TabGroups([][]string)是否長度大於0
	if info.TabGroups.Valid() {
		// aDataTable在plugins\admin\controller\show.go
		dataTable = aDataTable().
			// 將Thead設置至DataTableAttribute.Thead
			SetThead(panelInfo.Thead).
			// 將參數設置至DataTableAttribute.DeleteUrl
			SetDeleteUrl(deleteUrl).
			// 將參數設置至DataTableAttribute.NewUrl
			SetNewUrl(newUrl).
			// 將參數設置至ExportUrl
			SetExportUrl(exportUrl)

		var (
			tabsHtml    = make([]map[string]template2.HTML, len(info.TabHeaders))
			infoListArr = panelInfo.InfoList.GroupBy(info.TabGroups)
			theadArr    = panelInfo.Thead.GroupBy(info.TabGroups)
		)
		for key, header := range info.TabHeaders {
			tabsHtml[key] = map[string]template2.HTML{
				"title": template2.HTML(header),
				"content": aDataTable().
					SetInfoList(infoListArr[key]).
					SetInfoUrl(infoUrl).
					SetButtons(btns).
					SetActionJs(btnsJs + actionJs).
					SetHasFilter(len(panelInfo.FilterFormData) > 0).
					SetAction(actionBtns).
					SetIsTab(key != 0).
					SetPrimaryKey(panel.GetPrimaryKey().Name).
					SetThead(theadArr[key]).
					SetHideRowSelector(info.IsHideRowSelector).
					SetLayout(info.TableLayout).
					SetExportUrl(exportUrl).
					SetNewUrl(newUrl).
					SetSortUrl(params.GetFixedParamStrWithoutSort()).
					SetEditUrl(editUrl).
					SetUpdateUrl(updateUrl).
					SetDetailUrl(detailUrl).
					SetDeleteUrl(deleteUrl).
					GetContent(),
			}
		}
		body = aTab().SetData(tabsHtml).GetContent()
	} else {
		// aDataTable在plugins\admin\controller\show.go
		// 將參數值設置至DataAttribute(struct)
		// ---------新增、編輯頁面都執行-------------
		dataTable = aDataTable().
			SetInfoList(panelInfo.InfoList).                 // panelInfo.InfoList為頁面上所有取得的資料數值
			SetInfoUrl(infoUrl).                             // ex: /admin/info/manager(用戶)
			SetButtons(btns).                                // ex:空
			SetLayout(info.TableLayout).                     // ex:auto
			SetActionJs(btnsJs + actionJs).                  // ex:空
			SetAction(actionBtns).                           // ex:空
			SetHasFilter(len(panelInfo.FilterFormData) > 0). // ex:true(有可以篩選的條件)
			SetPrimaryKey(panel.GetPrimaryKey().Name).       // ex: id
			SetThead(panelInfo.Thead).                       // 頁面所有的欄位資訊(是否可排列、編輯、隱藏....等資訊)
			SetExportUrl(exportUrl).                         // ex: /admin/export/manager?__page=1&__pageSize=10&__sort=id&__sort_type=desc(用戶)
			SetHideRowSelector(info.IsHideRowSelector).      // ex: false(沒有隱藏選擇器)
			SetHideFilterArea(info.IsHideFilterArea).        // ex: true(隱藏過濾條件)
			SetNewUrl(newUrl).                               // ex: /admin/info/manager/new?__page=1&__pageSize=10&__sort=id&__sort_type=desc(用戶)
			SetEditUrl(editUrl).                             // ex: /admin/info/manager/edit?__page=1&__pageSize=10&__sort=id&__sort_type=desc(用戶)
			// 將__pageSize、__go_admin_no_animation_...等資訊加入url.Values(map[string][]string)後編碼回傳
			SetSortUrl(params.GetFixedParamStrWithoutSort()). // ex: &__go_admin_no_animation_=true&__pageSize=10
			SetUpdateUrl(updateUrl).                          // ex: /admin/update/manager(用戶)
			SetDetailUrl(detailUrl).                          // ex: /admin/info/manager/detail?__page=1&__pageSize=10&__sort=id&__sort_type=desc(用戶)
			SetDeleteUrl(deleteUrl)                           // ex: /admin/delete/manager(用戶)

		// 判斷條件DataTableAttribute.MinWidth是否為空，如果為空則設置DataTableAttribute.MinWidth = "1000px"
		// 接著將符合DataTableAttribute.TemplateList["table"](map[string]string)的值加入text(string)
		// 接著將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
		// 為尋找{{define "table"}}HTML語法(將所有資料數值也加入HTML中)
		body = dataTable.GetContent()

	}

	// IframeKey = __goadmin_iframe
	// Query取得Request url(在url中)裡的參數(key)
	isNotIframe := ctx.Query(constant.IframeKey) != "true" // ex: true(用戶)

	paginator := panelInfo.Paginator // 分頁器語法

	if !isNotIframe {
		paginator = paginator.SetHideEntriesInfo()
	}

	boxModel := aBox().
		SetBody(body). // 將上面取得的body設置至box的body
		// 將padding:0設置至BoxAttribute(struct).Padding
		SetNoPadding().
		// GetDataTableHeader首先將符合DataAttribute.TemplateList["components/table/box-header"](map[string]string)的值加入text(string)
		// 接著將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML(設置在header)
		// 還沒設定篩選條件的語法
		SetHeader(dataTable.GetDataTableHeader() + info.HeaderHtml).
		// 將"with-border"設置至BoxAttribute(struct).SecondHeadBorder
		WithHeadBorder().
		SetIframeStyle(!isNotIframe).
		// 設定分頁器語法
		SetFooter(paginator.GetContent() + info.FooterHtml)

	// 加入可篩選條件語法至第二個標頭("filter-area")
	if len(panelInfo.FilterFormData) > 0 {
		boxModel = boxModel.SetSecondHeaderClass("filter-area").
			SetSecondHeader(aForm().
				SetContent(panelInfo.FilterFormData).     // 可以篩選的欄位資訊
				SetPrefix(h.config.PrefixFixSlash()).     // ex: /admin
				SetInputWidth(info.FilterFormInputWidth). //ex: 10
				SetHeadWidth(info.FilterFormHeadWidth).   //ex: 2
				SetMethod("get").
				SetLayout(info.FilterFormLayout). // ex:LayoutDefault
				SetUrl(infoUrl).                  //  + params.GetFixedParamStrWithoutColumnsAndPage()
				SetHiddenFields(map[string]string{
					form.NoAnimationKey: "true",
				}).
				SetOperationFooter(filterFormFooter(infoUrl)).
				GetContent())
	}

	// 接著將符合TreeAttribute.TemplateList["components/box"](map[string]string)的值加入text(string)
	// 最後將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
	content := boxModel.GetContent()

	// ------------用戶介面不會執行------------------
	if info.Wrapper != nil {
		content = info.Wrapper(content)
	}

	// 將參數設置至ExecuteParam(struct)，接著將給定的數據(types.Page(struct))寫入buf(struct)並回傳
	return h.Execute(ctx, user, types.Panel{
		Content:     content,
		Description: template2.HTML(panelInfo.Description),
		Title:       modules.AorBHTML(isNotIframe, template2.HTML(panelInfo.Title), ""),
	}, params.Animation)
}

// Assets return front-end assets according the request path.
// 處理前端檔案
func (h *Handler) Assets(ctx *context.Context) {
	// URLRemovePrefix將URL的前綴(ex:/admin)去除
	filepath := h.config.URLRemovePrefix(ctx.Path())

	// aTemplate判斷templateMap(map[string]Template)的key鍵是否參數globalCfg.Theme，有則回傳Template(interface)
	// GetAsset對map[string]Component迴圈，對每一個Component(interface)執行GetAsset方法
	data, err := aTemplate().GetAsset(filepath)

	if err != nil {
		data, err = template.GetAsset(filepath)
		if err != nil {
			logger.Error("asset err", err)
			ctx.Write(http.StatusNotFound, map[string]string{}, "")
			return
		}
	}

	var contentType = mime.TypeByExtension(path.Ext(filepath))

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	etag := fmt.Sprintf("%x", md5.Sum(data))

	// 藉由參數key獲得Header
	if match := ctx.Headers("If-None-Match"); match != "" {
		if strings.Contains(match, etag) {
			ctx.SetStatusCode(http.StatusNotModified)
			return
		}
	}

	// 將code, headers and body(參數)存在Context.Response中
	ctx.DataWithHeaders(http.StatusOK, map[string]string{
		"Content-Type":   contentType,
		"Cache-Control":  "max-age=2592000",
		"Content-Length": strconv.Itoa(len(data)),
		"ETag":           etag,
	}, data)
}

// Export export table rows as excel object.
// 建立一個excel檔接著取得所有匯出的資料，最後將值加入至excel中
func (h *Handler) Export(ctx *context.Context) {
	// 取得Context.UserValue[export_param]的值並轉換成ExportParam(struct)
	param := guard.GetExportParam(ctx)

	tableName := "Sheet1"
	// PrefixKey = __prefix
	// prefix = manager、roles、permission
	prefix := ctx.Query(constant.PrefixKey)

	// BaseTable也屬於Table(interface)
	// 先透過參數prefix取得Table(interface)，接著判斷條件後將[]context.Node加入至Handler.operations後回傳
	panel := h.table(prefix, ctx)

	// 創建一個file
	f := excelize.NewFile()
	// 新增一個sheet名為Sheet1
	index := f.NewSheet(tableName)
	// 設定一個預設的工作表(sheet)
	f.SetActiveSheet(index)

	// TODO: support any numbers of fields.
	orders := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K",
		"L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	var (
		infoData table.PanelInfo
		fileName string
		err      error
		// 將參數值設置至base.Info(InfoPanel(struct)).primaryKey中後回傳InfoPanel(struct)
		tableInfo = panel.GetInfo()
	)

	// 判斷是否有選擇匯出特定資料，如選擇當頁或全部(則Id為空)
	if len(param.Id) == 0 {
		// GetParam取得頁面size、資料排列方式、選擇欄位...等資訊後設置至Parameters(struct)並回傳
		params := parameter.GetParam(ctx.Request.URL, tableInfo.DefaultPageSize, tableInfo.SortField,
			tableInfo.GetSort())

		// 透過參數處理sql語法後取得資料表資料並將值設置至PanelInfo(struct)後回傳，PanelInfo裡的資訊有主題、描述名稱、可以篩選條件的欄位、選擇顯示的欄位....等資訊
		infoData, err = panel.GetData(params.WithIsAll(param.IsAll))

		// ex: 权限管理-1594877943-page-1-pageSize-10.xlsx
		fileName = fmt.Sprintf("%s-%d-page-%s-pageSize-%s.xlsx", tableInfo.Title, time.Now().Unix(),
			params.Page, params.PageSize)
	} else {
		// 選擇匯出特定資料
		// 透過參數(選擇取得特定id資料)處理sql語法後取得資料表資料並將值設置至PanelInfo(struct)後回傳，PanelInfo裡的資訊有主題、描述名稱、可以篩選條件的欄位、選擇顯示的欄位....等資訊
		infoData, err = panel.GetDataWithIds(parameter.GetParam(ctx.Request.URL,
			// WithPKs將參數(多個string)結合並設置至Parameters.Fields["__pk"]後回傳
			tableInfo.DefaultPageSize, tableInfo.SortField, tableInfo.GetSort()).WithPKs(param.Id...))

		// ex:权限管理-1594876892-id-40_39_38.xlsx
		fileName = fmt.Sprintf("%s-%d-id-%s.xlsx", tableInfo.Title, time.Now().Unix(), strings.Join(param.Id, "_"))
	}
	if err != nil {
		response.Error(ctx, "export error")
		return
	}

	columnIndex := 0
	// PanelInfo.Thead為頁面中所有的欄位資訊
	for _, head := range infoData.Thead {
		// 如果不隱藏欄位則執行
		// 將欄位名稱設置至excel
		if !head.Hide {
			f.SetCellValue(tableName, orders[columnIndex]+"1", head.Head)
			columnIndex++
		}
	}

	count := 2
	// PanelInfo.InfoList為所有取得的資料
	for _, info := range infoData.InfoList {
		columnIndex = 0
		for _, head := range infoData.Thead {
			// 如果不隱藏欄位則執行
			// 將數值設置至excel
			if !head.Hide {
				if tableInfo.IsExportValue() {
					f.SetCellValue(tableName, orders[columnIndex]+strconv.Itoa(count), info[head.Field].Value)
				} else {
					f.SetCellValue(tableName, orders[columnIndex]+strconv.Itoa(count), info[head.Field].Content)
				}
				columnIndex++
			}
		}
		count++
	}

	buf, err := f.WriteToBuffer()

	if err != nil || buf == nil {
		response.Error(ctx, "export error")
		return
	}

	ctx.AddHeader("content-disposition", `attachment; filename=`+fileName)
	ctx.Data(200, "application/vnd.ms-excel", buf.Bytes())
}
