package route

import (
	"encoding/json"
	"errors"
	beegoSession "github.com/astaxie/beego/session"
	"github.com/kico0909/cgo/core/route/defaultPages"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
)

const (
	RegExp_Url_Set_String = "{[a-z|A-Z|0-9|_|-|.]*}"
	RegExp_Url_String     = "[a-z|A-Z|0-9|_|.|-]*"
)

var (
	Method_Error = errors.New("路由访问方式错误!")

	RegExp_Url_Set, _ = regexp.Compile(RegExp_Url_Set_String)
	RegExp_Url, _     = regexp.Compile(RegExp_Url_String)

	defaultApiCode = defaultApiCodeType{200, 400}

	default_methods       = []string{"POST", "GET", "PUT", "DELETE"}
	default_method_post   = "POST"
	default_method_get    = "GET"
	default_method_put    = "PUT"
	default_method_delete = "DELETE"

	REG_staticFilesTypes = regexp.MustCompile("\\.[A-Z|a-z|0-9]{2,5}$") // 静态路由文件名后缀的正则
)

// 一条路由的类型
type routerChip struct {
	H             *RouterHandler       // 传入的原生的路由数据
	Vars          map[string]string    // 路由的传值
	Methods       map[string]bool      // 路由可被访问的模式
	FilterFunc    func(*RouterHandler) // 当前路由的拦截器
	IsRouterValue bool                 // 是否是通过路由传值

	isWS bool	// websocket 的执行function
	valueName     []string             // 路由传值的变量名
	path           string               // 原始路由
	regPath        string               // 正则后的路由
	viewRenderFunc func(*RouterHandler) // 路由执行的视图
	wsFunc websocketFunc				// websocket 执行方法
}

// 设置路由的Method
func (_self *routerChip) Method(methods ...string) *routerChip {
	_self.Methods = make(map[string]bool)
	for _, v := range methods {
		_self.Methods[strings.ToUpper(v)] = true
	}
	return _self
}

// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----
type RouterHandler struct {
	W       http.ResponseWriter
	R       *http.Request
	path    string
	Vars    map[string]string
	Session beegoSession.Store
	Values  map[string]interface{}
}

// 路由传值的原型链
func (r *RouterHandler) ShowString(str string) {
	r.W.Write([]byte(str))
}

// 路由传值的原型链
func (r *RouterHandler) ShowByte(b []byte) {
	r.W.Write(b)
}

// 路由直接读取静态文件
// router: 指定的可被访问的路由, static: 映射的静态路径
func (r *RouterHandler) ShowForStaticFile(router, static string) {
	if len(static) < 1 {
		defaultPages.Page_404_get(r.W)
	}
	path := r.R.URL.Path
	if len(REG_staticFilesTypes.FindAllString(path, 1)) < 1 {
		path = path + "index.html"
	}
	path = strings.Replace(path, router, static, 1)
	fn := REG_staticFilesTypes.FindAllString(path, 1)

	switch fn[0] {

	case ".woff", ".eot", ".ttf", ".woff2":
		r.W.Header().Set("Content-Type", "application/font-"+strings.Replace(fn[0], ".", "", 1))
		break
	case ".css":
		r.W.Header().Set("Content-Type", "text/"+strings.Replace(fn[0], ".", "", 1))
		break
	case ".js":
		r.W.Header().Set("Content-Type", "text/javascript")
		break
	case ".jpg", ".png", ".gif", ".bmp":
		r.W.Header().Set("Content-Type", "image/"+strings.Replace(fn[0], ".", "", 1))
		if fn[0] == ".jpg" {
			r.W.Header().Set("Content-Type", "image/jpeg")
		}
		break
	default:
		if len(fn) != 1 {
			r.ShowString("不支持的文件格式!")
			return
		}
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		r.ShowString(err.Error())
		return
	}
	r.ShowByte(b)
}

// 获得json类型的body传值
func (r *RouterHandler) GetBodyValueToJson(res interface{}) error {
	defer r.R.Body.Close()
	b, err := ioutil.ReadAll(r.R.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, res)
}

// 获得form表单信息
func (r *RouterHandler) GetFormData(res interface{}) error {
	r.R.ParseForm()
	d := make(map[string]interface{})
	for k, v := range r.R.Form {
		d[k] = v
	}
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	json.Unmarshal(b, res)
	return nil
}

// 获得form表单 上传信息
func (r *RouterHandler) GetFormFile(valueKey string, res interface{}) (multipart.File, *multipart.FileHeader) {
	formFile, header, err := r.R.FormFile(valueKey)
	if err != nil {
		log.Println("Get form file failed: %s\n", err.Error())
		return nil, nil
	}
	return formFile, header
}

// API形式的json数据渲染页面(用于API的返回)
type showForApiModeType struct {
	Code    interface{} `json:"code"`
	Success bool        `json:"success"`
	Message interface{} `json:"message"`
	Data    interface{} `json:"data"`
}

func (r *RouterHandler) ShowForApiMode(success bool, err, data interface{}, code ...interface{}) {
	var result = showForApiModeType{ Success: success, Code: defaultApiCode.Success, Message: err, Data: data}
	if !result.Success {
		result.Code = defaultApiCode.Fail
		if len(code) > 0 {
			result.Code = code[0]
		}
	}
	strByte, _ := json.Marshal(result)
	r.W.Write(strByte)
}
