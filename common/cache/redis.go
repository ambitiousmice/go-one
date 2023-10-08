package cache

import (
	"context"
	"go-one/common/json"
	"go-one/common/log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient IRedisClient

var redisConfig *RedisConfig

func InitRedis(c *RedisConfig) {
	if c == nil || c.Addr == "" {
		return
	}
	redisConfig = c

	if c.ClientType == "sentinel" {
		RedisClient, _ = NewSentinelRedisClient()
	} else if c.ClientType == "cluster" {
		RedisClient, _ = NewClusterRedisClient()
	} else {
		RedisClient, _ = NewSingleNodeRedisClient()
	}

	log.Infof("redis client init success,addr:%s", c.Addr)
}

func Get(key string, v any) error {
	value, err := RedisClient.Get(key)
	if err == nil {
		return json.UnmarshalFromString(value, v)
	} else {
		return err
	}
}

func Set(key string, v any, expiration time.Duration) error {
	value, err := json.MarshalToString(v)
	if err == nil {
		return RedisClient.Set(key, value, expiration)
	} else {
		return err
	}
}

func Delete(key string) error {
	return RedisClient.Delete(key)
}

func AddToSortedSet(key string, members ...redis.Z) error {
	return RedisClient.AddToSortedSet(key, members...)
}

func GetSortedSetRange(key string, start, stop int64) ([]string, error) {
	return RedisClient.GetSortedSetRange(key, start, stop)
}

func SetHashField(key, field string, v any) error {
	value, err := json.MarshalToString(v)
	if err == nil {
		return RedisClient.SetHashField(key, field, value)
	} else {
		return err
	}
}

func GetHashField(key, field string, v any) error {
	value, err := RedisClient.GetHashField(key, field)
	if err == nil {
		return json.UnmarshalFromString(value, v)
	} else {
		return err
	}
}

func ExecuteLuaScript(script string, keys []string, args []any) (any, error) {
	return RedisClient.ExecuteLuaScript(script, keys, args)
}

// IRedisClient 是通用的 Redis 客户端接口
type IRedisClient interface {
	Close()
	Set(key, value string, expiration time.Duration) error
	Get(key string) (string, error)
	Delete(key string) error
	AddToSortedSet(key string, members ...redis.Z) error
	GetSortedSetRange(key string, start, stop int64) ([]string, error)
	SetHashField(key, field, value string) error
	GetHashField(key, field string) (string, error)
	ExecuteLuaScript(script string, keys []string, args []any) (any, error)
}

type RedisConfig struct {
	ClientType   string        `yaml:"client-type"`
	Addr         string        `yaml:"addr"`
	MasterName   string        `yaml:"master-name"`
	Password     string        `yaml:"password"`
	DB           int           `yaml:"DB"`
	MaxRetries   int           `yaml:"max-retries"`
	PoolSize     int           `yaml:"pool-size"`
	MinIdleConns int           `yaml:"min-idle-conns"`
	MaxIdleConns int           `yaml:"max-idle-conns"`
	IdleTimeout  time.Duration `yaml:"idle-timeout"`
	// 其他常用选项可以根据需要添加
}

// SingleNodeRedisClient 是单机模式 Redis 客户端结构体
type SingleNodeRedisClient struct {
	client *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSingleNodeRedisClient 创建一个新的单机模式 Redis 客户端
func NewSingleNodeRedisClient() (*SingleNodeRedisClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	client := redis.NewClient(&redis.Options{
		Addr:            redisConfig.Addr,
		Password:        redisConfig.Password,
		DB:              redisConfig.DB,
		MaxRetries:      redisConfig.MaxRetries,
		PoolSize:        redisConfig.PoolSize,
		MinIdleConns:    redisConfig.MinIdleConns,
		MaxIdleConns:    redisConfig.MaxIdleConns,
		ConnMaxIdleTime: redisConfig.IdleTimeout,
	})

	return &SingleNodeRedisClient{
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Close 关闭 Redis 客户端连接
func (r *SingleNodeRedisClient) Close() {
	r.client.Close()
	r.cancel()
}

// Set 设置键值对
func (r *SingleNodeRedisClient) Set(key, value string, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Get 获取键的值
func (r *SingleNodeRedisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// Delete 删除键
func (r *SingleNodeRedisClient) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// AddToSortedSet 向 Sorted Set 添加成员
func (r *SingleNodeRedisClient) AddToSortedSet(key string, members ...redis.Z) error {
	return r.client.ZAdd(r.ctx, key, members...).Err()
}

// GetSortedSetRange 获取 Sorted Set 指定范围的成员
func (r *SingleNodeRedisClient) GetSortedSetRange(key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(r.ctx, key, start, stop).Result()
}

// SetHashField 设置 Hash 中的字段值
func (r *SingleNodeRedisClient) SetHashField(key, field, value string) error {
	return r.client.HSet(r.ctx, key, field, value).Err()
}

// GetHashField 获取 Hash 中的字段值
func (r *SingleNodeRedisClient) GetHashField(key, field string) (string, error) {
	return r.client.HGet(r.ctx, key, field).Result()
}

// ExecuteLuaScript 执行 Lua 脚本
func (r *SingleNodeRedisClient) ExecuteLuaScript(script string, keys []string, args []any) (any, error) {
	return r.client.Eval(r.ctx, script, keys, args...).Result()
}

// SentinelRedisClient 是哨兵模式 Redis 客户端结构体
type SentinelRedisClient struct {
	client *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSentinelRedisClient 创建一个新的哨兵模式 Redis 客户端
func NewSentinelRedisClient() (*SentinelRedisClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	sentinelAddrs := strings.Split(redisConfig.Addr, ",")
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		SentinelAddrs:   sentinelAddrs,
		MasterName:      redisConfig.MasterName,
		Password:        redisConfig.Password,
		DB:              redisConfig.DB,
		MaxRetries:      redisConfig.MaxRetries,
		PoolSize:        redisConfig.PoolSize,
		MinIdleConns:    redisConfig.MinIdleConns,
		MaxIdleConns:    redisConfig.MaxIdleConns,
		ConnMaxIdleTime: redisConfig.IdleTimeout,
	})

	return &SentinelRedisClient{
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Close 关闭 Redis 客户端连接
func (r *SentinelRedisClient) Close() {
	r.client.Close()
	r.cancel()
}

// Set 设置键值对
func (r *SentinelRedisClient) Set(key, value string, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Get 获取键的值
func (r *SentinelRedisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// Delete 删除键
func (r *SentinelRedisClient) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// AddToSortedSet 向 Sorted Set 添加成员
func (r *SentinelRedisClient) AddToSortedSet(key string, members ...redis.Z) error {
	return r.client.ZAdd(r.ctx, key, members...).Err()
}

// GetSortedSetRange 获取 Sorted Set 指定范围的成员
func (r *SentinelRedisClient) GetSortedSetRange(key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(r.ctx, key, start, stop).Result()
}

// SetHashField 设置 Hash 中的字段值
func (r *SentinelRedisClient) SetHashField(key, field, value string) error {
	return r.client.HSet(r.ctx, key, field, value).Err()
}

// GetHashField 获取 Hash 中的字段值
func (r *SentinelRedisClient) GetHashField(key, field string) (string, error) {
	return r.client.HGet(r.ctx, key, field).Result()
}

// ExecuteLuaScript 执行 Lua 脚本
func (r *SentinelRedisClient) ExecuteLuaScript(script string, keys []string, args []any) (any, error) {
	return r.client.Eval(r.ctx, script, keys, args...).Result()
}

// ClusterRedisClient 是集群模式 Redis 客户端结构体
type ClusterRedisClient struct {
	client *redis.ClusterClient
	ctx    context.Context
	cancel context.CancelFunc
}

// NewClusterRedisClient 创建一个新的集群模式 Redis 客户端
func NewClusterRedisClient() (*ClusterRedisClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	clusterAddrs := strings.Split(redisConfig.Addr, ",")
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           clusterAddrs,
		Password:        redisConfig.Password,
		MaxRetries:      redisConfig.MaxRetries,
		PoolSize:        redisConfig.PoolSize,
		MinIdleConns:    redisConfig.MinIdleConns,
		MaxIdleConns:    redisConfig.MaxIdleConns,
		ConnMaxIdleTime: redisConfig.IdleTimeout,
	})

	return &ClusterRedisClient{
		client: client,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Close 关闭 Redis 客户端连接
func (r *ClusterRedisClient) Close() {
	r.client.Close()
	r.cancel()
}

// Set 设置键值对
func (r *ClusterRedisClient) Set(key, value string, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Get 获取键的值
func (r *ClusterRedisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// Delete 删除键
func (r *ClusterRedisClient) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// AddToSortedSet 向 Sorted Set 添加成员
func (r *ClusterRedisClient) AddToSortedSet(key string, members ...redis.Z) error {
	return r.client.ZAdd(r.ctx, key, members...).Err()
}

// GetSortedSetRange 获取 Sorted Set 指定范围的成员
func (r *ClusterRedisClient) GetSortedSetRange(key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(r.ctx, key, start, stop).Result()
}

// SetHashField 设置 Hash 中的字段值
func (r *ClusterRedisClient) SetHashField(key, field, value string) error {
	return r.client.HSet(r.ctx, key, field, value).Err()
}

// GetHashField 获取 Hash 中的字段值
func (r *ClusterRedisClient) GetHashField(key, field string) (string, error) {
	return r.client.HGet(r.ctx, key, field).Result()
}

// ExecuteLuaScript 执行 Lua 脚本
func (r *ClusterRedisClient) ExecuteLuaScript(script string, keys []string, args []any) (any, error) {
	return r.client.Eval(r.ctx, script, keys, args...).Result()
}
