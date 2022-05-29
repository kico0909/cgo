package command

import (
	cgoApp "github.com/kico0909/cgo/core/kernel/app"
	"github.com/kico0909/cgo/core/kernel/config"
	"github.com/kico0909/cgo/core/route"
)

// 服务器参数处理
func Run(comm *string, router *route.RouterManager, conf *config.ConfigData) {
	serverStart(router, conf)
}

// 服务器初始化与启动
func serverStart(router *route.RouterManager, conf *config.ConfigData) {
	// 启动服务器
	cgoApp.ServerStart(router, conf)
}

func init() {

}
