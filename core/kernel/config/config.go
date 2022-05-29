package config

import (
	"github.com/kico0909/cgo/core/kernel/logger"
	"github.com/kico0909/cgo/core/plugins/iniHandler"
	"github.com/kico0909/cgo/core/redis"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
	"time"
)

// 完整配置结构
type ConfigData struct {
	Server  ConfigServerOptions  `yaml:"server"`
	TLS     ConfigTLSOptions     `yaml:"tls"`
	Session ConfigSessionOptions `yaml:"session"`
	Log     ConfigLoggerOptions  `yaml:"log"`
	Mysql   ConfigMysqlOptions   `yaml:"mysql"`
	Redis   ConfgigRedisOptions  `yaml:"redis"`
	MongoDB MongoDBBase          `yaml:"mongodb"` // mongoDB配置信息
	Custom  Custom               `yaml:"custom"`
}

type Custom map[string]interface{}

// Mongo 配置
type MongoDBBase struct {
	Key                bool                    `yaml:"key"`                // 是否开启mongoDB
	Timeout            int64                   `yaml:"timeout"`            // mongo请求超时时间
	DefaultPoolMaxSize int                     `yaml:"defaultPoolMaxSize"` // 默认最大链接池数量
	Child              map[string]*MongoDBInfo `yaml:"child"`
}

// 一个mongoDB的配置
type MongoDBInfo struct {
	AuthMode    string `yaml:"authMode"`    // 认证形式 无,user,ca
	Uri         string `yaml:"uri"`         // mongo的地址 不加协议头
	Database    string `yaml:"database"`    // 数据库名称
	PoolMaxSize int    `yaml:"poolMaxSize"` // 最大链接池数量
	UserName    string `yaml:"userName"`    // 登录用户名
	Passwd      string `yaml:"passwd"`      // 用户登录密码
	CAPath      string `yaml:"CAPath"`      // 证书地址
}

// server 基本配置
type ConfigServerOptions struct {
	Port                 int64         `yaml:"port"`
	StaticRouter         string        `yaml:"staticRouter"`
	StaticPath           string        `yaml:"staticPath"`
	TemplatePath         string        `yaml:"templatePath"`
	ReadTimeout          time.Duration `yaml:"readTimeout"`
	WriteTimeout         time.Duration `yaml:"writeTimeout"`
	MaxHeaderBytes       int           `yaml:"maxHeaderBytes"`
	AllowOtherAjaxOrigin bool          `yaml:"allowOtherAjaxOrigin"`
}

// TLS配置
type ConfigTLSOptions struct {
	Key      bool   `yaml:"key"`
	KeyPath  string `yaml:"keyPath"`
	CertPath string `yaml:"certPath"`
}

// mysql 链接配置
type MysqlSetOpt struct {
	Tag      string `yaml:"tag"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int64  `yaml:"port"`
	Dbname   string `yaml:"dbname"`
	Socket   string `yaml:"socket"`
	Charset  string `yaml:"charset"`
}

// mysql 基础配置
type ConfigMysqlOptions struct {
	Key     bool                    `yaml:"key"`
	MaxOpen int                     `yaml:"maxOpen"`
	MaxIdle int                     `yaml:"maxIdle"`
	Default MysqlSetOpt             `yaml:"default"`
	Write   MysqlSetOpt             `yaml:"write"`
	Read    MysqlSetOpt             `yaml:"read"`
	Childs  map[string]*MysqlSetOpt `yaml:"childs"`
}

// redis 配置
type ConfgigRedisOptions struct {
	Key   bool                 `yaml:"key"`
	Setup redis.RedisSetupInfo `yaml:"setup"`
}

// session 配置信息
type ConfigSessionOptions struct {
	Key             bool                 `yaml:"key"`
	SessionType     string               `yaml:"sessionType"`
	SessionName     string               `yaml:"sessionName"`
	SessionLifeTime int64                `yaml:"sessionLifeTime"`
	Redis           redis.RedisSetupInfo `yaml:"redis"`
}

// 日志配置信息
type ConfigLoggerOptions struct {
	Key        bool   `yaml:"key"`
	Path       string `yaml:"path"`
	FileName   string `yaml:"fileName"`
	StopCutOff bool   `yaml:"stopCutOff"`
}

type ConfigModule struct {
	Conf ConfigData
}

func (_self *ConfigModule) Set(path string) bool {
	b, _ := ioutil.ReadFile(path)
	check, _ := regexp.MatchString(`^\S.*\.yaml$`, path)
	if check {
		err := yaml.Unmarshal(b, &(_self.Conf))
		if err != nil {
			log.Println("<功能初始化> 初始化配置文件失败 ", err)
			return false
		}
	} else {
		err := iniHandler.InitFile(path, &_self.Conf)
		if err != nil {
			log.Println("<功能初始化> 初始化配置文件失败 ", err)
			return false
		}
	}
	return true
}
