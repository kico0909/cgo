package cas

import (
	"github.com/kico0909/cgo"
	"github.com/kico0909/cgo/core/kernel/session"
	"github.com/kico0909/cgo/core/route"
)

type casSessionKeyType struct {
	Sid string `json:"sid"`
	Tgt string `json:"tgt"`
}
var casSessionName = "cas_router_box"

var casAndRouter map[string]string

type casHandler struct {
	sm *session.CgoSession
}

func NewCas(casServerUrl string) *casHandler {
	return &casHandler{
		sm: route.GetSessionManager()}
}

func CreateCasFilter(SessionType interface{}) (func(*cgo.RouterHandler) bool) {
	//ss := route.GetSessionManager()

	// 返回一个拦截器
	return func(handler *cgo.RouterHandler)bool {

		return false
	}
}