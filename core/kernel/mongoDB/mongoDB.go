package mongoDB

import (
	"context"
	"github.com/kico0909/cgo/core/kernel/config"
	log "github.com/kico0909/cgo/core/kernel/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"net/url"
	"os"
	"time"
)

func NewMongoDB (option *config.MongoDBBase) (mongos map[string]*mongo.Database) {
	mongos = make(map[string]*mongo.Database)
	var err error
	for k,v := range option.Child {
		mongos[k], err = connectToDB(option, v)
		if err != nil {
			log.Println("功能初始化: MongoDB	 --- [ fail ]")
			log.Println("MongoDB[ " + k + " ]连接池,初始化错误: ", err.Error())
			os.Exit(9)
		}
	}
	log.Println("功能初始化: MongoDB	 --- [ ok ]")
	return mongos
}

func makeMongoDBClientOption(main *config.MongoDBBase, sub *config.MongoDBInfo) (*options.ClientOptions) {
	uri := ""
	switch sub.AuthMode {
	case "user":
		uri = "mongodb://" + sub.UserName + ":" + url.QueryEscape(sub.Passwd) + "@" + sub.Uri
		break
	default:
		uri = "mongodb://" + sub.Uri
	}
	o := options.Client().ApplyURI(uri)
	num := sub.PoolMaxSize
	if num == 0 {
		num = main.DefaultPoolMaxSize
	}
	o.SetMaxPoolSize(uint64(num))
	return o
}

func connectToDB(main *config.MongoDBBase, sub *config.MongoDBInfo) ( *mongo.Database, error) {
	timeout :=  time.Duration(main.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client, err := mongo.Connect(ctx, makeMongoDBClientOption(main, sub))
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary());
	if err != nil {
		return nil, err
	}
	return client.Database(sub.Database), nil
}
