package auth

import (
	"context"
	"fmt"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

const CollectionName = "auth"

var PrefixAuthCacheKey = "cache:auth:"

var _ AuthMongoMapper = (*MongoMapper)(nil)

type (
	AuthMongoMapper interface {
		Insert(ctx context.Context, data *Auth) (string, error)
		FindOne(ctx context.Context, id string) (*Auth, error)
		Update(ctx context.Context, data *Auth) (*mongo.UpdateResult, error)
		UpdateByAuthKeyAndType(ctx context.Context, data *Auth) (*mongo.UpdateResult, error)
		Delete(ctx context.Context, id string) (int64, error)
		FindOneByAuthKeyAndType(ctx context.Context, authKey string, authType int32) (*Auth, error)
		FindOneByUserId(ctx context.Context, userId string) (*Auth, error)
	}
	Auth struct {
		ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
		PassWord string             `bson:"passWord,omitempty" json:"passWord,omitempty"`
		Role     int32              `bson:"role,omitempty" json:"role,omitempty"`
		Type     int32              `bson:"type,omitempty" json:"type,omitempty"`
		Key      string             `bson:"key,omitempty" json:"key,omitempty"`
		UserId   string             `bson:"userId,omitempty" json:"userId,omitempty"`
		UpdateAt time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
		CreateAt time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
	}

	MongoMapper struct {
		conn *monc.Model
	}
)

func (m *MongoMapper) UpdateByAuthKeyAndType(ctx context.Context, data *Auth) (*mongo.UpdateResult, error) {
	data.UpdateAt = time.Now()
	key := PrefixAuthCacheKey + fmt.Sprintf("%s:%d", data.Key, data.Type)
	res, err := m.conn.UpdateOne(ctx, key, bson.M{"key": data.Key, "type": data.Type}, bson.M{"$set": data})
	return res, err
}

func (m *MongoMapper) FindOneByAuthKeyAndType(ctx context.Context, authKey string, authType int32) (*Auth, error) {
	var data Auth
	key := PrefixAuthCacheKey + fmt.Sprintf("%s:%d", authKey, authType)
	err := m.conn.FindOne(ctx, key, &data, bson.M{"key": authKey, "type": authType})
	switch {
	case err == nil:
		return &data, nil
	default:
		return nil, err
	}
}

func (m *MongoMapper) FindOneByUserId(ctx context.Context, userId string) (*Auth, error) {
	var data Auth
	key := PrefixAuthCacheKey + userId
	err := m.conn.FindOne(ctx, key, &data, bson.M{"userId": userId})
	switch {
	case err == nil:
		return &data, nil
	default:
		return nil, err
	}
}

func NewMongoMapper(config *config.Config) AuthMongoMapper {
	conn := monc.MustNewModel(config.Mongo.URL, config.Mongo.DB, CollectionName, config.CacheConf)
	return &MongoMapper{
		conn: conn,
	}
}

func (m *MongoMapper) Insert(ctx context.Context, data *Auth) (string, error) {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}

	key := PrefixAuthCacheKey + data.ID.Hex()
	ID, err := m.conn.InsertOne(ctx, key, data)
	if err != nil {
		return "", err
	}
	return ID.InsertedID.(primitive.ObjectID).Hex(), err
}

func (m *MongoMapper) FindOne(ctx context.Context, id string) (*Auth, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var data Auth
	key := PrefixAuthCacheKey + id
	err = m.conn.FindOne(ctx, key, &data, bson.M{"_id": oid})
	switch {
	case err == nil:
		return &data, nil
	default:
		return nil, err
	}
}

func (m *MongoMapper) Update(ctx context.Context, data *Auth) (*mongo.UpdateResult, error) {
	data.UpdateAt = time.Now()
	key := PrefixAuthCacheKey + data.ID.Hex()
	res, err := m.conn.UpdateOne(ctx, key, bson.M{"_id": data.ID}, bson.M{"$set": data})
	return res, err
}

func (m *MongoMapper) Delete(ctx context.Context, id string) (int64, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}
	key := PrefixAuthCacheKey + id
	res, err := m.conn.DeleteOne(ctx, key, bson.M{"_id": oid})
	return res, err
}

func (m *MongoMapper) GetConn() *monc.Model {
	return m.conn
}
