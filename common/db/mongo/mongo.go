package mongo

import (
	"context"
	"sync"
	"time"

	"github.com/ambitiousmice/go-one/common/log"
	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/bson"
	mgoptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var once sync.Once
var sharedClient *qmgo.Client
var config *Config

type Config struct {
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

func (c *Config) BuildConfig() *mgoptions.ClientOptions {
	clientConfig := mgoptions.Client()

	if c.URI != "" {
		clientConfig = clientConfig.ApplyURI(c.URI)
	}

	if c.Username != "" && c.Password != "" {
		clientConfig = clientConfig.SetAuth(mgoptions.Credential{
			Username:      c.Username,
			Password:      c.Password,
			AuthMechanism: c.AuthMechanism,
		})
	}

	if c.ServerSelectionTO != 0 {
		clientConfig = clientConfig.SetServerSelectionTimeout(c.ServerSelectionTO)
	}

	if c.ConnectTimeout != 0 {
		clientConfig = clientConfig.SetConnectTimeout(c.ConnectTimeout)
	}

	if c.SocketTimeout != 0 {
		clientConfig = clientConfig.SetSocketTimeout(c.SocketTimeout)
	}

	if c.MinPoolSize != 0 {
		clientConfig = clientConfig.SetMinPoolSize(c.MinPoolSize)

	}

	if c.MaxPoolSize != 0 {
		clientConfig = clientConfig.SetMaxPoolSize(c.MaxPoolSize)
	}

	if c.ReadConcern != "" {
		clientConfig = clientConfig.SetReadConcern(&readconcern.ReadConcern{
			Level: c.ReadConcern,
		})
	}

	if c.WriteConcern > 0 {
		clientConfig = clientConfig.SetWriteConcern(&writeconcern.WriteConcern{
			W: c.WriteConcern,
		})
	}

	if c.AppName != "" {
		clientConfig = clientConfig.SetAppName(c.AppName)
	}

	clientConfig = clientConfig.SetRetryWrites(c.RetryWrites)

	return clientConfig
}

func InitMongo(c *Config) {
	if c == nil || c.URI == "" {
		return
	}

	config = c
	GetMongoClient()

	log.Infof("MongoDB init success")
}

func GetMongoClient() *qmgo.Client {
	once.Do(func() {
		// 创建 MongoDB 客户端，只会执行一次
		clientOptions := config.BuildConfig()
		client, err := qmgo.NewClient(context.Background(), nil, options.ClientOptions{ClientOptions: clientOptions})
		if err != nil {
			log.Fatal(err)
		}

		sharedClient = client
	})

	return sharedClient
}

type Mapper struct {
	client     *qmgo.Collection
	database   string
	collection string
}

func NewMapper(collection string) (*Mapper, error) {
	return &Mapper{
		client:     GetMongoClient().Database(config.Database).Collection(collection),
		database:   config.Database,
		collection: collection,
	}, nil
}

func (m *Mapper) Insert(document any) error {
	_, err := m.client.InsertOne(context.Background(), document)
	return err
}

func (m *Mapper) Find(filter bson.M, result any) error {
	return m.client.Find(context.Background(), filter).One(result)
}

func (m *Mapper) FindArray(filter bson.M, result any) error {

	err := m.client.Find(context.Background(), filter).All(result)
	if err != nil {
		return err
	}

	return nil
}

// FindPage 分页查询文档
func (m *Mapper) FindPage(filter bson.M, pageNum, pageSize int64, results []any) error {
	// 计算要跳过的文档数量
	skip := (pageNum) * pageSize

	// 执行查询
	err := m.client.Find(context.Background(), filter).Skip(skip).Limit(pageSize).All(&results)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mapper) Update(filter, update bson.M) error {
	_, err := m.client.UpdateAll(context.Background(), filter, update)
	return err
}

func (m *Mapper) UpdateOne(filter, update bson.M) error {
	err := m.client.UpdateOne(context.Background(), filter, update)
	return err
}

func (m *Mapper) Delete(filter bson.M) error {
	_, err := m.client.RemoveAll(context.Background(), filter)
	return err
}

func (m *Mapper) DeleteOne(filter bson.M) error {
	err := m.client.Remove(context.Background(), filter)
	return err
}
