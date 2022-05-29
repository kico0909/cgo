package session

import (
	"github.com/kico0909/cgo/core/kernel/config"
	"github.com/kico0909/cgo/core/kernel/logger"
	beegoSession "github.com/astaxie/beego/session"
	_ "github.com/astaxie/beego/session/redis"
	"net/http"
	"strconv"
)

// Cgo的session 封装 TODO 把beego的session做了二次封装,是否能自行封装(有没有必要)session? Beego的session 其实很好用
type CgoSession struct {
	manager *beegoSession.Manager
}

var (
	sessionEndName = "_glsessn_"
	sessionSetup   beegoSession.ManagerConfig
)

var sessionManager CgoSession

func initSessionResult(success bool, sessionType string) {
	if success {
		log.Println("功能初始化: SESSION(" + sessionType + ") --- [ ok ]")
	} else {
		log.Fatalln("功能初始化: SESSION(" + sessionType + ") --- [ ok ]")
	}
}

// 新建
func (_self *CgoSession) New(conf *config.ConfigSessionOptions) *CgoSession {

	var err error

	// 配置信息检测容错设置默认值
	if conf.SessionType == "" {
		conf.SessionType = "memory"
	}

	if conf.SessionName == "" {
		conf.SessionName = "_Cgo"
	}

	if conf.SessionLifeTime == 0 {
		conf.SessionLifeTime = 3600
	}

	sessionSetup.CookieName = conf.SessionName + sessionEndName
	sessionSetup.Gclifetime = conf.SessionLifeTime
	sessionSetup.EnableSetCookie = true

	// 初始化 session
	switch conf.SessionType {

	case "redis":
		srHost := conf.Redis.Host
		srPort := strconv.FormatInt(conf.Redis.Port, 10)
		srNumber := strconv.FormatInt(conf.Redis.Dbname, 10)
		srPassword := conf.Redis.Password
		sessionSetup.ProviderConfig = srHost + `:` + srPort + `,` + srNumber + `,` + srPassword
		break

	default:

	}

	sessionManager.manager, err = beegoSession.NewManager(conf.SessionType, &sessionSetup)

	if err != nil {
		initSessionResult(false, conf.SessionType)
		log.Println(333, err)
	}

	go sessionManager.manager.GC()

	initSessionResult(true, conf.SessionType)

	return &sessionManager
}

// 启动session
func (_self *CgoSession) SessionStart(w http.ResponseWriter, r *http.Request) (beegoSession.Store, error) {
	return _self.manager.SessionStart(w, r)
}

// 根据id 获得
func (_self *CgoSession) GetSessionStore(sid string) (beegoSession.Store, error) {
	return _self.manager.GetSessionStore(sid)
}

// 销毁全部
func (_self *CgoSession) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	_self.manager.SessionDestroy(w, r)
}

func (_self *CgoSession) SessionRegenerateID(w http.ResponseWriter, r *http.Request) beegoSession.Store {
	return _self.manager.SessionRegenerateID(w, r)
}

func (_self *CgoSession) GetActiveSession() int {
	return _self.manager.GetActiveSession()
}

func (_self *CgoSession) SetSecure(secure bool) {
	_self.manager.SetSecure(secure)
}
