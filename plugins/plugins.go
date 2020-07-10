// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package plugins

import (
	"bytes"
	"errors"
	template2 "html/template"
	"net/http"
	"plugin"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/menu"
	"github.com/GoAdminGroup/go-admin/modules/service"
	"github.com/GoAdminGroup/go-admin/modules/ui"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
)

// Plugin as one of the key components of goAdmin has three
// methods. GetRequest return all the path registered in the
// plugin. GetHandler according the url and method return the
// corresponding handler. InitPlugin init the plugin which do
// something like init the database and set the config and register
// the routes. The Plugin must implement the three methods.
// GetRequest回傳插件中的所有路徑
// InitPlugin初始化插件，類似於初始化資料庫並設置及配置路徑
type Plugin interface {
	GetHandler() context.HandlerMap
	InitPlugin(services service.List)
	Name() string
	Prefix() string
}

// Base(struct)也是Plugin(interface)
type Base struct {
	// context.App在context\context.go中
	App       *context.App
	Services  service.List
	Conn      db.Connection
	UI        *ui.Service
	PlugName  string
	URLPrefix string
}

// 回傳Base.App.Handlers
func (b *Base) GetHandler() context.HandlerMap {
	return b.App.Handlers
}

// 回傳Base.PlugName
func (b *Base) Name() string {
	return b.PlugName
}

// 回傳Base.URLPrefix
func (b *Base) Prefix() string {
	return b.URLPrefix
}

// 將參數srv(map[string]Service)設置至Base(struct)
func (b *Base) InitBase(srv service.List) {
	b.Services = srv
	// 將參數b.Services轉換為Connect(interface)回傳並回傳
	b.Conn = db.GetConnection(b.Services)
	// 將參數b.Services轉換成Service(struct)後回傳
	b.UI = ui.GetService(b.Services)
}

// 將參數設置至ExecuteParam(struct)，接著將給定的數據(types.Page(struct))寫入buf(struct)並回傳
func (b *Base) ExecuteTmpl(ctx *context.Context, panel types.Panel, animation ...bool) *bytes.Buffer {
	return Execute(ctx, b.Conn, *b.UI.NavButtons, auth.Auth(ctx), panel, animation...)
}

// 將參數設置至ExecuteParam(struct)，接著將給定的數據(types.Page(struct))寫入buf(struct)，並輸出HTML至Context.response.Body
func (b *Base) HTML(ctx *context.Context, panel types.Panel, animation ...bool) {
	// ExecuteTmpl將參數設置至ExecuteParam(struct)，接著將給定的數據(types.NewPageParam...)寫入buf(struct)並回傳
	buf := b.ExecuteTmpl(ctx, panel, animation...)
	// 在context/context.go
	// 輸出HTML，將body參數設置至Context.response.Body
	ctx.HTMLByte(http.StatusOK, buf.Bytes())
}

// 先解析檔案後將參數設置至ExecuteParam(struct)，接著將給定的數據(types.Page(struct))寫入buf(struct)，並輸出HTML至Context.response.Body
func (b *Base) HTMLFile(ctx *context.Context, path string, data map[string]interface{}, animation ...bool) {

	buf := new(bytes.Buffer)
	var panel types.Panel

	// html\template套件
	// ParseFiles解析檔案並創建一個Template(struct)
	t, err := template2.ParseFiles(path)
	if err != nil {
		// IsProductionEnvironment判斷globalCfg(Config).Env是否是"prod"
		panel = template.WarningPanel(err.Error()).GetContent(config.IsProductionEnvironment())
	} else {
		// Execute將參數data寫入參數buf中
		if err := t.Execute(buf, data); err != nil {
			panel = template.WarningPanel(err.Error()).GetContent(config.IsProductionEnvironment())
		} else {
			panel = types.Panel{
				// HTML轉換型態為string
				Content: template.HTML(buf.String()),
			}
		}
	}

	// 將參數設置至ExecuteParam(struct)，接著將給定的數據(types.Page(struct))寫入buf(struct)，並輸出HTML至Context.response.Body
	b.HTML(ctx, panel, animation...)
}

// 先解析多個檔案後將參數設置至ExecuteParam(struct)，接著將給定的數據(types.Page(struct))寫入buf(struct)，並輸出HTML至Context.response.Body
func (b *Base) HTMLFiles(ctx *context.Context, data map[string]interface{}, files []string, animation ...bool) {
	buf := new(bytes.Buffer)
	var panel types.Panel

	t, err := template2.ParseFiles(files...)
	if err != nil {
		panel = template.WarningPanel(err.Error()).GetContent(config.IsProductionEnvironment())
	} else {
		// Execute將參數data寫入參數buf中
		if err := t.Execute(buf, data); err != nil {
			panel = template.WarningPanel(err.Error()).GetContent(config.IsProductionEnvironment())
		} else {
			panel = types.Panel{
				Content: template.HTML(buf.String()),
				// HTML轉換型態為string(buf.String())
			}
		}
	}

	// 將參數設置至ExecuteParam(struct)，接著將給定的數據(types.Page(struct))寫入buf(struct)，並輸出HTML至Context.response.Body
	b.HTML(ctx, panel, animation...)
}

// 將參數mod打開取得Plugin(struct)後尋找"Plugin"的符號，最後轉換成Plugin(interface)類別回傳
func LoadFromPlugin(mod string) Plugin {

	// plugin套件
	// Open打開go plugin，如果參數mod已經存在則回傳Plugin(struct)
	plug, err := plugin.Open(mod)
	if err != nil {
		logger.Error("LoadFromPlugin err", err)
		panic(err)
	}

	// 尋找plug(struct)中參數"Plugin"的符號，回傳symPlugin(interface)
	symPlugin, err := plug.Lookup("Plugin")
	if err != nil {
		logger.Error("LoadFromPlugin err", err)
		panic(err)
	}

	var p Plugin
	//將symPlugin轉換成Plugin(interface)
	p, ok := symPlugin.(Plugin)
	if !ok {
		logger.Error("LoadFromPlugin err: unexpected type from module symbol")
		panic(errors.New("LoadFromPlugin err: unexpected type from module symbol"))
	}

	return p
}

// GetHandler is a help method for Plugin GetHandler.
// 回傳參數app(App.Handlers)
func GetHandler(app *context.App) context.HandlerMap { return app.Handlers }

// 將參數設置至ExecuteParam(struct)，接著將給定的數據(types.Page(struct))寫入buf(struct)並回傳
func Execute(ctx *context.Context, conn db.Connection, navButtons types.Buttons, user models.UserModel,
	panel types.Panel, animation ...bool) *bytes.Buffer {
	// GetTheme回傳globalCfg.Theme
	// IsPjax判斷是否header X-PJAX:true
	// Get判斷templateMap(map[string]Template)的key鍵是否參數config.GetTheme()，有則回傳Template(interface)
	// GetTemplate為Template(interface)的方法
	tmpl, tmplName := template.Get(config.GetTheme()).GetTemplate(ctx.IsPjax())

	// template\template.go中
	return template.Execute(template.ExecuteParam{
		User:       user,
		TmplName:   tmplName,
		Tmpl:       tmpl,
		Panel:      panel,
		// 複製globalCfg(Config struct)後將Config.Databases[key].Driver設置至Config.Databases[key]後回傳
		Config:     *config.Get(),
		// GetGlobalMenu回傳參數user(struct)的Menu(設置menuList、menuOption、MaxOrder)
		// SetActiveClass設定menu的active
		// URLRemovePrefixglobalCfg(Config struct).prefix將URL的前綴去除
		Menu:       menu.GetGlobalMenu(user, conn).SetActiveClass(config.URLRemovePrefix(ctx.Path())),
		Animation:  len(animation) > 0 && animation[0] || len(animation) == 0,
		Buttons:    navButtons.CheckPermission(user),
		NoCompress: len(animation) > 1 && animation[1],
	})
}
