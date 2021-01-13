package mongodb

import (
	"cloud/conf"
	"cloud/log"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client *mongo.Client
	DB *mongo.Database
	ctx context.Context
	cancelFunc context.CancelFunc
}

var (
	BaseMongo *MongoDB
)

func init() {
	var err error
	BaseMongo, err = NewMongoBD()
	if err != nil {
		log.Panic("err in init mongo, err=%v", err)
	}
	// 建立索引
	index := []mongo.IndexModel{
		{
			Keys:bson.D{{"name",1}},
			Options: options.Index().SetUnique(true),
		},
	}
	opts := options.CreateIndexes().SetCommitQuorumMajority()
	_, err = BaseMongo.DB.Collection("g_role").Indexes().CreateMany(BaseMongo.ctx, index, opts)
	if err != nil {
		log.Panic("error in createIndex, err=%v", err)
	}
}

func NewMongoBD() (*MongoDB, error) {
	ctx, cancel := context.WithCancel(context.Background())

	var err error
	Client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.MongoConfig.Uri))
	//Client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017").SetServerSelectionTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	db := Client.Database(conf.MongoConfig.DB)
	return &MongoDB{
		cancelFunc: cancel,
		Client: Client,
		DB: db,
		ctx: ctx,
	}, nil
}

func (m *MongoDB)StopMongoDB() {
	m.cancelFunc()
	if err := m.Client.Disconnect(m.ctx); err != nil {
		log.Error("mongodb stop err=%v", err)
	}
}

func (m *MongoDB) InsertOne(collection string, document interface{},
	opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error){
	return m.DB.Collection(collection).InsertOne(m.ctx, document, opts...)
}

func (m *MongoDB) InsertMany(collection string, documents []interface{},
	opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	return m.DB.Collection(collection).InsertMany(m.ctx, documents, opts...)
}

func (m *MongoDB) FindOne(collection string, filter interface{},
	opts ...*options.FindOneOptions) *mongo.SingleResult {
	return m.DB.Collection(collection).FindOne(m.ctx, filter, opts...)
}

func (m *MongoDB) Find(collection string, filter interface{},
	opts ...*options.FindOptions) (*mongo.Cursor, error) {
	return m.DB.Collection(collection).Find(m.ctx, filter, opts...)
}

func (m *MongoDB) UpdateOne(collection string, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return m.DB.Collection(collection).UpdateOne(m.ctx, filter, update, opts...)
}

func (m *MongoDB) UpdateMany(collection string, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return m.DB.Collection(collection).UpdateMany(m.ctx, filter, update, opts...)
}

func (m *MongoDB) DeleteOne(collection string, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return m.DB.Collection(collection).DeleteOne(m.ctx, filter, opts...)
}

func (m *MongoDB) DeleteMany(collection string, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return m.DB.Collection(collection).DeleteMany(m.ctx, filter, opts...)
}