package WechatMiniProgram

import (
	log "github.com/kico0909/cgo/core/kernel/logger"
	"github.com/kico0909/cgo/core/plugins/request"
	"encoding/json"
	"errors"
	"time"
)

const (
	errorCode_busy     = -1
	errorCode_success  = 0
	errorCode_appinfoFail = 40001	// AppSecret 错误或者 AppSecret 不属于这个小程序，请开发者确认 AppSecret 的正确性
	errorCode_granttypeError = 40002	// 请确保 grant_type 字段值为 client_credential
	errorCode_appidError = 40013	// 不合法的 AppID，请开发者检查 AppID 的正确性，避免异常字符，注意大小写
	errorCode_codeFail = 40029
	errorCode_outof    = 45011
)

type WechatMiniProgramType struct {
	appid     string `json:"appid"`
	appsecret string `json:"appsecret"`
	token     string `json:"token"`
	expires   int64  `json:"expires"`
}

func NewWechatMiniProgram(appid, appsecret string) *WechatMiniProgramType {
	rs := new(WechatMiniProgramType)
	rs.appid = appid
	rs.appsecret = appsecret
	rs.expires = time.Now().Unix()
	return rs
}

// 获得用户信息
type Code2SessionReturnType struct {
	Openid      string `json:"openid"`
	Session_key string `json:"session_key"`
	Unionid     string `json:"unionid"`
	Errcode     int64  `json:"errcode"`
	Errmsg      string `json:"errmsg"`
}

func (s *WechatMiniProgramType) Code2Session(code string) (rs Code2SessionReturnType, err error) {
	var req = request.NewRequest()
	url := "https://api.weixin.qq.com/sns/jscode2session?appid=" + s.appid + "&secret=" + s.appsecret + "&js_code=" + code + "&grant_type=authorization_code"
	_, err = req.Get(url, nil, nil)
	if err != nil {
		return rs, err
	}
	b, _ := req.GetBody()
	json.Unmarshal(b, &rs)
	if rs.Errcode == errorCode_success { // 请求成功
		return rs, nil
	}
	return rs, errors.New(rs.Errmsg)
}

func (s *WechatMiniProgramType) checkToken() {
	if s.expires >= time.Now().Unix() {
		s.updateAccessToken()
	}
}

// 获得token
type tokenResult struct {
	Access_token string `json:"access_token"`
	Expires_in int64 `json:"expires_in"`
	Errcode int64 `json:"errcode"`
	Errmsg string `json:"errmsg"`
}
func (s *WechatMiniProgramType) updateAccessToken() {
	var req = request.NewRequest()
	url := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid="+s.appid+"&secret=" + s.appsecret
	req.Get(url, nil, nil)
	var rs tokenResult
	b, err := req.GetBody()
	if err != nil {
		log.Println("微信小程序获得Token请求失败")
		return
	}
	json.Unmarshal(b, &rs)
	if rs.Errcode == errorCode_success {
		s.token = rs.Access_token
		s.expires = time.Now().Unix() + rs.Expires_in
		return
	}
	log.Println("微信小程序获得Token请求错误", rs.Errcode, rs.Errmsg)
}
