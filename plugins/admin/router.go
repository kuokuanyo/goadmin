package admin

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"
	"github.com/GoAdminGroup/go-admin/template"
)

// initRouter initialize the router and return the context.
// 初始化router
func (admin *Admin) initRouter() *Admin {
	// 回傳新的App(struct)，空的
	app := context.NewApp()

	// 將參數config.Prefix()、admin.globalErrorHandler(錯誤處理程序)新增至RouterGroup(struct)
	// Prefix回傳globalCfg(Config struct).prefix
	// globalErrorHandler判斷是否將站點關閉後執行迴圈Context.handlers[ctx.index](ctx)
	// 最後印出訪問訊息在終端機上並記錄所有操作行為至資料表(goadmin_operation_log)中
	route := app.Group(config.Prefix(), admin.globalErrorHandler)

	// auth
	// GetLoginUrl回傳globalCfg.LoginUrl
	// ShowLogin在plugins\admin\controller\auth.go
	// ShowLogin判斷map[string]Component(interface)是否有參數login(key)的值，接著執行template將data寫入buf並輸出HTML
	route.GET(config.GetLoginUrl(), admin.handler.ShowLogin)

	// 對輸入的username、password身分驗證後取得user的role、permission及可用menu，最後更新資料表(goadmin_users)的密碼值(加密)
	route.POST("/signin", admin.handler.Auth)

	// auto install
	// plugins\admin\controller\install.go
	// 建立buffer(bytes.Buffer)並輸出HTML
	route.GET("/install", admin.handler.ShowInstall)
	// 檢查資料庫連線參數是否正確(mysql)
	// 參數設置範例h=127.0.0.1,po=3306,u=root,pa=asdf4440,db=godmin(在multipart/form-data配置)
	route.POST("/install/database/check", admin.handler.CheckDatabase)

	// 處理前端的檔案
	// Get判斷templateMap(map[string]Template)的key鍵是否參數theme，有則回傳Template(interface)
	// GetTheme回傳globalCfg.Theme
	// GetAssetList為Template(interface)的方法
	checkRepeatedPath := make([]string, 0)
	for _, themeName := range template.Themes() {
		for _, path := range template.Get(themeName).GetAssetList() {
			if !utils.InArray(checkRepeatedPath, path) {
				checkRepeatedPath = append(checkRepeatedPath, path)
				route.GET("/assets"+path, admin.handler.Assets)
			}
		}
	}

	// GetComponentAsset檢查compMap(map[string]Component)的物件一一加入陣列([]string)中
	for _, path := range template.GetComponentAsset() {
		route.GET("/assets"+path, admin.handler.Assets)
	}

	// 將參數"/"、auth.middleware(admin.Conn)新增至RouterGroup(struct)
	// Middleware建立Invoker(Struct)並透過參數ctx取得UserModel，並且取得該user的role、權限與可用menu，最後檢查用戶權限
	// authRoute需要驗證user的role、權限與可用menu，最後檢查用戶權限
	authRoute := route.Group("/", auth.Middleware(admin.Conn))

	// auth
	// 登出並清除cookie後回到登入頁面
	authRoute.GET("/logout", admin.handler.Logout)

	// menus
	// 需要有參數id = ?
	// MenuDelete查詢url中參數id的值後將id設置至MenuDeleteParam(struct)，接著將值設置至Context.UserValue[delete_menu_param]中，最後執行迴圈Context.handlers[ctx.index](ctx)
	// DeleteMenu刪除條件MenuModel.id的資料，除了刪除goadmin_menu之外還要刪除goadmin_role_menu資料
	// 如果MenuModel.id是其他菜單的父級，也必須刪除
	authRoute.POST("/menu/delete", admin.guardian.MenuDelete, admin.handler.DeleteMenu).Name("menu_delete")

	// MenuNew在plugins\admin\modules\guard\menu_new.go
	// MenuNew藉由參數取得multipart/form-data中設置的值，接著驗證token並將multipart/form-data的key、value值設置至Context.UserValue[new_menu_param]，最後執行迴圈Context.handlers[ctx.index](ctx)
	// NewMenu將Context.UserValue(map[string]interface{})[new_menu_param]的值轉換成MenuNewParam(struct)類別，接著將MenuNewParam(struct)值新增至資料表(MenuModel.Base.TableName(goadmin_menu))中
	// 最後如果multipart/form-data有設定roles[]值，檢查條件後將參數roleId(role_id)與MenuModel.Id(menu_id)加入goadmin_role_menu資料表
	authRoute.POST("/menu/new", admin.guardian.MenuNew, admin.handler.NewMenu).Name("menu_new")

	// MenuEdit在plugins\admin\modules\guard\menu_edit.go中
	// MenuEdit藉由參數取得multipart/form-data中設置的值，接著驗證token並將multipart/form-data的key、value值設置至Context.UserValue[edit_menu_param]，最後執行迴圈Context.handlers[ctx.index](ctx)
	// EditMenu將Context.UserValue(map[string]interface{})[edit_menu_param]的值轉換成MenuEditParam(struct)類別
	// 先將goadmin_role_menu資料表中menu_id = MenuModel.Id的資料刪除，接著如果有在multipart/form-data有設定roles[]值，檢查條件後將參數roleId(role_id)與MenuModel.Id(menu_id)加入goadmin_role_menu資料表
	// 接著將goadmin_menu資料表條件為id = MenuModel.Id的資料透過參數(由multipart/form-data設置)更新
	authRoute.POST("/menu/edit", admin.guardian.MenuEdit, admin.handler.EditMenu).Name("menu_edit")

	// 取得multipart/form-data中的_order參數後更改menu順序
	// 參數設置範例(在multipart/form-data配置):_order: [{"id":7},{"id":1,"children":[{"id":2},{"id":3},{"id":4},{"id":5},{"id":6}]},{"id":15},{"id":17},{"id":16}])
	authRoute.POST("/menu/order", admin.handler.MenuOrder).Name("menu_order")

	//分別處理上下半部表單的HTML語法，最後結合並輸出HTML
	authRoute.GET("/menu", admin.handler.ShowMenu).Name("menu")

	// 先檢查設置的參數(id = ?)是否符合條件，接著透過id取得goadmin_menu資料表中的資料，然後設置值至FormInfo(struct)中
	// 最後以FormInfo(struct)匯出編輯介面的HTML語法
	authRoute.GET("/menu/edit/show", admin.handler.ShowEditMenu).Name("menu_edit_show")
	authRoute.GET("/menu/new", admin.handler.ShowNewMenu).Name("menu_new_show")

	// Group將參數"/"、auth.middleware(admin.Conn)、admin.guardian.CheckPrefix新增至RouterGroup(struct)
	// CheckPrefix在plugins\admin\modules\guard\guard.go
	// CheckPrefix查詢url裡的參數(__prefix)，如果Guard.tableList存在該prefix(key)則執行迴圈
	// authPrefixRoute需要驗證user的role、權限與可用menu，最後檢查用戶權限，以及查詢url裡的參數(__prefix)
	authPrefixRoute := route.Group("/", auth.Middleware(admin.Conn), admin.guardian.CheckPrefix)

	// add delete modify query
	authPrefixRoute.GET("/info/:__prefix/detail", admin.handler.ShowDetail).Name("detail")
	authPrefixRoute.GET("/info/:__prefix/edit", admin.guardian.ShowForm, admin.handler.ShowForm).Name("show_edit")
	authPrefixRoute.GET("/info/:__prefix/new", admin.guardian.ShowNewForm, admin.handler.ShowNewForm).Name("show_new")

	// EditForm(編輯表單)編輯用戶、角色、權限等表單資訊，首先取得multipart/form-data設定的參數值並驗證token是否正確
	// 接著取得頁面size、資料排列方式、選擇欄位...等資訊後設置至Parameters(struct)，最後設定Context.UserValue並執行編輯表單的動作
	authPrefixRoute.POST("/edit/:__prefix", admin.guardian.EditForm, admin.handler.EditForm).Name("edit")

	authPrefixRoute.POST("/new/:__prefix", admin.guardian.NewForm, admin.handler.NewForm).Name("new")
	authPrefixRoute.POST("/delete/:__prefix", admin.guardian.Delete, admin.handler.Delete).Name("delete")
	authPrefixRoute.POST("/export/:__prefix", admin.guardian.Export, admin.handler.Export).Name("export")
	authPrefixRoute.GET("/info/:__prefix", admin.handler.ShowInfo).Name("info")

	authPrefixRoute.POST("/update/:__prefix", admin.guardian.Update, admin.handler.Update).Name("update")

	authRoute.GET("/application/info", admin.handler.SystemInfo)

	route.ANY("/operation/:__goadmin_op_id", auth.Middleware(admin.Conn), admin.handler.Operation)

	if config.GetOpenAdminApi() {

		// crud json apis
		apiRoute := route.Group("/api", auth.Middleware(admin.Conn), admin.guardian.CheckPrefix)
		apiRoute.GET("/list/:__prefix", admin.handler.ApiList).Name("api_info")
		apiRoute.GET("/detail/:__prefix", admin.handler.ApiDetail).Name("api_detail")
		apiRoute.POST("/delete/:__prefix", admin.guardian.Delete, admin.handler.Delete).Name("api_delete")
		apiRoute.POST("/edit/:__prefix", admin.guardian.EditForm, admin.handler.ApiUpdate).Name("api_edit")
		apiRoute.GET("/edit/form/:__prefix", admin.guardian.ShowForm, admin.handler.ApiUpdateForm).Name("api_show_edit")
		apiRoute.POST("/create/:__prefix", admin.guardian.NewForm, admin.handler.ApiCreate).Name("api_new")
		apiRoute.GET("/create/form/:__prefix", admin.guardian.ShowNewForm, admin.handler.ApiCreateForm).Name("api_show_new")
		apiRoute.POST("/export/:__prefix", admin.guardian.Export, admin.handler.Export).Name("api_export")
		apiRoute.POST("/update/:__prefix", admin.guardian.Update, admin.handler.Update).Name("api_update")
	}

	admin.App = app
	return admin
}

// globalErrorHandler(錯誤處理程序)
// 判斷是否將站點關閉後執行迴圈Context.handlers[ctx.index](ctx)
// 最後印出訪問訊息在終端機上並記錄所有操作行為至資料表(goadmin_operation_log)中
func (admin *Admin) globalErrorHandler(ctx *context.Context) {
	// 在plugins\admin\controller\handler.go
	// 印出訪問訊息在終端機上並記錄所有操作行為至資料表(goadmin_operation_log)中
	defer admin.handler.GlobalDeferHandler(ctx)

	// OffLineHandler在plugins\admin\modules\response\response.go
	// OffLineHandler(離線處理程序)是func(ctx *context.Context)
	// OffLineHandler判斷站點是否要關閉，如要關閉，判斷method是否為get以及header裡包含accept:html後輸出HTML
	response.OffLineHandler(ctx)

	// 執行迴圈Context.handlers[ctx.index](ctx)
	ctx.Next()
}
