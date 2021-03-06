package types

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/modules/utils"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	"github.com/GoAdminGroup/go-admin/template/types/table"
)

// FieldModel is the single query result.
type FieldModel struct {
	// The primaryKey of the table.
	ID string

	// The value of the single query result.
	Value string

	// The current row data.
	Row map[string]interface{}

	// Post type
	PostType PostType
}

type PostType uint8

const (
	PostTypeCreate = iota
	PostTypeUpdate
)

func (m FieldModel) IsCreate() bool {
	return m.PostType == PostTypeCreate
}

func (m FieldModel) IsUpdate() bool {
	return m.PostType == PostTypeUpdate
}

// PostFieldModel contains ID and value of the single query result and the current row data.
type PostFieldModel struct {
	ID    string
	Value FieldModelValue
	Row   map[string]string
	// Post type
	PostType PostType
}

func (m PostFieldModel) IsCreate() bool {
	return m.PostType == PostTypeCreate
}

func (m PostFieldModel) IsUpdate() bool {
	return m.PostType == PostTypeUpdate
}

type InfoList []map[string]InfoItem

type InfoItem struct {
	Content template.HTML `json:"content"`
	Value   string        `json:"value"`
}

func (i InfoList) GroupBy(groups TabGroups) []InfoList {

	var res = make([]InfoList, len(groups))

	for key, value := range groups {
		var newInfoList = make(InfoList, len(i))

		for index, info := range i {
			var newRow = make(map[string]InfoItem)
			for mk, m := range info {
				if modules.InArray(value, mk) {
					newRow[mk] = m
				}
			}
			newInfoList[index] = newRow
		}

		res[key] = newInfoList
	}

	return res
}

type Callbacks []context.Node

func (c Callbacks) AddCallback(node context.Node) Callbacks {
	if node.Path != "" && node.Method != "" && len(node.Handlers) > 0 {
		for _, item := range c {
			if strings.ToUpper(item.Path) == strings.ToUpper(node.Path) &&
				strings.ToUpper(item.Method) == strings.ToUpper(node.Method) {
				return c
			}
		}
		parr := strings.Split(node.Path, "?")
		if len(parr) > 1 {
			node.Path = parr[0]
			return append(c, node)
		}
		return append(c, node)
	}
	return c
}

type FieldModelValue []string

func (r FieldModelValue) Value() string {
	return r.First()
}

func (r FieldModelValue) First() string {
	return r[0]
}

// FieldDisplay is filter function of data.
type FieldFilterFn func(value FieldModel) interface{}

// PostFieldFilterFn is filter function of data.
type PostFieldFilterFn func(value PostFieldModel) interface{}

// Field is the table field.
type Field struct {
	Head     string
	Field    string
	TypeName db.DatabaseType

	Joins Joins

	Width      int
	Sortable   bool
	EditAble   bool
	Fixed      bool
	Filterable bool
	Hide       bool

	EditType    table.Type
	EditOptions FieldOptions

	FilterFormFields []FilterFormField

	FieldDisplay
}

type QueryFilterFn func(param parameter.Parameters, conn db.Connection) (ids []string, stopQuery bool)

type FilterFormField struct {
	Type        form.Type
	Options     FieldOptions
	OptionTable OptionTable
	Width       int
	Operator    FilterOperator
	OptionExt   template.JS
	Head        string
	Placeholder string
	HelpMsg     template.HTML
	ProcessFn   func(string) string
}

// 將Field(struct)透過參數Parameters(struct)及string處理後回傳[]FormField
func (f Field) GetFilterFormFields(params parameter.Parameters, headField string, sql ...*db.SQL) []FormField {

	var (
		filterForm               = make([]FormField, 0)
		value, value2, keySuffix string
	)

	// 處理可以篩選條件的欄位
	for index, filter := range f.FilterFormFields {
		if index > 0 {
			keySuffix = parameter.FilterParamCountInfix + strconv.Itoa(index)
		}

		if filter.Type.IsRange() {
			value = params.GetFilterFieldValueStart(headField)
			value2 = params.GetFilterFieldValueEnd(headField)
		} else if filter.Type.IsMultiSelect() {
			value = params.GetFieldValuesStr(headField)
		} else {
			if filter.Operator == FilterOperatorFree {
				value2 = GetOperatorFromValue(params.GetFieldOperator(headField, keySuffix)).String()
			}
			// -------用戶頁面會執行(使用者、暱稱、角色欄位)，value回傳空值--------------
			// GetFieldValue透過參數field尋找Parameters.Fields[field]是否存在，如果存在則回傳第一個value值(string)
			value = params.GetFieldValue(headField + keySuffix)
		}

		var (
			optionExt1 = filter.OptionExt
			optionExt2 template.JS
		)

		if filter.OptionExt == template.JS("") {
			// ------------用戶頁面的三個篩選欄位會執行--------------
			op1, op2, js := filter.Type.GetDefaultOptions(headField + keySuffix)
			if op1 != nil {
				s, _ := json.Marshal(op1)
				optionExt1 = template.JS(string(s))
			}
			if op2 != nil {
				s, _ := json.Marshal(op2)
				optionExt2 = template.JS(string(s))
			}
			if js != template.JS("") {
				optionExt1 = js
			}
		}

		field := &FormField{
			Field:       headField + keySuffix,
			FieldClass:  headField + keySuffix,
			Head:        filter.Head,
			TypeName:    f.TypeName,
			HelpMsg:     filter.HelpMsg,
			FormType:    filter.Type,
			Editable:    true,
			Width:       filter.Width,
			Placeholder: filter.Placeholder,
			Value:       template.HTML(value),
			Value2:      value2,
			Options:     filter.Options,
			OptionExt:   optionExt1,
			OptionExt2:  optionExt2,
			OptionTable: filter.OptionTable,
			Label:       filter.Operator.Label(),
		}

		field.setOptionsFromSQL(sql[0])

		if filter.Type.IsSingleSelect() {
			field.Options = field.Options.SetSelected(params.GetFieldValue(f.Field), filter.Type.SelectedLabel())
		}

		if filter.Type.IsMultiSelect() {
			field.Options = field.Options.SetSelected(params.GetFieldValues(f.Field), filter.Type.SelectedLabel())
		}

		filterForm = append(filterForm, *field)

		if filter.Operator.AddOrNot() {
			filterForm = append(filterForm, FormField{
				Field:      headField + parameter.FilterParamOperatorSuffix + keySuffix,
				FieldClass: headField + parameter.FilterParamOperatorSuffix + keySuffix,
				Head:       f.Head,
				TypeName:   f.TypeName,
				Value:      template.HTML(filter.Operator.Value()),
				FormType:   filter.Type,
				Hide:       true,
			})
		}
	}

	return filterForm
}

func (f Field) Exist() bool {
	return f.Field != ""
}

type FieldList []Field

type TableInfo struct {
	Table      string
	PrimaryKey string
	Delimiter  string
	Driver     string
}

// 透過參數並且將欄位、join語法...等資訊處理後，回傳[]TheadItem、欄位名稱、joinFields(ex:group_concat(goadmin_roles.`name`...)、join語法(left join....)、合併的資料表、可篩選過濾的欄位
func (f FieldList) GetTheadAndFilterForm(info TableInfo, params parameter.Parameters, columns []string,
	sql ...func() *db.SQL) (Thead, string, string, string, []string, []FormField) {
	var (
		thead      = make(Thead, 0)
		fields     = ""                   // 欄位
		joinFields = ""                   // ex: group_concat(goadmin_roles.`name` separator 'CkN694kH') as goadmin_roles_goadmin_join_name,
		joins      = ""                   // join資料表語法，ex: left join `goadmin_role_users` on goadmin_role_users.`user_id` = goadmin_users.`id` left join....
		joinTables = make([]string, 0)    // ex:{goadmin_roles role_id id goadmin_role_users}(用戶頁面)
		filterForm = make([]FormField, 0) // 可以篩選過濾的欄位
	)

	// field為介面顯示的欄位
	for _, field := range f {

		if field.Field != info.PrimaryKey && modules.InArray(columns, field.Field) &&
			// Valid對joins([]join(struct))執行迴圈，假設Join的Table、Field、JoinField都不為空，回傳true
			!field.Joins.Valid() {
			// 欄位在columns裡以及不是primary key會執行，在欄位名稱前加入資料表名(ex: tablename.colname)
			// ex: goadmin_users.`username`,goadmin_users.`name`,goadmin_users.`created_at`,goadmin_users.`updated_at`,
			fields += info.Table + "." + modules.FilterField(field.Field, info.Delimiter) + ","
		}

		headField := field.Field

		// -------------編輯介面(用戶的roles欄位會執行)-------------
		// 處理join語法
		// 例如用戶頁面的role欄位需要與其他表join取得角色，因此Joins([]Join)為ex: [{goadmin_role_users id user_id } {goadmin_roles role_id id goadmin_role_users}]
		// Valid對joins([]join(struct))執行迴圈，假設Join的Table、Field、JoinField都不為空，回傳true
		if field.Joins.Valid() {
			// Last判斷Joins([]Join)長度，如果大於0回傳Joins[len(j)-1](最後一個數值)(struct)
			// FilterParamJoinInfix = _goadmin_join_
			// ex:goadmin_roles_goadmin_join_name
			headField = field.Joins.Last().Table + parameter.FilterParamJoinInfix + field.Field

			// GetAggregationExpression取得資料庫引擎的Aggregation表達式，將參數值加入表達式
			// FilterField判斷第二個參數符號，如果為[則回傳[field(第一個參數)]，否則回傳ex: 'field'(mysql)
			// ex: group_concat(goadmin_roles.`name` separator 'CkN694kH') as goadmin_roles_goadmin_join_name,
			joinFields += db.GetAggregationExpression(info.Driver, field.Joins.Last().Table+"."+
				modules.FilterField(field.Field, info.Delimiter), headField, JoinFieldValueDelimiter) + ","

			for _, join := range field.Joins {
				if !modules.InArray(joinTables, join.Table) {
					joinTables = append(joinTables, join.Table)
					if join.BaseTable == "" {
						// ex: goadmin_users(用戶頁面)
						join.BaseTable = info.Table
					}
					// FilterField判斷第二個參數符號，如果為[則回傳[field(第一個參數)]，否則回傳ex: 'field'(mysql)
					// ex: joins =  left join `goadmin_role_users` on goadmin_role_users.`user_id` = goadmin_users.`id` left join....
					joins += " left join " + modules.FilterField(join.Table, info.Delimiter) + " on " +
						join.Table + "." + modules.FilterField(join.JoinField, info.Delimiter) + " = " +
						join.BaseTable + "." + modules.FilterField(join.Field, info.Delimiter)

				}
			}
		}

		// 可以做篩選的欄位，例如用戶頁面的用戶名、暱稱、角色
		if field.Filterable {
			if len(sql) > 0 {
				// GetFilterFormFields透過參數Parameters(struct)及string處理後回傳[]FormField
				filterForm = append(filterForm, field.GetFilterFormFields(params, headField, sql[0]())...)
			} else {
				filterForm = append(filterForm, field.GetFilterFormFields(params, headField)...)
			}
		}

		// 檢查欄位是否隱藏
		if field.Hide {
			continue
		}

		// 將值設置至TheadItem(struct)並添加至thead([]TheadItem)
		thead = append(thead, TheadItem{
			Head:     field.Head,
			Sortable: field.Sortable, // 是否可以排序
			Field:    headField,
			// 判斷params.Columns([]string)長度如果為0回傳true，如果值與第二個參數(string)相等也回傳true，否則回傳false
			// params.Columns為顯示的欄位
			Hide:       !modules.InArrayWithoutEmpty(params.Columns, headField),
			Editable:   field.EditAble,
			EditType:   field.EditType.String(),
			EditOption: field.EditOptions,
			Width:      strconv.Itoa(field.Width) + "px",
		})

	}

	return thead, fields, joinFields, joins, joinTables, filterForm
}

func (f FieldList) GetThead(info TableInfo, params parameter.Parameters, columns []string) (Thead, string, string) {
	var (
		thead      = make(Thead, 0)
		fields     = ""
		joins      = ""
		joinTables = make([]string, 0)
	)
	for _, field := range f {
		if field.Field != info.PrimaryKey && modules.InArray(columns, field.Field) &&
			!field.Joins.Valid() {
			fields += info.Table + "." + modules.FilterField(field.Field, info.Delimiter) + ","
		}

		headField := field.Field

		if field.Joins.Valid() {
			headField = field.Joins.Last().Table + parameter.FilterParamJoinInfix + field.Field
			for _, join := range field.Joins {
				if !modules.InArray(joinTables, join.Table) {
					joinTables = append(joinTables, join.Table)
					if join.BaseTable == "" {
						join.BaseTable = info.Table
					}
					joins += " left join " + modules.FilterField(join.Table, info.Delimiter) + " on " +
						join.Table + "." + modules.FilterField(join.JoinField, info.Delimiter) + " = " +
						join.BaseTable + "." + modules.FilterField(join.Field, info.Delimiter)
				}
			}
		}

		if field.Hide {
			continue
		}
		thead = append(thead, TheadItem{
			Head:       field.Head,
			Sortable:   field.Sortable,
			Field:      headField,
			Hide:       !modules.InArrayWithoutEmpty(params.Columns, headField),
			Editable:   field.EditAble,
			EditType:   field.EditType.String(),
			EditOption: field.EditOptions,
			Width:      strconv.Itoa(field.Width) + "px",
		})
	}

	return thead, fields, joins
}

func (f FieldList) GetFieldFilterProcessValue(key, value, keyIndex string) string {

	field := f.GetFieldByFieldName(key)
	index := 0
	if keyIndex != "" {
		index, _ = strconv.Atoi(keyIndex)
	}
	if field.FilterFormFields[index].ProcessFn != nil {
		value = field.FilterFormFields[index].ProcessFn(value)
	}
	return value
}

func (f FieldList) GetFieldJoinTable(key string) string {
	field := f.GetFieldByFieldName(key)
	if field.Exist() {
		return field.Joins.Last().Table
	}
	return ""
}

func (f FieldList) GetFieldByFieldName(name string) Field {
	for _, field := range f {
		if field.Field == name {
			return field
		}
		if JoinField(field.Joins.Last().Table, field.Field) == name {
			return field
		}
	}
	return Field{}
}

// Join store join table info. For example:
//
// Join {
//     BaseTable:   "users",
//     Field:       "role_id",
//     Table:       "roles",
//     JoinField:   "id",
// }
//
// It will generate the join table sql like:
//
// ... left join roles on roles.id = users.role_id ...
//
type Join struct {
	Table     string
	Field     string
	JoinField string
	BaseTable string
}

type Joins []Join

func JoinField(table, field string) string {
	return table + parameter.FilterParamJoinInfix + field
}

func GetJoinField(field string) string {
	return strings.Split(field, parameter.FilterParamJoinInfix)[1]
}

// 對joins([]join(struct))執行迴圈，假設Join的Table、Field、JoinField都不為空，回傳true
func (j Joins) Valid() bool {
	for i := 0; i < len(j); i++ {
		// 假設Join的Table、Field、JoinField都不為空，回傳true
		if j[i].Valid() {
			return true
		}
	}
	return false
}

// 判斷Joins([]Join)長度，如果大於0回傳Joins[len(j)-1](最後一個數值)
func (j Joins) Last() Join {
	if len(j) > 0 {
		return j[len(j)-1]
	}
	return Join{}
}

// 假設Join的Table、Field、JoinField都不為空，回傳true
func (j Join) Valid() bool {
	return j.Table != "" && j.Field != "" && j.JoinField != ""
}

var JoinFieldValueDelimiter = utils.Uuid(8)

type TabGroups [][]string

// 判斷TabGroups([][]string)是否長度大於0
func (t TabGroups) Valid() bool {
	return len(t) > 0
}

func NewTabGroups(items ...string) TabGroups {
	var t = make(TabGroups, 0)
	return append(t, items)
}

func (t TabGroups) AddGroup(items ...string) TabGroups {
	return append(t, items)
}

type TabHeaders []string

func (t TabHeaders) Add(header string) TabHeaders {
	return append(t, header)
}

type GetDataFn func(param parameter.Parameters) ([]map[string]interface{}, int)

type DeleteFn func(ids []string) error
type DeleteFnWithRes func(ids []string, res error) error

type Sort uint8

const (
	SortDesc Sort = iota
	SortAsc
)

type primaryKey struct {
	Type db.DatabaseType
	Name string
}

// InfoPanel
type InfoPanel struct {
	FieldList         FieldList
	curFieldListIndex int

	Table       string
	Title       string
	Description string

	// Warn: may be deprecated future.
	TabGroups  TabGroups
	TabHeaders TabHeaders

	Sort      Sort
	SortField string

	PageSizeList    []int
	DefaultPageSize int

	ExportType int

	primaryKey primaryKey

	IsHideNewButton    bool
	IsHideExportButton bool
	IsHideEditButton   bool
	IsHideDeleteButton bool
	IsHideDetailButton bool
	IsHideFilterButton bool
	IsHideRowSelector  bool
	IsHidePagination   bool
	IsHideFilterArea   bool
	IsHideQueryInfo    bool
	FilterFormLayout   form.Layout

	FilterFormHeadWidth  int
	FilterFormInputWidth int

	Wheres    Wheres
	WhereRaws WhereRaw

	Callbacks Callbacks

	Buttons Buttons

	TableLayout string

	DeleteHook  DeleteFn
	PreDeleteFn DeleteFn
	DeleteFn    DeleteFn

	DeleteHookWithRes DeleteFnWithRes

	GetDataFn GetDataFn

	processChains DisplayProcessFnChains

	ActionButtons Buttons

	DisplayGeneratorRecords map[string]struct{}

	QueryFilterFn QueryFilterFn

	Wrapper ContentWrapper

	// column operation buttons
	Action     template.HTML
	HeaderHtml template.HTML
	FooterHtml template.HTML
}

type Where struct {
	Join     string
	Field    string
	Operator string
	Arg      interface{}
}

type Wheres []Where

func (whs Wheres) Statement(wheres, delimiter string, whereArgs []interface{}, existKeys, columns []string) (string, []interface{}) {
	pwheres := ""

	for k, wh := range whs {

		whFieldArr := strings.Split(wh.Field, ".")
		whField := ""
		whTable := ""
		if len(whFieldArr) > 1 {
			whField = whFieldArr[1]
			whTable = whFieldArr[0]
		} else {
			whField = whFieldArr[0]
		}

		if modules.InArray(existKeys, whField) {
			continue
		}

		// TODO: support like operation and join table
		if modules.InArray(columns, whField) {

			joinMark := ""
			if k != len(whs)-1 {
				joinMark = whs[k+1].Join
			}

			if whTable != "" {
				pwheres += whTable + "." + modules.FilterField(whField, delimiter) + " " + wh.Operator + " ? " + joinMark + " "
			} else {
				pwheres += modules.FilterField(whField, delimiter) + " " + wh.Operator + " ? " + joinMark + " "
			}
			whereArgs = append(whereArgs, wh.Arg)
		}
	}
	if wheres != "" && pwheres != "" {
		wheres += " and "
	}
	return wheres + pwheres, whereArgs
}

type WhereRaw struct {
	Raw  string
	Args []interface{}
}

func (wh WhereRaw) check() int {
	index := 0
	for i := 0; i < len(wh.Raw); i++ {
		if wh.Raw[i] == ' ' {
			continue
		} else {
			if wh.Raw[i] == 'a' {
				if len(wh.Raw) < i+3 {
					break
				} else {
					if wh.Raw[i+1] == 'n' && wh.Raw[i+2] == 'd' {
						index = i + 3
					}
				}
			} else if wh.Raw[i] == 'o' {
				if len(wh.Raw) < i+2 {
					break
				} else {
					if wh.Raw[i+1] == 'r' {
						index = i + 2
					}
				}
			} else {
				break
			}
		}
	}
	return index
}

func (wh WhereRaw) Statement(wheres string, whereArgs []interface{}) (string, []interface{}) {

	if wh.Raw == "" {
		return wheres, whereArgs
	}

	if wheres != "" {
		if wh.check() != 0 {
			wheres += wh.Raw + " "
		} else {
			wheres += " and " + wh.Raw + " "
		}

		whereArgs = append(whereArgs, wh.Args...)
	} else {
		wheres += wh.Raw[wh.check():] + " "
		whereArgs = append(whereArgs, wh.Args...)
	}

	return wheres, whereArgs
}

type Handler func(ctx *context.Context) (success bool, msg string, data interface{})

func (h Handler) Wrap() context.Handler {
	return func(ctx *context.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(err)
				ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
					"code": 500,
					"data": "",
					"msg":  "error",
				})
			}
		}()

		code := 0
		s, m, d := h(ctx)

		if !s {
			code = 500
		}
		ctx.JSON(http.StatusOK, map[string]interface{}{
			"code": code,
			"data": d,
			"msg":  m,
		})
	}
}

type ContentWrapper func(content template.HTML) template.HTML

type Action interface {
	Js() template.JS
	BtnAttribute() template.HTML
	BtnClass() template.HTML
	ExtContent() template.HTML
	FooterContent() template.HTML
	SetBtnId(btnId string)
	SetBtnData(data interface{})
	GetCallbacks() context.Node
}

type DefaultAction struct {
	Attr   template.HTML
	JS     template.JS
	Ext    template.HTML
	Footer template.HTML
}

func NewDefaultAction(attr, ext, footer template.HTML, js template.JS) *DefaultAction {
	return &DefaultAction{Attr: attr, Ext: ext, Footer: footer, JS: js}
}

func (def *DefaultAction) SetBtnId(btnId string)        {}
func (def *DefaultAction) SetBtnData(data interface{})  {}
func (def *DefaultAction) Js() template.JS              { return def.JS }
func (def *DefaultAction) BtnAttribute() template.HTML  { return def.Attr }
func (def *DefaultAction) BtnClass() template.HTML      { return "" }
func (def *DefaultAction) ExtContent() template.HTML    { return def.Ext }
func (def *DefaultAction) FooterContent() template.HTML { return def.Footer }
func (def *DefaultAction) GetCallbacks() context.Node   { return context.Node{} }

var _ Action = (*DefaultAction)(nil)

var DefaultPageSizeList = []int{10, 20, 30, 50, 100}

const DefaultPageSize = 10

func NewInfoPanel(pk string) *InfoPanel {
	return &InfoPanel{
		curFieldListIndex:       -1,
		PageSizeList:            DefaultPageSizeList,
		DefaultPageSize:         DefaultPageSize,
		processChains:           make(DisplayProcessFnChains, 0),
		Buttons:                 make(Buttons, 0),
		Callbacks:               make(Callbacks, 0),
		DisplayGeneratorRecords: make(map[string]struct{}),
		Wheres:                  make([]Where, 0),
		WhereRaws:               WhereRaw{},
		SortField:               pk,
		TableLayout:             "auto",
		FilterFormInputWidth:    10,
		FilterFormHeadWidth:     2,
	}
}

func (i *InfoPanel) Where(field string, operator string, arg interface{}) *InfoPanel {
	i.Wheres = append(i.Wheres, Where{Field: field, Operator: operator, Arg: arg, Join: "and"})
	return i
}

func (i *InfoPanel) WhereOr(field string, operator string, arg interface{}) *InfoPanel {
	i.Wheres = append(i.Wheres, Where{Field: field, Operator: operator, Arg: arg, Join: "or"})
	return i
}

func (i *InfoPanel) WhereRaw(raw string, arg ...interface{}) *InfoPanel {
	i.WhereRaws.Raw = raw
	i.WhereRaws.Args = arg
	return i
}

func (i *InfoPanel) AddSelectBox(placeholder string, options FieldOptions, action Action, width ...int) *InfoPanel {
	options = append(FieldOptions{{Value: "", Text: language.Get("All")}}, options...)
	action.SetBtnData(options)
	i.addButton(GetDefaultSelection(placeholder, options, action, width...)).
		addFooterHTML(action.FooterContent()).
		addCallback(action.GetCallbacks())

	return i
}

func (i *InfoPanel) ExportValue() *InfoPanel {
	i.ExportType = 1
	return i
}

// 是否輸出值
func (i *InfoPanel) IsExportValue() bool {
	return i.ExportType == 1
}

func (i *InfoPanel) AddButtonRaw(btn Button, action Action) *InfoPanel {
	i.Buttons = append(i.Buttons, btn)
	i.addFooterHTML(action.FooterContent()).addCallback(action.GetCallbacks())
	return i
}

func (i *InfoPanel) AddButton(title template.HTML, icon string, action Action, color ...template.HTML) *InfoPanel {
	i.addButton(GetDefaultButton(title, icon, action, color...)).
		addFooterHTML(action.FooterContent()).
		addCallback(action.GetCallbacks())
	return i
}

func (i *InfoPanel) AddActionButton(title template.HTML, action Action, ids ...string) *InfoPanel {
	i.addActionButton(GetActionButton(title, action, ids...)).
		addFooterHTML(action.FooterContent()).
		addCallback(action.GetCallbacks())

	return i
}

func (i *InfoPanel) AddActionButtonFront(title template.HTML, action Action, ids ...string) *InfoPanel {
	i.ActionButtons = append([]Button{GetActionButton(title, action, ids...)}, i.ActionButtons...)
	i.addFooterHTML(action.FooterContent()).
		addCallback(action.GetCallbacks())
	return i
}

func (i *InfoPanel) AddLimitFilter(limit int) *InfoPanel {
	i.processChains = addLimit(limit, i.processChains)
	return i
}

func (i *InfoPanel) AddTrimSpaceFilter() *InfoPanel {
	i.processChains = addTrimSpace(i.processChains)
	return i
}

func (i *InfoPanel) AddSubstrFilter(start int, end int) *InfoPanel {
	i.processChains = addSubstr(start, end, i.processChains)
	return i
}

func (i *InfoPanel) AddToTitleFilter() *InfoPanel {
	i.processChains = addToTitle(i.processChains)
	return i
}

func (i *InfoPanel) AddToUpperFilter() *InfoPanel {
	i.processChains = addToUpper(i.processChains)
	return i
}

func (i *InfoPanel) AddToLowerFilter() *InfoPanel {
	i.processChains = addToLower(i.processChains)
	return i
}

func (i *InfoPanel) AddXssFilter() *InfoPanel {
	i.processChains = addXssFilter(i.processChains)
	return i
}

func (i *InfoPanel) AddXssJsFilter() *InfoPanel {
	i.processChains = addXssJsFilter(i.processChains)
	return i
}

// 將參數設置至InfoPanel(struct).DeleteHook中並回傳
func (i *InfoPanel) SetDeleteHook(fn DeleteFn) *InfoPanel {
	i.DeleteHook = fn
	return i
}

// 將參數設置至InfoPanel(struct).DeleteHookWithRes中並回傳
func (i *InfoPanel) SetDeleteHookWithRes(fn DeleteFnWithRes) *InfoPanel {
	i.DeleteHookWithRes = fn
	return i
}

// 將參數設置至InfoPanel(struct).QueryFilterFn中並回傳
func (i *InfoPanel) SetQueryFilterFn(fn QueryFilterFn) *InfoPanel {
	i.QueryFilterFn = fn
	return i
}

// 將參數設置至InfoPanel(struct).Wrapper中並回傳
func (i *InfoPanel) SetWrapper(wrapper ContentWrapper) *InfoPanel {
	i.Wrapper = wrapper
	return i
}

// 將參數設置至InfoPanel(struct).PreDeleteFn中並回傳
func (i *InfoPanel) SetPreDeleteFn(fn DeleteFn) *InfoPanel {
	i.PreDeleteFn = fn
	return i
}

// 將參數設置至InfoPanel(struct).DeleteFn中並回傳
func (i *InfoPanel) SetDeleteFn(fn DeleteFn) *InfoPanel {
	i.DeleteFn = fn
	return i
}

// 將參數設置至InfoPanel(struct).GetDataFn中並回傳
func (i *InfoPanel) SetGetDataFn(fn GetDataFn) *InfoPanel {
	i.GetDataFn = fn
	return i
}

// 將參數設置至InfoPanel(struct).primaryKey中並回傳
func (i *InfoPanel) SetPrimaryKey(name string, typ db.DatabaseType) *InfoPanel {
	i.primaryKey = primaryKey{Name: name, Type: typ}
	return i
}

func (i *InfoPanel) SetTableFixed() *InfoPanel {
	i.TableLayout = "fixed"
	return i
}

func (i *InfoPanel) AddColumn(head string, fun FieldFilterFn) *InfoPanel {
	i.FieldList = append(i.FieldList, Field{
		Head:     head,
		Field:    head,
		TypeName: db.Varchar,
		Sortable: false,
		EditAble: false,
		EditType: table.Text,
		FieldDisplay: FieldDisplay{
			Display:              fun,
			DisplayProcessChains: chooseDisplayProcessChains(i.processChains),
		},
	})
	i.curFieldListIndex++
	return i
}

func (i *InfoPanel) AddColumnButtons(head string, buttons ...Button) *InfoPanel {
	var content, js template.HTML
	for _, btn := range buttons {
		btn.GetAction().SetBtnId(btn.ID())
		btnContent, btnJs := btn.Content()
		content += btnContent
		js += template.HTML(btnJs)
		i.FooterHtml += template.HTML(ParseTableDataTmpl(btn.GetAction().FooterContent()))
		i.Callbacks = i.Callbacks.AddCallback(btn.GetAction().GetCallbacks())
	}
	i.FooterHtml += template.HTML("<script>") + template.HTML(ParseTableDataTmpl(js)) + template.HTML("</script>")
	i.FieldList = append(i.FieldList, Field{
		Head:     head,
		Field:    head,
		TypeName: db.Varchar,
		Sortable: false,
		EditAble: false,
		EditType: table.Text,
		FieldDisplay: FieldDisplay{
			Display: func(value FieldModel) interface{} {
				pk := db.GetValueFromDatabaseType(i.primaryKey.Type, value.Row[i.primaryKey.Name], i.isFromJSON())
				var v = make(map[string]InfoItem)
				for key, item := range value.Row {
					itemValue := fmt.Sprintf("%v", item)
					v[key] = InfoItem{Value: itemValue, Content: template.HTML(itemValue)}
				}
				return template.HTML(ParseTableDataTmplWithID(pk.HTML(), string(content), v))
			},
			DisplayProcessChains: chooseDisplayProcessChains(i.processChains),
		},
	})
	i.curFieldListIndex++
	return i
}

func (i *InfoPanel) AddField(head, field string, typeName db.DatabaseType) *InfoPanel {
	i.FieldList = append(i.FieldList, Field{
		Head:     head,
		Field:    field,
		TypeName: typeName,
		Sortable: false,
		Joins:    make(Joins, 0),
		EditAble: false,
		EditType: table.Text,
		FieldDisplay: FieldDisplay{
			Display: func(value FieldModel) interface{} {
				return value.Value
			},
			DisplayProcessChains: chooseDisplayProcessChains(i.processChains),
		},
	})
	i.curFieldListIndex++
	return i
}

// Field attribute setting functions
// ====================================================

func (i *InfoPanel) FieldDisplay(filter FieldFilterFn) *InfoPanel {
	i.FieldList[i.curFieldListIndex].Display = filter
	return i
}

type FieldLabelParam struct {
	Color template.HTML
	Type  string
}

func (i *InfoPanel) FieldLabel(args ...FieldLabelParam) *InfoPanel {
	i.addDisplayChains(displayFnGens["label"].Get(args))
	return i
}

func (i *InfoPanel) FieldImage(width, height string, prefix ...string) *InfoPanel {
	i.addDisplayChains(displayFnGens["image"].Get(width, height, prefix))
	return i
}

func (i *InfoPanel) FieldBool(flags ...string) *InfoPanel {
	i.addDisplayChains(displayFnGens["bool"].Get(flags))
	return i
}

func (i *InfoPanel) FieldLink(src string, openInNewTab ...bool) *InfoPanel {
	i.addDisplayChains(displayFnGens["link"].Get(src, openInNewTab))
	return i
}

func (i *InfoPanel) FieldFileSize() *InfoPanel {
	i.addDisplayChains(displayFnGens["filesize"].Get())
	return i
}

func (i *InfoPanel) FieldDate(format string) *InfoPanel {
	i.addDisplayChains(displayFnGens["date"].Get())
	return i
}

func (i *InfoPanel) FieldIcon(icons map[string]string, defaultIcon string) *InfoPanel {
	i.addDisplayChains(displayFnGens["link"].Get(icons, defaultIcon))
	return i
}

type FieldDotColor string

const (
	FieldDotColorDanger  FieldDotColor = "danger"
	FieldDotColorInfo    FieldDotColor = "info"
	FieldDotColorPrimary FieldDotColor = "primary"
	FieldDotColorSuccess FieldDotColor = "success"
)

func (i *InfoPanel) FieldDot(icons map[string]FieldDotColor, defaultDot FieldDotColor) *InfoPanel {
	i.addDisplayChains(displayFnGens["dot"].Get(icons, defaultDot))
	return i
}

type FieldProgressBarData struct {
	Style string
	Size  string
	Max   int
}

func (i *InfoPanel) FieldProgressBar(data ...FieldProgressBarData) *InfoPanel {
	i.addDisplayChains(displayFnGens["progressbar"].Get(data))
	return i
}

func (i *InfoPanel) FieldLoading(data []string) *InfoPanel {
	i.addDisplayChains(displayFnGens["loading"].Get(data))
	return i
}

func (i *InfoPanel) FieldDownLoadable(prefix ...string) *InfoPanel {
	i.addDisplayChains(displayFnGens["downloadable"].Get(prefix))
	return i
}

func (i *InfoPanel) FieldCopyable(prefix ...string) *InfoPanel {
	i.addDisplayChains(displayFnGens["copyable"].Get(prefix))
	if _, ok := i.DisplayGeneratorRecords["copyable"]; !ok {
		i.addFooterHTML(`<script>` + displayFnGens["copyable"].JS() + `</script>`)
		i.DisplayGeneratorRecords["copyable"] = struct{}{}
	}
	return i
}

type FieldGetImgArrFn func(value string) []string

func (i *InfoPanel) FieldCarousel(fn FieldGetImgArrFn, size ...int) *InfoPanel {
	i.addDisplayChains(displayFnGens["carousel"].Get(fn, size))
	return i
}

func (i *InfoPanel) FieldQrcode() *InfoPanel {
	i.addDisplayChains(displayFnGens["qrcode"].Get())
	if _, ok := i.DisplayGeneratorRecords["qrcode"]; !ok {
		i.addFooterHTML(`<script>` + displayFnGens["qrcode"].JS() + `</script>`)
		i.DisplayGeneratorRecords["qrcode"] = struct{}{}
	}
	return i
}

func (i *InfoPanel) FieldWidth(width int) *InfoPanel {
	i.FieldList[i.curFieldListIndex].Width = width
	return i
}

func (i *InfoPanel) FieldSortable() *InfoPanel {
	i.FieldList[i.curFieldListIndex].Sortable = true
	return i
}

func (i *InfoPanel) FieldEditOptions(options FieldOptions, extra ...map[string]string) *InfoPanel {
	if i.FieldList[i.curFieldListIndex].EditType.IsSwitch() {
		if len(extra) == 0 {
			options[0].Extra = map[string]string{
				"size":     "small",
				"onColor":  "primary",
				"offColor": "default",
			}
		} else {
			if extra[0]["size"] == "" {
				extra[0]["size"] = "small"
			}
			if extra[0]["onColor"] == "" {
				extra[0]["onColor"] = "primary"
			}
			if extra[0]["offColor"] == "" {
				extra[0]["offColor"] = "default"
			}
			options[0].Extra = extra[0]
		}
	}
	i.FieldList[i.curFieldListIndex].EditOptions = options
	return i
}

func (i *InfoPanel) FieldEditAble(editType ...table.Type) *InfoPanel {
	i.FieldList[i.curFieldListIndex].EditAble = true
	if len(editType) > 0 {
		i.FieldList[i.curFieldListIndex].EditType = editType[0]
	}
	return i
}

func (i *InfoPanel) FieldFixed() *InfoPanel {
	i.FieldList[i.curFieldListIndex].Fixed = true
	return i
}

type FilterType struct {
	FormType    form.Type
	Operator    FilterOperator
	Head        string
	Placeholder string
	NoHead      bool
	Width       int
	HelpMsg     template.HTML
	Options     FieldOptions
	Process     func(string) string
	OptionExt   map[string]interface{}
}

func (i *InfoPanel) FieldFilterable(filterType ...FilterType) *InfoPanel {
	i.FieldList[i.curFieldListIndex].Filterable = true

	if len(filterType) == 0 {
		i.FieldList[i.curFieldListIndex].FilterFormFields = append(i.FieldList[i.curFieldListIndex].FilterFormFields,
			FilterFormField{
				Type:        form.Text,
				Head:        i.FieldList[i.curFieldListIndex].Head,
				Placeholder: language.Get("input") + " " + i.FieldList[i.curFieldListIndex].Head,
			})
	}

	for _, filter := range filterType {
		var ff FilterFormField
		ff.Operator = filter.Operator
		if filter.FormType == form.Default {
			ff.Type = form.Text
		} else {
			ff.Type = filter.FormType
		}
		ff.Head = modules.AorB(!filter.NoHead && filter.Head == "",
			i.FieldList[i.curFieldListIndex].Head, filter.Head)
		ff.Width = filter.Width
		ff.HelpMsg = filter.HelpMsg
		ff.ProcessFn = filter.Process
		ff.Placeholder = modules.AorB(filter.Placeholder == "", language.Get("input")+" "+ff.Head, filter.Placeholder)
		ff.Options = filter.Options
		if len(filter.OptionExt) > 0 {
			s, _ := json.Marshal(filter.OptionExt)
			ff.OptionExt = template.JS(s)
		}
		i.FieldList[i.curFieldListIndex].FilterFormFields = append(i.FieldList[i.curFieldListIndex].FilterFormFields, ff)
	}

	return i
}

func (i *InfoPanel) FieldFilterOptions(options FieldOptions) *InfoPanel {
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].Options = options
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].OptionExt = `{"allowClear": "true"}`
	return i
}

func (i *InfoPanel) FieldFilterOptionsFromTable(table, textFieldName, valueFieldName string, process ...OptionTableQueryProcessFn) *InfoPanel {
	var fn OptionTableQueryProcessFn
	if len(process) > 0 {
		fn = process[0]
	}
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].OptionTable = OptionTable{
		Table:          table,
		TextField:      textFieldName,
		ValueField:     valueFieldName,
		QueryProcessFn: fn,
	}
	return i
}

func (i *InfoPanel) FieldFilterProcess(process func(string) string) *InfoPanel {
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].ProcessFn = process
	return i
}

func (i *InfoPanel) FieldFilterOptionExt(m map[string]interface{}) *InfoPanel {
	s, _ := json.Marshal(m)
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].OptionExt = template.JS(s)
	return i
}

func (i *InfoPanel) FieldFilterOnSearch(url string, handler Handler, delay ...int) *InfoPanel {
	ext, callback := searchJS(i.FieldList[i.curFieldListIndex].FilterFormFields[0].OptionExt, url, handler, delay...)
	i.FieldList[i.curFieldListIndex].FilterFormFields[0].OptionExt = ext
	i.Callbacks = append(i.Callbacks, callback)
	return i
}

func (i *InfoPanel) FieldFilterOnChooseCustom(js template.HTML) *InfoPanel {
	i.FooterHtml += chooseCustomJS(i.FieldList[i.curFieldListIndex].Field, js)
	return i
}

func (i *InfoPanel) FieldFilterOnChooseMap(m map[string]LinkField) *InfoPanel {
	i.FooterHtml += chooseMapJS(i.FieldList[i.curFieldListIndex].Field, m)
	return i
}

func (i *InfoPanel) FieldFilterOnChoose(val, field string, value template.HTML) *InfoPanel {
	i.FooterHtml += chooseJS(i.FieldList[i.curFieldListIndex].Field, field, val, value)
	return i
}

func (i *InfoPanel) FieldFilterOnChooseAjax(field, url string, handler Handler) *InfoPanel {
	js, callback := chooseAjax(i.FieldList[i.curFieldListIndex].Field, field, url, handler)
	i.FooterHtml += js
	i.Callbacks = append(i.Callbacks, callback)
	return i
}

func (i *InfoPanel) FieldFilterOnChooseHide(value string, field ...string) *InfoPanel {
	i.FooterHtml += chooseHideJS(i.FieldList[i.curFieldListIndex].Field, value, field...)
	return i
}

func (i *InfoPanel) FieldFilterOnChooseShow(value string, field ...string) *InfoPanel {
	i.FooterHtml += chooseShowJS(i.FieldList[i.curFieldListIndex].Field, value, field...)
	return i
}

func (i *InfoPanel) FieldFilterOnChooseDisable(value string, field ...string) *InfoPanel {
	i.FooterHtml += chooseDisableJS(i.FieldList[i.curFieldListIndex].Field, value, field...)
	return i
}

func (i *InfoPanel) FieldHide() *InfoPanel {
	i.FieldList[i.curFieldListIndex].Hide = true
	return i
}

func (i *InfoPanel) FieldJoin(join Join) *InfoPanel {
	i.FieldList[i.curFieldListIndex].Joins = append(i.FieldList[i.curFieldListIndex].Joins, join)
	return i
}

func (i *InfoPanel) FieldLimit(limit int) *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddLimit(limit)
	return i
}

func (i *InfoPanel) FieldTrimSpace() *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddTrimSpace()
	return i
}

func (i *InfoPanel) FieldSubstr(start int, end int) *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddSubstr(start, end)
	return i
}

func (i *InfoPanel) FieldToTitle() *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddToTitle()
	return i
}

func (i *InfoPanel) FieldToUpper() *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddToUpper()
	return i
}

func (i *InfoPanel) FieldToLower() *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].AddToLower()
	return i
}

func (i *InfoPanel) FieldXssFilter() *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains = i.FieldList[i.curFieldListIndex].DisplayProcessChains.
		Add(func(value FieldModel) interface{} {
			return html.EscapeString(value.Value)
		})
	return i
}

// InfoPanel attribute setting functions
// ====================================================

func (i *InfoPanel) SetTable(table string) *InfoPanel {
	i.Table = table
	return i
}

func (i *InfoPanel) SetPageSizeList(pageSizeList []int) *InfoPanel {
	i.PageSizeList = pageSizeList
	return i
}

func (i *InfoPanel) SetDefaultPageSize(defaultPageSize int) *InfoPanel {
	i.DefaultPageSize = defaultPageSize
	return i
}

func (i *InfoPanel) GetPageSizeList() []string {
	var pageSizeList = make([]string, len(i.PageSizeList))
	for j := 0; j < len(i.PageSizeList); j++ {
		pageSizeList[j] = strconv.Itoa(i.PageSizeList[j])
	}
	return pageSizeList
}

// 判斷資料是升冪或降冪
func (i *InfoPanel) GetSort() string {
	switch i.Sort {
	case SortAsc:
		return "asc"
	default:
		return "desc"
	}
}

func (i *InfoPanel) SetTitle(title string) *InfoPanel {
	i.Title = title
	return i
}

func (i *InfoPanel) SetTabGroups(groups TabGroups) *InfoPanel {
	i.TabGroups = groups
	return i
}

func (i *InfoPanel) SetTabHeaders(headers ...string) *InfoPanel {
	i.TabHeaders = headers
	return i
}

func (i *InfoPanel) SetDescription(desc string) *InfoPanel {
	i.Description = desc
	return i
}

func (i *InfoPanel) SetFilterFormLayout(layout form.Layout) *InfoPanel {
	i.FilterFormLayout = layout
	return i
}

func (i *InfoPanel) SetFilterFormHeadWidth(w int) *InfoPanel {
	i.FilterFormHeadWidth = w
	return i
}

func (i *InfoPanel) SetFilterFormInputWidth(w int) *InfoPanel {
	i.FilterFormInputWidth = w
	return i
}

func (i *InfoPanel) SetSortField(field string) *InfoPanel {
	i.SortField = field
	return i
}

func (i *InfoPanel) SetSortAsc() *InfoPanel {
	i.Sort = SortAsc
	return i
}

func (i *InfoPanel) SetSortDesc() *InfoPanel {
	i.Sort = SortDesc
	return i
}

func (i *InfoPanel) SetAction(action template.HTML) *InfoPanel {
	i.Action = action
	return i
}

func (i *InfoPanel) SetHeaderHtml(header template.HTML) *InfoPanel {
	i.HeaderHtml += header
	return i
}

func (i *InfoPanel) SetFooterHtml(footer template.HTML) *InfoPanel {
	i.FooterHtml += footer
	return i
}

func (i *InfoPanel) HideNewButton() *InfoPanel {
	i.IsHideNewButton = true
	return i
}

func (i *InfoPanel) HideExportButton() *InfoPanel {
	i.IsHideExportButton = true
	return i
}

func (i *InfoPanel) HideFilterButton() *InfoPanel {
	i.IsHideFilterButton = true
	return i
}

func (i *InfoPanel) HideRowSelector() *InfoPanel {
	i.IsHideRowSelector = true
	return i
}

func (i *InfoPanel) HidePagination() *InfoPanel {
	i.IsHidePagination = true
	return i
}

func (i *InfoPanel) HideFilterArea() *InfoPanel {
	i.IsHideFilterArea = true
	return i
}

func (i *InfoPanel) HideQueryInfo() *InfoPanel {
	i.IsHideQueryInfo = true
	return i
}

func (i *InfoPanel) HideEditButton() *InfoPanel {
	i.IsHideEditButton = true
	return i
}

func (i *InfoPanel) HideDeleteButton() *InfoPanel {
	i.IsHideDeleteButton = true
	return i
}

func (i *InfoPanel) HideDetailButton() *InfoPanel {
	i.IsHideDetailButton = true
	return i
}

func (i *InfoPanel) addFooterHTML(footer template.HTML) *InfoPanel {
	i.FooterHtml += template.HTML(ParseTableDataTmpl(footer))
	return i
}

func (i *InfoPanel) addCallback(node context.Node) *InfoPanel {
	i.Callbacks = i.Callbacks.AddCallback(node)
	return i
}

func (i *InfoPanel) addButton(btn Button) *InfoPanel {
	i.Buttons = append(i.Buttons, btn)
	return i
}

func (i *InfoPanel) addActionButton(btn Button) *InfoPanel {
	i.ActionButtons = append(i.ActionButtons, btn)
	return i
}

func (i *InfoPanel) isFromJSON() bool {
	return i.GetDataFn != nil
}

func (i *InfoPanel) addDisplayChains(fn FieldFilterFn) *InfoPanel {
	i.FieldList[i.curFieldListIndex].DisplayProcessChains =
		i.FieldList[i.curFieldListIndex].DisplayProcessChains.Add(fn)
	return i
}
