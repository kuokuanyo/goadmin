package components

import (
	"github.com/GoAdminGroup/go-admin/modules/menu"
	"github.com/GoAdminGroup/go-admin/template/types"
	"html/template"
)

type TreeAttribute struct {
	Name      string
	Tree      []menu.Item
	EditUrl   string
	DeleteUrl string
	UrlPrefix string
	OrderUrl  string
	types.Attribute
}

// 將參數值設置至TreeAttribute(struct)
func (compo *TreeAttribute) SetTree(value []menu.Item) types.TreeAttribute {
	compo.Tree = value
	return compo
}

// 將參數值設置至TreeAttribute(struct)
func (compo *TreeAttribute) SetEditUrl(value string) types.TreeAttribute {
	compo.EditUrl = value
	return compo
}

// 將參數值設置至TreeAttribute(struct)
func (compo *TreeAttribute) SetUrlPrefix(value string) types.TreeAttribute {
	compo.UrlPrefix = value
	return compo
}

// 將參數值設置至TreeAttribute(struct)
func (compo *TreeAttribute) SetDeleteUrl(value string) types.TreeAttribute {
	compo.DeleteUrl = value
	return compo
}

// 將參數值設置至TreeAttribute(struct)
func (compo *TreeAttribute) SetOrderUrl(value string) types.TreeAttribute {
	compo.OrderUrl = value
	return compo
}

// 首先將符合TreeAttribute.TemplateList["components/tree"](map[string]string)的值加入text(string)
// 接著將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func (compo *TreeAttribute) GetContent() template.HTML {
	// 在template\components\composer.go
	// 首先將符合TreeAttribute.TemplateList["components/tree"](map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
	// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
	return ComposeHtml(compo.TemplateList, *compo, "tree")
}

// 首先將符合TreeAttribute.TemplateList["components/tree-header"](map[string]string)的值加入text(string)
// 接著將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func (compo *TreeAttribute) GetTreeHeader() template.HTML {
	// 在template\components\composer.go
	// 首先將符合TreeAttribute.TemplateList["components/tree-header"](map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
	// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
	return ComposeHtml(compo.TemplateList, *compo, "tree-header")
}
