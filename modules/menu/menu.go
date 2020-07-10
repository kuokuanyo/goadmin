// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package menu

import (
	"html/template"
	"regexp"
	"strconv"

	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
)

// Item is an menu item.
// Item為資料表goadmin_menu的欄位
type Item struct {
	Name         string
	ID           string
	Url          string
	Icon         string
	Header       string
	Active       string
	ChildrenList []Item
}

// Menu contains list of menu items and other info.
type Menu struct {
	List     []Item
	Options  []map[string]string
	MaxOrder int64
}

// SetMaxOrder set the max order of menu.
// 透過參數order設置Menu.MaxOrder
func (menu *Menu) SetMaxOrder(order int64) {
	menu.MaxOrder = order
}

// AddMaxOrder add the max order of menu.
// Menu.MaxOrder+1
func (menu *Menu) AddMaxOrder() {
	menu.MaxOrder++
}

// SetActiveClass set the active class of menu.
// 設定menu的active
func (menu *Menu) SetActiveClass(path string) *Menu {

	reg, _ := regexp.Compile(`\?(.*)`)
	path = reg.ReplaceAllString(path, "")

	for i := 0; i < len(menu.List); i++ {
		menu.List[i].Active = ""
	}

	for i := 0; i < len(menu.List); i++ {
		if menu.List[i].Url == path && len(menu.List[i].ChildrenList) == 0 {
			menu.List[i].Active = "active"
			return menu
		}

		for j := 0; j < len(menu.List[i].ChildrenList); j++ {
			if menu.List[i].ChildrenList[j].Url == path {
				menu.List[i].Active = "active"
				menu.List[i].ChildrenList[j].Active = "active"
				return menu
			}

			menu.List[i].Active = ""
			menu.List[i].ChildrenList[j].Active = ""
		}
	}

	return menu
}

// FormatPath get template.HTML for front-end.
func (menu Menu) FormatPath() template.HTML {
	res := template.HTML(``)
	for i := 0; i < len(menu.List); i++ {
		if menu.List[i].Active != "" {
			if menu.List[i].Url != "#" && menu.List[i].Url != "" && len(menu.List[i].ChildrenList) > 0 {
				res += template.HTML(`<li><a href="` + menu.List[i].Url + `">` + menu.List[i].Name + `</a></li>`)
			} else {
				res += template.HTML(`<li>` + menu.List[i].Name + `</li>`)
				if len(menu.List[i].ChildrenList) == 0 {
					return res
				}
			}
			for j := 0; j < len(menu.List[i].ChildrenList); j++ {
				if menu.List[i].ChildrenList[j].Active != "" {
					return res + template.HTML(`<li>`+menu.List[i].ChildrenList[j].Name+`</li>`)
				}
			}
		}
	}
	return res
}

// GetEditMenuList return menu items list.
// 回傳Menu.List
func (menu *Menu) GetEditMenuList() []Item {
	return menu.List
}

// GetGlobalMenu return Menu of given user model.
// 回傳參數user(struct)的Menu(設置menuList、menuOption、MaxOrder)
func GetGlobalMenu(user models.UserModel, conn db.Connection) *Menu {

	var (
		menus      []map[string]interface{}
		menuOption = make([]map[string]string, 0)
	)

	// WithRoles查詢role藉由user
	// WithMenus查詢menu藉由user
	user.WithRoles().WithMenus()

	//是否為超級管理員
	if user.IsSuperAdmin() {
		// 在modules\db\statement.go中，回傳SQL(struct)藉由給定的conn
		// 取得多筆資料(利用where、order等篩選)
		// 篩選函式在modules\db\statement.go
		// menus依照order欄位升冪排列
		menus, _ = db.WithDriver(conn).Table("goadmin_menu").
			Where("id", ">", 0).
			OrderBy("order", "asc").
			All()
	} else {
		//不是超級管理員
		var ids []interface{}
		for i := 0; i < len(user.MenuIds); i++ {
			ids = append(ids, user.MenuIds[i])
		}

		// menus依照order欄位升冪排列
		menus, _ = db.WithDriver(conn).Table("goadmin_menu").
			WhereIn("id", ids).
			OrderBy("order", "asc").
			All()
	}

	var title string
	for i := 0; i < len(menus); i++ {

		// 如果type = 1，將title轉換成該語言文字(例如將Perssion轉換為中文)
		if menus[i]["type"].(int64) == 1 {
			// Get回傳設定的語言(modules\language\language.go)
			title = language.Get(menus[i]["title"].(string))
		} else {
			title = menus[i]["title"].(string)
		}

		menuOption = append(menuOption, map[string]string{
			"id":    strconv.FormatInt(menus[i]["id"].(int64), 10),
			"title": title,
		})
	}

	// 將參數menus轉換[]Item(Item是資料表goadmin_menu的欄位)
	menuList := constructMenuTree(menus, 0)

	return &Menu{
		//清單
		List:     menuList, // 所有菜單資訊([]Item)，Item為資料表goadmin_menu的欄位
		//選項
		Options:  menuOption, // 設置每個菜單的id、title
		MaxOrder: menus[len(menus)-1]["parent_id"].(int64),
	}
}

// 將參數menus轉換成[]Item類別(Item(struct)是資料表goadmin_menu的欄位)
func constructMenuTree(menus []map[string]interface{}, parentID int64) []Item {

	branch := make([]Item, 0)

	var title string
	for j := 0; j < len(menus); j++ {
		if parentID == menus[j]["parent_id"].(int64) {
			if menus[j]["type"].(int64) == 1 {
				title = language.Get(menus[j]["title"].(string))
			} else {
				title = menus[j]["title"].(string)
			}

			header, _ := menus[j]["header"].(string)

			child := Item{
				Name:         title,
				ID:           strconv.FormatInt(menus[j]["id"].(int64), 10),
				Url:          menus[j]["uri"].(string),
				Icon:         menus[j]["icon"].(string),
				Header:       header,
				Active:       "",
				ChildrenList: constructMenuTree(menus, menus[j]["id"].(int64)),
			}

			branch = append(branch, child)
		}
	}

	return branch
}
