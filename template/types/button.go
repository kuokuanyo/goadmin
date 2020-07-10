package types

import (
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"html/template"
	"net/url"
)

type Button interface {
	Content() (template.HTML, template.JS)
	GetAction() Action
	URL() string
	METHOD() string
	ID() string
	GetName() string
	SetName(name string)
}

type BaseButton struct {
	Id, Url, Method, Name string
	Title                 template.HTML
	Action                Action
}

func (b *BaseButton) Content() (template.HTML, template.JS) { return "", "" }
func (b *BaseButton) GetAction() Action                     { return b.Action }
func (b *BaseButton) ID() string                            { return b.Id }
func (b *BaseButton) URL() string                           { return b.Url }
func (b *BaseButton) METHOD() string                        { return b.Method }
func (b *BaseButton) GetName() string                       { return b.Name }
func (b *BaseButton) SetName(name string)                   { b.Name = name }

type DefaultButton struct {
	*BaseButton
	Color     template.HTML
	TextColor template.HTML
	Icon      string
	Direction template.HTML
}

func GetDefaultButton(title template.HTML, icon string, action Action, colors ...template.HTML) *DefaultButton {
	return defaultButton(title, "right", icon, action, colors...)
}

func defaultButton(title, direction template.HTML, icon string, action Action, colors ...template.HTML) *DefaultButton {
	id := btnUUID()
	action.SetBtnId(id)

	var color, textColor template.HTML
	if len(colors) > 0 {
		color = colors[0]
	}
	if len(colors) > 1 {
		textColor = colors[1]
	}
	node := action.GetCallbacks()
	return &DefaultButton{
		BaseButton: &BaseButton{
			Id:     id,
			Title:  title,
			Action: action,
			Url:    node.Path,
			Method: node.Method,
		},
		Color:     color,
		TextColor: textColor,
		Icon:      icon,
		Direction: direction,
	}
}

func GetColumnButton(title template.HTML, icon string, action Action, colors ...template.HTML) *DefaultButton {
	return defaultButton(title, "", icon, action, colors...)
}

func (b *DefaultButton) Content() (template.HTML, template.JS) {

	color := template.HTML("")
	if b.Color != template.HTML("") {
		color = template.HTML(`background-color:`) + b.Color + template.HTML(`;`)
	}
	textColor := template.HTML("")
	if b.TextColor != template.HTML("") {
		textColor = template.HTML(`color:`) + b.TextColor + template.HTML(`;`)
	}

	style := template.HTML("")
	addColor := color + textColor

	if addColor != template.HTML("") {
		style = template.HTML(`style="`) + addColor + template.HTML(`"`)
	}

	h := `<div class="btn-group pull-` + b.Direction + `" style="margin-right: 10px">
                <a ` + style + ` class="` + template.HTML(b.Id) + ` btn btn-sm btn-default ` + b.Action.BtnClass() + `" ` + b.Action.BtnAttribute() + `>
                    <i class="fa ` + template.HTML(b.Icon) + `"></i>&nbsp;&nbsp;` + b.Title + `
                </a>
        </div>` + b.Action.ExtContent()
	return h, b.Action.Js()
}

type ActionButton struct {
	*BaseButton
}

func GetActionButton(title template.HTML, action Action, ids ...string) *ActionButton {

	id := ""
	if len(ids) > 0 {
		id = ids[0]
	} else {
		id = "action-info-btn-" + utils.Uuid(10)
	}

	action.SetBtnId(id)
	node := action.GetCallbacks()

	return &ActionButton{
		BaseButton: &BaseButton{
			Id:     id,
			Title:  title,
			Action: action,
			Url:    node.Path,
			Method: node.Method,
		},
	}
}

func (b *ActionButton) Content() (template.HTML, template.JS) {
	h := template.HTML(`<li style="cursor: pointer;"><a data-id="{{.Id}}" class="`+template.HTML(b.Id)+` `+
		b.Action.BtnClass()+`" `+b.Action.BtnAttribute()+`>`+b.Title+`</a></li>`) + b.Action.ExtContent()
	return h, b.Action.Js()
}

type Buttons []Button

func (b Buttons) Content() (template.HTML, template.JS) {
	h := template.HTML("")
	j := template.JS("")

	for _, btn := range b {
		hh, jj := btn.Content()
		h += hh
		j += jj
	}
	return h, j
}

func (b Buttons) FooterContent() template.HTML {
	footer := template.HTML("")

	for _, btn := range b {
		footer += btn.GetAction().FooterContent()
	}
	return footer
}

func (b Buttons) CheckPermission(user models.UserModel) Buttons {
	btns := make(Buttons, 0)
	for _, btn := range b {
		if user.CheckPermissionByUrlMethod(btn.URL(), btn.METHOD(), url.Values{}) {
			btns = append(btns, btn)
		}
	}
	return btns
}

func (b Buttons) AddNavButton(ico, name string, action Action) Buttons {
	if !b.CheckExist(name) {
		return append(b, GetNavButton("", ico, action, name))
	}
	return b
}

func (b Buttons) RemoveButtonByName(name string) Buttons {
	if name == "" {
		return b
	}

	for i := 0; i < len(b); i++ {
		if b[i].GetName() == name {
			b = append(b[:i], b[i+1:]...)
		}
	}
	return b
}

func (b Buttons) CheckExist(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(b); i++ {
		if b[i].GetName() == name {
			return true
		}
	}
	return false
}

const (
	NavBtnSiteName = "go_admin_site_navbtn"
	NavBtnInfoName = "go_admin_info_navbtn"
	NavBtnToolName = "go_admin_tool_navbtn"
)

func (b Buttons) RemoveSiteNavButton() Buttons {
	return b.RemoveButtonByName(NavBtnSiteName)
}

func (b Buttons) RemoveInfoNavButton() Buttons {
	return b.RemoveButtonByName(NavBtnInfoName)
}

func (b Buttons) RemoveToolNavButton() Buttons {
	return b.RemoveButtonByName(NavBtnToolName)
}

type NavButton struct {
	*BaseButton
	Icon string
}

func GetNavButton(title template.HTML, icon string, action Action, names ...string) *NavButton {

	id := btnUUID()
	action.SetBtnId(id)
	node := action.GetCallbacks()
	name := ""

	if len(names) > 0 {
		name = names[0]
	}

	return &NavButton{
		BaseButton: &BaseButton{
			Id:     id,
			Title:  title,
			Action: action,
			Url:    node.Path,
			Method: node.Method,
			Name:   name,
		},
		Icon: icon,
	}
}

func (n *NavButton) Content() (template.HTML, template.JS) {

	ico := template.HTML("")
	title := template.HTML("")

	if n.Icon != "" {
		ico = template.HTML(`<i class="fa ` + n.Icon + `"></i>`)
	}

	if n.Title != "" {
		title = `<span>` + n.Title + `</span>`
	}

	h := template.HTML(`<li>
    <a class="`+template.HTML(n.Id)+` `+n.Action.BtnClass()+`" `+n.Action.BtnAttribute()+`>
      `+ico+`
      `+title+`
    </a>
</li>`) + n.Action.ExtContent()
	return h, n.Action.Js()
}
