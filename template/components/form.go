package components

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	form2 "github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
)

type FormAttribute struct {
	Name            string
	Id              string
	Header          template.HTML
	Content         types.FormFields
	ContentList     []types.FormFields
	Layout          form.Layout
	TabContents     []types.FormFields
	TabHeaders      []string
	Footer          template.HTML
	Url             string
	FieldsHTML      template.HTML
	Method          string
	PrimaryKey      string
	Ajax            bool
	AjaxSuccessJS   template.JS
	AjaxErrorJS     template.JS
	HeadWidth       int
	InputWidth      int
	HiddenFields    map[string]string
	Title           template.HTML
	OperationFooter template.HTML
	Prefix          string
	CdnUrl          string
	types.Attribute
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetHeader(value template.HTML) types.FormAttribute {
	compo.Header = value
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetPrimaryKey(value string) types.FormAttribute {
	compo.PrimaryKey = value
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetContent(value types.FormFields) types.FormAttribute {
	compo.Content = value
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetId(id string) types.FormAttribute {
	compo.Id = id
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetAjax(successJS, errorJS template.JS) types.FormAttribute {
	if successJS != template.JS("") && errorJS != template.JS("") {
		compo.Ajax = true
		compo.AjaxErrorJS = errorJS
		compo.AjaxSuccessJS = successJS
	}
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetTabContents(value []types.FormFields) types.FormAttribute {
	compo.TabContents = value
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetTabHeaders(value []string) types.FormAttribute {
	compo.TabHeaders = value
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetHeadWidth(width int) types.FormAttribute {
	if width > 0 {
		if width > 12 {
			width = 12
		}
		compo.HeadWidth = width
	}
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetInputWidth(width int) types.FormAttribute {
	if width > 0 {
		if width > 12 {
			width = 12
		}
		compo.InputWidth = width
	}
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetFieldsHTML(html template.HTML) types.FormAttribute {
	compo.FieldsHTML = html
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetFooter(value template.HTML) types.FormAttribute {
	compo.Footer = value
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetLayout(layout form.Layout) types.FormAttribute {
	compo.Layout = layout
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetPrefix(value string) types.FormAttribute {
	compo.Prefix = value
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetUrl(value string) types.FormAttribute {
	compo.Url = value
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetHiddenFields(fields map[string]string) types.FormAttribute {
	compo.HiddenFields = fields
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetMethod(value string) types.FormAttribute {
	compo.Method = value
	return compo
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetTitle(value template.HTML) types.FormAttribute {
	compo.Title = value
	return compo
}

// GetDefaultBoxHeader判斷條件(是否隱藏返回按鍵)後取得class="box-title"HTML語法
func (compo *FormAttribute) GetDefaultBoxHeader(hideBack bool) template.HTML {
	if hideBack {

		return template.HTML(fmt.Sprintf(`<h3 class="box-title">%s</h3>`, language.GetFromHtml(compo.Title)))
	}

	return template.HTML(fmt.Sprintf(`<h3 class="box-title">%s</h3>
            <div class="box-tools">
                <div class="btn-group pull-right" style="margin-right: 10px">
                    <a href='%s' class="btn btn-sm btn-default form-history-back"><i
                                class="fa fa-arrow-left"></i> %s</a>
                </div>
            </div>`, language.GetFromHtml(compo.Title), compo.HiddenFields[form2.PreviousKey], language.Get("Back")))
}

// GetDetailBoxHeader(取得細節的BoxHeader)將編輯及刪除按鈕的HTML語法處理後加入class="box-title的HTML語法中
func (compo *FormAttribute) GetDetailBoxHeader(editUrl, deleteUrl string) template.HTML {

	var (
		editBtn   string
		deleteBtn string
	)

	if editUrl != "" {
		editBtn = fmt.Sprintf(`
                <div class="btn-group pull-right" style="margin-right: 10px">
                    <a href='%s' class="btn btn-sm btn-primary"><i
                                class="fa fa-edit"></i> %s</a>
                </div>`, editUrl, language.Get("Edit"))
	}

	if deleteUrl != "" {
		deleteBtn = fmt.Sprintf(`
                <div class="btn-group pull-right" style="margin-right: 10px">
                    <a href='javascript:;' class="btn btn-sm btn-danger delete-btn"><i
                                class="fa fa-trash"></i> %s</a>
                </div>`, language.Get("Delete"))
	}

	return template.HTML(`<h3 class="box-title">`) + language.GetFromHtml(compo.Title) + template.HTML(`</h3>
            <div class="box-tools">
				`+deleteBtn+editBtn+`
                <div class="btn-group pull-right" style="margin-right: 10px">
                    <a href='`+compo.HiddenFields[form2.PreviousKey]+`' class="btn btn-sm btn-default form-history-back"><i
                                class="fa fa-arrow-left"></i> `+language.Get("Back")+`</a>
                </div>
            </div>`)
}

// GetBoxHeaderNoButton(取得BoxHeader不要按鈕)的HTML語法
func (compo *FormAttribute) GetBoxHeaderNoButton() template.HTML {
	return template.HTML(fmt.Sprintf(`<h3 class="box-title">%s</h3>`, language.GetFromHtml(compo.Title)))
}

// 將參數值設置至FormAttribute(struct)
func (compo *FormAttribute) SetOperationFooter(value template.HTML) types.FormAttribute {
	compo.OperationFooter = value
	return compo
}

// 判斷條件後將FormFields添加至FormAttribute.ContentList([]FormFields)
// 首先將符合TreeAttribute.TemplateList["components/多個"](map[string]string)的值加入text(string)
// 接著將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func (compo *FormAttribute) GetContent() template.HTML {
	// GetAssetUrl return globalCfg.AssetUrl
	compo.CdnUrl = config.GetAssetUrl()
	if compo.Id == "" {
		compo.Id = utils.Uuid(10)
	}

	// Col在btemplate\types\form\form.go
	// type Layout uint8
	if col := compo.Layout.Col(); col > 0 {
		compo.ContentList = make([]types.FormFields, col)
		index := 0
		for i := 0; i < len(compo.Content); i++ {
			ii := index % col
			// 將FormFields添加至FormAttribute.ContentList([]FormFields)中
			compo.ContentList[ii] = append(compo.ContentList[ii], compo.Content[i])
			if i < len(compo.Content)-1 {
				if strings.Contains(compo.Content[i+1].Field, "__goadmin_operator__") {
					compo.ContentList[ii] = append(compo.ContentList[ii], compo.Content[i+1])
					i++
				}
			}
			index++
		}
	}

	// 在template\components\composer.go
	// 首先將符合TreeAttribute.TemplateList["components/多個"](map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
	// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
	return ComposeHtml(compo.TemplateList, *compo, "form",
		"form/default", "form/file", "form/multi_file", "form/textarea", "form/custom", "form/rate", "form/slider",
		"form/selectbox", "form/text", "form/table", "form/radio", "form/switch", "form/checkbox", "form/checkbox_single",
		"form/checkbox_stacked", "form/password", "form/code", "form/array", "form/select", "form/singleselect",
		"form/richtext", "form/iconpicker", "form/datetime", "form/number", "form/number_range",
		"form/email", "form/url", "form/ip", "form/color", "form/currency", "form_components", "form/datetime_range",
		"form_layout_default", "form_layout_two_col", "form_layout_tab", "form_components_layout", "form_layout_flow")
}
