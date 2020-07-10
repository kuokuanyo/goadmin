package models

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
)

// MenuModel is menu model structure.
type MenuModel struct {
	Base

	Id        int64
	Title     string
	ParentId  int64
	Icon      string
	Uri       string
	Header    string
	CreatedAt string
	UpdatedAt string
}

// Menu return a default menu model.
// 將MenuModel(struct).Base.TableName設置goadmin_menu後回傳
func Menu() MenuModel {
	return MenuModel{Base: Base{TableName: "goadmin_menu"}}
}

// MenuWithId return a default menu model of given id.
// 透過參數將id與tablename(goadmin_menu)設置至MenuModel(struct)後回傳
func MenuWithId(id string) MenuModel {
	idInt, _ := strconv.Atoi(id)
	return MenuModel{Base: Base{TableName: "goadmin_menu"}, Id: int64(idInt)}
}

// 將參數con設置至MenuModel.Base.Conn
func (t MenuModel) SetConn(con db.Connection) MenuModel {
	t.Conn = con
	return t
}

// Find return a default menu model of given id.
func (t MenuModel) Find(id interface{}) MenuModel {
	item, _ := t.Table(t.TableName).Find(id)
	return t.MapToModel(item)
}

// New create a new menu model.
// 將參數值新增至資料表(MenuModel.Base.TableName(goadmin_menu))中，最後將參數值都設置在MenuModel中
func (t MenuModel) New(title, icon, uri, header string, parentId, order int64) (MenuModel, error) {

	id, err := t.Table(t.TableName).Insert(dialect.H{
		"title":     title,
		"parent_id": parentId,
		"icon":      icon,
		"uri":       uri,
		"order":     order,
		"header":    header,
	})

	t.Id = id
	t.Title = title
	t.ParentId = parentId
	t.Icon = icon
	t.Uri = uri
	t.Header = header

	return t, err
}

// Delete delete the menu model.
// 刪除條件MenuModel.id的資料，除了刪除goadmin_menu之外還要刪除goadmin_role_menu資料
// 如果MenuModel.id是其他菜單的父級，也必須刪除
func (t MenuModel) Delete() {
	_ = t.Table(t.TableName).Where("id", "=", t.Id).Delete()
	_ = t.Table("goadmin_role_menu").Where("menu_id", "=", t.Id).Delete()
	items, _ := t.Table(t.TableName).Where("parent_id", "=", t.Id).All()

	// 如果MenuModel.id是其他菜單的父級，也必須刪除
	if len(items) > 0 {
		ids := make([]interface{}, len(items))
		for i := 0; i < len(ids); i++ {
			ids[i] = items[i]["id"]
		}
		_ = t.Table("goadmin_role_menu").WhereIn("menu_id", ids).Delete()
	}

	_ = t.Table(t.TableName).Where("parent_id", "=", t.Id).Delete()
}

// Update update the menu model.
// 將goadmin_menu資料表條件為id = MenuModel.Id的資料透過參數(由multipart/form-data設置)更新
func (t MenuModel) Update(title, icon, uri, header string, parentId int64) (int64, error) {
	return t.Table(t.TableName).
		Where("id", "=", t.Id).
		Update(dialect.H{
			"title":      title,
			"parent_id":  parentId,
			"icon":       icon,
			"uri":        uri,
			"header":     header,
			"updated_at": time.Now().Format("2006-01-02 15:04:05"),
		})
}

type OrderItems []OrderItem

type OrderItem struct {
	ID       uint       `json:"id"`
	Children OrderItems `json:"children"`
}

// ResetOrder update the order of menu models.
// 更改menu的順序
func (t MenuModel) ResetOrder(data []byte) {

	var items OrderItems
	// 將參數data解碼至items([]OrderItem(struct))
	_ = json.Unmarshal(data, &items)
	count := 1
	for _, v := range items {
		if len(v.Children) > 0 {
			_, _ = t.Table(t.TableName).
				Where("id", "=", v.ID).Update(dialect.H{
				"order":     count,
				"parent_id": 0,
			})

			for _, v2 := range v.Children {
				if len(v2.Children) > 0 {

					_, _ = t.Table(t.TableName).
						Where("id", "=", v2.ID).Update(dialect.H{
						"order":     count,
						"parent_id": v.ID,
					})

					for _, v3 := range v2.Children {
						_, _ = t.Table(t.TableName).
							Where("id", "=", v3.ID).Update(dialect.H{
							"order":     count,
							"parent_id": v2.ID,
						})
						count++
					}
				} else {
					_, _ = t.Table(t.TableName).
						Where("id", "=", v2.ID).Update(dialect.H{
						"order":     count,
						"parent_id": v.ID,
					})
					count++
				}
			}
		} else {
			_, _ = t.Table(t.TableName).
				Where("id", "=", v.ID).Update(dialect.H{
				"order":     count,
				"parent_id": 0,
			})
			count++
		}
	}
}

// CheckRole check the role if has permission to get the menu.
// 檢查goadmin_role_menu資料表裡是否有符合role_id = 參數roleId與menu_id = MenuModel.Id條件
func (t MenuModel) CheckRole(roleId string) bool {
	checkRole, _ := t.Table("goadmin_role_menu").
		Where("role_id", "=", roleId).
		Where("menu_id", "=", t.Id).
		First()
	return checkRole != nil
}

// AddRole add a role to the menu.
// 先檢查goadmin_role_menu條件，接著將參數roleId(role_id)與MenuModel.Id(menu_id)加入goadmin_role_menu資料表
func (t MenuModel) AddRole(roleId string) (int64, error) {
	if roleId != "" {
		// 檢查goadmin_role_menu資料表裡是否有符合role_id = 參數roleId與menu_id = MenuModel.Id條件
		if !t.CheckRole(roleId) {
			// 將參數roleId(role_id)與MenuModel.Id(menu_id)加入goadmin_role_menu資料表
			return t.Table("goadmin_role_menu").
				Insert(dialect.H{
					"role_id": roleId,
					"menu_id": t.Id,
				})
		}
	}
	return 0, nil
}

// DeleteRoles delete roles with menu.
// 刪除goadmin_role_menu資料表中menu_id = MenuModel.Id條件的資料
func (t MenuModel) DeleteRoles() error {
	return t.Table("goadmin_role_menu").
		Where("menu_id", "=", t.Id).
		Delete()
}

// MapToModel get the menu model from given map.
func (t MenuModel) MapToModel(m map[string]interface{}) MenuModel {
	t.Id = m["id"].(int64)
	t.Title, _ = m["title"].(string)
	t.ParentId = m["parent_id"].(int64)
	t.Icon, _ = m["icon"].(string)
	t.Uri, _ = m["uri"].(string)
	t.Header, _ = m["header"].(string)
	t.CreatedAt, _ = m["created_at"].(string)
	t.UpdatedAt, _ = m["updated_at"].(string)
	return t
}
