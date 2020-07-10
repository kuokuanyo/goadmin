// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package gin

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/GoAdminGroup/go-admin/adapter"
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/engine"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/plugins"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/gin-gonic/gin"
)

// Gin structure value is a Gin GoAdmin adapter.
// Gin同時也符合adapter.WebFrameWork(interface)
// Gin structure value is a Gin GoAdmin adapter.
type Gin struct {
	// adapter.BaseAdapter在adapter/adapter.go中
	// adapter.BaseAdapter(struct)裡面為db.Connection(interface)
	adapter.BaseAdapter
	// gin-gonic套件
	// gin.Context(struct)為gin最重要的部分，允許在middleware傳遞變數(例如驗證請求、管理流程)
	ctx *gin.Context
	// gin-gonic套件
	// app為框架中的實例，包含muxer,middleware ,configuration，藉由New() or Default()建立Engine
	app *gin.Engine
}

// 初始化
func init() {
	// 在engine\engine.go
	// 建立引擎預設的配適器
	engine.Register(new(Gin))
}

//-------------------------------------
// 下列為adapter.WebFrameWork(interface)的方法
// Gin(struct)也是adapter.WebFrameWork(interface)
//------------------------------------

// User implements the method Adapter.User.
// 從ctx中取得用戶模型(UserModel)，但UserModel.Base.Conn = nil(因ReleaseConn方法)
// User implements the method Adapter.User.
func (gins *Gin) User(ctx interface{}) (models.UserModel, bool) {
	// GetUser從ctx中取得用戶模型(為adapter.BaseAdapter的方法)
	// 取得用戶角色、權限以及可用menu
	return gins.GetUser(ctx, gins)
}

// plugins.Plugin在plugins/plugins.go中，是interface
// 增加處理程序(Handler)
// Use implements the method Adapter.Use.
func (gins *Gin) Use(app interface{}, plugs []plugins.Plugin) error {
	// GetUse增加插件至框架(為adapter.BaseAdapter的方法)
	// 增加處理程序(Handler)
	return gins.GetUse(app, plugs, gins)
}

// 利用cookie驗證使用者，取得role、permission、menu，接著檢查權限，執行模板並導入HTML
// Content implements the method Adapter.Content.
func (gins *Gin) Content(ctx interface{}, getPanelFn types.GetPanelFn, fn context.NodeProcessor, btns ...types.Button) {
	// GetContent在adapter\adapter.go
	// GetContent(Gin.adapter.BaseAdapter(struct)的方法)
	// 利用cookie驗證使用者，取得role、permission、menu，接著檢查權限，執行模板並導入HTML
	gins.GetContent(ctx, getPanelFn, gins, btns, fn)
}

type HandlerFunc func(ctx *gin.Context) (types.Panel, error)

// 添加html到框架中
// 利用cookie驗證使用者，接著檢查權限，建立模板並執行模板
func Content(handler HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Content在engine/engine.go中
		// Engine.Adapter不能為nil，接著添加html到框架中
		// 最後一樣執行上面的func (gins *Gin) Content(ctx inter...)函式
		engine.Content(ctx, func(ctx interface{}) (types.Panel, error) {
			return handler(ctx.(*gin.Context))
		})
	}
}

func (gins *Gin) Run() error                 { panic("not implement") }
func (gins *Gin) DisableLog()                { panic("not implement") }
func (gins *Gin) Static(prefix, path string) { panic("not implement") }

// 設置Gin.app(gin.Engine)，gin.Engine(gin-gonic套件)
// SetApp implements the method Adapter.SetApp.
func (gins *Gin) SetApp(app interface{}) error {
	var (
		eng *gin.Engine
		ok  bool
	)
	// app.(*gin.Engine)將interface{}轉換為gin.Engine型態
	if eng, ok = app.(*gin.Engine); !ok {
		return errors.New("gin adapter SetApp: wrong parameter")
	}
	gins.app = eng
	return nil
}

// 添加處理程序
// context.Handlers(context.context.go中)，Handlers類型為[]context.Handler，Handler類型為function(*context.Context)
// AddHandler藉由method、path增加處理程序(Handler)
// 設置context.Context增加handlers、處理url及寫入header，最後取得新的request handle與middleware
// AddHandler implements the method Adapter.AddHandler.
func (gins *Gin) AddHandler(method, path string, handlers context.Handlers) {

	// gins.app類型為gin.Engine
	// Handle方法為藉由path及method取得request handle與middleware，此功能為大量loading
	// Handle第三個參數(主要處理程序)為funcion(*gin.Context)，gin.Context為struct(gin-gonic套件)
	gins.app.Handle(strings.ToUpper(method), path, func(c *gin.Context) {

		// 設置新Context(struct)，設置Request(請求)以及UserValue、Response(預設的slice)
		// NewContext在context\context.go
		ctx := context.NewContext(c.Request)

		// Context.Params類型為[]Context.Param，Param裡有key以及value(他是url參數的鍵與值)
		// 將參數設置在url中
		for _, param := range c.Params {
			if c.Request.URL.RawQuery == "" {
				c.Request.URL.RawQuery += strings.Replace(param.Key, ":", "", -1) + "=" + param.Value
			} else {
				c.Request.URL.RawQuery += "&" + strings.Replace(param.Key, ":", "", -1) + "=" + param.Value
			}
		}

		// SetHandlers在context\context.go
		// 設置Handlers，將handlers設置至Context.handlers
		// Next只在middleware中使用
		ctx.SetHandlers(handlers).Next()

		// ctx.Response.Header 可能有多個鍵與值(map[string][]string)
		// Header()在gin-gonic套件中
		// Header()寫入header
		for key, head := range ctx.Response.Header {
			c.Header(key, head[0])
		}

		// 客戶與傳輸端Body不能為nil
		if ctx.Response.Body != nil {
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(ctx.Response.Body)
			c.String(ctx.Response.StatusCode, buf.String())
		} else {
			c.Status(ctx.Response.StatusCode)
		}
	})
}

// 回傳框架名稱
// Name implements the method Adapter.Name.
func (gins *Gin) Name() string {
	return "gin"
}

// 設置Gin.ctx(struct)
// SetContext implements the method Adapter.SetContext.
func (gins *Gin) SetContext(contextInterface interface{}) adapter.WebFrameWork {
	var (
		// gin.Context(struct)是gin-gonic套件
		ctx *gin.Context
		ok  bool
	)

	// 將contextInterface類別變成gin.Context(struct)
	if ctx, ok = contextInterface.(*gin.Context); !ok {
		panic("gin adapter SetContext: wrong parameter")
	}

	return &Gin{ctx: ctx}
}

// 重新導向至登入頁面(出現錯誤)
// Redirect implements the method Adapter.Redirect.
func (gins *Gin) Redirect() {
	// Redirect()為gin-gonic套件裡的方法
	// http.StatusFound = 302
	// config.GetLoginUrl()登入頁面的url
	gins.ctx.Redirect(http.StatusFound, config.Url(config.GetLoginUrl()))
	gins.ctx.Abort()
}

// SetContentType implements the method Adapter.SetContentType.
func (gins *Gin) SetContentType() {
	return
}

// Write implements the method Adapter.Write.
func (gins *Gin) Write(body []byte) {
	// Data方法在gin-gonic套件中
	// Data將資料寫入body並更新http代碼
	// gins.HTMLContentType() return "text/html; charset=utf-8"
	gins.ctx.Data(http.StatusOK, gins.HTMLContentType(), body)
}

// 取得cookie value藉由cookie命名尋找
// GetCookie implements the method Adapter.GetCookie.
func (gins *Gin) GetCookie() (string, error) {
	// Cookie()在gin-gonic套件裡Context(struct)的方法
	// Cookie()回傳cookie(藉由參數裡的命名回傳的)
	// gins.CookieKey()是利用Gin.adapter.BaseAdapter裡的CookieKey方法取得cookie的命名
	// gins.CookieKey() = go_admin_session
	return gins.ctx.Cookie(gins.CookieKey())
}

// 回傳路徑
// Path implements the method Adapter.Path.
func (gins *Gin) Path() string {
	return gins.ctx.Request.URL.Path
}

// 回傳方法
// Method implements the method Adapter.Method.
func (gins *Gin) Method() string {
	return gins.ctx.Request.Method
}

// 解析參數(multipart/form-data裡的)
// FormParam implements the method Adapter.FormParam.
func (gins *Gin) FormParam() url.Values {
	// http套件中
	// 解析multipart/form-data裡的參數
	_ = gins.ctx.Request.ParseMultipartForm(32 << 20)
	return gins.ctx.Request.PostForm
}

// 設置標頭 X-PJAX = true
// IsPjax implements the method Adapter.IsPjax.
func (gins *Gin) IsPjax() bool {
	// http套件中
	// constant.PjaxHeader = X-PJAX
	return gins.ctx.Request.Header.Get(constant.PjaxHeader) == "true"
}
