package app

/*
服务器的内核 用于启动服务器
*/

import (
	"net/http"
	"strconv"
	"time"

	"github.com/kico0909/cgo/core/kernel/config"
	"github.com/kico0909/cgo/core/kernel/logger"
	"github.com/kico0909/cgo/core/route"
	//"github.com/golang/crypto/acme/autocert"
	//"github.com/golang/net/http2"
)

type RouterType interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func ServerStart(router *route.RouterManager, conf *config.ConfigData) {
	// 不启用HTTPS
	if !conf.TLS.Key {
		normalServerStart(router, conf)
	}
	// 正常https证书使用
	httpsNormalServerStart(router, conf)
}

// 非HTTPS服务器
func normalServerStart(router *route.RouterManager, conf *config.ConfigData) {

	server := &http.Server{
		// 地址及端口号
		Addr: `:` + strconv.FormatInt(conf.Server.Port, 10),

		// 读取超时时间
		ReadTimeout: conf.Server.ReadTimeout * time.Second,

		// 写入超时时间
		WriteTimeout: conf.Server.WriteTimeout * time.Second,

		// 头字节限制
		MaxHeaderBytes: conf.Server.MaxHeaderBytes * 1024,

		// 路由
		Handler: router,
	}

	log.Println("服务器启动完成: (监听端口:" + strconv.FormatInt(conf.Server.Port, 10) + ") --- [ ok ]\n\n")

	log.Fatalln(server.ListenAndServe())

}

// 启动https服务器,需要填写证书路径
func httpsNormalServerStart(router *route.RouterManager, conf *config.ConfigData) {
	// 启用 HTTPS 直接加载证书
	server := &http.Server{

		// 地址及端口号
		Addr: `:` + strconv.FormatInt(conf.Server.Port, 10),

		// 读取超时时间
		ReadTimeout: conf.Server.ReadTimeout * time.Second,

		// 写入超时时间
		WriteTimeout: conf.Server.WriteTimeout * time.Second,

		// 头字节限制
		MaxHeaderBytes: conf.Server.MaxHeaderBytes * 1024,
		// 路由
		Handler: router,
	}

	log.Println("服务器启动完成:(https:" + strconv.FormatInt(conf.Server.Port, 10) + ") --- [ ok ]")

	log.Fatalln(server.ListenAndServeTLS(conf.TLS.CertPath, conf.TLS.KeyPath))
}
