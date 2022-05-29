package mysql

import (
	config "github.com/kico0909/cgo/core/kernel/config"
	"github.com/kico0909/cgo/core/kernel/logger"
	mysql "github.com/kico0909/cgo/core/mysql"
)

func New(conf *config.ConfigMysqlOptions) *mysql.DatabaseMysql {
	return mysql.New(conf, nil, func() {
		log.Println("功能初始化: MYSQL --- [ ok ]")
	})
}
