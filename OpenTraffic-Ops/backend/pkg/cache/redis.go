package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache Redis缓存封装
type Cache struct {
	client *redis.Client
	ctx    context.Context
}

// NewCache 创建缓存实例
func NewCache(client *redis.Client) *Cache {
	return &Cache{
		client: client,
		ctx:    context.Background(),
	}
}

// Set 设置缓存
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(c.ctx, key, value, expiration).Err()
}

// Get 获取缓存
func (c *Cache) Get(key string) (string, error) {
	return c.client.Get(c.ctx, key).Result()
}

// GetBytes 获取缓存字节
func (c *Cache) GetBytes(key string) ([]byte, error) {
	return c.client.Get(c.ctx, key).Bytes()
}

// Delete 删除缓存
func (c *Cache) Delete(keys ...string) error {
	return c.client.Del(c.ctx, keys...).Err()
}

// Exists 判断key是否存在
func (c *Cache) Exists(keys ...string) (int64, error) {
	return c.client.Exists(c.ctx, keys...).Result()
}

// Expire 设置过期时间
func (c *Cache) Expire(key string, expiration time.Duration) error {
	return c.client.Expire(c.ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (c *Cache) TTL(key string) (time.Duration, error) {
	return c.client.TTL(c.ctx, key).Result()
}

// SetJSON 设置JSON缓存
func (c *Cache) SetJSON(key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(c.ctx, key, value, expiration).Err()
}

// HSet Hash设置
func (c *Cache) HSet(key string, values ...interface{}) error {
	return c.client.HSet(c.ctx, key, values...).Err()
}

// HGet Hash获取
func (c *Cache) HGet(key, field string) (string, error) {
	return c.client.HGet(c.ctx, key, field).Result()
}

// HGetAll Hash获取全部
func (c *Cache) HGetAll(key string) (map[string]string, error) {
	return c.client.HGetAll(c.ctx, key).Result()
}

// HDel Hash删除字段
func (c *Cache) HDel(key string, fields ...string) error {
	return c.client.HDel(c.ctx, key, fields...).Err()
}

// SAdd 集合添加
func (c *Cache) SAdd(key string, members ...interface{}) error {
	return c.client.SAdd(c.ctx, key, members...).Err()
}

// SMembers 集合获取所有成员
func (c *Cache) SMembers(key string) ([]string, error) {
	return c.client.SMembers(c.ctx, key).Result()
}

// SRem 集合移除成员
func (c *Cache) SRem(key string, members ...interface{}) error {
	return c.client.SRem(c.ctx, key, members...).Err()
}

// Incr 自增
func (c *Cache) Incr(key string) (int64, error) {
	return c.client.Incr(c.ctx, key).Result()
}

// IncrBy 按指定值自增
func (c *Cache) IncrBy(key string, value int64) (int64, error) {
	return c.client.IncrBy(c.ctx, key, value).Result()
}

// Decr 自减
func (c *Cache) Decr(key string) (int64, error) {
	return c.client.Decr(c.ctx, key).Result()
}

// Keys 查找key
func (c *Cache) Keys(pattern string) ([]string, error) {
	return c.client.Keys(c.ctx, pattern).Result()
}

// DeleteKeys 批量删除匹配模式的key
func (c *Cache) DeleteKeys(pattern string) error {
	keys, err := c.Keys(pattern)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return c.Delete(keys...)
	}
	return nil
}

// Lock 分布式锁（基于SET NX EX）
func (c *Cache) Lock(key string, expiration time.Duration) (bool, error) {
	return c.client.SetNX(c.ctx, key, "1", expiration).Result()
}

// Unlock 释放分布式锁
func (c *Cache) Unlock(key string) error {
	return c.Delete(key)
}

// CacheKey 构建缓存key
func CacheKey(prefix string, parts ...interface{}) string {
	key := prefix
	for _, part := range parts {
		key = fmt.Sprintf("%s:%v", key, part)
	}
	return key
}
