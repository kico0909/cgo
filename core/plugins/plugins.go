package plugins

import (
	"github.com/kico0909/cgo/core/plugins/WechatMiniProgram"
	"github.com/kico0909/cgo/core/plugins/amap"
	"github.com/kico0909/cgo/core/plugins/qyWeChat"
	"github.com/kico0909/cgo/core/plugins/request"
	"github.com/kico0909/cgo/core/plugins/weichat"
)

type plugins struct {
	//Gaode func(string) amap.AmapType
}

func NewPlugins() *plugins {
	return new(plugins)
}

// 高德地图的golang实现
func (s *plugins) Gaode(key string) *amap.AmapType {
	return amap.NewAmap(key)
}

// 微信的golang实现
func (s *plugins) Wechat(appid, secret, redirectUri string) *weichat.WxAPI {
	return weichat.New(appid, secret, redirectUri)
}

// 企业微信的golang实现
func (s *plugins) QyWechat(CropID, Secret, REDIRECT_URI, STATE, TongXunLu_Secret string, Agentid int) *qyWeChat.QywxApi {
	return qyWeChat.New(CropID, Secret, REDIRECT_URI, STATE, TongXunLu_Secret, Agentid)
}

// 微信小程序
func (s *plugins) WechatMiniprogram(appid, appsecret string) *WechatMiniProgram.WechatMiniProgramType {
	return WechatMiniProgram.NewWechatMiniProgram(appid, appsecret)
}

// request请求的封装
func (s *plugins) Request() *request.Req {
	return request.NewRequest()
}
