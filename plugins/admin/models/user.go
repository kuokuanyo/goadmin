package models

import (
	"database/sql"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// UserModel is user model structure.
type UserModel struct {
	//plugins/admin/models/base.go中
	Base `json:"-"`

	Id            int64             `json:"id"`
	Name          string            `json:"name"`
	UserName      string            `json:"user_name"`
	Password      string            `json:"password"`
	Avatar        string            `json:"avatar"`
	RememberToken string            `json:"remember_token"`
	//plugins/admin/models/permission.go中
	Permissions   []PermissionModel `json:"permissions"`
	MenuIds       []int64           `json:"menu_ids"`
	//plugins/admin/models/role.go中
	Roles         []RoleModel       `json:"role"`
	Level         string            `json:"level"`
	LevelName     string            `json:"level_name"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// User return a default user model.
// 設置UserModel.Base.TableName(struct)並回傳設置UserModel(struct)
func User() UserModel {
	//config.GetAuthUserTable()在modules/config/config.go中
	//config.GetAuthUserTable()取得tablename(globalCfg.AuthUserTable)
	return UserModel{Base: Base{TableName: config.GetAuthUserTable()}}
}

// UserWithId return a default user model of given id.
func UserWithId(id string) UserModel {
	idInt, _ := strconv.Atoi(id)
	return UserModel{Base: Base{TableName: config.GetAuthUserTable()}, Id: int64(idInt)}
}

// 將參數con(db.Connection)設置至UserModel.conn(UserModel.Base.Conn)
// db.Connection(interface)
func (t UserModel) SetConn(con db.Connection) UserModel {
	//UserModel.Base.Conn
	t.Conn = con
	return t
}

func (t UserModel) WithTx(tx *sql.Tx) UserModel {
	t.Tx = tx
	return t
}

// Find return a default user model of given id.
// 取的user model藉由id
func (t UserModel) Find(id interface{}) UserModel {
	// Table 取得預設sql struct 並且設置tablename
	// Table /plugins/admin/models/base.go中
	// Find 在modules/db/statement.go中
	item, _ := t.Table(t.TableName).Find(id)
	//將item(map)設置至user model
	return t.MapToModel(item)
}

// FindByUserName return a default user model of given name.
// 透過參數username尋找符合的資料並設置至UserModel
func (t UserModel) FindByUserName(username interface{}) UserModel {
	// Table藉由給定的table回傳sql(struct)
	// sql 語法 where = ...，回傳 SQl struct
	// First回傳第一筆符合的資料
	item, _ := t.Table(t.TableName).Where("username", "=", username).First()
	// 將item資訊設置至UserModel後回傳
	return t.MapToModel(item)
}

// IsEmpty check the user model is empty or not.
// 判斷是否為空
func (t UserModel) IsEmpty() bool {
	return t.Id == int64(0)
}

// HasMenu check the user has visitable menu or not.
// 檢查用戶是否有可訪問的menu
func (t UserModel) HasMenu() bool {
	return len(t.MenuIds) != 0 || t.IsSuperAdmin()
}

// IsSuperAdmin check the user model is super admin or not.
// 判斷是否為超級管理員
func (t UserModel) IsSuperAdmin() bool {
	for _, per := range t.Permissions {
		if len(per.HttpPath) > 0 && per.HttpPath[0] == "*" && per.HttpMethod[0] == "" {
			return true
		}
	}
	return false
}

func (t UserModel) GetCheckPermissionByUrlMethod(path, method string) string {
	if !t.CheckPermissionByUrlMethod(path, method, url.Values{}) {
		return ""
	}
	return path
}

func (t UserModel) IsVisitor() bool {
	return !t.CheckPermissionByUrlMethod(config.Url("/info/normal_manager"), "GET", url.Values{})
}

func (t UserModel) HideUserCenterEntrance() bool {
	return t.IsVisitor() && config.GetHideVisitorUserCenterEntrance()
}

// 檢查權限(藉由url、method)
func (t UserModel) CheckPermissionByUrlMethod(path, method string, formParams url.Values) bool {

	// 檢查是否為超級管理員
	if t.IsSuperAdmin() {
		return true
	}

	// 登出檢查
	logoutCheck, _ := regexp.Compile(config.Url("/logout") + "(.*?)")

	if logoutCheck.MatchString(path) {
		return true
	}

	if path == "" {
		return false
	}

	if path != "/" && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	path = strings.Replace(path, constant.EditPKKey, "id", -1)
	path = strings.Replace(path, constant.DetailPKKey, "id", -1)

	// 取得路徑及參數
	path, params := getParam(path)
	for key, value := range formParams {
		if len(value) > 0 {
			params.Add(key, value[0])
		}
	}

	for _, v := range t.Permissions {

		if v.HttpMethod[0] == "" || inMethodArr(v.HttpMethod, method) {

			if v.HttpPath[0] == "*" {
				return true
			}

			for i := 0; i < len(v.HttpPath); i++ {

				matchPath := config.Url(strings.TrimSpace(v.HttpPath[i]))
				matchPath, matchParam := getParam(matchPath)

				if matchPath == path {
					if checkParam(params, matchParam) {
						return true
					}
				}

				reg, err := regexp.Compile(matchPath)

				if err != nil {
					logger.Error("CheckPermissions error: ", err)
					continue
				}

				if reg.FindString(path) == path {
					if checkParam(params, matchParam) {
						return true
					}
				}
			}
		}
	}

	return false
}

// 取得參數
func getParam(u string) (string, url.Values) {
	m := make(url.Values)
	urr := strings.Split(u, "?")
	if len(urr) > 1 {
		m, _ = url.ParseQuery(urr[1])
	}
	return urr[0], m
}

func checkParam(src, comp url.Values) bool {
	if len(comp) == 0 {
		return true
	}
	if len(src) == 0 {
		return false
	}
	for key, value := range comp {
		v, find := src[key]
		if !find {
			return false
		}
		if len(value) == 0 {
			continue
		}
		if len(v) == 0 {
			return false
		}
		for i := 0; i < len(v); i++ {
			if v[i] == value[i] {
				continue
			} else {
				return false
			}
		}
	}
	return true
}

func inMethodArr(arr []string, str string) bool {
	for i := 0; i < len(arr); i++ {
		if strings.EqualFold(arr[i], str) {
			return true
		}
	}
	return false
}

// UpdateAvatar update the avatar of user.
func (t UserModel) ReleaseConn() UserModel {
	t.Conn = nil
	return t
}

// UpdateAvatar update the avatar of user.
func (t UserModel) UpdateAvatar(avatar string) {
	t.Avatar = avatar
}

// WithRoles query the role info of the user.
// 查詢role藉由user
func (t UserModel) WithRoles() UserModel {
	// 查詢goadmin_role_users資料表
	// LeftJoin、Where、Select、All都在modules/db/statement.go中
	// 執行sql命令(join、wherew、select等篩選命令)
	// 結果會有(goadmin_roles.id", "goadmin_roles.name", "goadmin_roles.slug","goadmin_roles.created_at", "goadmin_roles.updated_at")五個欄位
	// roleModel取得role藉由user_id
	// 可能會有多個role(ex:user_id = 1可能設定多個role)
	roleModel, _ := t.Table("goadmin_role_users").
		LeftJoin("goadmin_roles", "goadmin_roles.id", "=", "goadmin_role_users.role_id").
		Where("user_id", "=", t.Id).
		Select("goadmin_roles.id", "goadmin_roles.name", "goadmin_roles.slug",
			"goadmin_roles.created_at", "goadmin_roles.updated_at").
		All()

	for _, role := range roleModel {
		//Role、MapToModel在plugins/admin/models/role.go中
		//Role取得初始化role model
		//MapToModel將role設置至role model中
		t.Roles = append(t.Roles, Role().MapToModel(role))
	}

	if len(t.Roles) > 0 {
		//設置slug、name至user model
		t.Level = t.Roles[0].Slug
		t.LevelName = t.Roles[0].Name
	}

	return t
}

// 取得該用戶的role_id
func (t UserModel) GetAllRoleId() []interface{} {

	var ids = make([]interface{}, len(t.Roles))

	for key, role := range t.Roles {
		ids[key] = role.Id
	}

	return ids
}

// WithPermissions query the permission info of the user.
// 查詢user的permission
func (t UserModel) WithPermissions() UserModel {

	var permissions = make([]map[string]interface{}, 0)

	//可能會有多個role id(可以設定多個role)
	roleIds := t.GetAllRoleId()

	//----------------------------------------------------------------------------------------------
	// permission會依照user_id以及role_id取得不同的權限，因此需要做下列兩次判斷
	//----------------------------------------------------------------------------------------------
	// 查詢goadmin_role_permissions資料表
	// LeftJoin、WhereIn、Select、All都在modules/db/statement.go中
	// 執行sql命令(join、where、select等篩選指令)
	// 結果會有("goadmin_permissions.http_method", "goadmin_permissions.http_path","goadmin_permissions.id", "goadmin_permissions.name", "goadmin_permissions.slug","goadmin_permissions.created_at", "goadmin_permissions.updated_at")七個欄位
	// permissions藉由role_id取得permission
	// 可能會有permission(ex:role_id = 1有兩個permission)
	// 假設該user有設定role才會執行下面指令取得該role的permission
	if len(roleIds) > 0 {
		//查詢role_id的permission
		permissions, _ = t.Table("goadmin_role_permissions").
			LeftJoin("goadmin_permissions", "goadmin_permissions.id", "=", "goadmin_role_permissions.permission_id").
			WhereIn("role_id", roleIds).
			Select("goadmin_permissions.http_method", "goadmin_permissions.http_path",
				"goadmin_permissions.id", "goadmin_permissions.name", "goadmin_permissions.slug",
				"goadmin_permissions.created_at", "goadmin_permissions.updated_at").
			All()
	}

	// 跟上面role的方式一樣(藉由user_id取得permission)
	// 可能有多個permission
	userPermissions, _ := t.Table("goadmin_user_permissions").
		LeftJoin("goadmin_permissions", "goadmin_permissions.id", "=", "goadmin_user_permissions.permission_id").
		Where("user_id", "=", t.Id).
		Select("goadmin_permissions.http_method", "goadmin_permissions.http_path",
			"goadmin_permissions.id", "goadmin_permissions.name", "goadmin_permissions.slug",
			"goadmin_permissions.created_at", "goadmin_permissions.updated_at").
		All()

	permissions = append(permissions, userPermissions...)

	for i := 0; i < len(permissions); i++ {
		exist := false
		//如果裡面已經有相同的權限加入，就停止迴圈
		for j := 0; j < len(t.Permissions); j++ {
			if t.Permissions[j].Id == permissions[i]["id"] {
				exist = true
				break
			}
		}

		//Permission、MapToModel在plugins/admin/models/permission.go中
		//Permission 為初始化Permission model
		//MapToModel 將role設置至Permission model
		if exist {
			continue
		}
		t.Permissions = append(t.Permissions, Permission().MapToModel(permissions[i]))
	}

	return t
}

// WithMenus query the menu info of the user.
// 查詢menu藉由user
func (t UserModel) WithMenus() UserModel {

	var menuIdsModel []map[string]interface{}

	// 判斷是否為超級管理員
	if t.IsSuperAdmin() {
		// 查詢goadmin_role_menu資料表
		// Table在plugins/admin/modules/base.go 中，初始化sql struct並設置tablename
		// LeftJoin、Select、All都在modules/db/statement.go中
		// LeftJoin、Select、All等sql篩選命令
		// 總共取得menu_id, parent_id兩個欄位
		menuIdsModel, _ = t.Table("goadmin_role_menu").
			LeftJoin("goadmin_menu", "goadmin_menu.id", "=", "goadmin_role_menu.menu_id").
			Select("menu_id", "parent_id").
			All()
	} else {
		// 取得該user的role_id
		// LeftJoin、WhereIn、Select、All都在modules/db/statement.go中
		// 總共取得menu_id, parent_id兩個欄位
		// 取得menuIdsModel藉由role_id
		rolesId := t.GetAllRoleId()
		if len(rolesId) > 0 {
			menuIdsModel, _ = t.Table("goadmin_role_menu").
				LeftJoin("goadmin_menu", "goadmin_menu.id", "=", "goadmin_role_menu.menu_id").
				WhereIn("goadmin_role_menu.role_id", rolesId).
				Select("menu_id", "parent_id").
				All()
		}
	}

	var menuIds []int64

	// 將menu_id加入menuIds中
	for _, mid := range menuIdsModel {
		if parentId, ok := mid["parent_id"].(int64); ok && parentId != 0 {
			for _, mid2 := range menuIdsModel {
				if mid2["menu_id"].(int64) == mid["parent_id"].(int64) {
					menuIds = append(menuIds, mid["menu_id"].(int64))
					break
				}
			}
		} else {
			menuIds = append(menuIds, mid["menu_id"].(int64))
		}
	}

	// 設置UserModel.MenuIds
	t.MenuIds = menuIds
	return t
}

// New create a user model.
func (t UserModel) New(username, password, name, avatar string) (UserModel, error) {

	id, err := t.WithTx(t.Tx).Table(t.TableName).Insert(dialect.H{
		"username": username,
		"password": password,
		"name":     name,
		"avatar":   avatar,
	})

	t.Id = id
	t.UserName = username
	t.Password = password
	t.Avatar = avatar
	t.Name = name

	return t, err
}

// Update update the user model.
func (t UserModel) Update(username, password, name, avatar string) (int64, error) {

	fieldValues := dialect.H{
		"username":   username,
		"name":       name,
		"avatar":     avatar,
		"updated_at": time.Now().Format("2006-01-02 15:04:05"),
	}

	if password != "" {
		fieldValues["password"] = password
	}

	return t.WithTx(t.Tx).Table(t.TableName).
		Where("id", "=", t.Id).
		Update(fieldValues)
}

// UpdatePwd update the password of the user model.
// 將參數password設置至UserModel.UserModel並且更新dialect.H{"password": password,}
func (t UserModel) UpdatePwd(password string) UserModel {

	_, _ = t.Table(t.TableName).
		Where("id", "=", t.Id).
		Update(dialect.H{
			"password": password,
		})

	t.Password = password
	return t
}

// CheckRole check the role of the user model.
func (t UserModel) CheckRoleId(roleId string) bool {
	checkRole, _ := t.Table("goadmin_role_users").
		Where("role_id", "=", roleId).
		Where("user_id", "=", t.Id).
		First()
	return checkRole != nil
}

// DeleteRoles delete all the roles of the user model.
func (t UserModel) DeleteRoles() error {
	return t.Table("goadmin_role_users").
		Where("user_id", "=", t.Id).
		Delete()
}

// AddRole add a role of the user model.
func (t UserModel) AddRole(roleId string) (int64, error) {
	if roleId != "" {
		if !t.CheckRoleId(roleId) {
			return t.WithTx(t.Tx).Table("goadmin_role_users").
				Insert(dialect.H{
					"role_id": roleId,
					"user_id": t.Id,
				})
		}
	}
	return 0, nil
}

// CheckRole check the role of the user.
func (t UserModel) CheckRole(slug string) bool {
	for _, role := range t.Roles {
		if role.Slug == slug {
			return true
		}
	}

	return false
}

// CheckPermission check the permission of the user.
func (t UserModel) CheckPermissionById(permissionId string) bool {
	checkPermission, _ := t.Table("goadmin_user_permissions").
		Where("permission_id", "=", permissionId).
		Where("user_id", "=", t.Id).
		First()
	return checkPermission != nil
}

// CheckPermission check the permission of the user.
func (t UserModel) CheckPermission(permission string) bool {
	for _, per := range t.Permissions {
		if per.Slug == permission {
			return true
		}
	}

	return false
}

// DeletePermissions delete all the permissions of the user model.
func (t UserModel) DeletePermissions() error {
	return t.WithTx(t.Tx).Table("goadmin_user_permissions").
		Where("user_id", "=", t.Id).
		Delete()
}

// AddPermission add a permission of the user model.
func (t UserModel) AddPermission(permissionId string) (int64, error) {
	if permissionId != "" {
		if !t.CheckPermissionById(permissionId) {
			return t.WithTx(t.Tx).Table("goadmin_user_permissions").
				Insert(dialect.H{
					"permission_id": permissionId,
					"user_id":       t.Id,
				})
		}
	}
	return 0, nil
}

// MapToModel get the user model from given map.
// 設置user model從map中
func (t UserModel) MapToModel(m map[string]interface{}) UserModel {
	t.Id, _ = m["id"].(int64)
	t.Name, _ = m["name"].(string)
	t.UserName, _ = m["username"].(string)
	t.Password, _ = m["password"].(string)
	t.Avatar, _ = m["avatar"].(string)
	t.RememberToken, _ = m["remember_token"].(string)
	t.CreatedAt, _ = m["created_at"].(string)
	t.UpdatedAt, _ = m["updated_at"].(string)
	return t
}
