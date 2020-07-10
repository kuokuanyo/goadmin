package components

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	template2 "github.com/GoAdminGroup/go-admin/template"
)

// 首先透過參數(templateName ...string)將符合temList(map[string]string)的值加入text(string)，接著將參數及功能添加給新的模板並解析為模板的主體
// 將參數compo寫入buffer(bytes.Buffer)中最後輸出HTML
func ComposeHtml(temList map[string]string, compo interface{}, templateName ...string) template.HTML {

	// 將map[string]string符合key的值加入text
	var text = ""
	
	// fmt.Println(temList["components/box"])
	for _, v := range templateName {
		text += temList["components/"+v]
	}

	// new將給定的參數分配給新的HTML模板
	// Funcs添加新的功能到模板
	// Parse將參數text解析為模板的主體
	tmpl, err := template.New("comp").Funcs(template2.DefaultFuncMap).Parse(text)
	if err != nil {
		panic("ComposeHtml Error:" + err.Error())
	}

	buffer := new(bytes.Buffer)

	defineName := strings.Replace(templateName[0], "table/", "", -1)
	defineName = strings.Replace(defineName, "form/", "", -1)

	// 與給定defineName的模板應用，將第三個參數(compo)寫入buffer中
	err = tmpl.ExecuteTemplate(buffer, defineName, compo)
	if err != nil {
		fmt.Println("ComposeHtml Error:", err)
	}
	
	// 將buffer輸出成HTML
	return template.HTML(buffer.String())
}
