package cache

import (
	"context"
	"github.com/ambitiousmice/go-one/common/json"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/utils"
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

func Gets(key string) (string, error) {
	return RedisClient.Get(key)
}

func Sets(key string, value string, expiration time.Duration) error {
	return RedisClient.Set(key, value, expiration)
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

// 添加一个元素, zset与set最大的区别就是每个元素都有一个score，因此有个排序的辅助功能;  zadd
func ZSetAdd(key string, value string, score float64) error {
	return RedisClient.ZSetAdd(key, value, score)
}

// 删除元素 zrem
func ZSetRemove(key string, value string) error {
	return RedisClient.ZSetRemove(key, value)
}

// score的增加or减少 zincrby
func ZSetIncrScore(key string, value string, score float64) (float64, error) {
	return RedisClient.ZSetIncrScore(key, value, score)
}

// 查询value对应的score   zscore
func ZSetScore(key string, value string) (float64, error) {
	return RedisClient.ZSetScore(key, value)
}

// 判断value在zset中的排名  zrank
func ZSetRankAsc(key string, value string) (int64, error) {
	return RedisClient.ZSetRankAsc(key, value)
}

// 判断value在zset中的排名  zrank
func ZSetRankDesc(key string, value string) (int64, error) {
	return RedisClient.ZSetRankDesc(key, value)
}

// 返回集合的长度
func ZSetSize(key string) (int64, error) {
	return RedisClient.ZSetSize(key)
}

// 询集合中指定顺序的值， 0 -1 表示获取全部的集合内容  zrange
func ZSetRangeWithScore(key string, start int64, end int64) ([]redis.Z, error) {
	return RedisClient.ZSetRangeWithScore(key, start, end)
}

// 询集合中指定顺序的值， 0 -1 表示获取全部的集合内容  倒序
func ZSetRevRangeWithScore(key string, start int64, end int64) ([]redis.Z, error) {
	return RedisClient.ZSetRevRangeWithScore(key, start, end)
}

// 询集合中指定顺序的值， 0 -1 表示获取全部的集合内容  倒序
func ZSetRevRange(key string, start int64, end int64) ([]string, error) {
	return RedisClient.ZSetRevRange(key, start, end)
}

// 询集合中指定顺序的值， 0 -1 表示获取全部的集合内容
func ZSetRange(key string, start int64, end int64) ([]string, error) {
	return RedisClient.ZSetRange(key, start, end)
}

// 询集合中指定score下的value
func ZSetRangeByScore(key string, min float64, max float64) ([]string, error) {
	return RedisClient.ZSetRangeByScore(key, min, max)
}

func ZSetRankAscWithScore(key string, value string) (redis.RankScore, error) {
	return RedisClient.ZSetRankAscWithScore(key, value)
}

// 判断value在zset中的排名  zrank
func ZSetRankDescWithScore(key string, value string) (redis.RankScore, error) {
	return RedisClient.ZSetRankDescWithScore(key, value)
}

// 分布式锁功能
func SetNx(key string, value string, expiration time.Duration) (bool, error) {
	return RedisClient.SetNx(key, value, expiration)
}

func Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	return RedisClient.Scan(cursor, match, count)
}

func ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	return RedisClient.ZRemRangeByRank(key, start, stop)
}

func SAdd(key string, members ...interface{}) (int64, error) {
	return RedisClient.SAdd(key, members)
}

func SIsMember(key string, member interface{}) (bool, error) {
	return RedisClient.SIsMember(key, member)
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

	//添加一个元素, zset与set最大的区别就是每个元素都有一个score，因此有个排序的辅助功能;  zadd
	ZSetAdd(key string, value string, score float64) error
	//删除元素 zrem
	ZSetRemove(key string, value string) error
	//score的增加or减少 zincrby
	ZSetIncrScore(key string, value string, score float64) (float64, error)
	//查询value对应的score   zscore
	ZSetScore(key string, value string) (float64, error)
	//判断value在zset中的排名  zrank
	ZSetRankAsc(key string, value string) (int64, error)
	//判断value在zset中的排名  zrank
	ZSetRankDesc(key string, value string) (int64, error)

	//判断value在zset中的排名  zrank
	ZSetRankAscWithScore(key string, value string) (redis.RankScore, error)
	//判断value在zset中的排名  zrank
	ZSetRankDescWithScore(key string, value string) (redis.RankScore, error)
	//返回集合的长度
	ZSetSize(key string) (int64, error)
	//询集合中指定顺序的值， 0 -1 表示获取全部的集合内容  zrange
	ZSetRangeWithScore(key string, start int64, end int64) ([]redis.Z, error)
	//询集合中指定顺序的值， 0 -1 表示获取全部的集合内容  倒序
	ZSetRevRangeWithScore(key string, start int64, end int64) ([]redis.Z, error)
	ZSetRevRange(key string, start int64, end int64) ([]string, error)
	ZSetRange(key string, start int64, end int64) ([]string, error)
	ZSetRangeByScore(key string, min float64, max float64) ([]string, error)

	ZRemRangeByRank(key string, start, stop int64) (int64, error)

	SetNx(key, value string, expiration time.Duration) (bool, error)
	Scan(cursor uint64, match string, count int64) ([]string, uint64, error)

	SAdd(key string, members ...interface{}) (int64, error)
	SIsMember(key string, member interface{}) (bool, error)
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

func (r *SingleNodeRedisClient) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	return r.client.ZRemRangeByRank(r.ctx, key, start, stop).Result()
}

func (r *SingleNodeRedisClient) SAdd(key string, members ...interface{}) (int64, error) {
	return r.client.SAdd(r.ctx, key, members).Result()
}

func (r *SingleNodeRedisClient) SIsMember(key string, member interface{}) (bool, error) {
	return r.client.SIsMember(r.ctx, key, member).Result()
}

func (r *SingleNodeRedisClient) Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.client.Scan(r.ctx, cursor, match, count).Result()
}

func (r *SingleNodeRedisClient) SetNx(key, value string, expiration time.Duration) (bool, error) {
	return r.client.SetNX(r.ctx, key, value, expiration).Result()
}

func (r *SingleNodeRedisClient) ZSetRankAscWithScore(key string, value string) (redis.RankScore, error) {
	return r.client.ZRankWithScore(r.ctx, key, value).Result()
}

func (r *SingleNodeRedisClient) ZSetRankDescWithScore(key string, value string) (redis.RankScore, error) {
	return r.client.ZRevRankWithScore(r.ctx, key, value).Result()
}

func (r *SingleNodeRedisClient) ZSetAdd(key string, value string, score float64) error {
	return r.client.ZAdd(r.ctx, key, redis.Z{
		Score:  score,
		Member: value,
	}).Err()
}

func (r *SingleNodeRedisClient) ZSetRemove(key string, value string) error {
	return r.client.ZRem(r.ctx, key, value).Err()
}

func (r *SingleNodeRedisClient) ZSetIncrScore(key string, value string, score float64) (float64, error) {
	return r.client.ZIncrBy(r.ctx, key, score, value).Result()
}

func (r *SingleNodeRedisClient) ZSetScore(key string, value string) (float64, error) {
	return r.client.ZScore(r.ctx, key, value).Result()
}

func (r *SingleNodeRedisClient) ZSetRankAsc(key string, value string) (int64, error) {
	return r.client.ZRank(r.ctx, key, value).Result()
}

func (r *SingleNodeRedisClient) ZSetRankDesc(key string, value string) (int64, error) {
	return r.client.ZRevRank(r.ctx, key, value).Result()
}

func (r *SingleNodeRedisClient) ZSetSize(key string) (int64, error) {
	return r.client.ZCard(r.ctx, key).Result()
}

func (r *SingleNodeRedisClient) ZSetRangeWithScore(key string, start int64, end int64) ([]redis.Z, error) {
	return r.client.ZRangeWithScores(r.ctx, key, start, end).Result()
}

func (r *SingleNodeRedisClient) ZSetRevRangeWithScore(key string, start int64, end int64) ([]redis.Z, error) {
	return r.client.ZRevRangeWithScores(r.ctx, key, start, end).Result()
}

func (r *SingleNodeRedisClient) ZSetRevRange(key string, start int64, end int64) ([]string, error) {
	return r.client.ZRevRange(r.ctx, key, start, end).Result()
}

func (r *SingleNodeRedisClient) ZSetRange(key string, start int64, end int64) ([]string, error) {
	return r.client.ZRange(r.ctx, key, start, end).Result()
}

func (r *SingleNodeRedisClient) ZSetRangeByScore(key string, min float64, max float64) ([]string, error) {
	return r.client.ZRangeByScore(r.ctx, key, &redis.ZRangeBy{
		Min: utils.ToString(min),
		Max: utils.ToString(max),
	}).Result()
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

func (r *SentinelRedisClient) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	return r.client.ZRemRangeByRank(r.ctx, key, start, stop).Result()
}

func (r *SentinelRedisClient) SAdd(key string, members ...interface{}) (int64, error) {
	return r.client.SAdd(r.ctx, key, members).Result()
}

func (r *SentinelRedisClient) SIsMember(key string, member interface{}) (bool, error) {
	return r.client.SIsMember(r.ctx, key, member).Result()
}

func (r *SentinelRedisClient) Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.client.Scan(r.ctx, cursor, match, count).Result()
}

func (r *SentinelRedisClient) SetNx(key, value string, expiration time.Duration) (bool, error) {
	return r.client.SetNX(r.ctx, key, value, expiration).Result()
}

func (r *SentinelRedisClient) ZSetRankAscWithScore(key string, value string) (redis.RankScore, error) {
	return r.client.ZRankWithScore(r.ctx, key, value).Result()
}

func (r *SentinelRedisClient) ZSetRankDescWithScore(key string, value string) (redis.RankScore, error) {
	return r.client.ZRevRankWithScore(r.ctx, key, value).Result()
}

func (r *SentinelRedisClient) ZSetAdd(key string, value string, score float64) error {
	return r.client.ZAdd(r.ctx, key, redis.Z{
		Score:  score,
		Member: value,
	}).Err()
}

func (r *SentinelRedisClient) ZSetRemove(key string, value string) error {
	return r.client.ZRem(r.ctx, key, value).Err()
}

func (r *SentinelRedisClient) ZSetIncrScore(key string, value string, score float64) (float64, error) {
	return r.client.ZIncrBy(r.ctx, key, score, value).Result()
}

func (r *SentinelRedisClient) ZSetScore(key string, value string) (float64, error) {
	return r.client.ZScore(r.ctx, key, value).Result()
}

func (r *SentinelRedisClient) ZSetRankAsc(key string, value string) (int64, error) {
	return r.client.ZRank(r.ctx, key, value).Result()
}

func (r *SentinelRedisClient) ZSetRankDesc(key string, value string) (int64, error) {
	return r.client.ZRevRank(r.ctx, key, value).Result()
}

func (r *SentinelRedisClient) ZSetSize(key string) (int64, error) {
	return r.client.ZCard(r.ctx, key).Result()
}

func (r *SentinelRedisClient) ZSetRangeWithScore(key string, start int64, end int64) ([]redis.Z, error) {
	return r.client.ZRangeWithScores(r.ctx, key, start, end).Result()
}

func (r *SentinelRedisClient) ZSetRevRangeWithScore(key string, start int64, end int64) ([]redis.Z, error) {
	return r.client.ZRevRangeWithScores(r.ctx, key, start, end).Result()
}

func (r *SentinelRedisClient) ZSetRevRange(key string, start int64, end int64) ([]string, error) {
	return r.client.ZRevRange(r.ctx, key, start, end).Result()
}

func (r *SentinelRedisClient) ZSetRange(key string, start int64, end int64) ([]string, error) {
	return r.client.ZRange(r.ctx, key, start, end).Result()
}

func (r *SentinelRedisClient) ZSetRangeByScore(key string, min float64, max float64) ([]string, error) {
	return r.client.ZRangeByScore(r.ctx, key, &redis.ZRangeBy{
		Min: utils.ToString(min),
		Max: utils.ToString(max),
	}).Result()
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

func (r *ClusterRedisClient) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	return r.client.ZRemRangeByRank(r.ctx, key, start, stop).Result()
}

func (r *ClusterRedisClient) SAdd(key string, members ...interface{}) (int64, error) {
	return r.client.SAdd(r.ctx, key, members).Result()
}

func (r *ClusterRedisClient) SIsMember(key string, member interface{}) (bool, error) {
	return r.client.SIsMember(r.ctx, key, member).Result()
}

func (r *ClusterRedisClient) Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.client.Scan(r.ctx, cursor, match, count).Result()
}

func (r *ClusterRedisClient) SetNx(key, value string, expiration time.Duration) (bool, error) {
	return r.client.SetNX(r.ctx, key, value, expiration).Result()
}

func (r *ClusterRedisClient) ZSetRankAscWithScore(key string, value string) (redis.RankScore, error) {
	return r.client.ZRankWithScore(r.ctx, key, value).Result()
}

func (r *ClusterRedisClient) ZSetRankDescWithScore(key string, value string) (redis.RankScore, error) {
	return r.client.ZRevRankWithScore(r.ctx, key, value).Result()
}

func (r *ClusterRedisClient) ZSetAdd(key string, value string, score float64) error {

	return r.client.ZAdd(r.ctx, key, redis.Z{
		Score:  score,
		Member: value,
	}).Err()
}

func (r *ClusterRedisClient) ZSetRemove(key string, value string) error {

	return r.client.ZRem(r.ctx, key, value).Err()
}

func (r *ClusterRedisClient) ZSetIncrScore(key string, value string, score float64) (float64, error) {
	return r.client.ZIncrBy(r.ctx, key, score, value).Result()
}

func (r *ClusterRedisClient) ZSetScore(key string, value string) (float64, error) {
	return r.client.ZScore(r.ctx, key, value).Result()
}

func (r *ClusterRedisClient) ZSetRankAsc(key string, value string) (int64, error) {
	return r.client.ZRank(r.ctx, key, value).Result()
}

func (r *ClusterRedisClient) ZSetRankDesc(key string, value string) (int64, error) {
	return r.client.ZRevRank(r.ctx, key, value).Result()
}

func (r *ClusterRedisClient) ZSetSize(key string) (int64, error) {
	return r.client.ZCard(r.ctx, key).Result()
}

func (r *ClusterRedisClient) ZSetRangeWithScore(key string, start int64, end int64) ([]redis.Z, error) {
	return r.client.ZRangeWithScores(r.ctx, key, start, end).Result()
}

func (r *ClusterRedisClient) ZSetRevRangeWithScore(key string, start int64, end int64) ([]redis.Z, error) {
	return r.client.ZRevRangeWithScores(r.ctx, key, start, end).Result()
}

func (r *ClusterRedisClient) ZSetRevRange(key string, start int64, end int64) ([]string, error) {
	return r.client.ZRevRange(r.ctx, key, start, end).Result()
}

func (r *ClusterRedisClient) ZSetRange(key string, start int64, end int64) ([]string, error) {
	return r.client.ZRange(r.ctx, key, start, end).Result()
}

func (r *ClusterRedisClient) ZSetRangeByScore(key string, min float64, max float64) ([]string, error) {
	return r.client.ZRangeByScore(r.ctx, key, &redis.ZRangeBy{
		Min: utils.ToString(min),
		Max: utils.ToString(max),
	}).Result()
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
