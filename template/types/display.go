package types

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	"html"
	"html/template"
	"strings"
)

type DisplayFnGenerator interface {
	Get(args ...interface{}) FieldFilterFn
	JS() template.HTML
	HTML() template.HTML
}

type BaseDisplayFnGenerator struct{}

func (base *BaseDisplayFnGenerator) JS() template.HTML   { return "" }
func (base *BaseDisplayFnGenerator) HTML() template.HTML { return "" }

var displayFnGens = make(map[string]DisplayFnGenerator)

func RegisterDisplayFnGenerator(key string, gen DisplayFnGenerator) {
	if _, ok := displayFnGens[key]; ok {
		panic("display function generator has been registered")
	}
	displayFnGens[key] = gen
}

type FieldDisplay struct {
	Display              FieldFilterFn
	DisplayProcessChains DisplayProcessFnChains
}

func (f FieldDisplay) ToDisplay(value FieldModel) interface{} {
	val := f.Display(value)

	if len(f.DisplayProcessChains) > 0 && f.IsNotSelectRes(val) {
		valStr := fmt.Sprintf("%v", val)
		for _, process := range f.DisplayProcessChains {
			valStr = fmt.Sprintf("%v", process(FieldModel{
				Row:   value.Row,
				Value: valStr,
				ID:    value.ID,
			}))
		}
		return valStr
	}

	return val
}

func (f FieldDisplay) IsNotSelectRes(v interface{}) bool {
	switch v.(type) {
	case template.HTML:
		return false
	case []string:
		return false
	case [][]string:
		return false
	default:
		return true
	}
}

func (f FieldDisplay) ToDisplayHTML(value FieldModel) template.HTML {
	v := f.ToDisplay(value)
	if h, ok := v.(template.HTML); ok {
		return h
	} else if s, ok := v.(string); ok {
		return template.HTML(s)
	} else if arr, ok := v.([]string); ok && len(arr) > 0 {
		return template.HTML(arr[0])
	} else if arr, ok := v.([]template.HTML); ok && len(arr) > 0 {
		return arr[0]
	} else if v != nil {
		return template.HTML(fmt.Sprintf("%v", v))
	} else {
		return ""
	}
}

func (f FieldDisplay) ToDisplayString(value FieldModel) string {
	v := f.ToDisplay(value)
	if h, ok := v.(template.HTML); ok {
		return string(h)
	} else if s, ok := v.(string); ok {
		return s
	} else if arr, ok := v.([]string); ok && len(arr) > 0 {
		return arr[0]
	} else if arr, ok := v.([]template.HTML); ok && len(arr) > 0 {
		return string(arr[0])
	} else if v != nil {
		return fmt.Sprintf("%v", v)
	} else {
		return ""
	}
}

func (f FieldDisplay) ToDisplayStringArray(value FieldModel) []string {
	v := f.ToDisplay(value)
	if h, ok := v.(template.HTML); ok {
		return []string{string(h)}
	} else if s, ok := v.(string); ok {
		return []string{s}
	} else if arr, ok := v.([]string); ok && len(arr) > 0 {
		return arr
	} else if arr, ok := v.([]template.HTML); ok && len(arr) > 0 {
		ss := make([]string, len(arr))
		for k, a := range arr {
			ss[k] = string(a)
		}
		return ss
	} else if v != nil {
		return []string{fmt.Sprintf("%v", v)}
	} else {
		return []string{}
	}
}

func (f FieldDisplay) ToDisplayStringArrayArray(value FieldModel) [][]string {
	v := f.ToDisplay(value)
	if h, ok := v.(template.HTML); ok {
		return [][]string{{string(h)}}
	} else if s, ok := v.(string); ok {
		return [][]string{{s}}
	} else if arr, ok := v.([]string); ok && len(arr) > 0 {
		return [][]string{arr}
	} else if arr, ok := v.([][]string); ok && len(arr) > 0 {
		return arr
	} else if arr, ok := v.([]template.HTML); ok && len(arr) > 0 {
		ss := make([]string, len(arr))
		for k, a := range arr {
			ss[k] = string(a)
		}
		return [][]string{ss}
	} else if v != nil {
		return [][]string{{fmt.Sprintf("%v", v)}}
	} else {
		return [][]string{}
	}
}

// 加入func(value string) string至FieldDisplay.DisplayProcessFnChains([]DisplayProcessFn)
// 透過參數limit判斷func(value string)回傳的值
func (f FieldDisplay) AddLimit(limit int) DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(func(value FieldModel) interface{} {
		if limit > len(value.Value) {
			return value
		} else if limit < 0 {
			return ""
		} else {
			return value.Value[:limit]
		}
	})
}

// 加入func(value string) string至FieldDisplay.DisplayProcessFnChains([]DisplayProcessFn)
// func(value string)回傳值為strings.TrimSpace(value)
func (f FieldDisplay) AddTrimSpace() DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(func(value FieldModel) interface{} {
		return strings.TrimSpace(value.Value)
	})
}

func (f FieldDisplay) AddSubstr(start int, end int) DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(func(value FieldModel) interface{} {
		if start > end || start > len(value.Value) || end < 0 {
			return ""
		}
		if start < 0 {
			start = 0
		}
		if end > len(value.Value) {
			end = len(value.Value)
		}
		return value.Value[start:end]
	})
}

// 加入func(value string) string至FieldDisplay.DisplayProcessFnChains([]DisplayProcessFn)
func (f FieldDisplay) AddToTitle() DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(func(value FieldModel) interface{} {
		return strings.Title(value.Value)
	})
}

func (f FieldDisplay) AddToUpper() DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(func(value FieldModel) interface{} {
		return strings.ToUpper(value.Value)
	})
}

func (f FieldDisplay) AddToLower() DisplayProcessFnChains {
	return f.DisplayProcessChains.Add(func(value FieldModel) interface{} {
		return strings.ToLower(value.Value)
	})
}

type DisplayProcessFnChains []FieldFilterFn

func (d DisplayProcessFnChains) Valid() bool {
	return len(d) > 0
}

// 將參數f(func(string) string)加入globalDisplayProcessChains([]DisplayProcessFn)
func (d DisplayProcessFnChains) Add(f FieldFilterFn) DisplayProcessFnChains {
	return append(d, f)
}

func (d DisplayProcessFnChains) Append(f DisplayProcessFnChains) DisplayProcessFnChains {
	return append(d, f...)
}

func (d DisplayProcessFnChains) Copy() DisplayProcessFnChains {
	if len(d) == 0 {
		return make(DisplayProcessFnChains, 0)
	} else {
		var newDisplayProcessFnChains = make(DisplayProcessFnChains, len(d))
		copy(newDisplayProcessFnChains, d)
		return newDisplayProcessFnChains
	}
}

func chooseDisplayProcessChains(internal DisplayProcessFnChains) DisplayProcessFnChains {
	if len(internal) > 0 {
		return internal
	}
	return globalDisplayProcessChains.Copy()
}

// globalDisplayProcessChains類別為[]DisplayProcessFn，DisplayProcessFn類別為func(string) string
var globalDisplayProcessChains = make(DisplayProcessFnChains, 0)

// 將參數f(func(string) string)加入globalDisplayProcessChains([]DisplayProcessFn)
func AddGlobalDisplayProcessFn(f FieldFilterFn) {
	// type DisplayProcessFn func(string) string
	globalDisplayProcessChains = globalDisplayProcessChains.Add(f)
}

// 加入func(value string) string至參數globalDisplayProcessChains([]DisplayProcessFn)
// 透過參數limit判斷func(value string)回傳的值
func AddLimit(limit int) DisplayProcessFnChains {
	return addLimit(limit, globalDisplayProcessChains)
}

// 加入func(value string) string至參數globalDisplayProcessChains([]DisplayProcessFn)
// func(value string)回傳值為strings.TrimSpace(value)
func AddTrimSpace() DisplayProcessFnChains {
	return addTrimSpace(globalDisplayProcessChains)
}

// 加入func(value string) string至參數globalDisplayProcessChains([]DisplayProcessFn)
// 透過參數start、end判斷func(value string)回傳的值
func AddSubstr(start int, end int) DisplayProcessFnChains {
	return addSubstr(start, end, globalDisplayProcessChains)
}

// 加入func(value string) string至globalDisplayProcessChains([]DisplayProcessFn)
// func(value string)回傳值為strings.Title(value)
func AddToTitle() DisplayProcessFnChains {
	return addToTitle(globalDisplayProcessChains)
}

// 加入func(value string) string至globalDisplayProcessChains([]DisplayProcessFn)
// func(value string)回傳值為strings.ToUpper(value)
func AddToUpper() DisplayProcessFnChains {
	return addToUpper(globalDisplayProcessChains)
}

// 加入func(value string) string至globalDisplayProcessChains([]DisplayProcessFn)
// func(value string)回傳值為strings.ToLower(value)
func AddToLower() DisplayProcessFnChains {
	return addToLower(globalDisplayProcessChains)
}

// 加入func(value string) string至globalDisplayProcessChains([]DisplayProcessFn)
// func(value string)回傳值為html.EscapeString(value)
func AddXssFilter() DisplayProcessFnChains {
	return addXssFilter(globalDisplayProcessChains)
}

// 加入func(value string) string至globalDisplayProcessChains([]DisplayProcessFn)
// func(value string)回傳值為replacer.Replace(value)
func AddXssJsFilter() DisplayProcessFnChains {
	return addXssJsFilter(globalDisplayProcessChains)
}

// 加入func(value string) string至參數chains([]DisplayProcessFn)
// 透過參數limit判斷func(value string)回傳的值
func addLimit(limit int, chains DisplayProcessFnChains) DisplayProcessFnChains {
	chains = chains.Add(func(value FieldModel) interface{} {
		if limit > len(value.Value) {
			return value
		} else if limit < 0 {
			return ""
		} else {
			return value.Value[:limit]
		}
	})
	return chains
}

// 加入func(value string) string至參數chains([]DisplayProcessFn)
// func(value string)回傳值為strings.TrimSpace(value)
func addTrimSpace(chains DisplayProcessFnChains) DisplayProcessFnChains {
	chains = chains.Add(func(value FieldModel) interface{} {
		return strings.TrimSpace(value.Value)
	})
	return chains
}

// 加入func(value string) string至參數chains([]DisplayProcessFn)
// 透過參數start、end判斷func(value string)回傳的值
func addSubstr(start int, end int, chains DisplayProcessFnChains) DisplayProcessFnChains {
	chains = chains.Add(func(value FieldModel) interface{} {
		if start > end || start > len(value.Value) || end < 0 {
			return ""
		}
		if start < 0 {
			start = 0
		}
		if end > len(value.Value) {
			end = len(value.Value)
		}
		return value.Value[start:end]
	})
	return chains
}

// 加入func(value string) string至參數chains([]DisplayProcessFn)
// func(value string)回傳值為strings.Title(value)
func addToTitle(chains DisplayProcessFnChains) DisplayProcessFnChains {
	chains = chains.Add(func(value FieldModel) interface{} {
		return strings.Title(value.Value)
	})
	return chains
}

// 加入func(value string) string至參數chains([]DisplayProcessFn)
// func(value string)回傳值為strings.ToUpper(value)
func addToUpper(chains DisplayProcessFnChains) DisplayProcessFnChains {
	chains = chains.Add(func(value FieldModel) interface{} {
		return strings.ToUpper(value.Value)
	})
	return chains
}

// 加入func(value string) string至參數chains([]DisplayProcessFn)
// func(value string)回傳值為strings.ToLower(value)
func addToLower(chains DisplayProcessFnChains) DisplayProcessFnChains {
	chains = chains.Add(func(value FieldModel) interface{} {
		return strings.ToLower(value.Value)
	})
	return chains
}

// 加入func(value string) string至參數chains([]DisplayProcessFn)
// func(value string)回傳值為html.EscapeString(value)
func addXssFilter(chains DisplayProcessFnChains) DisplayProcessFnChains {
	chains = chains.Add(func(value FieldModel) interface{} {
		return html.EscapeString(value.Value)
	})
	return chains
}

// 加入func(value string) string至參數chains([]DisplayProcessFn)
// func(value string)回傳值為replacer.Replace(value)
func addXssJsFilter(chains DisplayProcessFnChains) DisplayProcessFnChains {
	chains = chains.Add(func(value FieldModel) interface{} {
		replacer := strings.NewReplacer("<script>", "&lt;script&gt;", "</script>", "&lt;/script&gt;")
		return replacer.Replace(value.Value)
	})
	return chains
}

func setDefaultDisplayFnOfFormType(f *FormPanel, typ form.Type) {
	if typ.IsMultiFile() {
		f.FieldList[f.curFieldListIndex].Display = func(value FieldModel) interface{} {
			if value.Value == "" {
				return ""
			}
			arr := strings.Split(value.Value, ",")
			res := "["
			for i, item := range arr {
				if i == len(arr)-1 {
					res += "'" + config.GetStore().URL(item) + "']"
				} else {
					res += "'" + config.GetStore().URL(item) + "',"
				}
			}
			return res
		}
	}
	if typ.IsSelect() {
		f.FieldList[f.curFieldListIndex].Display = func(value FieldModel) interface{} {
			return strings.Split(value.Value, ",")
		}
	}
}
