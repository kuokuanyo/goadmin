// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package engine

import (
	"bytes"
	"encoding/json"
	errors2 "errors"
	"fmt"
	template2 "html/template"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/template/icon"
	"github.com/GoAdminGroup/go-admin/template/types/action"

	"github.com/GoAdminGroup/go-admin/adapter"
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/menu"
	"github.com/GoAdminGroup/go-admin/modules/service"
	"github.com/GoAdminGroup/go-admin/modules/system"
	"github.com/GoAdminGroup/go-admin/modules/ui"
	"github.com/GoAdminGroup/go-admin/plugins"
	"github.com/GoAdminGroup/go-admin/plugins/admin"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/response"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
)

// Engine is the core component of goAdmin. It has two attributes.
// PluginList is an array of plugin. Adapter is the adapter of
// web framework context and goAdmin context. The relationship of adapter and
// plugin is that the adapter use the plugin which contains routers and
// controller methods to inject into the framework entity and make it work.
// 核心組件，有PluginList及Adapter兩個屬性
type Engine struct {
	// GetRequest回傳插件中的所有路徑
	// InitPlugin初始化插件，類似於初始化資料庫並設置及配置路徑
	PluginList []plugins.Plugin
	Adapter    adapter.WebFrameWork
	//Services在modules\service\service.go
	Services   service.List //Services類別為map[string]Service，Service為interface(Name方法)
	NavButtons *types.Buttons
	config     *config.Config //struct
}

// Default return the default engine instance.
// 回傳預設的Engine(struct)
func Default() *Engine {
	engine = &Engine{
		//調適器
		//空的adapter.WebFrameWork(interface)
		Adapter:    defaultAdapter,
		//map[token_csrf_helper:0xc0003c5e20]
		Services:   service.GetServices(),
		// 預設的Buttons(interface)
		NavButtons: new(types.Buttons),
	}
	return engine
}

// Use enable the adapter.
// 尋找符合的plugin，接著設置context.Context(struct)與設置url與寫入header，取得新的request與middleware
func (eng *Engine) Use(router interface{}) error {
	if eng.Adapter == nil {
		panic("adapter is nil, import the default adapter or use AddAdapter method add the adapter")
	}

	// 透過參數(name)找到符合的plugin
	_, exist := eng.FindPluginByName("admin")

	// 不存在符合的plugin
	if !exist {
		// admin.NewAdmin()在plugins\admin\admin.go中
		eng.PluginList = append(eng.PluginList, admin.NewAdmin())
	}

	// init site setting
	site := models.Site().SetConn(eng.DefaultConnection())
	site.Init(eng.config.ToMap())
	_ = eng.config.Update(site.AllToMap())
	// 在modules\service\service.go
	// 藉由參數新增List(map[string]Service)，新增config
	// config.SrvWithConfig設置Service
	eng.Services.Add("config", config.SrvWithConfig(eng.config))

	errors.Init()

	// 隱藏設置入口
	if !eng.config.HideConfigCenterEntrance {
		*eng.NavButtons = (*eng.NavButtons).AddNavButton(icon.Gear, types.NavBtnSiteName,
			action.JumpInNewTab(config.Url("/info/site/edit"),
				language.GetWithScope("site setting", "config")))
	}

	// 隱藏App Info入口
	if !eng.config.HideAppInfoEntrance {
		*eng.NavButtons = (*eng.NavButtons).AddNavButton(icon.Info, types.NavBtnInfoName,
			action.JumpInNewTab(config.Url("/application/info"),
				language.GetWithScope("system info", "system")))
	}

	if !eng.config.HideToolEntrance {
		*eng.NavButtons = (*eng.NavButtons).AddNavButton(icon.Wrench, types.NavBtnToolName,
			action.JumpInNewTab(config.Url("/info/generate/new"),
				language.GetWithScope("tool", "tool")))
	}

	navButtons = eng.NavButtons

	// ui.ServiceKey = ui
	// 藉由參數新增List(map[string]Service)，新增ui
	// ui.NewService設置Service
	eng.Services.Add(ui.ServiceKey, ui.NewService(eng.NavButtons))

	// GetConnection在modules\db\connection.go
	// 取得匹配的eng.Services然後轉換成Connection(interface)類別
	defaultConnection := db.GetConnection(eng.Services)
	// SetConnection為WebFrameWork(interface)的方法
	//設定連線
	defaultAdapter.SetConnection(defaultConnection)
	eng.Adapter.SetConnection(defaultConnection)

	// Initialize plugins
	for i := range eng.PluginList {
		// 初始化每個plugin
		eng.PluginList[i].InitPlugin(eng.Services)
	}

	// Use在adapter/gin/gin.go中(WebFrameWork方法)
	// 增加處理程序(Handler)
	// 設置context.Context(struct)與設置url與寫入header，取得新的request與middleware
	return eng.Adapter.Use(router, eng.PluginList)
}

// AddPlugins add the plugins and initialize them.
// 增加plugins以及初始化
func (eng *Engine) AddPlugins(plugs ...plugins.Plugin) *Engine {

	if len(plugs) == 0 {
		panic("wrong plugins")
	}

	eng.PluginList = append(eng.PluginList, plugs...)

	return eng
}

// FindPluginByName find the register plugin by given name.
// 透過參數(name)找到符合的plugin
func (eng *Engine) FindPluginByName(name string) (plugins.Plugin, bool) {
	for _, plug := range eng.PluginList {
		if plug.Name() == name {
			return plug, true
		}
	}
	return nil, false
}

// AddAuthService customize the auth logic with given callback function.
// 增加身分驗證，回傳Engine
func (eng *Engine) AddAuthService(processor auth.Processor) *Engine {
	// Add藉由參數新增List(map[string]Service)
	// auth.NewService在modules\auth\auth.go，將參數設置並回傳Service(struct)
	eng.Services.Add("auth", auth.NewService(processor))
	return eng
}

// ============================
// Config APIs
// ============================

// AddConfig set the global config.
// 設置global config後初始化所有資料庫連線(設置Engine.Services)並啟動引擎
func (eng *Engine) AddConfig(cfg config.Config) *Engine {
	// setConfig設置Config(struct)title、theme、登入url、前綴url...資訊
	// InitDatabase初始化所有資料庫連線(將driver加入Engine.Services)並啟動引擎
	return eng.setConfig(cfg).InitDatabase()
}

// setConfig set the config of engine.
// 設置Engine.config
func (eng *Engine) setConfig(cfg config.Config) *Engine {
	eng.config = config.Set(cfg)
	sysCheck, themeCheck := template.CheckRequirements()
	if !sysCheck {
		panic(fmt.Sprintf("wrong GoAdmin version, theme %s required GoAdmin version are %s",
			eng.config.Theme, strings.Join(template.Default().GetRequirements(), ",")))
	}
	if !themeCheck {
		panic(fmt.Sprintf("wrong Theme version, GoAdmin %s required Theme version are %s",
			system.Version(), strings.Join(system.RequireThemeVersion()[eng.config.Theme], ",")))
	}
	return eng
}

// AddConfigFromJSON set the global config from json file.
func (eng *Engine) AddConfigFromJSON(path string) *Engine {
	return eng.setConfig(config.ReadFromJson(path)).InitDatabase()
}

// AddConfigFromYAML set the global config from yaml file.
func (eng *Engine) AddConfigFromYAML(path string) *Engine {
	return eng.setConfig(config.ReadFromYaml(path)).InitDatabase()
}

// AddConfigFromINI set the global config from ini file.
func (eng *Engine) AddConfigFromINI(path string) *Engine {
	return eng.setConfig(config.ReadFromINI(path)).InitDatabase()
}

// InitDatabase initialize all database connection.
// 初始化所有資料庫連線(將driver加入Engine.Services)並啟動引擎
func (eng *Engine) InitDatabase() *Engine {
	// GroupByDriver將資料庫依照資料庫引擎分組(ex:mysql一組mssql一組)
	// driver = mysql、mssql等引擎名稱
	for driver, databaseCfg := range eng.config.Databases.GroupByDriver() {
		// 列出所有資料庫引擎加入Services
		// Add藉由參數新增List(map[string]Service)，在List加入引擎(driver)
		// GetConnectionByDriver藉由參數(driver = mysql、mssql...)取得Connection(interface)
		// InitDB初始化資料庫連線並啟動引擎
		eng.Services.Add(driver, db.GetConnectionByDriver(driver).InitDB(databaseCfg))
	}
	if defaultAdapter == nil {
		panic("adapter is nil")
	}
	return eng
}

// AddAdapter add the adapter of engine.
// 設定Engine.Adapter與defaultAdapter，回傳設定Engine(struct)
func (eng *Engine) AddAdapter(ada adapter.WebFrameWork) *Engine {
	eng.Adapter = ada
	defaultAdapter = ada
	return eng
}

// defaultAdapter is the default adapter of engine.
// 預設的配飾器(adapter.WebFrameWork(interface))
var defaultAdapter adapter.WebFrameWork

var engine *Engine

// navButtons is the default buttons in the navigation bar.
// 預設的Buttons(interface)
var navButtons = new(types.Buttons)

// Register set default adapter of engine.
// 建立引擎預設的配適器
func Register(ada adapter.WebFrameWork) {
	if ada == nil {
		panic("adapter is nil")
	}
	defaultAdapter = ada
}

// User call the User method of defaultAdapter.
// 回傳adapter(adapter.WebFrameWork(interface))的User方法，在adapter\gin\gin.go中
// 從ctx中取得cookie，接著利用cookie可以取得用戶角色、權限以及可用menu，最後得到UserModel，但UserModel.Base.Conn = nil(因ReleaseConn方法)
func User(ctx interface{}) (models.UserModel, bool) {
	return defaultAdapter.User(ctx)
}

// User call the User method of engine adapter.
// 回傳Engine.Adapter(adapter.WebFrameWork(interface))的User方法，在adapter\gin\gin.go中
// 從ctx中取得cookie，接著利用cookie可以取得用戶角色、權限以及可用menu，最後得到UserModel，但UserModel.Base.Conn = nil(因ReleaseConn方法)
func (eng *Engine) User(ctx interface{}) (models.UserModel, bool) {
	return eng.Adapter.User(ctx)
}

// ============================
// DB Connection APIs
// ============================

// DB return the db connection of given driver.
// 透過參數(driver)找到匹配的Service(interface)後回傳Connection(interface)型態
// Service(interface)也具有db.Connection(interface)型態，因為都具有Name方法
func (eng *Engine) DB(driver string) db.Connection {
	// GetConnectionFromService在modules\db\connection.go
	// Get在modules\service\service.go中
	// GetConnectionFromService將參數型態轉換成Connection(interface)後回傳
	// Get藉由參數(driver)取得匹配的Service(interface)
	return db.GetConnectionFromService(eng.Services.Get(driver))
}

// DefaultConnection return the default db connection.
// 回傳預設的Connection(interface)
func (eng *Engine) DefaultConnection() db.Connection {
	// GetDefault() = DatabaseList["default"]
	// 參數為DatabaseList["default"].driver
	return eng.DB(eng.config.Databases.GetDefault().Driver)
}

// MysqlConnection return the mysql db connection of given driver.
// 取得匹配mysql的Service後回傳Connection(interface)
// Service(interface)也具有db.Connection(interface)型態，因為都具有Name方法
func (eng *Engine) MysqlConnection() db.Connection {
	return db.GetConnectionFromService(eng.Services.Get(db.DriverMysql))
}

// MssqlConnection return the mssql db connection of given driver.
// 取得匹配mssql的Service後回傳Connection(interface)
// Service(interface)也具有db.Connection(interface)型態，因為都具有Name方法
func (eng *Engine) MssqlConnection() db.Connection {
	return db.GetConnectionFromService(eng.Services.Get(db.DriverMssql))
}

// PostgresqlConnection return the postgresql db connection of given driver.
// 取得匹配postgresql的Service後回傳Connection(interface)
// Service(interface)也具有db.Connection(interface)型態，因為都具有Name方法
func (eng *Engine) PostgresqlConnection() db.Connection {
	return db.GetConnectionFromService(eng.Services.Get(db.DriverPostgresql))
}

// SqliteConnection return the sqlite db connection of given driver.
// 取得匹配sqlite的Service後回傳Connection(interface)
// Service(interface)也具有db.Connection(interface)型態，因為都具有Name方法
func (eng *Engine) SqliteConnection() db.Connection {
	return db.GetConnectionFromService(eng.Services.Get(db.DriverSqlite))
}

// 連接設置器
type ConnectionSetter func(db.Connection)

// ResolveConnection resolve the specified driver connection.
// 解決特別driver連接
func (eng *Engine) ResolveConnection(setter ConnectionSetter, driver string) *Engine {
	setter(eng.DB(driver))
	return eng
}

// ResolveMysqlConnection resolve the mysql connection.
// 解決mysql連接
func (eng *Engine) ResolveMysqlConnection(setter ConnectionSetter) *Engine {
	eng.ResolveConnection(setter, db.DriverMysql)
	return eng
}

// ResolveMssqlConnection resolve the mssql connection.
func (eng *Engine) ResolveMssqlConnection(setter ConnectionSetter) *Engine {
	// ConnectionSetter是func(db.Connection)
	// db.DsiverMysql = mssql
	eng.ResolveConnection(setter, db.DriverMssql)
	return eng
}

// ResolveSqliteConnection resolve the sqlite connection.
func (eng *Engine) ResolveSqliteConnection(setter ConnectionSetter) *Engine {
	// ConnectionSetter是func(db.Connection)
	// db.DriverSqlite = sqlite
	eng.ResolveConnection(setter, db.DriverSqlite)
	return eng
}

// ResolvePostgresqlConnection resolve the postgres connection.
func (eng *Engine) ResolvePostgresqlConnection(setter ConnectionSetter) *Engine {
	// ConnectionSetter是func(db.Connection)
	// db.DriverPostgresql = postgresql
	eng.ResolveConnection(setter, db.DriverPostgresql)
	return eng
}

type Setter func(*Engine)

// Clone copy a new Engine.
// 複製一個新Engine
func (eng *Engine) Clone(e *Engine) *Engine {
	e = eng
	return eng
}

// ClonedBySetter copy a new Engine by a setter callback function.
// 透過Setter(function)複製一個新Engine
func (eng *Engine) ClonedBySetter(setter Setter) *Engine {
	setter(eng)
	return eng
}

// 回傳記錄使用者操作行為的資料表後將參數(conn)設置至OperationLogModel(struct)，最後新增一筆資料(操作紀錄)
func (eng *Engine) deferHandler(conn db.Connection) context.Handler {
	return func(ctx *context.Context) {
		defer func(ctx *context.Context) {
			// 尋找cts(struct).UserValue裡user的值後轉換成UserModel(struct)型態
			if user, ok := ctx.UserValue["user"].(models.UserModel); ok {
				var input []byte
				// 解析form-data參數
				form := ctx.Request.MultipartForm
				if form != nil {
					input, _ = json.Marshal((*form).Value)
				}
				// 在plugins\admin\models\operation_log.go中
				// OperationLog回傳預設的OperationLogModel(struct)，記錄使用者操作行為，預設資料表為goadmin_operation_log
				// SetConn將參數conn(Connection(interface))設置至OperationLogModel.Base.Conn(struct)
				// New新增一筆使用者行為資料至資料表，回傳OperationLogModel(struct)
				models.OperationLog().SetConn(conn).New(user.Id, ctx.Path(), ctx.Method(), ctx.LocalIP(), string(input))
			}

			// 出現錯誤
			if err := recover(); err != nil {
				logger.Error(err)
				logger.Error(string(debug.Stack()[:]))

				var (
					errMsg string
					ok     bool
					e      error
				)

				if errMsg, ok = err.(string); !ok {
					if e, ok = err.(error); ok {
						errMsg = e.Error()
					}
				}

				if errMsg == "" {
					errMsg = "system error"
				}

				if ctx.WantJSON() {
					response.Error(ctx, errMsg)
					return
				}

				eng.errorPanelHTML(ctx, new(bytes.Buffer), errors2.New(errMsg))
			}
		}(ctx)

		// Next在context\context.go中
		// Next只在middleware中使用
		// 執行多次func(ctx *Context)
		ctx.Next()
	}
}

// wrapWithAuthMiddleware wrap a auth middleware to the given handler.
// 回傳的類別為Handlers([]Handler)，Handler類別為func(ctx *Context)，需身分驗證
func (eng *Engine) wrapWithAuthMiddleware(handler context.Handler) context.Handlers {
	// GetConnection取得匹配的service.List然後轉換成Connection(interface)類別
	conn := db.GetConnection(eng.Services)
	// deferHandler回傳記錄使用者操作行為的資料表並將參數(conn)設置至OperationLogModel(struct)，最後新增一筆資料(操作紀錄)
	return []context.Handler{eng.deferHandler(conn), response.OffLineHandler, auth.Middleware(conn), handler}
}

// wrapWithAuthMiddleware wrap a auth middleware to the given handler.
// 回傳的類別為Handlers([]Handler)，Handler類別為func(ctx *Context)，不須身分驗證
func (eng *Engine) wrap(handler context.Handler) context.Handlers {
	// GetConnection取得匹配的service.List然後轉換成Connection(interface)類別
	conn := db.GetConnection(eng.Services)
	// deferHandler回傳記錄使用者操作行為的資料表並將參數(conn)設置至OperationLogModel(struct)，最後新增一筆資料(操作紀錄)
	return []context.Handler{eng.deferHandler(conn), response.OffLineHandler, handler}
}

// ============================
// HTML Content Render APIs
// ============================

// AddNavButtons add the nav buttons.
// Action是interface
// 新增一個NavButton(struct)並設置至Engine.NavButtons
func (eng *Engine) AddNavButtons(title template2.HTML, icon string, action types.Action) *Engine {
	// GetNavButton在template\types\button.go中
	// GetNavButton回傳NavButton(struct)設置資訊
	btn := types.GetNavButton(title, icon, action)
	*eng.NavButtons = append(*eng.NavButtons, btn)
	return eng
}

// Content call the Content method of engine adapter.
// If adapter is nil, it will panic.
// Engine.Adapter(interface)不能為空，利用cookie驗證使用者，取得role、permission、menu，接著檢查權限，執行模板並導入HTML
func (eng *Engine) Content(ctx interface{}, panel types.GetPanelFn) {
	if eng.Adapter == nil {
		panic("adapter is nil")
	}
	// Content方法在gin/gin.go中
	// 添加html到框架中
	// 利用cookie驗證使用者，取得role、permission、menu，接著檢查權限，執行模板並導入HTML
	// Content為adapter.WebFrameWork(interface)方法(在gin/gin.go中)
	eng.Adapter.Content(ctx, panel, eng.AdminPlugin().GetAddOperationFn(), *eng.NavButtons...)
}

// Content call the Content method of defaultAdapter.
// If defaultAdapter is nil, it will panic.
// Engine.Adapter(interface)不能為空，利用cookie驗證使用者，取得role、permission、menu，接著檢查權限，執行模板並導入HTML
func Content(ctx interface{}, panel types.GetPanelFn) {
	if defaultAdapter == nil {
		panic("adapter is nil")
	}
	// Content方法在gin/gin.go中
	// Engine.Adapter(interface)不能為空，// 利用cookie驗證使用者，取得role、permission、menu，接著檢查權限，執行模板並導入HTML
	// Content為adapter.WebFrameWork(interface)方法(在gin/gin.go中)
	defaultAdapter.Content(ctx, panel, engine.AdminPlugin().GetAddOperationFn(), *navButtons...)
}

// Data inject the route and corresponding handler to the web framework.
// 將route以及相對應的處理程序加入Web框架
// 設置context.Context增加handlers、處理url及寫入header，最後取得新的request handle與middleware
func (eng *Engine) Data(method, url string, handler context.Handler, noAuth ...bool) {
	// AddHandler藉由method、url增加處理程序(Handler)
	// 設置context.Context增加handlers、處理url及寫入header，最後取得新的request handle 與middleware
	if len(noAuth) > 0 && noAuth[0] {
		// wrap回傳的類別為Handlers([]Handler)，Handler類別為func(ctx *Context)
		eng.Adapter.AddHandler(method, url, eng.wrap(handler))
	} else {
		// wrapWithAuthMiddleware回傳的類別為Handlers([]Handler)，Handler類別為func(ctx *Context)
		eng.Adapter.AddHandler(method, url, eng.wrapWithAuthMiddleware(handler))
	}
}


// HTML inject the route and corresponding handler wrapped by the given function to the web framework.
// type GetPanelInfoFn func(ctx *context.Context) (Panel, error)
// 透過function將route以及相對應的處理程序加入Web框架
// 建立一個handler後設置context.Context增加handlers、處理url及寫入header，最後取得新的request handle與middleware
func (eng *Engine) HTML(method, url string, fn types.GetPanelInfoFn, noAuth ...bool) {

	var handler = func(ctx *context.Context) {
		panel, err := fn(ctx)
		if err != nil {
			panel = template.WarningPanel(err.Error())
		}

		eng.AdminPlugin().GetAddOperationFn()(panel.Callbacks...)

		// Default如果主題名稱已經通過全局配置，取得預設的Template(interface)
		// GetTemplate為Template(interface)的方法
		tmpl, tmplName := template.Default().GetTemplate(ctx.IsPjax())

		// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
		user := auth.Auth(ctx)

		buf := new(bytes.Buffer)

		// ExecuteTemplate執行模板(html\template\template.go中Template的方法)
		// 藉由參數tmplName應用模板到指定的對象(第三個參數)，並將結果輸出寫入buf
		hasError := tmpl.ExecuteTemplate(buf, tmplName, types.NewPage(types.NewPageParam{
			User:         user,
			// GetGlobalMenu在modules\menu\menu.go中
			// GetGlobalMenu回傳給定user的Menu(設置menuList、menuOption、MaxOrder)
			// Menu(struct)包含List、Options、MaxOrder
			Menu:         menu.GetGlobalMenu(user, eng.Adapter.GetConnection()).SetActiveClass(config.URLRemovePrefix(ctx.Path())),
			// IsProductionEnvironment檢查生產環境
			// GetContent在template\types\page.go
			// Panel(struct)主要內容使用pjax的模板
			// GetContent獲取內容(設置前端HTML)，設置Panel並回傳
			Panel:        panel.GetContent(eng.config.IsProductionEnvironment()),
			// Assets類別為template.HTML(string)
			// 處理asset後並回傳HTML語法
			Assets:       template.GetComponentAssetImportHTML(),
			// 檢查權限，回傳Buttons([]Button(interface))
			// 在template\types\button.go
			Buttons:      eng.NavButtons.CheckPermission(user),
			TmplHeadHTML: template.Default().GetHeadHTML(),
			TmplFootJS:   template.Default().GetFootJS(),
		}))

		if hasError != nil {
			logger.Error(fmt.Sprintf("error: %s adapter content, ", eng.Adapter.Name()), hasError)
		}

		// 輸出HTML，將buf.Bytes()參數設置至Context.response.Body
		ctx.HTMLByte(http.StatusOK, buf.Bytes())
	}

	// AddHandler藉由method、url增加處理程序(Handler)
	// 設置context.Context增加handlers、處理url及寫入header，最後取得新的request與middleware
	if len(noAuth) > 0 && noAuth[0] {
		// wrap回傳的類別為Handlers([]Handler)，Handler類別為func(ctx *Context)
		eng.Adapter.AddHandler(method, url, eng.wrap(handler))
	} else {
		// wrapWithAuthMiddleware回傳的類別為Handlers([]Handler)，Handler類別為func(ctx *Context)
		eng.Adapter.AddHandler(method, url, eng.wrapWithAuthMiddleware(handler))
	}
}

// HTMLFile inject the route and corresponding handler which returns the panel content of given html file path
// to the web framework.
// 將route以及相對應的處理程序加入Web框架(該程序回傳html文件的面板(panel)內容)
// 建立一個handler後設置context.Context增加handlers、處理url及寫入header，最後取得新的request handle與middleware
func (eng *Engine) HTMLFile(method, url, path string, data map[string]interface{}, noAuth ...bool) {

	var handler = func(ctx *context.Context) {

		cbuf := new(bytes.Buffer)

		// ParseFiles在html/template套件
		// 解析文件並創建新的Template(struct)
		t, err := template2.ParseFiles(path)
		if err != nil {
			eng.errorPanelHTML(ctx, cbuf, err)
			return
		} else {
			if err := t.Execute(cbuf, data); err != nil {
				eng.errorPanelHTML(ctx, cbuf, err)
				return
			}
		}

		// Default如果主題名稱已經通過全局配置，取得預設的Template(interface)
		// GetTemplate為Template(interface)的方法
		tmpl, tmplName := template.Default().GetTemplate(ctx.IsPjax())

		// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
		user := auth.Auth(ctx)

		buf := new(bytes.Buffer)

		// ExecuteTemplate執行模板(html\template\template.go中Template的方法)
		// 藉由參數tmplName應用模板到指定的對象(第三個參數)，並將結果輸出寫入buf
		hasError := tmpl.ExecuteTemplate(buf, tmplName, types.NewPage(types.NewPageParam{
			User: user,
			// GetGlobalMenu在modules\menu\menu.go中
			// GetGlobalMenu回傳給定user的Menu(設置menuList、menuOption、MaxOrder)
			// Menu(struct)包含List、Options、MaxOrder
			Menu: menu.GetGlobalMenu(user, eng.Adapter.GetConnection()).SetActiveClass(eng.config.URLRemovePrefix(ctx.Path())),
			Panel: types.Panel{
				Content: template.HTML(cbuf.String()),
			},
			// Assets類別為template.HTML(string)
			// 處理asset後並回傳HTML語法
			Assets:       template.GetComponentAssetImportHTML(),
			// 檢查權限，回傳Buttons([]Button(interface))
			// 在template\types\button.go
			Buttons:      eng.NavButtons.CheckPermission(user),
			TmplHeadHTML: template.Default().GetHeadHTML(),
			TmplFootJS:   template.Default().GetFootJS(),
		}))

		if hasError != nil {
			logger.Error(fmt.Sprintf("error: %s adapter content, ", eng.Adapter.Name()), hasError)
		}

		// 輸出HTML，將buf.Bytes()參數設置至Context.response.Body
		ctx.HTMLByte(http.StatusOK, buf.Bytes())
	}

	// AddHandler藉由method、url增加處理程序(Handler)
	// 設置context.Context增加handlers、處理url及寫入header，最後取得新的request與middleware
	if len(noAuth) > 0 && noAuth[0] {
		// wrap回傳的類別為Handlers([]Handler)，Handler類別為func(ctx *Context)
		eng.Adapter.AddHandler(method, url, eng.wrap(handler))
	} else {
		// wrapWithAuthMiddleware回傳的類別為Handlers([]Handler)，Handler類別為func(ctx *Context)
		eng.Adapter.AddHandler(method, url, eng.wrapWithAuthMiddleware(handler))
	}
}

// HTMLFiles inject the route and corresponding handler which returns the panel content of given html files path
// to the web framework.
// 將route以及相對應的處理程序加入Web框架(該程序回傳html文件(多個文件)的面板(panel)內容)，需身分驗證
// 設置context.Context增加handlers、處理url及寫入header，最後取得新的request handle與middleware
func (eng *Engine) HTMLFiles(method, url string, data map[string]interface{}, files ...string) {
	// wrapWithAuthMiddleware回傳的類別為Handlers([]Handler)，Handler類別為func(ctx *Context)
	eng.Adapter.AddHandler(method, url, eng.wrapWithAuthMiddleware(eng.htmlFilesHandler(data, files...)))
}

// HTMLFilesNoAuth inject the route and corresponding handler which returns the panel content of given html files path
// to the web framework without auth check.
// 將route以及相對應的處理程序加入Web框架(該程序回傳html文件(多個文件)的面板(panel)內容)，無須身分驗證
// 設置context.Context增加handlers、處理url及寫入header，最後取得新的request handle與middleware
func (eng *Engine) HTMLFilesNoAuth(method, url string, data map[string]interface{}, files ...string) {
	// wrap回傳的類別為Handlers([]Handler)，Handler類別為func(ctx *Context)
	eng.Adapter.AddHandler(method, url, eng.wrap(eng.htmlFilesHandler(data, files...)))
}

// HTMLFiles inject the route and corresponding handler which returns the panel content of given html files path
// to the web framework.
// 解析多個文件建立並執行Template
func (eng *Engine) htmlFilesHandler(data map[string]interface{}, files ...string) context.Handler {
	return func(ctx *context.Context) {

		cbuf := new(bytes.Buffer)

		//解析多個文件
		t, err := template2.ParseFiles(files...)
		if err != nil {
			eng.errorPanelHTML(ctx, cbuf, err)
			return
		} else {
			if err := t.Execute(cbuf, data); err != nil {
				eng.errorPanelHTML(ctx, cbuf, err)
				return
			}
		}

		// Default如果主題名稱已經通過全局配置，取得預設的Template(interface)
		// GetTemplate為Template(interface)的方法
		tmpl, tmplName := template.Default().GetTemplate(ctx.IsPjax())

		// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
		user := auth.Auth(ctx)

		buf := new(bytes.Buffer)

		// ExecuteTemplate執行模板(html\template\template.go中Template的方法)
		// 藉由參數tmplName應用模板到指定的對象(第三個參數)，並將結果輸出寫入buf
		hasError := tmpl.ExecuteTemplate(buf, tmplName, types.NewPage(types.NewPageParam{
			User: user,
			// GetGlobalMenu在modules\menu\menu.go中
			// GetGlobalMenu回傳給定user的Menu(設置menuList、menuOption、MaxOrder)
			// Menu(struct)包含List、Options、MaxOrder
			Menu: menu.GetGlobalMenu(user, eng.Adapter.GetConnection()).SetActiveClass(eng.config.URLRemovePrefix(ctx.Path())),
			Panel: types.Panel{
				Content: template.HTML(cbuf.String()),
			},
			// Assets類別為template.HTML(string)
			// 處理asset後並回傳HTML語法
			Assets:       template.GetComponentAssetImportHTML(),
			Buttons:      eng.NavButtons.CheckPermission(user),
			// 檢查權限，回傳Buttons([]Button(interface))
			// 在template\types\button.go
			TmplHeadHTML: template.Default().GetHeadHTML(),
			TmplFootJS:   template.Default().GetFootJS(),
		}))

		if hasError != nil {
			logger.Error(fmt.Sprintf("error: %s adapter content, ", eng.Adapter.Name()), hasError)
		}

		// 輸出HTML，將buf.Bytes()參數設置至Context.response.Body
		ctx.HTMLByte(http.StatusOK, buf.Bytes())
	}
}

// errorPanelHTML add an error panel html to context response.
// 加入錯誤panel至response
func (eng *Engine) errorPanelHTML(ctx *context.Context, buf *bytes.Buffer, err error) {

	// 透過參數ctx回傳目前登入的用戶(Context.UserValue["user"])並轉換成UserModel
	user := auth.Auth(ctx)

	// Default如果主題名稱已經通過全局配置，取得預設的Template(interface)
	// GetTemplate為Template(interface)的方法
	tmpl, tmplName := template.Default().GetTemplate(ctx.IsPjax())

	// ExecuteTemplate執行模板(html\template\template.go中Template的方法)
	// 藉由參數tmplName應用模板到指定的對象(第三個參數)，並將結果輸出寫入buf
	hasError := tmpl.ExecuteTemplate(buf, tmplName, types.NewPage(types.NewPageParam{
		User:         user,
		// GetGlobalMenu在modules\menu\menu.go中
		// GetGlobalMenu回傳給定user的Menu(設置menuList、menuOption、MaxOrder)
		// Menu(struct)包含List、Options、MaxOrder
		Menu:         menu.GetGlobalMenu(user, eng.Adapter.GetConnection()).SetActiveClass(eng.config.URLRemovePrefix(ctx.Path())),
		Panel:        template.WarningPanel(err.Error()).GetContent(eng.config.IsProductionEnvironment()),
		// Assets類別為template.HTML(string)
		// 處理asset後並回傳HTML語法
		Assets:       template.GetComponentAssetImportHTML(),
		// 檢查權限，回傳Buttons([]Button(interface))
		// 在template\types\button.go
		Buttons:      (*eng.NavButtons).CheckPermission(user),
		TmplHeadHTML: template.Default().GetHeadHTML(),
		TmplFootJS:   template.Default().GetFootJS(),
	}))

	if hasError != nil {
		logger.Error(fmt.Sprintf("error: %s adapter content, ", eng.Adapter.Name()), hasError)
	}

	// 輸出HTML，將buf.Bytes()參數設置至Context.response.Body
	ctx.HTMLByte(http.StatusOK, buf.Bytes())
}

// ============================
// Admin Plugin APIs
// ============================

// AddGenerators add the admin generators.
// 判斷plug是否存在，如存在則透過參數LIST(多個)判斷GeneratorList已經有該key、value，如果不存在則加入該鍵與值至Admin.tableList
// 如不存在，設置的Admin(struct)加至Engine.PluginList
func (eng *Engine) AddGenerators(list ...table.GeneratorList) *Engine {
	// 透過參數(name)找到符合的plugin(interface)
	plug, exist := eng.FindPluginByName("admin")
	if exist {
		// plugins\admin\admin.go
		// 將plug轉換成Admin(struct)類別
		// 透過參數LIST(多個)判斷GeneratorList已經有該key、value，如果不存在則加入該鍵與值至Admin.tableList
		plug.(*admin.Admin).AddGenerators(list...)
		return eng
	}
	// NewAdmin設置Admin(struct)並回傳
	eng.PluginList = append(eng.PluginList, admin.NewAdmin(list...))
	return eng
}

// AdminPlugin get the admin plugin. if not exist, create one.
// 判斷plug是否存在，如存在則將plug轉換成Admin(struct)類別並回傳
// 如不存在，設置的Admin(struct)加至Engine.PluginList並回傳Admin(struct)
func (eng *Engine) AdminPlugin() *admin.Admin {
	// 透過參數(name)找到符合的plugin(interface)
	plug, exist := eng.FindPluginByName("admin")
	if exist {
		// plugins\admin\admin.go
		// 將plug轉換成Admin(struct)類別
		return plug.(*admin.Admin)
	}
	// NewAdmin設置Admin(struct)並回傳
	adm := admin.NewAdmin()
	eng.PluginList = append(eng.PluginList, adm)
	return adm
}

// SetCaptcha set the captcha config.
// 將參數captcha(驗證碼)設置至Admin.handler.captchaConfig(struct)
func (eng *Engine) SetCaptcha(captcha map[string]string) *Engine {
	// AdminPlugin判斷plug是否存在，如存在則將plug轉換成Admin(struct)類別並回傳
	// 如不存在，設置的Admin(struct)加至Engine.PluginList並回傳Admin(struct)
	// SetCaptcha將參數captcha(驗證碼)設置至Admin.handler.captchaConfig(struct)
	eng.AdminPlugin().SetCaptcha(captcha)
	return eng
}

// SetCaptchaDriver set the captcha config with driver.
// SetCaptchaDriver將參數map[string]string{"driver": driver}(驗證碼)設置至Admin.handler.captchaConfig(struct)
func (eng *Engine) SetCaptchaDriver(driver string) *Engine {
	// AdminPlugin判斷plug是否存在，如存在則將plug轉換成Admin(struct)類別並回傳
	// 如不存在，設置的Admin(struct)加至Engine.PluginList並回傳Admin(struct)
	// SetCaptcha將參數map[string]string{"driver": driver}(驗證碼)設置至Admin.handler.captchaConfig(struct)
	eng.AdminPlugin().SetCaptcha(map[string]string{"driver": driver})
	return eng
}

// AddGenerator add table model generator.
// AddGenerator將參數key及g(function)添加至Admin.tableList(map[string]Generator)
func (eng *Engine) AddGenerator(key string, g table.Generator) *Engine {
	// AdminPlugin判斷plug是否存在，如存在則將plug轉換成Admin(struct)類別並回傳
	// 如不存在，設置的Admin(struct)加至Engine.PluginList並回傳Admin(struct)
	// AddGenerator將參數key及g(function)添加至Admin.tableList(map[string]Generator)
	eng.AdminPlugin().AddGenerator(key, g)
	return eng
}

// AddGlobalDisplayProcessFn call types.AddGlobalDisplayProcessFn.
// 將參數f(func(string) string)加入globalDisplayProcessChains([]DisplayProcessFn)
func (eng *Engine) AddGlobalDisplayProcessFn(f types.FieldFilterFn) *Engine {
	// 在template\types\display.go
	// 將參數f(func(string) string)加入globalDisplayProcessChains([]DisplayProcessFn)
	types.AddGlobalDisplayProcessFn(f)
	return eng
}

// AddDisplayFilterLimit call types.AddDisplayFilterLimit.
// 加入func(value string) string至FieldDisplay.DisplayProcessFnChains([]DisplayProcessFn)
// 透過參數limit判斷func(value string)回傳的值
func (eng *Engine) AddDisplayFilterLimit(limit int) *Engine {
	// 在template\types\display.go
	// 加入func(value string) string至DisplayProcessFnChains([]DisplayProcessFn)
	types.AddLimit(limit)
	return eng
}

// AddDisplayFilterTrimSpace call types.AddDisplayFilterTrimSpace.
// 加入func(value string) string至DisplayProcessFnChains([]DisplayProcessFn)
// func(value string)回傳值為strings.TrimSpace(value)
func (eng *Engine) AddDisplayFilterTrimSpace() *Engine {
	types.AddTrimSpace()
	return eng
}

// AddDisplayFilterSubstr call types.AddDisplayFilterSubstr.
// 加入func(value string) string至參數globalDisplayProcessChains([]DisplayProcessFn)
// 透過參數start、end判斷func(value string)回傳的值
func (eng *Engine) AddDisplayFilterSubstr(start int, end int) *Engine {
	types.AddSubstr(start, end)
	return eng
}

// AddDisplayFilterToTitle call types.AddDisplayFilterToTitle.
// 加入func(value string) string至globalDisplayProcessChains([]DisplayProcessFn)
// func(value string)回傳值為strings.Title(value)
func (eng *Engine) AddDisplayFilterToTitle() *Engine {
	types.AddToTitle()
	return eng
}

// AddDisplayFilterToUpper call types.AddDisplayFilterToUpper.
// 加入func(value string) string至globalDisplayProcessChains([]DisplayProcessFn)
// func(value string)回傳值為strings.ToUpper(value)
func (eng *Engine) AddDisplayFilterToUpper() *Engine {
	types.AddToUpper()
	return eng
}

// AddDisplayFilterToLower call types.AddDisplayFilterToLower.
// 加入func(value string) string至globalDisplayProcessChains([]DisplayProcessFn)
// func(value string)回傳值為strings.ToLower(value)
func (eng *Engine) AddDisplayFilterToLower() *Engine {
	types.AddToUpper()
	return eng
}

// AddDisplayFilterXssFilter call types.AddDisplayFilterXssFilter.
// 加入func(value string) string至globalDisplayProcessChains([]DisplayProcessFn)
// func(value string)回傳值為html.EscapeString(value)
func (eng *Engine) AddDisplayFilterXssFilter() *Engine {
	types.AddXssFilter()
	return eng
}

// AddDisplayFilterXssJsFilter call types.AddDisplayFilterXssJsFilter.
// 加入func(value string) string至globalDisplayProcessChains([]DisplayProcessFn)
// func(value string)回傳值為replacer.Replace(value)
func (eng *Engine) AddDisplayFilterXssJsFilter() *Engine {
	types.AddXssJsFilter()
	return eng
}
