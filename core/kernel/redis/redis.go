package redis

import (
	reids "github.com/kico0909/cgo/core/redis"
	"github.com/kico0909/cgo/core/kernel/config"
	"github.com/kico0909/cgo/core/kernel/logger"
	"encoding/json"
)

func New(conf *config.ConfgigRedisOptions) *reids.DatabaseRedis {
	log.Println("功能初始化: Redis	 --- [ ok ]")

	var cgoRedis reids.DatabaseRedis
	strByte, _ := json.Marshal(conf.Setup)
	var redisSetupInfo reids.RedisSetupInfo
	json.Unmarshal(strByte, &redisSetupInfo)
	cgoRedis.Init(&redisSetupInfo)
	return &cgoRedis
}
