package services

import (
	"fmt"
	"sync"
	"time"

	"asd/conf"

	"github.com/beego/beego/v2/core/logs"
	"github.com/go-redis/redis/v8"
)

var (
	redisClient *RedisClient
	redisOnce   sync.Once
)

// RedisClient Redis客户端封装
type RedisClient struct {
	client *redis.Client
}

// GetRedisClient 获取Redis客户端实例（单例模式）
func GetRedisClient() *RedisClient {
	redisOnce.Do(func() {
		redisClient = newRedisClient()
	})
	return redisClient
}

// newRedisClient 创建新的Redis客户端
func newRedisClient() *RedisClient {
	// 从配置中读取Redis连接信息
	redisHost := conf.CONFIG.Redis.Host
	redisPort := conf.CONFIG.Redis.Port
	redisPassword := conf.CONFIG.Redis.Password
	redisDB := conf.CONFIG.Redis.DB

	// 如果没有配置Redis，使用默认值
	if redisHost == "" {
		redisHost = "127.0.0.1"
	}
	if redisPort == 0 {
		redisPort = 6379
	}

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPassword,
		DB:       redisDB,
	})

	logs.Info("Redis客户端已初始化，连接地址: %s:%d", redisHost, redisPort)

	return &RedisClient{
		client: client,
	}
}

// RPush 向列表右侧添加元素
func (r *RedisClient) RPush(key string, value interface{}) error {
	ctx := r.client.Context()
	return r.client.RPush(ctx, key, value).Err()
}

// LPop 从列表左侧弹出元素
func (r *RedisClient) LPop(key string) (string, error) {
	ctx := r.client.Context()
	return r.client.LPop(ctx, key).Result()
}

// BLPop 阻塞式从列表左侧弹出元素
func (r *RedisClient) BLPop(timeout int, key string) ([]string, error) {
	ctx := r.client.Context()
	return r.client.BLPop(ctx, time.Duration(timeout)*time.Second, key).Result()
}

// LRange 获取列表指定范围的元素
func (r *RedisClient) LRange(key string, start, stop int64) ([]string, error) {
	ctx := r.client.Context()
	return r.client.LRange(ctx, key, start, stop).Result()
}

// LLen 获取列表长度
func (r *RedisClient) LLen(key string) (int64, error) {
	ctx := r.client.Context()
	return r.client.LLen(ctx, key).Result()
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(key string) (bool, error) {
	ctx := r.client.Context()
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Expire 设置键的过期时间
func (r *RedisClient) Expire(key string, seconds int) error {
	ctx := r.client.Context()
	return r.client.Expire(ctx, key, time.Duration(seconds)*time.Second).Err()
}

// Del 删除键
func (r *RedisClient) Del(key string) error {
	ctx := r.client.Context()
	return r.client.Del(ctx, key).Err()
}

// HSet 设置哈希表字段的值
func (r *RedisClient) HSet(key, field string, value interface{}) error {
	ctx := r.client.Context()
	return r.client.HSet(ctx, key, field, value).Err()
}

// HGet 获取哈希表字段的值
func (r *RedisClient) HGet(key, field string) (string, error) {
	ctx := r.client.Context()
	return r.client.HGet(ctx, key, field).Result()
}

// SetEx 设置带过期时间的键值对
func (r *RedisClient) SetEx(key string, value interface{}, expiration time.Duration) error {
	ctx := r.client.Context()
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取键值
func (r *RedisClient) Get(key string) (string, error) {
	ctx := r.client.Context()
	return r.client.Get(ctx, key).Result()
}

// HGetAll 获取哈希表中的所有字段和值
func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	ctx := r.client.Context()
	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		logs.Error("Redis HGetAll 失败: %v, key: %s", err, key)
		return nil, err
	}
	return result, nil
}
