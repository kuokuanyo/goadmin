package components

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/template/icon"
	"github.com/GoAdminGroup/go-admin/template/types"
	"html/template"
)

type ButtonAttribute struct {
	Name        string
	Content     template.HTML
	Orientation string
	LoadingText template.HTML
	Theme       string
	Type        string
	Size        string
	Href        string
	Style       template.HTMLAttr
	MarginLeft  int
	MarginRight int
	types.Attribute
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetContent(value template.HTML) types.ButtonAttribute {
	compo.Content = value
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetOrientationRight() types.ButtonAttribute {
	compo.Orientation = "pull-right"
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetOrientationLeft() types.ButtonAttribute {
	compo.Orientation = "pull-left"
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetMarginLeft(px int) types.ButtonAttribute {
	compo.MarginLeft = px
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetSmallSize() types.ButtonAttribute {
	compo.Size = "btn-sm"
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetMiddleSize() types.ButtonAttribute {
	compo.Size = "btn-md"
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetMarginRight(px int) types.ButtonAttribute {
	compo.MarginRight = px
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetLoadingText(value template.HTML) types.ButtonAttribute {
	compo.LoadingText = value
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetThemePrimary() types.ButtonAttribute {
	compo.Theme = "primary"
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetThemeDefault() types.ButtonAttribute {
	compo.Theme = "default"
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetThemeWarning() types.ButtonAttribute {
	compo.Theme = "warning"
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetHref(href string) types.ButtonAttribute {
	compo.Href = href
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetTheme(value string) types.ButtonAttribute {
	compo.Theme = value
	return compo
}

// 將參數設置至ButtonAttribute(struct)
func (compo *ButtonAttribute) SetType(value string) types.ButtonAttribute {
	compo.Type = value
	return compo
}

// 首先處理ButtonAttribute.Style與ButtonAttribute.LoadingText後，接著將符合ButtonAttribute.TemplateList["components/button"](map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func (compo *ButtonAttribute) GetContent() template.HTML {

	if compo.MarginLeft != 0 {
		compo.Style = template.HTMLAttr(fmt.Sprintf(`style="margin-left:%dpx;"`, compo.MarginLeft))
	}

	if compo.MarginRight != 0 {
		compo.Style = template.HTMLAttr(fmt.Sprintf(`style="margin-right:%dpx;"`, compo.MarginRight))
	}

	if compo.LoadingText == "" {
		compo.LoadingText = icon.Icon(icon.Spinner, 1) + language.GetFromHtml(`Save`)
	}

	// 在template\components\composer.go
	// 首先將符合ButtonAttribute.TemplateList["components/button"](map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
	// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
	return ComposeHtml(compo.TemplateList, *compo, "button")
}
