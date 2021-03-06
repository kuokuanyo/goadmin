package components

import (
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/menu"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	"html/template"
)

type Base struct {
	Attribute types.Attribute
}

// 建立BoxAttribute(struct)並設置值後回傳
func (b Base) Box() types.BoxAttribute {
	return &BoxAttribute{
		Name:       "box",
		Header:     template.HTML(""),
		Body:       template.HTML(""),
		Footer:     template.HTML(""),
		Title:      "",
		HeadBorder: "",
		Attribute:  b.Attribute,
	}
}

// 建立ColAttribute(struct)並設置值後回傳
func (b Base) Col() types.ColAttribute {
	return &ColAttribute{
		Name:      "col",
		Size:      "col-md-2",
		Content:   "",
		Attribute: b.Attribute,
	}
}

// 建立FormAttribute(struct)並設置值後回傳
func (b Base) Form() types.FormAttribute {
	return &FormAttribute{
		Name:         "form",
		Content:      []types.FormField{},
		Url:          "/",
		Method:       "post",
		HiddenFields: make(map[string]string),
		Layout:       form.LayoutDefault,
		Title:        "edit",
		Attribute:    b.Attribute,
		CdnUrl:       config.GetAssetUrl(),
		HeadWidth:    2,
		InputWidth:   8,
	}
}

// 建立ImgAttribute(struct)並設置值後回傳
func (b Base) Image() types.ImgAttribute {
	return &ImgAttribute{
		Name:      "image",
		Width:     "50",
		Height:    "50",
		Src:       "",
		Attribute: b.Attribute,
	}
}

// 建立TabsAttribute(struct)並設置值後回傳
func (b Base) Tabs() types.TabsAttribute {
	return &TabsAttribute{
		Name:      "tabs",
		Attribute: b.Attribute,
	}
}

// 建立AlertAttribute(struct)並設置值後回傳
func (b Base) Alert() types.AlertAttribute {
	return &AlertAttribute{
		Name:      "alert",
		Attribute: b.Attribute,
	}
}

// 建立LabelAttribute(struct)並設置值後回傳
func (b Base) Label() types.LabelAttribute {
	return &LabelAttribute{
		Name:      "label",
		Type:      "",
		Content:   "",
		Attribute: b.Attribute,
	}
}

// 建立LinkAttribute(struct)並設置值後回傳
func (b Base) Link() types.LinkAttribute {
	return &LinkAttribute{
		Name:      "link",
		Content:   "",
		Attribute: b.Attribute,
	}
}

// 建立PopupAttribute(struct)並設置值後回傳
func (b Base) Popup() types.PopupAttribute {
	return &PopupAttribute{
		Name:      "popup",
		Attribute: b.Attribute,
	}
}

// 建立PaginatorAttribute(struct)並設置值後回傳
func (b Base) Paginator() types.PaginatorAttribute {
	return &PaginatorAttribute{
		Name:      "paginator",
		Attribute: b.Attribute,
	}
}

// 建立RowAttribute(struct)並設置值後回傳
func (b Base) Row() types.RowAttribute {
	return &RowAttribute{
		Name:      "row",
		Content:   "",
		Attribute: b.Attribute,
	}
}

// 建立ButtonAttribute(struct)並設置值後回傳
func (b Base) Button() types.ButtonAttribute {
	return &ButtonAttribute{
		Name:      "button",
		Content:   "",
		Href:      "",
		Attribute: b.Attribute,
	}
}

// 建立TableAttribute(struct)並設置值後回傳
func (b Base) Table() types.TableAttribute {
	return &TableAttribute{
		Name:      "table",
		Thead:     make(types.Thead, 0),
		InfoList:  make([]map[string]types.InfoItem, 0),
		Type:      "table",
		Style:     "hover",
		Layout:    "auto",
		Attribute: b.Attribute,
	}
}

// 建立DataTableAttribute(struct)並設置值後回傳
func (b Base) DataTable() types.DataTableAttribute {
	return &DataTableAttribute{
		TableAttribute: *(b.Table().
			SetStyle("hover").
			SetName("data-table").
			SetType("data-table").(*TableAttribute)),
		Attribute: b.Attribute,
	}
}

// 建立TreeAttribute(struct)並設置值後回傳
func (b Base) Tree() types.TreeAttribute {
	return &TreeAttribute{
		Name:      "tree",
		Tree:      make([]menu.Item, 0),
		Attribute: b.Attribute,
	}
}

// 建立TreeAttribute(struct)並設置值後回傳
func (b Base) TreeView() types.TreeViewAttribute {
	return &TreeViewAttribute{
		Name:      "treeview",
		Attribute: b.Attribute,
	}
}
