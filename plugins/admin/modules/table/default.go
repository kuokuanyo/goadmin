package table

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
	errs "github.com/GoAdminGroup/go-admin/modules/errors"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/logger"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/form"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/paginator"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/template/types"
)

type DefaultTable struct {
	*BaseTable
	connectionDriver string
	connection       string
	sourceURL        string
	getDataFun       GetDataFun
}

type GetDataFun func(params parameter.Parameters) ([]map[string]interface{}, int)

func NewDefaultTable(cfgs ...Config) Table {

	var cfg Config

	if len(cfgs) > 0 && cfgs[0].PrimaryKey.Name != "" {
		cfg = cfgs[0]
	} else {
		cfg = DefaultConfig()
	}

	return &DefaultTable{
		BaseTable: &BaseTable{
			Info:           types.NewInfoPanel(cfg.PrimaryKey.Name),
			Form:           types.NewFormPanel(),
			Detail:         types.NewInfoPanel(cfg.PrimaryKey.Name),
			CanAdd:         cfg.CanAdd,
			Editable:       cfg.Editable,
			Deletable:      cfg.Deletable,
			Exportable:     cfg.Exportable,
			PrimaryKey:     cfg.PrimaryKey,
			OnlyNewForm:    cfg.OnlyNewForm,
			OnlyUpdateForm: cfg.OnlyUpdateForm,
			OnlyDetail:     cfg.OnlyDetail,
			OnlyInfo:       cfg.OnlyInfo,
		},
		connectionDriver: cfg.Driver,
		connection:       cfg.Connection,
		sourceURL:        cfg.SourceURL,
		getDataFun:       cfg.GetDataFun,
	}
}

func (tb *DefaultTable) Copy() Table {
	return &DefaultTable{
		BaseTable: &BaseTable{
			Form: types.NewFormPanel().SetTable(tb.Form.Table).
				SetDescription(tb.Form.Description).
				SetTitle(tb.Form.Title),
			Info: types.NewInfoPanel(tb.PrimaryKey.Name).SetTable(tb.Info.Table).
				SetDescription(tb.Info.Description).
				SetTitle(tb.Info.Title).
				SetGetDataFn(tb.Info.GetDataFn),
			Detail: types.NewInfoPanel(tb.PrimaryKey.Name).SetTable(tb.Detail.Table).
				SetDescription(tb.Detail.Description).
				SetTitle(tb.Detail.Title).
				SetGetDataFn(tb.Detail.GetDataFn),
			CanAdd:     tb.CanAdd,
			Editable:   tb.Editable,
			Deletable:  tb.Deletable,
			Exportable: tb.Exportable,
			PrimaryKey: tb.PrimaryKey,
		},
		connectionDriver: tb.connectionDriver,
		connection:       tb.connection,
		sourceURL:        tb.sourceURL,
		getDataFun:       tb.getDataFun,
	}
}

// GetData query the data set.
func (tb *DefaultTable) GetData(params parameter.Parameters) (PanelInfo, error) {

	var (
		data      []map[string]interface{}
		size      int
		beginTime = time.Now()
	)

	if tb.Info.QueryFilterFn != nil {
		ids, stop := tb.Info.QueryFilterFn(params, tb.db())
		if stop {
			return tb.GetDataWithIds(params.WithPKs(ids...))
		}
	}

	if tb.getDataFun != nil {
		data, size = tb.getDataFun(params)
	} else if tb.sourceURL != "" {
		data, size = tb.getDataFromURL(params)
	} else if tb.Info.GetDataFn != nil {
		data, size = tb.Info.GetDataFn(params)
	} else if params.IsAll() {
		return tb.getAllDataFromDatabase(params)
	} else {
		return tb.getDataFromDatabase(params)
	}

	infoList := make(types.InfoList, 0)

	for i := 0; i < len(data); i++ {
		infoList = append(infoList, tb.getTempModelData(data[i], params, []string{}))
	}

	thead, _, _, _, _, filterForm := tb.getTheadAndFilterForm(params, []string{})

	endTime := time.Now()

	extraInfo := ""

	if !tb.Info.IsHideQueryInfo {
		extraInfo = fmt.Sprintf("<b>" + language.Get("query time") + ": </b>" +
			fmt.Sprintf("%.3fms", endTime.Sub(beginTime).Seconds()*1000))
	}

	return PanelInfo{
		Thead:    thead,
		InfoList: infoList,
		Paginator: paginator.Get(paginator.Config{
			Size:         size,
			Param:        params,
			PageSizeList: tb.Info.GetPageSizeList(),
		}).SetExtraInfo(template.HTML(extraInfo)),
		Title:          tb.Info.Title,
		FilterFormData: filterForm,
		Description:    tb.Info.Description,
	}, nil
}

type GetDataFromURLRes struct {
	Data []map[string]interface{}
	Size int
}

// getDataFromURL(從url中取得data)
func (tb *DefaultTable) getDataFromURL(params parameter.Parameters) ([]map[string]interface{}, int) {

	u := ""
	if strings.Contains(tb.sourceURL, "?") {
		u = tb.sourceURL + "&" + params.Join()
	} else {
		u = tb.sourceURL + "?" + params.Join()
	}
	res, err := http.Get(u + "&pk=" + strings.Join(params.PKs(), ","))

	if err != nil {
		return []map[string]interface{}{}, 0
	}

	defer func() {
		_ = res.Body.Close()
	}()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return []map[string]interface{}{}, 0
	}

	var data GetDataFromURLRes

	err = json.Unmarshal(body, &data)

	if err != nil {
		return []map[string]interface{}{}, 0
	}

	return data.Data, data.Size
}

// GetDataWithIds query the data set.
func (tb *DefaultTable) GetDataWithIds(params parameter.Parameters) (PanelInfo, error) {

	var (
		data      []map[string]interface{}
		size      int
		beginTime = time.Now()
	)

	if tb.getDataFun != nil {
		data, size = tb.getDataFun(params)
	} else if tb.sourceURL != "" {
		data, size = tb.getDataFromURL(params)
	} else if tb.Info.GetDataFn != nil {
		data, size = tb.Info.GetDataFn(params)
	} else {
		return tb.getDataFromDatabase(params)
	}

	infoList := make([]map[string]types.InfoItem, 0)

	for i := 0; i < len(data); i++ {
		infoList = append(infoList, tb.getTempModelData(data[i], params, []string{}))
	}

	thead, _, _, _, _, filterForm := tb.getTheadAndFilterForm(params, []string{})

	endTime := time.Now()

	return PanelInfo{
		Thead:    thead,
		InfoList: infoList,
		Paginator: paginator.Get(paginator.Config{
			Size:         size,
			Param:        params,
			PageSizeList: tb.Info.GetPageSizeList(),
		}).
			SetExtraInfo(template.HTML(fmt.Sprintf("<b>" + language.Get("query time") + ": </b>" +
				fmt.Sprintf("%.3fms", endTime.Sub(beginTime).Seconds()*1000)))),
		Title:          tb.Info.Title,
		FilterFormData: filterForm,
		Description:    tb.Info.Description,
	}, nil
}

func (tb *DefaultTable) getTempModelData(res map[string]interface{}, params parameter.Parameters, columns Columns) map[string]types.InfoItem {

	var tempModelData = make(map[string]types.InfoItem)
	headField := ""

	primaryKeyValue := db.GetValueFromDatabaseType(tb.PrimaryKey.Type, res[tb.PrimaryKey.Name], len(columns) == 0)

	for _, field := range tb.Info.FieldList {

		headField = field.Field

		if field.Joins.Valid() {
			headField = field.Joins.Last().Table + parameter.FilterParamJoinInfix + field.Field
		}

		if field.Hide {
			continue
		}
		if !modules.InArrayWithoutEmpty(params.Columns, headField) {
			continue
		}

		typeName := field.TypeName

		if field.Joins.Valid() {
			typeName = db.Varchar
		}

		combineValue := db.GetValueFromDatabaseType(typeName, res[headField], len(columns) == 0).String()

		// TODO: ToDisplay some same logic execute repeatedly, it can be improved.
		var value interface{}
		if len(columns) == 0 || modules.InArray(columns, headField) || field.Joins.Valid() {
			value = field.ToDisplay(types.FieldModel{
				ID:    primaryKeyValue.String(),
				Value: combineValue,
				Row:   res,
			})
		} else {
			value = field.ToDisplay(types.FieldModel{
				ID:    primaryKeyValue.String(),
				Value: "",
				Row:   res,
			})
		}
		if valueStr, ok := value.(string); ok {
			tempModelData[headField] = types.InfoItem{
				Content: template.HTML(valueStr),
				Value:   combineValue,
			}
		} else {
			tempModelData[headField] = types.InfoItem{
				Content: value.(template.HTML),
				Value:   combineValue,
			}
		}
	}

	primaryKeyField := tb.Info.FieldList.GetFieldByFieldName(tb.PrimaryKey.Name)
	value := primaryKeyField.ToDisplay(types.FieldModel{
		ID:    primaryKeyValue.String(),
		Value: primaryKeyValue.String(),
		Row:   res,
	})
	if valueStr, ok := value.(string); ok {
		tempModelData[tb.PrimaryKey.Name] = types.InfoItem{
			Content: template.HTML(valueStr),
			Value:   primaryKeyValue.String(),
		}
	} else {
		tempModelData[tb.PrimaryKey.Name] = types.InfoItem{
			Content: value.(template.HTML),
			Value:   primaryKeyValue.String(),
		}
	}
	return tempModelData
}

func (tb *DefaultTable) getAllDataFromDatabase(params parameter.Parameters) (PanelInfo, error) {
	var (
		connection     = tb.db()
		queryStatement = "select %s from %s %s %s %s order by " + modules.Delimiter(connection.GetDelimiter(), "%s") + " %s"
	)

	columns, _ := tb.getColumns(tb.Info.Table)

	thead, fields, joins := tb.Info.FieldList.GetThead(types.TableInfo{
		Table:      tb.Info.Table,
		Delimiter:  tb.db().GetDelimiter(),
		Driver:     tb.connectionDriver,
		PrimaryKey: tb.PrimaryKey.Name,
	}, params, columns)

	fields += tb.Info.Table + "." + modules.FilterField(tb.PrimaryKey.Name, connection.GetDelimiter())

	groupBy := ""
	if joins != "" {
		groupBy = " GROUP BY " + tb.Info.Table + "." + modules.Delimiter(connection.GetDelimiter(), tb.PrimaryKey.Name)
	}

	var (
		wheres    = ""
		whereArgs = make([]interface{}, 0)
		existKeys = make([]string, 0)
	)

	wheres, whereArgs, existKeys = params.Statement(wheres, tb.Info.Table, connection.GetDelimiter(), whereArgs, columns, existKeys,
		tb.Info.FieldList.GetFieldFilterProcessValue)
	wheres, whereArgs = tb.Info.Wheres.Statement(wheres, connection.GetDelimiter(), whereArgs, existKeys, columns)
	wheres, whereArgs = tb.Info.WhereRaws.Statement(wheres, whereArgs)

	if wheres != "" {
		wheres = " where " + wheres
	}

	if !modules.InArray(columns, params.SortField) {
		params.SortField = tb.PrimaryKey.Name
	}

	queryCmd := fmt.Sprintf(queryStatement, fields, tb.Info.Table, joins, wheres, groupBy, params.SortField, params.SortType)

	logger.LogSQL(queryCmd, []interface{}{})

	res, err := connection.QueryWithConnection(tb.connection, queryCmd, whereArgs...)

	if err != nil {
		return PanelInfo{}, err
	}

	infoList := make([]map[string]types.InfoItem, 0)

	for i := 0; i < len(res); i++ {
		infoList = append(infoList, tb.getTempModelData(res[i], params, columns))
	}

	return PanelInfo{
		InfoList:    infoList,
		Thead:       thead,
		Title:       tb.Info.Title,
		Description: tb.Info.Description,
	}, nil
}

// TODO: refactor
func (tb *DefaultTable) getDataFromDatabase(params parameter.Parameters) (PanelInfo, error) {

	var (
		connection     = tb.db()
		placeholder    = modules.Delimiter(connection.GetDelimiter(), "%s")
		queryStatement string
		countStatement string
		ids            = params.PKs()
		pk             = tb.Info.Table + "." + modules.Delimiter(connection.GetDelimiter(), tb.PrimaryKey.Name)
	)

	if connection.Name() == db.DriverPostgresql {
		placeholder = "%s"
	}

	beginTime := time.Now()

	if len(ids) > 0 {
		countExtra := ""
		if connection.Name() == db.DriverMssql {
			countExtra = "as [size]"
		}
		// %s means: fields, table, join table, pk values, group by, order by field,  order by type
		queryStatement = "select %s from " + placeholder + " %s where " + pk + " in (%s) %s ORDER BY %s." + placeholder + " %s"
		// %s means: table, join table, pk values
		countStatement = "select count(*) " + countExtra + " from " + placeholder + " %s where " + pk + " in (%s)"
	} else {
		if connection.Name() == db.DriverMssql {
			// %s means: order by field, order by type, fields, table, join table, wheres, group by
			queryStatement = "SELECT * FROM (SELECT ROW_NUMBER() OVER (ORDER BY %s." + placeholder + " %s) as ROWNUMBER_, %s from " +
				placeholder + "%s %s %s ) as TMP_ WHERE TMP_.ROWNUMBER_ > ? AND TMP_.ROWNUMBER_ <= ?"
			// %s means: table, join table, wheres
			countStatement = "select count(*) as [size] from " + placeholder + " %s %s"
		} else {
			// %s means: fields, table, join table, wheres, group by, order by field, order by type
			queryStatement = "select %s from " + placeholder + "%s %s %s order by %s." + placeholder + " %s LIMIT ? OFFSET ?"
			// %s means: table, join table, wheres
			countStatement = "select count(*) from " + placeholder + " %s %s"
		}
	}

	columns, _ := tb.getColumns(tb.Info.Table)

	thead, fields, joinFields, joins, joinTables, filterForm := tb.getTheadAndFilterForm(params, columns)

	fields += pk

	allFields := fields
	groupFields := fields

	if joinFields != "" {
		allFields += "," + joinFields[:len(joinFields)-1]
		if connection.Name() == db.DriverMssql {
			for _, field := range tb.Info.FieldList {
				if field.TypeName == db.Text || field.TypeName == db.Longtext {
					f := modules.Delimiter(connection.GetDelimiter(), field.Field)
					headField := tb.Info.Table + "." + f
					allFields = strings.Replace(allFields, headField, "CAST("+headField+" AS NVARCHAR(MAX)) as "+f, -1)
					groupFields = strings.Replace(groupFields, headField, "CAST("+headField+" AS NVARCHAR(MAX))", -1)
				}
			}
		}
	}

	if !modules.InArray(columns, params.SortField) {
		params.SortField = tb.PrimaryKey.Name
	}

	var (
		wheres    = ""
		whereArgs = make([]interface{}, 0)
		args      = make([]interface{}, 0)
		existKeys = make([]string, 0)
	)

	if len(ids) > 0 {
		for _, value := range ids {
			if value != "" {
				wheres += "?,"
				args = append(args, value)
			}
		}
		wheres = wheres[:len(wheres)-1]
	} else {

		// parameter
		wheres, whereArgs, existKeys = params.Statement(wheres, tb.Info.Table, connection.GetDelimiter(), whereArgs, columns, existKeys,
			tb.Info.FieldList.GetFieldFilterProcessValue)
		// pre query
		wheres, whereArgs = tb.Info.Wheres.Statement(wheres, connection.GetDelimiter(), whereArgs, existKeys, columns)
		wheres, whereArgs = tb.Info.WhereRaws.Statement(wheres, whereArgs)

		if wheres != "" {
			wheres = " where " + wheres
		}

		if connection.Name() == db.DriverMssql {
			args = append(whereArgs, (params.PageInt-1)*params.PageSizeInt, params.PageInt*params.PageSizeInt)
		} else {
			args = append(whereArgs, params.PageSizeInt, (params.PageInt-1)*params.PageSizeInt)
		}
	}

	groupBy := ""
	if len(joinTables) > 0 {
		if connection.Name() == db.DriverMssql {
			groupBy = " GROUP BY " + groupFields
		} else {
			groupBy = " GROUP BY " + pk
		}
	}

	queryCmd := ""
	if connection.Name() == db.DriverMssql && len(ids) == 0 {
		queryCmd = fmt.Sprintf(queryStatement, tb.Info.Table, params.SortField, params.SortType,
			allFields, tb.Info.Table, joins, wheres, groupBy)
	} else {
		queryCmd = fmt.Sprintf(queryStatement, allFields, tb.Info.Table, joins, wheres, groupBy,
			tb.Info.Table, params.SortField, params.SortType)
	}

	logger.LogSQL(queryCmd, args)

	res, err := connection.QueryWithConnection(tb.connection, queryCmd, args...)

	if err != nil {
		return PanelInfo{}, err
	}

	infoList := make([]map[string]types.InfoItem, 0)

	for i := 0; i < len(res); i++ {
		infoList = append(infoList, tb.getTempModelData(res[i], params, columns))
	}

	// TODO: use the dialect
	var size int

	if len(ids) == 0 {
		countCmd := fmt.Sprintf(countStatement, tb.Info.Table, joins, wheres)

		total, err := connection.QueryWithConnection(tb.connection, countCmd, whereArgs...)

		if err != nil {
			return PanelInfo{}, err
		}

		logger.LogSQL(countCmd, nil)

		if tb.connectionDriver == "postgresql" {
			size = int(total[0]["count"].(int64))
		} else if tb.connectionDriver == db.DriverMssql {
			size = int(total[0]["size"].(int64))
		} else {
			size = int(total[0]["count(*)"].(int64))
		}
	}

	endTime := time.Now()

	return PanelInfo{
		Thead:    thead,
		InfoList: infoList,
		Paginator: tb.GetPaginator(size, params,
			template.HTML(fmt.Sprintf("<b>"+language.Get("query time")+": </b>"+
				fmt.Sprintf("%.3fms", endTime.Sub(beginTime).Seconds()*1000)))),
		Title:          tb.Info.Title,
		FilterFormData: filterForm,
		Description:    tb.Info.Description,
	}, nil
}

// 假設參數list([]map[string]interface{})長度大於零則回傳list[0](map[string]interface{})
func getDataRes(list []map[string]interface{}, _ int) map[string]interface{} {
	if len(list) > 0 {
		return list[0]
	}
	return nil
}

// GetDataWithId query the single row of data.
// GetDataWithId(透過id取得資料)透過id取得goadmin_menu資料表中的資料，接著對有帶值的欄位更新並加入FormFields後回傳，最後設置值至FormInfo(struct)中
func (tb *DefaultTable) GetDataWithId(param parameter.Parameters) (FormInfo, error) {

	var (
		res     map[string]interface{}
		columns Columns
		// PK透過參數__pk尋找Parameters.Fields[__pk]是否存在，如果存在則回傳第一個value值(string)並且用","拆解成[]string，回傳第一個數值
		id      = param.PK()
	)

	// getDataRes假設參數list([]map[string]interface{})長度大於零則回傳list[0](map[string]interface{})
	if tb.getDataFun != nil {
		res = getDataRes(tb.getDataFun(param))
	} else if tb.sourceURL != "" {
		res = getDataRes(tb.getDataFromURL(param))
	} else if tb.Detail.GetDataFn != nil {
		res = getDataRes(tb.Detail.GetDataFn(param))
	} else if tb.Info.GetDataFn != nil {
		res = getDataRes(tb.Info.GetDataFn(param))
	} else {

		// getColumns(取得欄位)將欄位名稱加入columns([]string)
		// 如果有值是primary_key並且自動遞增則bool = true，最後回傳欄位名稱及bool
		// columns為goadmin_menu所有欄位名稱
		columns, _ = tb.getColumns(tb.Form.Table)

		var (
			fields, joinFields, joins, groupBy = "", "", "", ""

			err            error
			joinTables     = make([]string, 0)
			// args為編輯的id
			args           = []interface{}{id}
			connection     = tb.db()
			// db透過參數(k)取得匹配的Service(interface)，接著將參數services.Get(tb.connectionDriver)轉換為Connection(interface)回傳並回傳
			delimiter      = connection.GetDelimiter()
			// GetForm將參數值設置至BaseTable.Form(FormPanel(struct)).primaryKey中後回傳
			// tablename = goadmin_menu
			tableName      = tb.GetForm().Table
			// Delimiter在plugins\admin\modules\table\default.go
			// Delimiter判斷參數del後回傳del+s(參數)+del或[s(參數)]
			// pk = goadmin_menu.'id'
			pk             = tableName + "." + modules.Delimiter(delimiter, tb.PrimaryKey.Name)
			// queryStatement取得goadmin_menu某一筆資料(根據id)
			queryStatement = "select %s from " + modules.Delimiter(delimiter, "%s") + " %s where " + pk + " = ? %s "
		)

		if connection.Name() == db.DriverPostgresql {
			queryStatement = "select %s from %s %s where " + pk + " = ? %s "
		}

		// tb.Form.FieldList為表單所有欄位資訊
		for _, field := range tb.Form.FieldList {

			if field.Field != pk && modules.InArray(columns, field.Field) &&
			// Valid在template\types\info.go
			// 對joins([]join(struct))執行迴圈，假設Join的Table、Field、JoinField不為空，回傳true
				!field.Joins.Valid() {
				// 將所有欄位名稱�m加上資料表名(ex:tablename.colname)
				// ex:goadmin_menu.`id`,goadmin_menu.`parent_id`,goadmin_menu.`title`,...
				fields += tableName + "." + modules.FilterField(field.Field, delimiter) + ","
			}

			headField := field.Field

			// 在編輯頁面時不會執行下列判斷(沒有join)
			if field.Joins.Valid() {
				headField = field.Joins.Last().Table + parameter.FilterParamJoinInfix + field.Field
				joinFields += db.GetAggregationExpression(connection.Name(), field.Joins.Last().Table+"."+
					modules.FilterField(field.Field, delimiter), headField, types.JoinFieldValueDelimiter) + ","
				for _, join := range field.Joins {
					if !modules.InArray(joinTables, join.Table) {
						joinTables = append(joinTables, join.Table)
						if join.BaseTable == "" {
							join.BaseTable = tableName
						}
						joins += " left join " + modules.FilterField(join.Table, delimiter) + " on " +
							join.Table + "." + modules.FilterField(join.JoinField, delimiter) + " = " +
							join.BaseTable + "." + modules.FilterField(join.Field, delimiter)
					}
				}
			}
		}

		// fields再加上"goadmin_menu.`id`"
		fields += pk
		groupFields := fields

		// 在編輯頁面時不會執行下列判斷(沒有joinFields)
		if joinFields != "" {
			fields += "," + joinFields[:len(joinFields)-1]
			if connection.Name() == db.DriverMssql {
				for _, field := range tb.Form.FieldList {
					if field.TypeName == db.Text || field.TypeName == db.Longtext {
						f := modules.Delimiter(connection.GetDelimiter(), field.Field)
						headField := tb.Info.Table + "." + f
						fields = strings.Replace(fields, headField, "CAST("+headField+" AS NVARCHAR(MAX)) as "+f, -1)
						groupFields = strings.Replace(groupFields, headField, "CAST("+headField+" AS NVARCHAR(MAX))", -1)
					}
				}
			}
		}

		// 在編輯頁面時不會執行下列判斷(沒有joinTables)
		if len(joinTables) > 0 {
			if connection.Name() == db.DriverMssql {
				groupBy = " GROUP BY " + groupFields
			} else {
				groupBy = " GROUP BY " + pk
			}
		}

		queryCmd := fmt.Sprintf(queryStatement, fields, tableName, joins, groupBy)

		// 印出sql資料(編輯頁面時沒有印出)
		logger.LogSQL(queryCmd, args)

		// 取得單筆資料(利用id)
		// QueryWithConnection(connection方法)在admin\modules\db\mysql.go
		// QueryWithConnection有給定連接(tb.connection)名稱，透過參數tb.connection查詢db.DbList[tb.connection]資料並回傳
		result, err := connection.QueryWithConnection(tb.connection, queryCmd, args...)

		if err != nil {
			// tb.Form.Title主題左上角(ex:菜單管理)
			// tb.Form.Description主題旁邊的描述(ex:菜單管理)
			return FormInfo{Title: tb.Form.Title, Description: tb.Form.Description}, err
		}

		if len(result) == 0 {
			return FormInfo{Title: tb.Form.Title, Description: tb.Form.Description}, errors.New(errs.WrongID)
		}

		res = result[0]
	}

	var (
		// 編輯頁面時，groupFormList、groupHeaders都為空
		groupFormList = make([]types.FormFields, 0)
		groupHeaders  = make([]string, 0)
	)

	// 在編輯頁面時，沒有tb.Form.TabGroups(組標籤)
	if len(tb.Form.TabGroups) > 0 {
		groupFormList, groupHeaders = tb.Form.GroupFieldWithValue(tb.PrimaryKey.Name, id, columns, res, tb.sql)
		return FormInfo{
			FieldList:         tb.Form.FieldList,
			GroupFieldList:    groupFormList,
			GroupFieldHeaders: groupHeaders,
			// tb.Form.Title左上角標題
			Title:             tb.Form.Title,
			// tb.Form.Description標題旁的描述
			Description:       tb.Form.Description,
		}, nil
	}

	// tb.PrimaryKey.Name = id
	// columns = [id parent_id type order title icon uri header created_at updated_at]
	var fieldList = tb.Form.FieldsWithValue(tb.PrimaryKey.Name, id, columns, res, tb.sql)

	return FormInfo{
		FieldList:         fieldList,
		GroupFieldList:    groupFormList,
		GroupFieldHeaders: groupHeaders,
		Title:             tb.Form.Title,
		Description:       tb.Form.Description,
	}, nil
}

// UpdateData update data.
func (tb *DefaultTable) UpdateData(dataList form.Values) error {

	dataList.Add(form.PostTypeKey, "0")

	var (
		errMsg = ""
		err    error
	)

	if tb.Form.PostHook != nil {
		defer func() {
			dataList.Add(form.PostTypeKey, "0")
			dataList.Add(form.PostResultKey, errMsg)
			go func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Error(err)
					}
				}()

				err := tb.Form.PostHook(dataList)
				if err != nil {
					logger.Error(err)
				}
			}()
		}()
	}

	if tb.Form.Validator != nil {
		if err := tb.Form.Validator(dataList); err != nil {
			errMsg = "post error: " + err.Error()
			return err
		}
	}

	if tb.Form.PreProcessFn != nil {
		dataList = tb.Form.PreProcessFn(dataList)
	}

	if tb.Form.UpdateFn != nil {
		dataList.Delete(form.PostTypeKey)
		err = tb.Form.UpdateFn(dataList)
		if err != nil {
			errMsg = "post error: " + err.Error()
		}
		return err
	}

	_, err = tb.sql().Table(tb.Form.Table).
		Where(tb.PrimaryKey.Name, "=", dataList.Get(tb.PrimaryKey.Name)).
		Update(tb.getInjectValueFromFormValue(dataList, types.PostTypeUpdate))

	// NOTE: some errors should be ignored.
	if db.CheckError(err, db.UPDATE) {
		if err != nil {
			errMsg = "post error: " + err.Error()
		}
		return err
	}

	return nil
}

// InsertData insert data.
func (tb *DefaultTable) InsertData(dataList form.Values) error {

	dataList.Add(form.PostTypeKey, "1")

	var (
		id     = int64(0)
		err    error
		errMsg = ""
	)

	if tb.Form.PostHook != nil {
		defer func() {
			dataList.Add(form.PostTypeKey, "1")
			dataList.Add(tb.GetPrimaryKey().Name, strconv.Itoa(int(id)))
			dataList.Add(form.PostResultKey, errMsg)

			go func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Error(err)
					}
				}()

				err := tb.Form.PostHook(dataList)
				if err != nil {
					logger.Error(err)
				}
			}()
		}()
	}

	if tb.Form.Validator != nil {
		if err := tb.Form.Validator(dataList); err != nil {
			errMsg = "post error: " + err.Error()
			return err
		}
	}

	if tb.Form.PreProcessFn != nil {
		dataList = tb.Form.PreProcessFn(dataList)
	}

	if tb.Form.InsertFn != nil {
		dataList.Delete(form.PostTypeKey)
		err = tb.Form.InsertFn(dataList)
		if err != nil {
			errMsg = "post error: " + err.Error()
		}
		return err
	}

	id, err = tb.sql().Table(tb.Form.Table).Insert(tb.getInjectValueFromFormValue(dataList, types.PostTypeCreate))

	// NOTE: some errors should be ignored.
	if db.CheckError(err, db.INSERT) {
		errMsg = "post error: " + err.Error()
		return err
	}

	return nil
}

func (tb *DefaultTable) getInjectValueFromFormValue(dataList form.Values, typ types.PostType) dialect.H {

	var (
		value        = make(dialect.H)
		exceptString = make([]string, 0)

		columns, auto = tb.getColumns(tb.Form.Table)

		fun types.PostFieldFilterFn
	)

	// If a key is a auto increment primary key, it can`t be insert or update.
	if auto {
		exceptString = []string{tb.PrimaryKey.Name, form.PreviousKey, form.MethodKey, form.TokenKey,
			constant.IframeKey, constant.IframeIDKey}
	} else {
		exceptString = []string{form.PreviousKey, form.MethodKey, form.TokenKey,
			constant.IframeKey, constant.IframeIDKey}
	}

	if !dataList.IsSingleUpdatePost() {
		for _, field := range tb.Form.FieldList {
			if field.FormType.IsMultiSelect() {
				if _, ok := dataList[field.Field+"[]"]; !ok {
					dataList[field.Field+"[]"] = []string{""}
				}
			}
		}
	}

	dataList = dataList.RemoveRemark()

	for k, v := range dataList {
		k = strings.Replace(k, "[]", "", -1)
		if !modules.InArray(exceptString, k) {
			if modules.InArray(columns, k) {
				field := tb.Form.FieldList.FindByFieldName(k)
				delimiter := ","
				if field != nil {
					fun = field.PostFilterFn
					delimiter = modules.SetDefault(field.DefaultOptionDelimiter, ",")
				}
				vv := modules.RemoveBlankFromArray(v)
				if fun != nil {
					value[k] = fun(types.PostFieldModel{
						ID:       dataList.Get(tb.PrimaryKey.Name),
						Value:    vv,
						Row:      dataList.ToMap(),
						PostType: typ,
					})
				} else {
					if len(vv) > 1 {
						value[k] = strings.Join(vv, delimiter)
					} else if len(vv) > 0 {
						value[k] = vv[0]
					} else {
						value[k] = ""
					}
				}
			} else {
				field := tb.Form.FieldList.FindByFieldName(k)
				if field != nil && field.PostFilterFn != nil {
					field.PostFilterFn(types.PostFieldModel{
						ID:       dataList.Get(tb.PrimaryKey.Name),
						Value:    modules.RemoveBlankFromArray(v),
						Row:      dataList.ToMap(),
						PostType: typ,
					})
				}
			}
		}
	}
	return value
}

// DeleteData delete data.
func (tb *DefaultTable) DeleteData(id string) error {

	var (
		idArr = strings.Split(id, ",")
		err   error
	)

	if tb.Info.DeleteHook != nil {
		defer func() {
			go func() {
				defer func() {
					if recoverErr := recover(); recoverErr != nil {
						logger.Error(recoverErr)
					}
				}()

				if hookErr := tb.Info.DeleteHook(idArr); hookErr != nil {
					logger.Error(hookErr)
				}
			}()
		}()
	}

	if tb.Info.DeleteHookWithRes != nil {
		defer func() {
			go func() {
				defer func() {
					if recoverErr := recover(); recoverErr != nil {
						logger.Error(recoverErr)
					}
				}()

				if hookErr := tb.Info.DeleteHookWithRes(idArr, err); hookErr != nil {
					logger.Error(hookErr)
				}
			}()
		}()
	}

	if tb.Info.PreDeleteFn != nil {
		if err = tb.Info.PreDeleteFn(idArr); err != nil {
			return err
		}
	}

	if tb.Info.DeleteFn != nil {
		err = tb.Info.DeleteFn(idArr)
		return err
	}

	if len(idArr) == 0 || tb.Info.Table == "" {
		err = errors.New("delete error: wrong parameter")
		return err
	}

	err = tb.delete(tb.Info.Table, tb.PrimaryKey.Name, idArr)
	return err
}

// GetNewForm(取得新表單)判斷條件(TabGroups)後，設置FormInfo(struct)後並回傳
func (tb *DefaultTable) GetNewForm() FormInfo {

	if len(tb.Form.TabGroups) == 0 {
		// 在template\types\form.go
		// FillCustomContent(填寫自定義內容)對FormFields([]FormField)執行迴圈，判斷條件後設置FormField，最後回傳FormFields([]FormField)
		// FieldsWithDefaultValue判斷欄位是否允許添加，例如ID無法手動增加，接著將預設值更新後得到FormField(struct)並加入FormFields中，最後回傳FormFields
		// ----------/menu、/menu/new會執行----------------
		return FormInfo{FieldList: tb.Form.FieldsWithDefaultValue(tb.sql)}
	}
	newForm, headers := tb.Form.GroupField(tb.sql)

	return FormInfo{GroupFieldList: newForm, GroupFieldHeaders: headers}
}

// ***************************************
// helper function for database operation
// ***************************************

func (tb *DefaultTable) delete(table, key string, values []string) error {

	var vals = make([]interface{}, len(values))
	for i, v := range values {
		vals[i] = v
	}

	return tb.sql().Table(table).
		WhereIn(key, vals).
		Delete()
}

func (tb *DefaultTable) getTheadAndFilterForm(params parameter.Parameters, columns Columns) (types.Thead,
	string, string, string, []string, []types.FormField) {

	return tb.Info.FieldList.GetTheadAndFilterForm(types.TableInfo{
		Table:      tb.Info.Table,
		Delimiter:  tb.delimiter(),
		Driver:     tb.connectionDriver,
		PrimaryKey: tb.PrimaryKey.Name,
	}, params, columns, func() *db.SQL {
		return tb.sql()
	})
}

// db is a helper function return raw db connection.
// 透過參數(k)取得匹配的Service(interface)，接著將參數services.Get(tb.connectionDriver)轉換為Connection(interface)回傳並回傳
func (tb *DefaultTable) db() db.Connection {
	if tb.connectionDriver != "" && tb.getDataFromDB() {
		// GetConnectionFromService將參數services.Get(tb.connectionDriver)轉換為Connect(interface)回傳並回傳
		// Get透過參數(k)取得匹配的Service(interface)
		return db.GetConnectionFromService(services.Get(tb.connectionDriver))
	}
	return nil
}

func (tb *DefaultTable) delimiter() string {
	if tb.getDataFromDB() {
		return tb.db().GetDelimiter()
	}
	return ""
}

// getDataFromDB(從資料庫取得資料)判斷條件
func (tb *DefaultTable) getDataFromDB() bool {
	return tb.sourceURL == "" && tb.getDataFun == nil && tb.Info.GetDataFn == nil && tb.Detail.GetDataFn == nil
}

// sql is a helper function return db sql.
// 將參數設置(connName、conn)並回傳sql(struct)
func (tb *DefaultTable) sql() *db.SQL {
	// getDataFromDB(從資料庫取得資料)判斷條件
	if tb.connectionDriver != "" && tb.getDataFromDB() {
		// WithDriverAndConnection將參數設置(connName、conn)並回傳sql(struct)
		return db.WithDriverAndConnection(tb.connection, db.GetConnectionFromService(services.Get(tb.connectionDriver)))
	}
	return nil
}

type Columns []string

// getColumns(取得欄位)將欄位名稱加入columns([]string)
// 如果有值是primary_key並且自動遞增則bool = true，最後回傳欄位名稱及bool
func (tb *DefaultTable) getColumns(table string) (Columns, bool) {

	// sql將參數設置(connName、conn)並回傳sql(struct)
	// Table將SQL(struct)資訊清除後將參數table設置至SQL.TableName回傳
	// ShowColumns取得所有欄位資訊
	columnsModel, _ := tb.sql().Table(table).ShowColumns()

	columns := make(Columns, len(columnsModel))

	// 判斷資料庫引擎類型
	switch tb.connectionDriver {
	// 將欄位名稱加入columns([]string)，如果有值是primary_key並且自動遞增則bool = true，最後回傳欄位名稱及bool
	case db.DriverPostgresql:
		auto := false
		for key, model := range columnsModel {
			columns[key] = model["column_name"].(string)
			if columns[key] == tb.PrimaryKey.Name {
				if v, ok := model["column_default"].(string); ok {
					if strings.Contains(v, "nextval") {
						auto = true
					}
				}
			}
		}
		return columns, auto
	case db.DriverMysql:
		auto := false
		for key, model := range columnsModel {
			columns[key] = model["Field"].(string)
			if columns[key] == tb.PrimaryKey.Name {
				if v, ok := model["Extra"].(string); ok {
					if v == "auto_increment" {
						auto = true
					}
				}
			}
		}
		return columns, auto
	case db.DriverSqlite:
		for key, model := range columnsModel {
			columns[key] = string(model["name"].(string))
		}

		num, _ := tb.sql().Table("sqlite_sequence").
			Where("name", "=", tb.GetForm().Table).Count()

		return columns, num > 0
	case db.DriverMssql:
		for key, model := range columnsModel {
			columns[key] = string(model["column_name"].(string))
		}
		return columns, true
	default:
		panic("wrong driver")
	}
}
