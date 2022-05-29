package weichat

import (
	"github.com/kico0909/cgo/core/kernel/logger"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	Appid        = "wx0d3205d8375caa77"
	Secret       = "2ec272b4c03364e662bae6dd2cf38896"
	REDIRECT_URI = "http://qywx.cbim.org.cn/votes" // 投票调查回调地址
	STATE        = "EgaWeChat"                     // 随机定义子串
)

type WxAPI struct {
	Appid        string
	Secret       string
	REDIRECT_URI string
	STATE        string

	accessToken   string // 当前有效的token
	token_expires int64  // token 失效时间
}

// 创建微信API 实例
func New(Appid, Secret, REDIRECT_URI string) *WxAPI {
	tmp := &WxAPI{
		Appid:        Appid,
		Secret:       Secret,
		REDIRECT_URI: REDIRECT_URI,
		STATE:        STATE}
	token, err := getToken(Appid, Secret)
	if err != nil {
		log.Println("微信初始化失败: ", err)
		return nil
	}
	tmp.accessToken = token.Access_token
	log.Println(tmp.accessToken)
	tmp.token_expires = token.Expires_in + time.Now().Unix()
	return tmp
}

// 发送一个get请求
func get(url string) ([]byte, error) {
	var tmp []byte
	req, err := http.Get(url)
	if err != nil {
		return tmp, err
	}

	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return tmp, err
	}
	return body, nil
}

// 发送一个post请求
func post(url, jsonStr string) ([]byte, error) {
	req, err := http.Post(url, "application/json;charset=utf-8", strings.NewReader(jsonStr))
	var tmp []byte
	if err != nil {
		return tmp, err
	}

	defer req.Body.Close()
	str, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return tmp, err
	}
	return str, nil
}

// 刷新/获得access token
type accessTokenType struct {
	Errcode      int64
	Errmsg       string
	Access_token string
	Expires_in   int64
}

// 获得token
func getToken(AppId, AppSecret string) (accessTokenType, error) {
	var data accessTokenType
	requestUrl := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + AppId + "&secret=" + AppSecret
	str, err := get(requestUrl)
	if err != nil {
		return data, err
	}
	json.Unmarshal([]byte(str), &data)
	return data, nil
}

// 刷新 token
func (this *WxAPI) ResetToken() error {

	// 未超时情况下不刷新token
	if time.Now().Unix() < this.token_expires {
		return nil
	}
	token, err := getToken(this.Appid, this.Secret)
	if err != nil {
		return err
	}
	this.accessToken = token.Access_token
	this.token_expires = token.Expires_in + time.Now().Unix()
	return nil
}

// 重定向用户去登录
func (s *WxAPI) Login(w http.ResponseWriter, r *http.Request, state ...string) {
	log.Println("跳转微信登录=>")
	rurl := url.QueryEscape(s.REDIRECT_URI)

	STATE := strconv.FormatInt(time.Now().UnixNano(), 10)
	if len(state) > 0 {
		STATE = url.QueryEscape(state[0])
	}
	URL := "https://open.weixin.qq.com/connect/oauth2/authorize?appid=" + s.Appid + "&redirect_uri=" + rurl + "&response_type=code&scope=snsapi_userinfo&state=" + STATE + "#wechat_redirect"
	http.Redirect(w, r, URL, http.StatusFound)
	return
}

type openidgetResType struct {
	Access_token  string
	Expires_in    int64
	Refresh_token string
	Openid        string
	Scope         string
}

func (s *WxAPI) OpenID_get(code string) (openidgetResType, error) {
	// https://api.weixin.qq.com/sns/oauth2/access_token?appid=APPID&secret=SECRET&code=CODE&grant_type=authorization_code
	var res openidgetResType
	url := "https://api.weixin.qq.com/sns/oauth2/access_token?appid=" + s.Appid + "&secret=" + s.Secret + "&code=" + code + "&grant_type=authorization_code"
	resb, err := get(url)
	if err != nil {
		return res, err
	}
	if err := json.Unmarshal(resb, &res); err != nil {
		return res, err
	}

	return res, nil
}

type WxUserInfoType struct {
	Openid     string
	Nickname   string
	Sex        int64
	Province   string
	City       string
	Country    string
	Headimgurl string
	Privilege  []string
	Unionid    string
}

func (s *WxAPI) UserInfo_get(accessToken, openid string) (WxUserInfoType, error) {
	var res WxUserInfoType
	url := "https://api.weixin.qq.com/sns/userinfo?access_token=" + accessToken + "&openid=" + openid + "&lang=zh_CN"

	resb, err := get(url)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(resb, &res)
	log.Println("get User info URL => ", url, string(resb), err)
	return res, err
}
