package components

import (
	"html/template"

	"github.com/GoAdminGroup/go-admin/template/types"
)

type TableAttribute struct {
	Name       string
	Thead      types.Thead
	InfoList   []map[string]types.InfoItem
	Type       string
	PrimaryKey string
	Style      string
	HideThead  bool
	NoAction   bool
	Action     template.HTML
	EditUrl    string
	MinWidth   string
	DeleteUrl  string
	DetailUrl  string
	SortUrl    string
	UpdateUrl  string
	Layout     string
	IsTab      bool
	ExportUrl  string
	types.Attribute
}

// 將參數值設置至TableAttribute(struct)
func (compo *TableAttribute) SetThead(value types.Thead) types.TableAttribute {
	compo.Thead = value
	return compo
}

// 將參數值設置至TableAttribute(struct)
func (compo *TableAttribute) SetInfoList(value []map[string]types.InfoItem) types.TableAttribute {
	compo.InfoList = value
	return compo
}

// 將參數值設置至TableAttribute(struct)
func (compo *TableAttribute) SetType(value string) types.TableAttribute {
	compo.Type = value
	return compo
}

// 將參數值設置至TableAttribute(struct)
func (compo *TableAttribute) SetName(name string) types.TableAttribute {
	compo.Name = name
	return compo
}

// 將參數值設置至TableAttribute(struct)
func (compo *TableAttribute) SetHideThead() types.TableAttribute {
	compo.HideThead = true
	return compo
}

// 將參數值設置至TableAttribute(struct)
func (compo *TableAttribute) SetStyle(style string) types.TableAttribute {
	compo.Style = style
	return compo
}

// 將參數值設置至TableAttribute(struct)
func (compo *TableAttribute) SetMinWidth(value string) types.TableAttribute {
	compo.MinWidth = value
	return compo
}

// 將參數值設置至TableAttribute(struct)
func (compo *TableAttribute) SetLayout(value string) types.TableAttribute {
	compo.Layout = value
	return compo
}

// 判斷條件TableAttribute.MinWidth是否為空，如果為空則設置TableAttribute.MinWidth = "1000px"
// 接著將符合TableAttribute.TemplateList["components/table"](map[string]string)的值加入text(string)
// 接著將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func (compo *TableAttribute) GetContent() template.HTML {
	if compo.MinWidth == "" {
		compo.MinWidth = "1000px"
	}
	return ComposeHtml(compo.TemplateList, *compo, "table")
}

type DataTableAttribute struct {
	TableAttribute
	EditUrl           string
	NewUrl            string
	UpdateUrl         string
	HideThead         bool
	DetailUrl         string
	SortUrl           string
	DeleteUrl         string
	PrimaryKey        string
	IsTab             bool
	ExportUrl         string
	InfoUrl           string
	Buttons           template.HTML
	ActionJs          template.JS
	IsHideFilterArea  bool
	IsHideRowSelector bool
	NoAction          bool
	HasFilter         bool
	Action            template.HTML
	types.Attribute
}

// 首先將符合DataAttribute.TemplateList["components/table/box-header"](map[string]string)的值加入text(string)
// 接著將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func (compo *DataTableAttribute) GetDataTableHeader() template.HTML {
	return ComposeHtml(compo.TemplateList, *compo, "table/box-header")
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetThead(value types.Thead) types.DataTableAttribute {
	compo.Thead = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetLayout(value string) types.DataTableAttribute {
	compo.Layout = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetIsTab(value bool) types.DataTableAttribute {
	compo.IsTab = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetHideThead() types.DataTableAttribute {
	compo.HideThead = true
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetButtons(btns template.HTML) types.DataTableAttribute {
	compo.Buttons = btns
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetHideFilterArea(value bool) types.DataTableAttribute {
	compo.IsHideFilterArea = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetActionJs(aj template.JS) types.DataTableAttribute {
	compo.ActionJs = aj
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetHasFilter(hasFilter bool) types.DataTableAttribute {
	compo.HasFilter = hasFilter
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetInfoUrl(value string) types.DataTableAttribute {
	compo.InfoUrl = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetAction(action template.HTML) types.DataTableAttribute {
	compo.Action = action
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetStyle(style string) types.DataTableAttribute {
	compo.Style = style
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetExportUrl(value string) types.DataTableAttribute {
	compo.ExportUrl = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetHideRowSelector(value bool) types.DataTableAttribute {
	compo.IsHideRowSelector = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetUpdateUrl(value string) types.DataTableAttribute {
	compo.UpdateUrl = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetDetailUrl(value string) types.DataTableAttribute {
	compo.DetailUrl = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetSortUrl(value string) types.DataTableAttribute {
	compo.SortUrl = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetPrimaryKey(value string) types.DataTableAttribute {
	compo.PrimaryKey = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetInfoList(value []map[string]types.InfoItem) types.DataTableAttribute {
	compo.InfoList = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetEditUrl(value string) types.DataTableAttribute {
	compo.EditUrl = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetDeleteUrl(value string) types.DataTableAttribute {
	compo.DeleteUrl = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetNewUrl(value string) types.DataTableAttribute {
	compo.NewUrl = value
	return compo
}

// 將參數值設置至DataAttribute(struct)
func (compo *DataTableAttribute) SetNoAction() types.DataTableAttribute {
	compo.NoAction = true
	return compo
}

// 判斷條件DataTableAttribute.MinWidth是否為空，如果為空則設置DataTableAttribute.MinWidth = "1000px"
// 接著將符合DataTableAttribute.TemplateList["components/table"](map[string]string)的值加入text(string)
// 接著將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func (compo *DataTableAttribute) GetContent() template.HTML {
	if compo.MinWidth == "" {
		compo.MinWidth = "1000px"
	}
	if !compo.NoAction && compo.EditUrl == "" && compo.DeleteUrl == "" && compo.DetailUrl == "" && compo.Action == "" {
		compo.NoAction = true
	}

	return ComposeHtml(compo.TemplateList, *compo, "table")
}
