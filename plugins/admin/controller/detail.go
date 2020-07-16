package controller

import (
	"fmt"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	form2 "github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

// 首先取得所有欄位的資訊，接著設置按鈕的url，透過id取得資料表中的資料，接著對有帶值的欄位更新並加入FormFields
// 最後匯出HTML語法
func (h *Handler) ShowDetail(ctx *context.Context) {

	var (
		prefix = ctx.Query(constant.PrefixKey) // 從url中取得__prefix的值
		// DetailPKKey = __goadmin_detail_pk
		id = ctx.Query(constant.DetailPKKey) // 從url中取得__goadmin_detail_pk的值
		// BaseTable也屬於Table(interface)
		// 先透過參數prefix取得Table(interface)，接著判斷條件後將[]context.Node加入至Handler.operations後回傳
		panel = h.table(prefix, ctx)
		// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
		user     = auth.Auth(ctx)
		newPanel = panel.Copy()
		// 將參數值設置至base.Detail(InfoPanel(struct)).primaryKey中後回傳InfoPanel(struct)
		detail = panel.GetDetail()
		// 將參數值設置至base.Info(InfoPanel(struct)).primaryKey中後回傳InfoPanel(struct)
		// info為所有顯示的欄位資訊
		info = panel.GetInfo()
		// 將參數值(BaseTable.PrimaryKey)的值設置至BaseTable.Form(FormPanel(struct)).primaryKey中後回傳FormPanel(struct)
		formModel = newPanel.GetForm()
		fieldList = make(types.FieldList, 0)
	)

	// fieldList為所有欄位資訊
	if len(detail.FieldList) == 0 {
		// ------------角色、權限執行----------------
		fieldList = info.FieldList
	} else {
		// ------------用戶執行----------------
		fieldList = detail.FieldList
	}

	formModel.FieldList = make([]types.FormField, len(fieldList))

	// 將fieldList的資訊設置至formModel.FieldList
	for i, field := range fieldList {
		formModel.FieldList[i] = types.FormField{
			Field:        field.Field,
			FieldClass:   field.Field,
			TypeName:     field.TypeName, // 欄位類型
			Head:         field.Head,     // 欄位名稱(中文)
			Hide:         field.Hide,     // 是否隱藏
			Joins:        field.Joins,    // join
			FormType:     form.Default,   // ex: default
			FieldDisplay: field.FieldDisplay,
		}
	}

	// 將資料表名稱設置至formModel.Table
	if detail.Table != "" {
		formModel.Table = detail.Table
	} else {
		// -------用戶、角色、權限都執行---------
		formModel.Table = info.Table
	}

	// GetParam取得頁面size、資料排列方式、選擇欄位...等資訊後設置至Parameters(struct)並回傳
	param := parameter.GetParam(ctx.Request.URL,
		info.DefaultPageSize,
		info.SortField,
		info.GetSort())

	// DeleteDetailPk刪除Parameters.Fields[__goadmin_detail_pk]後回傳
	// GetRouteParamStr取得url.Values(map[string][]string)後加入__page(鍵)與值s，最後編碼並回傳
	// ex: ?__page=1&__pageSize=10&__sort=id&__sort_type=desc(最後要回傳的url)
	paramStr := param.DeleteDetailPk().GetRouteParamStr()

	// 下面三個url會用於頁面上的按鈕(返回、編輯、刪除)
	// AorEmpty判斷第一個(condition)參數，如果true則回傳第二個參數，否則回傳""
	// 編輯頁面url
	editUrl := modules.AorEmpty(!info.IsHideEditButton, h.routePathWithPrefix("show_edit", prefix)+paramStr+
		"&"+constant.EditPKKey+"="+ctx.Query(constant.DetailPKKey)) //ex: /admin/info/manager/edit?__page=1&__pageSize=10&__sort=id&__sort_type=desc&__goadmin_edit_pk=26(用戶)
	// 刪除資料url
	deleteUrl := modules.AorEmpty(!info.IsHideDeleteButton, h.routePathWithPrefix("delete", prefix)+paramStr) // ex:/admin/delete/manager?__page=1&__pageSize=10&__sort=id&__sort_type=desc
	// 回到資料頁面
	infoUrl := h.routePathWithPrefix("info", prefix) + paramStr //ex: /admin/info/manager?__page=1&__pageSize=10&__sort=id&__sort_type=desc
	// 藉由參數檢查權限，如果有權限回傳第一個參數(path)，反之回傳""
	editUrl = user.GetCheckPermissionByUrlMethod(editUrl, h.route("show_edit").Method())
	deleteUrl = user.GetCheckPermissionByUrlMethod(deleteUrl, h.route("delete").Method())

	deleteJs := ""
	// 刪除的js語法
	if deleteUrl != "" {
		deleteJs = fmt.Sprintf(`<script>
function DeletePost(id) {
	swal({
			title: '%s',
			type: "warning",
			showCancelButton: true,
			confirmButtonColor: "#DD6B55",
			confirmButtonText: '%s',
			closeOnConfirm: false,
			cancelButtonText: '%s',
		},
		function () {
			$.ajax({
				method: 'post',
				url: '%s',
				data: {
					id: id
				},
				success: function (data) {
					if (typeof (data) === "string") {
						data = JSON.parse(data);
					}
					if (data.code === 200) {
						location.href = '%s'
					} else {
						swal(data.msg, '', 'error');
					}
				}
			});
		});
}

$('.delete-btn').on('click', function (event) {
	DeletePost(%s)
});

</script>`, language.Get("are you sure to delete"), language.Get("yes"), language.Get("cancel"), deleteUrl, infoUrl, id)
	}

	title := "" // 左上角主題名稱
	desc := ""  // 主題名稱旁的描述

	// 如果url沒有設定__goadmin_iframe參數，回傳空字串
	// IframeKey = __goadmin_iframe
	isNotIframe := ctx.Query(constant.IframeKey) != "true" // ex: true

	if isNotIframe {
		title = detail.Title
		if title == "" {
			title = info.Title + language.Get("Detail")
		}

		desc = detail.Description
		if desc == "" {
			desc = info.Description + language.Get("Detail")
		}
	}

	// WithPKs將參數(多個string)結合並設置至Parameters.Fields["__pk"]後回傳
	// GetDataWithId(透過id取得資料)透過id取得資料表中的資料，接著對有帶值的欄位更新並加入FormFields後回傳，最後設置值至FormInfo(struct)中
	formInfo, err := newPanel.GetDataWithId(param.WithPKs(id))
	if err != nil {
		h.HTML(ctx, user, types.Panel{
			Content:     aAlert().Warning(err.Error()),
			Description: template.HTML(desc),
			Title:       template.HTML(title),
		}, param.Animation)
		return
	}

	h.HTML(ctx, user, types.Panel{
		Content: detailContent(aForm().
			SetTitle(template.HTML(title)). // 表單主題
			SetContent(formInfo.FieldList). // 將欄位資訊及數值設置至表單的content
			SetFooter(template.HTML(deleteJs)).
			SetHiddenFields(map[string]string{ // 設置隱藏的資訊(ex: __go_admin_previous_)
				form2.PreviousKey: infoUrl,
			}).
			SetPrefix(h.config.PrefixFixSlash()), editUrl, deleteUrl, !isNotIframe),
		Description: template.HTML(desc),
		Title:       template.HTML(title),
	}, param.Animation)
}
