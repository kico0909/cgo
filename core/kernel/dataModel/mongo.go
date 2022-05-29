package dataModel

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type Mongo struct {
	conn *mongo.Database `json:"conn"`
	collectionName string `json:"collection"`
}

func InitMongoCollectionModule () (func(*mongo.Database, string) *Mongo) {
	return func(db *mongo.Database, collectionName string) *Mongo {
		m := Mongo{
		conn: db,
		collectionName: collectionName}
		return &m
	}
}

// 插入单条
func (s *Mongo) InsertOne (d interface{}, opt ...*options.CollectionOptions) (*mongo.InsertOneResult, error) {
	return s.conn.Collection(s.collectionName, opt...).InsertOne(context.TODO(), d)
}
// 插入多条
func (s *Mongo) InsertMany (ds []interface{}, opt ...*options.CollectionOptions) (*mongo.InsertManyResult, error) {
	return s.conn.Collection(s.collectionName, opt...).InsertMany(context.TODO(), ds)
}
// 查找单条
func (s *Mongo) FindOne(filterForBsonStr primitive.D,decodeData interface{}) {
	s.conn.Collection(s.collectionName).FindOne(context.TODO(),filterForBsonStr).Decode(decodeData)
}

// 查找多条
func (s *Mongo) Find (filter primitive.D, options *options.FindOptions, result interface{}) (err error) {
	cur, err := s.conn.Collection(s.collectionName).Find(context.TODO(), filter, options)
	if err != nil {
		return  err
	}
	defer cur.Close(context.TODO())
	var res []interface{}
	for cur.Next(context.TODO()) {
		// 创建一个值，将单个文档解码为该值
		tmp := make(map[string]interface{})
		err := cur.Decode(&tmp)
		if err != nil {
			return err
		}
		res = append(res, tmp)
	}
	b, _ := json.Marshal(res)
	json.Unmarshal(b, result)
	return err
}

// 更新 单条
func (s *Mongo) UpdateOne(filter primitive.D, updateData primitive.D) (*mongo.UpdateResult, error) {
	ud := bson.D{{"$set", updateData}}
	return s.conn.Collection(s.collectionName).UpdateOne(context.TODO(),filter, ud)
}

// 批量更新
func (s *Mongo) UpdateMany(filter primitive.D, updateData primitive.D) (*mongo.UpdateResult, error) {
	log.Println(updateData.Map())
	ud := bson.D{{"$set", updateData}}
	return s.conn.Collection(s.collectionName).UpdateMany( context.TODO(), filter, ud, nil)
}

// 删除一条数据
func (s *Mongo) DeleteOne(filter primitive.D) (*mongo.DeleteResult, error) {
	return s.conn.Collection(s.collectionName).DeleteOne(context.TODO(), filter)
}

// 批量删除数据
func (s *Mongo) DeleteMany(filter primitive.D)(*mongo.DeleteResult, error) {
	return s.conn.Collection(s.collectionName).DeleteMany(context.TODO(),filter)
}