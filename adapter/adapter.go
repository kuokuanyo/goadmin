// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package adapter

import (
	"bytes"
	"fmt"
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/menu"
	"github.com/GoAdminGroup/go-admin/plugins"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"net/url"
)

// WebFrameWork功能都設定在框架中(使用/adapter/gin/gin.go框架)
// WebFrameWork is an interface which is used as an adapter of
// framework and goAdmin. It must implement two methods. Use registers
// the routes and the corresponding handlers. Content writes the
// response to the corresponding context of framework.
type WebFrameWork interface {
	// Name return the web framework name.
	// 回傳使用的web框架名稱
	Name() string

	// Use method inject the plugins to the web framework engine which is the
	// first parameter.
	// 將插件插入框架引擎中
	Use(app interface{}, plugins []plugins.Plugin) error

	// Content add the panel html response of the given callback function to
	// the web framework context which is the first parameter.
	// 添加html到框架中
	Content(ctx interface{}, fn types.GetPanelFn, fn2 context.NodeProcessor, navButtons ...types.Button)

	// User get the auth user model from the given web framework context.
	// 從給定的上下文中取得用戶模型
	User(ctx interface{}) (models.UserModel, bool)

	// AddHandler inject the route and handlers of GoAdmin to the web framework.
	// 將路由(路徑)及處理程式加入框架
	AddHandler(method, path string, handlers context.Handlers)

	DisableLog()

	Static(prefix, path string)

	Run() error

	// Helper functions
	// ================================

	SetApp(app interface{}) error
	SetConnection(db.Connection)
	GetConnection() db.Connection
	SetContext(ctx interface{}) WebFrameWork
	GetCookie() (string, error)
	Path() string
	Method() string
	FormParam() url.Values
	IsPjax() bool
	Redirect()
	SetContentType()
	Write(body []byte)
	CookieKey() string
	HTMLContentType() string
}

// 基本配飾器包含輔助功能
// db.Connection(interface)
// BaseAdapter is a base adapter contains some helper functions.
type BaseAdapter struct {
	db db.Connection
}

// 設定資料庫連接
// SetConnection set the db connection.
func (base *BaseAdapter) SetConnection(conn db.Connection) {
	base.db = conn
}

// 取得連線
// GetConnection get the db connection.
func (base *BaseAdapter) GetConnection() db.Connection {
	return base.db
}

// 回傳預設的content type
// HTMLContentType return the default content type header.
func (base *BaseAdapter) HTMLContentType() string {
	return "text/html; charset=utf-8"
}

// CookieKey return the cookie key.
func (base *BaseAdapter) CookieKey() string {
	// auth.DefaultCookieKey = go_admin_session
	return auth.DefaultCookieKey
}

// 從上下文(參數ctx)中取得用戶模型(UserModel)，但UserModel.Base.Conn = nil(因ReleaseConn方法)
// 取得用戶角色、權限以及可用menu
// GetUser is a helper function get the auth user model from the context.
func (base *BaseAdapter) GetUser(ctx interface{}, wf WebFrameWork) (models.UserModel, bool) {
	// 取得cookie
	cookie, err := wf.SetContext(ctx).GetCookie()

	// models.UserModel在plugins/admin/modules/user.go中
	if err != nil {
		return models.UserModel{}, false
	}

	// auth.GetCurUser在modules/auth/middleware.go中，回傳使用者資訊
	// WebFrameWork.GetConnection()回傳BaseAdapter.db
	// 藉由cookie、conn可以得到角色、權限以及可使用菜單
	user, exist := auth.GetCurUser(cookie, wf.GetConnection())

	// ReleaseConn在plugins/admin/modules/user.go中
	// 將UserModel.Conn(UserModel.Base.Conn) = nil
	return user.ReleaseConn(), exist
}

// 增加插件至框架
// plugins.Plugin在plugins/plugins.go中，是interface
// WebFrameWork interface
// 藉由method、url增加處理程序(Handler)
// GetUse is a helper function adds the plugins to the framework.
func (base *BaseAdapter) GetUse(app interface{}, plugin []plugins.Plugin, wf WebFrameWork) error {
	// adapter\gin\gin.go中
	// 設置Gin.app(gin.Engine)
	if err := wf.SetApp(app); err != nil {
		return err
	}

	// 在plugins/plugins.go
	// plug interface
	for _, plug := range plugin {
		// 返回路由和控制器方法
		// GetHandler()在 context\context.go，類型map[Path]Handlers
		// path struct，包含url、method
		// handlers類型[]Handler， Handler類型 func(ctx *Context)
		for path, handlers := range plug.GetHandler() {
			// 執行方法(WebFrameWork interface)， adapter\gin\gin.go有設置該方法
			// 藉由method、url增加處理程序(Handler)
			// 設置context.Context與設置url與寫入header，取得新的request與middleware
			wf.AddHandler(path.Method, path.URL, handlers)
		}
	}

	return nil
}

// 利用cookie驗證使用者，取得role、permission、menu，接著檢查權限，執行模板並導入HTML
// GetContent is a helper function of adapter.Content
func (base *BaseAdapter) GetContent(ctx interface{}, getPanelFn types.GetPanelFn, wf WebFrameWork,
	navButtons types.Buttons, fn context.NodeProcessor) {

	var (
		// SetContext設置Gin.ctx(struct)
		newBase          = wf.SetContext(ctx)
		// 取的cookie value
		cookie, hasError = newBase.GetCookie()
	)

	// 如出現錯誤重新導向至登入頁面
	if hasError != nil || cookie == "" {
		newBase.Redirect()
		return
	}

	// GetCurUser回傳使用者模型，取得role、permission、menu
	// wf.GetConnection()回傳BaseAdapter.db(interface)
	user, authSuccess := auth.GetCurUser(cookie, wf.GetConnection())

	// 如出現錯誤重新導向至登入頁面
	if !authSuccess {
		newBase.Redirect()
		return
	}

	var (
		panel types.Panel
		err   error
	)

	// CheckPermissions檢查用戶權限(在modules\auth\middleware.go)
	if !auth.CheckPermissions(user, newBase.Path(), newBase.Method(), newBase.FormParam()) {
		// 沒有權限
		// errors.NoPermission = no permission
		panel = template.WarningPanel(errors.NoPermission, template.NoPermission403Page)
	} else {
		panel, err = getPanelFn(ctx)
		if err != nil {
			panel = template.WarningPanel(err.Error())
		}
	}

	fn(panel.Callbacks...)

	// Default()取得預設的template(主題名稱已經通過全局配置)
	// tmpl類別為template.Template(interface)，在template/template.go中
	// template.Template為ui組件的方法，將在plugins中自定義ui
	// IsPjax()在gin/gin.go中，設置標頭 X-PJAX = true
	// GetTemplate(bool)為template.Template(interface)的方法
	tmpl, tmplName := template.Default().GetTemplate(newBase.IsPjax())

	buf := new(bytes.Buffer)

	// ExecuteTemplate執行模板(html\template\template.go中Template的方法)
	// 藉由給的tmplName應用模板到指定的對象(第三個參數)
	hasError = tmpl.ExecuteTemplate(buf, tmplName, types.NewPage(types.NewPageParam{
		User:         user,
		// GetGlobalMenu 返回user的menu(modules\menu\menu.go中)
		// Menu(struct包含)List、Options、MaxOrder
		Menu:         menu.GetGlobalMenu(user, wf.GetConnection()).SetActiveClass(config.URLRemovePrefix(newBase.Path())),
		// IsProductionEnvironment檢查生產環境
		// GetContent在template\types\page.go
		// Panel(struct)主要內容使用pjax的模板
		// GetContent獲取內容(設置前端HTML)，設置Panel並回傳
		Panel:        panel.GetContent(config.IsProductionEnvironment()),
		// Assets類別為template.HTML(string)
		// 處理asset後並回傳HTML語法
		Assets:       template.GetComponentAssetImportHTML(),
		// 檢查權限，回傳Buttons([]Button(interface))
		// 在template\types\button.go
		Buttons:      navButtons.CheckPermission(user),
		TmplHeadHTML: template.Default().GetHeadHTML(),
		TmplFootJS:   template.Default().GetFootJS(),
	}))


	if hasError != nil {
		logger.Error(fmt.Sprintf("error: %s adapter content, ", newBase.Name()), hasError)
	}

	// 設置ContentType
	newBase.SetContentType()
	// 寫入
	newBase.Write(buf.Bytes())
}
