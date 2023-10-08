package db

import (
	"context"
	"go-one/common/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"sync"
	"time"
)

var once sync.Once
var sharedClient *mongo.Client
var config *MongoDBConfig

type MongoDBConfig struct {
	URI               string        `yaml:"uri"`
	Database          string        `yaml:"database"`
	Username          string        `yaml:"username"`
	Password          string        `yaml:"password"`
	AuthMechanism     string        `yaml:"auth-mechanism"`
	ServerSelectionTO time.Duration `yaml:"server-selection-timeout"`
	ConnectTimeout    time.Duration `yaml:"connect-timeout"`
	SocketTimeout     time.Duration `yaml:"socket-timeout"`
	MaxPoolSize       uint64        `yaml:"max-pool-size"`
	MinPoolSize       uint64        `yaml:"min-pool-size"`
	ReadConcern       string        `yaml:"read-concern"`
	WriteConcern      int           `yaml:"write-concern"`
	AppName           string        `yaml:"app-name"`
	RetryWrites       bool          `yaml:"retry-writes"`
}

func (c *MongoDBConfig) BuildOptions() *options.ClientOptions {
	clientOptions := options.Client()

	if c.URI != "" {
		clientOptions = clientOptions.ApplyURI(c.URI)
	}

	if c.Username != "" && c.Password != "" {
		clientOptions = clientOptions.SetAuth(options.Credential{
			Username:      c.Username,
			Password:      c.Password,
			AuthMechanism: c.AuthMechanism,
		})
	}

	if c.ServerSelectionTO != 0 {
		clientOptions = clientOptions.SetServerSelectionTimeout(c.ServerSelectionTO)
	}

	if c.ConnectTimeout != 0 {
		clientOptions = clientOptions.SetConnectTimeout(c.ConnectTimeout)
	}

	if c.SocketTimeout != 0 {
		clientOptions = clientOptions.SetSocketTimeout(c.SocketTimeout)
	}

	if c.MinPoolSize != 0 {
		clientOptions = clientOptions.SetMinPoolSize(c.MinPoolSize)

	}

	if c.MaxPoolSize != 0 {
		clientOptions = clientOptions.SetMaxPoolSize(c.MaxPoolSize)
	}

	if c.ReadConcern != "" {
		clientOptions = clientOptions.SetReadConcern(&readconcern.ReadConcern{
			Level: c.ReadConcern,
		})
	}

	if c.WriteConcern > 0 {
		clientOptions = clientOptions.SetWriteConcern(&writeconcern.WriteConcern{
			W: c.WriteConcern,
		})
	}

	if c.AppName != "" {
		clientOptions = clientOptions.SetAppName(c.AppName)
	}

	clientOptions = clientOptions.SetRetryWrites(c.RetryWrites)

	return clientOptions
}

func InitMongo(c *MongoDBConfig) {
	if c == nil || c.URI == "" {
		return
	}

	GetMongoClient()

	log.Infof("MongoDB init success")
}

func GetMongoClient() *mongo.Client {
	once.Do(func() {
		// 创建 MongoDB 客户端，只会执行一次
		clientOptions := config.BuildOptions()
		client, err := mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			log.Fatal(err)
		}

		sharedClient = client
	})

	return sharedClient
}

type MongoMapper struct {
	client     *mongo.Client
	database   string
	collection string
}

func NewMongoMapper(collection string) (*MongoMapper, error) {
	return &MongoMapper{
		client:     GetMongoClient(),
		database:   config.Database,
		collection: collection,
	}, nil
}

func (m *MongoMapper) Close() {
	m.client.Disconnect(context.Background())
}

func (m *MongoMapper) Create(document any) error {
	collection := m.client.Database(m.database).Collection(m.collection)
	_, err := collection.InsertOne(context.Background(), document)
	return err
}

func (m *MongoMapper) Find(filter bson.M, result any) error {
	collection := m.client.Database(m.database).Collection(m.collection)
	return collection.FindOne(context.Background(), filter).Decode(result)
}

func (m *MongoMapper) FindArray(filter bson.M, result any) error {
	collection := m.client.Database(m.database).Collection(m.collection)

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	var results []any
	for cursor.Next(context.Background()) {
		var resultItem any
		err := cursor.Decode(&resultItem)
		if err != nil {
			return err
		}
		results = append(results, resultItem)
	}
	if err := cursor.Err(); err != nil {
		return err
	}

	// 将查找到的结果复制到指定的 result 接口中
	err = bson.Unmarshal(cursor.Current, result)
	if err != nil {
		return err
	}

	return nil
}

// FindPage 分页查询文档
func (m *MongoMapper) FindPage(filter bson.M, pageNum, pageSize int) ([]any, error) {
	collection := m.client.Database(m.database).Collection(m.collection)

	// 计算要跳过的文档数量
	skip := (pageNum) * pageSize

	// 设置分页选项
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(pageSize))

	// 执行查询
	cursor, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []any
	for cursor.Next(context.Background()) {
		var result any // 根据您的文档结构适配
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (m *MongoMapper) Update(filter, update bson.M) error {
	collection := m.client.Database(m.database).Collection(m.collection)
	_, err := collection.UpdateMany(context.Background(), filter, update)
	return err
}

func (m *MongoMapper) UpdateOne(filter, update bson.M) error {
	collection := m.client.Database(m.database).Collection(m.collection)
	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (m *MongoMapper) Delete(filter bson.M) error {
	collection := m.client.Database(m.database).Collection(m.collection)
	_, err := collection.DeleteMany(context.Background(), filter)
	return err
}

func (m *MongoMapper) DeleteOne(filter bson.M) error {
	collection := m.client.Database(m.database).Collection(m.collection)
	_, err := collection.DeleteOne(context.Background(), filter)
	return err
}
