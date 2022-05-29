package route

import (
	"github.com/kico0909/cgo/core/kernel/config"
	"github.com/kico0909/cgo/core/kernel/logger"
	session "github.com/kico0909/cgo/core/kernel/session"
	"github.com/kico0909/cgo/core/route/defaultPages"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	BEFORE_ROUTER = "beforeRouter"
	AFTER_RENDER  = "afterRender"
)

// 配置出session
var sess *session.CgoSession

func SetSession(s *session.CgoSession) {
	sess = s
}
func GetSessionManager() *session.CgoSession {
	return sess
}

// 获得全局的config
var conf config.ConfigData

func SetConfig(c config.ConfigData) {
	conf = c
}

// 路由管理员
type RouterManager struct {
	Routers []*routerChip // 所有的路由
	filter  struct {      // 过滤器
		beforeRoute []*filter // 匹配路由之前进行拦截
		afterRender []*filter // 渲染页面之后执行的拦截
	}
	httpStatus struct {
		notFound         func(http.ResponseWriter)
		notAllowedMethod func(http.ResponseWriter)
	}

	staticRouter string
	staticPath   string
}

// 接口实现方法
func (_self *RouterManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// 禁止方法继续执行
	var stopKey bool

	routerHandlerValue := &RouterHandler{
		W:       w,
		R:       r,
		path:    "",
		Vars:    make(map[string]string),
		Session: nil,
		Values:  make(map[string]interface{})}

	// Session 当前全局
	if conf.Session.Key {
		routerHandlerValue.Session, _ = sess.SessionStart(w, r)
	}

	// 检测路由前的拦截器
	for _, v := range _self.filter.beforeRoute {
		if v.rule.MatchString(r.URL.Path) { // 正则匹配
			tmp := (*(v.f))(routerHandlerValue)
			if !stopKey {
				stopKey = tmp
			}
			if v.BlockNext {
				break
			}
		}
	}

	// 拦截器是否阻塞执行
	if stopKey {
		return
	}

	for _, v := range _self.Routers {

		// 1. 匹配路由
		if checkRouter(r.URL.Path, v.regPath) {
			// 2. 匹配Method
			if v.Methods == nil || v.Methods[r.Method] {

				// 基础数据的赋值
				routerHandlerValue.path = v.path
				routerHandlerValue.Vars = _self.getRouterValue(r.URL.Path, v)
				v.H = routerHandlerValue

				// 3. 执行业务视图渲染
				if v.isWS {
					v.WsRun()
				} else {
					v.viewRenderFunc(v.H)
				}

				// 4. session 保存
				if v.H.Session != nil {
					v.H.Session.SessionRelease(v.H.W)
				}

				// 路由执行渲染后的拦截器
				for _, vv := range _self.filter.afterRender {
					if vv.rule.MatchString(r.URL.Path) { // 正则匹配
						(*(vv.f))(v.H)
						if vv.BlockNext {
							break
						}
					}
				}

			} else {

				// 访问模式拒绝页面
				switch strings.ToLower(r.Method) {

				case "get":
					defaultPages.Page_405_get(w)
					break

				case "post", "put", "delete":
					defaultPages.Page_405_post(w)
					break
				}

			}
			return
		}
	}
	// 没有路由情况下跳404
	// 访问模式拒绝页面
	switch strings.ToLower(r.Method) {
	case "get":
		defaultPages.Page_404_get(w)
		break
	case "post":
		defaultPages.Page_404_post(w)
		break
	}
}

// 注册一条新路由
func (_self *RouterManager) Register(path string, f routerHandlerFunc) *routerChip {
	if path == _self.staticRouter {
		log.Fatalln("路由地址与静态文件路由地址冲突!\n[", path, "==>", _self.staticPath, "]")
	}
	return _self.addRouter(path, f)
}

// 设置一条路由重定向到一条已有路由上
func (_self *RouterManager) Redirect(path, repath string) {
	if path == _self.staticRouter {
		log.Fatalln("重定向地址与静态文件路由地址冲突!\n[", path, "==>", _self.staticPath, "]")
	}

	for _, v := range _self.Routers {
		if repath == v.path {
			_self.addRouter(path, v.viewRenderFunc).Methods = v.Methods
			return
		}
	}

	log.Fatalln("重定向路由地址未被注册过")
}

// 针对路由的拦截器
// 参数: 拦截器位置, 拦截的路由, 拦截器执行方法(需要返回Bool 是否拦截), 被拦截后是否继续执行拦截器
func (_self *RouterManager) InsertFilter(position string, pathRule string, f func(handler *RouterHandler) bool, BlockNext ...bool) {
	re, err := regexp.Compile(handlerPathString2regexp(pathRule))
	if err != nil {
		re, _ = regexp.Compile(`[\D|\d]*`)
	}
	filterStruct := &filter{pathRule, re, &f, len(BlockNext) > 0 && BlockNext[0] == true}
	switch position {

	case BEFORE_ROUTER:
		_self.filter.beforeRoute = append(_self.filter.beforeRoute, filterStruct)
		break

	case AFTER_RENDER:
		_self.filter.afterRender = append(_self.filter.afterRender, filterStruct)
		break

	}

}

// 设置静态文件访问目录
func (_self *RouterManager) SetStaticPath(router string, path string) {
	// 检测是否有重复路由
	for _, v := range _self.Routers {
		if v.path == router {
			log.Fatalln("设置的静态文件地址路由路径冲突!")
		}
	}
	_self.staticRouter = router
	_self.staticPath = path

	// 注册一个静态文件的路由
	_self.addRouter(router+"**", _self.makeFileServe(http.StripPrefix(router, http.FileServer(http.Dir(_self.staticPath)))))
}

// 设置默认API 返回 success, fail 状态的返回code码
func (_self *RouterManager) SetDefaultApiCode(success, fail int64) {
	defaultApiCode.Success = success
	defaultApiCode.Fail = fail
}

// session manager 的获得
func (s *RouterManager) GetSessionManager() *session.CgoSession {
	return sess
}

// 注册一条新路由
func (_self *RouterManager) addRouter(path string, f routerHandlerFunc) *routerChip {
	valuename := _self.getRouterValueName(path)
	tmp := &routerChip{
		Vars:           make(map[string]string),
		path:           path,
		regPath:        handlerPathString2regexp(path),
		viewRenderFunc: f,
		IsRouterValue:  len(valuename) > 0,
		H:              &RouterHandler{path: path},
		valueName:      valuename}
	_self.Routers = append(_self.Routers, tmp)
	return tmp
}

// 解析传值路由的变量名
func (_self *RouterManager) getRouterValueName(path string) []string {

	replaceStr, _ := regexp.Compile("[{|}]")

	r := RegExp_Url_Set.FindAllString(path, 100)
	for i := range r {
		r[i] = replaceStr.ReplaceAllString(r[i], "")
	}
	return r
}

// 获得路由解析传值
func (_self *RouterManager) getRouterValue(url string, rh *routerChip) map[string]string {

	res := make(map[string]string)
	// 不是路由传值
	if !rh.IsRouterValue {
		return res
	}

	reg, _ := regexp.Compile("[{|}]")

	routeSet := strings.Split(rh.path, "/")
	urlSet := strings.Split(url, "/")

	for k, v := range routeSet {
		// 确认是路由上设置的变量
		if RegExp_Url_Set.MatchString(v) {
			res[reg.ReplaceAllString(v, "")] = urlSet[k]
		}
	}
	return res
}

// 文件路由的处理方法
func (_self *RouterManager) makeFileServe(handler http.Handler) routerHandlerFunc {
	return func(h *RouterHandler) {
		handler.ServeHTTP(h.W, h.R)
	}
}

// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----
// 创建新的路由
func NewRouter() *RouterManager {
	return &RouterManager{}
}

// ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ---- ----

// 把一个路由设置的URL转换成用于判断URL的正则
func handlerPath(path string) (string, bool) {
	// 创建一个正则
	return handlerPathString2regexp(path), !(RegExp_Url_Set.FindIndex([]byte(path)) == nil)
}

// 路由匹配
func checkRouter(url, path string) bool {
	re, err := regexp.Compile(path)
	if err != nil {
		return false
	}
	return re.MatchString(url)
}

// 解析路由路径为正则字符串
func handlerPathString2regexp(path string) string {
	temp_time := strconv.FormatInt(time.Now().Unix(), 10)

	// 双星
	doubleStartReg, _ := regexp.Compile(`\*\*`)

	// 单星
	startReg, _ := regexp.Compile(`\*`)

	// 双星转换的中间量
	swapReg, _ := regexp.Compile(temp_time)

	// 替换双星
	path = doubleStartReg.ReplaceAllString(path, `\S`+temp_time)

	// 替换单星
	path = "^" + startReg.ReplaceAllString(path, RegExp_Url_String) + "$"

	// 替换中间变量
	path = swapReg.ReplaceAllString(path, `*`)

	// 替换路由传值变量值
	path = RegExp_Url_Set.ReplaceAllString(path, RegExp_Url_String)

	return path
}
