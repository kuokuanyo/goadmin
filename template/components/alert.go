package components

import (
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/template/types"
	"html/template"
)

type AlertAttribute struct {
	Name    string
	Theme   string
	Title   template.HTML
	Content template.HTML
	types.Attribute
}

// 將參數值設置至AlertAttribute(struct)
func (compo *AlertAttribute) SetTheme(value string) types.AlertAttribute {
	compo.Theme = value
	return compo
}

// 將參數值設置至AlertAttribute(struct)
func (compo *AlertAttribute) SetTitle(value template.HTML) types.AlertAttribute {
	compo.Title = value
	return compo
}

// 將參數值設置至AlertAttribute(struct)
func (compo *AlertAttribute) SetContent(value template.HTML) types.AlertAttribute {
	compo.Content = value
	return compo
}

// 首先將參數設置至AlertAttribute(struct)後，接著將符合AlertAttribute.TemplateList["components/alert"](map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func (compo *AlertAttribute) Warning(msg string) template.HTML {
	// SetTitle、SetTheme、SetContent將參數設置至AlertAttribute(struct)後
	// GetContent首先將符合AlertAttribute.TemplateList["components/alert"](map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
	// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
	return compo.SetTitle(errors.MsgWithIcon).
		SetTheme("warning").
		SetContent(language.GetFromHtml(template.HTML(msg))).
		GetContent()
}

// 首先將符合AlertAttribute.TemplateList["components/alert"](map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func (compo *AlertAttribute) GetContent() template.HTML {
	// 在template\components\composer.go
	// 首先將符合AlertAttribute.TemplateList["components/alert"](map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
	// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
	return ComposeHtml(compo.TemplateList, *compo, "alert")
}
