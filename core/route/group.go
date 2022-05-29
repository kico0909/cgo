package route

import (
	"github.com/kico0909/cgo/core/kernel/logger"
	"strings"
)

/*

以组的形式注册路由

1. 组可以无限嵌套
2. 拦截器处理 由 拦截器单独处理, 组内不处理拦截器问题
3. 所谓的路由组可以简单的认为是快捷的批量的 组合路由进行 逐条注册

*/
// 路由组 每一个Chip的返回值 方便上一级的处理
type routerGroupFinilReturn struct {
	Path        string
	HandlerFunc routerHandlerFunc
	Methods     string
}

// 设置路由组, 支持多级分组, 暂时method 只支持 get,post 两种方式
func (c *RouterManager) Group(path string, params ...[]routerGroupFinilReturn) {
	if len(params) < 1 {
		log.Println("Router Group Set Error")
		return
	}
	for _, v := range params {
		for _, vv := range v {
			c.Register(path+vv.Path, vv.HandlerFunc).Method(strings.Split(vv.Methods, ",")...)
		}
	}
}
func (c *RouterManager) GPChip(path string, params ...interface{}) []routerGroupFinilReturn {
	return groupChipHandler(path, params)
}
func (c *RouterManager) GPChipForGet(path string, params ...interface{}) []routerGroupFinilReturn {
	return groupChipHandler(path, params, default_method_get)
}
func (c *RouterManager) GPChipForPost(path string, params ...interface{}) []routerGroupFinilReturn {
	return groupChipHandler(path, params, default_method_post)
}
func (c *RouterManager) GPChipForPut(path string, params ...interface{}) []routerGroupFinilReturn {
	return groupChipHandler(path, params, default_method_put)
}
func (c *RouterManager) GPChipForDelete(path string, params ...interface{}) []routerGroupFinilReturn {
	return groupChipHandler(path, params, default_method_delete)
}
func (c *RouterManager) GPChipForAny(path string, params ...interface{}) []routerGroupFinilReturn {
	return groupChipHandler(path, params, default_methods...)
}
func groupChipHandler(path string, routerGroup []interface{}, methods ...string) []routerGroupFinilReturn {
	var result []routerGroupFinilReturn
	for _, v := range routerGroup {
		if pArr, ok := v.([]routerGroupFinilReturn); ok { // 组方法返回情况
			for _, vv := range pArr {
				var p routerGroupFinilReturn
				p = vv
				p.Path = path + p.Path
				result = append(result, p)
			}
		} else {
			if len(methods) < 1 {
				log.Println("GroupChip use error")
				return result
			}
			var p routerGroupFinilReturn
			p.Path = path
			p.HandlerFunc = v.(func(*RouterHandler))
			p.Methods = strings.Join(methods, ",")
			result = append(result, p)
		}
	}

	return result
}
