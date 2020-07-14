package table

import (
	"html/template"
	"sync"
	"sync/atomic"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/service"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/paginator"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/template/types"
)

type Generator func(ctx *context.Context) Table

type GeneratorList map[string]Generator

// 透過參數key及gen(function)添加至GeneratorList(map[string]Generator)
func (g GeneratorList) Add(key string, gen Generator) {
	g[key] = gen
}

// 透過參數list判斷GeneratorList已經有該key、value，如果不存在則加入該鍵與值
func (g GeneratorList) Combine(list GeneratorList) GeneratorList {
	for key, gen := range list {
		if _, ok := g[key]; !ok {
			g[key] = gen
		}
	}
	return g
}

// 透過參數gens判斷GeneratorList已經有該key、value，如果不存在則加入該鍵與值
func (g GeneratorList) CombineAll(gens []GeneratorList) GeneratorList {
	for _, list := range gens {
		for key, gen := range list {
			if _, ok := g[key]; !ok {
				g[key] = gen
			}
		}
	}
	return g
}

type Table interface {
	GetInfo() *types.InfoPanel
	GetDetail() *types.InfoPanel
	GetDetailFromInfo() *types.InfoPanel
	GetForm() *types.FormPanel

	GetCanAdd() bool
	GetEditable() bool
	GetDeletable() bool
	GetExportable() bool

	GetPrimaryKey() PrimaryKey

	GetData(params parameter.Parameters) (PanelInfo, error)
	GetDataWithIds(params parameter.Parameters) (PanelInfo, error)
	GetDataWithId(params parameter.Parameters) (FormInfo, error)
	UpdateData(dataList form.Values) error
	InsertData(dataList form.Values) error
	DeleteData(pk string) error

	GetNewForm() FormInfo

	GetOnlyInfo() bool
	GetOnlyDetail() bool
	GetOnlyNewForm() bool
	GetOnlyUpdateForm() bool

	Copy() Table
}

type BaseTable struct {
	Info           *types.InfoPanel
	Form           *types.FormPanel
	Detail         *types.InfoPanel
	CanAdd         bool
	Editable       bool
	Deletable      bool
	Exportable     bool
	OnlyInfo       bool
	OnlyDetail     bool
	OnlyNewForm    bool
	OnlyUpdateForm bool
	PrimaryKey     PrimaryKey
}

// 將參數值設置至base.Info(InfoPanel(struct)).primaryKey中後回傳
func (base *BaseTable) GetInfo() *types.InfoPanel {
	// 在template\types\info.go中
	// 將參數值設置至InfoPanel(struct).primaryKey中後回傳
	return base.Info.SetPrimaryKey(base.PrimaryKey.Name, base.PrimaryKey.Type)
}

func (base *BaseTable) GetDetail() *types.InfoPanel {
	return base.Detail.SetPrimaryKey(base.PrimaryKey.Name, base.PrimaryKey.Type)
}

func (base *BaseTable) GetDetailFromInfo() *types.InfoPanel {
	detail := base.GetDetail()
	detail.FieldList = make(types.FieldList, len(base.Info.FieldList))
	copy(detail.FieldList, base.Info.FieldList)
	return detail
}

// 將參數值(BaseTable.PrimaryKey)的值設置至BaseTable.Form(FormPanel(struct)).primaryKey中後回傳FormPanel(struct)
func (base *BaseTable) GetForm() *types.FormPanel {
	// 在template\types\info.go中
	// 將參數值設置至FormPanel(struct).primaryKey中後回傳FormPanel(struct)
	return base.Form.SetPrimaryKey(base.PrimaryKey.Name, base.PrimaryKey.Type)
}

func (base *BaseTable) GetCanAdd() bool {
	return base.CanAdd
}

func (base *BaseTable) GetPrimaryKey() PrimaryKey { return base.PrimaryKey } // 回傳BaseTable.PrimaryKey

func (base *BaseTable) GetEditable() bool { return base.Editable } // 回傳BaseTable.Editable(是否可以編輯)

func (base *BaseTable) GetDeletable() bool { return base.Deletable } // 回傳BaseTable.Deletable(是否可以刪除)

func (base *BaseTable) GetExportable() bool { return base.Exportable } // 回傳BaseTable.Exportable(是否可以輸出)

func (base *BaseTable) GetOnlyInfo() bool { return base.OnlyInfo } // 回傳BaseTable.OnlyInfo(是否唯一資訊)

func (base *BaseTable) GetOnlyDetail() bool { return base.OnlyDetail } // 回傳BaseTable.OnlyDetail(是否取得detail)

func (base *BaseTable) GetOnlyNewForm() bool { return base.OnlyNewForm } // 回傳BaseTable.OnlyNewForm()

func (base *BaseTable) GetOnlyUpdateForm() bool { return base.OnlyUpdateForm } // 回傳BaseTable.OnlyUpdateForm

func (base *BaseTable) GetPaginator(size int, params parameter.Parameters, extraHtml ...template.HTML) types.PaginatorAttribute {

	var eh template.HTML

	if len(extraHtml) > 0 {
		eh = extraHtml[0]
	}

	return paginator.Get(paginator.Config{
		Size:         size,
		Param:        params,
		PageSizeList: base.Info.GetPageSizeList(),
	}).SetExtraInfo(eh)
}

type PanelInfo struct {
	Thead          types.Thead              `json:"thead"`
	InfoList       types.InfoList           `json:"info_list"`
	FilterFormData types.FormFields         `json:"filter_form_data"`
	Paginator      types.PaginatorAttribute `json:"-"`
	Title          string                   `json:"title"`
	Description    string                   `json:"description"`
}

type FormInfo struct {
	FieldList         types.FormFields        `json:"field_list"`
	GroupFieldList    types.GroupFormFields   `json:"group_field_list"`
	GroupFieldHeaders types.GroupFieldHeaders `json:"group_field_headers"`
	Title             string                  `json:"title"`
	Description       string                  `json:"description"`
}

type PrimaryKey struct {
	Type db.DatabaseType
	Name string
}

const (
	DefaultPrimaryKeyName = "id"
	DefaultConnectionName = "default"
)

var (
	services service.List
	count    uint32
	lock     sync.Mutex
)

func SetServices(srv service.List) {
	lock.Lock()
	defer lock.Unlock()

	if atomic.LoadUint32(&count) != 0 {
		panic("can not initialize twice")
	}

	services = srv
}
