package user

import (
	"context"
	"errors"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/config"
	"github.com/CloudStriver/cloudmind-sts/biz/infrastructure/consts"
	"github.com/zeromicro/go-zero/core/stores/monc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const CollectionName = "user"

var PrefixUserCacheKey = "cache:user:"

var _ IUserMongoMapper = (*MongoMapper)(nil)

type (
	IUserMongoMapper interface {
		Insert(ctx context.Context, data *User) (string, error)                             // 插入
		FindOne(ctx context.Context, id string) (*User, error)                              // 查找
		Update(ctx context.Context, data *User) (*mongo.UpdateResult, error)                // 修改
		UpdateById(ctx context.Context, auth *Auth, id string) (*mongo.UpdateResult, error) // 通过id修改授权信息
		Delete(ctx context.Context, id string) (int64, error)                               // 删除
		FindOneByAuth(ctx context.Context, auth *Auth) (*User, error)                       // 查找某个授权信息
		AppendAuth(ctx context.Context, id string, auth *Auth) error                        // 追加授权信息
	}
	Auth struct {
		Type    int32  `bson:"type" json:"type"`
		AppId   string `bson:"appId" json:"appId"`
		UnionId string `bson:"unionId" json:"unionId"`
	}
	User struct {
		ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
		PassWord string             `bson:"passWord,omitempty" json:"passWord,omitempty"`
		Role     int32              `bson:"role,omitempty" json:"role,omitempty"`
		Auths    []*Auth            `bson:"auths,omitempty" json:"auths,omitempty"`
		UpdateAt time.Time          `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
		CreateAt time.Time          `bson:"createAt,omitempty" json:"createAt,omitempty"`
	}

	MongoMapper struct {
		conn *monc.Model
	}
)

func (m *MongoMapper) AppendAuth(ctx context.Context, id string, auth *Auth) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	key := PrefixUserCacheKey + id
	_, err = m.conn.UpdateOne(ctx, key, bson.M{"_id": ID}, bson.M{"$push": bson.M{"auths": bson.M{"$each": []*Auth{auth}}}})
	if err != nil {
		return err
	}
	return nil
}

func NewMongoMapper(config *config.Config) IUserMongoMapper {
	conn := monc.MustNewModel(config.Mongo.URL, config.Mongo.DB, CollectionName, config.CacheConf)
	return &MongoMapper{
		conn: conn,
	}
}

func (m *MongoMapper) UpdateById(ctx context.Context, auth *Auth, id string) (*mongo.UpdateResult, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	key := PrefixUserCacheKey + id
	update := bson.M{
		"$set": bson.M{
			"auths.$[element]": auth,
		},
	}

	option := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"element.type": auth.Type}},
	})

	res, err := m.conn.UpdateOne(ctx, key, bson.M{"_id": ID}, update, option)
	return res, err
}

func (m *MongoMapper) FindOneByAuth(ctx context.Context, auth *Auth) (*User, error) {
	var data User
	filter := bson.M{
		"auths": bson.M{
			"$elemMatch": bson.M{
				"type":    auth.Type,
				"appId":   auth.AppId,
				"unionId": auth.UnionId,
			},
		},
	}

	err := m.conn.FindOneNoCache(ctx, &data, filter)
	switch {
	case err == nil:
		return &data, nil
	case errors.Is(err, monc.ErrNotFound):
		return nil, consts.ErrNotFound
	default:
		return nil, err
	}
}

func (m *MongoMapper) Insert(ctx context.Context, data *User) (string, error) {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}

	key := PrefixUserCacheKey + data.ID.Hex()
	ID, err := m.conn.InsertOne(ctx, key, data)
	if err != nil {
		return "", err
	}
	return ID.InsertedID.(primitive.ObjectID).Hex(), err
}

func (m *MongoMapper) FindOne(ctx context.Context, id string) (*User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, consts.ErrInvalidObjectId
	}
	var data User
	key := PrefixUserCacheKey + id
	err = m.conn.FindOne(ctx, key, &data, bson.M{"_id": oid})
	switch {
	case err == nil:
		return &data, nil
	case errors.Is(err, monc.ErrNotFound):
		return nil, consts.ErrNotFound
	default:
		return nil, err
	}
}

func (m *MongoMapper) Update(ctx context.Context, data *User) (*mongo.UpdateResult, error) {
	data.UpdateAt = time.Now()
	key := PrefixUserCacheKey + data.ID.Hex()
	res, err := m.conn.UpdateOne(ctx, key, bson.M{"_id": data.ID}, bson.M{"$set": data})
	return res, err
}

func (m *MongoMapper) Delete(ctx context.Context, id string) (int64, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}
	key := PrefixUserCacheKey + id
	res, err := m.conn.DeleteOne(ctx, key, bson.M{"_id": oid})
	return res, err
}
