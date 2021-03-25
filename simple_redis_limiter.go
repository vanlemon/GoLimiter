package limiter

import (
	"sync/atomic"
	"time"

	"github.com/go-redis/redis"

	logs "lmf.mortal.com/GoLogs"
)

// Redis 限流器
type SimpleRedisLimiter struct {
	Limiter
	key         string        // 限流器唯一标识
	limit       int64         // 限流器限额
	redisClient *redis.Client // redis 客户端
}

func NewSimpleRedisLimiter(key string, limit int64, redisClient *redis.Client) *SimpleRedisLimiter {
	if limit <= 0 { // 限额应大于零
		logs.CtxError(logs.SysCtx, "[Running Limiter] NewSimpleRedisLimiter bad params, key: %s, limit: %d", key, limit)
		return nil
	}

	return &SimpleRedisLimiter{
		key:         "simple-redis-limiter-" + key,
		limit:       limit,
		redisClient: redisClient,
	}
}

func (s *SimpleRedisLimiter) SetLimit(newLimit int64) (oldLimit int64) {
	oldLimit = atomic.SwapInt64(&s.limit, newLimit) // 原子操作
	logs.CtxInfo(logs.SysCtx, "[Running Limiter] SimpleRedisLimiter SetLimit key: %s, old: %d, new: %d", s.key, oldLimit, newLimit)
	return oldLimit
}

func (s *SimpleRedisLimiter) GetLimit() (oldLimit int64) {
	return s.limit
}

func (s *SimpleRedisLimiter) SetLimitAndStep(newLimit int64, newStep int64) (oldLimit int64, oldStep int64) {
	return s.SetLimit(newLimit), 0
}

func (s *SimpleRedisLimiter) CanPass() (pass bool) {
	// 以 redis 唯一标识为键，递增 redis 的值
	// key: limiter_name + service_name + method_name
	val := s.redisClient.Incr(s.key).Val()
	logs.CtxInfo(logs.SysCtx, "[Running Limiter] SimpleRedisLimiter val: %+v", val)
	if val == 1 { // 如果值为 1，则为新建限流器计数，初始化键的过期时间
		s.redisClient.PExpire(s.key, time.Second)
		logs.CtxInfo(logs.SysCtx, "[Running Limiter] SimpleRedisLimiter start new counter, val: %+v", val)
	}

	// 防御性代码，如果 key 没有被设置过期时间需要能修复
	dur := s.redisClient.PTTL(s.key).Val()
	logs.CtxInfo(logs.SysCtx, "[Running Limiter] SimpleRedisLimiter pttl: %+v", dur/time.Millisecond)
	if dur < 0 {
		// 没有 key 或未设置过期时间
		s.redisClient.PExpire(s.key, time.Second)
		logs.CtxInfo(logs.SysCtx, "[Running Limiter] SimpleRedisLimiter start new counter, val: %+v", val)
	}

	return val <= s.limit // 递增值小于限额值则请求通过
}
