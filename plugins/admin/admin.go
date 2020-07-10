package admin

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/service"
	"github.com/GoAdminGroup/go-admin/plugins"
	"github.com/GoAdminGroup/go-admin/plugins/admin/controller"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/guard"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/action"
	_ "github.com/GoAdminGroup/go-admin/template/types/display"
)

// Admin is a GoAdmin plugin.
type Admin struct {
	*plugins.Base
	// plugins\admin\modules\table\table.go
	// GeneratorList類別為map[string]Generator，Generator類別為func(ctx *context.Context) Table
	tableList table.GeneratorList
	// plugins\admin\modules\guard
	guardian  *guard.Guard
	// plugins\admin\controller
	handler   *controller.Handler
}

// InitPlugin implements Plugin.InitPlugin.
// TODO: find a better way to manage the dependencies
func (admin *Admin) InitPlugin(services service.List) {

	// DO NOT DELETE
	// 將參數services(map[string]Service)設置至Admin.Base(struct)
	admin.InitBase(services)

	// 透過參數("config")取得匹配的Service(interface)
	// GetService將參數services.Get("config")轉換成Service(struct)後回傳Service.C(Config struct)
	c := config.GetService(services.Get("config"))

	// 將參數conn、c設置至SystemTable(struct)後回傳
	st := table.NewSystemTable(admin.Conn, c)

	// GeneratorList類別為map[string]Generator，Generator類別為func(ctx *context.Context) Table
	// Combine透過參數判斷GeneratorList已經有該key、value，如果不存在則加入該鍵與值
	admin.tableList.Combine(table.GeneratorList{
		"manager":        st.GetManagerTable,
		"permission":     st.GetPermissionTable,
		"roles":          st.GetRolesTable,
		"op":             st.GetOpTable,
		"menu":           st.GetMenuTable,
		"normal_manager": st.GetNormalManagerTable,
		"site":           st.GetSiteTable,
		"generate":       st.GetGenerateForm,
	})

	// 將參數admin.Services, admin.Conn, admin.tableList設置Admin.guardian(struct)後回傳
	admin.guardian = guard.New(admin.Services, admin.Conn, admin.tableList, admin.UI.NavButtons)

	// 將參數設置至Config(struct)
	handlerCfg := controller.Config{
		Config:     c,
		Services:   services,
		Generators: admin.tableList,
		Connection: admin.Conn,
	}

	// 將參數handlerCfg(struct)參數設置至Admin.handler(struct)
	admin.handler.UpdateCfg(handlerCfg)

	// 初始化router
	admin.initRouter()
	admin.handler.SetRoutes(admin.App.Routers)
	admin.handler.AddNavButton(admin.UI.NavButtons)

	table.SetServices(services)

	action.InitOperationHandlerSetter(admin.GetAddOperationFn())
}

// NewAdmin return the global Admin plugin.
// 設置Admin(struct)後回傳
func NewAdmin(tableCfg ...table.GeneratorList) *Admin {
	return &Admin{
		tableList: make(table.GeneratorList).CombineAll(tableCfg),
		Base:      &plugins.Base{PlugName: "admin"},
		// 設置Handler(struct)後回傳
		handler:   controller.New(),
	}
}

func (admin *Admin) GetAddOperationFn() context.NodeProcessor {
	return admin.handler.AddOperation
}

// SetCaptcha set captcha driver.
// 將參數captcha(驗證碼)設置至Admin.handler.captchaConfig(struct)
func (admin *Admin) SetCaptcha(captcha map[string]string) *Admin {
	// SetCaptcha在plugins\admin\controller\common.go
	// 將參數captcha設置至Handler.captchaConfig(驗證碼配置)
	admin.handler.SetCaptcha(captcha)
	return admin
}

// AddGenerator add table model generator.
// 將參數key及gen(function)添加數值至至GeneratorList(map[string]Generator)
func (admin *Admin) AddGenerator(key string, g table.Generator) *Admin {
	// Add將參數key及g(function)添加至Admin.tableList(map[string]Generator)
	admin.tableList.Add(key, g)
	return admin
}

// AddGenerators add table model generators.

func (admin *Admin) AddGenerators(gen ...table.GeneratorList) *Admin {
	admin.tableList.CombineAll(gen)
	return admin
}

// AddGlobalDisplayProcessFn call types.AddGlobalDisplayProcessFn
func (admin *Admin) AddGlobalDisplayProcessFn(f types.FieldFilterFn) *Admin {
	types.AddGlobalDisplayProcessFn(f)
	return admin
}

// AddDisplayFilterLimit call types.AddDisplayFilterLimit
func (admin *Admin) AddDisplayFilterLimit(limit int) *Admin {
	types.AddLimit(limit)
	return admin
}

// AddDisplayFilterTrimSpace call types.AddDisplayFilterTrimSpace
func (admin *Admin) AddDisplayFilterTrimSpace() *Admin {
	types.AddTrimSpace()
	return admin
}

// AddDisplayFilterSubstr call types.AddDisplayFilterSubstr
func (admin *Admin) AddDisplayFilterSubstr(start int, end int) *Admin {
	types.AddSubstr(start, end)
	return admin
}

// AddDisplayFilterToTitle call types.AddDisplayFilterToTitle
func (admin *Admin) AddDisplayFilterToTitle() *Admin {
	types.AddToTitle()
	return admin
}

// AddDisplayFilterToUpper call types.AddDisplayFilterToUpper
func (admin *Admin) AddDisplayFilterToUpper() *Admin {
	types.AddToUpper()
	return admin
}

// AddDisplayFilterToLower call types.AddDisplayFilterToLower
func (admin *Admin) AddDisplayFilterToLower() *Admin {
	types.AddToUpper()
	return admin
}

// AddDisplayFilterXssFilter call types.AddDisplayFilterXssFilter
func (admin *Admin) AddDisplayFilterXssFilter() *Admin {
	types.AddXssFilter()
	return admin
}

// AddDisplayFilterXssJsFilter call types.AddDisplayFilterXssJsFilter
func (admin *Admin) AddDisplayFilterXssJsFilter() *Admin {
	types.AddXssJsFilter()
	return admin
}
