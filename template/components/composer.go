package components

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	template2 "github.com/GoAdminGroup/go-admin/template"
)

func ComposeHtml(temList map[string]string, compo interface{}, templateName ...string) template.HTML {
	var text = ""
	
	// fmt.Println(temList["components/box"])
	for _, v := range templateName {
		text += temList["components/"+v]
	}

	tmpl, err := template.New("comp").Funcs(template2.DefaultFuncMap).Parse(text)
	if err != nil {
		panic("ComposeHtml Error:" + err.Error())
	}

	buffer := new(bytes.Buffer)

	defineName := strings.Replace(templateName[0], "table/", "", -1)
	defineName = strings.Replace(defineName, "form/", "", -1)

	err = tmpl.ExecuteTemplate(buffer, defineName, compo)
	if err != nil {
		fmt.Println("ComposeHtml Error:", err)
	}
	return template.HTML(buffer.String())
}
