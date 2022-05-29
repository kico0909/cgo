package template

import (
	"github.com/kico0909/cgo/core/kernel/config"
	"github.com/kico0909/cgo/core/kernel/logger"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
)

type CgoTemplateType struct {
	templateObj  *template.Template
	templatePath string
}

func New(conf *config.ConfigData) *CgoTemplateType {

	log.Println("功能初始化: 缓存模板(" + conf.Server.TemplatePath + ") --- [ ok ]")
	tmp := &CgoTemplateType{templatePath: conf.Server.TemplatePath}
	tmp.CacheHtmlTemplate(conf.Server.TemplatePath)
	return tmp
}

// 缓存所有 html 模板
func (_self *CgoTemplateType) CacheHtmlTemplate(templatePath string) {
	var err error
	_self.templateObj, err = template.ParseGlob(templatePath + "/*.html")
	if err != nil {
		log.Println("模板缓存异常: 没有模板文件可被缓存!")
	}

}

// 渲染 基于缓存的模板
func (_self *CgoTemplateType) RenderHtml(w http.ResponseWriter, templateName string, data interface{}) {
	err := _self.templateObj.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Println("模板创建页面错误----->", templateName, data)
		log.Println(err)
	}
}

// 基于文件的模板渲染
func (_self *CgoTemplateType) RenderFile(w http.ResponseWriter, prefectPath string) {
	fileCont, _ := ioutil.ReadFile(prefectPath)
	fmt.Fprintf(w, string(fileCont))
}
