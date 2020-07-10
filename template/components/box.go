package components

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/template/types"
	"html/template"
)

type BoxAttribute struct {
	Name              string
	Header            template.HTML
	Body              template.HTML
	Footer            template.HTML
	Title             template.HTML
	Theme             string
	HeadBorder        string
	Attr              template.HTMLAttr
	HeadColor         string
	SecondHeaderClass string
	SecondHeader      template.HTML
	SecondHeadBorder  string
	SecondHeadColor   string
	Style             template.HTMLAttr
	Padding           string
	types.Attribute
}

// 將參數值設置至BoxAttribute(struct)
func (compo *BoxAttribute) SetTheme(value string) types.BoxAttribute {
	compo.Theme = value
	return compo
}

// 將參數值設置至BoxAttribute(struct)
func (compo *BoxAttribute) SetHeader(value template.HTML) types.BoxAttribute {
	compo.Header = value
	return compo
}

// 將參數值設置至BoxAttribute(struct)
func (compo *BoxAttribute) SetBody(value template.HTML) types.BoxAttribute {
	compo.Body = value
	return compo
}

// 將參數值設置至BoxAttribute(struct)
func (compo *BoxAttribute) SetStyle(value template.HTMLAttr) types.BoxAttribute {
	compo.Style = value
	return compo
}

// 將參數值設置至BoxAttribute(struct)
func (compo *BoxAttribute) SetAttr(attr template.HTMLAttr) types.BoxAttribute {
	compo.Attr = attr
	return compo
}

// 判斷條件後將值設置至BoxAttribute(struct).Attr
func (compo *BoxAttribute) SetIframeStyle(iframe bool) types.BoxAttribute {
	if iframe {
		compo.Attr = `style="border-radius: 0px;box-shadow:none;border-top:none;margin-bottom: 0px;"`
	}
	return compo
}

// 將參數值設置至BoxAttribute(struct)
func (compo *BoxAttribute) SetFooter(value template.HTML) types.BoxAttribute {
	compo.Footer = value
	return compo
}

// 將參數值設置至BoxAttribute(struct)
func (compo *BoxAttribute) SetTitle(value template.HTML) types.BoxAttribute {
	compo.Title = value
	return compo
}

// 將參數值設置至BoxAttribute(struct)
func (compo *BoxAttribute) SetHeadColor(value string) types.BoxAttribute {
	compo.HeadColor = value
	return compo
}

// 將"with-border"設置至BoxAttribute(struct).SecondHeadBorder
func (compo *BoxAttribute) WithHeadBorder() types.BoxAttribute {
	compo.HeadBorder = "with-border"
	return compo
}

// 將參數值設置至BoxAttribute(struct)
func (compo *BoxAttribute) SetSecondHeader(value template.HTML) types.BoxAttribute {
	compo.SecondHeader = value
	return compo
}

// 將參數值設置至BoxAttribute(struct)
func (compo *BoxAttribute) SetSecondHeadColor(value string) types.BoxAttribute {
	compo.SecondHeadColor = value
	return compo
}

// 將參數值設置至BoxAttribute(struct)
func (compo *BoxAttribute) SetSecondHeaderClass(value string) types.BoxAttribute {
	compo.SecondHeaderClass = value
	return compo
}

// 將padding:0設置至BoxAttribute(struct).Padding
func (compo *BoxAttribute) SetNoPadding() types.BoxAttribute {
	compo.Padding = "padding:0;"
	return compo
}

// 將"with-border"設置至BoxAttribute(struct).SecondHeadBorder
func (compo *BoxAttribute) WithSecondHeadBorder() types.BoxAttribute {
	compo.SecondHeadBorder = "with-border"
	return compo
}

// 先依條件判斷並設置BoxAttribute.Style
// 接著將符合TreeAttribute.TemplateList["components/box"](map[string]string)的值加入text(string)
// 最後將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func (compo *BoxAttribute) GetContent() template.HTML {

	// 判斷條件後設置BoxAttribute.Style
	if compo.Style == "" {
		compo.Style = template.HTMLAttr(fmt.Sprintf(`style="overflow-x: scroll;overflow-y: hidden;%s"`, compo.Padding))
	} else {
		compo.Style = template.HTMLAttr(fmt.Sprintf(`style="%s"`, string(compo.Style)+compo.Padding))
	}

	// 在template\components\composer.go
	// 首先將符合TreeAttribute.TemplateList["components/box"](map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
	// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
	return ComposeHtml(compo.TemplateList, *compo, "box")
}
