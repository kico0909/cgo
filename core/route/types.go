package route

import (
	"github.com/gobwas/ws"
	"net"
	"regexp"
)

// 拦截器结构
type filter struct {
	path      string                     // 原始过滤器路由
	rule      *regexp.Regexp             // 过滤规则
	f         *func(*RouterHandler) bool // 符合规则 执行方法 ; 返回是否阻塞路由的执行
	BlockNext bool                       // 过滤器被执行后是否阻塞过滤器判断, 默认false
}

type WSConn struct {
	conn net.Conn
	op ws.OpCode
	Msg []byte
	Error error
	id string
}

type routerHandlerFunc func(handler *RouterHandler)
type websocketFunc func(*WSConn)

type defaultApiCodeType struct {
	Success int64
	Fail    int64
}