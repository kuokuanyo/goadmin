package components

import (
	"github.com/GoAdminGroup/go-admin/template/types"
	"html/template"
)

type RowAttribute struct {
	Name    string
	Content template.HTML
	types.Attribute
}

// 將參數值設置至RowAttribute(struct)
func (compo *RowAttribute) SetContent(value template.HTML) types.RowAttribute {
	compo.Content = value
	return compo
}

// 將參數值設置至RowAttribute(struct)
func (compo *RowAttribute) AddContent(value template.HTML) types.RowAttribute {
	compo.Content += value
	return compo
}

// 首先將符合TreeAttribute.TemplateList["components/tree-header"](map[string]string)的值加入text(string)
// 接著將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func (compo *RowAttribute) GetContent() template.HTML {
	// 在template\components\composer.go
	// 首先將符合TreeAttribute.TemplateList["components/row"](map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
	// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
	return ComposeHtml(compo.TemplateList, *compo, "row")
}
