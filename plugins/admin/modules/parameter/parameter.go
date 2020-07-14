package parameter

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
)

type Parameters struct {
	Page        string
	PageInt     int
	PageSize    string
	PageSizeInt int
	SortField   string
	Columns     []string
	SortType    string
	Animation   bool
	URLPath     string
	Fields      map[string][]string
}

const (
	Page     = "__page"
	PageSize = "__pageSize"
	Sort     = "__sort"
	SortType = "__sort_type"
	Columns  = "__columns"
	Prefix   = "__prefix"
	Pjax     = "_pjax"

	sortTypeDesc = "desc"
	sortTypeAsc  = "asc"

	IsAll      = "__is_all"
	PrimaryKey = "__pk"

	True  = "true"
	False = "false"

	FilterRangeParamStartSuffix = "_start__goadmin"
	FilterRangeParamEndSuffix   = "_end__goadmin"
	FilterParamJoinInfix        = "_goadmin_join_"
	FilterParamOperatorSuffix   = "__goadmin_operator__"
	FilterParamCountInfix       = "__goadmin_index__"

	Separator = "__goadmin_separator__"
)

var operators = map[string]string{
	"like": "like",
	"gr":   ">",
	"gq":   ">=",
	"eq":   "=",
	"ne":   "!=",
	"le":   "<",
	"lq":   "<=",
	"free": "free",
}

var keys = []string{Page, PageSize, Sort, Columns, Prefix, Pjax, form.NoAnimationKey}

// 設置值(頁數及頁數Size)至Parameters(struct)並回傳
func BaseParam() Parameters {
	return Parameters{Page: "1", PageSize: "10", Fields: make(map[string][]string)}
}

// 取得頁面size、資料排列方式、選擇欄位...等資訊後設置至Parameters(struct)並回傳
func GetParam(u *url.URL, defaultPageSize int, p ...string) Parameters {

	// Query從url取得設定參數
	// ex: map[__columns:[id,username,name,goadmin_roles_goadmin_join_name,created_at,updated_at] __page:[1] __pageSize:[10]  __sort:[id] __sort_type:[desc] ...]
	values := u.Query()

	primaryKey := "id"
	defaultSortType := "desc"

	if len(p) > 0 {
		primaryKey = p[0]
		defaultSortType = p[1]
	}

	// getDefault透過參數key取得url中的值(value)，判斷是否為空，如果是空值回傳第三個參數def，如果不為空則回傳value
	page := getDefault(values, Page, "1")                                   // __page
	pageSize := getDefault(values, PageSize, strconv.Itoa(defaultPageSize)) // __pageSize
	sortField := getDefault(values, Sort, primaryKey)                       // __sort
	sortType := getDefault(values, SortType, defaultSortType)               // __sort_type
	// 選擇顯示的欄位
	columns := getDefault(values, Columns, "") // ex: id,username,name,goadmin_roles_goadmin_join_name,created_at,updated_at

	animation := true

	// form.NoAnimationKey = __go_admin_no_animation_
	// 判斷url中是否有動畫參數
	if values.Get(form.NoAnimationKey) == "true" {
		animation = false
	}

	// fields在下面迴圈處理後，ex:map[__goadmin_edit_pk:[4]] or []...等
	fields := make(map[string][]string)
	for key, value := range values {
		// keys []string{Page, PageSize, Sort, Columns, Prefix, Pjax, form.NoAnimationKey}
		if !modules.InArray(keys, key) && len(value) > 0 && value[0] != "" {
			// SortType = __sort_type
			if key == SortType {
				// sortTypeDesc = desc
				// sortTypeAsc = asc
				if value[0] != sortTypeDesc && value[0] != sortTypeAsc {
					fields[key] = []string{sortTypeDesc}
				}
			} else {
				// FilterParamOperatorSuffix = __goadmin_operator__
				if strings.Contains(key, FilterParamOperatorSuffix) &&
					values.Get(strings.Replace(key, FilterParamOperatorSuffix, "", -1)) == "" {
					continue
				}
				fields[strings.Replace(key, "[]", "", -1)] = value
			}
		}
	}

	columnsArr := make([]string, 0)

	// 如果有設定顯示欄位(則回傳欄位名稱至columnsArr，如果沒有設定則回傳空[])
	if columns != "" {
		columns, _ = url.QueryUnescape(columns)
		columnsArr = strings.Split(columns, ",")
	}

	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)

	return Parameters{
		Page:        page,
		PageSize:    pageSize,
		PageSizeInt: pageSizeInt,
		PageInt:     pageInt,
		URLPath:     u.Path,
		SortField:   sortField,
		SortType:    sortType,
		Fields:      fields,
		Animation:   animation,
		Columns:     columnsArr,
	}
}

// GetParamFromURL(從URL中取得參數)取得頁面size、資料排列方式、選擇欄位...等資訊後設置至Parameters(struct)並回傳
// defaultPageSize = 10,defaultSortType = desc or asc，primaryKey = id
func GetParamFromURL(urlStr string, defaultPageSize int, defaultSortType, primaryKey string) Parameters {

	// 解析url
	// ex: /admin/info/manager?__page=1&__pageSize=10&__sort=id&__sort_type=desc
	// 如果有選擇顯示欄位則還會得到ex:__columns=id,Cusername....
	u, err := url.Parse(urlStr)

	if err != nil {
		return BaseParam()
	}

	// 取得頁面size、資料排列方式、選擇欄位...等資訊後設置至Parameters(struct)並回傳
	return GetParam(u, defaultPageSize, primaryKey, defaultSortType)
}

// 將參數(多個string)結合並設置至Parameters.Fields["__pk"]後回傳
func (param Parameters) WithPKs(id ...string) Parameters {
	// PrimaryKey = __pk
	param.Fields[PrimaryKey] = []string{strings.Join(id, ",")}
	return param
}

// 透過參數__pk尋找Parameters.Fields[__pk]是否存在，如果存在則回傳第一個value值(string)並且用","拆解成[]string
func (param Parameters) PKs() []string {
	// 透過參數__pk尋找Parameters.Fields[__pk]是否存在，如果存在則回傳第一個value值(string)
	// PrimaryKey = PrimaryKey
	pk := param.GetFieldValue(PrimaryKey)
	if pk == "" {
		return []string{}
	}
	return strings.Split(param.GetFieldValue(PrimaryKey), ",")
}

// 刪除Parameters.Fields[__pk]後回傳
func (param Parameters) DeletePK() Parameters {
	// PrimaryKey = __pk
	delete(param.Fields, PrimaryKey)
	return param
}

// PK透過參數__pk尋找Parameters.Fields[__pk]是否存在，如果存在則回傳第一個value值(string)並且用","拆解成[]string，回傳第一個數值
func (param Parameters) PK() string {
	// PKs透過參數__pk尋找Parameters.Fields[__pk]是否存在，如果存在則回傳第一個value值(string)並且用","拆解成[]string
	return param.PKs()[0]
}

// 透過參數__is_all尋找Parameters.Fields[__is_all]是否存在，如果存在則回傳True，反之false
func (param Parameters) IsAll() bool {
	return param.GetFieldValue(IsAll) == True
}

func (param Parameters) WithURLPath(path string) Parameters {
	param.URLPath = path
	return param
}

// 判斷條件後，在param.Fields[參數](map[string][]string)加入值
func (param Parameters) WithIsAll(isAll bool) Parameters {
	if isAll {
		param.Fields[IsAll] = []string{True}
	} else {
		param.Fields[IsAll] = []string{False}
	}
	return param
}

func (param Parameters) DeleteIsAll() Parameters {
	delete(param.Fields, IsAll)
	return param
}

func (param Parameters) GetFilterFieldValueStart(field string) string {
	return param.GetFieldValue(field + FilterRangeParamStartSuffix)
}

func (param Parameters) GetFilterFieldValueEnd(field string) string {
	return param.GetFieldValue(field + FilterRangeParamEndSuffix)
}

// 透過參數field尋找Parameters.Fields[field]是否存在，如果存在則回傳第一個value值(string)，不存在則回傳""
func (param Parameters) GetFieldValue(field string) string {
	value, ok := param.Fields[field]
	if ok && len(value) > 0 {
		return value[0]
	}
	return ""
}

func (param Parameters) AddField(field, value string) Parameters {
	param.Fields[field] = []string{value}
	return param
}

func (param Parameters) DeleteField(field string) Parameters {
	delete(param.Fields, field)
	return param
}

// 刪除Parameters.Fields[__goadmin_edit_pk]後回傳
func (param Parameters) DeleteEditPk() Parameters {
	// EditPKKey = __goadmin_edit_pk
	delete(param.Fields, constant.EditPKKey)
	return param
}

func (param Parameters) DeleteDetailPk() Parameters {
	delete(param.Fields, constant.DetailPKKey)
	return param
}

func (param Parameters) GetFieldValues(field string) []string {
	return param.Fields[field]
}

func (param Parameters) GetFieldValuesStr(field string) string {
	return strings.Join(param.Fields[field], Separator)
}

// 透過參數field+ __goadmin_operator__+suffix尋找Parameters.Fields[field+ __goadmin_operator__+suffix]是否存在(op)，如果op為空則回傳eq，反之則回傳op(string)
func (param Parameters) GetFieldOperator(field, suffix string) string {
	// 透過參數field+ __goadmin_operator__+suffix尋找Parameters.Fields[field+ __goadmin_operator__+suffix]是否存在，如果存在則回傳第一個value值(string)
	// FilterParamOperatorSuffix = __goadmin_operator__
	op := param.GetFieldValue(field + FilterParamOperatorSuffix + suffix)
	if op == "" {
		return "eq"
	}
	return op
}

func (param Parameters) Join() string {
	p := param.GetFixedParamStr()
	p.Add(Page, param.Page)
	return p.Encode()
}

func (param Parameters) SetPage(page string) Parameters {
	param.Page = page
	return param
}

// 取得url.Values後加入__page(鍵)與值，最後編碼並回傳
func (param Parameters) GetRouteParamStr() string {
	// GetFixedParamStr將Parameters(struct)的鍵與值加入至url.Values並回傳
	p := param.GetFixedParamStr()
	p.Add(Page, param.Page)

	return "?" + p.Encode()
}

func (param Parameters) URL(page string) string {
	return param.URLPath + param.SetPage(page).GetRouteParamStr()
}

func (param Parameters) URLNoAnimation(page string) string {
	return param.URLPath + param.SetPage(page).GetRouteParamStr() + "&" + form.NoAnimationKey + "=true"
}

func (param Parameters) GetRouteParamStrWithoutPageSize(page string) string {
	p := url.Values{}
	p.Add(Sort, param.SortField)
	p.Add(Page, page)
	p.Add(SortType, param.SortType)
	if len(param.Columns) > 0 {
		p.Add(Columns, strings.Join(param.Columns, ","))
	}
	for key, value := range param.Fields {
		p[key] = value
	}
	return "?" + p.Encode()
}

func (param Parameters) GetLastPageRouteParamStr() string {
	p := param.GetFixedParamStr()
	p.Add(Page, strconv.Itoa(param.PageInt-1))
	return "?" + p.Encode()
}

func (param Parameters) GetNextPageRouteParamStr() string {
	p := param.GetFixedParamStr()
	p.Add(Page, strconv.Itoa(param.PageInt+1))
	return "?" + p.Encode()
}

// 將Parameters(struct)的鍵與值加入至url.Values並回傳
func (param Parameters) GetFixedParamStr() url.Values {
	p := url.Values{}
	p.Add(Sort, param.SortField)
	p.Add(PageSize, param.PageSize)
	p.Add(SortType, param.SortType)
	if len(param.Columns) > 0 {
		p.Add(Columns, strings.Join(param.Columns, ","))
	}
	for key, value := range param.Fields {
		p[key] = value
	}

	return p
}

func (param Parameters) GetFixedParamStrWithoutColumnsAndPage() string {
	p := url.Values{}
	p.Add(Sort, param.SortField)
	p.Add(PageSize, param.PageSize)
	if len(param.Columns) > 0 {
		p.Add(Columns, strings.Join(param.Columns, ","))
	}
	p.Add(SortType, param.SortType)
	return "?" + p.Encode()
}

func (param Parameters) GetFixedParamStrWithoutSort() string {
	p := url.Values{}
	p.Add(PageSize, param.PageSize)
	for key, value := range param.Fields {
		p[key] = value
	}
	p.Add(form.NoAnimationKey, "true")
	if len(param.Columns) > 0 {
		p.Add(Columns, strings.Join(param.Columns, ","))
	}
	return "&" + p.Encode()
}

// 處理param.Fields(map[string][]string)，接著依判斷條件處理參數，最後回傳string, []interface{}, []string
func (param Parameters) Statement(wheres, table, delimiter string, whereArgs []interface{}, columns, existKeys []string,
	filterProcess func(string, string, string) string) (string, []interface{}, []string) {
	var multiKey = make(map[string]uint8)

	// 處理param.Fields，例如用戶介面為ex: map[__is_all:[false]]
	for key, value := range param.Fields {
		keyIndexSuffix := ""

		// FilterParamCountInfix = __goadmin_index__
		keyArr := strings.Split(key, FilterParamCountInfix)

		if len(keyArr) > 1 {
			key = keyArr[0]
			keyIndexSuffix = FilterParamCountInfix + keyArr[1]
		}

		// -------用戶頁面不會執行-------
		if keyIndexSuffix != "" {
			multiKey[key] = 0
		} else if _, exist := multiKey[key]; !exist && modules.InArray(existKeys, key) {
			continue
		}

		var op string
		if strings.Contains(key, FilterRangeParamEndSuffix) {
			key = strings.Replace(key, FilterRangeParamEndSuffix, "", -1)
			op = "<="
		} else if strings.Contains(key, FilterRangeParamStartSuffix) {
			key = strings.Replace(key, FilterRangeParamStartSuffix, "", -1)
			op = ">="
		} else if len(value) > 1 {
			op = "in"
		} else if !strings.Contains(key, FilterParamOperatorSuffix) {
			// -------用戶頁面會執行-----------
			// GetFieldOperator透過參數key+ __goadmin_operator__+keyIndexSuffix尋找Parameters.Fields[key+ __goadmin_operator__+keyIndexSuffix]是否存在(op)，如果op為空則回傳eq，反之則回傳op(string)
			op = operators[param.GetFieldOperator(key, keyIndexSuffix)] // op = '='(用戶頁面)
		}

		if modules.InArray(columns, key) {
			if op == "in" {
				qmark := ""
				for range value {
					qmark += "?,"
				}
				wheres += table + "." + modules.FilterField(key, delimiter) + " " + op + " (" + qmark[:len(qmark)-1] + ") and "
			} else {
				wheres += table + "." + modules.FilterField(key, delimiter) + " " + op + " ? and "
			}
			if op == "like" && !strings.Contains(value[0], "%") {
				whereArgs = append(whereArgs, "%"+filterProcess(key, value[0], keyIndexSuffix)+"%")
			} else {
				for _, v := range value {
					whereArgs = append(whereArgs, filterProcess(key, v, keyIndexSuffix))
				}
			}
		} else {
			// --------用戶頁面會執行--------
			keys := strings.Split(key, FilterParamJoinInfix) // ex: [__is_all](用戶頁面)

			// ------用戶頁面不會執行--------
			if len(keys) > 1 {
				val := filterProcess(key, value[0], keyIndexSuffix)
				if op == "in" {
					qmark := ""
					for range value {
						qmark += "?,"
					}
					wheres += keys[0] + "." + modules.FilterField(keys[1], delimiter) + " " + op + " (" + qmark[:len(qmark)-1] + ") and "
				} else {
					wheres += keys[0] + "." + modules.FilterField(keys[1], delimiter) + " " + op + " ? and "
				}
				if op == "like" && !strings.Contains(val, "%") {
					whereArgs = append(whereArgs, "%"+val+"%")
				} else {
					for _, v := range value {
						whereArgs = append(whereArgs, filterProcess(key, v, keyIndexSuffix))
					}
				}
			}
		}

		existKeys = append(existKeys, key) // ex:[__is_all](用戶頁面)
	}

	if len(wheres) > 3 {
		wheres = wheres[:len(wheres)-4]
	}

	return wheres, whereArgs, existKeys
}

// 透過參數key取得url中的值(value)，判斷是否為空，如果是空值回傳第三個參數def，如果不為空則回傳value
func getDefault(values url.Values, key, def string) string {
	value := values.Get(key)
	if value == "" {
		return def
	}
	return value
}
